package model

import "gorm.io/gorm"

type Participant struct {
	gorm.Model
	Prename string
	Surname string
}

type Course struct {
	gorm.Model
	ID int
	// TODO: unique constraint does not go well with soft delete
	Name        string `gorm:"unique"`
	MaxCapacity int
	MinCapacity int
}
