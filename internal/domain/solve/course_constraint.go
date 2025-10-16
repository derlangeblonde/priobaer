package solve

import "softbaer.dev/ass/internal/domain"

type courseConstraint struct {
	CourseID          domain.CourseID
	GapToMinCapacity  int
	RemainingCapacity int
}

func newCourseConstraint(cid domain.CourseID, gapToMinCapacity, remainingCapacity int) courseConstraint {
	return courseConstraint{CourseID: cid, GapToMinCapacity: gapToMinCapacity, RemainingCapacity: remainingCapacity}
}
