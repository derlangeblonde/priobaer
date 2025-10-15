package loadsave

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
	"softbaer.dev/ass/internal/domain"
)

func SaveScenarioToExcelFile(scenario *domain.Scenario) ([]byte, error) {
	var buf bytes.Buffer
	var writer *sheetWriter

	file := excelize.NewFile()
	writer, err := newSheetWriter(file, courseSheetName)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writer.write(domain.CourseDataRecordHeader())
	if err != nil {
		return nil, err
	}
	for course := range scenario.AllCourses() {
		if err = writer.write(course.MarshalRecord()); err != nil {
			return nil, err
		}
	}

	if writer, err = newSheetWriter(file, participantsSheetName); err != nil {
		return buf.Bytes(), err
	}

	participantsSheetHeader := append(domain.ParticipantDataRecordHeader(), assignmentColumnHeader)
	for i := range scenario.MaxAmountOfPriorities() {
		participantsSheetHeader = append(participantsSheetHeader, nthPriorityColumnHeader(i+1))
	}

	if err := writer.write(participantsSheetHeader); err != nil {
		return nil, err
	}
	for participant := range scenario.AllParticipants() {
		assignedCourse, ok := scenario.AssignedCourse(participant.ID)
		courseIdMarshalled := "null"

		if ok {
			courseIdMarshalled = strconv.Itoa(int(assignedCourse.ID))
		}

		row := append(participant.MarshalRecord(), courseIdMarshalled)

		for course := range scenario.PrioritizedCoursesOrdered(participant.ID) {
			row = append(row, strconv.Itoa(int(course.ID)))
		}

		if err := writer.write(row); err != nil {
			return nil, err
		}
	}

	if writer, err = newSheetWriter(file, participantsSheetName); err != nil {
		return buf.Bytes(), err
	}

	if err = writer.write(append(domain.ParticipantDataRecordHeader(), assignmentColumnHeader)); err != nil {
		return nil, err
	}
	for participant := range scenario.AllParticipants() {
		assignedCourse, ok := scenario.AssignedCourse(participant.ID)
		courseIdMarshalled := ""

		if ok {
			courseIdMarshalled = strconv.Itoa(int(assignedCourse.ID))
		}

		if err = writer.write(append(participant.MarshalRecord(), courseIdMarshalled)); err != nil {
			return nil, err
		}
	}

	if writer, err = newSheetWriter(file, versionSheetName); err != nil {
		return buf.Bytes(), err
	}

	if err = writer.write([]string{"1.0"}); err != nil {
		return nil, err
	}

	if err := file.Write(&buf); err != nil {
		fmt.Printf("Error writing Excel file to buffer: %v\n", err)
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
