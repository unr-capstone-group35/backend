package lessons

import (
	"encoding/json"
	"io"
	"os"
)

type MapStore struct {
	paths map[string]*Path // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		paths: make(map[string]*Path),
	}
}

func (m *MapStore) ListPathNames() ([]string, error) {
	pathsSlice := make([]string, 0)

	for _, value := range m.paths {
		pathsSlice = append(pathsSlice, value.Name)
	}

	return pathsSlice, nil
}

func (m *MapStore) ListQuestionsInPath(name string) ([]*Question, error) {
	return m.paths[name].Questions, nil
}

func (m *MapStore) LoadPath(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	var path Path

	if err := json.Unmarshal(bytes, &path); err != nil {
		return err
	}

	m.paths[path.Name] = &path

	return nil
}
