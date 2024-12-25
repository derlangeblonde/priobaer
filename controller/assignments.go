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
	result = db.Model(&model.Course{}).Preload("Participants").Find(&courses).Debug()

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
		CoureseId     int `form:"course-id"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsUpdate", "err", err)
		return
	}

	var result *gorm.DB

	if req.CoureseId == 0 {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", nil)
	} else {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", req.CoureseId)
	}

	if result.Error != nil {
		slog.Error("Unexpected error while updating assignment relation", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	c.Data(http.StatusOK, "text/html", []byte(""))
}

func toViewCourses(models []model.Course, selectedId *int) (views []view.Course) {
	for _, model := range models {
		view := toViewCourse(model, selectedId, false)
		views = append(views, view)
	}

	return
}

func toViewCourse(model model.Course, selectedId *int, hxSwap bool) view.Course {
	return view.Course{
		ID:          model.ID,
		Name:        model.Name,
		MinCapacity: model.MinCapacity,
		MaxCapacity: model.MaxCapacity,
		Selected:    selectedId != nil && model.ID == *selectedId,
		Allocation:  model.Allocation(),
		HxSwap: hxSwap,
	}
}
