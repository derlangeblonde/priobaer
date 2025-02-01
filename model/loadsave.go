package model

import (
	"bytes"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
	"slices"
)

func ToExcelBytes(courses []Course, participants []Participant) ([]byte, error) {
	file := excelize.NewFile()
	writer, err := NewSheetWriter(file, "Kurse")
	if err != nil {
		return make([]byte, 0), err
	}

	writer.Write(Course{}.RecordHeader())
	for _, course := range courses {
		writer.Write(course.MarshalRecord())
	}

	writer, err = NewSheetWriter(file, "Teilnehmer")
	if err != nil {
		return make([]byte, 0), err
	}

	writer.Write(Participant{}.RecordHeader())
	for _, participant := range participants {
		writer.Write(participant.MarshalRecord())
	}

	writer, err = NewSheetWriter(file, "Version")
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

type UnmarshalExcelBytesError struct {
	Sheet string
	Inner error
}

func (e *UnmarshalExcelBytesError) Error() string {
	return "Tabellenblatt: " + e.Sheet 
}

func (e *UnmarshalExcelBytesError) Unwrap() error {
	return e.Inner
}

func unmarshalExcelBytesError(sheet string, err error) *UnmarshalExcelBytesError {
	return &UnmarshalExcelBytesError{Sheet: sheet, Inner: err}
}

func FromExcelBytes(fileReader io.Reader) (courses []Course, participants []Participant, err error) {
	const participantsSheet string = "Teilnehmer"
	file, err := excelize.OpenReader(fileReader)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create Excel file from bytes: %w", err)
	}
	reader, err := NewSheetReader(file, "Kurse")
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}

	courseHeader, err := reader.Read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(courseHeader, Course{}.RecordHeader()) {
		return courses, participants, fmt.Errorf("Headers for courses differ. Got=%v, Want=%v", courseHeader, Course{}.RecordHeader())
	}
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read() {
		if err != nil {
			return courses, participants, err
		}

		course := Course{}
		course.UnmarshalRecord(record)
		courses = append(courses, course)
	}

	reader, err = NewSheetReader(file, participantsSheet)
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	participantHeader, err := reader.Read()
	if err != nil && err != io.EOF {
		return courses, participants, err
	}
	if !slices.Equal(participantHeader, Participant{}.RecordHeader()) {
		return courses, participants, fmt.Errorf("Headers for participants differ. Got=%v, Want=%v", participantHeader, Participant{}.RecordHeader())
	}
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read() {
		if err != nil {
			return courses, participants, err
		}

		participant := Participant{}
		if err = participant.UnmarshalRecord(record); err != nil {
			return courses, participants, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheet, err)
		}
		participants = append(participants, participant)
	}

	return courses, participants, err
}
