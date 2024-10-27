package lessons

type QuestionType int

const (
	Choice QuestionType = iota
	TrueFalse
)

type Question struct {
	Name    string       `json:"Name"`
	Type    QuestionType `json:"Type"`
	Choices []string     `json:"Choices"`
}

type Path struct {
	Name      string      `json:"Name"`
	Level     int         `json:"Level"`
	Questions []*Question `json:"Questions"`
}

type Service interface {
	ListPathNames() ([]string, error)
	ListQuestionsInPath(name string) ([]*Question, error)
}
