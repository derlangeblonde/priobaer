package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/model/store"
	"softbaer.dev/ass/internal/ui"
)

func ParticipantsNew(c *gin.Context) {
	db := GetDB(c)
	var courses model.Courses
	if err := db.Select("id", "name").Find(&courses).Error; err != nil {
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

	var priorities []model.Course
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&participant).Error; err != nil {
			return err
		}

		if err := store.SetPriorities(tx, participant.ID, req.PrioritizedCourseIDs); err != nil {
			return err
		}

		priorities, err = store.GetPriorities(tx, participant.ID)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		DbError(c, err, "ParticipantsCreate")
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "participants/_show-with-new-button", toViewParticipant(participant, priorities))
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

func toViewParticipant(model model.Participant, priorities []model.Course) ui.Participant {
	result := ui.Participant{
		ID:         model.ID,
		Prename:    model.Prename,
		Surname:    model.Surname,
		Priorities: []ui.Priority{},
	}

	for i, prio := range priorities {
		result.Priorities = append(result.Priorities, ui.Priority{CourseName: prio.Name, Level: uint8(i + 1)})
	}

	return result
}

func toViewParticipants(models []model.Participant, prioritiesById map[int][]model.Course) (results []ui.Participant) {
	for _, m := range models {
		results = append(results, toViewParticipant(m, prioritiesById[m.ID]))
	}
	return
}
