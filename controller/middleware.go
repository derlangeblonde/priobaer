package controller 

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/dbdir"
)

const sessionIdKey = "session_id"
const dbKey = "db"


func InjectDB(dbDirectory *dbdir.DbDirectory) gin.HandlerFunc {

	return func(c *gin.Context) {
		whitelist := []string{
			"/health",
			"/static/*filepath",
			"favicon.png",
			"favicon.ico",
			"/sessions/new",
		}

		if slices.Contains(whitelist, c.FullPath()) {
			c.Next()

			return
		}

		sessionId, ok := getSessionId(c)

		if ok {
			conn, err := dbDirectory.Open(sessionId)

			if err != nil {
				slog.Error("SessionId existed, but there was an error when opening db-conn", "err", err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}

			c.Set(dbKey, conn)

			c.Next()

			return
		}

		c.Redirect(http.StatusSeeOther, "/sessions/new")

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
