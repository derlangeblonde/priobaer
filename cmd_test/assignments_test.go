package cmdtest

import (
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/util"
)

func TestParticpantsAreUnassignedIntially(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost8080)

	testClient.AcquireSessionCookie()

	expectedParticipant := RandomParticipant()

	testClient.ParticpantsCreateAction(expectedParticipant, nil)

	unassignedParticipants := testClient.AssignmentsIndexAction()

	is.Equal(len(unassignedParticipants), 1) // expect exactly one participant after creating one

	is.Equal(unassignedParticipants[0].Prename, expectedParticipant.Prename)
	is.Equal(unassignedParticipants[0].Surname, expectedParticipant.Surname)
}

func TestAssignParticipant(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost8080)

	testClient.AcquireSessionCookie()

	expectedParticipant := RandomParticipant()
	expectedCourse := RandomCourse()

	testClient.ParticpantsCreateAction(expectedParticipant, nil)
	testClient.CoursesCreateAction(expectedCourse, nil)

	unassignedParticipants := testClient.AssignmentsIndexAction()

	is.Equal(len(unassignedParticipants), 1) // expect exactly one participant after creating one

	idParticipantToAssign := unassignedParticipants[0].ID

	testClient.AssignmentsUpdateAction(idParticipantToAssign, util.JustInt(int(expectedCourse.ID)))

	unassignedParticipants = testClient.AssignmentsIndexAction()

	is.Equal(len(unassignedParticipants), 0) // expect exactly one participant after creating one
}
