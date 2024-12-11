// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		if err := database.Close(); err != nil {
			log.Fatalf("DB close error: %v", err)
		} else {
			fmt.Println("Closed database")
		}

		if err := srv.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		} else {
			fmt.Println("Closed server")
		}
	}()

	fmt.Printf("Running server on %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
