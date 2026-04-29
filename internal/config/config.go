package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Settings struct {
	WorkMinutes      int  `json:"workMinutes"`
	BreakMinutes     int  `json:"breakMinutes"`
	LongBreakMinutes int  `json:"longBreakMinutes"`
	Bell             bool `json:"bell"`
}

var Defaults = Settings{
	WorkMinutes:      25,
	BreakMinutes:     5,
	LongBreakMinutes: 15,
	Bell:             true,
}

func Validate(s Settings) error {
	if s.WorkMinutes < 5 || s.WorkMinutes > 180 {
		return errors.New("work must be between 5 and 180 minutes")
	}
	if s.BreakMinutes < 5 || s.BreakMinutes > 180 {
		return errors.New("break must be between 5 and 180 minutes")
	}
	if s.LongBreakMinutes < 5 || s.LongBreakMinutes > 180 {
		return errors.New("long-break must be between 5 and 180 minutes")
	}
	return nil
}

func ConfigPath() (string, error) {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			return "", errors.New("APPDATA is not set")
		}
		return filepath.Join(base, "pomodoro", "config.json"), nil
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base != "" {
		return filepath.Join(base, "pomodoro", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pomodoro", "config.json"), nil
}

func Load() (Settings, string, error) {
	path, err := ConfigPath()
	if err != nil {
		return Defaults, "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Defaults, "", nil
		}
		return Defaults, "", err
	}

	cfg := Defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Defaults, fmt.Sprintf("warning: invalid config at %s (%v); using defaults", path, err), nil
	}
	if err := Validate(cfg); err != nil {
		return Defaults, fmt.Sprintf("warning: invalid config at %s (%v); using defaults", path, err), nil
	}

	return cfg, "", nil
}

func Save(s Settings) error {
	if err := Validate(s); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
