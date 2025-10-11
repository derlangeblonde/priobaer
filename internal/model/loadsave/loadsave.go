package loadsave

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"softbaer.dev/ass/internal/domain"

	"github.com/xuri/excelize/v2"
)

const participantsSheetName = "Teilnehmer"
const courseSheetName = "Kurse"
const versionSheetName = "Version"

const assignmentColumnHeader = "Zuteilung (Kurs ID)"

type candidateAssignment struct {
	pid domain.ParticipantID
	cid domain.CourseID
}

type candidatePrioList struct {
	pid        domain.ParticipantID
	cidOrdered []domain.CourseID
}

func validateParticipantHeader(header []string) error {
	expected := domain.ParticipantDataRecordHeader()

	if len(header) < len(expected)+1 {
		return invalidHeaderError(participantsSheetName, header, expected)
	}

	for i, want := range expected {
		if header[i] != want {
			return invalidHeaderError(participantsSheetName, header, expected)
		}
	}

	if header[len(expected)] != assignmentColumnHeader {
		return invalidHeaderError(participantsSheetName, header, expected)
	}

	for i := len(expected) + 1; i < len(header); i++ {
		if header[i] != nthPriorityColumnHeader(i-len(expected)) {
			return invalidHeaderError(participantsSheetName, header, expected)
		}
	}

	return nil
}

func ParseExcelFile(fileReader io.Reader) (*domain.Scenario, error) {
	scenario := domain.EmptyScenario()
	var candidateAssignments []candidateAssignment
	var candidatePrioLists []candidatePrioList

	file, err := excelize.OpenReader(fileReader)
	if err != nil {
		return scenario, fmt.Errorf("failed to create Excel file from bytes: %w", err)
	}
	reader, err := newSheetReader(file, courseSheetName)
	if err != nil {
		return scenario, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}

	courseHeader, err := reader.read()
	if err != nil && err != io.EOF {
		return scenario, err
	}
	if !slices.Equal(courseHeader, domain.CourseDataRecordHeader()) {
		return scenario, invalidHeaderError(courseSheetName, courseHeader, domain.CourseDataRecordHeader())
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return scenario, err
		}

		course := domain.CourseData{}
		err := course.UnmarshalRecord(record)
		if err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: Kurse\n%w", err)
		}
		scenario.AddCourse(course)
	}

	reader, err = newSheetReader(file, participantsSheetName)
	if err != nil {
		return scenario, fmt.Errorf("failed to create excel sheet reader: %w", err)
	}
	participantHeader, err := reader.read()
	if err != nil && err != io.EOF {
		return scenario, err
	}
	if err = validateParticipantHeader(participantHeader); err != nil {
		return scenario, err
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return scenario, err
		}

		participant := domain.ParticipantData{}
		if err = participant.UnmarshalRecord(record); err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
		}

		scenario.AddParticipant(participant)

		minimumRecordLen := len(domain.ParticipantDataRecordHeader()) + 1

		if len(record) < minimumRecordLen {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, fmt.Errorf("zeile hat %d Werte. Es müssen mind. %d sein", len(record), minimumRecordLen))
		}

		assignedCourseIdStr := record[minimumRecordLen-1]

		if assignedCourseIdStr != "null" {

			if assignedCourseId, err := strconv.Atoi(assignedCourseIdStr); err != nil {
				return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
			} else {
				candidateAssignments = append(candidateAssignments, candidateAssignment{participant.ID, domain.CourseID(assignedCourseId)})
			}
		}

		var prioList []domain.CourseID
		for i := minimumRecordLen; i < len(record); i++ {
			prioStr := record[i]

			if prio, err := strconv.Atoi(prioStr); err != nil {
				return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
			} else {
				prioList = append(prioList, domain.CourseID(prio))
			}
		}
		candidatePrioLists = append(candidatePrioLists, candidatePrioList{participant.ID, prioList})
	}

	for _, a := range candidateAssignments {
		if err = scenario.Assign(a.pid, a.cid); err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: %s\nTeilnehmer %d kann Kurs %d nicht zugeordnet werden. Dieser Kurs existiert nicht", participantsSheetName, a.pid, a.cid)
		}
	}

	for _, p := range candidatePrioLists {
		if err = scenario.Prioritize(p.pid, p.cidOrdered); err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
		}
	}

	return scenario, err
}

func nthPriorityColumnHeader(n int) string {
	return fmt.Sprintf("Priorität %d (Kurs ID)", n)
}

func WriteScenarioDataToExcel(scenario *domain.Scenario) ([]byte, error) {
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
		courseIdMarshalled := "null"

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

func invalidHeaderError(sheetName string, gotHeader, wantHeader []string) error {
	return fmt.Errorf(
		"Tabellenblatt: %s\nKopfzeile anders als erwartet. Gefunden: '%v', Erwartet: '%v'",
		sheetName,
		strings.Join(gotHeader, ", "),
		strings.Join(wantHeader, ", "),
	)
}
