package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/mrz1836/postmark"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	db *sql.DB
}

var postmarkClient *postmark.Client

func SetPostmarkClient(client *postmark.Client) {
	postmarkClient = client
}

func NewService(database *sql.DB) Service {
	return &service{db: database}
}

func (s *service) List() ([]*User, error) {
	query := `
        SELECT id, username, email, created_at, updated_at 
        FROM users 
        ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Update the existing Get method to include the profile_pic_id
func (s *service) Get(username string) (*User, error) {
	user := &User{}

	query := `
		SELECT id, username, email, password_hash, profile_pic_id, created_at, updated_at
		FROM users
		WHERE username = $1`

	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.ProfilePicID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoUser
		}

		return nil, err
	}

	return user, nil
}

// Update the existing Create method to set default profile_pic_id
func (s *service) Create(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		ProfilePicID: "default", // Set default profile pic
	}

	query := `
		INSERT INTO users (username, email, password_hash, profile_pic_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err = s.db.QueryRow(query, username, email, string(hashedPassword), "default").Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, ErrUsernameTaken
			case "users_email_key":
				return nil, ErrEmailTaken
			}
		}
		return nil, err
	}

	return user, nil
}

func (s *service) DeleteUser(username string) error {
	query := `DELETE FROM users WHERE username = $1`
	result, err := s.db.Exec(query, username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

func (s *service) Authenticate(username, password string) (*User, error) {
	user, err := s.Get(username)
	if err != nil {
		return nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateProfilePic updates the user's profile picture ID
func (s *service) UpdateProfilePic(username, profilePicID string) error {
	query := `
		UPDATE users
		SET profile_pic_id = $1
		WHERE username = $2`

	result, err := s.db.Exec(query, profilePicID, username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// UploadProfilePic stores a custom profile picture for the user
func (s *service) UploadProfilePic(username string, imageData []byte) error {
	// For Phase 1, we'll implement basic functionality without image validation
	// In a production environment, you'd want to validate the image format and size

	query := `
		UPDATE users
		SET custom_profile_pic = $1, profile_pic_id = 'custom'
		WHERE username = $2`

	result, err := s.db.Exec(query, imageData, username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoUser
	}

	return nil
}

// GetProfilePic retrieves the user's profile picture
func (s *service) GetProfilePic(username string) ([]byte, string, error) {
	query := `
		SELECT custom_profile_pic, profile_pic_id
		FROM users
		WHERE username = $1`

	var customPic []byte
	var profilePicID string

	err := s.db.QueryRow(query, username).Scan(&customPic, &profilePicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrNoUser
		}
		return nil, "", err
	}

	return customPic, profilePicID, nil
}

// RequestPasswordReset creates a password reset token and sends an email
func (s *service) RequestPasswordReset(email string) error {
	var userID int
	var username string

	err := s.db.QueryRow("SELECT id, username FROM users WHERE email = $1", email).Scan(&userID, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	token := make([]byte, 32)
	_, err = rand.Read(token)
	if err != nil {
		return err
	}

	resetToken := base64.URLEncoding.EncodeToString(token)

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = s.db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, resetToken, expiresAt,
	)
	if err != nil {
		return err
	}

	if postmarkClient != nil {
		resetURL := fmt.Sprintf("http://localhost:3000/reset-password/%s", resetToken)

		email := postmark.Email{
			From:          "jasonparmar@unr.edu",
			To:            email,
			Subject:       "DevQuest Password Reset",
			TextBody:      fmt.Sprintf("Hello %s,\n\nWe received a request to reset your password. If you didn't make this request, you can ignore this email.\n\nTo reset your password, please click the link below:\n\n%s\n\nThis link will expire in 24 hours.\n\nThank you,\nDevQuest Team", username, resetURL),
			HTMLBody:      fmt.Sprintf("<p>Hello %s,</p><p>We received a request to reset your password. If you didn't make this request, you can ignore this email.</p><p>To reset your password, please click the link below:</p><p><a href=\"%s\">Reset Password</a></p><p>This link will expire in 24 hours.</p><p>Thank you,<br>DevQuest Team</p>", username, resetURL),
			MessageStream: "outbound",
		}

		_, err = postmarkClient.SendEmail(context.Background(), email)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Password reset requested for user ID %d. Token: %s", userID, resetToken)
	}

	return nil
}

// VerifyResetToken checks if a token is valid and returns the associated user
func (s *service) VerifyResetToken(token string) (*User, error) {
	var userID int
	var expiresAt time.Time

	err := s.db.QueryRow(
		"SELECT user_id, expires_at FROM password_reset_tokens WHERE token = $1",
		token,
	).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired token")
		}
		return nil, err
	}

	if time.Now().After(expiresAt) {
		return nil, errors.New("token has expired")
	}

	var user User
	err = s.db.QueryRow(
		"SELECT id, username, email, profile_pic_id, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePicID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ResetPassword resets a user's password using a valid token
func (s *service) ResetPassword(token string, newPassword string) error {
	user, err := s.VerifyResetToken(token)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		string(hashedPassword), user.ID,
	)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM password_reset_tokens WHERE token = $1", token)
	if err != nil {
		return err
	}

	return nil
}
