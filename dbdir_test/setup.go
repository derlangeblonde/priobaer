package dbdir_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"gorm.io/gorm"
	"softbaer.dev/ass/dbdir"
)

type config struct {
	TmpDir     string
	FakeClock  clockwork.FakeClock
	DbId       uuid.UUID
	Expiration time.Duration
	Models     []any
	T          *testing.T
}

type testData struct {
	gorm.Model
	Number int
}

func newConfig(t *testing.T) *config {
	return &config{
		TmpDir:     t.TempDir(),
		FakeClock:  clockwork.NewFakeClockAt(time.Date(2024, 9, 9, 22, 5, 0, 0, time.Local)),
		DbId:       uuid.New(),
		Expiration: 60 * time.Second,
		Models:     []any{&testData{}},
		T:          t,
	}
}

func (c *config) withExpiration(e time.Duration) *config {
	c.Expiration = e

	return c
}

func (c *config) createSut() *dbdir.DbDirectory {
	sut, err := dbdir.New(c.TmpDir, c.Expiration, c.FakeClock, c.Models)

	if err != nil {
		c.T.Fatalf("Could not create sut, err: %v", err)
	}

	return sut
}
