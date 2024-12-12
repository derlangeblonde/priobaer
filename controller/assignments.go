package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/model"
)

func AssignmentsIndex(c *gin.Context) {
	db := GetDB(c)

	participants := make([]model.Participant, 0)
	result := db.Find(&participants)

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

	participant := model.Participant{ID: req.ParticipantId}
	db.Delete(&participant)

	c.Redirect(http.StatusSeeOther, "/assignments")
}
