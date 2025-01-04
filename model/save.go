package model

import (
	"bytes"
	"encoding/csv"
)

// TODO: do we want to write a version to the saved files???
func toCsvBytes(courses []Course, participants []Participant) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	writer := csv.NewWriter(buf)

	for _, course := range courses {
		writer.Write(course.CsvRow())	
	}

	writer.Flush()

	return buf.Bytes(), nil
}
