package cmdtest

import (
	"slices"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/util"
)

func TestParticipantsAreUnassignedIntially(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost)

	expectedParticipant := RandomParticipant()

	testClient.ParticipantsCreateAction(expectedParticipant, nil)

	unassignedParticipants := testClient.AssignmentsIndexAction(util.NoneInt())

	is.Equal(len(unassignedParticipants), 1) // expect exactly one participant after creating one

	is.Equal(unassignedParticipants[0].Prename, expectedParticipant.Prename)
	is.Equal(unassignedParticipants[0].Surname, expectedParticipant.Surname)

}

func TestAssignParticipant(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost)

	expectedParticipant := RandomParticipant()
	expectedCourse := RandomCourse()

	testClient.ParticipantsCreateAction(expectedParticipant, nil)
	testClient.CoursesCreateAction(expectedCourse, nil)

	unassignedParticipants := testClient.AssignmentsIndexAction(util.NoneInt())
	allCourses := testClient.CoursesIndexAction()

	is.Equal(len(unassignedParticipants), 1) // expect exactly one participant after creating one
	is.Equal(len(allCourses), 1)             // expect exactly one course after creating one

	idParticipantToAssign := unassignedParticipants[0].ID
	courseIdToAssignTo := allCourses[0].ID

	testClient.AssignmentsUpdateAction(idParticipantToAssign, util.JustInt(courseIdToAssignTo))

	unassignedParticipants = testClient.AssignmentsIndexAction(util.NoneInt())

	is.Equal(len(unassignedParticipants), 0) // expect exactly no unassigned participant after assigning the only one

	participantsAssignedToCourse := testClient.AssignmentsIndexAction(util.JustInt(courseIdToAssignTo))

	is.Equal(len(participantsAssignedToCourse), 1) // expect exactly one participant after creating one

	testClient.AssignmentsUpdateAction(idParticipantToAssign, util.NoneInt())

	unassignedParticipants = testClient.AssignmentsIndexAction(util.NoneInt())

	is.Equal(len(unassignedParticipants), 1) // expect exactly one unassigned participant after unassigne participant again

	participantsAssignedToCourse = testClient.AssignmentsIndexAction(util.JustInt(courseIdToAssignTo))

	is.Equal(len(participantsAssignedToCourse), 0) // expect no participant assinged to course after unassigning
}

func TestDisplayCourseAllocation(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	expectedAllocations := []int{4, 2, 5, 11, 5}

	testClient.CreateCoursesWithAllocationsAction(expectedAllocations)

	actualCourses := testClient.CoursesIndexAction()

	is.Equal(len(actualCourses), len(expectedAllocations))

	var actualAllocations []int

	for _, actualCourse := range actualCourses {
		actualAllocations = append(actualAllocations, actualCourse.Allocation)
	}

	slices.Sort(actualAllocations)
	slices.Sort(expectedAllocations)

	is.Equal(actualAllocations, expectedAllocations)
}

func TestUpdateAssignmentUpdatesCourseAllocations(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	courseOld := testClient.CoursesCreateAction(RandomCourse(), nil)
	courseNew := testClient.CoursesCreateAction(RandomCourse(), nil)
	participant := testClient.ParticipantsCreateAction(RandomParticipant(), nil)

	testClient.AssignmentsUpdateAction(participant.ID, util.JustInt(courseOld.ID))

	// act
	viewUpdate := testClient.AssignmentsUpdateAction(participant.ID, util.JustInt(courseNew.ID))

	// assert
	is.Equal(len(viewUpdate.courses), 2) // expect exactly to courses to have updated allocation

	courseOldPresent, courseNewPresent := false, false

	for _, courseUpdated := range viewUpdate.courses{
		if courseUpdated.ID == courseOld.ID {
			courseOldPresent = true
			is.Equal(courseUpdated.Allocation, 0) // expect old course to have no one assignment after update
		}

		if courseUpdated.ID == courseNew.ID {
			courseNewPresent = true
			is.Equal(courseUpdated.Allocation, 1) // expect new course to have one participant after update
		}
	}

	is.True(courseOldPresent) // expect courseOld to be present in view-update
	is.True(courseNewPresent) // expect courseNew to be present in view-update
}

func TestAssignmentUpdateWithMultipleParticipantsUpdatesViewCorrectly(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	initialAllocations := []int{2, 2}
	assignmentMap := testClient.CreateCoursesWithAllocationsAction(initialAllocations)

	newCourseId, participantId, loopCounter := 0, 0, 0
	for courseId, participantsId := range assignmentMap {
		if loopCounter == 0 {
			participantId = participantsId[0]
		} else {
			newCourseId = courseId
		}

		loopCounter++
	}

	viewUpdate := testClient.AssignmentsUpdateAction(participantId, util.JustInt(newCourseId))

	var updatedAllocations []int
	for _, courseUpdated := range viewUpdate.courses {
		updatedAllocations = append(updatedAllocations, courseUpdated.Allocation)
	}

	slices.Sort(updatedAllocations)

	is.Equal(updatedAllocations, []int{1, 3})
}

func TestAssignmentUpdateInitialAssignUpdatesUnassignedCount(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	var participant model.Participant
	for i := 0; i < 3; i ++ {
		participant = testClient.ParticipantsCreateAction(RandomParticipant(), nil)
	}

	course := testClient.CoursesCreateAction(RandomCourse(), nil)

	// act
	viewUpdate := testClient.AssignmentsUpdateAction(participant.ID, util.JustInt(course.ID))

	// assert
	is.True(viewUpdate.UnassignedCount.Updated) // expect that unassigned count was updated
	is.Equal(viewUpdate.UnassignedCount.Value, 2)
}

func TestAssignmentUpdateUnassignUpdatesUnassignedCount(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	course := testClient.CoursesCreateAction(RandomCourse(), nil)

	var participant model.Participant
	for i := 0; i < 3; i ++ {
		participant = testClient.ParticipantsCreateAction(RandomParticipant(), nil)
		testClient.AssignmentsUpdateAction(participant.ID, util.JustInt(course.ID))
	}

	// act
	viewUpdate := testClient.AssignmentsUpdateAction(participant.ID, util.NoneInt())

	// assert
	is.True(viewUpdate.UnassignedCount.Updated) // expect that unassigned count was updated
	is.Equal(viewUpdate.UnassignedCount.Value, 1)
}
