package model

import (
	"bytes"
	"encoding/csv"
	"io"
)

// TODO: do we want to write a version to the saved files???
func toCsvBytes(courses []Course, participants []Participant) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	writer := csv.NewWriter(buf)

	for _, course := range courses {
		writer.Write(course.MarshalRecord())	
	}

	writer.Flush()

	return buf.Bytes(), nil
}

func fromCsvBytes(csvBytes []byte) (courses []Course, participants []Participant, err error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read(){
		course := Course{}
		course.UnmarshalRecord(record)
		courses = append(courses, course)
	}


	return courses, participants, err
}
