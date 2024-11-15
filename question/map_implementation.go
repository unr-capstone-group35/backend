package question

import (
	"encoding/json"
	"errors"
	"os"
)

type MapStore struct {
	questions map[string]*Question // maps are already passed by ref
}

func NewMapStore() *MapStore {
	return &MapStore{
		questions: make(map[string]*Question),
	}
}

func (m *MapStore) ListQuestionNames() ([]string, error) {
	questionsSlice := make([]string, 0)

	for _, value := range m.questions {
		questionsSlice = append(questionsSlice, value.Name)
	}

	return questionsSlice, nil
}

func (m *MapStore) GetQuestionByName(name string) (*Question, error) {
	question, ok := m.questions[name]
	if !ok {
		return nil, errors.New("question does not exist")
	}
	return question, nil
}

func (m *MapStore) LoadQuestion(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	var question Question

	if err := json.NewDecoder(file).Decode(&question); err != nil {
		return err
	}

	m.questions[question.Name] = &question

	return nil
}
