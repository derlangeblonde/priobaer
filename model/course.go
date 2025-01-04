package model

import (
	"strconv"

	"gorm.io/gorm"
)

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

func (c *Course) RemainingCapacity() int {
	return c.MaxCapacity - c.Allocation() 
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

	validateNonEmpty(c.Name, "name", "Name darf nicht leer sein", errors)

	return errors
}

// TODO: forbid quotes and/or csv-delimiters in all string props (participants too)
func (c *Course) CsvRow() []string {
	return []string{
		strconv.Itoa(c.ID),
		// fmt.Sprintf("\"%s\"", c.Name),
		c.Name,
		strconv.Itoa(c.MinCapacity),
		strconv.Itoa(c.MaxCapacity),
	}
}
