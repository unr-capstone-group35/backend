package api

import (
	"net/http"

	"github.com/tylerolson/capstone-backend/user"
)

type Server struct {
	Mux         *http.ServeMux
	UserService user.Service
}

func NewServer(userService user.Service) *Server {
	s := Server{
		UserService: userService,
		Mux:         http.NewServeMux(),
	}

	s.Mux.Handle("GET /users", s.handleListUsers())
	s.Mux.Handle("POST /users", s.handleCreateUser())

	return &s
}
