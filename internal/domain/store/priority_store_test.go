package store

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func setupTestDb(t *testing.T) *gorm.DB {
	is := is.New(t)
	dbPath := fmt.Sprintf("%s/%s", t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err)
	db.Exec("PRAGMA foreign_keys = ON;")

	db.AutoMigrate(&model.Participant{}, &model.Course{}, &model.Priority{})

	return db
}

func TestCanAddPriorityOfParticipantToCourse(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	participant := model.Participant{Prename: "hans", Surname: "klein"}
	is.NoErr(db.Create(&participant).Error) // could not create a participant

	course := model.Course{Name: "foo", MaxCapacity: 30, MinCapacity: 3}
	is.NoErr(db.Create(&course).Error) // could not create a course

	err := SetPriorities(db, participant.ID, []int{course.ID})
	is.NoErr(err) // SetPriorities failed

	prioritizedCourses, err := GetPriorities(db, participant.ID)
	is.NoErr(err)

	is.Equal(len(prioritizedCourses), 1)
	is.Equal(prioritizedCourses[0].ID, course.ID)
	is.Equal(prioritizedCourses[0].Name, course.Name)
}

func TestSetPrioritiesFailsWithTooManyPriorities(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	participant := model.Participant{Prename: "hans", Surname: "klein"}
	is.NoErr(db.Create(&participant).Error) // could not create a participant

	courseIds := make([]int, model.MaxPriorityLevel+1)

	err := SetPriorities(db, participant.ID, courseIds)
	is.True(err != nil) // want SetPriorities to fail
}

func TestSetPrioritiesOverwritesExistingPriorities(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	participant := model.RandomParticipant()
	is.NoErr(db.Create(&participant).Error) // could not create a participant

	oldPrioritizedCourses := model.RandomCourses(5)
	is.NoErr(db.CreateInBatches(&oldPrioritizedCourses, len(oldPrioritizedCourses)).Error) // could not create courses
	newPrioritizedCourses := model.RandomCourses(5)
	is.NoErr(db.CreateInBatches(&newPrioritizedCourses, len(newPrioritizedCourses)).Error) // could not create courses

	SetPriorities(db, participant.ID, model.MapToCourseId(oldPrioritizedCourses))
	SetPriorities(db, participant.ID, model.MapToCourseId(newPrioritizedCourses))

	prioritizedCourses, err := GetPriorities(db, participant.ID)
	is.NoErr(err)

	is.Equal(len(prioritizedCourses), 5) // want 5 priorities
	for i, course := range prioritizedCourses {
		is.Equal(course.ID, newPrioritizedCourses[i].ID)     // want the courseID of the priority to be the courseID of the new prioritized course
		is.Equal(course.Name, newPrioritizedCourses[i].Name) // want the courseName of the priority to be the courseName of the new prioritized course
	}
}

func TestSetPrioritiesFailsWithNonExistingParticipant(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	err := SetPriorities(db, 1, []int{1})
	is.True(err != nil) // SetPriorities did not fail
}

func TestSetPrioritiesToLengthZeroEffectivelyDeletesPriorities(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	participant := model.RandomParticipant()
	is.NoErr(db.Create(&participant).Error) // could not create a participant

	err := SetPriorities(db, participant.ID, []int{})
	is.NoErr(err) // SetPriorities failed

	prioritizedCourses, err := GetPriorities(db, participant.ID)
	is.NoErr(err)

	is.Equal(len(prioritizedCourses), 0) // want no priorities
}

func TestGetPrioritiesForMultipleReturnsPrioritiesInCorrectOrder(t *testing.T) {
	is := is.New(t)
	db := setupTestDb(t)

	participants := model.RandomParticipants(3)
	is.NoErr(db.Create(&participants).Error) // could not create a participant
	courses := model.RandomCourses(3)
	is.NoErr(db.Create(&courses).Error) // could not create a course

	wantMap := make(map[int][]model.Course)
	wantMap[participants[0].ID] = []model.Course{courses[0], courses[1], courses[2]}
	wantMap[participants[1].ID] = []model.Course{courses[1], courses[2], courses[0]}
	wantMap[participants[2].ID] = []model.Course{courses[2], courses[0], courses[1]}

	var participantIDs []int
	for participantID, wantCourses := range wantMap {
		err := SetPriorities(db, participantID, model.MapToCourseId(wantCourses))
		is.NoErr(err)
		participantIDs = append(participantIDs, participantID)
	}

	gotMap, err := GetPrioritiesForMultiple(db, participantIDs)
	is.NoErr(err)

	is.Equal(len(gotMap), 3) // want 3 entries

	for participantID, gotPrioritizedCourses := range gotMap {
		is.Equal(len(gotPrioritizedCourses), 3) // want 3 priorities per participant
		wantCourses := wantMap[participantID]
		for i, gotCourse := range gotPrioritizedCourses {
			is.Equal(gotCourse.ID, wantCourses[i].ID)     // prioritized courses in wrong order
			is.Equal(gotCourse.Name, wantCourses[i].Name) // prioritized courses in wrong order
		}
	}
}
