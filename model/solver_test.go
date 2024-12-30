package model

import (
	"slices"
	"testing"

	"github.com/matryer/is"
)

type AssignmentOnlyNames struct {
	participantName string
	courseName      string
}

func  TestSolveAssignmentWithExcessCapacity(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomNameCourse(0, 2),
		RandomNameCourse(0, 2),
		RandomNameCourse(0, 2),
	}

	var particpants []Participant

	for i := 0; i < 5; i++ {
		particpants = append(particpants, RandomParticipant())
	}

	assignments := SolveAssignment(courses, particpants)
	var assignmentsOnlyNames []AssignmentOnlyNames

	for _, assignment := range assignments {
		assignmentsOnlyNames = append(assignmentsOnlyNames,
			AssignmentOnlyNames{
				participantName: assignment.Participant.Prename + assignment.Participant.Surname,
				courseName: assignment.Course.Name,
		})
	}

	assignmentsOnlyNames = slices.Compact(assignmentsOnlyNames)

	is.Equal(len(assignmentsOnlyNames), len(particpants))
}
