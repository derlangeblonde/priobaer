package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func LoadScenario(db *gorm.DB) (scenario *Scenario, err error) {
	scenario = EmptyScenario()
	var participants []model.Participant
	var courses []model.Course

	if err := db.Find(&participants).Error; err != nil {
		return nil, err
	}
	if err := db.Find(&courses).Error; err != nil {
		return nil, err
	}
	scenario.participants = ParticipantsFromDbModel(participants)
	scenario.courses = CoursesFromDbModels(courses)

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

func OverwriteScenario(db *gorm.DB, scenario *Scenario) error {
	modelsToDelete := []any{
		&model.Priority{},
		&model.Participant{},
		&model.Course{},
	}
	for _, model := range modelsToDelete {
		if err := db.Unscoped().Delete(model, "deleted_at is null").Error; err != nil {
			return err
		}

	}

	courseRecords := CoursesToDbModels(scenario.courses)
	db.CreateInBatches(courseRecords, 100)

	participantRecords := scenario.allParticipantsAsDbModels()
	db.CreateInBatches(participantRecords, 100)

	savePriorities(db, scenario.AllPriorities())

	return nil
}
