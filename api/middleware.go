package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/tylerolson/capstone-backend/services/user"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	tokenKey    contextKey = "token"
	usernameKey contextKey = "username" // Add username key for profile pic handlers
)

type Middleware func(http.Handler) http.Handler

func (s *Server) DbAuthMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")

			if len(token) == 0 {
				http.Error(w, "Token not provided", http.StatusUnauthorized)
				return
			}

			session, err := s.SessionService.GetSessionByToken(token)
			if err != nil || session == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}

			// Get the username for the user ID
			// We need this for profile picture handlers
			user, err := s.getUsernameByID(session.UserID)
			if err != nil {
				s.logger.Error("Failed to get username for user ID", "userID", session.UserID, "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add user ID and username to request context
			ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
			ctx = context.WithValue(ctx, tokenKey, token)
			ctx = context.WithValue(ctx, usernameKey, user.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper function to get user by ID and return username
func (s *Server) getUsernameByID(userID int) (*user.User, error) {
	// Get all users (this could be optimized with a GetByID method)
	users, err := s.UserService.List()
	if err != nil {
		return nil, err
	}

	// Find the user with the given ID
	for _, u := range users {
		if u.ID == userID {
			return u, nil
		}
	}

	return nil, user.ErrNoUser
}

// GetUserID retrieves user ID from context
func (s *Server) GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

// GetToken retrieves token from context
func (s *Server) GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

// GetUsername retrieves username from context
func (s *Server) GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}
