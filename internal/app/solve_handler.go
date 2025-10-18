package app

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/app/respond"
	"softbaer.dev/ass/internal/domain/solve"
)

func SolveAssignments(c *gin.Context) {
	logger := slog.With("Func", "SolveAssignments")
	db := GetDB(c)

	err := db.Transaction(
		func(tx *gorm.DB) error {
			return solve.ComputeAndApplyOptimalAssignments(c.Request.Context(), tx)
		},
	)

	switch {
	case errors.Is(err, solve.Timeout):
		logger.Info("solve timed out", "err", err)
		c.HTML(http.StatusOK, "dialogs/solve-timed-out", gin.H{})

		return
	case errors.Is(err, solve.UserCancelled):
		// Client probably already closed connection, so does not matter too much what we return here.
		c.AbortWithStatus(500)
		return
	case errors.Is(err, solve.NotSolvable):
		logger.Info("Could not solve assignment", "err", err)
		c.HTML(http.StatusOK, "dialogs/not-solvable", gin.H{})

		return
	}

	if err != nil {
		respond.InternalServerError(c, "Error while trying to solve assignment", err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/scenario")
}
