package model

import (
	"slices"
	"testing"

	"github.com/matryer/is"
)

func TestSolveAssignmentWithExcessCapacity(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomCourse(WithCourseId(1), WithCapacity(0, 2)),
		RandomCourse(WithCourseId(2), WithCapacity(0, 2)),
		RandomCourse(WithCourseId(3), WithCapacity(0, 2)),
	}

	var particpants []Participant

	for i := 0; i < 5; i++ {
		particpants = append(particpants, RandomParticipant(WithParticipantId(i+1)))
	}

	assignments, err := SolveAssignment(courses, particpants)
	is.NoErr(err)

	is.Equal(countUniqueAssignments(assignments), len(particpants))
}

func TestSolveAssignmentWithScarceCapacity(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomCourse(WithCourseId(1), WithCapacity(0, 4)),
		RandomCourse(WithCourseId(2), WithCapacity(0, 3)),
		RandomCourse(WithCourseId(3), WithCapacity(0, 2)),
	}

	capacityTotal := 0

	for _, c := range courses {
		capacityTotal += c.MaxCapacity
	}

	var particpants []Participant

	for i := 0; i < 10; i++ {
		particpants = append(particpants, RandomParticipant(WithParticipantId(i+1)))
	}

	assignments, err := SolveAssignment(courses, particpants)
	is.NoErr(err)

	is.Equal(countUniqueAssignments(assignments), capacityTotal)
}

type assignmentOnlyNames struct {
	participantName string
	courseName      string
}

func countUniqueAssignments(assignments []Assignment) int {
	var assignmentsOnlyNames []assignmentOnlyNames

	for _, assignment := range assignments {
		assignmentsOnlyNames = append(assignmentsOnlyNames,
			assignmentOnlyNames{
				participantName: assignment.Participant.Prename + assignment.Participant.Surname,
				courseName:      assignment.Course.Name,
			})
	}

	assignmentsOnlyNames = slices.Compact(assignmentsOnlyNames)

	return len(assignmentsOnlyNames)
}
