# cliPomodoro

Lightweight, platform-independent CLI pomodoro timer in Go.

## Features (v1)

- `pomodoro start` runs until you quit.
- Interactive controls (TTY mode):
  - `p` pause/resume
  - `q` quit immediately
- Long break every 4 completed work sessions.
- Terminal bell (`\a`) on phase transitions (default on, configurable).
- Config stored per-platform:
  - Linux/macOS: `$XDG_CONFIG_HOME/pomodoro/config.json` (fallback `~/.config/pomodoro/config.json`)
  - Windows: `%APPDATA%\\pomodoro\\config.json`
- Config precedence: built-in defaults < config file < CLI flags.
- Malformed config warns and falls back to defaults.

## Defaults

- work: 25 minutes
- break: 5 minutes
- long break: 15 minutes
- bell: true

## Usage

```bash
# run without installing
go run ./cmd/pomodoro start

# install locally
go install ./cmd/pomodoro
pomodoro start
```

### Commands

```bash
pomodoro start [--work <min>] [--break <min>] [--long-break <min>] [--bell|--no-bell]
pomodoro config set [--work <min>] [--break <min>] [--long-break <min>] [--bell|--no-bell]
pomodoro config get [--json]
```

### Validation

All durations must be integers from **5 to 180** minutes.

## Notes

- If stdin/stdout is not an interactive TTY, keyboard controls are disabled and you can stop with Ctrl+C.
- `q` and Ctrl+C both quit immediately and print a summary.
