package model

import (
	"slices"
	"testing"

	"github.com/matryer/is"
)

func TestSolveAssignmentWithExcessCapacity(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomNameCourse(1, 0, 2),
		RandomNameCourse(2, 0, 2),
		RandomNameCourse(3, 0, 2),
	}

	var particpants []Participant

	for i := 0; i < 5; i++ {
		particpants = append(particpants, RandomNameParticipant(i+1))
	}

	assignments := SolveAssignment(courses, particpants)

	is.Equal(countUniqueAssignments(assignments), len(particpants))
}

func TestSolveAssignmentWithScarceCapacity(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomNameCourse(1, 0, 4),
		RandomNameCourse(2, 0, 3),
		RandomNameCourse(3, 0, 2),
	}

	capacityTotal := 0

	for _, c := range courses {
		capacityTotal += c.MaxCapacity
	}

	var particpants []Participant

	for i := 0; i < 10; i++ {
		particpants = append(particpants, RandomNameParticipant(i+1))
	}

	assignments := SolveAssignment(courses, particpants)

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
