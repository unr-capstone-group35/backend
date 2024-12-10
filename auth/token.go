// auth/token.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/tylerolson/capstone-backend/db"
)

// Define custom type for context keys to avoid collisions
type contextKey string

const (
	TokenLength    = 32 // 256 bits
	TokenLifetime  = 24 * time.Hour
	TokenHeaderKey = "X-Session-Token"

	// Define context key for user ID
	userIDKey contextKey = "userID"
)

// GenerateToken creates a cryptographically secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateSession generates a new session token and expiration time
func CreateSession() (token string, expiresAt time.Time, err error) {
	token, err = GenerateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(TokenLifetime)
	return token, expiresAt, nil
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

// Middleware represents the interface for auth middleware
type Middleware interface {
	RequireAuth(next http.HandlerFunc) http.HandlerFunc
}

type middleware struct {
	db *db.Database
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(db *db.Database) Middleware {
	return &middleware{db: db}
}

// RequireAuth wraps handlers requiring authentication
func (m *middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(TokenHeaderKey)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := m.db.GetSessionByToken(token)
		if err != nil || session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if session is expired - no need to parse since it's already time.Time
		if time.Now().After(session.ExpiresAt) {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
