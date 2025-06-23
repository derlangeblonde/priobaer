package store

import (
	"fmt"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func SetPriorities(tx *gorm.DB, participantID int, courseIDs []int) error {
	if len(courseIDs) > model.MaxPriorityLevel {
		return fmt.Errorf("die Priorität in Höhe von %d übersteigt das Maximum von %d", len(courseIDs), model.MaxPriorityLevel)
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

func GetPriorities(tx *gorm.DB, participantID int) (courses []model.Course, err error) {
	var priorities model.Priorities
	err = tx.Select("course_id").Where("participant_id = ?", participantID).Order("level ASC").Find(&priorities).Error
	if err != nil {
		return
	}

	if err = tx.Select("id, name").Where("id IN ?", priorities.CourseIDs()).Find(&courses).Error; err != nil {
		return
	}

	return courses, nil
}

func GetPrioritiesForMultiple(tx *gorm.DB, participantIDs []int) (map[int][]model.Course, error) {
	result := make(map[int][]model.Course)
	var priorities model.Priorities
	err := tx.
		Preload("Course").
		Select("course_id, participant_id, level").
		Where("participant_id in ?", participantIDs).
		Order("level ASC").
		Find(&priorities).Error

	if err != nil {
		return result, err
	}

	for _, priority := range priorities {
		result[priority.ParticipantID] = append(result[priority.ParticipantID], priority.Course)
	}

	return result, nil
}
