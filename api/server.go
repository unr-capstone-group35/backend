package api

import (
	"log/slog"
	"net/http"

	"github.com/tylerolson/capstone-backend/auth"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/user"
)

type Server struct {
	Mux           *http.ServeMux
	UserService   user.Service
	CourseService course.Service
	DB            *db.Database
	logger        *slog.Logger
}

func NewServer(userService user.Service, courseService course.Service, database *db.Database, logger *slog.Logger) *Server {
	s := &Server{
		UserService:   userService,
		CourseService: courseService,
		DB:            database,
		Mux:           http.NewServeMux(),
		logger:        logger,
	}

	// Public routes
	s.Mux.Handle("POST /api/users", s.handleCreateUser())
	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())
	s.Mux.Handle("POST /api/logout", s.handleLogout())

	// Protected routes (require authentication)
	dbAuth := auth.DbAuthMiddleware(s.DB)
	s.Mux.Handle("GET /api/users", dbAuth(s.handleListUsers()))
	s.Mux.Handle("GET /api/courses", dbAuth(s.handleListCourses()))
	s.Mux.Handle("GET /api/courses/{name}", dbAuth(s.handleGetCourse()))
	s.Mux.Handle("GET /api/courses/{name}/lessons/{lessonId}", dbAuth(s.handleGetLesson()))

	// Progress tracking routes (protected)
	s.Mux.Handle("GET /api/courses/{name}/progress", dbAuth(s.handleGetCourseProgress()))
	s.Mux.Handle("GET /api/courses/{name}/lessons/{lessonId}/progress", dbAuth(s.handleGetLessonProgress()))
	s.Mux.Handle("POST /api/courses/{name}/lessons/{lessonId}/progress", dbAuth(s.handleUpdateLessonProgress()))
	s.Mux.Handle("POST /api/courses/{name}/lessons/{lessonId}/exercises/{exerciseId}/attempt", dbAuth(s.handleExerciseAttempt()))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Serve actual request
	s.Mux.ServeHTTP(w, r)
}
