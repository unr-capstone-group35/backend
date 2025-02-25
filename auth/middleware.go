package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/tylerolson/capstone-backend/db"
)

type Middleware func(http.Handler) http.Handler

func DbAuthMiddleware(db *db.Database) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")

			if len(token) == 0 {
				http.Error(w, "Token not provided", http.StatusUnauthorized)
				return
			}

			session, err := db.GetSessionByToken(token)
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
