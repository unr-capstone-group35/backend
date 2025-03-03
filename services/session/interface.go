package session

import (
	"time"
)

const (
	TokenLength   = 32
	TokenLifetime = 24 * time.Hour
)

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type Service interface {
	CreateSession(userID int) (*Session, error)
	GetSessionByToken(token string) (*Session, error)
	DeleteSession(token string) error
}
