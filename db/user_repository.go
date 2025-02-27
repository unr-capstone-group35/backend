package db

import (
	"database/sql"
	"fmt"
)

func (d *Database) CreateUser(user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return d.DB.QueryRow(query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (d *Database) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	err := d.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user does not exist")
		}

		return nil, err
	}

	return user, nil
}

func (d *Database) ListUsers() ([]*User, error) {
	query := `
        SELECT id, username, email, created_at, updated_at 
        FROM users 
        ORDER BY created_at DESC`

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	//users := make([]*User, 0)
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

func (d *Database) DeleteUserByUsername(username string) error {
	query := `DELETE FROM users WHERE username = $1`
	result, err := d.DB.Exec(query, username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user does not exist")
	}

	return nil
}
