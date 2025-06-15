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

	participants := make([]Participant, 0, 5)
	priorities   := make([]Priority,     0, 5*len(courses))

	for i := 1; i <= 5; i++ {
		p := RandomParticipant(WithParticipantId(i))
		participants = append(participants, p)

		for level, c := range courses {
			priorities = append(
				priorities,
				NewPriority(PriorityLevel(level+1), c, p),
			)
		}
	}

	assignments, err := SolveAssignment(priorities)
	is.NoErr(err)

	is.Equal(countUniqueAssignments(assignments), len(participants))
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

	participants := make([]Participant, 0, 5)
	priorities   := make([]Priority,     0, 10*len(courses))

	for i := 1; i <= 10; i++ {
		p := RandomParticipant(WithParticipantId(i))
		participants = append(participants, p)

		for level, c := range courses {
			priorities = append(
				priorities,
				NewPriority(PriorityLevel(level+1), c, p),
			)
		}
	}

	assignments, err := SolveAssignment(priorities) 
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

participants := make([]Participant, 0, 6)
for i := 1; i <= 6; i++ {
	participants = append(participants, RandomParticipant(WithParticipantId(i)))
}

	priorities := []Priority{
		NewPriority(1, courses[0], participants[0]),
		NewPriority(2, courses[1], participants[0]),
		NewPriority(3, courses[2], participants[0]),

		NewPriority(1, courses[0], participants[1]),
		NewPriority(2, courses[1], participants[1]),
		NewPriority(3, courses[2], participants[1]),

		NewPriority(1, courses[1], participants[2]),
		NewPriority(2, courses[2], participants[2]),
		NewPriority(3, courses[0], participants[2]),

		NewPriority(1, courses[1], participants[3]),
		NewPriority(2, courses[2], participants[3]),
		NewPriority(3, courses[0], participants[3]),

		NewPriority(1, courses[2], participants[4]),
		NewPriority(2, courses[0], participants[4]),
		NewPriority(3, courses[1], participants[4]),

		NewPriority(1, courses[2], participants[5]),
		NewPriority(2, courses[0], participants[5]),
		NewPriority(3, courses[1], participants[5]),
	}

	assignments, err := SolveAssignment(priorities)
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
