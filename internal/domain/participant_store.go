package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

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

// DeleteParticipant Deletes a participant together with all associations that require the participant.
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
