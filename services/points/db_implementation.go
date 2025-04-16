package points

import (
	"database/sql"
	"fmt"
	"time"
)

type service struct {
	db     *sql.DB
	config PointsConfig
}

func NewService(database *sql.DB) Service {
	return &service{
		db:     database,
		config: DefaultPointsConfig,
	}
}

// SetPointsConfig allows for customizing the point values
func (s *service) SetPointsConfig(config PointsConfig) {
	s.config = config
}

func (s *service) AwardPointsForCorrectAnswer(userID int, courseID, lessonID, exerciseID string, isCorrect bool) (*PointTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if !isCorrect {
		_, err = tx.Exec(`
			UPDATE user_lesson_progress
			SET current_streak = 0
			WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`,
			userID, courseID, lessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to reset streak: %v", err)
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %v", err)
		}

		return nil, nil
	}

	var currentStreak int
	err = tx.QueryRow(`
		SELECT current_streak
		FROM user_lesson_progress
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`,
		userID, courseID, lessonID).Scan(&currentStreak)
	if err != nil {
		if err == sql.ErrNoRows {
			currentStreak = 0
		} else {
			return nil, fmt.Errorf("failed to get current streak: %v", err)
		}
	}

	newStreak := currentStreak + 1

	streakBonus := 0
	if newStreak > 1 {
		streakBonus = newStreak * s.config.StreakBonusMultiplier
		if streakBonus > s.config.MaxStreakBonus {
			streakBonus = s.config.MaxStreakBonus
		}
	}

	basePoints := s.config.CorrectAnswerPoints
	totalPoints := basePoints + streakBonus

	description := fmt.Sprintf("Correct answer (+%d points)", basePoints)
	if streakBonus > 0 {
		description += fmt.Sprintf(" with streak bonus of %d consecutive correct answers (+%d points)", newStreak, streakBonus)
	}

	_, err = tx.Exec(`
		UPDATE user_lesson_progress
		SET current_streak = $1, 
			max_streak = GREATEST(max_streak, $1),
			total_lesson_points = total_lesson_points + $2,
			last_accessed_at = NOW()
		WHERE user_id = $3 AND course_id = $4 AND lesson_id = $5`,
		newStreak, totalPoints, userID, courseID, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to update streak: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE user_exercise_attempts
		SET streak_at_attempt = $1, 
			points_earned = $2
		WHERE user_id = $3 AND course_id = $4 AND lesson_id = $5 AND exercise_id = $6
		AND id = (
			SELECT MAX(id) FROM user_exercise_attempts 
			WHERE user_id = $3 AND course_id = $4 AND lesson_id = $5 AND exercise_id = $6
		)`,
		newStreak, totalPoints, userID, courseID, lessonID, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to update exercise attempt: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE users
		SET total_points = total_points + $1
		WHERE id = $2`,
		totalPoints, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user total points: %v", err)
	}

	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow(`
		INSERT INTO user_point_transactions 
		(user_id, course_id, lesson_id, exercise_id, transaction_type, points, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`,
		userID, courseID, lessonID, exerciseID, TransactionTypeCorrectAnswer, totalPoints, description).
		Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert point transaction: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &PointTransaction{
		ID:              transactionID,
		UserID:          userID,
		CourseID:        courseID,
		LessonID:        lessonID,
		ExerciseID:      exerciseID,
		TransactionType: TransactionTypeCorrectAnswer,
		Points:          totalPoints,
		Description:     description,
		CreatedAt:       createdAt,
	}, nil
}

func (s *service) AwardLessonCompletionBonus(userID int, courseID, lessonID string) (*PointTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var status string
	err = tx.QueryRow(`
		SELECT status 
		FROM user_lesson_progress 
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`,
		userID, courseID, lessonID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to check lesson completion status: %v", err)
	}

	if status == "completed" {
		var exists bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM user_point_transactions 
				WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3 
				AND transaction_type = $4
			)`,
			userID, courseID, lessonID, TransactionTypeLessonCompleted).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing bonus: %v", err)
		}
		if exists {
			return nil, nil
		}
	}

	bonusPoints := s.config.LessonCompletionBonus
	description := fmt.Sprintf("Lesson completion bonus (+%d points)", bonusPoints)

	_, err = tx.Exec(`
		UPDATE users 
		SET total_points = total_points + $1 
		WHERE id = $2`,
		bonusPoints, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user total points: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE user_lesson_progress 
		SET total_lesson_points = total_lesson_points + $1,
			status = 'completed',
			completed_at = CASE WHEN completed_at IS NULL THEN NOW() ELSE completed_at END
		WHERE user_id = $2 AND course_id = $3 AND lesson_id = $4`,
		bonusPoints, userID, courseID, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to update lesson points: %v", err)
	}

	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow(`
		INSERT INTO user_point_transactions 
		(user_id, course_id, lesson_id, transaction_type, points, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		userID, courseID, lessonID, TransactionTypeLessonCompleted, bonusPoints, description).
		Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert point transaction: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &PointTransaction{
		ID:              transactionID,
		UserID:          userID,
		CourseID:        courseID,
		LessonID:        lessonID,
		TransactionType: TransactionTypeLessonCompleted,
		Points:          bonusPoints,
		Description:     description,
		CreatedAt:       createdAt,
	}, nil
}

func (s *service) AwardCourseCompletionBonus(userID int, courseID string) (*PointTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var status string
	err = tx.QueryRow(`
		SELECT status 
		FROM user_course_progress 
		WHERE user_id = $1 AND course_id = $2`,
		userID, courseID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to check course completion status: %v", err)
	}

	if status == "completed" {
		var exists bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM user_point_transactions 
				WHERE user_id = $1 AND course_id = $2 
				AND transaction_type = $3
			)`,
			userID, courseID, TransactionTypeCourseCompleted).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing bonus: %v", err)
		}
		if exists {
			return nil, nil
		}
	}

	bonusPoints := s.config.CourseCompletionBonus
	description := fmt.Sprintf("Course completion bonus (+%d points)", bonusPoints)

	_, err = tx.Exec(`
		UPDATE users 
		SET total_points = total_points + $1 
		WHERE id = $2`,
		bonusPoints, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user total points: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE user_course_progress 
		SET total_course_points = total_course_points + $1,
			status = 'completed',
			completed_at = CASE WHEN completed_at IS NULL THEN NOW() ELSE completed_at END
		WHERE user_id = $2 AND course_id = $3`,
		bonusPoints, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to update course points: %v", err)
	}

	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow(`
		INSERT INTO user_point_transactions 
		(user_id, course_id, transaction_type, points, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		userID, courseID, TransactionTypeCourseCompleted, bonusPoints, description).
		Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert point transaction: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &PointTransaction{
		ID:              transactionID,
		UserID:          userID,
		CourseID:        courseID,
		TransactionType: TransactionTypeCourseCompleted,
		Points:          bonusPoints,
		Description:     description,
		CreatedAt:       createdAt,
	}, nil
}

func (s *service) ResetLessonStreak(userID int, courseID, lessonID string) error {
	_, err := s.db.Exec(`
		UPDATE user_lesson_progress
		SET current_streak = 0
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`,
		userID, courseID, lessonID)
	if err != nil {
		return fmt.Errorf("failed to reset streak: %v", err)
	}
	return nil
}

func (s *service) GetUserTotalPoints(userID int) (*UserPoints, error) {
	userPoints := &UserPoints{UserID: userID}

	err := s.db.QueryRow(`
		SELECT total_points, updated_at
		FROM users
		WHERE id = $1`,
		userID).Scan(&userPoints.TotalPoints, &userPoints.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user total points: %v", err)
	}

	err = s.db.QueryRow(`
		SELECT COALESCE(MAX(current_streak), 0), COALESCE(MAX(max_streak), 0)
		FROM user_lesson_progress
		WHERE user_id = $1`,
		userID).Scan(&userPoints.CurrentStreak, &userPoints.MaxStreak)
	if err != nil {
		return nil, fmt.Errorf("failed to get user streaks: %v", err)
	}

	return userPoints, nil
}

func (s *service) GetLessonPoints(userID int, courseID, lessonID string) (*LessonPoints, error) {
	lessonPoints := &LessonPoints{
		UserID:   userID,
		CourseID: courseID,
		LessonID: lessonID,
	}

	err := s.db.QueryRow(`
		SELECT 
			total_lesson_points, 
			current_streak, 
			max_streak, 
			COALESCE(last_accessed_at, started_at) AS last_attempt_at
		FROM user_lesson_progress
		WHERE user_id = $1 AND course_id = $2 AND lesson_id = $3`,
		userID, courseID, lessonID).Scan(
		&lessonPoints.TotalPoints,
		&lessonPoints.CurrentStreak,
		&lessonPoints.MaxStreak,
		&lessonPoints.LastAttemptAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			lessonPoints.TotalPoints = 0
			lessonPoints.CurrentStreak = 0
			lessonPoints.MaxStreak = 0
			lessonPoints.LastAttemptAt = time.Now()
			return lessonPoints, nil
		}
		return nil, fmt.Errorf("failed to get lesson points: %v", err)
	}

	return lessonPoints, nil
}

func (s *service) GetRecentTransactions(userID int, limit int) ([]*PointTransaction, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	rows, err := s.db.Query(`
		SELECT 
			id, 
			user_id, 
			course_id, 
			lesson_id, 
			exercise_id, 
			transaction_type, 
			points, 
			description, 
			created_at
		FROM user_point_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent transactions: %v", err)
	}
	defer rows.Close()

	var transactions []*PointTransaction
	for rows.Next() {
		transaction := &PointTransaction{}
		var exerciseID sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CourseID,
			&transaction.LessonID,
			&exerciseID,
			&transaction.TransactionType,
			&transaction.Points,
			&transaction.Description,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %v", err)
		}

		if exerciseID.Valid {
			transaction.ExerciseID = exerciseID.String
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %v", err)
	}

	return transactions, nil
}
