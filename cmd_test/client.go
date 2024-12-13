package cmdtest

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/matryer/is"
	"golang.org/x/net/html"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/util"
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

func (c *TestClient) ParticpantsCreateAction(participant model.Participant, finish *sync.WaitGroup) {
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	form := url.Values{}
	form.Add("prename", participant.Prename)
	form.Add("surname", participant.Surname)
	resp, err := c.client.PostForm(c.Endpoint("participants"), form)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 303)
	location, err := resp.Location()
	is.NoErr(err) // could not get location of the redirect response

	is.Equal(location.Path, "/assignments")
}

func (c *TestClient) CoursesCreateAction(course model.Course, finish *sync.WaitGroup) {
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	form := url.Values{}
	form.Add("name", course.Name)
	form.Add("max-capacity", strconv.Itoa(course.MaxCapacity))
	form.Add("min-capacity", strconv.Itoa(course.MinCapacity))
	resp, err := c.client.PostForm(c.Endpoint("courses"), form)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 303)
	location, err := resp.Location()
	is.NoErr(err) // could not get location of the redirect response

	is.Equal(location.Path, "/assignments")
}

func (c *TestClient) CoursesIndexAction() []model.Course {
	is := is.New(c.T)

	resp, err := c.client.Get(c.Endpoint("courses"))
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get courses did not return 200
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	is.NoErr(err) // could not parse response html

	courses, err := unmarshalAll[model.Course](doc, "course-")

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

	doc, err := html.Parse(resp.Body)
	is.NoErr(err) // could not parse response html

	participants, err := unmarshalAll[model.Participant](doc, "participant-")

	return participants
}

func (c *TestClient) AssignmentsUpdateAction(participantId int, courseId util.MaybeInt) {
	is := is.New(c.T)

	data := url.Values{}
	data.Add("participant-id", strconv.Itoa(participantId))

	if courseId.Valid {
		data.Add("course-id", strconv.Itoa(courseId.Value))
	}

	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("PUT", c.Endpoint("assignments"), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	is.NoErr(err) // could not assemble put request to "assignments"

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
