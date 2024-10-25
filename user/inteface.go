package user

type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Service interface {
	List() ([]*User, error)
	Get(username string) (*User, error)
	Create(username string, password string) (*User, error)
}
