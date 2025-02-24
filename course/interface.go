package course

import "github.com/tylerolson/capstone-backend/db"

// Constants for exercise types
type ExerciseType string

const (
	ExerciseTypeMultipleChoice ExerciseType = "multiple_choice"
	ExerciseTypeMatching       ExerciseType = "matching"
	ExerciseTypeOrdering       ExerciseType = "ordering"
	ExerciseTypeTrueFalse      ExerciseType = "true_false"
	ExerciseTypeFillBlank      ExerciseType = "fill_blank"
)

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
	ID            string       `json:"id"`
	Type          ExerciseType `json:"type"`
	Question      string       `json:"question"`
	Choices       []string     `json:"choices,omitempty"`
	CorrectAnswer interface{}  `json:"correctAnswer,omitempty"`
	Pairs         [][]string   `json:"pairs,omitempty"`
	Items         []string     `json:"items,omitempty"`
	CorrectOrder  []int        `json:"correctOrder,omitempty"`
}

type Service interface {
	// Course Management
	ListCourseNames() ([]string, error)
	GetCourseByID(courseID string) (*Course, error)
	GetLessonByID(courseID, lessonID string) (*Lesson, error)

	// Progress Tracking
	GetCourseProgress(userID int, courseID string) (*db.CourseProgress, error)
	GetLessonProgress(userID int, courseID string, lessonID string) (*db.LessonProgress, error)
	UpdateLessonProgress(userID int, courseID string, lessonID string, status db.Status) error

	// Exercise Management
	VerifyExerciseAnswer(courseID, lessonID, exerciseID string, answer interface{}) (bool, error)

	// Data Loading
	LoadCourse(filename string) error
	LoadCourseDir() error
}
