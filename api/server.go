// api/server.go
package api

import (
	"net/http"

	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/user"
)

type Server struct {
	Mux           *http.ServeMux
	UserService   user.Service
	CourseService course.Service
	Port          string
}

func NewServer(userService user.Service, courseService course.Service, port string) *Server {
	s := &Server{
		UserService:   userService,
		CourseService: courseService,
		Mux:           http.NewServeMux(),
		Port:          port,
	}

	// Register routes
	s.Mux.Handle("GET /api/users", s.handleListUsers())
	s.Mux.Handle("POST /api/users", s.handleCreateUser())
	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())
	s.Mux.Handle("POST /api/logout", s.handleLogout())
	s.Mux.Handle("GET /api/courses", s.handleListCourses())
	s.Mux.Handle("GET /api/courses/{name}", s.handleGetCourse())

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
