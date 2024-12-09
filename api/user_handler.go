// api/user_handler.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *Server) handleListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.UserService.List()
		if err != nil {
			fmt.Printf("Error listing users: %v\n", err)
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
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if request.Username == "" || request.Password == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create user
		user, err := s.UserService.Create(request.Username, request.Username, request.Password)
		if err != nil {
			// Check for specific errors
			switch err.Error() {
			case "username already exists":
				http.Error(w, "Username already exists", http.StatusConflict)
			case "email already exists":
				http.Error(w, "Email already exists", http.StatusConflict)
			default:
				fmt.Printf("Error creating user: %v\n", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
			}
			return
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		response := UserResponse{
			Username: user.Username,
			Email:    request.Email,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Printf("Error encoding response: %v\n", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate user and create session
		user, token, expiresAt, err := s.UserService.AuthenticateAndCreateSession(request.Username, request.Password)
		if err != nil {
			fmt.Printf("Authentication failed: %v\n", err)
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
			fmt.Printf("Error encoding response: %v\n", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" {
			http.Error(w, "No session token provided", http.StatusBadRequest)
			return
		}

		// Delete the session from database
		err := s.UserService.DeleteSession(token)
		if err != nil {
			http.Error(w, "Failed to logout", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
