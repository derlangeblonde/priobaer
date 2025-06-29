package domain

import "softbaer.dev/ass/internal/model"

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
