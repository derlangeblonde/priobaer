package solve

import "softbaer.dev/ass/internal/domain"

type priorityConstraint struct {
	level            domain.PriorityLevel
	courseConstraint courseConstraint
	participantID    domain.ParticipantID
}

func newPriorityConstraint(level domain.PriorityLevel, courseConstraint courseConstraint, pid domain.ParticipantID) priorityConstraint {
	return priorityConstraint{level: level, courseConstraint: courseConstraint, participantID: pid}
}
