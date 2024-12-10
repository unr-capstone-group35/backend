// course/interface.go
package course

import "github.com/tylerolson/capstone-backend/db"

// Course represents a complete course with its lessons
type Course struct {
	Name    string   `json:"Name"`
	Lessons []Lesson `json:"Lessons"`
}

// Lesson represents a single lesson within a course
type Lesson struct {
	LessonID    string     `json:"lessonId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Exercises   []Exercise `json:"exercises"`
}

// Exercise represents a learning exercise within a lesson
type Exercise struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Question      string      `json:"question"`
	Choices       []string    `json:"choices,omitempty"`
	CorrectAnswer interface{} `json:"correctAnswer,omitempty"`
	Pairs         [][]string  `json:"pairs,omitempty"`
	Items         []string    `json:"items,omitempty"`
	CorrectOrder  []int       `json:"correctOrder,omitempty"`
}

// Service defines the interface for course-related operations
type Service interface {
	// Course Management
	ListCourseNames() ([]string, error)
	GetCourseByName(name string) (*Course, error)
	GetLessonByID(courseName, lessonID string) (*Lesson, error)

	// Progress Tracking
	GetCourseProgress(userID int, courseName string) (*db.CourseProgress, error)
	GetLessonProgress(userID int, courseName, lessonID string) (*db.LessonProgress, error)
	UpdateLessonProgress(userID int, courseName, lessonID string, status string) error

	// Exercise Management
	VerifyExerciseAnswer(courseName, lessonID, exerciseID string, answer interface{}) (bool, error)

	// Data Loading
	LoadCourse(filename string) error
	LoadCourseDir() error
}

// Constants for status values
const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

// Constants for exercise types
const (
	ExerciseTypeMultipleChoice = "multiple_choice"
	ExerciseTypeMatching       = "matching"
	ExerciseTypeOrdering       = "ordering"
	ExerciseTrueFalse          = "true_false"
)
