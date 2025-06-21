package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type CourseID int

type Course struct {
	ID          CourseID
	Name        string
	MinCapacity int
	MaxCapacity int
}

func (c Course) RecordHeader() []string {
	return []string{"ID", "Name", "Minimale Kapazität", "Maximale Kapazität"}
}

func (c *Course) MarshalRecord() []string {
	return []string{
		strconv.Itoa(int(c.ID)),
		c.Name,
		strconv.Itoa(c.MinCapacity),
		strconv.Itoa(c.MaxCapacity),
	}
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

	c.TrimFields()

	return errors
}

func (c *Course) TrimFields() {
	c.Name = strings.TrimSpace(c.Name)
}
func (c *Course) UnmarshalRecord(record []string) error {
	const recordLen int = 4
	if len(record) != recordLen {
		return fmt.Errorf("Die Zeile hat %d Werte bzw. Spalten. Genau %d sind erwartet.", len(record), recordLen)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		c.ID = CourseID(id)
	} else {
		return err
	}

	c.Name = record[1]

	if minCap, err := strconv.Atoi(record[2]); err == nil {
		c.MinCapacity = minCap
	} else {
		return err
	}

	if maxCap, err := strconv.Atoi(record[3]); err == nil {
		c.MaxCapacity = maxCap
	} else {
		return err
	}

	return stackValidationErrors(c.Valid())
}
