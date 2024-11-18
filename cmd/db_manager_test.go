package cmd

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
)

func TestDbManagerIgnoresAndRemovesExpiredDbs(t *testing.T) {
	_ = is.New(t)

	testDir := t.TempDir()

	fakeClock := clockwork.NewFakeClock()

	_ = NewDbManager(testDir, time.Second * time.Duration(60), fakeClock)
}


