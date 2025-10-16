package solve

import (
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/domain"
)

func TestSolveAssignmentSolvesDifferentScenariosCorrectly(t *testing.T) {
	testcases := []struct {
		name                      string
		courseConstraints         []courseConstraint
		participantsPriosBuilders []participantPriosBuilder
		testResultingAssignment   func(t *testing.T, computedAssignments []computedAssignment, err error)
		printInsteadOfAssert      bool
	}{
		{
			"Solver respects max capacity of a course",
			[]courseConstraint{newCourseConstraint(1, 0, 2), newCourseConstraint(2, 0, 2), newCourseConstraint(3, 0, 2)},
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{0, 1, 2}},
				{3, []int{0, 1, 2}},
				{4, []int{0, 1, 2}},
				{5, []int{0, 1, 2}},
			},
			assertAllocations(map[domain.CourseID]int{
				1: 2,
				2: 2,
				3: 2,
			}),
			false,
		},
		{
			"Equal capacity & harmonic priorities should give everyone their first prio",
			[]courseConstraint{newCourseConstraint(1, 0, 2), newCourseConstraint(2, 0, 2), newCourseConstraint(3, 0, 2)},
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{1, 2, 0}},
				{3, []int{1, 2, 0}},
				{4, []int{2, 0, 1}},
				{5, []int{2, 0, 1}},
			},
			assertExactAssignment(map[domain.ParticipantID]domain.CourseID{
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
			[]courseConstraint{newCourseConstraint(1, 0, 1), newCourseConstraint(2, 0, 1), newCourseConstraint(3, 0, 1)},
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{1, 2, 0}},
				{3, []int{2, 0, 1}},
			},
			assertIsNotSolvable(),
			false,
		},
		{
			"Assignment with excess capacity assigns all participants",
			[]courseConstraint{newCourseConstraint(1, 0, 2), newCourseConstraint(2, 0, 2), newCourseConstraint(3, 0, 2)},
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{0, 1, 2}},
				{3, []int{0, 1, 2}},
				{4, []int{0, 1, 2}},
			},
			assertAllParticipantsAssigned(5),
			false,
		},
		{
			"Min capacity respected",
			[]courseConstraint{newCourseConstraint(1, 2, 6), newCourseConstraint(2, 2, 6), newCourseConstraint(3, 2, 6)},
			[]participantPriosBuilder{
				{0, []int{0, 1, 2}},
				{1, []int{0, 1, 2}},
				{2, []int{0, 1, 2}},
				{3, []int{0, 1, 2}},
				{4, []int{0, 1, 2}},
				{5, []int{1, 2, 0}},
			},
			assertAllocations(map[domain.CourseID]int{
				1: 4,
				2: 2,
			}),
			false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			priorityConstraints := buildPriorityConstraints(tc.participantsPriosBuilders, tc.courseConstraints)

			assignments, err := computeOptimalAssignments(priorityConstraints)

			if tc.printInsteadOfAssert {
				assignmentsMap := make(map[domain.ParticipantID]domain.CourseID)
				for _, a := range assignments {
					assignmentsMap[a.participantID] = a.courseID
				}

				t.Logf("Results for '%s'\n", tc.name)
				t.Logf("%+v\n", assignmentsMap)

				return
			}

			tc.testResultingAssignment(t, assignments, err)
		})
	}
}

type assignmentAsserter func(t *testing.T, resultingAssignments []computedAssignment, err error)

func assertAllParticipantsAssigned(expectedParticipantCount int) assignmentAsserter {
	return func(t *testing.T, got []computedAssignment, err error) {
		is := is.New(t)

		is.NoErr(err)

		// Count unique participants assigned
		participantSet := make(map[domain.ParticipantID]bool)
		for _, assignment := range got {
			participantSet[assignment.participantID] = true
		}

		is.Equal(len(participantSet), expectedParticipantCount) // all participants should be assigned
	}
}

// assertAllocations returns a func that checks whether some assignments have the same allocation(-distribution)
// as the wantAllocations (maps CourseIds to allocation) passed to this builder func.
func assertAllocations(wantAllocations map[domain.CourseID]int) assignmentAsserter {
	return func(t *testing.T, got []computedAssignment, err error) {
		is := is.New(t)

		is.NoErr(err) // want assignments to be successfully solved
		actualAllocations := make(map[domain.CourseID]int)
		for _, assignment := range got {
			count, _ := actualAllocations[assignment.courseID]
			actualAllocations[assignment.courseID] = count + 1
		}

		for courseId, expectedAllocation := range wantAllocations {
			actualAllocation, ok := actualAllocations[courseId]
			is.True(ok)                                    // Want an actual allocation for each expected allocation
			is.Equal(expectedAllocation, actualAllocation) // Want actual and expected allocation to match
		}
	}
}

func assertIsNotSolvable() assignmentAsserter {
	return func(t *testing.T, resultingAssignments []computedAssignment, err error) {
		is := is.New(t)

		is.Equal(err, NotSolvable)
		is.Equal(len(resultingAssignments), 0)
	}
}

// assertExactAssignment returns a func that checks whether some assignments match the wantAssignments passed
// to this builder func.
func assertExactAssignment(wantAssignments map[domain.ParticipantID]domain.CourseID) assignmentAsserter {
	return func(t *testing.T, got []computedAssignment, err error) {
		is := is.New(t)

		is.NoErr(err)
		is.Equal(len(got), len(wantAssignments)) // each participant has gotten assigned

		for _, assignment := range got {
			gotCourseId := assignment.courseID
			wantCourseId, ok := wantAssignments[assignment.participantID]
			is.True(ok) // there has to exist an assignment for each participant
			if gotCourseId != wantCourseId {
				t.Fatalf("Participant %d should have been assigned to Course %d, but is assigned to Course %d", assignment.participantID, wantCourseId, gotCourseId)
			}
		}
	}
}

type participantPriosBuilder struct {
	participantIndex          int
	prioritizedCoursesIndices []int
}

func buildPriorityConstraints(prioMappings []participantPriosBuilder, courseConstraints []courseConstraint) []priorityConstraint {
	var result []priorityConstraint
	for _, prioMapping := range prioMappings {
		for i, courseIndex := range prioMapping.prioritizedCoursesIndices {
			result = append(result, newPriorityConstraint(
				domain.PriorityLevel(i+1),
				courseConstraints[courseIndex],
				domain.ParticipantID(prioMapping.participantIndex+1),
			))
		}
	}

	return result
}
