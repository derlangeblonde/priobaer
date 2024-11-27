package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
)

func TestReadExistingDbs_IngnoresAlreadyExpired(t *testing.T) {
	is := is.New(t)

	testDir := t.TempDir()

	dbName := uuid.New().String();

	fakeClock := clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local))

	dbDirectoryPrevious := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)

	dbDirectoryPrevious.OpenDB(dbName)

	errs := dbDirectoryPrevious.Close()
	is.Equal(len(errs), 0) // could not close dbs properly

	fakeClock.Advance(time.Second * time.Duration(61))

	dbDirectoryCurrent := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
	err := dbDirectoryCurrent.ReadExistingDbs()  
	is.NoErr(err) // ReadExistingDbs failed

	_, ok := dbDirectoryCurrent.Get(dbName)
	is.True(!ok) // new db manager did not read existing instance
}

func TestReadExistingDbs_SchedulesRemoval(t *testing.T) {
	is := is.New(t)

	testDir := t.TempDir()

	dbId := uuid.New().String();

	fakeClock := clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local))

	dbDirectoryPrevious := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)

	dbDirectoryPrevious.OpenDB(dbId)

	errs := dbDirectoryPrevious.Close()
	is.Equal(len(errs), 0) // could not close dbs properly

	fakeClock.Advance(time.Second * time.Duration(30))

	dbDirectoryCurrent := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
	err := dbDirectoryCurrent.ReadExistingDbs()
	is.NoErr(err) // error during reading existing dbs

	_, ok := dbDirectoryCurrent.Get(dbId)
	is.True(ok) // new db manager did not read existing instance

	exists := FileExists(dbDirectoryCurrent.Path(dbId))
	is.True(exists) // file did not exist but should have

	fakeClock.Advance(31 * time.Second)
	time.Sleep(10 * time.Microsecond)
	exists = FileExists(dbDirectoryCurrent.Path(dbId))
	is.True(!exists) // file did exist but should not have
}

func FileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}


