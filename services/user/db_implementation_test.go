package user_test

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/lib/pq"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/services/user"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := db.NewDatabase("dbuser", "dbpassword", "capstone_db", "", "")

	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	service := user.NewService(db)

	username := "testuser"
	email := "test@example.com"
	password := "password123"

	createdUser, err := service.Create(username, email, password)
	if err != nil {
		if errors.Is(err, user.ErrUsernameTaken) || errors.Is(err, user.ErrEmailTaken) {
			t.Skip(err)
		} else {
			t.Fatal(err)
		}
	}

	if createdUser.Username != username || createdUser.Email != email {
		t.Errorf("unexpected user data: %+v", createdUser)
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	service := user.NewService(db)

	username := "testuser"

	retrievedUser, err := service.Get(username)
	if err != nil {
		t.Fatal(err)
	}

	if retrievedUser.Username != username {
		t.Errorf("unexpected user data: %+v", retrievedUser)
	}
}

func TestAuthenticateUser(t *testing.T) {
	db := setupTestDB(t)
	service := user.NewService(db)

	username := "testuser"
	password := "password123"

	authUser, err := service.Authenticate(username, password)
	if err != nil {
		t.Fatal(err)
	}

	if authUser.Username != username {
		t.Errorf("unexpected authenticated user data: %+v", authUser)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	service := user.NewService(db)

	username := "testuser"

	err := service.DeleteUser(username)

	if err != nil {
		if errors.Is(err, user.ErrNoUser) {
			t.Skip(err)
		} else {
			t.Fatal(err)
		}

	}

	_, err = service.Get(username)
	if !errors.Is(err, user.ErrNoUser) {
		t.Errorf("expected ErrNoUser, got %v", err)
	}
}
