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
}

func NewServer(userService user.Service, courseService course.Service) *Server {
	s := &Server{
		UserService:   userService,
		CourseService: courseService,
		Mux:           http.NewServeMux(),
	}

	// Register routes
	s.routes()

	return s
}

func (s *Server) routes() {
	s.Mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case "GET":
			s.handleListUsers()(w, r)
		case "POST":
			s.handleCreateUser()(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.Mux.HandleFunc("/api/courses", s.handleListCourses())
	s.Mux.HandleFunc("/api/courses/", s.handleGetCourse())

	s.Mux.HandleFunc("/api/signin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			s.handleSignIn()(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	s.Mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			s.handleLogout()(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}
