package db

import (
	"database/sql"
	"fmt"
	"time"
)

func (d *Database) GetOrCreateCourseProgress(userID int, courseName string) (*CourseProgress, error) {
	dbProgress := &dbCourseProgress{}

	// Try to get existing progress
	query := `
        SELECT id, user_id, course_name, started_at, last_accessed_at, completed_at
        FROM user_course_progress
        WHERE user_id = $1 AND course_name = $2`

	err := d.DB.QueryRow(query, userID, courseName).Scan(
		&dbProgress.ID,
		&dbProgress.UserID,
		&dbProgress.CourseName,
		&dbProgress.StartedAt,
		&dbProgress.LastAccessedAt,
		&dbProgress.CompletedAt,
	)

	if err == sql.ErrNoRows {
		// Create new progress record
		now := time.Now()
		insertQuery := `
            INSERT INTO user_course_progress (user_id, course_name, started_at, last_accessed_at)
            VALUES ($1, $2, $3, $4)
            RETURNING id, user_id, course_name, started_at, last_accessed_at, completed_at`

		err = d.DB.QueryRow(
			insertQuery,
			userID,
			courseName,
			now,
			now,
		).Scan(
			&dbProgress.ID,
			&dbProgress.UserID,
			&dbProgress.CourseName,
			&dbProgress.StartedAt,
			&dbProgress.LastAccessedAt,
			&dbProgress.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create course progress: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get course progress: %v", err)
	}

	return dbProgress.toCourseProgress(), nil
}

func (d *Database) UpdateLessonProgress(userID int, courseName, lessonID, status string) error {
	query := `
		INSERT INTO user_lesson_progress (user_id, course_name, lesson_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, course_name, lesson_id) DO UPDATE
		SET status = EXCLUDED.status,
		    completed_at = CASE 
				WHEN EXCLUDED.status = 'completed' AND user_lesson_progress.status != 'completed'
				THEN CURRENT_TIMESTAMP
				ELSE user_lesson_progress.completed_at
			END`

	_, err := d.DB.Exec(query, userID, courseName, lessonID, status)
	return err
}

func (d *Database) GetLessonProgress(userID int, courseName, lessonID string) (*LessonProgress, error) {
	progress := &LessonProgress{}
	query := `
		SELECT id, user_id, course_name, lesson_id, status, started_at, completed_at
		FROM user_lesson_progress
		WHERE user_id = $1 AND course_name = $2 AND lesson_id = $3`

	err := d.DB.QueryRow(query, userID, courseName, lessonID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseName,
		&progress.LessonID,
		&progress.Status,
		&progress.StartedAt,
		&progress.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lesson does not exist")
	}
	if err != nil {
		return nil, err
	}
	return progress, nil
}

// GetCourseProgressWithPercentage gets course progress including completion percentage
func (d *Database) GetCourseProgressWithPercentage(userID int, courseName string) (*CourseProgressWithPercentage, error) {
	// First get the basic progress
	progress, err := d.GetOrCreateCourseProgress(userID, courseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get course progress: %v", err)
	}

	// Get total lessons count and completed lessons count
	query := `
        WITH lesson_counts AS (
            SELECT 
                COUNT(DISTINCT lesson_id) as total_lessons,
                COUNT(DISTINCT CASE 
                    WHEN status = 'completed' THEN lesson_id 
                    ELSE NULL 
                END) as completed_lessons
            FROM user_lesson_progress
            WHERE user_id = $1 AND course_name = $2
        )
        SELECT 
            CASE 
                WHEN total_lessons = 0 THEN 0
                ELSE (completed_lessons::float / total_lessons::float) * 100
            END as progress_percentage
        FROM lesson_counts
    `

	var progressPercentage float64
	err = d.DB.QueryRow(query, userID, courseName).Scan(&progressPercentage)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate progress percentage: %v", err)
	}

	return &CourseProgressWithPercentage{
		CourseProgress:     progress,
		ProgressPercentage: progressPercentage,
	}, nil
}
