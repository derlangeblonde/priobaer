package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbDirectory struct {
	rootDir string
	maxAge  time.Duration
	dbMap   map[string]*gorm.DB
	stopHandleMap map[string]clockwork.Timer
	clock clockwork.Clock
	models []any
}

func NewDbDirectory(rootDir string, maxAge time.Duration, clock clockwork.Clock, models []any) *DbDirectory {
	return &DbDirectory{rootDir: rootDir, maxAge: maxAge, dbMap: make(map[string]*gorm.DB, 0), stopHandleMap: make(map[string]clockwork.Timer), clock: clock, models: models}
}

func (d *DbDirectory) Open(dbId string) (*gorm.DB, error) {
	existingConn, ok := d.dbMap[dbId]

	if ok {
		return existingConn, nil
	}

	dbPath := d.Path(dbId)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	d.dbMap[dbId] = db
	db.AutoMigrate(d.models...)
	db.AutoMigrate(&Session{})

	var count int64
	db.Model(&Session{}).Count(&count)

	if count == 0 {
		db.Create(&Session{ExpiresAt: d.clock.Now().Add(d.maxAge)})
	} else {
		slog.Info("Session already set, this seems to be an existing db", "dbId", dbId)
	}

	if count > 1 {
		panic(fmt.Sprintf("Critical! Found multiple session entries in session table. dbId=%s, count=%d", dbId, count))
	}

	d.scheduleRemoval(dbId)

	return db, err
}

func (d *DbDirectory) Get(dbId string) (*gorm.DB, bool) {
	db, ok := d.dbMap[dbId]

	return db, ok
}

func (d *DbDirectory) GetExpirationDate(dbId string) (time.Time, error) {
	db, ok := d.Get(dbId)

	if !ok {
		return time.Time{}, fmt.Errorf("Requested expiration date for db that is not known to dbDirectory. dbId=%s", dbId)
	}

	var session Session
	result := db.First(&session)

	if result.Error != nil {
		return time.Time{}, result.Error
	}

	return session.ExpiresAt, nil
}

func (d *DbDirectory) ReadExistingDbs() error {
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

		_, err = d.Open(candidateUuid)

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DbDirectory) Close() []error {
	errs := make([]error, 0)
	for dbId, db := range d.dbMap {
		conn, err := db.DB()

		if err != nil {
			errs = append(errs, err)
		}

		err = conn.Close()

		if err != nil {
			errs = append(errs, err)
		}

		stopHandle, ok := d.stopHandleMap[dbId]

		if !ok {
			errs = append(errs, fmt.Errorf("No stop handle found for %s.", dbId))
		}

		stopHandle.Stop()
	}

	return errs
}

func (d *DbDirectory) Remove(dbId string) error {
	db, ok := d.dbMap[dbId] 
	defer delete(d.dbMap, dbId)

	if !ok {
		slog.Warn("Tried to remove db, but was not in map", "dbId", dbId)
	}

	conn, err := db.DB()

	if err == nil {
		err = conn.Close()
	}

	if err != nil {
		slog.Warn("Tried to close connection do db, but got an error", "err", err)
	}

	dbPath := d.Path(dbId)
	err = os.Remove(dbPath)

	if err != nil {
		return err
	}

	return nil
}


func (d *DbDirectory) Path(dbId string) string {
	return path.Join(d.rootDir, fmt.Sprintf("%s.sqlite", dbId))
}

func (d *DbDirectory) scheduleRemoval(dbId string) {
	expirationDate, err := d.GetExpirationDate(dbId)

	if err != nil {
		slog.Error("Could not get expiration date. This db will not be scheduled for removal", "err", err)	

		return
	}

	now := d.clock.Now()

	if now == expirationDate || now.After(expirationDate) {
		slog.Info("Called scheduleRemoval for DB that is already expired. Removing now")
		err := d.Remove(dbId)

		if err != nil {
			slog.Error("Could not remove db :(", "err", err)
		}

		return
	}

	expireIn := expirationDate.Sub(d.clock.Now())

	stopHandle := d.clock.AfterFunc(expireIn, func() {
		err := d.Remove(dbId)

		if err != nil {
			slog.Error("Could not remove db :(", "err", err)
		}
	})

	d.stopHandleMap[dbId] = stopHandle
}

