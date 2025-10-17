package domain

import (
	"slices"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/model"
)

func LoadScenario(db *gorm.DB, secret crypt.Secret) (scenario *Scenario, err error) {
	scenario = EmptyScenario()
	var participants []model.Participant
	var courses []model.Course

	if err := db.Find(&participants).Error; err != nil {
		return nil, err
	}
	if err := db.Find(&courses).Error; err != nil {
		return nil, err
	}
	if scenario.participants, err = participantsFromDbModel(participants, secret); err != nil {
		return nil, err
	}
	scenario.courses = coursesFromDbModels(courses)

	for _, participant := range participants {
		if participant.CourseID.Valid {
			if err := scenario.Assign(ParticipantID(participant.ID), CourseID(participant.CourseID.Int64)); err != nil {
				return nil, err
			}
		}
	}

	var priorities []model.Priority
	db.Find(&priorities)

	priosPerParticipantId := make(map[int][]int)
	for _, prio := range priorities {
		priosPerParticipantId[prio.ParticipantID] = append(priosPerParticipantId[prio.ParticipantID], prio.CourseID)
	}

	for participantId, prios := range priosPerParticipantId {
		if err = scenario.Prioritize(ParticipantID(participantId), toCourseIds(prios)); err != nil {
			return nil, err
		}
	}

	return
}

func OverwriteScenario(db *gorm.DB, scenario *Scenario, secret crypt.Secret) error {
	tablesToDelete := []any{
		&model.Priority{},
		model.EmptyParticipantPointer(),
		&model.Course{},
	}
	for _, table := range tablesToDelete {
		if err := db.Unscoped().Delete(table, "deleted_at is null").Error; err != nil {
			return err
		}

	}

	courseRecords := coursesToDbModels(scenario.courses)
	db.CreateInBatches(courseRecords, 100)

	participantRecords, err := scenario.allParticipantsAsDbModels(secret)
	if err != nil {
		return err
	}

	db.CreateInBatches(participantRecords, 100)

	err = savePriorities(db, slices.Collect(scenario.AllPriorities()))
	if err != nil {
		return err
	}

	return nil
}
