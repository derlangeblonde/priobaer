package cmd

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const sessionIdKey = "session_id"
const dbKey = "db"


type SessionDBMapper struct {
	dbMap map[string]*gorm.DB
}

func NewSessionDBMapper() SessionDBMapper {
	return SessionDBMapper{dbMap: make(map[string]*gorm.DB, 0)}
}

func (d *SessionDBMapper) New(sessionId string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file::memory:?%s", sessionId)), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	d.dbMap[sessionId] = db
	db.AutoMigrate(&Session{})
	db.Create(&Session{SessionId: sessionId})

	return db, err
}

func (d *SessionDBMapper) Get(sessionId string) (*gorm.DB, bool) {
	db, ok := d.dbMap[sessionId]

	return db, ok
}


type Session struct {
	gorm.Model
	SessionId string
}

func (d *SessionDBMapper) InjectDB() gin.HandlerFunc {

	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionId, ok := getSessionId(c)

		if ok {
			conn, ok := d.Get(sessionId)

			if !ok {
				slog.Error("SessionId existed, but no db-connection was found.")
				c.AbortWithStatus(http.StatusInternalServerError)
			}

			c.Set(dbKey, conn)
			
			c.Next()

			return
		}

		newSessionId, err := uuid.NewRandom()

		if err != nil {
			slog.Error("Failed while generating uuid", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		session.Set(sessionIdKey, newSessionId.String())
		err = session.Save()

		if err != nil {
			slog.Error("Failed while saving session", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		conn, err := d.New(newSessionId.String()) 

		if err != nil {
			slog.Error("Failed while opening new sqlite in-memory connection", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Set(dbKey, conn)

		c.Next()
	}

}

func GetDb(c *gin.Context) (*gorm.DB, bool) {
	maybeDb, ok := c.Get(dbKey)

	if !ok {
		return nil, ok 
	}

	db, ok := maybeDb.(*gorm.DB)	

	return db, ok
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
