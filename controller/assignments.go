package controller

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/view"
)

func AssignmentsIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsIndex", "err", err)
		return
	}

	participants := make([]model.Participant, 0)
	var result *gorm.DB

	if req.CourseIdSelected == nil {
		result = db.Where("course_id is null").Find(&participants)
	} else {
		courseID := *req.CourseIdSelected
		result = db.Where("course_id = ?", courseID).Find(&participants)
	}

	if result.Error != nil {
		slog.Error("Unexpected error while gettting assigned particpants from db", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	var courses []model.Course
	result = db.Model(&model.Course{}).Preload("Participants").Find(&courses)

	if result.Error != nil {
		slog.Error("Unexpected error while getting all courses from db", "err", result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	noCourseSelected := req.CourseIdSelected == nil
	viewCourses := toViewCourses(courses, pointerToNullable(req.CourseIdSelected), false)

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": false, "participants": participants, "courses": viewCourses, "noCourseSelected": noCourseSelected})

		return
	}

	c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": true, "participants": participants, "courses": viewCourses, "noCourseSelected": noCourseSelected})
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

	result := db.First(&participant, req.ParticipantId)

	courseIdUnassigned := participant.CourseID

	if result.Error != nil {
		slog.Error("Unexpected error in AssignmentUpdate while fetching participant from db", "err", result.Error)
		c.AbortWithStatus(500)

		return
	}

	if req.IsUnassign() {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", nil)
	} else {
		result = db.Model(model.Participant{}).Where("ID = ?", req.ParticipantId).Update("course_id", req.CourseId)
		// TODO: we might be overriding an error here
		result = db.Preload("Participants").First(&courseAssigned, req.CourseId)
		coursesToUpdate = append(coursesToUpdate, courseAssigned)
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
	}

	viewUpdates := toViewCourses(coursesToUpdate, courseIdUnassigned, true)
	c.HTML(http.StatusOK, "courses", viewUpdates)
}

func toViewCourses(models []model.Course, selectedId sql.NullInt64, allAsOobSwap bool) (views []view.Course) {
	for _, model := range models {
		view := toViewCourse(model, selectedId, allAsOobSwap)
		views = append(views, view)
	}

	return
}

func toViewCourse(model model.Course, selectedId sql.NullInt64, asOobSwap bool) view.Course {
	return view.Course{
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
