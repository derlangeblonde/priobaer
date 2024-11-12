package cmdtest

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
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

func TestCreateAndReadCourse(t *testing.T) {
	is := is.New(t)

	err, cancel := StartupSystemUnderTest(t)
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

func StartupSystemUnderTest(t *testing.T) (error, context.CancelFunc) {
	tempDir := t.TempDir()
	dbDir := path.Join(tempDir, "db")

	if err := os.Mkdir(dbDir, fs.ModePerm); err != nil {
		t.Fatalf("Could not make db root dir: %v", err)
	}

	mockEnv := func(s string) string {
		switch s {
		case "DB_ROOT_DIR":
			return dbDir
		default:
			return ""
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, mockEnv)

	return waitForReady(time.Millisecond*200, 4, "http://localhost:8080/health"), cancel
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

func (ctx *TestContext) AssertContains(r io.Reader, substr string) {
	is := is.New(ctx.T)
	content, err := io.ReadAll(r)
	is.NoErr(err)
	if !strings.Contains(string(content), substr) {
		ctx.T.Fatalf("%s does not contain %s", string(content), substr)
	}
}
