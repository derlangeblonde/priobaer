package model

import "gorm.io/gorm"

const MaxPriorityLevel int = 24

type PriorityLevel uint8

type Priority struct {
	gorm.Model
	Level PriorityLevel 
	CourseID int
	ParticipantID int
	Course Course
}

type Priorities []Priority

func (ps Priorities) CourseIDs() []int {
	var result []int
	for _, p := range ps {
		result = append(result, p.CourseID)
	}

	return result
}
