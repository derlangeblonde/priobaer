package apptest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
	"softbaer.dev/ass/internal/ui"
)

func TestThatParticipantsNamesAreStoredEncrypted(t *testing.T) {
	is := is.New(t)
	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	client := NewTestClient(t, localhost)

	createdParticipant := ui.RandomParticipant()
	client.ParticipantsCreateAction(createdParticipant, make([]int, 0), nil)

	var sqlFiles []string
	err := filepath.Walk(sut.dbDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".sqlite" {
			sqlFiles = append(sqlFiles, path)
		}

		return nil
	})

	is.NoErr(err)              // Want to walk dbDir without errors
	is.Equal(len(sqlFiles), 1) // Want exactly one sql file after creating one session in the sut
	db, err := gorm.Open(sqlite.Open(sqlFiles[0]), &gorm.Config{})
	is.NoErr(err) // Want to be able to open sql file without problems

	var participants []model.Participant
	is.NoErr(db.Find(&participants).Error)                                  // Want querying against db to be successful
	is.Equal(len(participants), 1)                                          // Want exactly one participant
	is.True(participants[0].EncryptedPrename != createdParticipant.Prename) // Want that Name stored in DB is not what we created via client (i.e. value in DB is encrypted)
	is.True(participants[0].EncryptedSurname != createdParticipant.Surname)
}
