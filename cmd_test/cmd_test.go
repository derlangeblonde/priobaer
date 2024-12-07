package cmdtest

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/matryer/is"
	"golang.org/x/net/html"
	"softbaer.dev/ass/cmd"
)

const localhost8080 string = "http://localhost:8080"
const aDayInSeconds = 60 * 60 * 24

func defaultFakeClock() clockwork.FakeClock {
	return clockwork.NewFakeClockAt(time.Date(2001, 1, 1, 12, 5, 0, 0, time.Local))
}

type TestContext struct {
	T       *testing.T
	client  *http.Client
	baseUrl *url.URL
}

func TestConcurrentRequests(t *testing.T) {
	requestCount := 10

	is := is.New(t)

	err, cancel := StartupSystemUnderTest(t, nil)
	defer cancel()
	is.NoErr(err)

	wg := sync.WaitGroup{}
	wg.Add(requestCount)

	testCtx := NewTestContext(t, localhost8080)

	testCtx.AcquireSessionCookie()

	for i := 0; i < requestCount; i++ {
		go testCtx.CoursesCreateAction(fmt.Sprintf("test%d", i), 5, 10, &wg)	
	}
	
	wg.Wait()
}

func TestDbsAreDeletedAfterSessionExpired(t *testing.T) {
	is := is.New(t)
	fakeClock := defaultFakeClock()

	dbDir := MakeTestingDbDir(t)

	mockEnv := setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE", strconv.Itoa(aDayInSeconds))

	err, cancel := StartupSystemUnderTestWithFakeClock(t, mockEnv, fakeClock)
	defer cancel()
	is.NoErr(err)

	testCtx := NewTestContext(t, localhost8080)

	testCtx.AcquireSessionCookie()

	dbFilesCount, err := countSQLiteFiles(dbDir)
	is.NoErr(err) // failure while counting sqlite files
	is.Equal(dbFilesCount, 1) // there should be exactly *one* db-file after first user request

	fakeClock.Advance(aDayInSeconds * time.Second)
	time.Sleep(50 * time.Microsecond)

	dbFilesCount, err = countSQLiteFiles(dbDir)
	is.NoErr(err) // failure while counting sqlite files
	is.Equal(dbFilesCount, 0) // there should be *no* db-file after expiration period
} 

func TestDataIsPersistedBetweenDeployments(t *testing.T) {
	is := is.New(t)

	dbDir := MakeTestingDbDir(t)

	mockEnv := setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE", strconv.Itoa(aDayInSeconds))

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, mockEnv, defaultFakeClock())

	err := defaultWaitForReady()
	is.NoErr(err) // Service was not ready

	testCtx := NewTestContext(t, "http://localhost:8080")

	testCtx.AcquireSessionCookie()
	testCtx.CoursesCreateAction("foo", 5, 25, nil)

	cancel()
	waitForTermination(time.Millisecond*200, 8, "http://localhost:8080/health")

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go cmd.Run(ctx, mockEnv, defaultFakeClock())
	err = defaultWaitForReady()
	is.NoErr(err) // Service was not ready

	courses := testCtx.CoursesIndexAction()

	is.Equal(len(courses), 1)
	is.Equal(courses[0].Name, "foo")
	is.Equal(courses[0].MinCapacity, 5)
	is.Equal(courses[0].MaxCapacity, 25)
}

func TestCreateAndReadCourse(t *testing.T) {
	is := is.New(t)

	err, cancel := StartupSystemUnderTest(t, nil)
	defer cancel()
	is.NoErr(err)

	ctx := NewTestContext(t, "http://localhost:8080")

	ctx.AcquireSessionCookie()
	ctx.CoursesCreateAction("foo", 5, 25, nil)
	courses := ctx.CoursesIndexAction()

	is.Equal(len(courses), 1)
	is.Equal(courses[0].Name, "foo")
	is.Equal(courses[0].MinCapacity, 5)
	is.Equal(courses[0].MaxCapacity, 25)
}

func (c *TestContext) AcquireSessionCookie() {
	is := is.New(c.T)

	resp, err := c.client.Get("http://localhost:8080/index")
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 200)

	cookies := resp.Cookies()
	is.Equal(len(cookies), 1)

	// Workaround to send cookies along although we are testing with a non-secure local http-server
	for _, cookie := range cookies {
		cookie.Secure = false
	}

	c.client.Jar.SetCookies(c.baseUrl, cookies)
}

func (c *TestContext) CoursesCreateAction(name string, minCap, maxCap int, finish *sync.WaitGroup) {
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	form := url.Values{}
	form.Add("name", name)
	form.Add("max-capacity", strconv.Itoa(maxCap))
	form.Add("min-capacity", strconv.Itoa(minCap))
	resp, err := c.client.PostForm("http://localhost:8080/courses", form)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 303)
	location, err := resp.Location()
	is.NoErr(err) // could not get location of the redirect response

	is.Equal(location.Path, "/courses")
}

func (c *TestContext) CoursesIndexAction() []cmd.Course {
	is := is.New(c.T)

	resp, err := c.client.Get("http://localhost:8080/courses")
	is.NoErr(err) // get request failed
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	is.NoErr(err) // could not parse response html

	divs := findCoursesDivs(doc)
	is.Equal(len(divs), 1)

	courses := make([]cmd.Course, 0)

	for _, div := range divs {
		var course cmd.Course
		err := unmarshal(&course, div)
		is.NoErr(err) // something went wrong during unmarshalling from html (duh!)

		courses = append(courses, course)
	}

	return courses
}

func StartupSystemUnderTest(t *testing.T, env func(string) string) (error, context.CancelFunc) {
	return StartupSystemUnderTestWithFakeClock(t, env, defaultFakeClock())
}

func StartupSystemUnderTestWithFakeClock(t *testing.T, env func(string) string, fakeClock clockwork.Clock) (error, context.CancelFunc) {
	dbDir := MakeTestingDbDir(t)

	if env == nil {
		env = setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE", strconv.Itoa(aDayInSeconds))
	}

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, env, fakeClock)

	return waitForReady(time.Millisecond*200, 8, "http://localhost:8080/health"), cancel
}

func MakeTestingDbDir(t *testing.T) string {
	tempDir := t.TempDir()
	dbDir := path.Join(tempDir, "db")

	if err := os.Mkdir(dbDir, fs.ModePerm); err != nil {
		t.Fatalf("Could not make db root dir: %v", err)
	}

	return dbDir
}

func NewTestContext(t *testing.T, baseUrl string) *TestContext {
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err) // create cookie jar failed

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	baseUrlParsed, err := url.Parse(baseUrl)
	is.NoErr(err) // could not parse baseUrl

	testCtx := TestContext{T: t, client: &client, baseUrl: baseUrlParsed}

	return &testCtx
}

func defaultWaitForReady() error {
	return waitForReady(time.Millisecond*200, 20, "http://localhost:8080/health")
}

func waitForReady(
	interval time.Duration,
	retries int,
	endpoint string,
) error {
	client := http.Client{}
	for i := 0; i < retries; i++ {
		timer := time.NewTimer(interval)
		req, err := http.NewRequest(
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		<-timer.C
	}

	return fmt.Errorf("timeout reached while waiting for endpoint")
}

func waitForTermination(
	interval time.Duration,
	retries int,
	endpoint string,
) error {
	client := http.Client{}
	for i := 0; i < retries; i++ {
		timer := time.NewTimer(interval)
		req, err := http.NewRequest(
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Request unsuccessful, server probably shutdown. Err :%v", err.Error())

			return nil
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is still available!")
			resp.Body.Close()

			continue
		}
		resp.Body.Close()

		<-timer.C
	}

	return fmt.Errorf("timeout reached while waiting for server to shut down")
}

func setupMockEnv(pairs ...string) func(string) string {
	envMap := make(map[string]string)

	for i := 0; i < len(pairs)-1; i += 2 {
		key := pairs[i]
		value := pairs[i+1]
		envMap[key] = value
	}

	return func(s string) string {
		if value, exists := envMap[s]; exists {
			return value
		}
		return ""
	}
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
