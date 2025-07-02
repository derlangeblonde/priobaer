package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"softbaer.dev/ass/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
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
		SelectedCourseID     *int   `form:"course-id"`
	}

	db := GetDB(c)

	var req request
	if err := c.Bind(&req); err != nil {
		return
	}

	if len(req.PrioritizedCourseIDs) > model.MaxPriorityLevel {
		c.HTML(422,
			"participants/_new",
			gin.H{"Errors": map[string]string{"priorities": fmt.Sprintf("Maximale Anzahl an Prioritäten (%d) überschritten", len(req.PrioritizedCourseIDs))}},
		)
		return
	}

	candidate := domain.NewParticipantCandidate(req.Prename, req.Surname)
	candidate.Prioritize(req.PrioritizedCourseIDs)
	candidate.Assign(req.SelectedCourseID)
	validationErrors := candidate.Valid()

	if len(validationErrors) > 0 {
		c.HTML(422, "participants/_new", gin.H{"Errors": validationErrors, "Value": candidateToViewParticipant(*candidate)})

		return
	}

	var createdParticipant domain.Participant
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		createdParticipant, err = candidate.Save(tx)
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
		c.HTML(http.StatusOK, "participants/_show-with-new-button", domainToViewParticipant(createdParticipant))
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

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.DeleteParticipant(db, domain.ParticipantID(req.ID))
	})

	if err != nil {
		slog.Error("Database error occurred when trying to delete participant", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	c.Data(http.StatusOK, "text/html", []byte(""))
}

func ParticipantsButtonNew(c *gin.Context) {
	c.HTML(http.StatusOK, "participants/_new-button", nil)
}

// Copies toViewParticipant, which can be removed eventually, when domain refactoring is done.
func candidateToViewParticipant(model domain.ParticipantCandidate) ui.Participant {
	result := ui.Participant{
		Prename: model.Prename,
		Surname: model.Surname,
	}

	return result
}

// Copies toViewParticipant, which can be removed eventually, when domain refactoring is done.
func domainToViewParticipant(participant domain.Participant) ui.Participant {
	result := ui.Participant{
		ID:         int(participant.ID),
		Prename:    participant.Prename,
		Surname:    participant.Surname,
		Priorities: make([]ui.Priority, len(participant.PrioritizedCourses)),
	}

	for i, prio := range participant.PrioritizedCourses {
		result.Priorities[i] = ui.Priority{CourseName: prio.Name, Level: uint8(i + 1)}
	}

	return result
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
