package cmdtest

import (
	"slices"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/util"
)

func TestParticpantsAreUnassignedIntially(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost)

	testClient.AcquireSessionCookie()

	expectedParticipant := RandomParticipant()

	testClient.ParticpantsCreateAction(expectedParticipant, nil)

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

	testClient.AcquireSessionCookie()

	expectedParticipant := RandomParticipant()
	expectedCourse := RandomCourse()

	testClient.ParticpantsCreateAction(expectedParticipant, nil)
	testClient.CoursesCreateAction(expectedCourse, nil)

	unassignedParticipants := testClient.AssignmentsIndexAction(util.NoneInt())
	allCourses := testClient.CoursesIndexAction()

	is.Equal(len(unassignedParticipants), 1) // expect exactly one participant after creating one 123
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
	testClient.AcquireSessionCookie()

	expectedAllocations := []int{4, 2, 5, 11, 5}

	for _, expectedAlloc := range expectedAllocations {
		course := testClient.CoursesCreateAction(RandomCourse(), nil)
		for i := 0; i < expectedAlloc; i++ {
			participant := testClient.ParticpantsCreateAction(RandomParticipant(), nil)
			testClient.AssignmentsUpdateAction(participant.ID, util.JustInt(course.ID))
		}
	}

	actualCourses := testClient.CoursesIndexAction()

	is.Equal(len(actualCourses), len(expectedAllocations))

	var actualAllocations []int

	for _, actualCourse := range actualCourses{
		actualAllocations = append(actualAllocations, actualCourse.Allocation)
	}

	slices.Sort(actualAllocations)
	slices.Sort(expectedAllocations)

	is.Equal(actualAllocations, expectedAllocations)
}
