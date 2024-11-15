package lesson

import "github.com/tylerolson/capstone-backend/question"

type Lesson struct {
	Name            string            `json:"Name"`
	QuestionService *question.Service `json:"Lessons"`
}

type Service interface {
	ListLessonNames() ([]string, error)
	GetLessonByName(name string) (*Lesson, error)
	LoadLesson(filename string) error
}
