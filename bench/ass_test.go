package bench

import (
	"os"
	"path"
	"testing"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
)

const nCourses int = 500
const nParticipants int = 10000
const nAssignedParticipants = 9000

func BenchmarkGetAllAssByCourses(t *testing.B) {
	is := is.New(t)

	testDir, err := os.MkdirTemp(".", "")
	defer os.RemoveAll(testDir)
	is.NoErr(err)
	db := setup(testDir, t)
	defer func() {
		conn, _ := db.DB()
		conn.Close()
	}()

	t.ResetTimer()

	var courses []model.Course
	result := db.Preload("Participants").Find(&courses)
	is.NoErr(result.Error)

	var unassignedParticipants []model.Participant
	result = db.Where("course_id is null").Find(&unassignedParticipants)
	is.NoErr(result.Error)

	t.StopTimer()

	count := 0
	for _, course := range courses {
		count += len(course.Participants)
	}

	is.True(count == nAssignedParticipants)
}

func setup(testDir string, t *testing.B) *gorm.DB {
	is := is.New(t)

	dbPath := path.Join(testDir, "db.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	is.NoErr(err)

	db.AutoMigrate(&model.Course{}, &model.Participant{})

	courses := model.RandomCourses(nCourses)
	participants := model.RandomParticipants(nParticipants)

	db.CreateInBatches(&courses, 100)
	db.CreateInBatches(&participants, 100)

	participantsToAssign := participants[:nAssignedParticipants]

	db.Transaction(func(tx *gorm.DB) error {
		for _, toAssign := range participantsToAssign {
			selectedCourseIndex := model.SeededRand.Intn(nCourses)
			selectedCourseId := courses[selectedCourseIndex].ID
			// toAssign.CourseID = sql.NullInt64{Valid: true, Int64: int64()}
			tx.Model(&toAssign).Update("course_id", selectedCourseId)
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
