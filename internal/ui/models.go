package ui

type Course struct {
	ID          int
	Name        string
	MaxCapacity int
	MinCapacity int
	Allocation  int
	Selected    bool
	AsOobSwap   bool
}

func (c Course) Id() int {
	return c.ID
}

type UnassignedEntry struct {
	ParticipantsCount int
	ShouldRender      bool
	AsOobSwap         bool
	Selected          bool
}

type CourseList struct {
	CourseEntries    []Course
	UnassignedEntry  UnassignedEntry
	NoCourseSelected bool
}

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
