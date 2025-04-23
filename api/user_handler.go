package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	vision "cloud.google.com/go/vision/apiv1"

	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb" // Import the correct protobuf package path

	"github.com/tylerolson/capstone-backend/services/points"
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

		// Update daily streak
		transaction, err := s.PointsService.UpdateDailyStreak(user.ID)
		if err != nil {
			s.logger.Error("Failed to update daily streak", "error", err)
			// Don't fail the whole sign-in process if streak update fails
		}

		// Prepare response
		response := SignInResponse{
			Username:  user.Username,
			Email:     user.Email,
			Token:     session.Token,
			ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		}

		// If streak update resulted in points, include that info
		if transaction != nil {
			response := struct {
				SignInResponse
				StreakTransaction *points.PointTransaction `json:"streakTransaction,omitempty"`
			}{
				SignInResponse:    response,
				StreakTransaction: transaction,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.logger.Error("Error encoding sign in response", "err", err)
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
				return
			}
			return
		}

		// Regular response without streak transaction
		w.Header().Set("Content-Type", "application/json")
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

		// Detect safe search results
		safeSearchResults, err := s.detectSafeSearch(imageData)
		if err != nil {
			s.logger.Error("Error detecting safe search", "error", err)
			http.Error(w, "Failed to detect safe search", http.StatusInternalServerError)
			return
		}

		// Check if the results violate our tolerance, possible or higher should be flagged
		// Since we are an education app for all ages, we should have a low tolerance
		violation := safeSearchThreshhold(safeSearchResults, 3)
		if violation {
			s.logger.Warn("Found offensive content")
			http.Error(w, "Found offensive content in profile picture.", http.StatusForbidden)
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

func (s *Server) detectSafeSearch(imageData []byte) (*visionpb.SafeSearchAnnotation, error) {
	ctx := context.Background()

	// Create a new Vision client
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	image := &visionpb.Image{Content: imageData}

	// Create a feature request for Safe Search detection using the visionpb type
	features := []*visionpb.Feature{
		{Type: visionpb.Feature_SAFE_SEARCH_DETECTION},
	}

	// Create the AnnotateImageRequest using the visionpb type
	req := &visionpb.AnnotateImageRequest{
		Image:    image,
		Features: features,
	}

	resp, err := client.AnnotateImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to annotate image: %w", err)
	}
	if resp.SafeSearchAnnotation != nil {
		props := resp.SafeSearchAnnotation
		s.logger.Debug("Safe Search properties:", "Adult:", props.Adult, "Medical:", props.Medical, "Spoof:", props.Spoof, "Violence:", props.Violence, "Racy:", props.Racy)
		return props, nil
	} else {
		s.logger.Debug("No Safe Search annotation found.")
	}

	return nil, nil
}

// safeSearchThreshhold() takes in a safeSearchThreshhold and check if any of its resuls violate
// the threshhold. If they do then return true.
func safeSearchThreshhold(a *visionpb.SafeSearchAnnotation, threshhold visionpb.Likelihood) bool {
	if a.Adult >= threshhold {
		return true
	} else if a.Medical >= threshhold {
		return true
	} else if a.Racy >= threshhold {
		return true
	} else if a.Spoof >= threshhold {
		return true
	} else if a.Violence >= threshhold {
		return true
	}

	return false
}
