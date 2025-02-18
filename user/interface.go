package user

import (
	"errors"
	"time"
)

var (
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
)

type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Service interface {
	Create(username, email, password string) (*User, error)
	List() ([]*User, error)
	Get(username string) (*User, error)
	Authenticate(username, password string) (*User, error)
	AuthenticateAndCreateSession(username, password string) (*User, string, time.Time, error)
	DeleteSession(token string) error
}
