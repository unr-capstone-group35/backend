// user/errors.go
package user

import "errors"

var (
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
)
