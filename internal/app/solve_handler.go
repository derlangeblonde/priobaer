package app

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func SolveAssignments(c *gin.Context) {
	logger := slog.With("Func", "SolveAssignments")
	db := GetDB(c)

	err := db.Transaction(
		func(tx *gorm.DB) error {
			var availableCourses []model.Course
			if result := tx.Preload("Participants").Find(&availableCourses); result.Error != nil {
				return result.Error
			}

			var unassignedParticipants []model.Participant
			if result := tx.Where("course_id is null").Find(&unassignedParticipants); result.Error != nil {
				return result.Error
			}

			var relevantPriorities []model.Priority
			// TODO: optimize querying
			if result := tx.Preload("Participant").Preload("Course").Where("participant_id in ?", model.ParticipantIds(unassignedParticipants)).Find(&relevantPriorities); result.Error != nil {
				return result.Error
			}

			assignments, err := model.SolveAssignment(relevantPriorities)
			if err != nil {
				return err
			}

			err = model.ApplyAssignments(tx, assignments)
			if err != nil {
				return err
			}

			return nil
		},
	)

	if errors.Is(err, model.NotSolvable) {
		logger.Info("Could not solve assignment", "err", err)
		c.HTML(http.StatusOK, "dialogs/not-solvable", gin.H{})

		return
	}

	if err != nil {
		logger.Error("Error while trying to solve assignment", "err", err)
		c.AbortWithStatus(500)

		return
	}

	c.Redirect(http.StatusSeeOther, "/scenario")
}
