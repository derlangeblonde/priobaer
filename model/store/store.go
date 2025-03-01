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

func GetPriorities(tx *gorm.DB, participantID int) (courseIDs []int, err error) {
	var priorities model.Priorities
	err = tx.Select("course_id").Where("participant_id = ?", participantID).Order("level ASC").Find(&priorities).Error
	if err != nil {
		return 
	}

	return priorities.CourseIDs(), nil
}

func PopulatePrioritizedCourseNames(tx *gorm.DB, participant *model.Participant) error {
	var courses []model.Course
	if err := tx.Select("id", "name").Where("id IN ?", participant.PrioritizedCourseIDs()).Find(&courses).Error; err != nil {
		return err
	}

	for i := range participant.Priorities {
		for _, course := range courses {
			if course.ID == participant.Priorities[i].CourseID {
				participant.Priorities[i].Course.Name = course.Name
				break
			}
		}
	}

	return nil
}
