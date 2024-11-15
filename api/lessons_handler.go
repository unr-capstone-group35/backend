package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NameRequest struct {
	Name string `json:"name,omitempty"`
}

func (s *Server) handleListCourses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paths, err := s.CourseService.ListCourseNames()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(paths); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("Returned list of paths")
	}
}

func (s *Server) handleGetCourse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("id")

		path, err := s.CourseService.GetCourseByName(name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(path); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
