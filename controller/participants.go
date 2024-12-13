package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
)

func ParticipantsIndex(c *gin.Context) {
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

func ParticipantsNew(c *gin.Context) {
	c.HTML(http.StatusOK, "participants/_new", nil)
}

func ParticipantsCreate(c *gin.Context) {
	type request struct {
		Prename string `form:"prename" binding:"required"`
		Surname string `form:"surname" binding:"required"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		return
	}

	participant := model.Participant{Prename: req.Prename, Surname: req.Surname}
	result := db.Create(&participant)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrCheckConstraintViolated) {
			slog.Error("Constraint violated while creating particpiant",
				"err", err,
				"particpaints.Prename", participant.Prename,
				"participants.Surname", participant.Surname)
			c.AbortWithStatus(http.StatusConflict)

			return
		}

		slog.Error("Unexpected error while creating participant", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "participants/_show-with-new-button", participant)
	} else {
		c.Redirect(http.StatusSeeOther, "/assignments")
	}
}

func ParticipantsDelete(c *gin.Context) {
	type request struct {
		ID uint `uri:"id" binding:"required"`
	}
	db := GetDB(c)

	var req request
	err := c.BindUri(&req)

	if err != nil {
		slog.Error("Could not parse id from uri in ParticipantsDelete", "err", err)
		c.AbortWithStatus(http.StatusNotFound)

		return
	}

	participant := model.Participant{ID: int(req.ID)}
	result := db.Delete(&participant)

	if result.Error != nil {
		slog.Error("Delete of participants failed on db level", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	c.Data(http.StatusOK, "text/html", []byte(""))
}
