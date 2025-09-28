package apptest

import (
	"github.com/matryer/is"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/ui"
	"softbaer.dev/ass/internal/util"
	"strconv"
	"testing"
)

func TestCreateAndReadParticpantWithPrios(t *testing.T) {
	is := is.New(t)
	tcs := []struct {
		nPrioritizedCourses   int
		nOtherCourses         int
		nBackgroundCharacters int
	}{
		{1, 1, 4},
		{3, 4, 6},
	}

	for _, tc := range tcs {
		func() {
			sut := StartupSystemUnderTest(t, nil)
			defer waitForTerminationDefault(sut.cancel)

			ctx := NewTestClient(t, localhost)

			wantParticipant := model.RandomParticipant()

			var wantCourses []ui.Course
			var wantPrioritizedCourseIds []int
			for i := 0; i < tc.nPrioritizedCourses; i++ {
				wantCourses = append(wantCourses, ctx.CoursesCreateAction(model.RandomCourse(
					model.WithCourseName(strconv.Itoa(i+1)),
				), nil))
				wantPrioritizedCourseIds = append(wantPrioritizedCourseIds, wantCourses[i].ID)
			}
			for i := 0; i < tc.nOtherCourses; i++ {
				ctx.CoursesCreateAction(model.RandomCourse(), nil)
			}

			ctx.ParticipantsCreateAction(wantParticipant, wantPrioritizedCourseIds, nil)

			for i := 0; i < tc.nBackgroundCharacters; i++ {
				ctx.ParticipantsCreateAction(model.RandomParticipant(), make([]int, 0), nil)
			}

			_, renderedParticipants := ctx.AssignmentsIndexAction()

			is.Equal(len(renderedParticipants), tc.nBackgroundCharacters+1)

			var gotParticipant ui.Participant
			for _, renderedParticipant := range renderedParticipants {
				if renderedParticipant.Surname == wantParticipant.Surname {
					gotParticipant = renderedParticipant
					break
				}
			}

			is.Equal(gotParticipant.Prename, wantParticipant.Prename)        // created and retrieved participant should be the same.
			is.Equal(gotParticipant.Surname, wantParticipant.Surname)        // created and retrieved participant should be the same.
			is.Equal(len(gotParticipant.Priorities), tc.nPrioritizedCourses) // want as many priorities as were created

			for i := 0; i < tc.nPrioritizedCourses; i++ {
				is.Equal(gotParticipant.Priorities[i].CourseName, wantCourses[i].Name)
			}

		}()
	}

}

func TestCreateParticipantWithAssignmentAndPriosCanBeDeleted(t *testing.T) {
	is := is.New(t)
	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	nPrioritizedCourses := 3
	nOtherCourses := 4

	client := NewTestClient(t, localhost)

	var wantCourses []ui.Course
	var wantPrioritizedCourseIds []int
	for i := 0; i < nPrioritizedCourses; i++ {
		wantCourses = append(wantCourses, client.CoursesCreateAction(model.RandomCourse(
			model.WithCourseName(strconv.Itoa(i+1)),
		), nil))
		wantPrioritizedCourseIds = append(wantPrioritizedCourseIds, wantCourses[i].ID)
	}
	for i := 0; i < nOtherCourses; i++ {
		client.CoursesCreateAction(model.RandomCourse(), nil)
	}

	wantParticipant := client.ParticipantsCreateAction(model.RandomParticipant(), wantPrioritizedCourseIds, nil)

	client.AssignmentsUpdateAction(wantParticipant.ID, util.JustInt(wantCourses[0].ID))

	_, renderedParticipants := client.AssignmentsIndexAction("selected-course", strconv.Itoa(wantCourses[0].ID))

	is.Equal(len(renderedParticipants), 1)

	client.ParticipantsDeleteAction(wantParticipant.ID)

	_, renderedParticipants = client.AssignmentsIndexAction("selected-course", strconv.Itoa(wantCourses[0].ID))
	is.Equal(len(renderedParticipants), 0)

	_, renderedParticipants = client.AssignmentsIndexAction()
	is.Equal(len(renderedParticipants), 0)
}
