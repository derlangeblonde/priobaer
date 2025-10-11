package domain

import (
	"strings"

	"softbaer.dev/ass/internal/crypt"
)

type EncryptedParticipantName ParticipantName

type ParticipantName struct {
	Prename string
	Surname string
}

func (p *EncryptedParticipantName) Decrypt(secret crypt.Secret) (ParticipantName, error) {
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

func (p *ParticipantName) Encrypt(secret crypt.Secret) (EncryptedParticipantName, error) {
	encryptedPrename, err := crypt.Encrypt(p.Prename, secret)
	if err != nil {
		return EncryptedParticipantName{}, err
	}
	encryptedSurname, err := crypt.Encrypt(p.Surname, secret)
	if err != nil {
		return EncryptedParticipantName{}, err
	}

	return EncryptedParticipantName{
		Prename: encryptedPrename,
		Surname: encryptedSurname,
	}, nil
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
