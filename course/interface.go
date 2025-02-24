package course

import "github.com/tylerolson/capstone-backend/db"

type Course struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Lessons []Lesson `json:"lessons"`
}

type Lesson struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Exercises   []Exercise `json:"exercises"`
}

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

type Service interface {
	// Course Management
	ListCourseNames() ([]string, error)
	GetCourseByID(courseID string) (*Course, error)
	GetLessonByID(courseID, lessonID string) (*Lesson, error)

	// Progress Tracking
	GetCourseProgress(userID int, courseID string) (*db.CourseProgress, error)
	GetLessonProgress(userID int, courseID string, lessonID string) (*db.LessonProgress, error)
	UpdateLessonProgress(userID int, courseID string, lessonID string, status string) error

	// Exercise Management
	VerifyExerciseAnswer(courseID, lessonID, exerciseID string, answer interface{}) (bool, error)

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
	ExerciseTypeTrueFalse      = "true_false"
	ExerciseTypeFillBlank      = "fill_blank"
)
