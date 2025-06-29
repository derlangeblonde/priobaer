package domain

import "softbaer.dev/ass/internal/model"

func CourseFromDbModel(model model.Course) CourseData {
	return CourseData{ID: CourseID(model.ID),
		Name:        model.Name,
		MaxCapacity: model.MaxCapacity,
		MinCapacity: model.MinCapacity,
	}
}

func CoursesFromDbModels(models []model.Course) []CourseData {
	courses := make([]CourseData, len(models))
	for i, model := range models {
		courses[i] = CourseFromDbModel(model)
	}
	return courses
}

func CourseToDbModel(course CourseData) model.Course {
	return model.Course{
		ID:          int(course.ID),
		Name:        course.Name,
		MaxCapacity: course.MaxCapacity,
		MinCapacity: course.MinCapacity,
	}
}

func CoursesToDbModels(courses []CourseData) []model.Course {
	dbModels := make([]model.Course, len(courses))
	for i, course := range courses {
		dbModels[i] = CourseToDbModel(course)
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
