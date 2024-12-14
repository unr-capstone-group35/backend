// db/models.go
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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type CourseProgress struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id"`
	CourseName     string     `json:"course_name"`
	StartedAt      time.Time  `json:"started_at"`
	LastAccessedAt time.Time  `json:"last_accessed_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type LessonProgress struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	CourseName  string     `json:"course_name"`
	LessonID    string     `json:"lesson_id"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ExerciseAttempt struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	CourseName    string    `json:"course_name"`
	LessonID      string    `json:"lesson_id"`
	ExerciseID    string    `json:"exercise_id"`
	AttemptNumber int       `json:"attempt_number"`
	Answer        string    `json:"answer"`
	IsCorrect     bool      `json:"is_correct"`
	AttemptedAt   time.Time `json:"attempted_at"`
}

// Database scanning structs
type dbCourseProgress struct {
	ID             int
	UserID         int
	CourseName     string
	StartedAt      time.Time
	LastAccessedAt time.Time
	CompletedAt    sql.NullTime
}

type dbLessonProgress struct {
	ID          int
	UserID      int
	CourseName  string
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
		CourseName:     dbp.CourseName,
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
		ID:         dbp.ID,
		UserID:     dbp.UserID,
		CourseName: dbp.CourseName,
		LessonID:   dbp.LessonID,
		Status:     dbp.Status,
		StartedAt:  dbp.StartedAt,
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
