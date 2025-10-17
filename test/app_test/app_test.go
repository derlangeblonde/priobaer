package apptest

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"softbaer.dev/ass/internal/util"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/app/server"
	"softbaer.dev/ass/internal/ui"
)

func TestConcurrentRequestsDontCorruptData(t *testing.T) {
	clientCount := 10
	requestCount := 5

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	wg := sync.WaitGroup{}
	wg.Add(clientCount)

	for i := 0; i < clientCount; i++ {
		go CoursesCreateActionConcurrent(requestCount, &wg, t)
	}

	wg.Wait()
}

func CoursesCreateActionConcurrent(requestCount int, outerWg *sync.WaitGroup, t *testing.T) {
	is := is.New(t)

	wg := sync.WaitGroup{}
	wg.Add(requestCount)

	testClient := NewTestClient(t, localhost)

	var expectedCourses []ui.Course

	for i := 0; i < requestCount; i++ {
		expectedCourses = append(expectedCourses, ui.RandomCourse())
	}

	for _, course := range expectedCourses {
		go testClient.CoursesCreateAction(course, &wg)
	}

	wg.Wait()

	actualCourses := testClient.CoursesIndexAction()

	is.Equal(len(actualCourses), len(expectedCourses)) // all courses were created

	var expectedNames []string

	for _, expected := range expectedCourses {
		expectedNames = append(expectedNames, expected.Name)
	}

	for _, actualCourse := range actualCourses {
		actualCourse.ID = 0
		is.True(slices.Contains(expectedNames, actualCourse.Name)) // actualCourse not in expectedCourses
	}

	outerWg.Done()
}

func TestDbsAreDeletedAfterSessionExpired(t *testing.T) {
	is := is.New(t)
	fakeClock := defaultFakeClock()

	sut := StartupSystemUnderTestWithFakeClock(t, nil, fakeClock)
	defer waitForTerminationDefault(sut.cancel)

	NewTestClient(t, localhost)

	dbFilesCount, err := countSQLiteFiles(sut.dbDir)
	is.NoErr(err)             // failure while counting sqlite files
	is.Equal(dbFilesCount, 1) // there should be exactly *one* db-file after first user request

	fakeClock.Advance(maxAgeDefault * time.Second)
	time.Sleep(100 * time.Microsecond)

	dbFilesCount, err = countSQLiteFiles(sut.dbDir)
	is.NoErr(err)             // failure while counting sqlite files
	is.Equal(dbFilesCount, 0) // there should be *no* db-file after expiration period
}

func TestDataIsPersistedBetweenDeployments(t *testing.T) {
	is := is.New(t)

	dbDir := MakeTestingDbDir(t)

	mockEnv := setupMockEnv("PRIOBAER_DB_ROOT_DIR", dbDir, "PRIOBAER_SESSION_MAX_AGE", strconv.Itoa(maxAgeDefault), "PRIOBAER_PORT", strconv.Itoa(port), "PRIOBAER_SECRET", "secret")

	ctx, cancel := context.WithCancel(context.Background())

	go server.Run(ctx, mockEnv, defaultFakeClock())

	err := defaultWaitForReady()
	is.NoErr(err) // Service was not ready

	testClient := NewTestClient(t, localhost)

	expectedCourse := ui.RandomCourse()
	testClient.CoursesCreateAction(expectedCourse, nil)

	waitForTerminationDefault(cancel)

	ctx, cancel = context.WithCancel(context.Background())
	defer waitForTerminationDefault(cancel)

	go server.Run(ctx, mockEnv, defaultFakeClock())
	err = defaultWaitForReady()
	is.NoErr(err) // Service was not ready

	courses := testClient.CoursesIndexAction()

	is.Equal(len(courses), 1)
	is.Equal(courses[0].Name, expectedCourse.Name)
	is.Equal(courses[0].MinCapacity, expectedCourse.MinCapacity)
	is.Equal(courses[0].MaxCapacity, expectedCourse.MaxCapacity)
}

func TestCreateAndReadCourse(t *testing.T) {
	is := is.New(t)

	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	ctx := NewTestClient(t, localhost)

	expectedCourse := ui.RandomCourse()
	ctx.CoursesCreateAction(expectedCourse, nil)
	courses := ctx.CoursesIndexAction()

	is.Equal(len(courses), 1)

	is.True(reflect.DeepEqual(courses[0].Name, expectedCourse.Name)) // created and retrieved course should be the same
}

func TestRoundtripViaSaveLoad(t *testing.T) {
	is := is.New(t)
	sut := StartupSystemUnderTest(t, nil)
	defer waitForTerminationDefault(sut.cancel)

	client1 := NewTestClient(t, localhost)

	var createdCourses []ui.Course
	for range 3 {
		createdCourses = append(createdCourses, client1.CoursesCreateAction(ui.RandomCourse(), nil))
	}
	courseIDs := util.IDs(createdCourses)
	wantParticipant := ui.RandomParticipant()
	client1.ParticipantsCreateAction(wantParticipant, courseIDs, nil)

	savedData := client1.DataSaveAction()

	client2 := NewTestClient(t, localhost)
	is.Equal(len(client2.CoursesIndexAction()), 0) // new session should be empty
	client2.DataLoadAction(savedData)

	gotCourses, gotParticipants := client2.AssignmentsIndexAction()

	is.Equal(len(gotCourses), len(createdCourses))
	// Sort to compare
	sort.Slice(createdCourses, func(i, j int) bool { return createdCourses[i].Name < createdCourses[j].Name })
	sort.Slice(gotCourses, func(i, j int) bool { return gotCourses[i].Name < gotCourses[j].Name })
	for i := range createdCourses {
		is.Equal(gotCourses[i].Name, createdCourses[i].Name)
		is.Equal(gotCourses[i].MinCapacity, createdCourses[i].MinCapacity)
		is.Equal(gotCourses[i].MaxCapacity, createdCourses[i].MaxCapacity)
	}
	// Verify participant
	is.Equal(len(gotParticipants), 1)
	gotParticipant := gotParticipants[0]
	is.Equal(gotParticipant.Prename, wantParticipant.Prename)
	is.Equal(gotParticipant.Surname, wantParticipant.Surname)
	// TODO: Verify priorities - to do that priorities would have to be parsed from the html response first
}

func countSQLiteFiles(dir string) (int, error) {
	count := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".sqlite" {
			count++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}
