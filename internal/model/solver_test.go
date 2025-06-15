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
	var priorities []Priority

	for i := range 5 {
		particpants = append(particpants, RandomParticipant(WithParticipantId(i+1)))
		priorities = append(priorities, Priority{Level: 1, CourseID: 1, ParticipantID: i+1})
		priorities = append(priorities, Priority{Level: 2, CourseID: 2, ParticipantID: i+1})
		priorities = append(priorities, Priority{Level: 3, CourseID: 3, ParticipantID: i+1})
	}

	assignments, err := SolveAssignment(courses, particpants, priorities)
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
	var priorities []Priority

	for i := 0; i < 10; i++ {
		particpants = append(particpants, RandomParticipant(WithParticipantId(i+1)))
		priorities = append(priorities, Priority{Level: 1, CourseID: 1, ParticipantID: i+1})
		priorities = append(priorities, Priority{Level: 2, CourseID: 2, ParticipantID: i+1})
		priorities = append(priorities, Priority{Level: 3, CourseID: 3, ParticipantID: i+1})
	}

	assignments, err := SolveAssignment(courses, particpants, priorities)
	is.NoErr(err)

	is.Equal(countUniqueAssignments(assignments), capacityTotal)
}

func TestSolveAssignmentWithRespectToPriorities(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		RandomCourse(WithCourseId(1), WithCapacity(0, 2)),
		RandomCourse(WithCourseId(2), WithCapacity(0, 2)),
		RandomCourse(WithCourseId(3), WithCapacity(0, 2)),
	}

	var particpants []Participant

	for i := range 6 {
		particpants = append(particpants, RandomParticipant(WithParticipantId(i+1)))
	}

	priorities := []Priority{
		{Level: 1, CourseID: 1, ParticipantID: 1},
		{Level: 2, CourseID: 2, ParticipantID: 1},
		{Level: 3, CourseID: 3, ParticipantID: 1},

		{Level: 1, CourseID: 1, ParticipantID: 2},
		{Level: 2, CourseID: 2, ParticipantID: 2},
		{Level: 3, CourseID: 3, ParticipantID: 2},

		{Level: 1, CourseID: 2, ParticipantID: 3},
		{Level: 2, CourseID: 3, ParticipantID: 3},
		{Level: 3, CourseID: 1, ParticipantID: 3},
		 
		{Level: 1, CourseID: 2, ParticipantID: 4},
		{Level: 2, CourseID: 3, ParticipantID: 4},
		{Level: 3, CourseID: 1, ParticipantID: 4},

		{Level: 1, CourseID: 3, ParticipantID: 5},
		{Level: 2, CourseID: 1, ParticipantID: 5},
		{Level: 3, CourseID: 2, ParticipantID: 5},

		{Level: 1, CourseID: 3, ParticipantID: 6},
		{Level: 2, CourseID: 1, ParticipantID: 6},
		{Level: 3, CourseID: 2, ParticipantID: 6},
	}

	assignments, err := SolveAssignment(courses, particpants, priorities)
	is.NoErr(err)

	// ParticipantIds to CourseIds
	expectedAssignments := map[int]int{
		1: 1,
		2: 1,
		3: 2,
		4: 2,
		5: 3,
		6: 3,
	}

	is.Equal(len(assignments), 6) // each participant has gotten assigned

	for _, assignment := range assignments {
		gotCourseId := assignment.Course.ID
		wantCourseId, ok := expectedAssignments[assignment.Participant.ID]
		is.True(ok) // there has to exist an assignment for each participant
		is.Equal(gotCourseId, wantCourseId)
	}
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
