package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const sessionIdKey = "session_id"
const dbKey = "db"

type SessionDBMapper struct {
	rootDir string
	maxAge time.Duration 
	dbMap   map[string]*gorm.DB
}

func NewSessionDBMapper(rootDir string, maxAge time.Duration) SessionDBMapper {
	return SessionDBMapper{rootDir: rootDir, maxAge: maxAge, dbMap: make(map[string]*gorm.DB, 0)}
}

func (d *SessionDBMapper) TryCloseAllDbs() []error {
	errs := make([]error, 0)
	for _, db := range d.dbMap {
		conn, err := db.DB()

		if err != nil {
			errs = append(errs, err)
		}

		err = conn.Close()

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (d *SessionDBMapper) NewDB(dbId string) (*gorm.DB, error) {
	dbPath := d.formatDbPath(dbId)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	d.dbMap[dbId] = db
	db.AutoMigrate(&Session{}, &Course{}, &Participant{})
	db.Create(&Session{ExpiresAt: time.Now().Add(d.maxAge)})

	d.scheduleDbRemovalAfterExpiration(dbId)

	return db, err
}


func (d *SessionDBMapper) formatDbPath(dbId string) string {
	return path.Join(d.rootDir, fmt.Sprintf("%s.sqlite", dbId))
}


func (d *SessionDBMapper) scheduleDbRemovalAfterExpiration(dbId string) {
	go func() {
		<-time.After(d.maxAge)

		db, _ := d.dbMap[dbId]

		var session Session
		db.First(&session)

		now := time.Now()

		if now == session.ExpiresAt || now.After(session.ExpiresAt) {
			// TODO: error handling
			_ = d.removeDb(dbId)
		}
	}()
}

func (d *SessionDBMapper) removeDb(dbId string) error {
	db, ok := d.dbMap[dbId]

	if !ok {
		return fmt.Errorf("Tried to remove db, but dbId=%s was not in map", dbId)
	}

	conn, _ := db.DB()
	_ = conn.Close()

	dbPath := d.formatDbPath(dbId)
	_ = os.Remove(dbPath)

	delete(d.dbMap, dbId)

	return nil
}

func (d *SessionDBMapper) ReadExistingSessions() error {
	fsEntries, err := os.ReadDir(d.rootDir)

	if err != nil {
		return err
	}

	for _, entry := range fsEntries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sqlite") {
			continue
		}

		candidateUuid := strings.Replace(entry.Name(), ".sqlite", "", 1)
		_, err := uuid.Parse(candidateUuid)

		if err != nil {
			slog.Info("There was a file with sqlite file-extension, which name was not parsable as uuid", "filname", entry.Name())
			continue
		}

		dbPath := path.Join(d.rootDir, fmt.Sprintf("%s.sqlite", candidateUuid))
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

		if err != nil {
			return err
		}

		d.dbMap[candidateUuid] = db
	}

	return nil
}

func (d *SessionDBMapper) Get(dbId string) (*gorm.DB, bool) {
	db, ok := d.dbMap[dbId]

	return db, ok
}

type Session struct {
	gorm.Model
	ExpiresAt time.Time
}

func (d *SessionDBMapper) InjectDB() gin.HandlerFunc {

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
			conn, ok := d.Get(sessionId)

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

		db, err := d.NewDB(newDbId.String())

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
