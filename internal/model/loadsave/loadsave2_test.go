package loadsave

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/xuri/excelize/v2"
	"softbaer.dev/ass/internal/model"
)

func TestMarshalModelsIsRoundTripConsistent2(t *testing.T) {
	is := is.New(t)

	testcases := []struct {
		coursesInput      []model.Course
		participantsInput []model.Participant
	}{
		{
			[]model.Course{
				{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
			},
			[]model.Participant{
				{ID: 1, Prename: "mady", Surname: "morison"},
				{ID: 2, Prename: "Breathe", Surname: "Flow"},
			},
		},
		{
			[]model.Course{},
			[]model.Participant{{ID: 143920, Prename: "we have", Surname: "no courses"}},
		},
		{
			[]model.Course{{ID: 42904, Name: "I am", MinCapacity: 741, MaxCapacity: 4920}},
			[]model.Participant{},
		},
		{
			[]model.Course{},
			[]model.Participant{},
		},
		{
			[]model.Course{{ID: 23, Name: "\"", MinCapacity: 482, MaxCapacity: 34213}},
			[]model.Participant{{ID: 1, Prename: "\\ \"", Surname: "''"}},
		},
		{
			[]model.Course{{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25}},
			[]model.Participant{
				{ID: 1, Prename: "nicht", Surname: "zugeteilt"},
				{ID: 2, Prename: "der", Surname: "schon", CourseID: sql.NullInt64{Valid: true, Int64: 1}},
			},
		},
	}

	for _, tc := range testcases {
		scenario := EmptyScenario()
		for _, mc := range tc.coursesInput {
			scenario.AddCourse(Course{
				ID:          CourseID(mc.ID),
				Name:        mc.Name,
				MinCapacity: mc.MinCapacity,
				MaxCapacity: mc.MaxCapacity,
			})
		}
		for _, mp := range tc.participantsInput {
			scenario.AddParticipant(Participant{
				ID:      ParticipantID(mp.ID),
				Prename: mp.Prename,
				Surname: mp.Surname,
			})
			if mp.CourseID.Valid {
				_ = scenario.Assign(ParticipantID(mp.ID), CourseID(mp.CourseID.Int64))
			}
		}

		excelBytes, err := Export(scenario)
		is.NoErr(err) // exporting should not error
		is.True(len(excelBytes) > 0)

		imported, err := Import(bytes.NewReader(excelBytes))
		is.NoErr(err) // importing should not error

		var gotCourses []Course
		for c := range imported.AllCourses() {
			gotCourses = append(gotCourses, c)
		}
		is.Equal(len(tc.coursesInput), len(gotCourses))
		for i, want := range tc.coursesInput {
			got := gotCourses[i]
			is.Equal(CourseID(want.ID), got.ID)
			is.Equal(want.Name, got.Name)
			is.Equal(want.MinCapacity, got.MinCapacity)
			is.Equal(want.MaxCapacity, got.MaxCapacity)
		}

		// Compare participants
		var gotParts []Participant
		for p := range imported.AllParticipants() {
			gotParts = append(gotParts, p)
		}
		is.Equal(len(tc.participantsInput), len(gotParts))
		for i, want := range tc.participantsInput {
			got := gotParts[i]
			is.Equal(ParticipantID(want.ID), got.ID)
			is.Equal(want.Prename, got.Prename)
			is.Equal(want.Surname, got.Surname)

			if want.CourseID.Valid {
				c, ok := imported.AssignedCourse(ParticipantID(want.ID))
				is.True(ok)
				is.Equal(CourseID(want.CourseID.Int64), c.ID)
			} else {
				_, ok := imported.AssignedCourse(ParticipantID(want.ID))
				is.True(!ok)
			}
		}
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
		is.NoErr(sheetWriter.write(model.Course{}.RecordHeader()))
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

