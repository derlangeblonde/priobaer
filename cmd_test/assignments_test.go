package cmdtest

import (
	"testing"
)

func TestParticpantsAreUnassignedIntially(t *testing.T) {
	// is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	testClient := NewTestClient(t, localhost8080)

	participant := RandomParticipant()

	testClient.ParticpantCreateAction(participant, nil)

	_ = testClient.AssignmentsIndexAction()

	// is.Equal(len(actualParticipants), 1)	
}
