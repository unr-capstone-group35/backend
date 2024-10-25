package user

type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Service interface {
	Get(username string) (*User, bool)
	Create(username string, password string) bool
}
