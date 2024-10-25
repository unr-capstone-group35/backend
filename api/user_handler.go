package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateUserRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (s *Server) handleListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.UserService.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(users); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("Returned list of users")
	}
}

func (s *Server) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request CreateUserRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err := s.UserService.Get(request.Username); err == nil {
			w.WriteHeader(http.StatusConflict)
			return
		}

		user, err := s.UserService.Create(request.Username, request.Password)
		if err != nil {
			fmt.Println("failed to create user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(user); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("created user")
	}
}
