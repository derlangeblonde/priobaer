package app

import (
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/ui"
)

func newSelectedUiCourse(courseData domain.CourseData, allocation int) ui.Course {
	result := newUiCourse(courseData, allocation)
	result.Selected = true

	return result
}

func newUiCourse(courseData domain.CourseData, allocation int) ui.Course {
	return ui.Course{
		ID:          int(courseData.ID),
		Name:        courseData.Name,
		MaxCapacity: courseData.MaxCapacity,
		MinCapacity: courseData.MinCapacity,
		Allocation:  allocation,
	}
}
