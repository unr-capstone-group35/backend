package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/services/user"

	"log/slog"
)

func TestRegisterLoginEndpoint(t *testing.T) {
	os.Setenv("POSTGRES_USER", "dbuser")
	os.Setenv("POSTGRES_PASSWORD", "dbpassword")
	os.Setenv("POSTGRES_DB", "capstone_db")

	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		}),
	)

	//this needs to be a test database, not sure how yet
	database, err := db.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	userService := user.NewService(database)
	coursesStore := course.NewJSONStore("../data", database)

	server := api.NewServer(
		userService,
		coursesStore,
		database,
		logger,
	)

	testEmail := "user@example.com"
	testUsername := "User"
	testPassword := "password123"

	err = database.DeleteUserByUsername(testUsername)
	if err != nil && err.Error() != "user does not exist" {
		t.Fatalf("Could not delete test user: %v", err)
	}

	t.Run("Register Test User", func(t *testing.T) {
		createUserBody, err := json.Marshal(api.CreateUserRequest{
			Email:    testEmail,
			Username: testUsername,
			Password: testPassword,
		})

		if err != nil {
			t.Fatalf("Failed to marshal create user request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(createUserBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		server.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated && status != http.StatusOK {
			t.Fatalf("Failed to create test user: got status %v, response: %s", status, rr.Body.String())
		}

		t.Logf("Successfully created test user with email: %s", testEmail)
	})

	t.Run("Login with Created User", func(t *testing.T) {
		loginBody, err := json.Marshal(api.SignInRequest{
			Username: testUsername,
			Password: testPassword,
		})

		if err != nil {
			t.Fatalf("Failed to marshal login request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		server.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Login failed: got status %v, want %v", status, http.StatusOK)
			t.Logf("Response body: %s", rr.Body.String())
			return
		}

		// Parse the response to check for token
		var response api.SignInResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify token exists in response
		token := response.Token
		if len(token) == 0 {
			t.Error("No valid token in login response")
			t.Logf("Response body: %s", rr.Body.String())
			return
		}

		t.Logf("Successfully logged in with created user")
	})

	t.Run("Login with Incorrect Password", func(t *testing.T) {
		loginBody, err := json.Marshal(api.SignInRequest{
			Username: testUsername,
			Password: "wrong_password",
		})
		if err != nil {
			t.Fatalf("Failed to marshal login request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		server.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Invalid login test failed: got status %v, want %v", status, http.StatusUnauthorized)
			t.Logf("Response body: %s", rr.Body.String())
		}

		t.Log("Successfully rejected login with incorrect password")
	})
}
