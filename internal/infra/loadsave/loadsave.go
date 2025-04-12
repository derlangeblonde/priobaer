package loadsave

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"slices"

	"github.com/xuri/excelize/v2"
	"softbaer.dev/ass/internal/infra"
)

const participantsSheetName = "Teilnehmer"
const courseSheetName = "Kurse"
const versionSheetName = "Version"

func ToExcelBytes(courses []infra.Course, participants []infra.Participant) ([]byte, error) {
	file := excelize.NewFile()
	writer, err := newSheetWriter(file, courseSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.write(infra.Course{}.RecordHeader())
	for _, course := range courses {
		writer.write(course.MarshalRecord())
	}

	writer, err = newSheetWriter(file, participantsSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.write(infra.Participant{}.RecordHeader())
	for _, participant := range participants {
		writer.write(participant.MarshalRecord())
	}

	writer, err = newSheetWriter(file, versionSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.write([]string{"1.0"})

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		fmt.Printf("Error writing Excel file to buffer: %v\n", err)
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func FromExcelBytes(fileReader io.Reader) (courses []infra.Course, participants []infra.Participant, err error) {
	exisingCourseIds := make(map[int]bool)

	file, err := excelize.OpenReader(fileReader)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create Excel file from bytes: %w", err)
	}
	reader, err := newSheetReader(file, courseSheetName)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}

	courseHeader, err := reader.read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(courseHeader, infra.Course{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(courseSheetName, courseHeader, infra.Course{}.RecordHeader())
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return courses, participants, err
		}

		course := infra.Course{}
		err := course.UnmarshalRecord(record)
		if err != nil {
			return courses, participants, fmt.Errorf("Tabellenblatt: Kurse\n%w", err)
		}
		courses = append(courses, course)
		exisingCourseIds[course.ID] = true
	}

	reader, err = newSheetReader(file, participantsSheetName)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	participantHeader, err := reader.read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(participantHeader, infra.Participant{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(participantsSheetName, participantHeader, infra.Participant{}.RecordHeader()) 
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return courses, participants, err
		}

		participant := infra.Participant{}
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
