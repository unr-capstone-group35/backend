// api/server.go
package api

import (
	"net/http"

	"github.com/tylerolson/capstone-backend/auth"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/user"
)

type Server struct {
	Mux            *http.ServeMux
	UserService    user.Service
	CourseService  course.Service
	AuthMiddleware auth.Middleware
	DB             *db.Database
	Port           string
}

func NewServer(userService user.Service, courseService course.Service, authMiddleware auth.Middleware, database *db.Database, port string) *Server {
	s := &Server{
		UserService:    userService,
		CourseService:  courseService,
		AuthMiddleware: authMiddleware,
		DB:             database,
		Mux:            http.NewServeMux(),
		Port:           port,
	}

	// Public routes
	s.Mux.Handle("POST /api/users", s.handleCreateUser())
	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())
	s.Mux.Handle("POST /api/logout", s.handleLogout())

	// Protected routes (require authentication)
	s.Mux.Handle("GET /api/users", s.AuthMiddleware.RequireAuth(s.handleListUsers()))
	s.Mux.Handle("GET /api/courses", s.AuthMiddleware.RequireAuth(s.handleListCourses()))
	s.Mux.Handle("GET /api/courses/{name}", s.AuthMiddleware.RequireAuth(s.handleGetCourse()))
	s.Mux.Handle("GET /api/courses/{name}/lessons/{lessonId}", s.AuthMiddleware.RequireAuth(s.handleGetLesson()))

	// Progress tracking routes (protected)
	s.Mux.Handle("GET /api/courses/{name}/progress",
		s.AuthMiddleware.RequireAuth(s.handleGetCourseProgress()))
	s.Mux.Handle("GET /api/courses/{name}/lessons/{lessonId}/progress",
		s.AuthMiddleware.RequireAuth(s.handleGetLessonProgress()))
	s.Mux.Handle("POST /api/courses/{name}/lessons/{lessonId}/progress",
		s.AuthMiddleware.RequireAuth(s.handleUpdateLessonProgress()))
	s.Mux.Handle("POST /api/courses/{name}/lessons/{lessonId}/exercises/{exerciseId}/attempt",
		s.AuthMiddleware.RequireAuth(s.handleExerciseAttempt()))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Serve the actual request
	s.Mux.ServeHTTP(w, r)
}
