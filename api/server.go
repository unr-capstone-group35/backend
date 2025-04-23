package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/services/points"
	"github.com/tylerolson/capstone-backend/services/progress"
	"github.com/tylerolson/capstone-backend/services/session"
	"github.com/tylerolson/capstone-backend/services/user"
)

type Server struct {
	Mux             *http.ServeMux
	UserService     user.Service
	CourseService   course.Service
	ProgressService progress.Service
	SessionService  session.Service
	PointsService   points.Service
	logger          *slog.Logger
	db              *sql.DB
}

func NewServer(userService user.Service, courseService course.Service, progressService progress.Service, sessionService session.Service, pointsService points.Service, database *sql.DB, logger *slog.Logger) *Server {
	s := &Server{
		UserService:     userService,
		CourseService:   courseService,
		ProgressService: progressService,
		SessionService:  sessionService,
		PointsService:   pointsService,
		Mux:             http.NewServeMux(),
		logger:          logger,
		db:              database,
	}

	// Public routes
	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())

	// Protected routes (require authentication)
	dbAuth := s.DbAuthMiddleware()

	// User profile endpoints
	s.Mux.Handle("GET /api/users/profilepic", dbAuth(s.handleGetProfilePic()))
	s.Mux.Handle("PUT /api/users/profilepic", dbAuth(s.handleUpdateProfilePic()))
	s.Mux.Handle("POST /api/users/profilepic/upload", dbAuth(s.handleUploadProfilePic()))

	// Other protected routes
	s.Mux.Handle("POST /api/logout", dbAuth(s.handleLogout()))
	s.Mux.Handle("GET /api/users", dbAuth(s.handleListUsers()))
	s.Mux.Handle("GET /api/courses", dbAuth(s.handleListCourses()))
	s.Mux.Handle("GET /api/courses/{courseID}", dbAuth(s.handleGetCourse()))
	s.Mux.Handle("GET /api/courses/{courseID}/lessons/{lessonID}", dbAuth(s.handleGetLesson()))

	// Progress tracking routes (protected)
	s.Mux.Handle("GET /api/courses/{courseID}/progress", dbAuth(s.handleGetCourseProgress()))
	s.Mux.Handle("GET /api/courses/{courseID}/lessons/{lessonID}/progress", dbAuth(s.handleGetLessonProgress()))
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/progress", dbAuth(s.handleUpdateLessonProgress()))

	// Exercise attempt handler with points included
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/exercises/{exerciseID}/attempt", dbAuth(s.handleExerciseAttempt()))
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/exercises/{exerciseID}/points", dbAuth(s.handleExerciseAttemptPoints()))

	// Points and gamification routes
	s.Mux.Handle("GET /api/points/summary", dbAuth(s.handleGetPointsSummary()))
	s.Mux.Handle("GET /api/courses/{courseID}/lessons/{lessonID}/points", dbAuth(s.handleGetLessonPoints()))
	s.Mux.Handle("POST /api/courses/{courseID}/complete", dbAuth(s.handleCompleteCourse()))
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/complete", dbAuth(s.handleCompleteLesson()))
	s.Mux.Handle("GET /api/leaderboard", dbAuth(s.handleGetLeaderboard()))
	s.Mux.Handle("GET /api/stats/daily-streak", dbAuth(s.handleGetDailyStreak()))
	s.Mux.Handle("GET /api/stats/accuracy", dbAuth(s.handleGetAccuracyStats()))

	// Make sure the profile pictures directory exists
	if err := EnsureProfilePicDirectory(); err != nil {
		logger.Error("Failed to create profile pictures directory", "error", err)
		os.Exit(1)
	}

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle OPTIONS requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle static file serving for profile pictures
	if strings.HasPrefix(r.URL.Path, "/static/profile-pics/") {
		// Security check to prevent directory traversal
		cleanPath := filepath.Clean(r.URL.Path)
		if !strings.HasPrefix(cleanPath, "/static/profile-pics/") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		filePath := "." + cleanPath // Convert URL path to file path
		http.ServeFile(w, r, filePath)
		return
	}

	s.Mux.ServeHTTP(w, r)
}
