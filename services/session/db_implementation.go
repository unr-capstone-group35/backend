package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

type service struct {
	db *sql.DB
}

func NewService(database *sql.DB) Service {
	return &service{db: database}
}

func (s *service) CreateSession(userID int) (*Session, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	expiresAt := time.Now().Add(TokenLifetime)

	session := &Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := s.db.QueryRow(query, userID, token, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *service) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1`

	err := s.db.QueryRow(query, token).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}

		return nil, err
	}

	return session, nil
}

func (s *service) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := s.db.Exec(query, token)

	return err
}
