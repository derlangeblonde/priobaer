package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
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
		result = db.Where("course_id is null").Find(&participants).Debug()
	} else {
		courseID := *req.CourseIdSelected
		result = db.Where("course_id = ?", courseID).Find(&participants).Debug()
	}

	if result.Error != nil {
		slog.Error("Unexpected error while showing course index", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "participants/index", gin.H{"fullPage": false, "participants": participants})
	} else {
		c.HTML(http.StatusOK, "participants/index", gin.H{"fullPage": true, "participants": participants})
	}

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

	c.Redirect(http.StatusSeeOther, "/assignments")
}
