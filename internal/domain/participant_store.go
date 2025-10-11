package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/model"
)

func participantDataFromDbModel(dbModel model.Participant, secret crypt.Secret) (ParticipantData, error) {
	encryptName := encryptedParticipantName{
		Prename: dbModel.Prename,
		Surname: dbModel.Surname,
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
