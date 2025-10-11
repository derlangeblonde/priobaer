package app

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/domain/store"
	"softbaer.dev/ass/internal/model"
)

func ScenarioIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
		Solve            bool `form:"solve"`
	}
	fnName := "ScenarioIndex"

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on ScenarioIndex", "err", err)
		return
	}

	if req.Solve {
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
			slog.Info("Could not solve assignment", "err", err)
			c.HTML(http.StatusOK, "dialogs/not-solvable", gin.H{})

			return
		}

		if err != nil {
			slog.Error("Error while trying to solve assignment", "err", err)
			c.AbortWithStatus(500)

			return
		}
	}

	var courses []model.Course
	err = db.Preload("Participants").Find(&courses).Error

	if err != nil {
		DbError(c, err, fnName)
		return
	}

	var participants []model.Participant
	err = db.Where("course_id is null").Find(&participants).Error
	unassignedCount := len(participants)
	participantsSet := false
	// TODO: logic in this block and the block above might be overly complex
	// Let's see if we find sth more simple.
	if req.CourseIdSelected != nil {
		for _, course := range courses {
			if course.ID == *req.CourseIdSelected {
				participants = course.Participants
				participantsSet = true
				break
			}
		}
	}

	if !participantsSet && req.CourseIdSelected != nil {
		slog.Error("Could not find course with id. Defaulting to unassigned", "course_id", req.CourseIdSelected)
		req.CourseIdSelected = nil
	}

	viewCourses := toViewCourses(courses, pointerToNullable(req.CourseIdSelected), false)
	viewCourses.UnassignedEntry.Selected = req.CourseIdSelected == nil
	viewCourses.UnassignedEntry.ParticipantsCount = unassignedCount
	viewCourses.UnassignedEntry.ShouldRender = true

	prioritiesByParticipantIds, err := store.GetPrioritiesForMultiple(db, model.ParticipantIds(participants))

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": false, "participants": toViewParticipants(participants, prioritiesByParticipantIds), "courseList": viewCourses})

		return
	}

	c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": true, "participants": toViewParticipants(participants, prioritiesByParticipantIds), "courseList": viewCourses})
}
