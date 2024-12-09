// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tylerolson/capstone-backend/api"
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
	coursesStore := course.NewJSONStore("./data")
	if err := coursesStore.LoadCourseDir(); err != nil {
		log.Fatalf("Failed to load courses: %v", err)
	}
	log.Printf("Successfully connected to database")

	// Initialize server
	server := api.NewServer(userService, coursesStore, "8080")

	port := ":8080"
	fmt.Printf("Running server on %s\n", port)
	http.ListenAndServe(port, server.Mux)
}
