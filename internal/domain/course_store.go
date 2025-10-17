package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func FindSingleCourseData(db *gorm.DB, cid CourseID) (CourseData, error) {
	var model model.Course
	if err := db.First(&model, "id = ?", cid).Error; err != nil {
		return CourseData{}, err
	}

	return courseFromDbModel(model), nil
}

func findCourseDataById(db *gorm.DB, cids []CourseID) (result []CourseData, err error) {
	var models []model.Course
	if err := db.Where("id IN ?", cids).Find(&models).Error; err != nil {
		return result, err
	}

	return coursesFromDbModels(models), nil
}

// DeleteCourse deletes the course with the specified id together with all existing associations.
// I.e. participants assigned to that course will be unassigned and priorities to that courses will be deleted.
// Prefer passing a transaction, so that partial changes will be rolled back in case of an error.
func DeleteCourse(tx *gorm.DB, courseId int) error {
	err := tx.Model(model.EmptyParticipantPointer()).Where("course_id = ?", courseId).Update("course_id", nil).Error
	if err != nil {
		return err
	}

	if err := tx.Unscoped().Delete(&model.Priority{}, "course_id = ?", courseId).Error; err != nil {
		return err
	}

	course := model.Course{ID: courseId}
	return tx.Unscoped().Delete(&course).Error
}

func courseFromDbModel(model model.Course) CourseData {
	return CourseData{ID: CourseID(model.ID),
		Name:        model.Name,
		MaxCapacity: model.MaxCapacity,
		MinCapacity: model.MinCapacity,
	}
}

func coursesFromDbModels(models []model.Course) []CourseData {
	courses := make([]CourseData, len(models))
	for i, dbCourse := range models {
		courses[i] = courseFromDbModel(dbCourse)
	}
	return courses
}

func courseToDbModel(course CourseData) model.Course {
	return model.Course{
		ID:          int(course.ID),
		Name:        course.Name,
		MaxCapacity: course.MaxCapacity,
		MinCapacity: course.MinCapacity,
	}
}

func coursesToDbModels(courses []CourseData) []model.Course {
	dbModels := make([]model.Course, len(courses))
	for i, course := range courses {
		dbModels[i] = courseToDbModel(course)
	}
	return dbModels
}

func toCourseIds(ids []int) []CourseID {
	courseIds := make([]CourseID, len(ids))
	for i, id := range ids {
		courseIds[i] = CourseID(id)
	}
	return courseIds
}
