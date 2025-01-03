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
		Solve bool `form:"solve"`
	}

	db := GetDB(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsIndex", "err", err)
		return
	}

	if req.Solve {
		slog.Error("do solve")
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

				assignments, err := model.SolveAssignment(availableCourses, unassignedParticipants)
				if err != nil {
					return err
				}

				for _, assignment := range assignments {
					slog.Error("applying assignment", "partID", assignment.Participant.ID, "courseID", assignment.Course.ID)
					if result := tx.Model(model.Participant{}).Where("ID = ?", assignment.Participant.ID).Update("course_id", assignment.Course.ID); result.Error != nil {
						return result.Error
					}
				}

				return nil	
			},
		)		

		if err != nil {
			slog.Error("Error while trying to solve assignment", "err", err)
			c.AbortWithStatus(500)

			return
		}
	} else {
		slog.Error("do not solve--------------------------------------")
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

	var unassignedParticipantsCount int64
	result = db.Model(model.Participant{}).Where("course_id is null").Count(&unassignedParticipantsCount)

	if result.Error != nil {
		slog.Error("Error while fetching unassigned participants count from db", "err", err)
		c.AbortWithStatus(500)

		return
	}

	viewCourses := toViewCourses(courses, pointerToNullable(req.CourseIdSelected), false)
	viewCourses.UnassignedEntry.Selected = req.CourseIdSelected == nil
	viewCourses.UnassignedEntry.ParticipantsCount = int(unassignedParticipantsCount)
	viewCourses.UnassignedEntry.ShouldRender = true

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": false, "participants": participants, "courseList": viewCourses})

		return
	}

	c.HTML(http.StatusOK, "assignments/index", gin.H{"fullPage": true, "participants": participants, "courseList": viewCourses})
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

		viewUpdates.UnassignedEntry = view.UnassignedEntry{ShouldRender: true, ParticipantsCount: int(unassignedParticipantsCount), AsOobSwap: true}
	}
	c.HTML(http.StatusOK, "assignments/course-list", viewUpdates)
}

func toViewCourses(models []model.Course, selectedId sql.NullInt64, allAsOobSwap bool) view.CourseList {
	var courseViews []view.Course

	for _, model := range models {
		view := toViewCourse(model, selectedId, allAsOobSwap)
		courseViews = append(courseViews, view)
	}

	return view.CourseList{CourseEntries: courseViews}
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
