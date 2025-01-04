package model

import (
	"bytes"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// TODO: do we want to write a version to the saved files???
func toExcelBytes(courses []Course, participants []Participant) ([]byte, error) {
	file := excelize.NewFile()
	writer, err := NewSheetWriter(file, "Kurse")
	if err != nil {
		return make([]byte, 0), err
	}

	for _, course := range courses {
		writer.Write(course.MarshalRecord())	
	}

	writer, err = NewSheetWriter(file, "Teilnehmer")
	if err != nil {
		return make([]byte, 0), err
	}

	for _, participant := range participants {
		writer.Write(participant.MarshalRecord())	
	}

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		fmt.Printf("Error writing Excel file to buffer: %v\n", err)
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func fromExcelBytes(csvBytes []byte) (courses []Course, participants []Participant, err error) {
	file, err := excelize.OpenReader(bytes.NewReader(csvBytes))
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create Excel file from bytes: %w", err)
	}
	reader, err := NewSheetReader(file, "Kurse")
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read(){
		if err != nil {
			return courses, participants, err
		}

		course := Course{}
		course.UnmarshalRecord(record)
		courses = append(courses, course)
	}

	reader, err = NewSheetReader(file, "Teilnehmer")
	if err != nil {
		return courses, participants, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read(){
		if err != nil {
			return courses, participants, err
		}

		participant := Participant{}
		participant.UnmarshalRecord(record)
		participants = append(participants, participant)
	}


	return courses, participants, err
}
