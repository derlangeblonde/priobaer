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

type CourseList struct {
	CourseEntries    []Course
	NoCourseSelected bool
}
