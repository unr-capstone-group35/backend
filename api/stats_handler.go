package api

import (
	"encoding/json"
	"net/http"
)

// Stats response types
type DailyStreakResponse struct {
	CurrentStreak   int    `json:"currentStreak"`
	MaxStreak       int    `json:"maxStreak"`
	LastLoginDate   string `json:"lastLoginDate,omitempty"`
	NextMilestone   int    `json:"nextMilestone,omitempty"`
	DaysToMilestone int    `json:"daysToMilestone,omitempty"`
}

type AccuracyResponse struct {
	TotalAttempts   int     `json:"totalAttempts"`
	CorrectAttempts int     `json:"correctAttempts"`
	AccuracyRate    float64 `json:"accuracyRate"` // Percentage (0-100)
}

// GET /api/stats/daily-streak
func (s *Server) handleGetDailyStreak() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		streakInfo, err := s.PointsService.GetDailyStreak(userID)
		if err != nil {
			s.logger.Error("Failed to get daily streak", "error", err)
			http.Error(w, "Failed to get daily streak information", http.StatusInternalServerError)
			return
		}

		response := DailyStreakResponse{
			CurrentStreak:   streakInfo.CurrentStreak,
			MaxStreak:       streakInfo.MaxStreak,
			NextMilestone:   streakInfo.NextMilestone,
			DaysToMilestone: streakInfo.DaysToMilestone,
		}

		if !streakInfo.LastLoginDate.IsZero() {
			response.LastLoginDate = streakInfo.LastLoginDate.Format("2006-01-02")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode daily streak response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// GET /api/stats/accuracy
func (s *Server) handleGetAccuracyStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := s.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		stats, err := s.PointsService.GetAccuracyStats(userID)
		if err != nil {
			s.logger.Error("Failed to get accuracy stats", "error", err)
			http.Error(w, "Failed to get accuracy statistics", http.StatusInternalServerError)
			return
		}

		response := AccuracyResponse{
			TotalAttempts:   stats.TotalAttempts,
			CorrectAttempts: stats.CorrectAttempts,
			AccuracyRate:    stats.AccuracyRate,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode accuracy response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
