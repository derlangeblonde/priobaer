package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/model"
)

func isNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

const nCourses int = 5

func main() {
	dbPath := "asgalig.sqlite"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	isNoErr(err)
	defer func() {
		conn, err := db.DB()
		isNoErr(err)
		isNoErr(conn.Close())
		isNoErr(os.Remove(dbPath))
	}()

	db.AutoMigrate(&model.Course{}, model.EmptyParticipantPointer(), &model.Priority{})

	courses := model.RandomCourses(nCourses)

	isNoErr(db.CreateInBatches(&courses, 100).Error)

	participant, err := model.NewParticipant("Prename", "Surname", crypt.GenerateSecret())
	isNoErr(err)
	isNoErr(db.Create(&participant).Error)
	prios := []model.Priority{{ParticipantID: participant.ID, CourseID: courses[0].ID, Level: 1}, {ParticipantID: participant.ID, CourseID: courses[2].ID, Level: 2}, {ParticipantID: participant.ID, CourseID: courses[1].ID, Level: 3}}
	isNoErr(db.Preload("Courses").CreateInBatches(prios, 100).Error)

	for _, prio := range prios {
		fmt.Printf("P-ID:%d\nC-ID:%d\nL:%d\n'%s'\n", prio.ParticipantID, prio.CourseID, prio.Level, prio.Course.Name)
	}
}
