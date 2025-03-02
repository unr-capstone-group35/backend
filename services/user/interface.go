package user

import (
	"errors"
	"time"
)

var (
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
	ErrNoUser        = errors.New("user does not exist")
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means this won't be included in JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Service interface {
	Create(username, email, password string) (*User, error)
	List() ([]*User, error)
	Get(username string) (*User, error)
	DeleteUser(username string) error
	Authenticate(username, password string) (*User, error)
}
