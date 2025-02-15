package store

import (
	"fmt"

	"gorm.io/gorm"
	"softbaer.dev/ass/model"
)


func SetPriorities(tx *gorm.DB, participantID int, courseIDs []int) error {
	if len(courseIDs) > model.MaxPriorityLevel {
		return fmt.Errorf("Die Priorität in Höhe von %d übersteigt das Maximum von %d", len(courseIDs), model.MaxPriorityLevel)
	}

	var priorities []model.Priority

	for i, courseID := range courseIDs {
		priority := model.Priority{Level: model.PriorityLevel(i + 1), ParticipantID: participantID, CourseID: courseID}
		priorities = append(priorities, priority)
	}

	if err := tx.Where("participant_id = ?", participantID).Delete(&model.Priority{}).Error; err != nil {
		return model.DefaultDbError(err)
	}

	if err := tx.CreateInBatches(&priorities, model.MaxPriorityLevel).Error; err != nil {
		return model.DefaultDbError(err) 
	}

	return nil
}
