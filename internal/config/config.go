package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	defaultInterval  = 1
	defaultLogOutput = "console"
)

// Config is the root watchdog configuration loaded from JSON.
type Config struct {
	Interval  int    `json:"interval"`
	Delay     int    `json:"delay"`
	LogOutput string `json:"log_output"`
	Apps      []App  `json:"apps"`
}

// App describes a single process to supervise.
type App struct {
	Path    string   `json:"path"`
	Workdir string   `json:"workdir"`
	Args    []string `json:"args"`
}

// Load reads and validates configuration from a JSON file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate applies defaults and checks required fields.
func (c *Config) Validate() error {
	if c.Interval <= 0 {
		c.Interval = defaultInterval
	}
	if c.Delay < 0 {
		return fmt.Errorf("delay must be >= 0")
	}
	if c.LogOutput == "" {
		c.LogOutput = defaultLogOutput
	}
	switch c.LogOutput {
	case "console", "file":
	default:
		return fmt.Errorf("log_output must be console or file")
	}
	if len(c.Apps) == 0 {
		return fmt.Errorf("apps must not be empty")
	}
	for i, app := range c.Apps {
		if app.Path == "" {
			return fmt.Errorf("apps[%d].path is required", i)
		}
	}
	return nil
}
