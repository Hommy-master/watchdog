package monitor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"watchdog/internal/config"
)

var (
	helperOnce sync.Once
	helperPath string
	helperErr  error
)

func TestMain(m *testing.M) {
	helperOnce.Do(func() {
		helperPath, helperErr = buildHelper()
	})
	if helperErr != nil {
		panic(helperErr)
	}
	os.Exit(m.Run())
}

func buildHelper() (string, error) {
	root, err := moduleRoot()
	if err != nil {
		return "", err
	}

	src := filepath.Join(root, "internal", "testutil", "helper.go")
	out := filepath.Join(os.TempDir(), "watchdog-test-helper")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", out, src)
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return "", err
	} else if len(outBytes) > 0 {
		_ = outBytes
	}
	return out, nil
}

func moduleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
	}
}

func testConfig(t *testing.T, interval, delay int) *config.Config {
	t.Helper()
	cfg := &config.Config{
		Interval: interval,
		Delay:    delay,
		Apps: []config.App{
			{Path: helperPath},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestMonitorStartsProcess(t *testing.T) {
	cfg := testConfig(t, 1, 1)
	mon := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mon.Run(ctx)
	}()

	waitFor(t, 3*time.Second, func() bool {
		return mon.IsRunning(0) && mon.PID(0) > 0
	})

	cancel()
	<-errCh
}

func TestMonitorRestartsAfterExit(t *testing.T) {
	cfg := testConfig(t, 1, 1)
	cfg.Apps[0].Args = []string{"-duration=1"}

	mon := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mon.Run(ctx)
	}()

	waitFor(t, 3*time.Second, func() bool {
		return mon.IsRunning(0)
	})
	firstPID := mon.PID(0)
	if firstPID == 0 {
		t.Fatal("expected first PID")
	}

	// Process exits after 1s; with delay=1s it should restart within a few seconds.
	waitFor(t, 5*time.Second, func() bool {
		pid := mon.PID(0)
		return pid > 0 && pid != firstPID
	})

	cancel()
	<-errCh
}

func TestMonitorStopsChildrenOnShutdown(t *testing.T) {
	cfg := testConfig(t, 1, 5)
	mon := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- mon.Run(ctx)
	}()

	waitFor(t, 3*time.Second, func() bool {
		return mon.IsRunning(0)
	})
	pid := mon.PID(0)

	cancel()
	<-errCh

	waitFor(t, 3*time.Second, func() bool {
		return !processAlive(pid)
	})
}

func TestProcessAlive(t *testing.T) {
	cfg := testConfig(t, 1, 1)
	mon := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mon.Run(ctx)
	}()

	waitFor(t, 3*time.Second, func() bool {
		return mon.PID(0) > 0
	})
	pid := mon.PID(0)

	if !processAlive(pid) {
		t.Fatalf("processAlive(%d) = false, want true", pid)
	}

	cancel()
	<-errCh

	waitFor(t, 3*time.Second, func() bool {
		return !processAlive(pid)
	})
}

func TestIsAliveNilCommand(t *testing.T) {
	if isAlive(nil) {
		t.Fatal("isAlive(nil) = true, want false")
	}
}

func TestMonitorRespectsDelay(t *testing.T) {
	cfg := testConfig(t, 1, 3)
	cfg.Apps[0].Args = []string{"-duration=1"}

	mon := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mon.Run(ctx)
	}()

	waitFor(t, 3*time.Second, func() bool {
		return mon.IsRunning(0)
	})

	waitFor(t, 3*time.Second, func() bool {
		return !mon.IsRunning(0)
	})

	// Within the delay window the process must stay down.
	time.Sleep(1500 * time.Millisecond)
	if mon.IsRunning(0) {
		t.Fatal("process restarted before delay elapsed")
	}

	waitFor(t, 4*time.Second, func() bool {
		return mon.IsRunning(0)
	})

	cancel()
	<-errCh
}

func TestMonitorPIDBounds(t *testing.T) {
	mon := New(&config.Config{
		Interval: 1,
		Delay:    1,
		Apps:     []config.App{{Path: helperPath}},
	})

	if mon.IsRunning(-1) || mon.IsRunning(1) {
		t.Fatal("out-of-range index should not be running")
	}
	if mon.PID(-1) != 0 || mon.PID(1) != 0 {
		t.Fatal("out-of-range PID should be 0")
	}
}
