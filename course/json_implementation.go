// course/json_implementation.go
package course

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/tylerolson/capstone-backend/db"
)

// JSONStore implements the course.Service interface
type JSONStore struct {
	courses map[string]*Course
	dataDir string
	db      *db.Database
}

func NewJSONStore(dataDir string, database *db.Database) *JSONStore {
	return &JSONStore{
		courses: make(map[string]*Course),
		dataDir: dataDir,
		db:      database,
	}
}

// LoadCourseDir loads all courses from the data directory
func (j *JSONStore) LoadCourseDir() error {
	// Read all course directories
	entries, err := os.ReadDir(j.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := j.loadCourseFromDir(entry.Name()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (j *JSONStore) loadCourseFromDir(courseName string) error {
	// Load root.json first
	rootPath := filepath.Join(j.dataDir, courseName, "root.json")
	var course Course

	data, err := os.ReadFile(rootPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &course); err != nil {
		return err
	}

	// Load all lesson files
	lessonFiles, err := filepath.Glob(filepath.Join(j.dataDir, courseName, "*.json"))
	if err != nil {
		return err
	}

	lessons := make([]Lesson, 0)
	for _, file := range lessonFiles {
		if filepath.Base(file) != "root.json" {
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			var lesson Lesson
			if err := json.Unmarshal(data, &lesson); err != nil {
				return err
			}
			lessons = append(lessons, lesson)
		}
	}

	course.Lessons = lessons
	j.courses[course.Name] = &course

	return nil
}

// Implement Service interface methods
func (j *JSONStore) ListCourseNames() ([]string, error) {
	names := make([]string, 0, len(j.courses))
	for name := range j.courses {
		names = append(names, name)
	}
	return names, nil
}

func (j *JSONStore) GetCourseByName(name string) (*Course, error) {
	course, ok := j.courses[name]
	if !ok {
		return nil, errors.New("course not found")
	}
	return course, nil
}

func (j *JSONStore) GetLessonByID(courseName string, lessonID string) (*Lesson, error) {
	course, ok := j.courses[courseName]
	if !ok {
		return nil, errors.New("course not found")
	}

	for i := range course.Lessons {
		if course.Lessons[i].LessonID == lessonID {
			return &course.Lessons[i], nil
		}
	}

	return nil, errors.New("lesson not found")
}

func (j *JSONStore) LoadCourse(filename string) error {
	return j.loadCourseFromDir(filepath.Base(filepath.Dir(filename)))
}

// Progress tracking methods
func (j *JSONStore) GetCourseProgress(userID int, courseName string) (*db.CourseProgress, error) {
	return j.db.GetOrCreateCourseProgress(userID, courseName)
}

func (j *JSONStore) GetLessonProgress(userID int, courseName, lessonID string) (*db.LessonProgress, error) {
	progress, err := j.db.GetLessonProgress(userID, courseName, lessonID)
	if err != nil {
		if err.Error() == "lesson does not exist" {
			return &db.LessonProgress{
				UserID:     userID,
				CourseName: courseName,
				LessonID:   lessonID,
				Status:     "not_started",
				StartedAt:  time.Now().Format(time.RFC3339),
			}, nil
		}
		return nil, err
	}
	return progress, nil
}

func (j *JSONStore) UpdateLessonProgress(userID int, courseName, lessonID, status string) error {
	return j.db.UpdateLessonProgress(userID, courseName, lessonID, status)
}

func (j *JSONStore) VerifyExerciseAnswer(courseName, lessonID, exerciseID string, answer interface{}) (bool, error) {
	course, ok := j.courses[courseName]
	if !ok {
		return false, errors.New("course not found")
	}

	var targetExercise *Exercise
	for _, lesson := range course.Lessons {
		if lesson.LessonID == lessonID {
			for _, ex := range lesson.Exercises {
				if ex.ID == exerciseID {
					targetExercise = &ex
					break
				}
			}
			break
		}
	}

	if targetExercise == nil {
		return false, errors.New("exercise not found")
	}

	switch targetExercise.Type {
	case "multiple_choice":
		if choiceIdx, ok := answer.(float64); ok {
			correctAnswer, ok := targetExercise.CorrectAnswer.(float64)
			if !ok {
				return false, errors.New("invalid correct answer format for multiple choice")
			}
			return choiceIdx == correctAnswer, nil
		}
		return false, errors.New("invalid answer format for multiple choice")
	case "true_false":
		if boolAnswer, ok := answer.(bool); ok {
			return boolAnswer == targetExercise.CorrectAnswer.(bool), nil
		}
		return false, errors.New("invalid answer format for true/false")

	case "matching":
		answerPairs, ok := answer.([]interface{})
		if !ok {
			return false, errors.New("invalid answer format for matching")
		}
		return verifyMatchingAnswer(answerPairs, targetExercise.Pairs)

	case "ordering":
		answerOrder, ok := answer.([]interface{})
		if !ok {
			return false, errors.New("invalid answer format for ordering")
		}
		return verifyOrderingAnswer(answerOrder, targetExercise.CorrectOrder)

	default:
		return false, errors.New("unsupported exercise type")
	}
}

// Helper functions for exercise verification
func verifyMatchingAnswer(answer []interface{}, correctPairs [][]string) (bool, error) {
	if len(answer) != len(correctPairs) {
		return false, nil
	}

	for i, pair := range answer {
		answerPair, ok := pair.([]interface{})
		if !ok || len(answerPair) != 2 {
			return false, errors.New("invalid matching pair format")
		}

		left, ok1 := answerPair[0].(string)
		right, ok2 := answerPair[1].(string)
		if !ok1 || !ok2 {
			return false, errors.New("invalid matching pair values")
		}

		if left != correctPairs[i][0] || right != correctPairs[i][1] {
			return false, nil
		}
	}
	return true, nil
}

func verifyOrderingAnswer(answer []interface{}, correctOrder []int) (bool, error) {
	if len(answer) != len(correctOrder) {
		return false, nil
	}

	for i, val := range answer {
		idx, ok := val.(float64)
		if !ok {
			return false, errors.New("invalid ordering value format")
		}
		if int(idx) != correctOrder[i] {
			return false, nil
		}
	}
	return true, nil
}
