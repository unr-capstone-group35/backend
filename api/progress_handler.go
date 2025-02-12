package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tylerolson/capstone-backend/auth"
	"github.com/tylerolson/capstone-backend/db"
)

type UpdateProgressRequest struct {
	Status string `json:"status"`
}

type ExerciseAttemptRequest struct {
	Answer interface{} `json:"answer"`
}

func (s *Server) handleGetCourseProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Starting handleGetCourseProgress")

		courseName := r.PathValue("name")
		log.Printf("Course name from request: %s", courseName)

		if courseName == "" {
			log.Printf("Error: Course name is empty")
			http.Error(w, "Course name is required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		log.Printf("User ID from context: %d, ok: %v", userId, ok)

		if !ok {
			log.Printf("Error: User not found in context")
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		// Check if course exists
		_, err := s.CourseService.GetCourseByName(courseName)
		if err != nil {
			log.Printf("Error verifying course existence: %v", err)
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		log.Printf("Course verification successful")

		// Get or create progress
		progress, err := s.CourseService.GetCourseProgress(userId, courseName)
		if err != nil {
			log.Printf("Error getting course progress: %v", err)

			if err.Error() == "record not found" {
				log.Printf("Creating new progress record for user %d, course %s", userId, courseName)
				now := time.Now()
				progress = &db.CourseProgress{
					UserID:         userId,
					CourseName:     courseName,
					StartedAt:      now,
					LastAccessedAt: now,
				}
			} else {
				log.Printf("Database error getting course progress: %v", err)
				http.Error(w, "Failed to get course progress", http.StatusInternalServerError)
				return
			}
		}

		log.Printf("Sending progress response: %+v", progress)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(progress); err != nil {
			log.Printf("Error encoding progress response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully completed handleGetCourseProgress")
	}
}

func (s *Server) handleGetLessonProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")

		if courseName == "" || lessonId == "" {
			http.Error(w, "Course name and lesson ID are required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		progress, err := s.CourseService.GetLessonProgress(userId, courseName, lessonId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get lesson progress: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(progress); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleUpdateLessonProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateProgressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Status == "" {
			http.Error(w, "Status is required", http.StatusBadRequest)
			return
		}

		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")

		if courseName == "" || lessonId == "" {
			http.Error(w, "Course name and lesson ID are required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		if err := s.CourseService.UpdateLessonProgress(userId, courseName, lessonId, req.Status); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update lesson progress: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleExerciseAttempt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExerciseAttemptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")
		exerciseId := r.PathValue("exerciseId")

		log.Printf("Received exercise attempt - Course: %s, Lesson: %s, Exercise: %s, Answer: %v",
			courseName, lessonId, exerciseId, req.Answer)

		if courseName == "" || lessonId == "" || exerciseId == "" {
			http.Error(w, "Course name, lesson ID, and exercise ID are required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		// Convert answer to JSON string
		answerJSON, err := json.Marshal(req.Answer)
		if err != nil {
			http.Error(w, "Failed to process answer data", http.StatusInternalServerError)
			return
		}

		attempt := &db.ExerciseAttempt{
			UserID:     userId,
			CourseName: courseName,
			LessonID:   lessonId,
			ExerciseID: exerciseId,
			Answer:     string(answerJSON),
		}

		// Verify answer and record attempt
		isCorrect, err := s.CourseService.VerifyExerciseAnswer(courseName, lessonId, exerciseId, req.Answer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify exercise answer: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Exercise verification result - isCorrect: %v", isCorrect)
		attempt.IsCorrect = isCorrect

		// Record the attempt using the DB field
		if err := s.DB.RecordExerciseAttempt(attempt); err != nil {
			http.Error(w, fmt.Sprintf("Failed to record exercise attempt: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"isCorrect": isCorrect,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
