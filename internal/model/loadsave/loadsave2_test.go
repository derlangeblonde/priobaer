package loadsave

import (
	"bytes"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/model"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/xuri/excelize/v2"
)

func buildScenario(
	courses []domain.Course,
	participants []domain.Participant,
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

func TestMarshalModelsIsRoundTripConsistent2(t *testing.T) {
	testcases := []struct {
		name              string
		coursesInput      []domain.Course
		participantsInput []domain.Participant
		assignments       map[domain.ParticipantID]domain.CourseID
		priorities        map[domain.ParticipantID][]domain.CourseID
	}{
		{
			name: "Two courses and two participants without assignments",
			coursesInput: []domain.Course{
				{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
			},
			participantsInput: []domain.Participant{
				{ID: 1, Prename: "mady", Surname: "morison"},
				{ID: 2, Prename: "Breathe", Surname: "Flow"},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{},
			priorities:  map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "No courses but one participant",
			coursesInput:      []domain.Course{},
			participantsInput: []domain.Participant{{ID: 143920, Prename: "we have", Surname: "no courses"}},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "One course but no participants",
			coursesInput:      []domain.Course{{ID: 42904, Name: "I am", MinCapacity: 741, MaxCapacity: 4920}},
			participantsInput: []domain.Participant{},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "Empty scenario",
			coursesInput:      []domain.Course{},
			participantsInput: []domain.Participant{},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:              "Course and participant with special characters",
			coursesInput:      []domain.Course{{ID: 23, Name: "\"", MinCapacity: 482, MaxCapacity: 34213}},
			participantsInput: []domain.Participant{{ID: 1, Prename: "\\ \"", Surname: "''"}},
			assignments:       map[domain.ParticipantID]domain.CourseID{},
			priorities:        map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name:         "One course with one assignment",
			coursesInput: []domain.Course{{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25}},
			participantsInput: []domain.Participant{
				{ID: 1, Prename: "nicht", Surname: "zugeteilt"},
				{ID: 2, Prename: "der", Surname: "schon"},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{
				2: 1,
			},
			priorities: map[domain.ParticipantID][]domain.CourseID{},
		},
		{
			name: "Single participant with ordered priorities",
			coursesInput: []domain.Course{
				{ID: 1, Name: "Math", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "Physics", MinCapacity: 3, MaxCapacity: 20},
				{ID: 3, Name: "Chemistry", MinCapacity: 4, MaxCapacity: 15},
			},
			participantsInput: []domain.Participant{
				{ID: 1, Prename: "John", Surname: "Doe"},
			},
			assignments: map[domain.ParticipantID]domain.CourseID{},
			priorities: map[domain.ParticipantID][]domain.CourseID{
				1: {2, 3, 1},
			},
		},
		{
			name: "Multiple participants with priorities and one assignment",
			coursesInput: []domain.Course{
				{ID: 1, Name: "English", MinCapacity: 5, MaxCapacity: 30},
				{ID: 2, Name: "History", MinCapacity: 3, MaxCapacity: 25},
				{ID: 3, Name: "Art", MinCapacity: 4, MaxCapacity: 20},
				{ID: 4, Name: "Music", MinCapacity: 2, MaxCapacity: 15},
			},
			participantsInput: []domain.Participant{
				{ID: 1, Prename: "Alice", Surname: "Smith"},
				{ID: 2, Prename: "Bob", Surname: "Jones"},
				{ID: 3, Prename: "Carol", Surname: "Wilson"},
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
			coursesInput: []domain.Course{
				{ID: 1, Name: "Biology", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "Geography", MinCapacity: 4, MaxCapacity: 20},
			},
			participantsInput: []domain.Participant{
				{ID: 1, Prename: "David", Surname: "Brown"},
				{ID: 2, Prename: "Eva", Surname: "Green"},
				{ID: 3, Prename: "Frank", Surname: "White"},
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

			excelBytes, err := Export(scenario)
			is.NoErr(err) // exporting should not error
			is.True(len(excelBytes) > 0)

			imported, err := Import(bytes.NewReader(excelBytes))
			is.NoErr(err) // importing should not error

			var gotCourses []domain.Course
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

			var gotParts []domain.Participant
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

func TestUnmarshalInvalidExcelFileReturnsSpecificError2(t *testing.T) {
	testcases := []struct {
		wantErrorMsgKeywords []string
		excelBytes           []byte
		name                 string
	}{
		{[]string{"Spalte", "ID", "valide"}, scenarioOnlyStringValuesInParticipantsSheet2(t), "OnlyStringValuesInParticipantsSheet"},
		{[]string{"Teilnehmer", "Kopfzeile", "Vorname"}, scenarioInvalidHeaderParticipantsSheet2(t), "InvalidHeaderParticipantsSheet"},
		{[]string{"Kurse", "Kopfzeile", "Name"}, scenarioInvalidHeaderCourseSheet2(t), "InvalidHeaderCourseSheet"},
		{[]string{"Teilnehmer", "Zeile", "Werte"}, scenarioInvalidRowLengthInParticipantsSheet2(t), "InvalidRowLengthInParticipantsSheet"},
		{[]string{"Kurse", "Zeile", "Werte"}, scenarioInvalidRowLengthInCoursesSheet2(t), "InvalidRowLengthInCoursesSheet"},
		{[]string{"Kurse", "maximal", "minimale", "Kapazität", "größer"}, scenarioMaxCapacitySmallerThanMinCapacity2(t), "MaxCapacitySmallerThanMinCapacity"},
		{[]string{"Nachname", "nicht", "leer"}, scenarioSurnameEmpty2(t), "SurnameEmpty"},
		{[]string{"Kurs", "existiert", "nicht"}, scenarioAssignmentToNonExistentCourse2(t), "AssignmentToNonExistentCourse"},
	}

	for _, tc := range testcases {
		_, err := Import(bytes.NewReader(tc.excelBytes))
		if err == nil {
			t.Fatal("Want err (because we tried to Import an invalid file), but got nil")
		}
		for _, wantKeyword := range tc.wantErrorMsgKeywords {
			if !strings.Contains(err.Error(), wantKeyword) {
				t.Fatalf("'%s': Want keyword '%s' in error, Got: %s", tc.name, wantKeyword, err.Error())
			}
		}
	}
}

func scenarioAssignmentToNonExistentCourse2(t *testing.T) []byte {
	return buildExcelFile2(
		t,
		[][]string{
			{"1", "foo", "5", "25"},
			{"2", "bar", "25", "30"},
			{"3", "baz", "10", "20"},
		},
		[][]string{
			{"1", "foo", "bar", "2"},
			{"2", "bar", "baz", "3"},
			{"3", "baz", "foo", "1"},
			{"4", "qux", "quux", "5"},
		},
		true,
		true,
	)
}

func scenarioSurnameEmpty2(t *testing.T) []byte {
	participantWithEmptySurname := []string{"1", "foo", "", "null"}
	return buildExcelFile2(t, [][]string{{}}, [][]string{participantWithEmptySurname}, true, true)
}

func scenarioMaxCapacitySmallerThanMinCapacity2(t *testing.T) []byte {
	courseWithInvalidCapacity := []string{"1", "foo", "25", "5"}
	return buildExcelFile2(t, [][]string{courseWithInvalidCapacity}, [][]string{}, true, true)
}

func scenarioInvalidRowLengthInCoursesSheet2(t *testing.T) []byte {
	return buildExcelFile2(t, [][]string{{"1", "foo", "bar", "baz", "qux", "more", "than", "expected", "values"}}, [][]string{}, true, true)
}

func scenarioInvalidRowLengthInParticipantsSheet2(t *testing.T) []byte {
	invalidRowLengthParticipant := []string{"1", "foo", "bar"}
	return buildExcelFile2(t, [][]string{{}}, [][]string{invalidRowLengthParticipant}, true, true)
}

func scenarioInvalidHeaderParticipantsSheet2(t *testing.T) []byte {
	invalidHeaderParticipant := []string{"das", "ist", "kein", "header"}
	return buildExcelFile2(t, [][]string{{}}, [][]string{invalidHeaderParticipant}, true, false)
}

func scenarioInvalidHeaderCourseSheet2(t *testing.T) []byte {
	invalidHeaderCourse := []string{"das", "ist", "kein", "header"}
	return buildExcelFile2(t, [][]string{invalidHeaderCourse}, [][]string{}, false, true)
}

func scenarioOnlyStringValuesInParticipantsSheet2(t *testing.T) []byte {
	onlyStringParticipant := []string{"id", "foo", "bar", "baz"}
	return buildExcelFile2(t, [][]string{{}}, [][]string{onlyStringParticipant}, true, true)
}

func buildParticipantSheet2(t *testing.T, excelFile *excelize.File, participants [][]string, writeHeader bool) {
	is := is.New(t)

	sheetWriter, err := newSheetWriter(excelFile, participantsSheetName)
	is.NoErr(err)

	if writeHeader {
		is.NoErr(sheetWriter.write(model.Participant{}.RecordHeader()))
	}

	for _, participant := range participants {
		is.NoErr(sheetWriter.write(participant))
	}
}

func buildCourseSheet2(t *testing.T, excelFile *excelize.File, courses [][]string, writeHeader bool) {
	is := is.New(t)

	sheetWriter, err := newSheetWriter(excelFile, courseSheetName)
	is.NoErr(err)

	if writeHeader {
		is.NoErr(sheetWriter.write(domain.Course{}.RecordHeader()))
	}

	for _, course := range courses {
		is.NoErr(sheetWriter.write(course))
	}
}

func buildExcelFile2(t *testing.T, courses, participants [][]string, writeCourseHeader, writeParticipantHeader bool) []byte {
	is := is.New(t)

	excelFile := excelize.NewFile()

	buildCourseSheet2(t, excelFile, courses, writeCourseHeader)
	buildParticipantSheet2(t, excelFile, participants, writeParticipantHeader)

	var buf bytes.Buffer
	err := excelFile.Write(&buf)
	is.NoErr(err)

	return buf.Bytes()
}
