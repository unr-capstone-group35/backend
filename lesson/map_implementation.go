// lesson/map_implementation.go
package lesson

import (
	"encoding/json"
	"errors"
	"os"
)

type MapStore struct {
	lessons map[string]*Lesson // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		lessons: make(map[string]*Lesson),
	}
}

func (m *MapStore) ListLessonNames() ([]string, error) {
	lessonsSlice := make([]string, 0)

	for _, value := range m.lessons {
		lessonsSlice = append(lessonsSlice, value.Name)
	}

	return lessonsSlice, nil
}

func (m *MapStore) GetLessonByName(name string) (*Lesson, error) {
	lesson, ok := m.lessons[name]
	if !ok {
		return nil, errors.New("lesson does not exist")
	}
	return lesson, nil
}

func (m *MapStore) LoadLesson(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	var lesson Lesson

	if err := json.NewDecoder(file).Decode(&lesson); err != nil {
		return err
	}

	m.lessons[lesson.Name] = &lesson

	return nil
}
