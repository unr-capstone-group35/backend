package api

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	tokenKey  contextKey = "token"
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

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
			ctx = context.WithValue(ctx, tokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves user ID from context
func (s *Server) GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

func (s *Server) GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}
