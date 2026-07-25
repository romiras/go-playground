package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds benchmark settings loaded from YAML.
type Config struct {
	Hosts      []string `yaml:"hosts"`
	Model      string   `yaml:"model"`
	Output     string   `yaml:"output"`
	TimeoutSec int      `yaml:"timeout_sec"`
	Cleanup    bool     `yaml:"cleanup"`
}

// Load reads and validates a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks required fields.
func (c *Config) Validate() error {
	if len(c.Hosts) == 0 {
		return fmt.Errorf("config: hosts must not be empty")
	}
	for i, h := range c.Hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			return fmt.Errorf("config: hosts[%d] is empty", i)
		}
		c.Hosts[i] = strings.TrimRight(h, "/")
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("config: model must not be empty")
	}
	if c.TimeoutSec <= 0 {
		return fmt.Errorf("config: timeout_sec must be > 0")
	}
	if strings.TrimSpace(c.Output) == "" {
		c.Output = "results.json"
	}
	return nil
}
