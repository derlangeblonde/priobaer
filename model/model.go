package model

import (
	"database/sql"

	"gorm.io/gorm"
)

type Participant struct {
	gorm.Model
	ID       int
	Prename  string
	Surname  string
	CourseID sql.NullInt64
	Course   Course
}

type Course struct {
	gorm.Model
	ID int
	// TODO: unique constraint does not go well with soft delete
	Name        string `gorm:"unique"`
	MaxCapacity int
	MinCapacity int
}
