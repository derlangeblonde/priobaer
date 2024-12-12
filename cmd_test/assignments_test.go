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
	unassignedParticipants[0].ID = 0
	is.Equal(unassignedParticipants[0], expectedParticipant)
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

	expectedParticipant.ID = unassignedParticipants[0].ID

	testClient.AssignmentsUpdateAction(int(expectedParticipant.ID), util.JustInt(int(expectedCourse.ID)))
}
