package apptest

import (
	"slices"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/model"
)

func TestSolveAssignment(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	courses := []model.Course{
		model.RandomCourse(model.WithCapacity(0, 2)),
		model.RandomCourse(model.WithCapacity(0, 2)),
		model.RandomCourse(model.WithCapacity(0, 2)),
	}

	for _, course := range courses {
		testClient.CoursesCreateAction(course, nil)
	}

	participantCount := 5
	for i := 0; i < participantCount; i++ {
		testClient.ParticipantsCreateAction(model.RandomParticipant(), nil)
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
