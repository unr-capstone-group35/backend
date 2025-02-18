package db

import (
	"database/sql"
	"fmt"
)

func (d *Database) CreateSession(session *Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return d.DB.QueryRow(query, session.UserID, session.Token, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
}

func (d *Database) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1`

	err := d.DB.QueryRow(query, token).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}

		return nil, err
	}

	return session, nil
}

func (d *Database) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := d.DB.Exec(query, token)

	return err
}
