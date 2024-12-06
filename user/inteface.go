// user/interface.go
package user

type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Service interface {
	List() ([]*User, error)
	Get(username string) (*User, error)
	Create(username string, password string) (*User, error)
	Authenticate(username, password string) (*User, error)
}
