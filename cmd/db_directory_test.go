package cmd

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
	"gorm.io/gorm"
)

type defaultTestSetup struct {
	TmpDir string
	FakeClock clockwork.FakeClock
	DbId uuid.UUID 
	Sut *DbDirectory
}

type testData struct {
	gorm.Model
	Number int
}

type config struct {
	TmpDir string
	FakeClock clockwork.FakeClock
	DbId uuid.UUID 
	Expiration time.Duration
	Models []any
	T *testing.T
}

func newConfig(t *testing.T) *config {
	return &config{
		TmpDir: t.TempDir(),
		FakeClock:  clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local)),
		DbId: uuid.New(),
		Expiration: 60 * time.Second,
		Models:[]any{&testData{}},
		T: t ,
	}
}

func (c *config) withExpiration(e time.Duration) *config {
	c.Expiration = e

	return c
}

func (c *config) CreateSut() *DbDirectory {
	sut, err := NewDbDirectory(c.TmpDir, c.Expiration, c.FakeClock, c.Models)

	if err != nil {
		c.T.Fatalf("Could not create sut, err: %v", err)
	}

	return sut
}


func TestOpen_ReturnsSameConnection_WhenCalledMultipleTimes(t *testing.T) {
	is := is.New(t)
	
	c := newConfig(t)
	sut := c.CreateSut()
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
	sut := c.CreateSut()
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

	sut := c.CreateSut()

	conn1, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	c.FakeClock.Advance(expiration)
	time.Sleep(40 * time.Microsecond)

	conn2, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn2) // after db expired open should return a new connection
}


func TestOpen_DataIsErased_AfterExpired(t *testing.T) {
	is := is.New(t)

	expiration := 60 * 24 * time.Minute
	expectedNumber := 666

	c := newConfig(t).withExpiration(expiration)

	sut := c.CreateSut()

	conn1, err := sut.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	c.FakeClock.Advance(expiration)
	time.Sleep(40 * time.Microsecond)

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

	sutOld := c.CreateSut()

	conn1, err := sutOld.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	errs := sutOld.Close()
	is.Equal(len(errs), 0) // closing sut yielded some errors
	
	sutNew := c.CreateSut()

	conn2, err := sutNew.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn2) // because conn's were retrieved from different sut, they should be different

	var actualData testData
	result = conn2.First(&actualData)
	is.NoErr(result.Error)

	is.Equal(actualData.Number, expectedNumber) // did not got the same number set earlier

	c.FakeClock.Advance(expiration)
	time.Sleep(40 * time.Microsecond)

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

	sutOld := c.CreateSut()

	conn1, err := sutOld.Open(c.DbId.String())
	is.NoErr(err)

	result := conn1.Create(&testData{Number: expectedNumber})
	is.NoErr(result.Error)

	errs := sutOld.Close()
	is.Equal(len(errs), 0) // closing sut yielded some errors
	
	sutNew := c.CreateSut()

	c.FakeClock.Advance(expiration)
	time.Sleep(40 * time.Microsecond)

	conn3, err := sutNew.Open(c.DbId.String())
	is.NoErr(err)

	is.True(conn1 != conn3) //conn2 should be expired by now, therefore conn3 should be a new connection

	connHasNoRows(conn3, is)
}

func connHasNoRows(conn *gorm.DB, is *is.I) {
	var datas []testData
	result := conn.Find(&datas)
	is.NoErr(result.Error)

	is.Equal(len(datas), 0)
}


// func TestReadExistingDbs_IngnoresAlreadyExpired(t *testing.T) {
// 	is := is.New(t)
//
// 	testDir := t.TempDir()
//
// 	dbName := uuid.New().String();
//
// 	fakeClock := clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local))
//
// 	dbDirectoryPrevious := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
//
// 	dbDirectoryPrevious.Open(dbName)
//
// 	errs := dbDirectoryPrevious.Close()
// 	is.Equal(len(errs), 0) // could not close dbs properly
//
// 	fakeClock.Advance(time.Second * time.Duration(61))
//
// 	dbDirectoryCurrent := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
// 	err := dbDirectoryCurrent.ReadExistingDbs()
// 	is.NoErr(err) // ReadExistingDbs failed
//
// 	_, ok := dbDirectoryCurrent.Get(dbName)
// 	is.True(!ok) // new db manager did not read existing instance
// }
//
// func TestReadExistingDbs_SchedulesRemoval(t *testing.T) {
// 	is := is.New(t)
//
// 	testDir := t.TempDir()
//
// 	dbId := uuid.New().String();
//
// 	fakeClock := clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local))
//
// 	dbDirectoryPrevious := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
//
// 	dbDirectoryPrevious.Open(dbId)
//
// 	errs := dbDirectoryPrevious.Close()
// 	is.Equal(len(errs), 0) // could not close dbs properly
//
// 	fakeClock.Advance(time.Second * time.Duration(30))
//
// 	dbDirectoryCurrent := NewDbDirectory(testDir, time.Second * time.Duration(60), fakeClock)
// 	err := dbDirectoryCurrent.ReadExistingDbs()
// 	is.NoErr(err) // error during reading existing dbs
//
// 	_, ok := dbDirectoryCurrent.Get(dbId)
// 	is.True(ok) // new db manager did not read existing instance
//
// 	exists := FileExists(dbDirectoryCurrent.Path(dbId))
// 	is.True(exists) // file did not exist but should have
//
// 	fakeClock.Advance(31 * time.Second)
// 	time.Sleep(10 * time.Microsecond)
// 	exists = FileExists(dbDirectoryCurrent.Path(dbId))
// 	is.True(!exists) // file did exist but should not have
// }
//
// func FileExists(path string) bool {
// 	_, err := os.Stat(path)
//
// 	return err == nil
// }


