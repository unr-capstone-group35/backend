package course

import (
	"encoding/json"
	"errors"
	"os"
)

type MapStore struct {
	courses map[string]*Course // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		courses: make(map[string]*Course),
	}
}

func (m *MapStore) ListCourseNames() ([]string, error) {
	coursesSlice := make([]string, 0)

	for _, value := range m.courses {
		coursesSlice = append(coursesSlice, value.Name)
	}

	return coursesSlice, nil
}

func (m *MapStore) GetCourseByName(name string) (*Course, error) {
	course, ok := m.courses[name]
	if !ok {
		return nil, errors.New("questions do not exist not exist")
	}
	return course, nil
}

func (m *MapStore) LoadCourse(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	var course Course

	if err := json.NewDecoder(file).Decode(&course); err != nil {
		return err
	}

	m.courses[course.Name] = &course

	return nil
}
