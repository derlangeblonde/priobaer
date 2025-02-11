package model

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
)


func TestCanAddPriorityOfParticipantToCourse(t *testing.T) {
	is := is.New(t)

	dbPath := fmt.Sprintf("%s/%s" ,t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err)

	db.AutoMigrate(&Participant{}, &Course{}, &Priority{})

	participant := Participant{Prename: "hans", Surname: "klein"}
	is.NoErr(db.Create(&participant).Error) // could not create a participant

	course := Course{Name: "foo", MaxCapacity: 30, MinCapacity: 3}
	is.NoErr(db.Create(&course).Error) // could not create a course

	prio := Priority{CourseID: course.ID, ParticipantID: participant.ID, Level: 1}
	is.NoErr(db.Create(&prio).Error) // could not create a priority 

	db.Preload("Priorities").First(&participant)

	is.Equal(participant.Priorities[0].CourseID, course.ID) // want courseID of participants first priority to be the course for which we added the priority
	is.Equal(participant.Priorities[0].Level , PriorityLevel(1)) // want priority level to be 1 
}
