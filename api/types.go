// api/types.go
package api

type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignInResponse struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}
