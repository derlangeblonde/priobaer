package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
)

func TestReadExistingDbs_SchedulesRemoval(t *testing.T) {
	is := is.New(t)

	testDir := t.TempDir()

	dbName := uuid.New().String();

	fullDbPath := path.Join(testDir, dbName)

	fakeClock := clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local))

	dbManagerPrevious := NewDbManager(testDir, time.Second * time.Duration(60), fakeClock)

	dbManagerPrevious.OpenDB(dbName)

	errs := dbManagerPrevious.Close()
	is.Equal(len(errs), 0) // could not close dbs properly

	fakeClock.Advance(time.Second * time.Duration(30))

	dbManagerCurrent := NewDbManager(testDir, time.Second * time.Duration(60), fakeClock)
	err := dbManagerCurrent.ReadExistingDbs()
	is.NoErr(err) // error during reading existing dbs

	_, ok := dbManagerCurrent.Get(dbName)
	is.True(ok) // new db manager did not read existing instance

	exists := FileExists(fmt.Sprintf("%s.sqlite", fullDbPath))
	is.True(exists) // file did not exist but should have

	fakeClock.Advance(31 * time.Second)
	exists = FileExists(fullDbPath)
	is.True(!exists) // file did exist but should not have
}

func FileExists(path string) bool {
	fileInfo, err := os.Stat(path)

	slog.Info("File Exists", "fileInfo", fileInfo)

	return err == nil
}


