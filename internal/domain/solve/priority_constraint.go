package solve

import "softbaer.dev/ass/internal/domain"

type priorityConstraint struct {
	Level            domain.PriorityLevel
	CourseConstraint courseConstraint
	ParticipantID    domain.ParticipantID
}

func newPriorityConstraint(level domain.PriorityLevel, courseConstraint courseConstraint, pid domain.ParticipantID) priorityConstraint {
	return priorityConstraint{Level: level, CourseConstraint: courseConstraint, ParticipantID: pid}
}
