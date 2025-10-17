package solve

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func applyAssignments(tx *gorm.DB, assignments []computedAssignment) error {
	for _, assignment := range assignments {
		if err := tx.Model(model.EmptyParticipantPointer()).Where("ID = ?", assignment.participantID).Update("course_id", assignment.courseID).Error; err != nil {
			return err
		}
	}

	return nil
}
