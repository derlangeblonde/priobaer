package apptest

import (
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/ui"
)

func TestScenarioRemainsValidAfterCourseWithAssignmentsAndPriosWasDeleted(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer sut.cancel()

	testClient := NewTestClient(t, localhost)

	initialCourses := []ui.Course{
		ui.RandomCourse(ui.WithCapacity(0, 2)),
		ui.RandomCourse(ui.WithCapacity(0, 2)),
		ui.RandomCourse(ui.WithCapacity(0, 2)),
	}

	var createdCourses []ui.Course
	for _, course := range initialCourses {
		createdCourse := testClient.CoursesCreateAction(course, nil)
		createdCourses = append(createdCourses, createdCourse)
	}
	courseToDelete := createdCourses[0]

	participantCount := 5
	fixedPriorities := []int{1, 2, 3}
	var createdParticipants []ui.Participant
	for range participantCount {
		participant := testClient.ParticipantsCreateAction(ui.RandomParticipant(), fixedPriorities, nil)
		createdParticipants = append(createdParticipants, participant)
		testClient.InitialAssignAction(participant.ID, courseToDelete.ID)
	}

	testClient.CoursesDeleteAction(courseToDelete.ID)

	coursesAfterDelete, participantsAfterDelete := testClient.AssignmentsIndexAction()
	is.Equal(len(participantsAfterDelete), participantCount) // want all participants to be unassigned, after assigned course was deleted
	is.Equal(len(coursesAfterDelete), len(initialCourses)-1) // want one course less, since we deleted one

	for _, participant := range participantsAfterDelete {
		is.Equal(len(participant.Priorities), 2) // initially all participants had 3 prios after delete we want 2
	}
}
