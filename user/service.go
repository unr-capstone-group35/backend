// user/service.go
package user

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/tylerolson/capstone-backend/auth"
	"github.com/tylerolson/capstone-backend/db"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	db *db.Database
}

func NewService(database *db.Database) Service {
	return &service{db: database}
}

func (s *service) List() ([]*User, error) {
	dbUsers, err := s.db.ListUsers()
	if err != nil {
		return nil, err
	}

	users := make([]*User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = &User{
			Username: dbUser.Username,
			Email:    dbUser.Email,
		}
	}
	return users, nil
}

func (s *service) Create(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	dbUser := &db.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.db.CreateUser(dbUser)
	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, ErrUsernameTaken
			case "users_email_key":
				return nil, ErrEmailTaken
			}
		}
		return nil, err
	}

	return &User{
		Username: dbUser.Username,
		Email:    dbUser.Email,
	}, nil
}

func (s *service) Get(username string) (*User, error) {
	dbUser, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return &User{
		Username: dbUser.Username,
		Email:    dbUser.Email,
	}, nil
}

func (s *service) Authenticate(username, password string) (*User, error) {
	dbUser, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return &User{
		Username: dbUser.Username,
		Email:    dbUser.Email,
	}, nil
}

func (s *service) AuthenticateAndCreateSession(username, password string) (*User, string, time.Time, error) {
	// First authenticate the user
	dbUser, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	fmt.Println(username)
	fmt.Println(dbUser)
	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Generate session token
	token, expiresAt, err := auth.CreateSession()
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("error creating session: %v", err)
	}

	// Store session in database
	session := &db.Session{
		UserID:    dbUser.ID,
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}

	if err := s.db.CreateSession(session); err != nil {
		return nil, "", time.Time{}, fmt.Errorf("error storing session: %v", err)
	}

	return &User{
		Username: dbUser.Username,
		Email:    dbUser.Email,
	}, token, expiresAt, nil
}

func (s *service) DeleteSession(token string) error {
	return s.db.DeleteSession(token)
}
