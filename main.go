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
	"github.com/mrz1836/postmark"
	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/services/points"
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
		logger.Warn(".env not found")
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	logger.Debug("Got env variables", "user", dbUser, "password", dbPassword, "name", dbName, "host", dbHost, "port", dbPort)

	database, err := db.NewDatabase(dbUser, dbPassword, dbName, dbHost, dbPort)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services with database
	userService := user.NewService(database)
	progressService := progress.NewService(database)
	sessionService := session.NewService(database)

	// Initialize the new points service
	pointsService := points.NewService(database)

	// Configure point values (OPTIONAL - uses defaults if not set)
	pointsConfig := points.PointsConfig{
		CorrectAnswerPoints:   10,
		StreakBonusMultiplier: 2,   // 2 points per streak level
		MaxStreakBonus:        50,  // Maximum 50 bonus points for streaks
		LessonCompletionBonus: 100, // 100 points for completing a lesson
		CourseCompletionBonus: 500, // 500 points for completing a course
	}
	pointsService.SetPointsConfig(pointsConfig)

	postmarkAPIKey := os.Getenv("POSTMARK_API_KEY")
	if postmarkAPIKey != "" {
		logger.Info("Initializing Postmark client")
		// Create a new Postmark client with the server token
		postmarkClient := postmark.NewClient(postmarkAPIKey, "")
		user.SetPostmarkClient(postmarkClient)
	} else {
		logger.Warn("No Postmark API key provided - password reset emails will only be logged")
	}

	// Initialize course store with database
	coursesStore := course.NewJSONStore("./data")
	if err := coursesStore.LoadCourseDir(); err != nil {
		logger.Error("Failed to load courses", "error", err)
		os.Exit(1)
	}

	if err := api.EnsureProfilePicDirectory(); err != nil {
		logger.Error("Error creating profile pictures directory", "error", err)
		os.Exit(1)
	}

	// Initialize server with all required dependencies
	server := api.NewServer(
		userService,
		coursesStore,
		progressService,
		sessionService,
		pointsService,
		database,
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
