package user

import (
	"database/sql"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	db *sql.DB
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

func (s *service) Create(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err = s.db.QueryRow(query, username, email, string(hashedPassword)).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
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

func (s *service) Get(username string) (*User, error) {
	user := &User{}

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoUser
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
