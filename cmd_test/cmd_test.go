package cmdtest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"golang.org/x/net/html"
	"softbaer.dev/ass/cmd"
)

type TestContext struct {
	T      *testing.T
	client *http.Client
	baseUrl *url.URL
}

func TestFlow(t *testing.T) {
	is := is.New(t)

	is.NoErr(StartupSystemUnderTest(t))

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
	ctx.CoursesIndexAction()
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

func (c *TestContext) CoursesCreateAction(name string, maxCap, minCap int) {
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

func (c *TestContext) CoursesIndexAction() {
	is := is.New(c.T)

	resp, err := c.client.Get("http://localhost:8080/courses")
	is.NoErr(err) // get request failed
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	is.NoErr(err) // could not parse response html

	divs := findCoursesDivs(doc)
	is.Equal(len(divs), 1)

	// nodeText := getNodeText(divs[0])
}


func StartupSystemUnderTest(t *testing.T) error {
	go cmd.Run()
	return waitForReady(time.Millisecond*200, 4, "http://localhost:8080/health")
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

func getNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(getNodeText(c))
	}
	return text.String()
}


func findCoursesDivs(current *html.Node) []*html.Node {
	if current.Type == html.ElementNode && current.Data == "div" {
		for _, attr := range current.Attr {
			if attr.Key == "id" && strings.Contains(attr.Val, "course-"){
				return []*html.Node{current} 
			}
		}
	}

	alreadyFound := make([]*html.Node, 0)
	for c := current.FirstChild; c != nil; c = c.NextSibling {
		alreadyFound = append(alreadyFound, findCoursesDivs(c)...)
	}

	return alreadyFound 
}

func (ctx *TestContext) AssertContains(r io.Reader, substr string) {
	is := is.New(ctx.T)
	content, err := io.ReadAll(r)
	is.NoErr(err)
	if !strings.Contains(string(content), substr) {
		ctx.T.Fatalf("%s does not contain %s", string(content), substr)
	}
}
