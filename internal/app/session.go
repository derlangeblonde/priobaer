package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"softbaer.dev/ass/internal/dbdir"
)

func SessionNew(c *gin.Context) {
	c.HTML(http.StatusOK, "sessions/new", nil)
}

func SessionCreate(dbDirectory *dbdir.DbDirectory) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		newDbId, err := uuid.NewRandom()

		if err != nil {
			slog.Error("Failed while generating uuid", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		session.Set(sessionIdKey, newDbId.String())
		err = session.Save()

		if err != nil {
			slog.Error("Failed while saving session", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		_, err = dbDirectory.Open(newDbId.String())

		if err != nil {
			slog.Error("Failed to open new db", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Redirect(http.StatusSeeOther, "/assignments")
	}
}
