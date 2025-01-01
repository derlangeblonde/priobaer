package cmdtest

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
	"softbaer.dev/ass/cmd"
	"softbaer.dev/ass/model"
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


	var expectedCourses []model.Course

	for i := 0; i < requestCount; i++ {
		expectedCourses = append(expectedCourses, model.RandomCourse())
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

	mockEnv := setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE", strconv.Itoa(maxAgeDefault), "PORT", strconv.Itoa(port))

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, mockEnv, defaultFakeClock())

	err := defaultWaitForReady()
	is.NoErr(err) // Service was not ready

	testClient := NewTestClient(t, localhost)


	expectedCourse := model.RandomCourse()
	testClient.CoursesCreateAction(expectedCourse, nil)

	waitForTerminationDefault(cancel)

	ctx, cancel = context.WithCancel(context.Background())
	defer waitForTerminationDefault(cancel)

	go cmd.Run(ctx, mockEnv, defaultFakeClock())
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

	expectedCourse := model.RandomCourse()
	ctx.CoursesCreateAction(expectedCourse, nil)
	courses := ctx.CoursesIndexAction()

	is.Equal(len(courses), 1)

	is.True(reflect.DeepEqual(courses[0].Name, expectedCourse.Name)) // created and retrieved course should be the same
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
