package app

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/ui"
)

func AssignmentsCreate(c *gin.Context) {
	logger := slog.With("Func", "AssignmentsCreate")
	db := GetDB(c)

	var uriParams assignUriParams

	if err := c.ShouldBindUri(&uriParams); err != nil {
		logger.Error("Failed to bind uri request", "err", err)
		return
	}

	var participantID = domain.ParticipantID(uriParams.ParticipantID)
	var courseID = domain.CourseID(uriParams.CourseID)

	err := db.Transaction(func(tx *gorm.DB) error {
		return domain.InitialAssign(tx, participantID, courseID)
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrParticipantNotFound):
			logger.Error("Received non-existing ParticipantID", "id", participantID)
			emptyBadRequestResponse(c)
			return
		case errors.Is(err, domain.ErrCourseNotFound):
			logger.Error("Received non-existing CourseID", "id", courseID)
			emptyBadRequestResponse(c)
			return
		default:
			logger.Error("Writing initial assignment to db failed", "err", err)
			internalServerErrorResponse(c)
			return
		}
	}

	unassignedCount, err := domain.CountUnassigned(db)
	if err != nil {
		logger.Error("Counting unassigned participants failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	courseData, err := domain.FindSingleCourseData(db, courseID)
	if err != nil {
		logger.Error("Finding course data failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	newCourseAllocation, err := domain.CountAllocation(db, courseID)
	if err != nil {
		logger.Error("Counting allocation of assign target failed", "err", err)
		internalServerErrorResponse(c)
		return
	}
	uiUpdate := ui.NewOutOfBandCourseListUpdate().SelectUnassignedEntry().SetUnassignedCount(unassignedCount)
	uiUpdate.AppendCourse(
		ui.Course{
			ID:          int(courseData.ID),
			Name:        courseData.Name,
			MaxCapacity: courseData.MaxCapacity,
			MinCapacity: courseData.MinCapacity,
			Allocation:  newCourseAllocation,
		},
	)

	c.HTML(http.StatusOK, "scenario/course-list", uiUpdate)
}

func AssignmentsUpdate(c *gin.Context) {
	logger := slog.With("Func", "AssignmentsCreate")
	db := GetDB(c)

	var uriParams assignUriParams

	if err := c.ShouldBindUri(&uriParams); err != nil {
		logger.Error("Failed to bind uri request", "err", err)
		return
	}

	participantID := domain.ParticipantID(uriParams.ParticipantID)
	targetID := domain.CourseID(uriParams.CourseID)
	var source, target domain.CourseData

	target, err := domain.FindSingleCourseData(db, targetID)
	if err != nil {
		logger.Error("Finding target course data failed", "err", err)
		internalServerErrorResponse(c)
		return
	}
	source, err = domain.FindAssignedCourse(db, participantID)
	if err != nil {
		logger.Error("Finding assigned course data failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.Reassign(tx, participantID, targetID)
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrParticipantNotFound):
			logger.Error("Received non-existing ParticipantID", "id", participantID)
			emptyBadRequestResponse(c)
			return
		case errors.Is(err, domain.ErrCourseNotFound):
			logger.Error("Received non-existing CourseID", "id", targetID)
			emptyBadRequestResponse(c)
			return
		default:
			logger.Error("Writing initial assignment to db failed", "err", err)
			internalServerErrorResponse(c)
			return
		}
	}
	uiUpdate := ui.NewOutOfBandCourseListUpdate()
	for _, course := range []domain.CourseData{source, target} {
		newCourseAllocation, err := domain.CountAllocation(db, course.ID)
		if err != nil {
			logger.Error("Counting allocation of assigned target failed", "err", err, "courseID", course.ID)
			internalServerErrorResponse(c)
			return
		}

		uiUpdate.AppendCourse(
			ui.Course{
				ID:          int(course.ID),
				Name:        course.Name,
				MaxCapacity: course.MaxCapacity,
				MinCapacity: course.MinCapacity,
				Allocation:  newCourseAllocation,
			},
		)
	}

	c.HTML(http.StatusOK, "scenario/course-list", uiUpdate)
}

func AssignmentsDelete(c *gin.Context) {
	type unassignUriParams struct {
		ParticipantID uint `uri:"id" binding:"required"`
	}

	logger := slog.With("Func", "AssignmentsDelete")
	db := GetDB(c)

	var uriParams unassignUriParams
	if err := c.ShouldBindUri(&uriParams); err != nil {
		logger.Error("Failed to bind uri request", "err", err)
		return
	}

	participantID := domain.ParticipantID(uriParams.ParticipantID)
	source, err := domain.FindAssignedCourse(db, participantID)
	if err != nil {
		logger.Error("Finding currently assigned course data failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.Unassign(tx, participantID)
	})

	switch {
	case errors.Is(err, domain.ErrParticipantNotFound):
		logger.Error("Received non-existing ParticipantID", "id", participantID)
		emptyBadRequestResponse(c)
		return
	case err != nil:
		logger.Error("Deleting assignment in db failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	unassignedCount, err := domain.CountUnassigned(db)
	if err != nil {
		logger.Error("Counting unassigned failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	sourceAllocation, err := domain.CountAllocation(db, source.ID)
	if err != nil {
		logger.Error("Counting allocation of formerly assigned course failed", "err", err)
		internalServerErrorResponse(c)
		return
	}

	uiUpdate := ui.NewOutOfBandCourseListUpdate().SetUnassignedCount(unassignedCount)
	uiUpdate.AppendCourse(
		ui.Course{
			ID:          int(source.ID),
			Name:        source.Name,
			MaxCapacity: source.MaxCapacity,
			MinCapacity: source.MinCapacity,
			Allocation:  sourceAllocation,
		},
	)

	c.HTML(http.StatusOK, "scenario/course-list", uiUpdate)
}

func pointerToNullable(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Valid: true, Int64: int64(*i)}
}

type assignUriParams struct {
	ParticipantID uint `uri:"id" binding:"required"`
	CourseID      uint `uri:"course-id" binding:"required"`
}
