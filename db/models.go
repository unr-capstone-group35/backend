package db

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means this won't be included in JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type CourseProgress struct {
	ID             int        `json:"id"`
	UserID         int        `json:"userId"`
	CourseID       string     `json:"courseId"`
	StartedAt      time.Time  `json:"startedAt"`
	LastAccessedAt time.Time  `json:"lastAccessedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
}

type LessonProgress struct {
	ID          int        `json:"id"`
	UserID      int        `json:"userId"`
	CourseID    string     `json:"courseId"`
	LessonID    string     `json:"lessonId"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
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

// Database scanning structs
type dbCourseProgress struct {
	ID             int
	UserID         int
	CourseID       string
	StartedAt      time.Time
	LastAccessedAt time.Time
	CompletedAt    sql.NullTime
}

type dbLessonProgress struct {
	ID          int
	UserID      int
	CourseID    string
	LessonID    string
	Status      string
	StartedAt   time.Time
	CompletedAt sql.NullTime
}

// Conversion methods
func (dbp *dbCourseProgress) toCourseProgress() *CourseProgress {
	cp := &CourseProgress{
		ID:             dbp.ID,
		UserID:         dbp.UserID,
		CourseID:       dbp.CourseID,
		StartedAt:      dbp.StartedAt,
		LastAccessedAt: dbp.LastAccessedAt,
	}
	if dbp.CompletedAt.Valid {
		cp.CompletedAt = &dbp.CompletedAt.Time
	}
	return cp
}

func (dbp *dbLessonProgress) toLessonProgress() *LessonProgress {
	lp := &LessonProgress{
		ID:        dbp.ID,
		UserID:    dbp.UserID,
		CourseID:  dbp.CourseID,
		LessonID:  dbp.LessonID,
		Status:    dbp.Status,
		StartedAt: dbp.StartedAt,
	}
	if dbp.CompletedAt.Valid {
		lp.CompletedAt = &dbp.CompletedAt.Time
	}
	return lp
}

type CourseProgressWithPercentage struct {
	*CourseProgress
	ProgressPercentage float64 `json:"progress_percentage"`
}
