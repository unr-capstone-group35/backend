package api

import (
	"log/slog"
	"net/http"

	"github.com/tylerolson/capstone-backend/course"
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
	logger          *slog.Logger
}

func NewServer(userService user.Service, courseService course.Service, progressService progress.Service, sessionService session.Service, logger *slog.Logger) *Server {
	s := &Server{
		UserService:     userService,
		CourseService:   courseService,
		ProgressService: progressService,
		SessionService:  sessionService,
		Mux:             http.NewServeMux(),
		logger:          logger,
	}

	// Public routes
	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())

	// Protected routes (require authentication)
	dbAuth := s.DbAuthMiddleware()
	s.Mux.Handle("POST /api/logout", dbAuth(s.handleLogout()))
	s.Mux.Handle("GET /api/users", dbAuth(s.handleListUsers()))
	s.Mux.Handle("GET /api/courses", dbAuth(s.handleListCourses()))
	s.Mux.Handle("GET /api/courses/{courseID}", dbAuth(s.handleGetCourse()))
	s.Mux.Handle("GET /api/courses/{courseID}/lessons/{lessonID}", dbAuth(s.handleGetLesson()))

	// Progress tracking routes (protected)
	s.Mux.Handle("GET /api/courses/{courseID}/progress", dbAuth(s.handleGetCourseProgress()))
	s.Mux.Handle("GET /api/courses/{courseID}/lessons/{lessonID}/progress", dbAuth(s.handleGetLessonProgress()))
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/progress", dbAuth(s.handleUpdateLessonProgress()))
	s.Mux.Handle("POST /api/courses/{courseID}/lessons/{lessonID}/exercises/{exerciseID}/attempt", dbAuth(s.handleExerciseAttempt()))

	// Profile picture endpoints (protected)
	s.Mux.Handle("GET /api/users/profilepic", dbAuth(s.handleGetProfilePic()))
	s.Mux.Handle("PUT /api/users/profilepic", dbAuth(s.handleUpdateProfilePic()))
	s.Mux.Handle("POST /api/users/profilepic/upload", dbAuth(s.handleUploadProfilePic()))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// need this or nuxt screams???
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Serve actual request
	s.Mux.ServeHTTP(w, r)
}
