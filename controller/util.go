package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DbError(c *gin.Context, err error, function string) {
	slog.Error("DB-Operation Failed", "function", function, "err", err)
	c.HTML(http.StatusInternalServerError, "dialogs/generic-error", err)
}
