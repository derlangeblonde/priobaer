package view

type Course struct {
	ID          int
	Name        string
	MaxCapacity int
	MinCapacity int
	Allocation  int
	Selected    bool
	AsOobSwap   bool
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
