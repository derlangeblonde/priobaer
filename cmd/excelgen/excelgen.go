package main

import (
	"fmt"
	"os"
	"slices"

	"golang.org/x/exp/rand"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/dbdir"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/model/loadsave"
)

func main() {
	nParticipants := 1200
	nCourses := 48
	nPrioPerParticipant := 3
	minCap := 5
	maxCap := 25
	secret := crypt.GenerateSecret()
	db, err := dbdir.NewDb(":memory:", []any{model.EmptyParticipantPointer(), &model.Course{}, &model.Priority{}})
	if err != nil {
		panic(err)
	}

	var participants []model.Participant
	for i := range nParticipants {
		participant, err := model.NewParticipant(fmt.Sprintf("Vorname%d", i), fmt.Sprintf("Nachname%d", i), secret)
		if err != nil {
			panic(err)
		}
		if err := db.Create(&participant).Error; err != nil {
			panic(err)
		}
		participants = append(participants, participant)
	}

	var courses []model.Course
	var courseIds []int
	for i := range nCourses {
		course := model.Course{Name: fmt.Sprintf("Kurs%d", i), MinCapacity: minCap, MaxCapacity: maxCap}
		if err := db.Create(&course).Error; err != nil {
			panic(err)
		}
		courses = append(courses, course)
		courseIds = append(courseIds, course.ID)
	}

	for _, p := range participants {
		var chosenCourseIds []int
		for len(chosenCourseIds) < nPrioPerParticipant {
			randomCourseId := courseIds[rand.Intn(len(courseIds))]
			if !slices.Contains(chosenCourseIds, randomCourseId) {
				randomPrio := model.Priority{ParticipantID: p.ID, CourseID: randomCourseId, Level: model.PriorityLevel(len(chosenCourseIds) + 1)}
				chosenCourseIds = append(chosenCourseIds, randomCourseId)
				if err := db.Create(&randomPrio).Error; err != nil {
					panic(err)
				}
			}
		}
	}

	scenario, err := domain.LoadScenario(db, secret)
	if err != nil {
		panic(err)
	}

	excelBytes, err := loadsave.SaveScenarioToExcelFile(scenario)
	err = os.WriteFile("generated.xlsx", excelBytes, 0666)
	if err != nil {
		panic(err)
	}
}
