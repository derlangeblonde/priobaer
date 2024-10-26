package cmdtest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"softbaer.dev/ass/cmd"
)

type TestContext struct {
	T      *testing.T
	client *http.Client
}

func (c *TestContext) RegisterAction(email, password string) {
	is := is.New(c.T)

	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	resp, err := c.client.PostForm("http://localhost:8080/register", form)

	is.NoErr(err) // post request failed

	is.Equal(resp.StatusCode, 302)

	// c.AssertContains(resp.Body, "login")
}

func (c *TestContext) LoginAction(email, password string) {
	is := is.New(c.T)

	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	resp, err := c.client.PostForm("http://localhost:8080/login", form)
	is.NoErr(err)
	is.Equal(len(resp.Cookies()), 1)

	// is.NoErr(err) // post request failed
	//
	// is.Equal(resp.StatusCode, 200)
	//
	// c.AssertContains(resp.Body, "Index")
}

func TestFlow(t *testing.T) {
	is := is.New(t)

	is.NoErr(StartupSystemUnderTest(t))

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	ctx := TestContext{T: t, client: &client}
	ctx.RegisterAction("user@example.com", "a$$word")
	ctx.LoginAction("user@example.com", "a$$word")
}

func StartupSystemUnderTest(t *testing.T) error {
	db_path := fmt.Sprintf("%s/test.db", t.TempDir())
	go cmd.Run(db_path)
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

func (ctx *TestContext) AssertContains(r io.Reader, substr string) {
	is := is.New(ctx.T)
	content, err := io.ReadAll(r)
	is.NoErr(err)
	if !strings.Contains(string(content), substr) {
		ctx.T.Fatalf("%s does not contain %s", string(content), substr)
	}
}
