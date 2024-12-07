// course/interface.go
package course

import "github.com/tylerolson/capstone-backend/lesson"

type Course struct {
	Name          string          `json:"Name"`
	LessonService *lesson.Service `json:"Lessons"`
}

type Service interface {
	ListCourseNames() ([]string, error)
	GetCourseByName(name string) (*Course, error)
	LoadCourse(filename string) error
}
