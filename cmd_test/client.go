package cmdtest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/util"
	"softbaer.dev/ass/view"
)

type TestClient struct {
	T       *testing.T
	client  *http.Client
	baseUrl *url.URL
}

func NewTestClient(t *testing.T, baseUrl string) *TestClient {
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

	testClient := TestClient{T: t, client: &client, baseUrl: baseUrlParsed}

	return &testClient
}

func (c *TestClient) AcquireSessionCookie() {
	is := is.New(c.T)

	resp, err := c.client.Get(c.Endpoint("index"))
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

func (c *TestClient) ParticpantsCreateAction(participant model.Participant, finish *sync.WaitGroup) model.Participant {
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	req := c.RequestWithFormBody(
		"POST", c.Endpoint("participants"),
		"prename", participant.Prename,
		"surname", participant.Surname,
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("HX-Request", "true")

	resp, err := c.client.Do(req)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	participants, err := unmarshalAll[model.Participant](resp.Body, "participant-")
	is.NoErr(err)

	is.Equal(len(participants), 1)

	return participants[0]
}

func (c *TestClient) CoursesCreateAction(course model.Course, finish *sync.WaitGroup) view.Course {
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	req := c.RequestWithFormBody(
		"POST", c.Endpoint("courses"),
		"name", course.Name,
		"max-capacity", strconv.Itoa(course.MaxCapacity),
		"min-capacity", strconv.Itoa(course.MinCapacity),
	)

	SetHxRequest(req)

	resp, err := c.client.Do(req)

	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 200)
	defer resp.Body.Close()

	courses, err := unmarshalAll[view.Course](resp.Body, "course-")
	is.Equal(len(courses), 1)

	return courses[0]
}

func (c *TestClient) CoursesIndexAction() []view.Course {
	is := is.New(c.T)

	resp, err := c.client.Get(c.Endpoint("assignments"))
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get courses did not return 200
	defer resp.Body.Close()

	courses, err := unmarshalAll[view.Course](resp.Body, "course-")
	c.T.Log(courses)
	is.NoErr(err)

	return courses
}

func (c *TestClient) AssignmentsIndexAction(courseIdSelected util.MaybeInt) []model.Participant {
	is := is.New(c.T)

	endpoint := c.Endpoint("assignments")

	if courseIdSelected.Valid {
		endpoint = endpoint + fmt.Sprintf("?selected-course=%d", courseIdSelected.Value)
	}

	resp, err := c.client.Get(endpoint)
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get assignments did not return 200
	defer resp.Body.Close()

	participants, err := unmarshalAll[model.Participant](resp.Body, "participant-")

	return participants
}

func (c *TestClient) AssignmentsUpdateAction(participantId int, courseId util.MaybeInt) {
	is := is.New(c.T)

	data := []string{"participant-id", strconv.Itoa(participantId)}

	if courseId.Valid {
		data = slices.Concat(data, []string{"course-id", strconv.Itoa(courseId.Value)})
	}

	req := c.RequestWithFormBody("PUT", c.Endpoint("assignments"), data...)

	resp, err := c.client.Do(req)
	is.NoErr(err) // error while doing put request to "assignments"

	is.Equal(resp.StatusCode, 200)
}

func (c *TestClient) Endpoint(path string) string {
	url := url.URL{
		Scheme: c.baseUrl.Scheme,
		Host:   c.baseUrl.Host,
		Path:   path,
	}

	return url.String()
}

func (c *TestClient) RequestWithFormBody(method, url string, args ...string) *http.Request{
	body := BuildFormBody(args...)
	is := is.New(c.T)
	req, err := http.NewRequest(method, url, body)
	is.NoErr(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req 
}

func BuildFormBody(args ...string) io.Reader {
	if len(args)%2 != 0 {
		panic("Expected even number of args for BuildFormBody")
	}

	data := url.Values{}

	for i := 0; i < len(args); i += 2 {
		data.Add(args[i], args[i+1])
	}

	return strings.NewReader(data.Encode())
}

func SetHxRequest(req *http.Request) {
	req.Header.Add("HX-Request", "true")
}
