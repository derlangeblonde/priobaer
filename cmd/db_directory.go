package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbDirectory struct {
	rootDir string
	maxAge  time.Duration
	entries   map[string]*dbEntry
	clock clockwork.Clock
	models []any
	bigLock sync.Mutex
}

type dbEntry struct {
	conn *gorm.DB
	expirationTimer clockwork.Timer
}

func NewDbDirectory(rootDir string, maxAge time.Duration, clock clockwork.Clock, models []any) (*DbDirectory, error) {
	dbdir := &DbDirectory{rootDir: rootDir, maxAge: maxAge, entries: make(map[string]*dbEntry, 0), clock: clock, models: models}

	err := dbdir.ReadExistingDbs()

	return dbdir, err
}

func (d *DbDirectory) Open(dbId string) (*gorm.DB, error) {
	d.bigLock.Lock()
	defer d.bigLock.Unlock()
	existingEntry, ok := d.entries[dbId]

	if ok {
		return existingEntry.conn, nil
	}

	dbPath := d.Path(dbId)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	d.entries[dbId] = &dbEntry{conn: db}
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
	entry, ok := d.entries[dbId]

	return entry.conn, ok
}

func (d *DbDirectory) getExpirationDate(dbId string) (time.Time, error) {
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
	d.bigLock.Lock()
	defer d.bigLock.Unlock()
	errs := make([]error, 0)
	for dbId, entry := range d.entries {
		conn, err := entry.conn.DB()

		if err != nil {
			errs = append(errs, err)
		}

		err = conn.Close()

		if err != nil {
			errs = append(errs, err)
		}

		if entry.expirationTimer == nil {
			errs = append(errs, fmt.Errorf("No expiration timer found for %s.", dbId))
		}

		entry.expirationTimer.Stop()
	}

	return errs
}

func (d *DbDirectory) Remove(dbId string) error {
	d.bigLock.Lock()
	defer d.bigLock.Unlock()
	entry, ok := d.entries[dbId] 

	if !ok {
		slog.Warn("Tried to remove db, but was not in map", "dbId", dbId)
	}

	defer delete(d.entries, dbId)

	conn, err := entry.conn.DB()

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
	slog.Error("about to schedule")
	expirationDate, err := d.getExpirationDate(dbId)
	slog.Error("determined exp", "exp", expirationDate)

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

	expirationTimer := d.clock.AfterFunc(expireIn, func() {
		err := d.Remove(dbId)

		if err != nil {
			slog.Error("Could not remove db :(", "err", err)
		}
	})

	d.entries[dbId].expirationTimer = expirationTimer
}

