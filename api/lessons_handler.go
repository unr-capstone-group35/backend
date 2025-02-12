package api

import (
	"encoding/json"
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
		s.logger.Debug("Received request for course", "name", name)

		if name == "" {
			http.Error(w, "Course name is required", http.StatusBadRequest)
			return
		}

		course, err := s.CourseService.GetCourseByName(name)
		if err != nil {
			s.logger.Warn("Error fetching course by name", "name", name, "error", err)
			http.Error(w, "Failed to get course", http.StatusNotFound)
			return
		}

		s.logger.Debug("Course found", "course", course)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(course); err != nil {
			s.logger.Error("Error encoding response for course", "name", name, "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		s.logger.Debug("Response successfully sent for course", "name", name)
	}
}

func (s *Server) handleGetLesson() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseName := r.PathValue("name")
		lessonId := r.PathValue("lessonId")

		s.logger.Debug("Recieved get lesson request", "courseName", courseName, "lessonId", lessonId)
		if courseName == "" || lessonId == "" {
			http.Error(w, "Course name and lesson ID are required", http.StatusBadRequest)
			return
		}

		lesson, err := s.CourseService.GetLessonByID(courseName, lessonId)
		if err != nil {
			s.logger.Debug("Lesson not found", "courseName", courseName, "lessonId", lessonId)
			http.Error(w, "Lesson not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(lesson); err != nil {
			s.logger.Error("Failed to encode get lesson response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
