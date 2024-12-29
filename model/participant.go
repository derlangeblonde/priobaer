package model

import (
	"database/sql"

	"gorm.io/gorm"
)

type Participant struct {
	gorm.Model
	ID       int
	Prename  string
	Surname  string
	CourseID sql.NullInt64
	Course   Course `gorm:"constraint:OnDelete:SET NULL;"`
}

