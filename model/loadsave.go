package model

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"slices"

	"github.com/xuri/excelize/v2"
)

const participantsSheetName = "Teilnehmer"
const courseSheetName = "Kurse"
const versionSheetName = "Version"

func ToExcelBytes(courses []Course, participants []Participant) ([]byte, error) {
	file := excelize.NewFile()
	writer, err := NewSheetWriter(file, courseSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.Write(Course{}.RecordHeader())
	for _, course := range courses {
		writer.Write(course.MarshalRecord())
	}

	writer, err = NewSheetWriter(file, participantsSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.Write(Participant{}.RecordHeader())
	for _, participant := range participants {
		writer.Write(participant.MarshalRecord())
	}

	writer, err = NewSheetWriter(file, versionSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.Write([]string{"1.0"})

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		fmt.Printf("Error writing Excel file to buffer: %v\n", err)
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func FromExcelBytes(fileReader io.Reader) (courses []Course, participants []Participant, err error) {
	exisingCourseIds := make(map[int]bool)

	file, err := excelize.OpenReader(fileReader)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create Excel file from bytes: %w", err)
	}
	reader, err := NewSheetReader(file, courseSheetName)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}

	courseHeader, err := reader.Read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(courseHeader, Course{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(courseSheetName, courseHeader, Course{}.RecordHeader())
	}
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read() {
		if err != nil {
			return courses, participants, err
		}

		course := Course{}
		err := course.UnmarshalRecord(record)
		if err != nil {
			return courses, participants, fmt.Errorf("Tabellenblatt: Kurse\n%w", err)
		}
		courses = append(courses, course)
		exisingCourseIds[course.ID] = true
	}

	reader, err = NewSheetReader(file, participantsSheetName)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	participantHeader, err := reader.Read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(participantHeader, Participant{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(participantsSheetName, participantHeader, Participant{}.RecordHeader()) 
	}
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read() {
		if err != nil {
			return courses, participants, err
		}

		participant := Participant{}
		if err = participant.UnmarshalRecord(record); err != nil {
			return courses, participants, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
		}
		if _, exists := exisingCourseIds[int(participant.CourseID.Int64)]; participant.CourseID.Valid && !exists {
			return courses, participants, fmt.Errorf("Tabellenblatt: %s\nTeilnehmer %d kann Kurs %d nicht zugeordnet werden. Dieser Kurs existiert nicht", participantsSheetName, participant.ID, participant.CourseID.Int64)
		}
		participants = append(participants, participant)
	}

	return courses, participants, err
}

func invalidHeaderError(sheetName string, gotHeader, wantHeader []string) error {
	return fmt.Errorf(
		"Tabellenblatt: %s\nKopfzeile anders als erwartet. Gefunden: '%v', Erwartet: '%v'",
		sheetName,
		strings.Join(gotHeader, ", "),
		strings.Join(wantHeader, ", "),
	)
}
