package app

import (
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/ui"
)

func candidateToViewParticipant(model domain.ParticipantCandidate) ui.Participant {
	result := ui.Participant{
		Prename: model.Prename,
		Surname: model.Surname,
	}

	return result
}

func domainToViewParticipant(participant domain.Participant) ui.Participant {
	result := ui.Participant{
		ID:         int(participant.ID),
		Prename:    participant.Prename,
		Surname:    participant.Surname,
		Priorities: make([]ui.Priority, len(participant.PrioritizedCourses)),
	}

	for i, prio := range participant.PrioritizedCourses {
		result.Priorities[i] = ui.Priority{CourseName: prio.Name, Level: uint8(i + 1)}
	}

	return result
}

func toViewParticipant(model domain.ParticipantData, priorities []domain.CourseData) ui.Participant {
	result := ui.Participant{
		ID:         int(model.ID),
		Prename:    model.Prename,
		Surname:    model.Surname,
		Priorities: []ui.Priority{},
	}

	for i, prio := range priorities {
		result.Priorities = append(result.Priorities, ui.Priority{CourseName: prio.Name, Level: uint8(i + 1)})
	}

	return result
}

func toViewParticipants(participants []domain.ParticipantData, prioritiesById map[domain.ParticipantID][]domain.CourseData) (results []ui.Participant) {
	for _, participant := range participants {
		results = append(results, toViewParticipant(participant, prioritiesById[participant.ID]))
	}

	return
}
