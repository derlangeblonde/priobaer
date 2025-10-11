package domain

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
	ErrCourseNotFound      = errors.New("course not found")
)

func InitialAssign(tx *gorm.DB, pid ParticipantID, cid CourseID) error {
	result := tx.Model(model.Participant{}).Where("ID = ?", pid).Update("course_id", cid)

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

func ParticipantDataFromDbModel(dbModel model.Participant) ParticipantData {
	return ParticipantData{
		ID: ParticipantID(dbModel.ID),
		ParticipantName: ParticipantName{
			Prename: dbModel.Prename,
			Surname: dbModel.Surname,
		},
	}
}

func ParticipantsFromDbModel(dbModels []model.Participant) []ParticipantData {
	participants := make([]ParticipantData, len(dbModels))
	for i, dbModel := range dbModels {
		participants[i] = ParticipantDataFromDbModel(dbModel)
	}
	return participants
}

// DeleteParticipant deletes a participant together with all associations that require the participant.
// Parameter tx should be a transaction. Otherwise, we could delete some but not all data.
func DeleteParticipant(tx *gorm.DB, ParticipantID ParticipantID) error {
	if err := tx.Unscoped().Where("participant_id = ?", int(ParticipantID)).Delete(&model.Priority{}).Error; err != nil {
		return err
	}

	if err := tx.Unscoped().Where("id = ?", int(ParticipantID)).Delete(&model.Participant{}).Error; err != nil {
		return err
	}

	return nil
}
