package api

import (
	"encoding/json"
	"net/http"

	"github.com/tylerolson/capstone-backend/course"
)

type CourseInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	LessonAmount int    `json:"lessonAmount"`
}

type ListCoursesResponse []CourseInfo

// GET /api/courses
func (s *Server) handleListCourses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Got list course request")

		courses, err := s.CourseService.ListCourses()
		if err != nil {
			s.logger.Warn("Failed to list courses", "err", err)
			http.Error(w, "Failed to list courses", http.StatusInternalServerError)
			return
		}

		response := make([]CourseInfo, 0)

		for _, course := range courses {
			response = append(response, CourseInfo{
				ID:           course.ID,
				Name:         course.Name,
				Description:  course.Description,
				LessonAmount: len(course.Lessons),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode response", "err", err)
			http.Error(w, "Failed to encode response:", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleGetCourse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		s.logger.Debug("Received request for course", "courseID", courseID)

		if courseID == "" {
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		course, err := s.CourseService.GetCourseByID(courseID)
		if err != nil {
			s.logger.Warn("Error fetching course by courseID", "courseID", courseID, "error", err)
			http.Error(w, "Failed to get course", http.StatusNotFound)
			return
		}

		s.logger.Debug("Course found", "course", course.Name)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(course); err != nil {
			s.logger.Error("Error encoding response for course", "courseID", courseID, "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		s.logger.Debug("Response successfully sent for course", "courseID", courseID)
	}
}

func (s *Server) handleGetLesson() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := r.PathValue("courseID")
		lessonID := r.PathValue("lessonID")

		s.logger.Debug("Recieved get lesson request", "courseID", courseID, "lessonID", lessonID)
		if courseID == "" || lessonID == "" {
			http.Error(w, "Course ID and lesson ID are required", http.StatusBadRequest)
			return
		}

		lesson, err := s.CourseService.GetLessonByID(courseID, lessonID)
		if err != nil {
			s.logger.Debug("Lesson not found", "courseID", courseID, "lessonID", lessonID)
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
