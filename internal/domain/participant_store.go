package domain

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/model"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
	ErrCourseNotFound      = errors.New("course not found")
)

func FindAssignedCourse(db *gorm.DB, pid ParticipantID) (CourseData, error) {
	var courseID int64
	if err := db.Model(model.EmptyParticipantPointer()).Where("id = ?", pid).Select("course_id").First(&courseID).Error; err != nil {
		return CourseData{}, err
	}

	return FindSingleCourseData(db, CourseID(courseID))
}

func InitialAssign(tx *gorm.DB, pid ParticipantID, cid CourseID) error {
	result := tx.Model(model.EmptyParticipantPointer()).Where("id = ?", pid).Update("course_id", cid)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "FOREIGN KEY constraint failed") {
			return ErrParticipantNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

func Reassign(tx *gorm.DB, pid ParticipantID, cid CourseID) error {
	result := tx.Model(model.EmptyParticipantPointer()).Where("ID = ?", pid).Update("course_id", cid)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "FOREIGN KEY constraint failed") {
			return ErrParticipantNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

func Unassign(tx *gorm.DB, pid ParticipantID) error {
	result := tx.Model(model.EmptyParticipantPointer()).Where("ID = ?", pid).Update("course_id", nil)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}

	return nil
}

// DeleteParticipant deletes a participant together with all associations that require the participant.
// Parameter tx should be a transaction. Otherwise, we could delete some but not all data.
func DeleteParticipant(tx *gorm.DB, ParticipantID ParticipantID) error {
	if err := tx.Unscoped().Where("participant_id = ?", int(ParticipantID)).Delete(&model.Priority{}).Error; err != nil {
		return err
	}

	if err := tx.Unscoped().Where("id = ?", int(ParticipantID)).Delete(model.EmptyParticipantPointer()).Error; err != nil {
		return err
	}

	return nil
}

func participantDataFromDbModel(dbModel model.Participant, secret crypt.Secret) (ParticipantData, error) {
	encryptName := encryptedParticipantName{
		Prename: dbModel.EncryptedPrename,
		Surname: dbModel.EncryptedSurname,
	}
	decryptedName, err := encryptName.decrypt(secret)
	if err != nil {
		return ParticipantData{}, err
	}

	return ParticipantData{
		ID:              ParticipantID(dbModel.ID),
		ParticipantName: decryptedName,
	}, nil
}

func participantsFromDbModel(dbModels []model.Participant, secret crypt.Secret) ([]ParticipantData, error) {
	var participants []ParticipantData
	for _, dbModel := range dbModels {
		participant, err := participantDataFromDbModel(dbModel, secret)
		if err != nil {
			return participants, err
		}
		participants = append(participants, participant)
	}

	return participants, nil
}
