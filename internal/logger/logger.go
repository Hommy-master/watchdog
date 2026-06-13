package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const logFileName = "watchdog.log"

var (
	mu       sync.Mutex
	file     *os.File
	stdoutWG sync.WaitGroup
)

// Init opens watchdog.log and enables logging to both console and file.
func Init() error {
	disableQuickEdit()

	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	file = f
	return nil
}

// Close closes the log file after pending console writes finish.
func Close() error {
	syncStdout()

	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return nil
	}
	err := file.Close()
	file = nil
	return err
}

// SyncStdout waits until queued console writes finish. It is intended for tests.
func SyncStdout() {
	syncStdout()
}

func syncStdout() {
	stdoutWG.Wait()
}

// Printf writes a formatted log line.
func Printf(format string, args ...any) {
	write(2, fmt.Sprintf(format, args...))
}

// Println writes a log line from values.
func Println(args ...any) {
	write(2, fmt.Sprintln(args...))
}

// Fatalf writes a log line and exits with status 1.
func Fatalf(format string, args ...any) {
	write(2, fmt.Sprintf(format, args...))
	syncStdout()
	os.Exit(1)
}

func write(skip int, message string) {
	_, srcFile, line, ok := runtime.Caller(skip)
	fileName := "?"
	if ok {
		fileName = filepath.Base(srcFile)
	}

	entry := fmt.Sprintf("[%s] %s:%d %s", time.Now().Format("2006-01-02 15:04:05"), fileName, line, message)
	if entry[len(entry)-1] != '\n' {
		entry += "\n"
	}

	mu.Lock()
	if file != nil {
		_, _ = io.WriteString(file, entry)
	}
	mu.Unlock()

	stdoutWG.Add(1)
	go func() {
		defer stdoutWG.Done()
		_, _ = io.WriteString(os.Stdout, entry)
	}()
}
