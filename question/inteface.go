package question

type QuestionType int

const (
	Choice QuestionType = iota
	TrueFalse
)

type ChoiceContent struct {
	Choices []string `json:"Choices"`
}

type TrueFalseContent struct {
	Answer int `json:"Answer"`
}

type Question struct {
	Name    string       `json:"Name"`
	Type    QuestionType `json:"Type"`
	Content interface{}  `json:"Content"`
	Choices []string     `json:"Choices"`
}

type Service interface {
	ListQuestionNames() ([]string, error)
	GetQuestionsByName(name string) (*Question, error)
	LoadQuestion(filename string) error
}
