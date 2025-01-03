package dbdir

import (
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (d *DbDirectory) getEntry(dbId string) (*entry, bool) {
	entryUntyped, ok := d.entries.Load(dbId)
	entry, ok := entryUntyped.(*entry)

	return entry, ok
}

func (d *DbDirectory) getAndDeleteEntry(dbId string) (*entry, bool) {
	entryUntyped, ok := d.entries.LoadAndDelete(dbId)
	entry, ok := entryUntyped.(*entry)

	return entry, ok
}

func (d *DbDirectory) setEntry(dbId string, entry *entry) {
	d.entries.Store(dbId, entry)
}

func (d *DbDirectory) iterEntries() iter.Seq2[string, *entry] {
	return func(yield func(k string, v *entry) bool) {
		d.entries.Range(func(key, value any) bool {
			keyTyped, ok := key.(string)
			if !ok {
				return false
			}
			valueTyped, ok := key.(*entry)
			if !ok {
				return false
			}

			return yield(keyTyped, valueTyped)
		})
	}
}

func (d *DbDirectory) remove(dbId string) error {
	entry, ok := d.getAndDeleteEntry(dbId)

	if !ok {
		slog.Warn("Tried to remove db, but was not in map", "dbId", dbId)

		return nil
	}

	conn, err := entry.conn.DB()

	if err == nil {
		err = conn.Close()
	}

	if err != nil {
		slog.Warn("Tried to close connection do db, but got an error", "err", err)
	}

	dbPath := d.path(dbId)
	err = os.Remove(dbPath)

	if err != nil {
		return err
	}

	return nil
}

func (d *DbDirectory) scheduleRemoval(dbId string) {
	expirationDate, err := d.getExpirationDate(dbId)

	if err != nil {
		slog.Error("Could not get expiration date. This db will not be scheduled for removal", "err", err)

		return
	}

	now := d.clock.Now()

	if now == expirationDate || now.After(expirationDate) {
		slog.Info("Called scheduleRemoval for DB that is already expired. Removing now")
		err := d.remove(dbId)

		if err != nil {
			slog.Error("Could not remove db :(", "err", err)
		}

		return
	}

	expireIn := expirationDate.Sub(d.clock.Now())

	expirationTimer := d.clock.AfterFunc(expireIn, func() {
		err := d.remove(dbId)

		if err != nil {
			slog.Error("Could not remove db :(", "err", err)
		}
	})

	entry, _ := d.getEntry(dbId)
	entry.expirationTimer = expirationTimer
}

func (d *DbDirectory) restoreExistingDbs() error {
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

func (d *DbDirectory) getExpirationDate(dbId string) (time.Time, error) {
	entry, ok := d.getEntry(dbId)

	if !ok {
		return time.Time{}, fmt.Errorf("Requested expiration date for db that is not known to dbDirectory. dbId=%s", dbId)
	}

	var session Session
	result := entry.conn.First(&session)

	if result.Error != nil {
		return time.Time{}, result.Error
	}

	return session.ExpiresAt, nil
}

func (d *DbDirectory) path(dbId string) string {
	return path.Join(d.rootDir, fmt.Sprintf("%s.sqlite", dbId))
}
