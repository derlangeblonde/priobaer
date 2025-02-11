package model

import "gorm.io/gorm"

type PriorityLevel uint32

type Priority struct {
	gorm.Model
	Level PriorityLevel 
	CourseID int
	ParticipantID int
	Course Course
}
