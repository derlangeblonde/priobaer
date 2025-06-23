package domain

import (
	"iter"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

const batchSize = 100

func savePriorities(db *gorm.DB, prios iter.Seq[Priority]) error {
	var records []model.Priority

	for prio := range prios {
		record := model.Priority{Level: model.PriorityLevel(prio.Level), CourseID: int(prio.Course.ID), ParticipantID: int(prio.Participant.ID)}

		records = append(records, record)
	}

	return db.CreateInBatches(records, batchSize).Error
}
