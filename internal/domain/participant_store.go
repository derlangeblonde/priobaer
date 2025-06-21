package domain

import "softbaer.dev/ass/internal/model"

func ParticipantFromDbModel(dbModel model.Participant) Participant {
	return Participant{
		ID:      ParticipantID(dbModel.ID),
		Prename: dbModel.Prename,
		Surname: dbModel.Surname,
	}
}

func ParticipantsFromDbModel(dbModels []model.Participant) []Participant {
	participants := make([]Participant, len(dbModels))
	for i, dbModel := range dbModels {
		participants[i] = ParticipantFromDbModel(dbModel)
	}
	return participants
}
