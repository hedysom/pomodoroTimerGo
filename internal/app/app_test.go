package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSetAndGetJSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	if code := Execute([]string{"config", "set", "--work", "30"}, bytes.NewBuffer(nil), out, errOut); code != 0 {
		t.Fatalf("config set failed: code=%d err=%s", code, errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Execute([]string{"config", "get", "--json"}, bytes.NewBuffer(nil), out, errOut); code != 0 {
		t.Fatalf("config get failed: code=%d err=%s", code, errOut.String())
	}

	var cfg map[string]any
	if err := json.Unmarshal(out.Bytes(), &cfg); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if int(cfg["workMinutes"].(float64)) != 30 {
		t.Fatalf("expected workMinutes=30, got %v", cfg["workMinutes"])
	}
	if int(cfg["breakMinutes"].(float64)) != 5 {
		t.Fatalf("expected breakMinutes to remain default 5, got %v", cfg["breakMinutes"])
	}
}

func TestConfigSetRequiresAtLeastOneFlag(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	if code := Execute([]string{"config", "set"}, bytes.NewBuffer(nil), out, errOut); code == 0 {
		t.Fatalf("expected non-zero exit code, got 0")
	}
	if !strings.Contains(errOut.String(), "provide at least one") {
		t.Fatalf("unexpected error: %s", errOut.String())
	}
}

func TestConfigGetWarnsOnMalformedConfigAndUsesDefaults(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	path := filepath.Join(base, "pomodoro", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	if code := Execute([]string{"config", "get"}, bytes.NewBuffer(nil), out, errOut); code != 0 {
		t.Fatalf("config get failed: code=%d err=%s", code, errOut.String())
	}

	if !strings.Contains(errOut.String(), "warning: invalid config") {
		t.Fatalf("expected warning, got: %s", errOut.String())
	}
	if !strings.Contains(out.String(), "workMinutes: 25") {
		t.Fatalf("expected defaults in output, got: %s", out.String())
	}
}
