package main

import (
	"path"
	"slices"
	"testing"

	"softbaer.dev/ass/internal/domain/store"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

const nCourses int = 500
const nParticipants int = 10000
const nPriosPerParticipant = 4
const nAssignedParticipants = 9000

// func BenchmarkGetAllAssByCoursesVariant2(b *testing.B) {
// 	is := is.New(b)
//
// 	testDir := b.TempDir()
// 	db := populateDB(testDir, b)
// 	defer func() {
// 		conn, _ := db.DB()
// 		conn.Close()
// 	}()
//
// 	b.ResetTimer()
//
// 	var courses []model.Course
// 	result := db.Joins("Participants").Joins("Priorities.Course").Find(&courses)
// 	is.NoErr(result.Error)
//
// 	var unassignedParticipants []model.Participant
// 	result = db.Where("course_id is null").Find(&unassignedParticipants)
// 	is.NoErr(result.Error)
//
// 	b.StopTimer()
//
// 	count := 0
// 	for _, course := range courses {
// 		for _, participant := range course.Participants {
// 			is.Equal(len(participant.Priorities), nPriosPerParticipant)
//
// 			count += 1
// 		}
// 	}
//
// 	is.True(count == nAssignedParticipants)
// }

func populateDB(testDir string, t *testing.B) *gorm.DB {
	is := is.New(t)

	dbPath := path.Join(testDir, "db.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err)

	db.AutoMigrate(&model.Course{}, &model.Participant{}, &model.Priority{})

	courses := model.RandomCourses(nCourses)
	participants := model.RandomParticipants(nParticipants)

	db.CreateInBatches(&courses, 100)
	db.CreateInBatches(&participants, 100)

	participantsToAssign := participants[:nAssignedParticipants]

	db.Transaction(func(tx *gorm.DB) error {
		for _, toAssign := range participantsToAssign {
			selectedCourseId := selectRandomCourseId(courses)
			if err := tx.Model(&toAssign).Update("course_id", selectedCourseId).Error; err != nil {
				return err
			}

			var prioritizedCourseIDs []int
			for len(prioritizedCourseIDs) < nPriosPerParticipant {
				selectedCourseId := selectRandomCourseId(courses)
				if slices.Contains(prioritizedCourseIDs, selectedCourseId) {
					continue
				}
				prioritizedCourseIDs = append(prioritizedCourseIDs, selectedCourseId)
			}
			if err := store.SetPriorities(tx, toAssign.ID, prioritizedCourseIDs); err != nil {
				return err
			}
		}

		return nil
	})

	conn, err := db.DB()
	is.NoErr(err)

	conn.Close()

	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err)

	return db
}

func selectRandomCourseId(courses []model.Course) int {
	selectedCourseIndex := model.SeededRand.Intn(nCourses)
	return courses[selectedCourseIndex].ID
}
