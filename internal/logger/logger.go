package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const logFileName = "watchdog.log"

var (
	mu   sync.Mutex
	file *os.File
)

// Init opens watchdog.log in the current working directory for append.
func Init() error {
	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	file = f
	return nil
}

// Close closes the log file.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return nil
	}
	err := file.Close()
	file = nil
	return err
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
	defer mu.Unlock()
	if file != nil {
		_, _ = file.WriteString(entry)
	}
}
