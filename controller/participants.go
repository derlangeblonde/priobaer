package controller

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/model/store"
	"softbaer.dev/ass/view"
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
		c.HTML(http.StatusOK, "participants/index", gin.H{"fullPage": false, "participants": toViewParticipants(participants, make(map[int][]model.Course))})
	} else {
		c.HTML(http.StatusOK, "participants/index", gin.H{"fullPage": true, "participants": toViewParticipants(participants, make(map[int][]model.Course))})
	}

}

func ParticipantsNew(c *gin.Context) {
	db := GetDB(c)
	var courses model.Courses
	if err := db.Select("id", "name").Find(&courses).Error; err != nil {
		// TODO: für sqlite3.Error vereinheitlichen!
		// Ich möchte einen Mechanismus, sodass ich mit wenig Boilerplate und mental-overhead
		// (automatisch) für den Typ sqlite3.Error einen Dialog im Fronted rendere.
		// Dialog mit einer entsprechend generischen Fehlermeldung.
		DbError(c, err, "ParticipantsNew")

		return
	}

	c.HTML(http.StatusOK, "participants/_new", gin.H{"Errors": make(map[string]string, 0), "Courses": courses})
}

func ParticipantsCreate(c *gin.Context) {
	type request struct {
		Prename              string `form:"prename"`
		Surname              string `form:"surname"`
		PrioritizedCourseIDs []int  `form:"prio[]"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		return
	}

	if len(req.PrioritizedCourseIDs) > model.MaxPriorityLevel {
		c.HTML(422,
			"participants/_new",
			gin.H{"Errors": map[string]string{"priorities": fmt.Sprintf("Maximale Anzahl an Prioritäten (%d) überschritten", len(req.PrioritizedCourseIDs))}},
		)
		return
	}

	participant := model.Participant{Prename: req.Prename, Surname: req.Surname}
	validationErrors := participant.Valid()

	if len(validationErrors) > 0 {
		c.HTML(422, "participants/_new", gin.H{"Errors": validationErrors, "Value": toViewParticipant(participant, make([]model.Course, 0))})

		return
	}

	for i, courseID := range req.PrioritizedCourseIDs {
		prio := model.Priority{CourseID: courseID, Level: model.PriorityLevel(i + 1)}
		participant.Priorities = append(participant.Priorities, prio)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&participant)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrCheckConstraintViolated) {
				slog.Error("Constraint violated while creating particpiant",
					"err", err,
					"particpaints.ID", participant.ID,
				)
				c.AbortWithStatus(http.StatusConflict)

				return result.Error
			}

			slog.Error("Unexpected error while creating participant", "err", result.Error)
			c.AbortWithStatus(http.StatusInternalServerError)

			return result.Error
		}

		if err := store.PopulatePrioritizedCourseNames(tx, &participant); err != nil {
			DbError(c, err, "ParticipantsCreate")
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "participants/_show-with-new-button", toViewParticipant(participant, make([]model.Course, 0)))
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
	result := db.Unscoped().Delete(&participant)
	// TODO:
	// 2025/02/28 00:04:31 /home/joni/dev/ass/controller/participants.go:147 FOREIGN KEY constraint failed
	// [0.681ms] [rows:0] DELETE FROM `participants` WHERE `participants`.`id` = 5
	// 2025/02/28 00:04:31 ERROR Delete of participants failed on db level err="FOREIGN KEY constraint failed"
	// [GIN-debug] [WARNING] Headers were already written. Wanted to override status code 500 with 200


	if result.Error != nil {
		slog.Error("Delete of participants failed on db level", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

	}

	c.Data(http.StatusOK, "text/html", []byte(""))
}

func ParticipantsButtonNew(c *gin.Context) {
	c.HTML(http.StatusOK, "participants/_new-button", nil)
}

func toViewParticipant(model model.Participant, priorities []model.Course) view.Participant{
	result := view.Participant{
		ID:         model.ID,
		Prename:    model.Prename,
		Surname:    model.Surname,
		Priorities: []view.Priority{},
	}

	for i, prio := range priorities{
		result.Priorities = append(result.Priorities, view.Priority{CourseName: prio.Name, Level: uint8(i + 1)})
	}

	return result
}

func toViewParticipants(models []model.Participant, prioritiesById map[int][]model.Course) (results []view.Participant) {
	for _, model := range models {
		results = append(results, toViewParticipant(model, prioritiesById[model.ID]))
	}
	return
}
