package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/infra"
)

// Solver wants a way to easily lookup (for all unassigned Participants) which course is prioritized at what prio level

// App wants the same thing acutally

// Solver is only interested in course_ids and app primarily in course_names

type ParticipantID struct {
	Val int
}

func NewParticipantID(id int) (ParticipantID, ValidationErrs) {
	errs := EmptyValidationErrs()
	GreaterEqualZero(id, "ID", errs)

	return ParticipantID{Val: id}, errs
}

type CourseID struct {
	Val int
}

func NewCouresID(id int) (CourseID, ValidationErrs) {
	errs := EmptyValidationErrs()
	GreaterEqualZero(id, "ID", errs)

	return CourseID{Val: id}, errs
}

type CourseData struct {
	ID   CourseID
	name string
}

func NewCourseData(id int, name string) (CourseData, ValidationErrs) {
	courseID, errs := NewCouresID(id)
	NotEmpty(name, "Name", errs)

	return CourseData{ID: courseID, name: name}, errs
}

type PriorityList struct {
	ParticipantID     ParticipantID
	CoursesDescending []CourseData
}

func (list *PriorityList) Append(courseData CourseData) {
	list.CoursesDescending = append(list.CoursesDescending, courseData)
}

type PriorityLookup struct {
	innerMap map[ParticipantID]*PriorityList
}

func NewPriorityLookup() *PriorityLookup {
	return &PriorityLookup{
		innerMap: make(map[ParticipantID]*PriorityList),
	}
}

func (pl *PriorityLookup) PrioritiesFor(id ParticipantID) (PriorityList, bool) {
	result, ok := pl.innerMap[id]

	return *result, ok
}

func LoadPriorityLookupFor(tx *gorm.DB, participantIDs []ParticipantID) (*PriorityLookup, error) {
	var priorities infra.Priorities
	lookup := NewPriorityLookup()

	err := tx.
		Preload("Course").
		Select("course_id, participant_id, level").
		Where("participant_id in ?", participantIDs).
		Order("level ASC").
		Find(&priorities).Error

	if err != nil {
		return lookup, err
	}

	for _, priority := range priorities {
		participantID, errs := NewParticipantID(priority.ParticipantID)

		if !errs.Empty() {
			return lookup, errs
		}

		courseData, errs := NewCourseData(priority.CourseID, priority.Course.Name)

		if !errs.Empty() {
			return lookup, errs
		}

		lookup.innerMap[participantID].Append(courseData)
	}

	return lookup, nil
}
