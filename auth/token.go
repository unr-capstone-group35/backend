package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Custom type for context keys to avoid collisions
type contextKey string

const (
	TokenLength   = 32
	TokenLifetime = 24 * time.Hour

	userIDKey contextKey = "userID"
	tokenKey  contextKey = "token"
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

func GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}
