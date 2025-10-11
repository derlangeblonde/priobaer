package domain

import (
	"database/sql"
	"errors"
	"iter"
	"slices"

	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/util"
)

type Scenario struct {
	courses         []CourseData
	participants    []ParticipantData
	assignmentTable map[ParticipantID]*CourseData
	priorityTable   map[ParticipantID][]*CourseData
}

func EmptyScenario() *Scenario {
	return &Scenario{
		courses:         make([]CourseData, 0),
		participants:    make([]ParticipantData, 0),
		assignmentTable: make(map[ParticipantID]*CourseData),
		priorityTable:   make(map[ParticipantID][]*CourseData),
	}
}

func (s *Scenario) AddCourse(c CourseData) {
	s.courses = append(s.courses, c)
}

func (s *Scenario) AddParticipant(p ParticipantData) {
	s.participants = append(s.participants, p)
}

var ErrNotFound = errors.New("not found")

func (s *Scenario) course(cid CourseID) (*CourseData, bool) {
	for i := range s.courses {
		if s.courses[i].ID == cid {
			return &s.courses[i], true
		}
	}
	return nil, false
}

func (s *Scenario) participant(pid ParticipantID) (*ParticipantData, bool) {
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

	prioCourses := make([]*CourseData, 0, len(cids))
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

func (s *Scenario) AllCourses() iter.Seq[CourseData] {
	return slices.Values(s.courses)
}

func (s *Scenario) AllParticipants() iter.Seq[ParticipantData] {
	return slices.Values(s.participants)
}

func (s *Scenario) ParticipantsAssignedTo(cid CourseID) (result []ParticipantData) {
	for _, p := range s.participants {
		assignedCourse, ok := s.assignmentTable[p.ID]
		if !ok {
			continue
		}

		if assignedCourse.ID == cid {
			result = append(result, p)
		}
	}

	return
}

func (s *Scenario) Unassigned() (result []ParticipantData) {
	for _, p := range s.participants {
		if _, ok := s.assignmentTable[p.ID]; !ok {
			result = append(result, p)
		}
	}

	return
}

func (s *Scenario) allPrioListsIter() iter.Seq2[ParticipantID, []CourseData] {

	return func(yield func(ParticipantID, []CourseData) bool) {
		for pid, coursePointers := range s.priorityTable {
			var courses []CourseData
			for _, coursePointer := range coursePointers {
				courses = append(courses, *coursePointer)
			}

			if !yield(pid, courses) {
				return
			}
		}
	}
}

func (s *Scenario) AllPrioLists() map[ParticipantID][]CourseData {
	return util.Seq2ToMap(s.allPrioListsIter())
}

func (s *Scenario) AllPriorities() iter.Seq[Priority] {
	return func(yield func(Priority) bool) {
		for pid, courses := range s.priorityTable {
			participant, ok := s.participant(pid)

			if !ok {
				continue
			}

			for i, course := range courses {
				current := Priority{Level: PriorityLevel(i + 1), Participant: *participant, Course: *course}
				if !yield(current) {
					return
				}
			}

		}
	}
}

func (s *Scenario) AllocationOf(cid CourseID) (allocation int) {
	for _, course := range s.assignmentTable {
		if course.ID == cid {
			allocation++
		}
	}

	return
}

func (s *Scenario) AssignedCourse(pid ParticipantID) (CourseData, bool) {
	course, ok := s.assignmentTable[pid]

	if !ok {
		return CourseData{}, false
	}

	return *course, true
}

func (s *Scenario) PrioritizedCoursesOrdered(pid ParticipantID) iter.Seq[CourseData] {
	courses := s.priorityTable[pid]

	return func(yield func(CourseData) bool) {
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
