package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFormat(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Printf("hello %s", "world")

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "[") {
		t.Fatalf("line = %q, want timestamp prefix", line)
	}
	if !strings.Contains(line, "logger_test.go:") {
		t.Fatalf("line = %q, want source file and line", line)
	}
	if !strings.HasSuffix(line, "hello world") {
		t.Fatalf("line = %q, want message suffix", line)
	}
	if strings.Contains(line, "INFO") || strings.Contains(line, "ERROR") {
		t.Fatalf("line = %q, should not contain log level", line)
	}
}
