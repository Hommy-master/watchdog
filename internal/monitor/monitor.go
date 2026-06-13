package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"watchdog/internal/config"
	"watchdog/internal/logger"
)

// Monitor supervises configured processes on a fixed interval.
type Monitor struct {
	interval time.Duration
	delay    time.Duration
	states   []*appState

	stopOnce sync.Once
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

type appState struct {
	app    config.App
	cmd    *exec.Cmd
	deadAt time.Time
	mu     sync.Mutex
}

// New creates a monitor from validated configuration.
func New(cfg *config.Config) *Monitor {
	states := make([]*appState, len(cfg.Apps))
	for i, app := range cfg.Apps {
		states[i] = &appState{app: app}
	}

	return &Monitor{
		interval: time.Duration(cfg.Interval) * time.Second,
		delay:    time.Duration(cfg.Delay) * time.Second,
		states:   states,
		stopCh:   make(chan struct{}),
	}
}

// Run starts all apps, watches them until ctx is cancelled, then stops children.
func (m *Monitor) Run(ctx context.Context) error {
	for _, state := range m.states {
		if err := state.start(); err != nil {
			m.stopAll()
			return fmt.Errorf("start %q: %w", state.app.Path, err)
		}
	}

	m.wg.Add(1)
	go m.loop()

	<-ctx.Done()
	m.shutdown()
	m.wg.Wait()
	return ctx.Err()
}

func (m *Monitor) loop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAll()
		}
	}
}

func (m *Monitor) checkAll() {
	now := time.Now()
	for _, state := range m.states {
		state.check(now, m.delay)
	}
}

func (m *Monitor) shutdown() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
	})
	m.stopAll()
}

func (m *Monitor) stopAll() {
	for _, state := range m.states {
		state.stop()
	}
}

func (s *appState) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if isAlive(s.cmd) {
		return nil
	}

	cmd := exec.Command(s.app.Path, s.app.Args...)
	if s.app.Workdir != "" {
		cmd.Dir = s.app.Workdir
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	s.cmd = cmd
	s.deadAt = time.Time{}
	logger.Printf("started process %q pid=%d", s.app.Path, cmd.Process.Pid)
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

func (s *appState) check(now time.Time, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if isAlive(s.cmd) {
		s.deadAt = time.Time{}
		return
	}

	if s.deadAt.IsZero() {
		s.deadAt = now
		logger.Printf("process %q is not running", s.app.Path)
		return
	}

	if now.Sub(s.deadAt) < delay {
		return
	}

	cmd := exec.Command(s.app.Path, s.app.Args...)
	if s.app.Workdir != "" {
		cmd.Dir = s.app.Workdir
	}
	if err := cmd.Start(); err != nil {
		logger.Printf("failed to restart process %q: %v", s.app.Path, err)
		return
	}

	s.cmd = cmd
	s.deadAt = time.Time{}
	logger.Printf("restarted process %q pid=%d", s.app.Path, cmd.Process.Pid)
	go func() {
		_ = cmd.Wait()
	}()
}

func (s *appState) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd != nil && s.cmd.Process != nil {
		logger.Printf("stopping process %q pid=%d", s.app.Path, s.cmd.Process.Pid)
	}
	stopProcess(s.cmd)
	s.cmd = nil
	s.deadAt = time.Time{}
}

func isAlive(cmd *exec.Cmd) bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}
	return processAlive(cmd.Process.Pid)
}

func stopProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}

// IsRunning reports whether the supervised process at index i is alive.
// It is intended for tests.
func (m *Monitor) IsRunning(index int) bool {
	if index < 0 || index >= len(m.states) {
		return false
	}
	m.states[index].mu.Lock()
	defer m.states[index].mu.Unlock()
	return isAlive(m.states[index].cmd)
}

// PID returns the OS process ID for the supervised process at index i.
// It returns 0 when the process is not running. Intended for tests.
func (m *Monitor) PID(index int) int {
	if index < 0 || index >= len(m.states) {
		return 0
	}
	m.states[index].mu.Lock()
	defer m.states[index].mu.Unlock()
	if m.states[index].cmd == nil || m.states[index].cmd.Process == nil {
		return 0
	}
	return m.states[index].cmd.Process.Pid
}
