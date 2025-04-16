package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tylerolson/capstone-backend/services/points"
	"github.com/tylerolson/capstone-backend/services/progress"
)

type PointsSummaryResponse struct {
	TotalPoints   int                        `json:"totalPoints"`
	CurrentStreak int                        `json:"currentStreak"`
	MaxStreak     int                        `json:"maxStreak"`
	Transactions  []*points.PointTransaction `json:"recentTransactions,omitempty"`
}

type LessonPointsResponse struct {
	CourseID      string `json:"courseId"`
	LessonID      string `json:"lessonId"`
	TotalPoints   int    `json:"totalPoints"`
	CurrentStreak int    `json:"currentStreak"`
	MaxStreak     int    `json:"maxStreak"`
}

// GET /api/points/summary
func (s *Server) handleGetPointsSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		userPoints, err := s.PointsService.GetUserTotalPoints(userID)
		if err != nil {
			s.logger.Error("Failed to get user points", "error", err)
			http.Error(w, "Failed to get user points", http.StatusInternalServerError)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 5 // Default limit
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		transactions, err := s.PointsService.GetRecentTransactions(userID, limit)
		if err != nil {
			s.logger.Error("Failed to get recent transactions", "error", err)
			transactions = []*points.PointTransaction{}
		}

		response := PointsSummaryResponse{
			TotalPoints:   userPoints.TotalPoints,
			CurrentStreak: userPoints.CurrentStreak,
			MaxStreak:     userPoints.MaxStreak,
			Transactions:  transactions,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode points summary response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// GET /api/courses/{courseID}/lessons/{lessonID}/points
func (s *Server) handleGetLessonPoints() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")

		if courseID == "" || lessonID == "" {
			http.Error(w, "Course ID and lesson ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		lessonPoints, err := s.PointsService.GetLessonPoints(userID, courseID, lessonID)
		if err != nil {
			s.logger.Error("Failed to get lesson points", "error", err)
			http.Error(w, "Failed to get lesson points", http.StatusInternalServerError)
			return
		}

		response := LessonPointsResponse{
			CourseID:      lessonPoints.CourseID,
			LessonID:      lessonPoints.LessonID,
			TotalPoints:   lessonPoints.TotalPoints,
			CurrentStreak: lessonPoints.CurrentStreak,
			MaxStreak:     lessonPoints.MaxStreak,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode lesson points response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// POST /api/courses/{courseID}/complete
func (s *Server) handleCompleteCourse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		if courseID == "" {
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		_, err := s.CourseService.GetCourseByID(courseID)
		if err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		err = s.ProgressService.UpdateCourseProgress(userID, courseID, progress.StatusCompleted)
		if err != nil {
			s.logger.Error("Failed to update course progress", "error", err)
			http.Error(w, "Failed to update course progress", http.StatusInternalServerError)
			return
		}

		transaction, err := s.PointsService.AwardCourseCompletionBonus(userID, courseID)
		if err != nil {
			s.logger.Error("Failed to award course completion bonus", "error", err)
			http.Error(w, "Failed to award completion bonus", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(transaction); err != nil {
			s.logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// POST /api/courses/{courseID}/lessons/{lessonID}/complete
func (s *Server) handleCompleteLesson() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")

		if courseID == "" || lessonID == "" {
			http.Error(w, "Course ID and lesson ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		err := s.ProgressService.UpdateLessonProgress(userID, courseID, lessonID, progress.StatusCompleted)
		if err != nil {
			s.logger.Error("Failed to update lesson progress", "error", err)
			http.Error(w, "Failed to update lesson progress", http.StatusInternalServerError)
			return
		}

		transaction, err := s.PointsService.AwardLessonCompletionBonus(userID, courseID, lessonID)
		if err != nil {
			s.logger.Error("Failed to award lesson completion bonus", "error", err)
			http.Error(w, "Failed to award completion bonus", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(transaction); err != nil {
			s.logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Enhance the existing exercise attempt handler to award points
// This modifies the existing function in progress_handler.go
func (s *Server) handleExerciseAttemptWithPoints() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExerciseAttemptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.logger.Warn("Error decoding request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")
		exerciseID := r.PathValue("exerciseID")

		s.logger.Debug("Received exercise attempt", "courseID", courseID, "lessonID", lessonID, "exerciseID", exerciseID, "answer", req.Answer)

		if courseID == "" || lessonID == "" || exerciseID == "" {
			http.Error(w, "Course ID, lesson ID, and exercise ID are required", http.StatusBadRequest)
			return
		}

		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		answerJSON, err := json.Marshal(req.Answer)
		if err != nil {
			http.Error(w, "Failed to process answer data", http.StatusInternalServerError)
			return
		}

		attempt := &progress.ExerciseAttempt{
			UserID:     userID,
			CourseID:   courseID,
			LessonID:   lessonID,
			ExerciseID: exerciseID,
			Answer:     string(answerJSON),
		}

		isCorrect, err := s.CourseService.VerifyExerciseAnswer(courseID, lessonID, exerciseID, req.Answer)
		if err != nil {
			http.Error(w, "Failed to verify exercise answer", http.StatusInternalServerError)
			return
		}

		attempt.IsCorrect = isCorrect

		if err := s.ProgressService.RecordExerciseAttempt(attempt); err != nil {
			http.Error(w, "Failed to record exercise attempt", http.StatusInternalServerError)
			return
		}

		var transaction *points.PointTransaction
		if isCorrect {
			transaction, err = s.PointsService.AwardPointsForCorrectAnswer(userID, courseID, lessonID, exerciseID, true)
			if err != nil {
				s.logger.Error("Failed to award points for correct answer", "error", err)
			}
		} else {
			_, err = s.PointsService.AwardPointsForCorrectAnswer(userID, courseID, lessonID, exerciseID, false)
			if err != nil {
				s.logger.Error("Failed to update streak for incorrect answer", "error", err)
			}
		}

		lessonPoints, err := s.PointsService.GetLessonPoints(userID, courseID, lessonID)
		if err != nil {
			s.logger.Error("Failed to get updated lesson points", "error", err)
		}

		response := struct {
			IsCorrect     bool                     `json:"isCorrect"`
			Points        int                      `json:"points,omitempty"`
			Transaction   *points.PointTransaction `json:"transaction,omitempty"`
			CurrentStreak int                      `json:"currentStreak"`
			MaxStreak     int                      `json:"maxStreak"`
		}{
			IsCorrect: isCorrect,
		}

		if transaction != nil {
			response.Points = transaction.Points
			response.Transaction = transaction
		}

		if lessonPoints != nil {
			response.CurrentStreak = lessonPoints.CurrentStreak
			response.MaxStreak = lessonPoints.MaxStreak
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// GET /api/leaderboard
func (s *Server) handleGetLeaderboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		limit := 10 // Default limit
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		rows, err := s.db.Query(`
			SELECT id, username, total_points, profile_picture
			FROM users
			ORDER BY total_points DESC
			LIMIT $1`, limit)
		if err != nil {
			s.logger.Error("Failed to query leaderboard", "error", err)
			http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type LeaderboardEntry struct {
			UserID         int    `json:"userId"`
			Username       string `json:"username"`
			TotalPoints    int    `json:"totalPoints"`
			ProfilePicture string `json:"profilePicture,omitempty"`
			Rank           int    `json:"rank"`
		}

		var leaderboard []LeaderboardEntry
		rank := 1

		for rows.Next() {
			var entry LeaderboardEntry
			var profilePicture sql.NullString

			if err := rows.Scan(&entry.UserID, &entry.Username, &entry.TotalPoints, &profilePicture); err != nil {
				s.logger.Error("Failed to scan leaderboard row", "error", err)
				continue
			}

			if profilePicture.Valid {
				entry.ProfilePicture = profilePicture.String
			}

			entry.Rank = rank
			leaderboard = append(leaderboard, entry)
			rank++
		}

		if err := rows.Err(); err != nil {
			s.logger.Error("Error iterating leaderboard rows", "error", err)
			http.Error(w, "Failed to process leaderboard data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(leaderboard); err != nil {
			s.logger.Error("Failed to encode leaderboard response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
