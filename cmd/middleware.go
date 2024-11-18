package cmd

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const sessionIdKey = "session_id"
const dbKey = "db"


type Session struct {
	gorm.Model
	ExpiresAt time.Time
}

func InjectDB(sessionDbMapper *SessionDBMapper) gin.HandlerFunc {

	return func(c *gin.Context) {
		whitelist := []string{
			"/health",
			"/static/*filepath",
			"favicon.png",
			"favicon.ico",
		}

		if slices.Contains(whitelist, c.FullPath()) {
			c.Next()

			return
		}

		session := sessions.Default(c)
		sessionId, ok := getSessionId(c)

		if ok {
			conn, ok := sessionDbMapper.Get(sessionId)

			if !ok {
				slog.Error("SessionId existed, but no db-connection was found.")
				c.AbortWithStatus(http.StatusInternalServerError)
			}

			c.Set(dbKey, conn)

			c.Next()

			return
		}

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

		db, err := sessionDbMapper.NewDB(newDbId.String())

		if err != nil {
			slog.Error("Failed while opening new sqlite in-memory connection", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Set(dbKey, db)

		c.Next()
	}

}

func GetDB(c *gin.Context) (db *gorm.DB) {
	if val, ok := c.Get(dbKey); ok && val != nil {
		db, _ = val.(*gorm.DB)
	}
	return
}

func getSessionId(c *gin.Context) (string, bool) {
	session := sessions.Default(c)

	maybeSessionId := session.Get(sessionIdKey)

	if maybeSessionId == nil {
		return "", false
	}

	sessionId, ok := maybeSessionId.(string)

	return sessionId, ok
}
