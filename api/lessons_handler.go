// api/lessons_handler.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleListCourses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paths, err := s.CourseService.ListCourseNames()
		if err != nil {
			http.Error(w, "Failed to list courses", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(paths); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleGetCourse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "Course name is required", http.StatusBadRequest)
			return
		}

		course, err := s.CourseService.GetCourseByName(name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get course: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(course); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleGetLesson() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")

		if courseName == "" || lessonId == "" {
			http.Error(w, "Course name and lesson ID are required", http.StatusBadRequest)
			return
		}

		lesson, err := s.CourseService.GetLessonByID(courseName, lessonId)
		if err != nil {
			http.Error(w, "Lesson not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(lesson); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
