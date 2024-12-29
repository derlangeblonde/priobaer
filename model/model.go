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
	Course   Course `gorm:"constraint:OnDelete:SET NULL;"`
}

type Course struct {
	gorm.Model
	ID int
	// TODO: unique constraint does not go well with soft delete
	Name         string `gorm:"unique"`
	MaxCapacity  int
	MinCapacity  int
	Participants []Participant
}

func (c *Course) Allocation() int {
	return len(c.Participants)
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

	return errors
}
