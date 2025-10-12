package respond

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func InternalServerError(c *gin.Context, logMessage string, err error, args ...any) {
	logger := slog.With("Path", c.Request.URL.RawPath, "ResponseType", "InternalServerError")
	logger.Error(logMessage, "err", err, args)
	c.HTML(500, "general/500", gin.H{})
}

func BadRequest(c *gin.Context, logMessage string, args ...any) {
	logger := slog.With("Path", c.Request.URL.RawPath, "ResponseType", "InternalServerError")
	logger.Error(logMessage, args)
	c.HTML(400, "general/400", gin.H{})
}
