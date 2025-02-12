package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/user"
)

const PORT = "8080"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Initialize database connection
	database, err := db.NewDatabase()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services with database
	userService := user.NewService(database)

	// Initialize course store with database
	coursesStore := course.NewJSONStore("./data", database)
	if err := coursesStore.LoadCourseDir(); err != nil {
		logger.Error("Failed to load courses", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully connected to database")

	// Initialize server with all required dependencies
	server := api.NewServer(
		userService,
		coursesStore,
		database,
		logger,
	)

	// Create an HTTP server with adjusted timeouts
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", PORT),
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
