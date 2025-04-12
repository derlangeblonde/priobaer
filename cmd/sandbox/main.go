package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/infra"
)

func isNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

const nCourses int = 5 

func main(){
	dbPath := "asgalig.sqlite"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	isNoErr(err)
	defer func() {
		conn, err := db.DB()
		isNoErr(err)
		isNoErr(conn.Close())
		isNoErr(os.Remove(dbPath))
	}()

	db.AutoMigrate(&infra.Course{}, &infra.Participant{}, &infra.Priority{})

	courses := infra.RandomCourses(nCourses)

	isNoErr(db.CreateInBatches(&courses, 100).Error)

	participant := infra.Participant{Prename: "Prename", Surname: "Surname"}
	isNoErr(db.Create(&participant).Error)
	prios := []infra.Priority{{ParticipantID: participant.ID, CourseID: courses[0].ID, Level: 1}, {ParticipantID: participant.ID, CourseID: courses[2].ID, Level: 2}, {ParticipantID: participant.ID, CourseID: courses[1].ID, Level: 3}}
	isNoErr(db.Preload("Courses").CreateInBatches(prios, 100).Error)

	for _, prio := range prios {
		fmt.Printf("P-ID:%d\nC-ID:%d\nL:%d\n'%s'\n", prio.ParticipantID, prio.CourseID, prio.Level, prio.Course.Name)
	}
}
