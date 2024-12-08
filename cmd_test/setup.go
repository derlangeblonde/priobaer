package cmdtest

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"softbaer.dev/ass/cmd"
)

type SystemUnderTest struct {
	cancel context.CancelFunc
	dbDir  string
}

func StartupSystemUnderTest(t *testing.T, env func(string) string) SystemUnderTest {
	return StartupSystemUnderTestWithFakeClock(t, env, defaultFakeClock())
}

func StartupSystemUnderTestWithFakeClock(t *testing.T, env func(string) string, fakeClock clockwork.Clock) SystemUnderTest {
	dbDir := MakeTestingDbDir(t)

	if env == nil {
		env = setupMockEnv("DB_ROOT_DIR", dbDir, "SESSION_MAX_AGE", strconv.Itoa(maxAgeDefault))
	}

	ctx, cancel := context.WithCancel(context.Background())

	go cmd.Run(ctx, env, fakeClock)

	err := waitForReady(time.Millisecond*200, 8, "http://localhost:8080/health")

	if err != nil {
		t.Fatalf("Application did not boot in specified time span")
	}

	return SystemUnderTest{dbDir: dbDir, cancel: cancel}

}

func MakeTestingDbDir(t *testing.T) string {
	tempDir := t.TempDir()
	dbDir := path.Join(tempDir, "db")

	if err := os.Mkdir(dbDir, fs.ModePerm); err != nil {
		t.Fatalf("Could not make db root dir: %v", err)
	}

	return dbDir
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
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		<-timer.C
	}

	return fmt.Errorf("timeout reached while waiting for endpoint")
}

func waitForTerminationDefault(cancel context.CancelFunc) error {
	return waitForTermination(time.Millisecond*200, 8, "http://localhost:8080/health", cancel)
}

func waitForTermination(
	interval time.Duration,
	retries int,
	endpoint string,
	cancel context.CancelFunc,
) error {
	cancel()
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
			return nil
		}
		if resp.StatusCode == http.StatusOK {
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

func defaultFakeClock() clockwork.FakeClock {
	return clockwork.NewFakeClockAt(time.Date(2001, 1, 1, 12, 5, 0, 0, time.Local))
}
