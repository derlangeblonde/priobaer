package model

import (
	"bytes"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/xuri/excelize/v2"
)

func TestMarshalCourseIsRoundTripConsistent(t *testing.T) {
	is := is.New(t)

	testcases := []struct {
		coursesInput      []Course
		participantsInput []Participant
	}{
		{
			[]Course{
				{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
			},
			[]Participant{
				{ID: 1, Prename: "mady", Surname: "morison"},
				{ID: 2, Prename: "Breathe", Surname: "Flow"},
			},
		},
		{
			[]Course{},
			[]Participant{{ID: 143920, Prename: "we have", Surname: "no courses"}},
		},
		{
			[]Course{{ID: 42904, Name: "I am", MinCapacity: 741, MaxCapacity: 4920}},
			[]Participant{},
		},
		{
			[]Course{},
			[]Participant{},
		},
		{
			[]Course{{ID: 23, Name: "\"", MinCapacity: 482, MaxCapacity: 34213}},
			[]Participant{{ID: 1, Prename: "\\ \"", Surname: "''"}},
		},
		{
			[]Course{{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25}},
			[]Participant{{ID: 1, Prename: "nicht", Surname: "zugeteilt"}, {ID: 2, Prename: "der", Surname: "schon", CourseID: sql.NullInt64{Valid: true, Int64: 1}}},
		},
	}

	for _, tc := range testcases {
		excelBytes, err := ToExcelBytes(tc.coursesInput, tc.participantsInput)
		is.NoErr(err) // err while marshall to excel

		coursesOutput, participantsOutput, err := FromExcelBytes(bytes.NewReader(excelBytes))
		is.NoErr(err) // err while unmarshalling from excel

		is.Equal(len(tc.coursesInput), len(coursesOutput)) // count of courses same after marshal-roundtrip

		for i := 0; i < len(tc.coursesInput); i++ {
			is.True(reflect.DeepEqual(tc.coursesInput[i], coursesOutput[i]))
		}

		is.Equal(len(tc.participantsInput), len(participantsOutput)) // count of participants same after marshal-roundtrip

		for i := 0; i < len(tc.participantsInput); i++ {
			if !reflect.DeepEqual(tc.participantsInput[i], participantsOutput[i]) {
				t.Fatalf("Participant not equal. Got=%v, Want=%v", participantsOutput[i], tc.participantsInput[i])
			}
		}
	}
}

func TestUnmarshalInvalidExcelFileReturnsSpecificError(t *testing.T) {
	testcases := []struct {
		wantErrorMsgKeywords []string
		excelBytes           []byte
	}{
		{[]string{"Spalte", "ID", "valide"}, scenarioOnlyStringValuesInParticipantsSheet(t)},
		{[]string{"Teilnehmer", "Kopfzeile", "Vorname"}, scenarioInvalidHeaderParticipantsSheet(t)},
		{[]string{"Kurse", "Kopfzeile", "Name"}, scenarioInvalidHeaderCourseSheet(t)},
		{[]string{"Teilnehmer", "Zeile", "Werte"}, scenarioInvalidRowLengthInParticipantsSheet(t)},
		{[]string{"Kurse", "Zeile", "Werte"}, scenarioInvalidRowLengthInCoursesSheet(t)},
		{[]string{"Kurse", "maximal", "minimale", "Kapazität", "größer"}, scenarioMaxCapacitySmallerThanMinCapacity(t)},
		{[]string{"Nachname", "nicht", "leer"}, scenarioSurnameEmpty(t)},
		{[]string{"Kurs", "existiert", "nicht"}, scenarioAssignmentToExistentCourse(t)},
	}

	for _, tc := range testcases {
		_, _, err := FromExcelBytes(bytes.NewReader(tc.excelBytes))
		if err == nil {
			t.Fatal("Want err (because we tried to Unmarshal an invalid file), but got nil")
		}

		for _, wantKeyword := range tc.wantErrorMsgKeywords {
			if !strings.Contains(err.Error(), wantKeyword) {
				t.Fatalf("Want: '%s' (to be contained), Got: %s", wantKeyword, err.Error())
			}
		}
	}

}

func scenarioAssignmentToExistentCourse(t *testing.T) []byte {
	return buildExcelFile(
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

func scenarioSurnameEmpty(t *testing.T) []byte {
	participantWithEmptySurname := []string{"1", "foo", "", "null"}

	return buildExcelFile(t, [][]string{{}}, [][]string{participantWithEmptySurname}, true, true)
}

func scenarioMaxCapacitySmallerThanMinCapacity(t *testing.T) []byte {
	courseWithInvalidCapacity := []string{"1", "foo", "25", "5"}

	return buildExcelFile(t, [][]string{courseWithInvalidCapacity}, [][]string{}, true, true)
}

func scenarioInvalidRowLengthInCoursesSheet(t *testing.T) []byte {
	return buildExcelFile(t, [][]string{{"1", "foo", "bar", "baz", "qux", "more", "than", "expected", "values"}}, [][]string{}, true, true)
}

func scenarioInvalidRowLengthInParticipantsSheet(t *testing.T) []byte {
	invalidRowLengthParticipant := []string{"1", "foo", "bar"}
	return buildExcelFile(t, [][]string{{}}, [][]string{invalidRowLengthParticipant}, true, true)
}

func scenarioInvalidHeaderParticipantsSheet(t *testing.T) []byte {
	invalidHeaderParticipant := []string{"das", "ist", "kein", "header"}

	return buildExcelFile(t, [][]string{{}}, [][]string{invalidHeaderParticipant}, true, false)
}

func scenarioInvalidHeaderCourseSheet(t *testing.T) []byte {
	invalidHeaderCourse := []string{"das", "ist", "kein", "header"}

	return buildExcelFile(t, [][]string{invalidHeaderCourse}, [][]string{}, false, true)
}

func scenarioOnlyStringValuesInParticipantsSheet(t *testing.T) []byte {
	onlyStringParticipant := []string{"id", "foo", "bar", "baz"}

	return buildExcelFile(t, [][]string{{}}, [][]string{onlyStringParticipant}, true, true)
}

func buildParticipantSheet(t *testing.T, excelFile *excelize.File, participants [][]string, writeHeader bool) {
	is := is.New(t)

	sheetWriter, err := NewSheetWriter(excelFile, "Teilnehmer")
	is.NoErr(err)

	if writeHeader {
		is.NoErr(sheetWriter.Write(Participant{}.RecordHeader()))
	}

	for _, participant := range participants {
		is.NoErr(sheetWriter.Write(participant))
	}
}

func buildCourseSheet(t *testing.T, excelFile *excelize.File, courses [][]string, writeHeader bool) {
	is := is.New(t)

	sheetWriter, err := NewSheetWriter(excelFile, "Kurse")
	is.NoErr(err)

	if writeHeader {
		is.NoErr(sheetWriter.Write(Course{}.RecordHeader()))
	}

	for _, course := range courses {
		is.NoErr(sheetWriter.Write(course))
	}
}

func buildExcelFile(t *testing.T, courses, participants [][]string, writeCourseHeader, writeParticipantHeader bool) []byte {
	is := is.New(t)

	excelFile := excelize.NewFile()

	buildCourseSheet(t, excelFile, courses, writeCourseHeader)
	buildParticipantSheet(t, excelFile, participants, writeParticipantHeader)

	var buf bytes.Buffer
	err := excelFile.Write(&buf)
	is.NoErr(err)

	return buf.Bytes()
}
