package model

import (
	"database/sql"
	"fmt"
	"strconv"

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

func (p *Participant) Valid() map[string]string {
	validationErrors := make(map[string]string, 0)

	validateNonEmpty(p.Surname, "surname", "Nachname darf nicht leer sein", validationErrors)
	validateNonEmpty(p.Prename, "prename", "Vorname darf nicht leer sein", validationErrors)

	return validationErrors
}

func (p *Participant) UnmarshalRecord(record []string) error {
	const recordLen int = 3
	if len(record) != recordLen {
		return fmt.Errorf("Record to construct participant from has to have length: %d, this one has length: %d", recordLen, len(record))
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		p.ID = id
	} else {
		return err
	}

	p.Prename = record[1]
	p.Surname = record[2]

	return nil
}

func (p *Participant) MarshalRecord() []string {
	return []string{
		strconv.Itoa(p.ID),
		p.Prename,
		p.Surname,
	}
}

