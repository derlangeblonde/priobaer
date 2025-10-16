package solve

import "softbaer.dev/ass/internal/domain"

type courseConstraint struct {
	courseId          domain.CourseID
	gapToMinCapacity  int
	remainingCapacity int
}

func newCourseConstraint(cid domain.CourseID, gapToMinCapacity, remainingCapacity int) courseConstraint {
	return courseConstraint{courseId: cid, gapToMinCapacity: gapToMinCapacity, remainingCapacity: remainingCapacity}
}
