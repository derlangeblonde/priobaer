package solve

import "softbaer.dev/ass/internal/domain"

type computedAssignment struct {
	participantID domain.ParticipantID
	courseID      domain.CourseID
}

func newComputedAssignment(participantId domain.ParticipantID, courseId domain.CourseID) computedAssignment {
	return computedAssignment{participantID: participantId, courseID: courseId}
}
