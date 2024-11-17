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
	"strconv"
	"testing"
	"time"

	"github.com/matryer/is"
	"golang.org/x/net/html"
	"softbaer.dev/ass/cmd"
)

type TestContext struct {
	T       *testing.T
	client  *http.Client
	baseUrl *url.URL
}

func TestDbsAreDeletedAfterSessionExpired(t *testing.T) {
	is := is.New(t)

	dbDir := MakeTestingDbDir(t)

	mockEnv := setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE_SECONDS", "1")

	err, cancel := StartupSystemUnderTest(t, mockEnv)
	defer cancel()
	is.NoErr(err)
}

func TestDataIsPersistedBetweenDeployments(t *testing.T) {
	is := is.New(t)

	dbDir := MakeTestingDbDir(t)

	mockEnv := setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE_SECONDS", "3600")

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, mockEnv)

	err := waitForReady(time.Millisecond*200, 4, "http://localhost:8080/health")
	is.NoErr(err) // Service was not ready

	client := CreateTestHttpClient(t)

	baseUrl, err := url.Parse("http://localhost:8080")
	is.NoErr(err) // could not parse baseUrl

	testCtx := TestContext{T: t, client: client, baseUrl: baseUrl}

	testCtx.AcquireSessionCookie()
	testCtx.CoursesCreateAction("foo", 5, 25)

	cancel()
	waitForTermination(time.Millisecond*200, 4, "http://localhost:8080/health")

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go cmd.Run(ctx, mockEnv)
	err = waitForReady(time.Millisecond*200, 4, "http://localhost:8080/health")
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

	jar, err := cookiejar.New(nil)
	is.NoErr(err) // create cookie jar failed

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	baseUrl, err := url.Parse("http://localhost:8080")
	is.NoErr(err) // could not parse baseUrl

	ctx := TestContext{T: t, client: &client, baseUrl: baseUrl}

	ctx.AcquireSessionCookie()
	ctx.CoursesCreateAction("foo", 5, 25)
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

func (c *TestContext) CoursesCreateAction(name string, minCap, maxCap int) {
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
		err := unmarshal[cmd.Course](&course, div)
		is.NoErr(err) // something went wrong during unmarshalling from html (duh!)

		courses = append(courses, course)
	}

	return courses
}

func StartupSystemUnderTest(t *testing.T, env func (string) string) (error, context.CancelFunc) {
	dbDir := MakeTestingDbDir(t)

	if env == nil {
		env = setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE_SECONDS", "3600")
	}

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, env)

	return waitForReady(time.Millisecond*200, 4, "http://localhost:8080/health"), cancel
}

func MakeTestingDbDir(t *testing.T) string {
	tempDir := t.TempDir()
	dbDir := path.Join(tempDir, "db")

	if err := os.Mkdir(dbDir, fs.ModePerm); err != nil {
		t.Fatalf("Could not make db root dir: %v", err)
	}

	return dbDir
}

func CreateTestHttpClient(t *testing.T) *http.Client{
	is := is.New(t)
	jar, err := cookiejar.New(nil)
	is.NoErr(err) // create cookie jar failed

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	return &client
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
