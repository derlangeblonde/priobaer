package apptest

import (
	"slices"
	"strconv"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/util"
)

func TestSolveAssignmentDoesNotOverbookWhenAssignmentsAlreadyExist(t *testing.T) {
	is := is.New(t)
	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)
	courseCount := 2
	maxCapacity := 2
	var courseIds []int
	for range courseCount {
		course := testClient.CoursesCreateAction(model.RandomCourse(model.WithCapacity(0, maxCapacity)), nil)
		courseIds = append(courseIds, course.ID)
	}

	prioLists := [][]int{
		{courseIds[0], courseIds[1]},
		{courseIds[0], courseIds[1]},
		{courseIds[0], courseIds[1]},
		{courseIds[0], courseIds[1]},
	}
	var participantIds []int
	for _, prios := range prioLists {
		participant := testClient.ParticipantsCreateAction(model.RandomParticipant(), prios, nil)
		participantIds = append(participantIds, participant.ID)
	}

	testClient.InitialAssignAction(participantIds[0], courseIds[0])
	testClient.SolveAssignmentsAction()
	courses, unassigned := testClient.AssignmentsIndexAction()

	is.Equal(len(unassigned), 0)
	for _, course := range courses {
		is.Equal(course.Allocation, maxCapacity) // want that courses are allocated to exactly their max capacity
	}
}

func TestSolveAssignmentDontReassignParticipants(t *testing.T) {
	is := is.New(t)
	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	coursesCount := 3
	var courseIds []int
	for range coursesCount {
		course := testClient.CoursesCreateAction(model.RandomCourse(model.WithCapacity(0, 2)), nil)
		courseIds = append(courseIds, course.ID)
	}

	prioLists := [][]int{
		{courseIds[0], courseIds[1], courseIds[2]},
		{courseIds[0], courseIds[1], courseIds[2]},
		{courseIds[0], courseIds[1], courseIds[2]},
		{courseIds[1], courseIds[2], courseIds[0]},
		{courseIds[2], courseIds[0], courseIds[1]},
		{courseIds[1], courseIds[0], courseIds[2]},
	}
	var participantIds []int
	for _, prios := range prioLists {
		participant := testClient.ParticipantsCreateAction(model.RandomParticipant(), prios, nil)
		participantIds = append(participantIds, participant.ID)
	}

	alreadyAssignedParticipantIndex := 0
	alreadyAssignedCourseIndex := 2
	testClient.InitialAssignAction(participantIds[alreadyAssignedParticipantIndex], courseIds[alreadyAssignedCourseIndex])

	testClient.SolveAssignmentsAction()
	_, unassigned := testClient.AssignmentsIndexAction()
	is.Equal(len(unassigned), 0)

	_, participantsAssignedToCourse := testClient.AssignmentsIndexAction("selected-course", strconv.Itoa(courseIds[alreadyAssignedCourseIndex]))
	participantIDsAssignedToCourse := util.IDs(participantsAssignedToCourse)

	is.True(slices.Contains(participantIDsAssignedToCourse, participantIds[alreadyAssignedParticipantIndex]))
}

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
	fixedPriorities := []int{1, 2, 3}
	for range participantCount {
		testClient.ParticipantsCreateAction(model.RandomParticipant(), fixedPriorities, nil)
	}

	expectedAllocations := []int{1, 2, 2}

	testClient.SolveAssignmentsAction()
	coursesSolved, unassignedParticipants := testClient.AssignmentsIndexAction()
	is.Equal(len(unassignedParticipants), 0) // after solving no unassigned participants should be left
	var actualAllocations []int
	for _, course := range coursesSolved {
		actualAllocations = append(actualAllocations, course.Allocation)
	}
	slices.Sort(actualAllocations)

	is.Equal(actualAllocations, expectedAllocations)
}
