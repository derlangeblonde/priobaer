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
	c.Redirect(http.StatusSeeOther, "/assignments")
}

