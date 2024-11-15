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
	s := Server{
		UserService:   userService,
		CourseService: courseService,
		Mux:           http.NewServeMux(),
	}

	s.Mux.Handle("GET /users", s.handleListUsers())
	s.Mux.Handle("POST /users", s.handleCreateUser())

	s.Mux.Handle("GET /courses", s.handleListCourses())
	s.Mux.Handle("GET /courses/{name}", s.handleGetCourse())

	return &s
}
