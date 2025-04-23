package points

import "time"

// Define transaction types
const (
	TransactionTypeCorrectAnswer    = "correct_answer"
	TransactionTypeLessonCompleted  = "lesson_completed"
	TransactionTypeCourseCompleted  = "course_completed"
	TransactionTypeDailyStreakBonus = "daily_streak_bonus"
)

// PointsConfig defines the point values for different actions
type PointsConfig struct {
	CorrectAnswerPoints      int
	StreakBonusMultiplier    int
	MaxStreakBonus           int
	LessonCompletionBonus    int
	CourseCompletionBonus    int
	DailyStreakBonusPoints   int
	DailyStreakMilestones    []int // Milestones for extra bonuses (e.g., [7, 30, 365])
	MilestoneBonusMultiplier int   // Multiplier for milestone bonuses
}

// Default point configuration
var DefaultPointsConfig = PointsConfig{
	CorrectAnswerPoints:      10,
	StreakBonusMultiplier:    2,
	MaxStreakBonus:           50,
	LessonCompletionBonus:    50,
	CourseCompletionBonus:    200,
	DailyStreakBonusPoints:   20,
	DailyStreakMilestones:    []int{7, 30, 100, 365},
	MilestoneBonusMultiplier: 5,
}

// PointTransaction represents a points transaction record
type PointTransaction struct {
	ID              int       `json:"id"`
	UserID          int       `json:"userId"`
	CourseID        string    `json:"courseId"`
	LessonID        string    `json:"lessonId,omitempty"`
	ExerciseID      string    `json:"exerciseId,omitempty"`
	TransactionType string    `json:"transactionType"`
	Points          int       `json:"points"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"createdAt"`
}

// UserPoints represents a user's total points and streaks
type UserPoints struct {
	UserID        int       `json:"userId"`
	TotalPoints   int       `json:"totalPoints"`
	CurrentStreak int       `json:"currentStreak"`
	MaxStreak     int       `json:"maxStreak"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// LessonPoints represents points and streak for a specific lesson
type LessonPoints struct {
	UserID        int       `json:"userId"`
	CourseID      string    `json:"courseId"`
	LessonID      string    `json:"lessonId"`
	TotalPoints   int       `json:"totalPoints"`
	CurrentStreak int       `json:"currentStreak"`
	MaxStreak     int       `json:"maxStreak"`
	LastAttemptAt time.Time `json:"lastAttemptAt"`
}

// DailyStreakInfo represents a user's daily streak information
type DailyStreakInfo struct {
	UserID          int       `json:"userId"`
	CurrentStreak   int       `json:"currentStreak"`
	MaxStreak       int       `json:"maxStreak"`
	LastLoginDate   time.Time `json:"lastLoginDate"`
	NextMilestone   int       `json:"nextMilestone,omitempty"`
	DaysToMilestone int       `json:"daysToMilestone,omitempty"`
}

// AccuracyStats represents a user's accuracy statistics
type AccuracyStats struct {
	UserID          int     `json:"userId"`
	TotalAttempts   int     `json:"totalAttempts"`
	CorrectAttempts int     `json:"correctAttempts"`
	AccuracyRate    float64 `json:"accuracyRate"` // Percentage (0-100)
}

// Service defines the points service interface
type Service interface {
	// Existing methods
	SetPointsConfig(config PointsConfig)
	AwardPointsForCorrectAnswer(userID int, courseID, lessonID, exerciseID string, isCorrect bool) (*PointTransaction, error)
	AwardLessonCompletionBonus(userID int, courseID, lessonID string) (*PointTransaction, error)
	AwardCourseCompletionBonus(userID int, courseID string) (*PointTransaction, error)
	ResetLessonStreak(userID int, courseID, lessonID string) error
	GetUserTotalPoints(userID int) (*UserPoints, error)
	GetLessonPoints(userID int, courseID, lessonID string) (*LessonPoints, error)
	GetRecentTransactions(userID int, limit int) ([]*PointTransaction, error)

	// New methods for daily streak
	UpdateDailyStreak(userID int) (*PointTransaction, error)
	GetDailyStreak(userID int) (*DailyStreakInfo, error)

	// New methods for accuracy tracking
	UpdateAccuracyStats(userID int, isCorrect bool) error
	GetAccuracyStats(userID int) (*AccuracyStats, error)
}
