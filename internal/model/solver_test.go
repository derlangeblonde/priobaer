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
	priorities := make([]Priority, 0, 5*len(courses))

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

type participantPriosBuilder struct {
	participantIndex          int
	prioritizedCoursesIndices []int
}

func TestSolveAssignmentWithRespectToPriorities(t *testing.T) {
	testcases := []struct {
		name             string
		courseCapacities []CourseOption
		participantCount int
		participantsPriosBuilders     []participantPriosBuilder
		// ParticipantIds to CourseIds
		expectedAssignments map[int]int
	}{
		{
			"Equal capacity & harmonic priorities should give everyone their first prio",
			[]CourseOption{WithCapacity(0, 2), WithCapacity(0, 2), WithCapacity(0, 2)},
			6,
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{1, 2, 0}},
				{3, []int{1, 2, 0}},
				{4, []int{2, 0, 1}},
				{5, []int{2, 0, 1}},
			},
			map[int]int{
				1: 1,
				2: 1,
				3: 2,
				4: 2,
				5: 3,
				6: 3,
			},
		},
		{
			"Not enough capacity should trigger not solvable error",
			[]CourseOption{WithCapacity(0, 2), WithCapacity(0, 1), WithCapacity(0, 2)},
			6,
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{1, 2, 0}},
				{3, []int{1, 2, 0}},
				{4, []int{2, 0, 1}},
				{5, []int{2, 0, 1}},
			},
			map[int]int{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			var courses []Course
			for i, capacityOption := range tc.courseCapacities {
				courses = append(courses, RandomCourse(WithCourseId(i+1), capacityOption))
			}

			var participants []Participant
			for i := 1; i <= tc.participantCount; i++ {
				participants = append(participants, RandomParticipant(WithParticipantId(i)))
			}

			var priorities []Priority
			for _, participantPrios := range tc.participantsPriosBuilders {
				for i, prioritizedCourseIndex := range participantPrios.prioritizedCoursesIndices {
					priorities = append(priorities, NewPriority(PriorityLevel(i+1), courses[prioritizedCourseIndex], participants[participantPrios.participantIndex]))
				}
			}

			assignments, err := SolveAssignment(priorities)
			
			if len(tc.expectedAssignments) == 0 {
				is.Equal(err, notSolvable)
				is.Equal(len(assignments), 0)

				return
			}

			is.NoErr(err)

			is.Equal(len(assignments), len(tc.expectedAssignments)) // each participant has gotten assigned

			for _, assignment := range assignments {
				gotCourseId := assignment.Course.ID
				wantCourseId, ok := tc.expectedAssignments[assignment.Participant.ID]
				is.True(ok) // there has to exist an assignment for each participant
				is.Equal(gotCourseId, wantCourseId)
			}
		})
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
