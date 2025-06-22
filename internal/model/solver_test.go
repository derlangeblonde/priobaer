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

func TestSolveAssignmentAssertAssignmentCounts(t *testing.T) {
	testcases := []struct {
		name                      string
		courseCapacities          []CourseOption
		participantCount          int
		participantsPriosBuilders []participantPriosBuilder
		// ParticipantIds to CourseIds
		expectedAssignmentCountsPerCourseId map[int]int
	}{
		{
			"Min capacity respected",
			[]CourseOption{
				WithCapacity(2, 6),
				WithCapacity(2, 6),
				WithCapacity(2, 6),
			},
			6,
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{0, 1, 2}},
				{3, []int{0, 1, 2}},
				{4, []int{0, 1, 2}},
				{5, []int{1, 2, 0}},
			},
			map[int]int{
				1: 4,
				2: 2,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			courses := buildCourses(tc.courseCapacities)
			participants := buildParticipants(tc.participantCount)
			priorities := buildPriorities(tc.participantsPriosBuilders, courses, participants)

			assignments, err := SolveAssignment(priorities)

			is.NoErr(err)

			gotAssignmentCountPerCourseId := make(map[int]int, 0)
			for _, assignment := range assignments {
				gotAssignmentCountPerCourseId[assignment.Course.ID] += 1
			}

			for wantCourseId, wantAssignmentCount := range tc.expectedAssignmentCountsPerCourseId {
				gotAssignmentCount, ok := gotAssignmentCountPerCourseId[wantCourseId]
				is.True(ok) // no assignments at all

				if gotAssignmentCount != wantAssignmentCount {
					t.Fatalf("Want %d Particpants assigned to Course %d, but got %d assigned", wantAssignmentCount, wantCourseId, gotAssignmentCount)
				}
			}
		})
	}
}

func TestSolveAssignmentAssertExactAssignments(t *testing.T) {
	testcases := []struct {
		name                      string
		courseCapacities          []CourseOption
		participantCount          int
		participantsPriosBuilders []participantPriosBuilder
		// ParticipantIds to CourseIds
		expectedAssignments  map[int]int
		printInsteadOfAssert bool
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
			false,
		},
		{
			"Not enough capacity should trigger not solvable error",
			[]CourseOption{WithCapacity(0, 1), WithCapacity(0, 1), WithCapacity(0, 1)},
			4,
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{1, 2, 0}},
				{3, []int{2, 0, 1}},
			},
			map[int]int{},
			false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			courses := buildCourses(tc.courseCapacities)
			participants := buildParticipants(tc.participantCount)
			priorities := buildPriorities(tc.participantsPriosBuilders, courses, participants)

			assignments, err := SolveAssignment(priorities)

			if tc.printInsteadOfAssert {
				assignmentsMap := make(map[int]int, 0)
				for _, a := range assignments {
					assignmentsMap[a.Participant.ID] = a.Course.ID
				}

				t.Logf("Results for '%s'\n", tc.name)
				t.Logf("%+v\n", assignmentsMap)

				return
			}

			if len(tc.expectedAssignments) == 0 {
				is.Equal(err, NotSolvable)
				is.Equal(len(assignments), 0)

				return
			}

			is.NoErr(err)

			is.Equal(len(assignments), len(tc.expectedAssignments)) // each participant has gotten assigned

			for _, assignment := range assignments {
				gotCourseId := assignment.Course.ID
				wantCourseId, ok := tc.expectedAssignments[assignment.Participant.ID]
				is.True(ok) // there has to exist an assignment for each participant
				if gotCourseId != wantCourseId {
					t.Fatalf("Participant %d should have been assigned to Course %d, but is assigned to Course %d", assignment.Participant.ID, wantCourseId, gotCourseId)
				}
			}
		})
	}
}

type participantPriosBuilder struct {
	participantIndex          int
	prioritizedCoursesIndices []int
}

func buildCourses(courseCapacities []CourseOption) (courses []Course) {
	for i, capacityOption := range courseCapacities {
		courses = append(courses, RandomCourse(WithCourseId(i+1), capacityOption))
	}

	return
}

func buildParticipants(count int) (participants []Participant) {
	for i := 1; i <= count; i++ {
		participants = append(participants, RandomParticipant(WithParticipantId(i)))
	}

	return
}

func buildPriorities(participantPriosBuilders []participantPriosBuilder, courses []Course, participants []Participant) (prios Priorities) {
	for _, participantPrios := range participantPriosBuilders {
		for i, prioritizedCourseIndex := range participantPrios.prioritizedCoursesIndices {
			prios = append(prios, NewPriority(PriorityLevel(i+1), courses[prioritizedCourseIndex], participants[participantPrios.participantIndex]))
		}
	}

	return
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
