package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
hosts:
  - http://127.0.0.1:11434/
model: llama3.2:latest
output: out.json
timeout_sec: 60
cleanup: false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Hosts[0] != "http://127.0.0.1:11434" {
		t.Errorf("host trim = %q", cfg.Hosts[0])
	}
	if cfg.Model != "llama3.2:latest" {
		t.Errorf("model = %q", cfg.Model)
	}
	if cfg.TimeoutSec != 60 {
		t.Errorf("timeout = %d", cfg.TimeoutSec)
	}
	if cfg.Cleanup {
		t.Error("cleanup should be false")
	}
}

func TestValidateEmptyHosts(t *testing.T) {
	cfg := &Config{Model: "m", TimeoutSec: 1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateBadTimeout(t *testing.T) {
	cfg := &Config{Hosts: []string{"http://x"}, Model: "m", TimeoutSec: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
