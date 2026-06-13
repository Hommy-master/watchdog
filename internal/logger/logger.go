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
	mu  sync.Mutex
	out io.Writer
)

// Init configures log output. Supported values: console, file.
func Init(output string) error {
	mu.Lock()
	defer mu.Unlock()

	switch output {
	case "console":
		out = os.Stdout
	case "file":
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		out = f
	default:
		return fmt.Errorf("unsupported log output: %q", output)
	}
	return nil
}

// Close closes the log file when output is configured to file.
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	f, ok := out.(*os.File)
	if !ok || f == os.Stdout || f == os.Stderr {
		out = nil
		return nil
	}
	err := f.Close()
	out = nil
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
	if out != nil {
		_, _ = io.WriteString(out, entry)
	}
}
