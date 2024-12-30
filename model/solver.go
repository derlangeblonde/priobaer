package model

type Assignment struct {
	Participant Participant
	Course      Course
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment) {

	availableCoursesMap := make(map[int]Course, 0)

	for _, course := range availableCourses {
		if course.RemainingCapacity() > 0 {
			availableCoursesMap[course.ID] = course
		}
	}

	return assignments
}
