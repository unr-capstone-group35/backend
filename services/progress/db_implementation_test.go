package progress_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/lib/pq"
	"github.com/tylerolson/capstone-backend/db"
	"github.com/tylerolson/capstone-backend/services/progress"
	"github.com/tylerolson/capstone-backend/services/user"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := db.NewDatabase("dbuser", "dbpassword", "capstone_db", "", "")

	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestGetOrCreateCourseProgress(t *testing.T) {
	db := setupTestDB(t)
	service := progress.NewService(db)
	userService := user.NewService(db)

	user, err := userService.Get("testuser")
	if err != nil {
		t.Fatal(err)
	}

	courseID := "testCourse"

	progress, err := service.GetOrCreateCourseProgress(user.ID, courseID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.UserID != user.ID || progress.CourseID != courseID || progress.StartedAt.IsZero() {
		t.Errorf("unexpected progress data: %+v", progress)
	}

	jsonBytes, _ := json.MarshalIndent(progress, "", "  ")
	t.Log(string(jsonBytes))
}

func TestUpdateCourseProgress(t *testing.T) {
	db := setupTestDB(t)
	service := progress.NewService(db)
	userService := user.NewService(db)

	user, err := userService.Get("testuser")
	if err != nil {
		t.Fatal(err)
	}

	courseID := "testCourse"
	status := progress.StatusCompleted

	err = service.UpdateCourseProgress(user.ID, courseID, status)
	if err != nil {
		t.Fatal(err)
	}

	updatedProgress, err := service.GetOrCreateCourseProgress(user.ID, courseID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedProgress.Status != progress.StatusCompleted || updatedProgress.CompletedAt.IsZero() {
		t.Errorf("unexpected progress update: %+v", updatedProgress)
	}

	jsonBytes, _ := json.MarshalIndent(updatedProgress, "", "  ")
	t.Log(string(jsonBytes))
}

func TestGetOrCreateLessonProgress(t *testing.T) {
	db := setupTestDB(t)
	service := progress.NewService(db)
	userService := user.NewService(db)

	user, err := userService.Get("testuser")
	if err != nil {
		t.Fatal(err)
	}

	courseID := "testCourse"
	lessonID := "Lesson1"

	lessonProgress, err := service.GetOrCreateLessonProgress(user.ID, courseID, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if lessonProgress.UserID != user.ID || lessonProgress.CourseID != courseID || lessonProgress.LessonID != lessonID || lessonProgress.StartedAt.IsZero() {
		t.Errorf("unexpected lesson progress data: %+v", lessonProgress)
	}

	jsonBytes, _ := json.MarshalIndent(lessonProgress, "", "  ")
	t.Log(string(jsonBytes))
}

func TestUpdateLessonProgress(t *testing.T) {
	db := setupTestDB(t)
	service := progress.NewService(db)
	userService := user.NewService(db)

	user, err := userService.Get("testuser")
	if err != nil {
		t.Fatal(err)
	}

	courseID := "testCourse"
	lessonID := "Lesson1"
	status := progress.StatusCompleted

	err = service.UpdateLessonProgress(user.ID, courseID, lessonID, status)
	if err != nil {
		t.Fatal(err)
	}

	updatedLessonProgress, err := service.GetOrCreateLessonProgress(user.ID, courseID, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedLessonProgress.Status != progress.StatusCompleted || updatedLessonProgress.CompletedAt.IsZero() {
		t.Errorf("unexpected lesson progress update: %+v", updatedLessonProgress)
	}

	jsonBytes, _ := json.MarshalIndent(updatedLessonProgress, "", "  ")
	t.Log(string(jsonBytes))
}
