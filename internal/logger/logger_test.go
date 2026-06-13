package logger

import (
	"bytes"
	"io"
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

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		Close()
		os.Stdout = origStdout
	}()

	Printf("hello %s", "world")
	w.Close()

	var consoleBuf bytes.Buffer
	if _, err := io.Copy(&consoleBuf, r); err != nil {
		t.Fatal(err)
	}

	consoleLine := strings.TrimSpace(consoleBuf.String())
	assertLogLine(t, consoleLine, "hello world")

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatal(err)
	}
	fileLine := strings.TrimSpace(string(data))
	assertLogLine(t, fileLine, "hello world")
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
