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
	AsOobSwap        bool
}

func NewOutOfBandCourseListUpdate() *CourseList {
	return &CourseList{
		AsOobSwap: true,
	}
}

func (cl *CourseList) SelectUnassignedEntry() *CourseList {
	cl.UnassignedEntry.Selected = true
	cl.UnassignedEntry.ShouldRender = true
	cl.UnassignedEntry.AsOobSwap = cl.AsOobSwap

	return cl
}

func (cl *CourseList) SetUnassignedCount(count int) *CourseList {
	cl.UnassignedEntry.ParticipantsCount = count
	cl.UnassignedEntry.ShouldRender = true
	cl.UnassignedEntry.AsOobSwap = cl.AsOobSwap

	return cl
}

func (cl *CourseList) AppendCourse(course Course) *CourseList {
	course.AsOobSwap = cl.AsOobSwap
	cl.CourseEntries = append(cl.CourseEntries, course)

	return cl
}
