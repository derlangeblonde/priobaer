package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
)

type Participant struct {
	gorm.Model
	ID               int
	EncryptedPrename string
	EncryptedSurname string
	CourseID         sql.NullInt64
	Course           Course `gorm:"constraint:OnDelete:SET NULL;"`
}

type ParticipantOption func(*Participant)

// NewParticipant creates a new participant model it will encrypt the names passed as arguments, so that they
// will be stored in encrypted form on the db. To set anything other than the names use the opts.
func NewParticipant(plainPrename, plainSurname string, secret crypt.Secret, opts ...ParticipantOption) (Participant, error) {
	encryptedPrename, err := crypt.Encrypt(plainPrename, secret)
	if err != nil {
		return Participant{}, err
	}
	encryptedSurname, err := crypt.Encrypt(plainSurname, secret)
	if err != nil {
		return Participant{}, err
	}

	result := &Participant{EncryptedSurname: encryptedSurname, EncryptedPrename: encryptedPrename}

	for _, opt := range opts {
		opt(result)
	}

	return *result, nil
}

func WithParticipantId(id int) ParticipantOption {
	return func(participant *Participant) {
		participant.ID = id
	}
}

func WithSomeCourseId(courseId int64) ParticipantOption {
	return func(participant *Participant) {
		participant.CourseID = sql.NullInt64{Valid: true, Int64: int64(courseId)}
	}
}

func WithNoCourseId() ParticipantOption {
	return func(participant *Participant) {
		participant.CourseID = sql.NullInt64{Valid: false}
	}
}

func EmptyParticipantPointer() *Participant {
	return &Participant{}
}

func (p *Participant) Valid() map[string]string {
	validationErrors := make(map[string]string)

	validateNonEmpty(p.EncryptedSurname, "surname", "Nachname darf nicht leer sein", validationErrors)
	validateNonEmpty(p.EncryptedPrename, "prename", "Vorname darf nicht leer sein", validationErrors)

	p.TrimFields()

	return validationErrors
}

func (p *Participant) Prename(secret crypt.Secret) (string, error) {
	return crypt.Decrypt(p.EncryptedPrename, secret)
}

func (p *Participant) Surname(secret crypt.Secret) (string, error) {
	return crypt.Decrypt(p.EncryptedSurname, secret)
}

func (p *Participant) TrimFields() {
	p.EncryptedPrename = strings.TrimSpace(p.EncryptedPrename)
	p.EncryptedSurname = strings.TrimSpace(p.EncryptedSurname)
}

func (p *Participant) UnmarshalRecord(record []string) error {
	const fn string = "UnmarshalRecord"

	const recordLen int = 4
	if len(record) != recordLen {
		return fmt.Errorf("die Zeile hat %d Werte bzw. Spalten. Genau %d sind erwartet", len(record), recordLen)
	}

	if id, err := strconv.Atoi(record[0]); err == nil {
		p.ID = id
	} else {
		slog.Error(fn+": Could not parse int", "err", err)
		return errors.New(fmt.Sprintf("Spalte: ID\n%s ist keine valide Zahl", record[0]))
	}

	p.EncryptedPrename = record[1]
	p.EncryptedSurname = record[2]

	if record[3] != "null" {
		if courseId, err := strconv.Atoi(record[3]); err == nil {
			p.CourseID = sql.NullInt64{Valid: true, Int64: int64(courseId)}
		} else {
			return err
		}
	}

	return stackValidationErrors(p.Valid())
}

func (p *Participant) MarshalRecord() []string {
	courseIdMarshalled := "null"
	if p.CourseID.Valid {
		courseIdMarshalled = strconv.Itoa(int(p.CourseID.Int64))
	}

	return []string{
		strconv.Itoa(p.ID),
		p.EncryptedPrename,
		p.EncryptedSurname,
		courseIdMarshalled,
	}
}

func ParticipantIds(ps []Participant) (result []int) {
	for _, p := range ps {
		result = append(result, p.ID)
	}
	return
}
