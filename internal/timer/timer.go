package timer

import "time"

type Phase string

const (
	PhaseWork      Phase = "WORK"
	PhaseBreak     Phase = "BREAK"
	PhaseLongBreak Phase = "LONG BREAK"
)

type Settings struct {
	WorkMinutes      int
	BreakMinutes     int
	LongBreakMinutes int
}

type Session struct {
	settings Settings

	phase     Phase
	deadline  time.Time
	lastTick  time.Time
	paused    bool
	remaining time.Duration

	CompletedWork int
	Focused       time.Duration
}

func NewSession(settings Settings, now time.Time) *Session {
	s := &Session{
		settings: settings,
		phase:    PhaseWork,
		lastTick: now,
	}
	s.deadline = now.Add(s.phaseDuration(PhaseWork))
	return s
}

func (s *Session) phaseDuration(p Phase) time.Duration {
	switch p {
	case PhaseWork:
		return time.Duration(s.settings.WorkMinutes) * time.Minute
	case PhaseLongBreak:
		return time.Duration(s.settings.LongBreakMinutes) * time.Minute
	default:
		return time.Duration(s.settings.BreakMinutes) * time.Minute
	}
}

func (s *Session) Phase() Phase {
	return s.phase
}

func (s *Session) IsPaused() bool {
	return s.paused
}

func (s *Session) Remaining(now time.Time) time.Duration {
	if s.paused {
		if s.remaining < 0 {
			return 0
		}
		return s.remaining
	}
	rem := s.deadline.Sub(now)
	if rem < 0 {
		return 0
	}
	return rem
}

func (s *Session) NextBreakLabel() string {
	nextWorkCount := s.CompletedWork + 1
	if nextWorkCount%4 == 0 {
		return string(PhaseLongBreak)
	}
	return string(PhaseBreak)
}

func (s *Session) PhaseElapsed(now time.Time) time.Duration {
	elapsed := s.phaseDuration(s.phase) - s.Remaining(now)
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func (s *Session) FocusedAt(now time.Time) time.Duration {
	focused := s.Focused
	if s.phase != PhaseWork || s.paused || !now.After(s.lastTick) {
		return focused
	}

	end := now
	if end.After(s.deadline) {
		end = s.deadline
	}
	if end.After(s.lastTick) {
		focused += end.Sub(s.lastTick)
	}
	return focused
}

func (s *Session) Pause(now time.Time) {
	if s.paused {
		return
	}
	s.Tick(now)
	s.remaining = s.deadline.Sub(now)
	if s.remaining < 0 {
		s.remaining = 0
	}
	s.paused = true
}

func (s *Session) Resume(now time.Time) {
	if !s.paused {
		return
	}
	s.paused = false
	s.deadline = now.Add(s.remaining)
	s.lastTick = now
}

func (s *Session) TogglePause(now time.Time) {
	if s.paused {
		s.Resume(now)
		return
	}
	s.Pause(now)
}

func (s *Session) Tick(now time.Time) (transitioned bool) {
	if s.paused {
		s.lastTick = now
		return false
	}

	for {
		if s.phase == PhaseWork && now.After(s.lastTick) {
			end := now
			if end.After(s.deadline) {
				end = s.deadline
			}
			if end.After(s.lastTick) {
				s.Focused += end.Sub(s.lastTick)
			}
		}

		if now.Before(s.deadline) {
			s.lastTick = now
			return transitioned
		}

		transitioned = true
		if s.phase == PhaseWork {
			s.CompletedWork++
			if s.CompletedWork%4 == 0 {
				s.phase = PhaseLongBreak
			} else {
				s.phase = PhaseBreak
			}
		} else {
			s.phase = PhaseWork
		}

		s.lastTick = s.deadline
		s.deadline = s.deadline.Add(s.phaseDuration(s.phase))
	}
}
