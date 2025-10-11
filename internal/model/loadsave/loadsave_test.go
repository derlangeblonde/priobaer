package loadsave

import (
	"bytes"
	"testing"

	"softbaer.dev/ass/internal/domain"

	"github.com/matryer/is"
)

func buildScenario(
	courses []domain.CourseData,
	participants []domain.ParticipantData,
	assignments map[domain.ParticipantID]domain.CourseID,
	priorities map[domain.ParticipantID][]domain.CourseID,
) *domain.Scenario {
	scenario := domain.EmptyScenario()
	for _, course := range courses {
		scenario.AddCourse(course)
	}
	for _, participant := range participants {
		scenario.AddParticipant(participant)
	}
	for pid, cid := range assignments {
		_ = scenario.Assign(pid, cid)
	}
	for pid, cids := range priorities {
		_ = scenario.Prioritize(pid, cids)
	}
	return scenario
}

func TestMarshalModelsIsRoundTripConsistent(t *testing.T) {
	testcases := []struct {
		name              string
		coursesInput      []domain.CourseData
		participantsInput []domain.ParticipantData
		assignments       map[domain.ParticipantID]domain.CourseID
		priorities        map[domain.ParticipantID][]domain.CourseID
	}{
		{
			name: "Two courses and two participants without assignments",
			coursesInput: []domain.CourseData{
				{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
			},
			participantsInput: []domain.ParticipantData{
				{ID: 1, ParticipantName: domain.ParticipantName{Prename: "mady", Surname: "morison"}},
				{ID: 2, ParticipantName: domain.ParticipantName{Prename: "Breathe", Surname: "Flow"}},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{},
			priorities:  map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "No courses but one participant",
			coursesInput:      []domain.CourseData{},
			participantsInput: []domain.ParticipantData{{ID: 143920, ParticipantName: domain.ParticipantName{Prename: "we have", Surname: "no courses"}}},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "One course but no participants",
			coursesInput:      []domain.CourseData{{ID: 42904, Name: "I am", MinCapacity: 741, MaxCapacity: 4920}},
			participantsInput: []domain.ParticipantData{},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "Empty scenario",
			coursesInput:      []domain.CourseData{},
			participantsInput: []domain.ParticipantData{},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "CourseData and participant with special characters",
			coursesInput:      []domain.CourseData{{ID: 23, Name: "\"", MinCapacity: 482, MaxCapacity: 34213}},
			participantsInput: []domain.ParticipantData{{ID: 1, ParticipantName: domain.ParticipantName{Prename: "\\ \"", Surname: "''"}}},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:         "One course with one assignment",
			coursesInput: []domain.CourseData{{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25}},
			participantsInput: []domain.ParticipantData{
				{ID: 1, ParticipantName: domain.ParticipantName{Prename: "nicht", Surname: "zugeteilt"}},
				{ID: 2, ParticipantName: domain.ParticipantName{Prename: "der", Surname: "schon"}},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{
				2: 1,
			},
			priorities: map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name: "Single participant with ordered priorities",
			coursesInput: []domain.CourseData{
				{ID: 1, Name: "Math", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "Physics", MinCapacity: 3, MaxCapacity: 20},
				{ID: 3, Name: "Chemistry", MinCapacity: 4, MaxCapacity: 15},
			},
			participantsInput: []domain.ParticipantData{
				{ID: 1, ParticipantName: domain.ParticipantName{Prename: "John", Surname: "Doe"}},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{},
			priorities: map[domain.ParticipantID][]domain.CourseID{
				1: {2, 3, 1},
			},
		},
		{
			name: "Multiple participants with priorities and one assignment",
			coursesInput: []domain.CourseData{
				{ID: 1, Name: "English", MinCapacity: 5, MaxCapacity: 30},
				{ID: 2, Name: "History", MinCapacity: 3, MaxCapacity: 25},
				{ID: 3, Name: "Art", MinCapacity: 4, MaxCapacity: 20},
				{ID: 4, Name: "Music", MinCapacity: 2, MaxCapacity: 15},
			},
			participantsInput: []domain.ParticipantData{
				{ID: 1, ParticipantName: domain.ParticipantName{Prename: "Alice", Surname: "Smith"}},
				{ID: 2, ParticipantName: domain.ParticipantName{Prename: "Bob", Surname: "Jones"}},
				{ID: 3, ParticipantName: domain.ParticipantName{Prename: "Carol", Surname: "Wilson"}},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{
				1: 3,
			},
			priorities: map[domain.ParticipantID][]domain.CourseID{
				1: {3, 4, 1, 2},
				2: {2, 1, 4, 3},
				3: {4, 3, 2, 1},
			},
		},
		{
			name: "Some participants with priorities and one assignment",
			coursesInput: []domain.CourseData{
				{ID: 1, Name: "Biology", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "Geography", MinCapacity: 4, MaxCapacity: 20},
			},
			participantsInput: []domain.ParticipantData{
				{ID: 1, ParticipantName: domain.ParticipantName{Prename: "David", Surname: "Brown"}},
				{ID: 2, ParticipantName: domain.ParticipantName{Prename: "Eva", Surname: "Green"}},
				{ID: 3, ParticipantName: domain.ParticipantName{Prename: "Frank", Surname: "White"}},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{
				2: 1,
			},
			priorities: map[domain.ParticipantID][]domain.CourseID{
				1: {2, 1},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			scenario := buildScenario(tc.coursesInput, tc.participantsInput, tc.assignments, tc.priorities)

			excelBytes, err := WriteScenarioDataToExcel(scenario)
			is.NoErr(err) // exporting should not error
			is.True(len(excelBytes) > 0)

			imported, err := ParseExcelFile(bytes.NewReader(excelBytes))
			is.NoErr(err) // importing should not error

			var gotCourses []domain.CourseData
			for c := range imported.AllCourses() {
				gotCourses = append(gotCourses, c)
			}
			is.Equal(len(tc.coursesInput), len(gotCourses))
			for i, want := range tc.coursesInput {
				got := gotCourses[i]
				is.Equal(want.ID, got.ID)
				is.Equal(want.Name, got.Name)
				is.Equal(want.MinCapacity, got.MinCapacity)
				is.Equal(want.MaxCapacity, got.MaxCapacity)
			}

			var gotParts []domain.ParticipantData
			for p := range imported.AllParticipants() {
				gotParts = append(gotParts, p)
			}

			is.Equal(len(tc.participantsInput), len(gotParts))
			for i, want := range tc.participantsInput {
				got := gotParts[i]
				is.Equal(want.ID, got.ID)
				is.Equal(want.Prename, got.Prename)
				is.Equal(want.Surname, got.Surname)
			}

			for pid, wantCID := range tc.assignments {
				c, ok := imported.AssignedCourse(pid)
				is.True(ok)
				is.Equal(wantCID, c.ID)
			}

			for _, p := range tc.participantsInput {
				if _, exists := tc.assignments[p.ID]; !exists {
					_, ok := imported.AssignedCourse(p.ID)
					is.True(!ok)
				}
			}

			for pid, wantPrios := range tc.priorities {
				var gotPrios []domain.CourseID
				for c := range imported.PrioritizedCoursesOrdered(pid) {
					gotPrios = append(gotPrios, c.ID)
				}

				is.Equal(len(wantPrios), len(gotPrios)) // Same number of priorities

				for i, wantCID := range wantPrios {
					if i < len(gotPrios) {
						is.Equal(wantCID, gotPrios[i])
					}
				}
			}

			for _, p := range tc.participantsInput {
				if _, hasPrios := tc.priorities[p.ID]; !hasPrios {
					var gotPrios []domain.CourseID
					for c := range imported.PrioritizedCoursesOrdered(p.ID) {
						gotPrios = append(gotPrios, c.ID)
					}
					is.Equal(0, len(gotPrios)) // Should have no priorities
				}
			}
		})
	}
}
