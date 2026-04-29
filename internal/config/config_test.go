package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, warn, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if warn != "" {
		t.Fatalf("expected no warning, got %q", warn)
	}
	if cfg != Defaults {
		t.Fatalf("expected defaults, got %+v", cfg)
	}
}

func TestLoadMalformedWarnsAndDefaults(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	path := filepath.Join(base, "pomodoro", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, warn, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if warn == "" {
		t.Fatal("expected warning for malformed config")
	}
	if cfg != Defaults {
		t.Fatalf("expected defaults, got %+v", cfg)
	}
}

func TestValidateRange(t *testing.T) {
	bad := Defaults
	bad.WorkMinutes = 4
	if err := Validate(bad); err == nil {
		t.Fatal("expected error for work < 5")
	}

	bad = Defaults
	bad.BreakMinutes = 181
	if err := Validate(bad); err == nil {
		t.Fatal("expected error for break > 180")
	}
}
