// db/exercise_repository.go
package db

import (
	"encoding/json"
	"fmt"
)

func (d *Database) RecordExerciseAttempt(attempt *ExerciseAttempt) error {
	// Convert answer to JSON string before storing
	answerJSON, err := json.Marshal(attempt.Answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %v", err)
	}

	query := `
        INSERT INTO user_exercise_attempts 
        (user_id, course_name, lesson_id, exercise_id, attempt_number, answer, is_correct)
        VALUES ($1, $2, $3, $4, (
            SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM user_exercise_attempts 
            WHERE user_id = $5 AND course_name = $6 AND lesson_id = $7 AND exercise_id = $8
        ), $9, $10)
        RETURNING id, attempted_at`

	return d.DB.QueryRow(
		query,
		attempt.UserID,
		attempt.CourseName,
		attempt.LessonID,
		attempt.ExerciseID,
		attempt.UserID,
		attempt.CourseName,
		attempt.LessonID,
		attempt.ExerciseID,
		string(answerJSON), // Store as JSON string
		attempt.IsCorrect,
	).Scan(&attempt.ID, &attempt.AttemptedAt)
}
