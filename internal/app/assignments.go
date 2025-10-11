package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/ui"
)

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

	// Fetch affected Pp
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
	c.HTML(http.StatusOK, "scenario/course-list", viewUpdates)
}

func toViewCourses(models []model.Course, selectedId sql.NullInt64, allAsOobSwap bool) ui.CourseList {
	var courseViews []ui.Course

	for _, m := range models {
		view := toViewCourse(m, selectedId, allAsOobSwap)
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
