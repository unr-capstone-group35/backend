// user/service.go
package user

import (
	"github.com/lib/pq"
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

func (s *service) Create(username, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	dbUser := &db.User{
		Username:     username,
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
