package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	ID int
	Name         string `gorm:"unique"`
	MaxCapacity  int
	MinCapacity  int
	Participants []Participant
}

func (c Course) RecordHeader() []string {
	return []string{"ID", "Name", "Minimale Kapazität", "Maximale Kapazität"}
}

func (c *Course) Allocation() int {
	return len(c.Participants)
}

func (c *Course) RemainingCapacity() int {
	return c.MaxCapacity - c.Allocation() 
}

func (c *Course) GapToMinCapacity() int {
	return c.MinCapacity - c.Allocation()
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

func (c *Course) TrimFields()  {
	c.Name = strings.TrimSpace(c.Name)
}

func (c *Course) MarshalRecord() []string {
	return []string{
		strconv.Itoa(c.ID),
		c.Name,
		strconv.Itoa(c.MinCapacity),
		strconv.Itoa(c.MaxCapacity),
	}
}

func (c *Course) UnmarshalRecord(record []string) error {
	const recordLen int = 4
	if len(record) != recordLen {
		return fmt.Errorf("Die Zeile hat %d Werte bzw. Spalten. Genau %d sind erwartet.", len(record), recordLen)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		c.ID = id
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

type Courses []Course

func (cs Courses) Names() []string {
        var names []string
        for _, course := range cs {
                names = append(names, course.Name)
        }
        return names
}

func MapToCourseId(courses []Course) []int{
	courseIds := make([]int, 0)
	for _, course := range courses {
		courseIds = append(courseIds, course.ID)
	}

	return courseIds
}

func stackValidationErrors(validationErrors map[string]string) error {
	if len(validationErrors) == 0 {
		return nil
	}

	validErrMessages := make([]string, 0)
	for _, value := range validationErrors {
		validErrMessages= append(validErrMessages, value)
	}			

	return errors.New(strings.Join(validErrMessages, "\n"))
}
