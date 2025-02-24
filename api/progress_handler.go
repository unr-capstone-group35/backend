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
	Status db.Status `json:"status"`
}

type ExerciseAttemptRequest struct {
	Answer interface{} `json:"answer"`
}

func (s *Server) handleGetCourseProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Recieved GET CourseProgress")

		courseID := r.PathValue("courseID")
		s.logger.Debug("Course ID from request", "courseID", courseID)

		if courseID == "" {
			s.logger.Warn("Error: Course ID is empty")
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		userID, ok := auth.GetUserID(r.Context())
		s.logger.Debug("User ID from context", "userID", userID, "ok", ok)

		if !ok {
			s.logger.Warn("Error: User not found in context")
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		// Check if course exists
		_, err := s.CourseService.GetCourseByID(courseID)
		if err != nil {
			s.logger.Warn("Error verifying course existence", "error", err)
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		s.logger.Debug("Course verification successful")

		// Get or create progress
		progress, err := s.CourseService.GetCourseProgress(userID, courseID)
		if err != nil {
			s.logger.Warn("Error getting course progress", "error", err)

			if err.Error() == "record not found" {
				s.logger.Debug("Creating new progress record", "userID", userID, "courseID", courseID)
				now := time.Now()
				progress = &db.CourseProgress{
					UserID:         userID,
					CourseID:       courseID,
					StartedAt:      now,
					LastAccessedAt: now,
				}
			} else {
				s.logger.Warn("Database error getting course progress", "error", err)
				http.Error(w, "Failed to get course progress", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(progress); err != nil {
			s.logger.Error("Error encoding progress response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleGetLessonProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")

		s.logger.Debug("Got GET lesson progress", "courseID", courseID, "lessonID", lessonID)
		if courseID == "" || lessonID == "" {
			http.Error(w, "Course ID and lesson ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := auth.GetUserID(r.Context())
		if !ok {
			s.logger.Warn("User not found in context")
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		progress, err := s.CourseService.GetLessonProgress(userID, courseID, lessonID)
		if err != nil {
			s.logger.Warn("Failed to get lesson progress", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get lesson progress: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(progress); err != nil {
			s.logger.Error("Failed to encode response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleUpdateLessonProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")

		s.logger.Debug("Got POST update lesson progress", "courseID", courseID, "lessonID", lessonID)

		var req UpdateProgressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.logger.Warn("Invalid request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Status == "" {
			s.logger.Warn("Status is required")
			http.Error(w, "Status is required", http.StatusBadRequest)
			return
		}

		if courseID == "" || lessonID == "" {
			s.logger.Warn("Course ID and lesson ID are required")
			http.Error(w, "Course ID and lesson ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := auth.GetUserID(r.Context())
		if !ok {
			s.logger.Warn("User not found in context")
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		if err := s.CourseService.UpdateLessonProgress(userID, courseID, lessonID, req.Status); err != nil {
			s.logger.Error("Failed to update lesson progress", "error", err)
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

		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")
		exerciseID := r.PathValue("exerciseID")

		log.Printf("Received exercise attempt - Course: %s, Lesson: %s, Exercise: %s, Answer: %v",
			courseID, lessonID, exerciseID, req.Answer)

		if courseID == "" || lessonID == "" || exerciseID == "" {
			http.Error(w, "Course ID, lesson ID, and exercise ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := auth.GetUserID(r.Context())
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
			UserID:     userID,
			CourseID:   courseID,
			LessonID:   lessonID,
			ExerciseID: exerciseID,
			Answer:     string(answerJSON),
		}

		// Verify answer and record attempt
		isCorrect, err := s.CourseService.VerifyExerciseAnswer(courseID, lessonID, exerciseID, req.Answer)
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
