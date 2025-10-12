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

			gotAssignmentCountPerCourseId := make(map[int]int)
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
		testResultingAssignment   func(t *testing.T, resultingAssignments []Assignment, err error)
		printInsteadOfAssert      bool
	}{
		{
			"Solver respects max capacity of a course",
			[]CourseOption{WithCapacity(0, 2), WithCapacity(0, 2), WithCapacity(0, 2)},
			6,
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{0, 1, 2}},
				{3, []int{0, 1, 2}},
				{4, []int{0, 1, 2}},
				{5, []int{0, 1, 2}},
			},
			assertAllocations(map[int]int{
				1: 2,
				2: 2,
				3: 2,
			}),
			false,
		},
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
			assertExactAssignment(map[int]int{
				1: 1,
				2: 1,
				3: 2,
				4: 2,
				5: 3,
				6: 3,
			}),
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
			assertIsNotSolvable(),
			false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			courses := buildCourses(tc.courseCapacities)
			participants := buildParticipants(tc.participantCount)
			priorities := buildPriorities(tc.participantsPriosBuilders, courses, participants)

			assignments, err := SolveAssignment(priorities)

			if tc.printInsteadOfAssert {
				assignmentsMap := make(map[int]int)
				for _, a := range assignments {
					assignmentsMap[a.Participant.ID] = a.Course.ID
				}

				t.Logf("Results for '%s'\n", tc.name)
				t.Logf("%+v\n", assignmentsMap)

				return
			}

			tc.testResultingAssignment(t, assignments, err)
		})
	}
}

type assignmentAsserter func(t *testing.T, resultingAssignments []Assignment, err error)

// assertAllocations returns a func that checks whether some assignments have the same allocation(-distribution)
// as the wantAllocations (maps CourseIds to allocation) passed to this builder func.
func assertAllocations(wantAllocations map[int]int) assignmentAsserter {
	return func(t *testing.T, got []Assignment, err error) {
		is := is.New(t)

		is.NoErr(err) // want assignments to be successfully solved
		actualAllocations := make(map[int]int)
		for _, assignment := range got {
			count, _ := actualAllocations[assignment.Course.ID]
			actualAllocations[assignment.Course.ID] = count + 1
		}

		for courseId, expectedAllocation := range wantAllocations {
			actualAllocation, ok := actualAllocations[courseId]
			is.True(ok)                                    // Want an actual allocation for each expected allocation
			is.Equal(expectedAllocation, actualAllocation) // Want actual and expected allocation to match
		}
	}
}

func assertIsNotSolvable() assignmentAsserter {
	return func(t *testing.T, resultingAssignments []Assignment, err error) {
		is := is.New(t)

		is.Equal(err, NotSolvable)
		is.Equal(len(resultingAssignments), 0)
	}
}

// assertExactAssignment returns a func that checks whether some assignments match the wantAssignments passed
// to this builder func. ExpectedAssignments maps ParticipantIds to CourseIds.
func assertExactAssignment(wantAssignments map[int]int) assignmentAsserter {
	return func(t *testing.T, got []Assignment, err error) {
		is := is.New(t)

		is.NoErr(err)
		is.Equal(len(got), len(wantAssignments)) // each participant has gotten assigned

		for _, assignment := range got {
			gotCourseId := assignment.Course.ID
			wantCourseId, ok := wantAssignments[assignment.Participant.ID]
			is.True(ok) // there has to exist an assignment for each participant
			if gotCourseId != wantCourseId {
				t.Fatalf("Participant %d should have been assigned to Course %d, but is assigned to Course %d", assignment.Participant.ID, wantCourseId, gotCourseId)
			}
		}
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
