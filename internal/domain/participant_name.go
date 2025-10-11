package domain

import (
	"strings"

	"softbaer.dev/ass/internal/crypt"
)

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

type encryptedParticipantName ParticipantName

func (p *encryptedParticipantName) decrypt(secret crypt.Secret) (ParticipantName, error) {
	decryptedPrename, err := crypt.Decrypt(p.Prename, secret)
	if err != nil {
		return ParticipantName{}, err
	}
	decryptedSurname, err := crypt.Decrypt(p.Surname, secret)
	if err != nil {
		return ParticipantName{}, err
	}

	return ParticipantName{
		Prename: decryptedPrename,
		Surname: decryptedSurname,
	}, nil
}

func (p *ParticipantName) encrypt(secret crypt.Secret) (encryptedParticipantName, error) {
	encryptedPrename, err := crypt.Encrypt(p.Prename, secret)
	if err != nil {
		return encryptedParticipantName{}, err
	}
	encryptedSurname, err := crypt.Encrypt(p.Surname, secret)
	if err != nil {
		return encryptedParticipantName{}, err
	}

	return encryptedParticipantName{
		Prename: encryptedPrename,
		Surname: encryptedSurname,
	}, nil
}
