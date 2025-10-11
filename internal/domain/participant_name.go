package domain

import "strings"

type ParticipantName struct {
	Prename string
	Surname string
}

func (p *ParticipantName) ParticipantNameValid() map[string]string {
	validationErrors := make(map[string]string)

	validateNonEmpty(p.Surname, "surname", "Nachname darf nicht leer sein", validationErrors)
	validateNonEmpty(p.Prename, "prename", "Vorname darf nicht leer sein", validationErrors)

	p.TrimFields()

	return validationErrors
}

func (p *ParticipantName) TrimFields() {
	p.Prename = strings.TrimSpace(p.Prename)
	p.Surname = strings.TrimSpace(p.Surname)
}
