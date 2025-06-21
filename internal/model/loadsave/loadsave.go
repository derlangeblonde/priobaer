package loadsave

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"softbaer.dev/ass/internal/model"
)

const participantsSheetName = "Teilnehmer"
const courseSheetName = "Kurse"
const versionSheetName = "Version"

const assignmentColumnHeader = "Zuteilung (Kurs ID)" 

type ParticipantID int
type CourseID int

type Course struct {
	ID          CourseID
	Name        string
	MinCapacity int
	MaxCapacity int
}

func (c Course) RecordHeader() []string {
	return []string{"ID", "Name", "Minimale Kapazität", "Maximale Kapazität"}
}

func (c *Course) MarshalRecord() []string {
	return []string{
		strconv.Itoa(int(c.ID)),
		c.Name,
		strconv.Itoa(c.MinCapacity),
		strconv.Itoa(c.MaxCapacity),
	}
}

func (c *Course) Valid() map[string]string {
	errors := make(map[string]string, 0)

	if c.MinCapacity > c.MaxCapacity {
		errors["min-capacity"] = "Die minmale Kapazität muss kleiner oder gleich der maximalen Kapazität sein"
		errors["max-capacity"] = "Die maxmale Kapazität muss größer oder gleich der minimalen Kapazität sein"
	}

	if c.MaxCapacity <= 0 {
		errors["max-capacity"] = "Die maximale Kapazität muss größer null sein"
	}

	validateNonEmpty(c.Name, "name", "Name darf nicht leer sein", errors)

	c.TrimFields()

	return errors
}

func (c *Course) TrimFields()  {
	c.Name = strings.TrimSpace(c.Name)
}

func validateNonEmpty(field, mapFieldName, errorMessage string, errorMap map[string]string) {
	if len(field) == 0 {
		errorMap[mapFieldName] = errorMessage
	}
}

func stackValidationErrors(validationErrors map[string]string) error {
	if len(validationErrors) == 0 {
		return nil
	}

	validErrMessages := make([]string, 0)
	for _, value := range validationErrors {
		validErrMessages= append(validErrMessages, value)
	}			

	return errors.New(strings.Join(validErrMessages, "\n"))
}

func (c *Course) UnmarshalRecord(record []string) error {
	const recordLen int = 4
	if len(record) != recordLen {
		return fmt.Errorf("Die Zeile hat %d Werte bzw. Spalten. Genau %d sind erwartet.", len(record), recordLen)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		c.ID = CourseID(id)
	} else {
		return err
	}

	c.Name = record[1]

	if minCap, err := strconv.Atoi(record[2]); err == nil {
		c.MinCapacity = minCap
	} else {
		return err
	}

	if maxCap, err := strconv.Atoi(record[3]); err == nil {
		c.MaxCapacity = maxCap
	} else {
		return err
	}

	return stackValidationErrors(c.Valid())
}

type Participant struct {
	ID      ParticipantID
	Prename string
	Surname string
}

func (p Participant) RecordHeader() []string {
	return []string{"ID", "Vorname", "Nachname"}
}
func (p *Participant) MarshalRecord() []string {

	return []string{
		strconv.Itoa(int(p.ID)),
		p.Prename,
		p.Surname,
	}
}

func (p *Participant) Valid() map[string]string {
	validationErrors := make(map[string]string, 0)

	validateNonEmpty(p.Surname, "surname", "Nachname darf nicht leer sein", validationErrors)
	validateNonEmpty(p.Prename, "prename", "Vorname darf nicht leer sein", validationErrors)

	p.TrimFields()

	return validationErrors
}

func (p *Participant) TrimFields() {
	p.Prename = strings.TrimSpace(p.Prename)
	p.Surname = strings.TrimSpace(p.Surname)
}

func (p *Participant) UnmarshalRecord(record []string) error {
	const fn string = "UnmarshalRecord"
	const structType string = "Teilnehmer"

	const requiredValueCount int = 3
	if len(record) < requiredValueCount {
		return fmt.Errorf("Die Zeile hat %d Werte bzw. Spalten. Mindestens %d sind erwartet.", len(record), requiredValueCount)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		p.ID = ParticipantID(id)
	} else {
		return errors.New(fmt.Sprintf("Spalte: ID\n%s ist keine valide Zahl", record[0])) 
	}

	p.Prename = record[1]
	p.Surname = record[2]

	return stackValidationErrors(p.Valid()) 
}

type Scenario struct {
	courses      []Course
	participants []Participant
	assignmentTable map[ParticipantID]*Course
	priorityTable map[ParticipantID][]*Course
}

func EmptyScenario() *Scenario {
	return &Scenario{
		courses: make([]Course, 0),
		participants: make([]Participant, 0),
		assignmentTable: make(map[ParticipantID]*Course, 0),
		priorityTable: make(map[ParticipantID][]*Course, 0),
	}
}

func (s *Scenario) AddCourse(c Course) {
	s.courses = append(s.courses, c)
}

func (s *Scenario) AddParticipant(p Participant) {
	s.participants = append(s.participants, p)
}

var ErrNotFound = errors.New("not found")

func (s *Scenario) course(cid CourseID) (*Course, bool) {
	for i := range s.courses {
		if s.courses[i].ID == cid {
			return &s.courses[i], true
		}
	}
	return nil, false
}

func (s *Scenario) participant(pid ParticipantID) (*Participant, bool) {
	for i := range s.participants {
		if s.participants[i].ID == pid {
			return &s.participants[i], true
		}
	}
	return nil, false
}

func (s *Scenario) Assign(pid ParticipantID, cid CourseID) error {
	if _, ok := s.participant(pid); !ok {
		return ErrNotFound
	}

	c, ok := s.course(cid)
	if !ok {
		return ErrNotFound
	}
	s.assignmentTable[pid] = c
	return nil
}

func (s *Scenario) Unassign(pid ParticipantID) error {
	if _, ok := s.participant(pid); !ok {
		return ErrNotFound
	}

	delete(s.assignmentTable, pid)
	return nil
}

func (s *Scenario) Prioritize(pid ParticipantID, cids []CourseID) error {
	if _, ok := s.participant(pid); !ok {
		return ErrNotFound
	}

	prioCourses := make([]*Course, 0, len(cids))
	for _, cid := range cids {
		c, ok := s.course(cid)
		if !ok {
			return ErrNotFound
		}
		prioCourses = append(prioCourses, c)
	}
	s.priorityTable[pid] = prioCourses
	return nil
}

func (s *Scenario) AllCourses() iter.Seq[Course]{
	return slices.Values(s.courses)
}

func (s *Scenario) AllParticipants() iter.Seq[Participant] {
	return slices.Values(s.participants)
}

func (s *Scenario) AssignedCourse(pid ParticipantID) (Course, bool) {
	course, ok := s.assignmentTable[pid] 

	if !ok {
		return Course{}, false
	}

	return *course, true 
}

func (s *Scenario) PrioritizedCoursesOrdered(pid ParticipantID) iter.Seq[Course] {
	courses := s.priorityTable[pid]

	return func(yield func(Course) bool) {
		for _, course := range courses {
			if !yield(*course) {
				return
			}
		}
	}	
}

func (s *Scenario) MaxAmountOfPriorities() (result int) {
	for _, courses := range s.priorityTable {
		count := len(courses)
		if count > result {
			result = count 
		}
	}

	return
}

type candiateAssignment struct {
	pid ParticipantID
	cid CourseID
}

type candidatePrioList struct {
	pid ParticipantID
	cidOrdered []CourseID
}

func validateParticipantHeader(header []string) error {
    expected := Participant{}.RecordHeader()

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

func Import(fileReader io.Reader) (*Scenario, error) {
	scenario := EmptyScenario()
	var candidateAssignments []candiateAssignment
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
	if !slices.Equal(courseHeader, Course{}.RecordHeader()) {
		return scenario, invalidHeaderError(courseSheetName, courseHeader, Course{}.RecordHeader())
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return scenario, err
		}

		course := Course{}
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

		participant := Participant{}
		if err = participant.UnmarshalRecord(record); err != nil {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
		}

		scenario.AddParticipant(participant)

		minimumRecordLen := len(Participant{}.RecordHeader()) + 1

		if len(record) < minimumRecordLen {
			return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, fmt.Errorf("Zeile hat %d Werte. Es müssen mind. %d sein", len(record), minimumRecordLen))
		}

		assignedCourseIdStr := record[minimumRecordLen - 1]

		if assignedCourseIdStr != "null" {
			
			if assignedCourseId, err := strconv.Atoi(assignedCourseIdStr); err != nil {
				return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
			} else {
				candidateAssignments = append(candidateAssignments, candiateAssignment{participant.ID, CourseID(assignedCourseId)})
			}
		}


		var prioList []CourseID
		for i := minimumRecordLen; i < len(record); i++ {
			prioStr := record[i]

			if prio, err := strconv.Atoi(prioStr); err != nil {
				return scenario, fmt.Errorf("Tabellenblatt: %s\n%w", participantsSheetName, err)
			} else {
				prioList = append(prioList, CourseID(prio))
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


func Export(scenario *Scenario) ([]byte, error) {
	var buf bytes.Buffer
	var writer *sheetWriter

	file := excelize.NewFile()
	writer, err := newSheetWriter(file, courseSheetName)
	if err != nil {
		return buf.Bytes(), err
	}

	writer.write(Course{}.RecordHeader())
	for course := range scenario.AllCourses() {
		writer.write(course.MarshalRecord())
	}

	if writer, err = newSheetWriter(file, participantsSheetName); err != nil {
		return buf.Bytes(), err
	}

	participantsSheetHeader := append(Participant{}.RecordHeader(), assignmentColumnHeader)
	for i := range scenario.MaxAmountOfPriorities() {
		participantsSheetHeader = append(participantsSheetHeader, nthPriorityColumnHeader(i+1))
	}

	writer.write(participantsSheetHeader)
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

		writer.write(row)
	}

	if writer, err = newSheetWriter(file, participantsSheetName); err != nil {
		return buf.Bytes(), err
	}

	writer.write(append(Participant{}.RecordHeader(), assignmentColumnHeader))
	for participant := range scenario.AllParticipants() {
		assignedCourse, ok := scenario.AssignedCourse(participant.ID)
		courseIdMarshalled := "null" 

		if ok {
			courseIdMarshalled = strconv.Itoa(int(assignedCourse.ID))
		}

		writer.write(append(participant.MarshalRecord(), courseIdMarshalled))
	}

	if writer, err = newSheetWriter(file, versionSheetName); err != nil {
		return buf.Bytes(), err
	}

	writer.write([]string{"1.0"})

	if err := file.Write(&buf); err != nil {
		fmt.Printf("Error writing Excel file to buffer: %v\n", err)
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func ToExcelBytes(courses []model.Course, participants []model.Participant) ([]byte, error) {
	file := excelize.NewFile()
	writer, err := newSheetWriter(file, courseSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.write(model.Course{}.RecordHeader())
	for _, course := range courses {
		writer.write(course.MarshalRecord())
	}

	writer, err = newSheetWriter(file, participantsSheetName)
	if err != nil {
		return make([]byte, 0), err
	}

	writer.write(model.Participant{}.RecordHeader())
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

func FromExcelBytes(fileReader io.Reader) (courses []model.Course, participants []model.Participant, err error) {
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
	if !slices.Equal(courseHeader, model.Course{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(courseSheetName, courseHeader, model.Course{}.RecordHeader())
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return courses, participants, err
		}

		course := model.Course{}
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
	if !slices.Equal(participantHeader, model.Participant{}.RecordHeader()) {
		return courses, participants, invalidHeaderError(participantsSheetName, participantHeader, model.Participant{}.RecordHeader())
	}
	for record, err := reader.read(); err != io.EOF; record, err = reader.read() {
		if err != nil {
			return courses, participants, err
		}

		participant := model.Participant{}
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
