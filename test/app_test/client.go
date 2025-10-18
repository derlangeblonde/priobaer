package apptest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/matryer/is"
	"softbaer.dev/ass/internal/ui"
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
	testClient.AcquireSessionCookie()

	return &testClient
}

func (c *TestClient) AcquireSessionCookie() {
	is := is.New(c.T)

	resp, err := c.client.Post(c.Endpoint("sessions"), "application/x-www-form-urlencoded", bytes.NewReader([]byte{}))
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	is.Equal(resp.StatusCode, 303)

	cookies := resp.Cookies()
	is.Equal(len(cookies), 1)

	// Workaround to send cookies along although we are testing with a non-secure local http-server
	for _, cookie := range cookies {
		cookie.Secure = false
	}

	c.client.Jar.SetCookies(c.baseUrl, cookies)
}

func (c *TestClient) ParticipantsCreateAction(participant ui.Participant, prioritizedCourseIDs []int, finish *sync.WaitGroup) ui.Participant {
	// TODO: I want to get rid of paricipant.Priority member and replace it with a map[int]int, therefore I need to change this functions signature and its usages
	if finish != nil {
		defer finish.Done()
	}

	is := is.New(c.T)

	var requestParameters = []string{"prename", participant.Prename, "surname", participant.Surname}
	for _, courseID := range prioritizedCourseIDs {
		requestParameters = append(requestParameters, "prio[]")
		requestParameters = append(requestParameters, strconv.Itoa(courseID))
	}

	req := c.RequestWithFormBody(
		"POST", c.Endpoint("participants"),
		requestParameters...,
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("HX-Request", "true")

	resp, err := c.client.Do(req)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()

	participants, err := unmarshalAll[ui.Participant](resp.Body, "participant-")
	is.NoErr(err)

	is.Equal(len(participants), 1)

	return participants[0]
}

func (c *TestClient) ParticipantsIndexAction() []ui.Participant {
	is := is.New(c.T)

	resp, err := c.client.Get(c.Endpoint("scenario"))
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get participants did not return 200
	defer resp.Body.Close()

	participants, err := unmarshalAll[ui.Participant](resp.Body, "participant-")
	is.NoErr(err)

	return participants
}

func (c *TestClient) ParticipantsDeleteAction(participantId int) {
	is := is.New(c.T)

	req, err := http.NewRequest("DELETE", c.Endpoint(fmt.Sprintf("/participants/%d", participantId)), nil)
	is.NoErr(err)
	resp, err := c.client.Do(req)
	is.NoErr(err)

	is.Equal(resp.StatusCode, 200)
	defer resp.Body.Close()
}

func (c *TestClient) CoursesCreateAction(course ui.Course, finish *sync.WaitGroup) ui.Course {
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

	courses, err := unmarshalAll[ui.Course](resp.Body, "course-")
	is.Equal(len(courses), 1)

	return courses[0]
}

func (c *TestClient) CoursesIndexAction() []ui.Course {
	is := is.New(c.T)

	resp, err := c.client.Get(c.Endpoint("scenario"))
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get courses did not return 200
	defer resp.Body.Close()

	courses, err := unmarshalAll[ui.Course](resp.Body, "course-")
	is.NoErr(err)

	return courses
}

func (c *TestClient) CoursesDeleteAction(courseId int) {
	is := is.New(c.T)

	req, err := http.NewRequest("DELETE", c.Endpoint(fmt.Sprintf("/courses/%d", courseId)), nil)
	is.NoErr(err)

	resp, err := c.client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, 204)
}

func (c *TestClient) AssignmentsIndexAction(queryParams ...string) ([]ui.Course, []ui.Participant) {
	if len(queryParams)%2 != 0 {
		c.T.Fatal("Number of queryParams has to be even")
	}
	is := is.New(c.T)

	endpoint := c.Endpoint("scenario")

	var keyValuePairs []string
	for keyValueSlice := range slices.Chunk(queryParams, 2) {
		keyValuePair := fmt.Sprintf("%s=%s", keyValueSlice[0], keyValueSlice[1])
		keyValuePairs = append(keyValuePairs, keyValuePair)
	}

	queryString := "?" + strings.Join(keyValuePairs, "&")

	endpoint += queryString

	resp, err := c.client.Get(endpoint)
	is.NoErr(err)                  // get request failed
	is.Equal(resp.StatusCode, 200) // get scenario did not return 200
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	is.NoErr(err) // error while reading resp.Body to bytes

	participants, err := unmarshalAll[ui.Participant](bytes.NewReader(bodyBytes), "participant-")
	is.NoErr(err) // error while unmarshalling pariticpants
	courses, err := unmarshalAll[ui.Course](bytes.NewReader(bodyBytes), "course-")
	is.NoErr(err) // error while unmarshalling courses

	return courses, participants
}

func (c *TestClient) SolveAssignmentsAction() {
	is := is.New(c.T)

	req, err := http.NewRequest("PUT", c.Endpoint("assignments"), nil)
	is.NoErr(err) // want to create request successfully
	resp, err := c.client.Do(req)
	defer resp.Body.Close()
	is.NoErr(err)                  // want request to be successful
	is.Equal(resp.StatusCode, 303) // want to be redirected with 303

	loc := resp.Header.Get("Location")
	is.Equal(loc, "/scenario") //  want to redirected to '/scenario'
}

type AssignmentViewUpdate struct {
	courses         []ui.Course
	UnassignedCount UnassignedCount
}

type UnassignedCount struct {
	Updated bool
	Value   int
}

func (c *TestClient) InitialAssignAction(participantId int, courseId int) AssignmentViewUpdate {
	is := is.New(c.T)

	err, bodyBytes := c.ChangeAssignmentRequest("POST", participantId, courseId, is)

	coursesUpdated, err := unmarshalAll[ui.Course](bytes.NewReader(bodyBytes), "course-")
	is.NoErr(err)

	unassignedCount, err := unmarshalUnassignedCount(bytes.NewReader(bodyBytes))
	is.NoErr(err)

	return AssignmentViewUpdate{courses: coursesUpdated, UnassignedCount: unassignedCount}
}

func (c *TestClient) ReassignAction(participantId int, courseId int) AssignmentViewUpdate {
	is := is.New(c.T)

	err, bodyBytes := c.ChangeAssignmentRequest("PUT", participantId, courseId, is)

	coursesUpdated, err := unmarshalAll[ui.Course](bytes.NewReader(bodyBytes), "course-")
	is.NoErr(err)

	return AssignmentViewUpdate{courses: coursesUpdated}
}

func (c *TestClient) UnassignAction(participantId int) AssignmentViewUpdate {
	is := is.New(c.T)

	err, bodyBytes := c.ChangeAssignmentRequest("DELETE", participantId, 0, is)

	coursesUpdated, err := unmarshalAll[ui.Course](bytes.NewReader(bodyBytes), "course-")
	is.NoErr(err)

	unassignedCount, err := unmarshalUnassignedCount(bytes.NewReader(bodyBytes))
	is.NoErr(err)

	return AssignmentViewUpdate{courses: coursesUpdated, UnassignedCount: unassignedCount}
}

// ChangeAssignmentRequest performs a request to either do an initial assign, a reassign or an unassign.
// The method defines which of these. POST is an initial assign, PUT is a reassign and DELETE is an unassign.
func (c *TestClient) ChangeAssignmentRequest(method string, participantId int, courseId int, is *is.I) (error, []byte) {
	var route string
	if method == "DELETE" {
		route = fmt.Sprintf("participants/%d/assignments", participantId)
	} else {
		route = fmt.Sprintf("participants/%d/assignments/%d", participantId, courseId)
	}

	req := c.RequestWithFormBody(method, c.Endpoint(route))

	resp, err := c.client.Do(req)
	is.NoErr(err) // error while doing put request to "scenario"

	is.Equal(resp.StatusCode, 200)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	is.NoErr(err) // error while reading resp.Body to bytes
	return err, bodyBytes
}

func (c *TestClient) CreateCoursesWithAllocationsAction(expectedAllocations []int) map[int][]int {
	courseIdToAssignedParticipantId := make(map[int][]int)

	for _, expectedAlloc := range expectedAllocations {
		course := c.CoursesCreateAction(ui.RandomCourse(), nil)
		courseIdToAssignedParticipantId[course.ID] = make([]int, 0)

		for i := 0; i < expectedAlloc; i++ {
			participant := c.ParticipantsCreateAction(ui.RandomParticipant(), make([]int, 0), nil)
			c.InitialAssignAction(participant.ID, course.ID)

			courseIdToAssignedParticipantId[course.ID] = append(courseIdToAssignedParticipantId[course.ID], participant.ID)
		}
	}

	return courseIdToAssignedParticipantId
}

func (c *TestClient) DataSaveAction() []byte {
	is := is.New(c.T)
	resp, err := c.client.Get(c.Endpoint("save"))
	is.NoErr(err) // get request failed
	is.Equal(resp.StatusCode, 200)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	is.NoErr(err)

	return data
}

func (c *TestClient) DataLoadAction(data []byte) {
	is := is.New(c.T)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "data.blob")
	is.NoErr(err)
	_, err = io.Copy(part, bytes.NewReader(data))
	is.NoErr(err)
	err = writer.Close()
	is.NoErr(err)
	req, err := http.NewRequest("POST", c.Endpoint("load"), body)
	is.NoErr(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.client.Do(req)
	is.NoErr(err) // post request failed
	defer resp.Body.Close()
	is.Equal(resp.StatusCode, 303)
}

func (c *TestClient) Endpoint(path string) string {
	url := url.URL{
		Scheme: c.baseUrl.Scheme,
		Host:   c.baseUrl.Host,
		Path:   path,
	}

	return url.String()
}

func (c *TestClient) RequestWithFormBody(method, url string, args ...string) *http.Request {
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
