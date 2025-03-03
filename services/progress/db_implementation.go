package progress

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type service struct {
	db *sql.DB
}

func NewService(database *sql.DB) Service {
	return &service{db: database}
}

func (s *service) GetOrCreateCourseProgress(userID int, courseID string) (*CourseProgress, error) {
	progress := &CourseProgress{}

	// Try to get existing progress
	query := `
			SELECT id, user_id, course_id, status, started_at, last_accessed_at, completed_at
			FROM user_course_progress
			WHERE user_id = $1 AND course_id = $2`

	err := s.db.QueryRow(query, userID, courseID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseID,
		&progress.Status,
		&progress.StartedAt,
		&progress.LastAccessedAt,
		&progress.CompletedAt,
	)

	if err == sql.ErrNoRows {
		now := time.Now()
		insertQuery := `
				INSERT INTO user_course_progress (user_id, course_id, status, started_at)
				VALUES ($1, $2, $3, $4)
				RETURNING id, user_id, course_id, status, started_at, last_accessed_at, completed_at`

		err = s.db.QueryRow(
			insertQuery,
			userID,
			courseID,
			StatusNotStarted,
			now,
		).Scan(
			&progress.ID,
			&progress.UserID,
			&progress.CourseID,
			&progress.Status,
			&progress.StartedAt,
			&progress.LastAccessedAt,
			&progress.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create course progress: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get course progress: %v", err)
	}

	return progress, nil
}

func (s *service) UpdateCourseProgress(userID int, courseID string, status Status) error {
	query := `
		INSERT INTO user_course_progress (user_id, course_id, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, course_id) DO UPDATE
		SET status = EXCLUDED.status,
		    completed_at = CASE 
				WHEN EXCLUDED.status = 'completed' AND user_course_progress.status != 'completed'
				THEN CURRENT_TIMESTAMP
				ELSE user_course_progress.completed_at
			END`

	_, err := s.db.Exec(query, userID, courseID, status)
	return err

}

func (s *service) GetOrCreateLessonProgress(userID int, courseID string, lessonID string) (*LessonProgress, error) {
	progress := &LessonProgress{}

	query := `
		SELECT id, user_id, course_id, lesson_id, status, started_at, last_accessed_at, completed_at
		FROM user_lesson_progress
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`

	err := s.db.QueryRow(query, userID, courseID, lessonID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseID,
		&progress.LessonID,
		&progress.Status,
		&progress.StartedAt,
		&progress.LastAccessedAt,
		&progress.CompletedAt,
	)

	if err == sql.ErrNoRows {
		now := time.Now()
		insertQuery := `
				INSERT INTO user_lesson_progress (user_id, course_id, status, lesson_id, started_at)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, user_id, course_id, lesson_id, status, started_at, last_accessed_at, completed_at`

		err = s.db.QueryRow(
			insertQuery,
			userID,
			courseID,
			StatusNotStarted,
			lessonID,
			now,
		).Scan(
			&progress.ID,
			&progress.UserID,
			&progress.CourseID,
			&progress.LessonID,
			&progress.Status,
			&progress.StartedAt,
			&progress.LastAccessedAt,
			&progress.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create course progress: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get course progress: %v", err)
	}

	return progress, nil
}

func (s *service) GetLessonProgress(userID int, courseID, lessonID string) (*LessonProgress, error) {
	progress := &LessonProgress{}
	query := `
		SELECT id, user_id, course_id, lesson_id, status, started_at, completed_at
		FROM user_lesson_progress
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`

	err := s.db.QueryRow(query, userID, courseID, lessonID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseID,
		&progress.LessonID,
		&progress.Status,
		&progress.StartedAt,
		&progress.CompletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoProgress
		}

		return nil, err
	}

	return progress, nil
}

func (s *service) UpdateLessonProgress(userID int, courseID string, lessonID string, status Status) error {
	query := `
		INSERT INTO user_lesson_progress (user_id, course_id, lesson_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, course_id, lesson_id) DO UPDATE
		SET status = EXCLUDED.status,
		    completed_at = CASE 
				WHEN EXCLUDED.status = 'completed' AND user_lesson_progress.status != 'completed'
				THEN CURRENT_TIMESTAMP
				ELSE user_lesson_progress.completed_at
			END`

	_, err := s.db.Exec(query, userID, courseID, lessonID, status)
	return err
}

func (s *service) RecordExerciseAttempt(attempt *ExerciseAttempt) error {
	// Convert answer to JSON string before storing
	answerJSON, err := json.Marshal(attempt.Answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %v", err)
	}

	query := `
        INSERT INTO user_exercise_attempts 
        (user_id, course_id, lesson_id, exercise_id, attempt_number, answer, is_correct)
        VALUES ($1, $2, $3, $4, (
            SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM user_exercise_attempts 
            WHERE user_id = $5 AND course_id = $6 AND lesson_id = $7 AND exercise_id = $8
        ), $9, $10)
        RETURNING id, attempted_at`

	return s.db.QueryRow(
		query,
		attempt.UserID,
		attempt.CourseID,
		attempt.LessonID,
		attempt.ExerciseID,
		attempt.UserID,
		attempt.CourseID,
		attempt.LessonID,
		attempt.ExerciseID,
		string(answerJSON), // Store as JSON string
		attempt.IsCorrect,
	).Scan(&attempt.ID, &attempt.AttemptedAt)
}
