package domain

type Participant struct {
	ParticipantData
	assignedCourse     CourseData
	isAssigned         bool
	PrioritizedCourses []CourseData
}
