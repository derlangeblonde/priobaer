package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func CountUnassigned(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(model.EmptyParticipantPointer()).Where("course_id is null").Count(&count).Error

	return int(count), err
}

func CountAllocation(db *gorm.DB, cid CourseID) (int, error) {
	var count int64
	err := db.Model(model.EmptyParticipantPointer()).Where("course_id = ?", cid).Count(&count).Error

	return int(count), err
}
