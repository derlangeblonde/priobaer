package app

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/ui"
)

func ScenarioIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
		Solve            bool `form:"solve"`
	}

	db := GetDB(c)
	secret := crypt.GetSecret(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsIndex", "err", err)
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

	scenario, err := domain.LoadScenario(db, secret)
	if err != nil {
		slog.Error("Error while loading scenario", "err", err)
		c.HTML(500, "general/500", gin.H{})
		return
	}

	var selectedParticipants []domain.ParticipantData
	if req.CourseIdSelected == nil {
		selectedParticipants = scenario.Unassigned()
	} else {
		selectedParticipants = scenario.ParticipantsAssignedTo(domain.CourseID(*req.CourseIdSelected))
	}

	viewCourses := scenarioToViewCourses(scenario, pointerToNullable(req.CourseIdSelected), false)
	viewParticipants := toViewParticipants(selectedParticipants, scenario.AllPrioLists())

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "scenario/index", gin.H{"fullPage": false, "participants": viewParticipants, "courseList": viewCourses})

		return
	}

	c.HTML(http.StatusOK, "scenario/index", gin.H{"fullPage": true, "participants": viewParticipants, "courseList": viewCourses})
}

func scenarioToViewCourses(scenario *domain.Scenario, selectedId sql.NullInt64, allAsOobSwap bool) ui.CourseList {
	var courseViews []ui.Course

	for course := range scenario.AllCourses() {
		view := scenarioToViewCourse(course, scenario.AllocationOf(course.ID), selectedId, allAsOobSwap)
		courseViews = append(courseViews, view)
	}

	unassignedCount := len(scenario.Unassigned())
	return ui.CourseList{CourseEntries: courseViews, UnassignedEntry: ui.UnassignedEntry{
		ParticipantsCount: unassignedCount,
		ShouldRender:      true,
		AsOobSwap:         allAsOobSwap,
		Selected:          !selectedId.Valid,
	}}
}

func scenarioToViewCourse(courseData domain.CourseData, allocation int, selectedId sql.NullInt64, asOobSwap bool) ui.Course {
	return ui.Course{
		ID:          int(courseData.ID),
		Name:        courseData.Name,
		MinCapacity: courseData.MinCapacity,
		MaxCapacity: courseData.MaxCapacity,
		Selected:    selectedId.Valid && int(courseData.ID) == int(selectedId.Int64),
		Allocation:  allocation,
		AsOobSwap:   asOobSwap,
	}
}
