package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	User "github.com/tylerolson/capstone-backend/user"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
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

func (s *Server) handleListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.UserService.List()
		if err != nil {
			s.logger.Error("Error listing users", "error", err)
			http.Error(w, "Failed to list users", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func (s *Server) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			s.logger.Debug("Could not decode create user request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		s.logger.Debug("Got create user request", "request", request)

		// Validate required fields
		if request.Username == "" || request.Password == "" || request.Email == "" {
			s.logger.Debug("Missing required user fields")
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create unique user with username and email
		user, err := s.UserService.Create(request.Username, request.Email, request.Password)
		if err != nil {
			if errors.Is(err, User.ErrUsernameTaken) {
				s.logger.Debug("Username already exists", "error", err)
				http.Error(w, "Username already exists", http.StatusConflict)
			} else if errors.Is(err, User.ErrEmailTaken) {
				s.logger.Debug("Email already exists", "error", err)
				http.Error(w, "Email already exists", http.StatusConflict)
			} else { // unexpected error, we should log it
				s.logger.Error("Error creating user", "error", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
			}

			return
		}

		s.logger.Debug("Created user", "username", request.Username, "email", request.Email, "password", request.Password)
		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := CreateUserResponse{
			Username: user.Username,
			Email:    user.Email,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Error encoding response", "error", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			s.logger.Debug("Could not decode sign in request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate user and create session
		user, token, expiresAt, err := s.UserService.AuthenticateAndCreateSession(request.Username, request.Password)
		if err != nil {
			s.logger.Debug("Sign in authentication failed", "error", err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Send response with session token
		w.Header().Set("Content-Type", "application/json")
		response := SignInResponse{
			Username:  user.Username,
			Email:     user.Email,
			Token:     token,
			ExpiresAt: expiresAt.Format(time.RFC3339),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Error encoding sign in response", "err", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if len(token) == 0 {
			s.logger.Debug("No token provided in logout")
			http.Error(w, "No session token provided", http.StatusBadRequest)
			return
		}

		// Delete the session from database
		if err := s.UserService.DeleteSession(token); err != nil {
			s.logger.Error("Failed to delete session", "error", err)
			http.Error(w, "Failed to logout", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
