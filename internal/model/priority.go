package model

import "gorm.io/gorm"

const MaxPriorityLevel int = 24

type PriorityLevel uint8

type Priority struct {
	gorm.Model
	Level         PriorityLevel
	CourseID      int
	ParticipantID int
	Course        Course
	Participant   Participant
}

func NewPriority(level PriorityLevel, course Course, participant Participant) Priority {
	return Priority{Level: level, Course: course, CourseID: course.ID, Participant: participant, ParticipantID: participant.ID}
}

func (p Priority) AssignmentID() AssignmentID {
	return AssignmentID{CourseId: p.CourseID, ParticipantId: p.ParticipantID}
}

type Priorities []Priority

func (ps Priorities) CourseIDs() []int {
	var result []int
	for _, p := range ps {
		result = append(result, p.CourseID)
	}

	return result
}
