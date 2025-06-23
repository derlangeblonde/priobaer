package dbdir

import (
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(rootDir string, maxAge time.Duration, clock clockwork.Clock, models []any) (*DbDirectory, error) {
	dbdir := &DbDirectory{rootDir: rootDir, maxAge: maxAge, entries: sync.Map{}, clock: clock, models: models}

	err := dbdir.restoreExistingDbs()

	return dbdir, err
}

func (d *DbDirectory) Open(dbId string) (*gorm.DB, error) {
	existingEntry, ok := d.getEntry(dbId)

	if ok {
		return existingEntry.conn, nil
	}

	dbPath := d.path(dbId)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	d.setEntry(dbId, &entry{conn: db})
	db.AutoMigrate(d.models...)
	db.AutoMigrate(&Session{})
	db.Exec("PRAGMA foreign_keys = ON;")

	var count int64
	db.Model(&Session{}).Count(&count)

	if count == 0 {
		db.Create(&Session{ExpiresAt: d.clock.Now().Add(d.maxAge)})
	}

	if count > 1 {
		return nil, fmt.Errorf("critical! Found multiple session entries in session table. dbId=%s, count=%d", dbId, count)
	}

	d.scheduleRemove(dbId)

	return db, err
}

func (d *DbDirectory) Close() []error {
	errs := make([]error, 0)
	for dbId, entry := range d.iterEntries() {
		conn, err := entry.conn.DB()

		if err != nil {
			errs = append(errs, err)
		}

		err = conn.Close()

		if err != nil {
			errs = append(errs, err)
		}

		if entry.expirationTimer == nil {
			errs = append(errs, fmt.Errorf("no expiration timer found for %s", dbId))
		}

		entry.expirationTimer.Stop()
	}

	return errs
}
