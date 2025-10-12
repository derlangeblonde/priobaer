package ui

type Priority struct {
	Level      uint8
	CourseName string
}

type Participant struct {
	ID         int
	Prename    string
	Surname    string
	Priorities []Priority
}

func (p Participant) Id() int {
	return p.ID
}
