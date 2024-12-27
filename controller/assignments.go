package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/view"
)

func AssignmentsIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsIndex", "err", err)
		return
	}

	participants := make([]model.Participant, 0)
	var result *gorm.DB

	if req.CourseIdSelected == nil {
		result = db.Where("course_id is null").Find(&participants)
	} else {
		courseID := *req.CourseIdSelected
		result = db.Where("course_id = ?", courseID).Find(&participants)
	}

	if result.Error != nil {
		slog.Error("Unexpected error while gettting assigned particpants from db", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	var courses []model.Course
	result = db.Model(&model.Course{}).Preload("Participants").Find(&courses)

	if result.Error != nil {
		slog.Error("Unexpected error while getting all courses from db", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	noCourseSelected := req.CourseIdSelected == nil
	viewCourses := toViewCourses(courses, req.CourseIdSelected)

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": false, "participants": participants, "courses": viewCourses, "noCourseSelected": noCourseSelected})

		return
	}

	c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": true, "participants": participants, "courses": viewCourses, "noCourseSelected": noCourseSelected})
}

func AssignmentsUpdate(c *gin.Context) {
	type request struct {
		ParticipantId int `form:"participant-id" binding:"required"`
		CourseId      int `form:"course-id"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsUpdate", "err", err)
		return
	}

	var participant model.Participant
	result := db.Preload("Course").First(&participant, req.ParticipantId)

	courseUnassigned := participant.Course

	if result.Error != nil {
		slog.Error("Unexpected error in AssignmentUpdate while fetching participant from db", "err", result.Error)
	}

	if req.CourseId == 0 {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", nil)
	} else {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", req.CourseId)
	}

	if result.Error != nil {
		slog.Error("Unexpected error while updating assignment relation", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	var courseAssigned model.Course
	result = db.Preload("Participants").First(&courseAssigned, req.CourseId)

	if result.Error != nil {
		slog.Error("Error when querying for courseAssigned", "err", result.Error)
	}

	result = db.Preload("Participants").First(&courseUnassigned, courseUnassigned.ID)

	if result.Error != nil {
		slog.Error("Unexpected error in AssignmentUpdate while courseUnassigned from db", "err", result.Error)
	}

	viewUpdates := []view.Course{
		toViewCourse(courseAssigned, &courseUnassigned.ID, true),
		toViewCourse(courseUnassigned, &courseUnassigned.ID, true),
	}

	c.HTML(http.StatusOK, "courses", viewUpdates)

}

func toViewCourses(models []model.Course, selectedId *int) (views []view.Course) {
	for _, model := range models {
		view := toViewCourse(model, selectedId, false)
		views = append(views, view)
	}

	return
}

func toViewCourse(model model.Course, selectedId *int, asOobSwap bool) view.Course {
	return view.Course{
		ID:          model.ID,
		Name:        model.Name,
		MinCapacity: model.MinCapacity,
		MaxCapacity: model.MaxCapacity,
		Selected:    selectedId != nil && model.ID == *selectedId,
		Allocation:  model.Allocation(),
		AsOobSwap:   asOobSwap,
	}
}
