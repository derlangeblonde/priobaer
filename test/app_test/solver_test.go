package apptest

import (
	"slices"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/infra"
)

func TestSolveAssignment(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	courses := []infra.Course{
		infra.RandomCourse(infra.WithCapacity(0, 2)),
		infra.RandomCourse(infra.WithCapacity(0, 2)),
		infra.RandomCourse(infra.WithCapacity(0, 2)),
	}

	for _, course := range courses {
		testClient.CoursesCreateAction(course, nil)
	}

	participantCount := 5
	for i := 0; i < participantCount; i++ {
		testClient.ParticipantsCreateAction(infra.RandomParticipant(), make([]int, 0), nil)
	}

	expectedAllocations := []int{1, 2, 2}

	coursesSolved, unassignedParticipants := testClient.AssignmentsIndexAction("solve", "true")
	is.Equal(len(unassignedParticipants), 0) // after solving no unassigned participants should be left
	var actualAllocations []int
	for _, course := range coursesSolved {
		actualAllocations = append(actualAllocations, course.Allocation)
	}
	slices.Sort(actualAllocations)

	is.Equal(actualAllocations, expectedAllocations)
}
