package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/services/progress"
	"github.com/tylerolson/capstone-backend/services/session"
	"github.com/tylerolson/capstone-backend/services/user"
)

func main() {
	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}),
	)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic(".env not found")
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	database, err := db.NewDatabase(dbUser, dbPassword, dbName, "", "")
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services with database
	userService := user.NewService(database)

	// Initialize course store with database
	coursesStore := course.NewJSONStore("./data")
	if err := coursesStore.LoadCourseDir(); err != nil {
		logger.Error("Failed to load courses", "error", err)
		os.Exit(1)
	}

	progressService := progress.NewService(database)
	sessionService := session.NewService(database)

	// Initialize server with all required dependencies
	server := api.NewServer(
		userService,
		coursesStore,
		progressService,
		sessionService,
		logger,
	)

	// Create an HTTP server with adjusted timeouts
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		if err := database.Close(); err != nil {
			logger.Error("DB close error", "error", err)
			os.Exit(1)
		} else {
			logger.Info("Closed database")
		}

		if err := srv.Close(); err != nil {
			logger.Error("HTTP close error", "error", err)
			os.Exit(1)
		} else {
			logger.Info("Closed server")
		}
	}()

	logger.Info("Started server", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Listen error", "error", err)
		os.Exit(1)
	}
}
