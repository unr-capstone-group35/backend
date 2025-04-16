package points

import (
	"errors"
	"time"
)

var (
	ErrNoTransaction = errors.New("transaction does not exist")
)

// TransactionType defines the different ways users can earn points
type TransactionType string

const (
	// Different types of point transactions
	TransactionTypeCorrectAnswer   TransactionType = "correct_answer"
	TransactionTypeStreakBonus     TransactionType = "streak_bonus"
	TransactionTypeLessonCompleted TransactionType = "lesson_completed"
	TransactionTypeCourseCompleted TransactionType = "course_completed"
)

// PointsConfig defines the point values for different actions
type PointsConfig struct {
	CorrectAnswerPoints   int // Base points for a correct answer
	StreakBonusMultiplier int // Multiplier for consecutive correct answers
	MaxStreakBonus        int // Maximum streak bonus possible
	LessonCompletionBonus int // Bonus points for completing a lesson
	CourseCompletionBonus int // Bonus points for completing a course
}

// Default point values
var DefaultPointsConfig = PointsConfig{
	CorrectAnswerPoints:   10,
	StreakBonusMultiplier: 1,
	MaxStreakBonus:        50,
	LessonCompletionBonus: 100,
	CourseCompletionBonus: 500,
}

type PointTransaction struct {
	ID              int             `json:"id"`
	UserID          int             `json:"userId"`
	CourseID        string          `json:"courseId"`
	LessonID        string          `json:"lessonId"`
	ExerciseID      string          `json:"exerciseId,omitempty"`
	TransactionType TransactionType `json:"transactionType"`
	Points          int             `json:"points"`
	Description     string          `json:"description,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type UserPoints struct {
	UserID        int       `json:"userId"`
	TotalPoints   int       `json:"totalPoints"`
	CurrentStreak int       `json:"currentStreak"`
	MaxStreak     int       `json:"maxStreak"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type LessonPoints struct {
	UserID        int       `json:"userId"`
	CourseID      string    `json:"courseId"`
	LessonID      string    `json:"lessonId"`
	TotalPoints   int       `json:"totalPoints"`
	CurrentStreak int       `json:"currentStreak"`
	MaxStreak     int       `json:"maxStreak"`
	LastAttemptAt time.Time `json:"lastAttemptAt"`
}

type Service interface {
	// Award points for a correct answer and update streak
	AwardPointsForCorrectAnswer(userID int, courseID, lessonID, exerciseID string, isCorrect bool) (*PointTransaction, error)

	// Award bonus points for completing a lesson
	AwardLessonCompletionBonus(userID int, courseID, lessonID string) (*PointTransaction, error)

	// Award bonus points for completing a course
	AwardCourseCompletionBonus(userID int, courseID string) (*PointTransaction, error)

	// Reset streak for a specific lesson (typically called when starting a new lesson)
	ResetLessonStreak(userID int, courseID, lessonID string) error

	// Get user's total points
	GetUserTotalPoints(userID int) (*UserPoints, error)

	// Get user's points for a specific lesson
	GetLessonPoints(userID int, courseID, lessonID string) (*LessonPoints, error)

	// Get recent point transactions for a user
	GetRecentTransactions(userID int, limit int) ([]*PointTransaction, error)

	// Set points configuration
	SetPointsConfig(config PointsConfig)
}
