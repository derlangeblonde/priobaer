package cmdtest

import (
	"testing"

	"github.com/matryer/is"
)

func TestParticpantsAreUnassignedIntially(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost8080)

	testClient.AcquireSessionCookie()

	participant := RandomParticipant()

	testClient.ParticpantsCreateAction(participant, nil)

	actualParticipants := testClient.AssignmentsIndexAction()

	is.Equal(len(actualParticipants), 1) // expect exactly one participant after creating one
}
