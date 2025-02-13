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
