package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/tylerolson/capstone-backend/services/user"
)

// Request/response types for user management
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

// Request/response types for profile pictures
type ProfilePicRequest struct {
	ProfilePicID string `json:"profilePicId"`
}

type ProfilePicResponse struct {
	ProfilePicID string `json:"profilePicId"`
}

// handleListUsers returns a list of all users (admin functionality)
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

// handleCreateUser creates a new user account
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
		u, err := s.UserService.Create(request.Username, request.Email, request.Password)
		if err != nil {
			if errors.Is(err, user.ErrUsernameTaken) {
				s.logger.Debug("Username already exists", "error", err)
				http.Error(w, "Username already exists", http.StatusConflict)
			} else if errors.Is(err, user.ErrEmailTaken) {
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
			Username: u.Username,
			Email:    u.Email,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Error encoding response", "error", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// handleSignIn authenticates a user and creates a session
func (s *Server) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			s.logger.Debug("Could not decode sign in request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate user and create session
		user, err := s.UserService.Authenticate(request.Username, request.Password)
		if err != nil {
			s.logger.Debug("Sign in authentication failed", "error", err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session, err := s.SessionService.CreateSession(user.ID)
		if err != nil {
			s.logger.Debug("Sign in authentication failed", "error", err)
			http.Error(w, "Could not create session", http.StatusInternalServerError)
			return
		}

		// Send response with session token
		w.Header().Set("Content-Type", "application/json")
		response := SignInResponse{
			Username:  user.Username,
			Email:     user.Email,
			Token:     session.Token,
			ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Error encoding sign in response", "err", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// handleLogout ends a user session
func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := s.GetToken(r.Context())

		if !ok {
			s.logger.Debug("No token provided in logout")
			http.Error(w, "No session token provided", http.StatusBadRequest)
			return
		}

		// Delete the session from database
		if err := s.SessionService.DeleteSession(token); err != nil {
			s.logger.Error("Failed to delete session", "error", err)
			http.Error(w, "Failed to logout", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// === PROFILE PICTURE HANDLERS ===

// handleGetProfilePic returns the user's profile picture ID or serves a custom image
func (s *Server) handleGetProfilePic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get username from context (set by auth middleware)
		username, ok := s.GetUsername(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get profile pic from service
		customPic, profilePicID, err := s.UserService.GetProfilePic(username)
		if err != nil {
			s.logger.Error("Error getting profile pic", "error", err)
			http.Error(w, "Failed to get profile picture", http.StatusInternalServerError)
			return
		}

		// If specifically requesting the image binary and there's a custom pic
		if r.URL.Query().Get("type") == "image" && len(customPic) > 0 && profilePicID == "custom" {
			// Set appropriate content type based on image format
			// For Phase 1, we'll skip advanced image type detection
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
			w.Write(customPic)
			return
		}

		// Otherwise just return the profile pic ID as JSON
		response := ProfilePicResponse{
			ProfilePicID: profilePicID,
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			s.logger.Error("Error marshaling response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}
}

// handleUpdateProfilePic updates the user's profile picture ID
func (s *Server) handleUpdateProfilePic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get username from context (set by auth middleware)
		username, ok := s.GetUsername(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var req ProfilePicRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			s.logger.Error("Error decoding request", "error", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Update profile pic
		err = s.UserService.UpdateProfilePic(username, req.ProfilePicID)
		if err != nil {
			s.logger.Error("Error updating profile pic", "error", err)
			http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
			return
		}

		// Return success response
		response := ProfilePicResponse(req)

		responseJSON, err := json.Marshal(response)
		if err != nil {
			s.logger.Error("Error marshaling response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}
}

// handleUploadProfilePic handles custom image uploads
func (s *Server) handleUploadProfilePic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get username from context (set by auth middleware)
		username, ok := s.GetUsername(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse multipart form (max 2MB)
		err := r.ParseMultipartForm(2 << 20) // 2MB max
		if err != nil {
			s.logger.Error("Error parsing form", "error", err)
			http.Error(w, "Invalid form data or file too large (max 2MB)", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, _, err := r.FormFile("profilePic")
		if err != nil {
			s.logger.Error("Error getting file", "error", err)
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file data
		imageData, err := io.ReadAll(file)
		if err != nil {
			s.logger.Error("Error reading file", "error", err)
			http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
			return
		}

		// Store image
		err = s.UserService.UploadProfilePic(username, imageData)
		if err != nil {
			s.logger.Error("Error storing profile pic", "error", err)
			http.Error(w, "Failed to store profile picture", http.StatusInternalServerError)
			return
		}

		// Return success response
		response := ProfilePicResponse{
			ProfilePicID: "custom",
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			s.logger.Error("Error marshaling response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}
}
