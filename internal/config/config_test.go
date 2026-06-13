package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"interval": 2,
		"delay": 3,
		"apps": [
			{
				"path": "/usr/bin/demo",
				"workdir": "/opt/demo",
				"args": ["--port=8080"]
			}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Interval != 2 {
		t.Fatalf("Interval = %d, want 2", cfg.Interval)
	}
	if cfg.Delay != 3 {
		t.Fatalf("Delay = %d, want 3", cfg.Delay)
	}
	if cfg.LogOutput != defaultLogOutput {
		t.Fatalf("LogOutput = %q, want %q", cfg.LogOutput, defaultLogOutput)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("len(Apps) = %d, want 1", len(cfg.Apps))
	}
	if cfg.Apps[0].Path != "/usr/bin/demo" {
		t.Fatalf("Apps[0].Path = %q", cfg.Apps[0].Path)
	}
	if cfg.Apps[0].Workdir != "/opt/demo" {
		t.Fatalf("Apps[0].Workdir = %q", cfg.Apps[0].Workdir)
	}
	if len(cfg.Apps[0].Args) != 1 || cfg.Apps[0].Args[0] != "--port=8080" {
		t.Fatalf("Apps[0].Args = %v", cfg.Apps[0].Args)
	}
}

func TestLoadDefaultInterval(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"delay": 1,
		"apps": [{"path": "/bin/app"}]
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Interval != defaultInterval {
		t.Fatalf("Interval = %d, want default %d", cfg.Interval, defaultInterval)
	}
	if cfg.LogOutput != defaultLogOutput {
		t.Fatalf("LogOutput = %q, want %q", cfg.LogOutput, defaultLogOutput)
	}
}

func TestLoadLogOutputFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"log_output": "file",
		"apps": [{"path": "/bin/app"}]
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LogOutput != "file" {
		t.Fatalf("LogOutput = %q, want file", cfg.LogOutput)
	}
}

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty apps",
			cfg:     Config{Interval: 1, Delay: 1, Apps: nil},
			wantErr: "apps must not be empty",
		},
		{
			name:    "missing path",
			cfg:     Config{Interval: 1, Delay: 1, Apps: []App{{Path: ""}}},
			wantErr: "apps[0].path is required",
		},
		{
			name:    "negative delay",
			cfg:     Config{Interval: 1, Delay: -1, Apps: []App{{Path: "/bin/app"}}},
			wantErr: "delay must be >= 0",
		},
		{
			name:    "invalid log output",
			cfg:     Config{Interval: 1, Delay: 1, LogOutput: "syslog", Apps: []App{{Path: "/bin/app"}}},
			wantErr: "log_output must be console or file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Fatal("Validate() expected error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
