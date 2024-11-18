package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SessionDBMapper struct {
	rootDir string
	maxAge  time.Duration
	dbMap   map[string]*gorm.DB
}

func NewSessionDBMapper(rootDir string, maxAge time.Duration) *SessionDBMapper {
	return &SessionDBMapper{rootDir: rootDir, maxAge: maxAge, dbMap: make(map[string]*gorm.DB, 0)}
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

func (d *SessionDBMapper) Get(dbId string) (*gorm.DB, bool) {
	db, ok := d.dbMap[dbId]

	return db, ok
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
			err := d.removeDb(dbId)

			if err != nil {
				slog.Error("Could not remove db :(", "err", err)
			}
		}
	}()
}

func (d *SessionDBMapper) removeDb(dbId string) error {
	db, ok := d.dbMap[dbId]

	if !ok {
		return fmt.Errorf("Tried to remove db, but dbId=%s was not in map", dbId)
	}

	conn, err := db.DB()

	if err != nil {
		return err
	}

	err = conn.Close()

	if err != nil {
		return err
	}

	dbPath := d.formatDbPath(dbId)
	err = os.Remove(dbPath)

	if err != nil {
		return err
	}

	delete(d.dbMap, dbId)

	return nil
}

