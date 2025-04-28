package cronos

import (
	"net/http"
	"os/exec"
	"sync"
	"time"

	gjallarhorn "github.com/kashari/gjallarhorn/engine"
	"github.com/kashari/golog"
)

var (
	mu           sync.Mutex
	running      bool
	stopTaskChan chan struct{}
	cronMap      = map[string]time.Duration{
		"@every 1s":  time.Second,
		"@every 1m":  time.Minute,
		"@every 5m":  time.Minute * 5,
		"@every 10m": time.Minute * 10,
		"@every 1h":  time.Hour,
		"@every 1d":  time.Hour * 24,
		"@every 1w":  time.Hour * 24 * 7,
		"@every 1M":  time.Hour * 24 * 30,
		"@every 1y":  time.Hour * 24 * 365,
		"@every 1ms": time.Millisecond,
		"@every 1us": time.Microsecond,
		"@every 1ns": time.Nanosecond,
		"@every 1µs": time.Microsecond,
	}
)

// StartCronos starts a cron job that executes a task at specified intervals.
// It takes a gjallarhorn.Context as an argument and retrieves the interval from the query parameters.
// If the interval is not provided or is invalid, or if the job is already running, it returns a 400 Bad Request otherwise it returns a 200 OK status.
//
// The function uses a mutex to ensure that the job is started safely and prevents
// concurrent access to the running variable.
//
// Can be improved by adding a local function map or a cache of function executions for statistical purposes.
// The task to be executed is defined as a function that returns an error.
// Uses a channel to signal when the job should stop.
// The task is executed in a separate goroutine, and the function waits for the specified interval
// before executing the task again.
//
// Here i will be using this in a single scenario but it can be enhanced to be used in multiple scenarios.
func StartCronos(c *gjallarhorn.Context) {
	interval := c.QueryParam("interval")
	if interval == "" {
		c.String(http.StatusBadRequest, "Interval is required")
		return
	}
	if _, ok := cronMap[interval]; !ok {
		c.String(http.StatusBadRequest, "Invalid interval")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if running {
		c.String(http.StatusBadRequest, "Job is already running")
		return
	}

	stopTaskChan = make(chan struct{})

	go startJob(stopTaskChan, func() error {

		cmd := exec.Command("systemctl", "restart", "GoCinema.service")
		if err := cmd.Run(); err != nil {
			golog.Info("Failed to restart service: %v\n", err.Error())
			return err
		}

		golog.Info("Service restarted successfully")
		return nil
	}, interval)

	running = true
	c.String(http.StatusOK, "Job started successfully")
}

// StopCronos stops the cron job if it is running.
// It sends a signal to the stopTaskChan channel to stop the job gracefully.
// If the job is not running, it returns a 400 Bad Request status.
// If the job is stopped successfully, it returns a 200 OK status.
//
// If there is an error while stopping the job, it returns a 500 Internal Server Error status.
// It is important to note that this function should be called from a separate goroutine
// to avoid blocking the main thread.
//
// The function uses a mutex to ensure that the job is stopped safely and prevents
// concurrent access to the running variable.
// It is also important to handle the case where the job is not running
// to avoid sending a signal to a nil channel.
func StopCronos(c *gjallarhorn.Context) {
	mu.Lock()
	defer mu.Unlock()

	if !running {
		c.String(http.StatusBadRequest, "Job is not running")
		return
	}

	close(stopTaskChan)
	running = false
	c.String(http.StatusOK, "Job stopped successfully")
}

func startJob(stopChan chan struct{}, task func() error, interval string) {
	for {
		timer := time.NewTimer(cronMap[interval])

		select {
		case <-timer.C:
			golog.Info("Task executed")
			if err := task(); err != nil {
				golog.Info("Task execution failed: {}", err.Error())
			}
		case <-stopChan:
			timer.Stop()
			golog.Info("Task stopped")
			return
		}
	}
}
