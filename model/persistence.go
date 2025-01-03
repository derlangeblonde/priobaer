package model

import "gorm.io/gorm"

func ApplyAssignment(db *gorm.DB, assignment Assignment) error {
	if result := db.Model(Participant{}).Where("ID = ?", assignment.Participant.ID).Update("course_id", assignment.Course.ID); result.Error != nil {
		return result.Error
	}

	return nil
}

func ApplyAssignments(db *gorm.DB, assignments []Assignment) error {
	for _, assignment := range assignments {
		if err := ApplyAssignment(db, assignment); err != nil {
			return err
		}
	}

	return nil
}
