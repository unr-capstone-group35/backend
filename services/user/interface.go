package user

import (
	"errors"
	"time"
)

var (
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
	ErrNoUser        = errors.New("user does not exist")
	ErrInvalidImage  = errors.New("invalid image format or size")
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means this won't be included in JSON
	ProfilePicID string    `json:"profilePicId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Service interface {
	Create(username, email, password string) (*User, error)
	List() ([]*User, error)
	Get(username string) (*User, error)
	DeleteUser(username string) error
	Authenticate(username, password string) (*User, error)

	// New methods for profile pictures
	UpdateProfilePic(username, profilePicID string) error
	UploadProfilePic(username string, imageData []byte) error
	GetProfilePic(username string) ([]byte, string, error)
}
