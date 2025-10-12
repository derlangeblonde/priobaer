package app

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/app/respond"
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

	var participantID = uriParams.ParticipantID
	var courseID = uriParams.CourseID

	err := db.Transaction(func(tx *gorm.DB) error {
		return domain.InitialAssign(tx, participantID, courseID)
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrParticipantNotFound):
			respond.BadRequest(c, "Received non-existing ParticipantID", "id", participantID)
			return
		case errors.Is(err, domain.ErrCourseNotFound):
			respond.BadRequest(c, "Received non-existing CourseID", "id", courseID)
			return
		default:
			respond.InternalServerError(c, "Writing initial assignment to db failed", err)
			return
		}
	}

	unassignedCount, err := domain.CountUnassigned(db)
	if err != nil {
		respond.InternalServerError(c, "Counting unassigned participants failed", err)
		return
	}

	courseData, err := domain.FindSingleCourseData(db, courseID)
	if err != nil {
		respond.InternalServerError(c, "Finding course data failed", err)
		return
	}

	newCourseAllocation, err := domain.CountAllocation(db, courseID)
	if err != nil {
		respond.InternalServerError(c, "Counting allocation of assign target failed", err)
		return
	}
	uiUpdate := ui.NewOutOfBandCourseListUpdate().
		SelectUnassignedEntry().
		SetUnassignedCount(unassignedCount)

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

	participantID := uriParams.ParticipantID
	targetID := uriParams.CourseID
	var source, target domain.CourseData

	target, err := domain.FindSingleCourseData(db, targetID)
	if err != nil {
		respond.InternalServerError(c, "Finding target course data failed", err)
		return
	}
	source, err = domain.FindAssignedCourse(db, participantID)
	if err != nil {
		respond.InternalServerError(c, "Finding assigned course data failed", err)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.Reassign(tx, participantID, targetID)
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrParticipantNotFound):
			respond.BadRequest(c, "Received non-existing ParticipantID", "id", participantID)
			return
		case errors.Is(err, domain.ErrCourseNotFound):
			respond.BadRequest(c, "Received non-existing CourseID", "id", targetID)
			return
		default:
			respond.InternalServerError(c, "Writing initial assignment to db failed", err)
			return
		}
	}
	uiUpdate := ui.NewOutOfBandCourseListUpdate()
	for _, course := range []domain.CourseData{source, target} {
		newCourseAllocation, err := domain.CountAllocation(db, course.ID)
		if err != nil {
			respond.InternalServerError(c, "Counting allocation of assigned target failed", err, "courseID", course.ID)
			return
		}

		uiUpdate.AppendCourse(newUiCourse(course, newCourseAllocation))
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
		respond.InternalServerError(c, "Finding currently assigned course data failed", err)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.Unassign(tx, participantID)
	})

	switch {
	case errors.Is(err, domain.ErrParticipantNotFound):
		respond.BadRequest(c, "Received non-existing ParticipantID", "id", participantID)
		return
	case err != nil:
		respond.InternalServerError(c, "Deleting assignment in db failed", err)
		return
	}

	unassignedCount, err := domain.CountUnassigned(db)
	if err != nil {
		respond.InternalServerError(c, "Counting unassigned failed", err)
		return
	}

	sourceAllocation, err := domain.CountAllocation(db, source.ID)
	if err != nil {
		respond.InternalServerError(c, "Counting allocation of formerly assigned course failed", err)
		return
	}

	uiUpdate := ui.NewOutOfBandCourseListUpdate().
		SetUnassignedCount(unassignedCount)
	uiUpdate.AppendCourse(newUiCourse(source, sourceAllocation))

	c.HTML(http.StatusOK, "scenario/course-list", uiUpdate)
}

func pointerToNullable(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Valid: true, Int64: int64(*i)}
}

func newUiCourse(courseData domain.CourseData, allocation int) ui.Course {
	return ui.Course{
		ID:          int(courseData.ID),
		Name:        courseData.Name,
		MaxCapacity: courseData.MaxCapacity,
		MinCapacity: courseData.MinCapacity,
		Allocation:  allocation,
	}
}

type assignUriParams struct {
	ParticipantID domain.ParticipantID `uri:"id" binding:"required"`
	CourseID      domain.CourseID      `uri:"course-id" binding:"required"`
}
