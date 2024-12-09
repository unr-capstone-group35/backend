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

	s.Mux.Handle("GET /api/users", s.handleListUsers())

	s.Mux.Handle("POST /api/signin", s.handleSignIn())
	s.Mux.Handle("POST /api/register", s.handleCreateUser())
	s.Mux.Handle("POST /api/logout", s.handleLogout())

	s.Mux.Handle("GET /api/courses", s.handleListCourses())
	s.Mux.Handle("GET /api/courses/{name}", s.handleGetCourse())

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}
