// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/auth"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/user"
)

func main() {
	// Initialize database connection
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize services with database
	userService := user.NewService(database)

	// Initialize auth middleware
	authMiddleware := auth.NewMiddleware(database)

	// Initialize course store with database
	coursesStore := course.NewJSONStore("./data", database)
	if err := coursesStore.LoadCourseDir(); err != nil {
		log.Fatalf("Failed to load courses: %v", err)
	}
	log.Printf("Successfully connected to database")

	// Initialize server with all required dependencies
	server := api.NewServer(
		userService,
		coursesStore,
		authMiddleware,
		database,
		"8080",
	)

	// Create an HTTP server with adjusted timeouts
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("Running server on %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
