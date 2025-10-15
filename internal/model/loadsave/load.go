package loadsave

import (
	"fmt"
	"io"
	"slices"
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

func LoadScenarioFromExcelFile(fileReader io.Reader) (*domain.Scenario, error) {
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
		colsRead, err := participant.UnmarshalRecord(record)
		if err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
		}
		scenario.AddParticipant(participant)
		record = record[colsRead:]

		if len(record) <= 0 {
			continue
		}

		assignedCourseName := record[0]
		assignedCourseName = strings.TrimSpace(assignedCourseName)

		if assignedCourseName != "" {
			if assignedCourse, ok := scenario.FindCourseByName(assignedCourseName); !ok {
				return scenario, fmt.Errorf("Tabellenblatt: %s\nKeine Zuteilung möglich da Kurs nicht existiert: '%s'", participantsSheetName, assignedCourseName)
			} else {
				candidateAssignments = append(candidateAssignments, candidateAssignment{participant.ID, assignedCourse.ID})
			}
		}
		record = record[1:]

		if len(record) <= 0 {
			continue
		}

		var prioList []domain.CourseID
		for _, prioName := range record {
			if prio, ok := scenario.FindCourseByName(prioName); !ok {
				return scenario, fmt.Errorf("Tabellenblatt: %s\nKeine Priorisierung möglich da Kurs nicht existiert: '%s'", participantsSheetName, prioName)
			} else {
				prioList = append(prioList, prio.ID)
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
