package domain

import (
	"database/sql"
	"errors"
	"iter"
	"slices"
	"softbaer.dev/ass/internal/model"
)

type Scenario struct {
	courses         []Course
	participants    []Participant
	assignmentTable map[ParticipantID]*Course
	priorityTable   map[ParticipantID][]*Course
}

func EmptyScenario() *Scenario {
	return &Scenario{
		courses:         make([]Course, 0),
		participants:    make([]Participant, 0),
		assignmentTable: make(map[ParticipantID]*Course, 0),
		priorityTable:   make(map[ParticipantID][]*Course, 0),
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

func (s *Scenario) AllCourses() iter.Seq[Course] {
	return slices.Values(s.courses)
}

func (s *Scenario) AllParticipants() iter.Seq[Participant] {
	return slices.Values(s.participants)
}

func (s *Scenario) AllPrioLists() iter.Seq2[ParticipantID, []Course] {

	return func (yield func(ParticipantID, []Course) bool)  {
		for pid, coursePointers := range s.priorityTable{
			var courses []Course
			for _, coursePointer := range coursePointers {
				courses = append(courses, *coursePointer)
			}

			if !yield(pid, courses) {
				return
			}
		}
	}
}


func (s *Scenario) AllPriorities() iter.Seq[Priority] {
	return func (yield func(Priority) bool)  {
		for pid, courses := range s.priorityTable{
			participant, ok := s.participant(pid)

			if !ok {
				continue
			}

			for i, course := range courses {
				current := Priority{Level: PriorityLevel(i+1), Participant: *participant,  Course: *course}
				if !yield(current) {
					return
				}
			}

		}
	}
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

func (s *Scenario) allParticipantsAsDbModels() []model.Participant {
	result := make([]model.Participant, len(s.participants))
	for i, p := range s.participants {
		assignedCourse, ok := s.assignmentTable[p.ID]
		var nullableAssignedId sql.NullInt64
		if ok {
			nullableAssignedId = sql.NullInt64{Valid: ok, Int64: int64(assignedCourse.ID)}
		} else {
			nullableAssignedId = sql.NullInt64{Valid: false}
		}
		result[i] = model.Participant{ID: int(p.ID), Prename: p.Prename, Surname: p.Surname, CourseID: nullableAssignedId}
	}

	return result
}
