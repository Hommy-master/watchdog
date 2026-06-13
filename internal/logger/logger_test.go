package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFormatFile(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWD)

	if err := Init("file"); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Printf("hello %s", "world")

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatal(err)
	}

	assertLogLine(t, strings.TrimSpace(string(data)), "hello world")
}

func TestLogFormatConsole(t *testing.T) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	if err := Init("console"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		Close()
		os.Stdout = orig
	}()

	Printf("console %s", "log")
	w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}

	assertLogLine(t, strings.TrimSpace(buf.String()), "console log")
}

func assertLogLine(t *testing.T, line, wantMessage string) {
	t.Helper()

	if !strings.HasPrefix(line, "[") {
		t.Fatalf("line = %q, want timestamp prefix", line)
	}
	if !strings.Contains(line, "logger_test.go:") {
		t.Fatalf("line = %q, want source file and line", line)
	}
	if !strings.HasSuffix(line, wantMessage) {
		t.Fatalf("line = %q, want message suffix %q", line, wantMessage)
	}
	if strings.Contains(line, "INFO") || strings.Contains(line, "ERROR") {
		t.Fatalf("line = %q, should not contain log level", line)
	}
}
