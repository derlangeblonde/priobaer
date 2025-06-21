package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ParticipantID int

type Participant struct {
	ID      ParticipantID
	Prename string
	Surname string
}

func (p Participant) RecordHeader() []string {
	return []string{"ID", "Vorname", "Nachname"}
}
func (p *Participant) MarshalRecord() []string {

	return []string{
		strconv.Itoa(int(p.ID)),
		p.Prename,
		p.Surname,
	}
}

func (p *Participant) Valid() map[string]string {
	validationErrors := make(map[string]string, 0)

	validateNonEmpty(p.Surname, "surname", "Nachname darf nicht leer sein", validationErrors)
	validateNonEmpty(p.Prename, "prename", "Vorname darf nicht leer sein", validationErrors)

	p.TrimFields()

	return validationErrors
}

func (p *Participant) TrimFields() {
	p.Prename = strings.TrimSpace(p.Prename)
	p.Surname = strings.TrimSpace(p.Surname)
}

func (p *Participant) UnmarshalRecord(record []string) error {
	const requiredValueCount int = 3
	if len(record) < requiredValueCount {
		return fmt.Errorf("Die Zeile hat %d Werte bzw. Spalten. Mindestens %d sind erwartet.", len(record), requiredValueCount)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		p.ID = ParticipantID(id)
	} else {
		return errors.New(fmt.Sprintf("Spalte: ID\n%s ist keine valide Zahl", record[0]))
	}

	p.Prename = record[1]
	p.Surname = record[2]

	return stackValidationErrors(p.Valid())
}
