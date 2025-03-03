package progress

import (
	"errors"
	"time"
)

var (
	ErrNoProgress = errors.New("progress does not exist")
)

type Status string

const (
	StatusNotStarted Status = "not_started"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

type Progress struct {
	ID             int        `json:"id"`
	UserID         int        `json:"userId"`
	CourseID       string     `json:"courseId"`
	Status         Status     `json:"status"`
	StartedAt      time.Time  `json:"startedAt"`
	LastAccessedAt time.Time  `json:"lastAccessedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
}

type CourseProgress struct {
	Progress
}

type LessonProgress struct {
	Progress
	LessonID string `json:"lessonId"`
}

type ExerciseAttempt struct {
	ID            int       `json:"id"`
	UserID        int       `json:"userId"`
	CourseID      string    `json:"courseId"`
	LessonID      string    `json:"lessonId"`
	ExerciseID    string    `json:"exerciseId"`
	AttemptNumber int       `json:"attemptNumber"`
	Answer        string    `json:"answer"`
	IsCorrect     bool      `json:"is_correct"`
	AttemptedAt   time.Time `json:"attempted_at"`
}

type Service interface {
	GetOrCreateCourseProgress(userID int, courseID string) (*CourseProgress, error)
	UpdateCourseProgress(userID int, courseID string, status Status) error

	GetOrCreateLessonProgress(userID int, courseID string, lessonID string) (*LessonProgress, error)
	UpdateLessonProgress(userID int, courseID string, lessonID string, status Status) error

	RecordExerciseAttempt(attempt *ExerciseAttempt) error
}
