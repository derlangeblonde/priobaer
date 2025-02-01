package dbdir

import (
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"gorm.io/gorm"
)

type DbDirectory struct {
	rootDir                            string
	maxAge                             time.Duration
	gracePeriodBetweenMapAndDiskRemove time.Duration
	entries                            sync.Map
	clock                              clockwork.Clock
	models                             []any
}

type entry struct {
	conn            *gorm.DB
	expirationTimer clockwork.Timer
}

type Session struct {
	gorm.Model
	ExpiresAt time.Time
}
