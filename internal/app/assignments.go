package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/model/store"
	"softbaer.dev/ass/internal/ui"
)

func AssignmentsIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
		Solve            bool `form:"solve"`
	}
	fnName := "AssignmentsIndex"

	db := GetDB(c)

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

type assignmentUpdateRequest struct {
	ParticipantId int `form:"participant-id" binding:"required"`
	CourseId      int `form:"course-id"`
}

func (r *assignmentUpdateRequest) IsUnassign() bool {
	return r.CourseId == 0
}

func AssignmentsUpdate(c *gin.Context) {
	db := GetDB(c)

	var req assignmentUpdateRequest
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsUpdate", "err", err)
		return
	}

	var participant model.Participant
	var courseUnassigned, courseAssigned model.Course
	var coursesToUpdate []model.Course
	var updateUnassignedEntry bool

	result := db.First(&participant, req.ParticipantId)

	courseIdUnassigned := participant.CourseID

	if result.Error != nil {
		slog.Error("Unexpected error in AssignmentUpdate while fetching participant from db", "err", result.Error)
		c.AbortWithStatus(500)

		return
	}

	if req.IsUnassign() {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", nil)
		updateUnassignedEntry = true
	} else {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", req.CourseId)

		if result.Error == nil {
			result = db.Preload("Participants").First(&courseAssigned, req.CourseId)
		}

		if result.Error == nil {
			coursesToUpdate = append(coursesToUpdate, courseAssigned)
		}
	}

	if result.Error != nil {
		slog.Error("Unexpected error while updating assignment relation", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	if courseIdUnassigned.Valid {
		result = db.Preload("Participants").First(&courseUnassigned, courseIdUnassigned)

		if result.Error != nil {
			slog.Error("Unexpected error in AssignmentUpdate while courseUnassigned from db", "err", result.Error)
			c.AbortWithStatus(500)

			return
		}

		coursesToUpdate = append(coursesToUpdate, courseUnassigned)
	} else {
		updateUnassignedEntry = true
	}

	viewUpdates := toViewCourses(coursesToUpdate, courseIdUnassigned, true)

	if updateUnassignedEntry {
		var unassignedParticipantsCount int64
		result = db.Model(model.Participant{}).Where("course_id is null").Count(&unassignedParticipantsCount)

		if result.Error != nil {
			slog.Error("Error while fetching unassigned participants count from db", "err", err)
			c.AbortWithStatus(500)

			return
		}

		viewUpdates.UnassignedEntry = ui.UnassignedEntry{ShouldRender: true, ParticipantsCount: int(unassignedParticipantsCount), AsOobSwap: true}
	}
	c.HTML(http.StatusOK, "assignments/course-list", viewUpdates)
}

func toViewCourses(models []model.Course, selectedId sql.NullInt64, allAsOobSwap bool) ui.CourseList {
	var courseViews []ui.Course

	for _, model := range models {
		view := toViewCourse(model, selectedId, allAsOobSwap)
		courseViews = append(courseViews, view)
	}

	return ui.CourseList{CourseEntries: courseViews}
}

func toViewCourse(model model.Course, selectedId sql.NullInt64, asOobSwap bool) ui.Course {
	return ui.Course{
		ID:          model.ID,
		Name:        model.Name,
		MinCapacity: model.MinCapacity,
		MaxCapacity: model.MaxCapacity,
		Selected:    selectedId.Valid && model.ID == int(selectedId.Int64),
		Allocation:  model.Allocation(),
		AsOobSwap:   asOobSwap,
	}
}

func pointerToNullable(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Valid: true, Int64: int64(*i)}
}
