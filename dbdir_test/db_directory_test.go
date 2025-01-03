package dbdir_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/dbdir"
)

const timeWaitingForRemoval = time.Microsecond * 100

func TestOpen_ReturnsSameConnection_WhenCalledMultipleTimes(t *testing.T) {
	is := is.New(t)

	c := newConfig(t)
	sut := c.createSut()
	defer sut.Close()

	id := c.DbId

	conn1, err := sut.Open(id.String())
	is.NoErr(err)

	conn2, err := sut.Open(id.String())
	is.NoErr(err)

	is.Equal(conn1, conn2)
}

func TestOpen_ConnectionPersistsData_BetweenMultipleOpens(t *testing.T) {
	is := is.New(t)

	c := newConfig(t)
	sut := c.createSut()
	defer sut.Close()

	expectedNumber := 42

	conn1, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	conn2, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	var actualData testData
	result = conn2.First(&actualData)
	is.NoErr(result.Error)

	is.Equal(actualData.Number, expectedNumber) // did not got the same number set earlier
}

func TestOpen_YieldsNewConnection_AfterExpired(t *testing.T) {
	is := is.New(t)

	expiration := 60 * 24 * time.Minute

	c := newConfig(t).withExpiration(expiration)

	sut := c.createSut()
	defer sut.Close()

	conn1, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	c.FakeClock.Advance(expiration)
	time.Sleep(timeWaitingForRemoval)

	conn2, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn2) // after db expired open should return a new connection
}

func TestOpen_DataIsErased_AfterExpired(t *testing.T) {
	is := is.New(t)

	expiration := 60 * 24 * time.Minute
	expectedNumber := 666

	c := newConfig(t).withExpiration(expiration)

	sut := c.createSut()
	defer sut.Close()

	conn1, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	c.FakeClock.Advance(expiration)
	time.Sleep(timeWaitingForRemoval)

	conn2, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn2) // after db expired open should return a new connection

	connHasNoRows(conn2, is)
}

func TestNewDbDirectory_RestoresDataAndExpirationFromExistingDbFiles(t *testing.T) {
	is := is.New(t)

	expectedNumber := 282

	expiration := 60 * time.Second
	c := newConfig(t).withExpiration(expiration)

	sutOld := c.createSut()

	conn1, err := sutOld.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	errs := sutOld.Close()
	is.Equal(len(errs), 0) // closing sut yielded some errors

	sutNew := c.createSut()
	defer sutNew.Close()

	conn2, err := sutNew.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn2) // because conn's were retrieved from different sut, they should be different

	var actualData testData
	result = conn2.First(&actualData)
	is.NoErr(result.Error)

	is.Equal(actualData.Number, expectedNumber) // did not got the same number set earlier

	c.FakeClock.Advance(expiration)
	time.Sleep(timeWaitingForRemoval)

	conn3, err := sutNew.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn2 != conn3) //conn2 should be expired by now, therefore conn3 should be a new connection

	connHasNoRows(conn3, is)
}

func TestNewDbDirectory_RestoresExpiration_AlthoughDbNeverAccessed(t *testing.T) {
	is := is.New(t)

	expectedNumber := 282

	expiration := 60 * time.Second
	c := newConfig(t).withExpiration(expiration)

	sutOld := c.createSut()

	conn1, err := sutOld.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	errs := sutOld.Close()
	is.Equal(len(errs), 0) // closing sut yielded some errors

	sutNew := c.createSut()
	defer sutNew.Close()

	c.FakeClock.Advance(expiration)
	time.Sleep(timeWaitingForRemoval)

	conn3, err := sutNew.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn3) //conn2 should be expired by now, therefore conn3 should be a new connection

	connHasNoRows(conn3, is)
}

func TestNewDbDirectory_RemovesExpiredDbs(t *testing.T) {
	is := is.New(t)

	clock := clockwork.NewFakeClockAt(time.Date(2025, 1, 1, 9, 5, 30, 0, time.Local))
	tmpDir := t.TempDir()
	dbPath := fmt.Sprintf("%s/%s.sqlite", tmpDir, uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err) // error while creating db

	session := dbdir.Session{ExpiresAt: clock.Now()}
	db.AutoMigrate(&dbdir.Session{})
	db.Create(&session)

	clock.Advance(time.Second * 10)

	deadlockPrevention := time.After(time.Second * 1)
	done := make(chan bool, 0)

	go func() {
		_, err = dbdir.New(tmpDir, time.Second*60, clock, []any{})
		is.NoErr(err) // err while creating dbdir
		done <- true
	}()

	select {
	case <-deadlockPrevention:
		t.Fatalf("dbdir.New took too long, probably deadlocked")
	case <-done:
		t.Log("dbdir.New terminated")
	}

	if _, err := os.Stat(dbPath); err == nil {
		t.Fatal("db-file still exists")

	} else if errors.Is(err, os.ErrNotExist) {
		t.Log("all good, file is gone :)")
	} else {
		t.Fatalf("could not open file for other reason then 'ErrNotExist': %v", err)
	}
}
