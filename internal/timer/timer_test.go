package timer

import (
	"testing"
	"time"
)

func TestLongBreakEveryFourthWorkSession(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	s := NewSession(Settings{WorkMinutes: 5, BreakMinutes: 5, LongBreakMinutes: 15}, now)

	// complete 3 full work+break pairs
	for i := 0; i < 3; i++ {
		now = now.Add(5 * time.Minute)
		s.Tick(now) // work -> break
		now = now.Add(5 * time.Minute)
		s.Tick(now) // break -> work
	}

	// 4th work should go to long break
	now = now.Add(5 * time.Minute)
	transitioned := s.Tick(now)
	if !transitioned {
		t.Fatal("expected transition")
	}
	if s.Phase() != PhaseLongBreak {
		t.Fatalf("expected long break, got %s", s.Phase())
	}
}

func TestPauseResumeFreezesRemaining(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	s := NewSession(Settings{WorkMinutes: 5, BreakMinutes: 5, LongBreakMinutes: 15}, now)

	now = now.Add(2 * time.Minute)
	s.Tick(now)
	s.Pause(now)
	remaining := s.Remaining(now)

	now = now.Add(30 * time.Second)
	s.Tick(now)
	if got := s.Remaining(now); got != remaining {
		t.Fatalf("remaining changed while paused: got=%v want=%v", got, remaining)
	}

	s.Resume(now)
	now = now.Add(30 * time.Second)
	s.Tick(now)
	if got := s.Remaining(now); got >= remaining {
		t.Fatalf("expected remaining to decrease after resume, got=%v want<%v", got, remaining)
	}
}

func TestFocusedAtIncludesUntickedTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	s := NewSession(Settings{WorkMinutes: 5, BreakMinutes: 5, LongBreakMinutes: 15}, now)

	now = now.Add(90 * time.Second)
	focused := s.FocusedAt(now)
	if focused != 90*time.Second {
		t.Fatalf("expected 90s focused, got %v", focused)
	}
}
