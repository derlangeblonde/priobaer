package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

const batchSize = 100

func savePriorities(db *gorm.DB, prios []PriorityData) error {
	records := make([]model.Priority, len(prios))

	for i, prio := range prios {
		records[i] = model.Priority{Level: model.PriorityLevel(prio.Level), CourseID: int(prio.CourseID), ParticipantID: int(prio.ParticipantID)}
	}

	return db.CreateInBatches(records, batchSize).Error
}
