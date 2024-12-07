// course/json_implementation.go
package course

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

type Lesson struct {
	LessonID    string     `json:"lessonId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Exercises   []Exercise `json:"exercises"`
}

type Exercise struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Question      string      `json:"question"`
	Choices       []string    `json:"choices,omitempty"`
	CorrectAnswer interface{} `json:"correctAnswer,omitempty"`
	Pairs         [][]string  `json:"pairs,omitempty"`
	Items         []string    `json:"items,omitempty"`
	CorrectOrder  []int       `json:"correctOrder,omitempty"`
}

// JSONStore implements the course.Service interface
type JSONStore struct {
	courses map[string]*Course
	dataDir string
}

func NewJSONStore(dataDir string) *JSONStore {
	return &JSONStore{
		courses: make(map[string]*Course),
		dataDir: dataDir,
	}
}

// LoadCourseDir loads all courses from the data directory
func (j *JSONStore) LoadCourseDir() error {
	// Read all course directories
	dirs, err := ioutil.ReadDir(j.dataDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			if err := j.loadCourseFromDir(dir.Name()); err != nil {
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

	data, err := ioutil.ReadFile(rootPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &course); err != nil {
		return err
	}

	// Load all lesson files
	lessonFiles, err := filepath.Glob(filepath.Join(j.dataDir, courseName, "*.js"))
	if err != nil {
		return err
	}

	var lessons []Lesson
	for _, file := range lessonFiles {
		if filepath.Base(file) != "root.json" {
			data, err := ioutil.ReadFile(file)
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

	// Store the course
	j.courses[course.Name] = &course

	return nil
}

// Implement the Service interface methods
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

func (j *JSONStore) LoadCourse(filename string) error {
	return j.loadCourseFromDir(filepath.Base(filepath.Dir(filename)))
}
