package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"cliPomodoro/internal/timer"
)

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func RenderLine(w io.Writer, s *timer.Session, now time.Time, interactive bool) {
	rem := FormatDuration(s.Remaining(now))
	phase := string(s.Phase())
	controls := ""
	if interactive {
		if s.IsPaused() {
			controls = " | (p resume, q quit)"
		} else {
			controls = " | (p pause/resume, q quit)"
		}
	}

	pausedTag := ""
	if s.IsPaused() {
		pausedTag = "[PAUSED]"
	}

	next := ""
	if s.Phase() == timer.PhaseWork {
		next = fmt.Sprintf(" | next break: %s", s.NextBreakLabel())
	}

	line := fmt.Sprintf("[%s]%s %s | completed: %d%s%s", phase, pausedTag, rem, s.CompletedWork, next, controls)
	fmt.Fprintf(w, "\r%-110s", trimLine(line, 110))
}

func ClearLine(w io.Writer) {
	fmt.Fprint(w, "\r\n")
}

func trimLine(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return strings.TrimSpace(s[:max-1]) + "…"
}
