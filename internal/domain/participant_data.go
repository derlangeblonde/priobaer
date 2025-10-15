package domain

import (
	"errors"
	"fmt"
	"strconv"
)

type ParticipantID int

type ParticipantData struct {
	ID ParticipantID
	ParticipantName
}

func ParticipantDataRecordHeader() []string {
	return []string{"ID", "Vorname", "Nachname"}
}
func (p *ParticipantData) MarshalRecord() []string {

	return []string{
		strconv.Itoa(int(p.ID)),
		p.Prename,
		p.Surname,
	}
}

func (p *ParticipantData) Valid() map[string]string {
	return p.ParticipantNameValid()
}

func (p *ParticipantData) UnmarshalRecord(record []string) (columnsRead int, err error) {
	const requiredValueCount int = 3
	if len(record) < requiredValueCount {
		return 0, fmt.Errorf("die Zeile hat %d Werte bzw. Spalten. Mindestens %d sind erwartet", len(record), requiredValueCount)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		p.ID = ParticipantID(id)
	} else {
		return 0, errors.New(fmt.Sprintf("Spalte: ID\n%s ist keine valide Zahl", record[0]))
	}

	p.Prename = record[1]
	p.Surname = record[2]

	return requiredValueCount, stackValidationErrors(p.Valid())
}
