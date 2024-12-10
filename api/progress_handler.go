// api/progress_handler.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		courseName := r.PathValue("name")
		if courseName == "" {
			http.Error(w, "Course name is required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		progress, err := s.CourseService.GetCourseProgress(userId, courseName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get course progress: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(progress); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
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
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Answer == nil {
			http.Error(w, "Answer is required", http.StatusBadRequest)
			return
		}

		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")
		exerciseId := r.PathValue("exerciseId")

		if courseName == "" || lessonId == "" || exerciseId == "" {
			http.Error(w, "Course name, lesson ID, and exercise ID are required", http.StatusBadRequest)
			return
		}

		userId, ok := auth.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		attempt := &db.ExerciseAttempt{
			UserID:     userId,
			CourseName: courseName,
			LessonID:   lessonId,
			ExerciseID: exerciseId,
			Answer:     req.Answer,
		}

		// Verify answer and record attempt
		isCorrect, err := s.CourseService.VerifyExerciseAnswer(courseName, lessonId, exerciseId, req.Answer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify exercise answer: %v", err), http.StatusInternalServerError)
			return
		}

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
