package api

import (
	"os"
)

// EnsureProfilePicDirectory creates the directory for storing profile pictures if it doesn't exist
func EnsureProfilePicDirectory() error {
	dir := "./static/profile-pics"
	return os.MkdirAll(dir, 0755)
}
