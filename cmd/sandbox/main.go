package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
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

	db.AutoMigrate(&model.Course{}, &model.Participant{}, &model.Priority{})

	courses := model.RandomCourses(nCourses)

	isNoErr(db.CreateInBatches(&courses, 100).Error)

	participant := model.Participant{Prename: "Prename", Surname: "Surname", Priorities: []model.Priority{{CourseID: courses[0].ID, Level: 1}, {CourseID: courses[2].ID, Level: 2}, {CourseID: courses[1].ID, Level: 3}}}
	isNoErr(db.Create(&participant).Error)

	fmt.Println(participant)
	for _, prio := range participant.Priorities {
		fmt.Printf("P-ID:%d\nC-ID:%d\nL:%d\n'%s'\n", prio.ParticipantID, prio.CourseID, prio.Level, prio.Course.Name)
	}
}
