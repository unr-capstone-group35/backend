// db/db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase() (*Database, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5433/%s?sslmode=disable",
		dbUser, dbPassword, dbName)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

// User represents a user in the database
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // The "-" means this won't be included in JSON
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Session represents a user session in the database
type Session struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// CreateUser creates a new user in the database
func (d *Database) CreateUser(user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return d.DB.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetUserByUsername retrieves a user by their username
func (d *Database) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	err := d.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateSession creates a new session for a user
func (d *Database) CreateSession(session *Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return d.DB.QueryRow(
		query,
		session.UserID,
		session.Token,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

// GetSessionByToken retrieves a session by its token
func (d *Database) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1`

	err := d.DB.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (d *Database) ListUsers() ([]*User, error) {
	query := `
        SELECT id, username, email, created_at, updated_at 
        FROM users 
        ORDER BY created_at DESC`

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (d *Database) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := d.DB.Exec(query, token)
	return err
}
