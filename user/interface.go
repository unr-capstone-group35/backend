// user/interface.go
package user

import "time"

type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Service interface {
	Create(username, password string) (*User, error)
	List() ([]*User, error)
	Get(username string) (*User, error)
	Authenticate(username, password string) (*User, error)
	AuthenticateAndCreateSession(username, password string) (*User, string, time.Time, error)
	DeleteSession(token string) error
}
