package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"cliPomodoro/internal/config"
	"cliPomodoro/internal/timer"
	"cliPomodoro/internal/ui"
)

func Execute(args []string, in io.Reader, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		printRootHelp(out)
		return 0
	}

	switch args[0] {
	case "start":
		return runStart(args[1:], in, out, errOut)
	case "config":
		return runConfig(args[1:], out, errOut)
	case "help", "--help", "-h":
		printRootHelp(out)
		return 0
	default:
		fmt.Fprintf(errOut, "unknown command: %s\n\n", args[0])
		printRootHelp(errOut)
		return 1
	}
}

func runStart(args []string, in io.Reader, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(errOut)

	work := fs.Int("work", 0, "work duration in minutes (5-180)")
	breakMin := fs.Int("break", 0, "break duration in minutes (5-180)")
	longBreak := fs.Int("long-break", 0, "long break duration in minutes (5-180)")
	bell := fs.Bool("bell", false, "enable bell")
	noBell := fs.Bool("no-bell", false, "disable bell")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *bell && *noBell {
		fmt.Fprintln(errOut, "cannot use --bell and --no-bell together")
		return 1
	}

	cfg, warn, err := config.Load()
	if err != nil {
		fmt.Fprintf(errOut, "failed to load config: %v\n", err)
		return 1
	}
	if warn != "" {
		fmt.Fprintln(errOut, warn)
	}

	if *work != 0 {
		cfg.WorkMinutes = *work
	}
	if *breakMin != 0 {
		cfg.BreakMinutes = *breakMin
	}
	if *longBreak != 0 {
		cfg.LongBreakMinutes = *longBreak
	}
	if *bell {
		cfg.Bell = true
	}
	if *noBell {
		cfg.Bell = false
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(errOut, "invalid settings: %v\n", err)
		return 1
	}

	timerSession := timer.NewSession(timer.Settings{
		WorkMinutes:      cfg.WorkMinutes,
		BreakMinutes:     cfg.BreakMinutes,
		LongBreakMinutes: cfg.LongBreakMinutes,
	}, time.Now())

	interactive := false
	if f, ok := in.(*os.File); ok {
		if outF, ok2 := out.(*os.File); ok2 {
			interactive = term.IsTerminal(int(f.Fd())) && term.IsTerminal(int(outF.Fd()))
		}
	}

	var restore func() = func() {}
	keyChan := make(chan byte, 1)
	if interactive {
		if f, ok := in.(*os.File); ok {
			state, err := term.MakeRaw(int(f.Fd()))
			if err != nil {
				fmt.Fprintf(errOut, "warning: could not enable interactive controls (%v); running non-interactive\n", err)
				interactive = false
			} else {
				restore = func() { _ = term.Restore(int(f.Fd()), state) }
				go readKeys(f, keyChan)
			}
		}
	}
	defer restore()

	if !interactive {
		fmt.Fprintln(errOut, "interactive controls unavailable; use Ctrl+C to quit")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	ui.RenderLine(out, timerSession, time.Now(), interactive)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			transitioned := timerSession.Tick(now)
			if transitioned && cfg.Bell {
				fmt.Fprint(out, "\a")
			}
			ui.RenderLine(out, timerSession, now, interactive)
		case key := <-keyChan:
			now := time.Now()
			switch key {
			case 'p', 'P':
				timerSession.TogglePause(now)
				ui.RenderLine(out, timerSession, now, interactive)
			case 'q', 'Q':
				printSummary(out, timerSession, now)
				return 0
			}
		case <-sigChan:
			printSummary(out, timerSession, time.Now())
			return 0
		}
	}
}

func readKeys(f *os.File, keyChan chan<- byte) {
	buf := make([]byte, 1)
	for {
		n, err := f.Read(buf)
		if err != nil || n == 0 {
			return
		}
		keyChan <- buf[0]
	}
}

func printSummary(out io.Writer, s *timer.Session, now time.Time) {
	ui.ClearLine(out)
	fmt.Fprintln(out, "Session summary")
	fmt.Fprintf(out, "- Completed work sessions: %d\n", s.CompletedWork)
	fmt.Fprintf(out, "- Total focused time: %dm\n", int(s.FocusedAt(now).Minutes()))
	fmt.Fprintf(out, "- Current phase: %s (%s elapsed)\n", s.Phase(), ui.FormatDuration(s.PhaseElapsed(now)))
}

func runConfig(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		printConfigHelp(out)
		return 0
	}

	switch args[0] {
	case "set":
		return runConfigSet(args[1:], out, errOut)
	case "get":
		return runConfigGet(args[1:], out, errOut)
	case "help", "--help", "-h":
		printConfigHelp(out)
		return 0
	default:
		fmt.Fprintf(errOut, "unknown config subcommand: %s\n", args[0])
		printConfigHelp(errOut)
		return 1
	}
}

func runConfigSet(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("config set", flag.ContinueOnError)
	fs.SetOutput(errOut)

	work := fs.Int("work", 0, "work duration in minutes (5-180)")
	breakMin := fs.Int("break", 0, "break duration in minutes (5-180)")
	longBreak := fs.Int("long-break", 0, "long break duration in minutes (5-180)")
	bell := fs.Bool("bell", false, "enable bell")
	noBell := fs.Bool("no-bell", false, "disable bell")

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *bell && *noBell {
		fmt.Fprintln(errOut, "cannot use --bell and --no-bell together")
		return 1
	}

	provided := 0
	if *work != 0 {
		provided++
	}
	if *breakMin != 0 {
		provided++
	}
	if *longBreak != 0 {
		provided++
	}
	if *bell || *noBell {
		provided++
	}
	if provided == 0 {
		fmt.Fprintln(errOut, "provide at least one of --work, --break, --long-break, --bell, --no-bell")
		return 1
	}

	cfg, warn, err := config.Load()
	if err != nil {
		fmt.Fprintf(errOut, "failed to load config: %v\n", err)
		return 1
	}
	if warn != "" {
		fmt.Fprintln(errOut, warn)
	}

	if *work != 0 {
		cfg.WorkMinutes = *work
	}
	if *breakMin != 0 {
		cfg.BreakMinutes = *breakMin
	}
	if *longBreak != 0 {
		cfg.LongBreakMinutes = *longBreak
	}
	if *bell {
		cfg.Bell = true
	}
	if *noBell {
		cfg.Bell = false
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(errOut, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintln(out, "config updated")
	return 0
}

func runConfigGet(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("config get", flag.ContinueOnError)
	fs.SetOutput(errOut)
	asJSON := fs.Bool("json", false, "output JSON")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, warn, err := config.Load()
	if err != nil {
		fmt.Fprintf(errOut, "failed to load config: %v\n", err)
		return 1
	}
	if warn != "" {
		fmt.Fprintln(errOut, warn)
	}

	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cfg); err != nil {
			fmt.Fprintf(errOut, "failed to encode json: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintf(out, "workMinutes: %d\n", cfg.WorkMinutes)
	fmt.Fprintf(out, "breakMinutes: %d\n", cfg.BreakMinutes)
	fmt.Fprintf(out, "longBreakMinutes: %d\n", cfg.LongBreakMinutes)
	fmt.Fprintf(out, "bell: %t\n", cfg.Bell)
	return 0
}

func printRootHelp(out io.Writer) {
	msg := []string{
		"pomodoro - lightweight CLI pomodoro timer",
		"",
		"Usage:",
		"  pomodoro start [--work <min>] [--break <min>] [--long-break <min>] [--bell|--no-bell]",
		"  pomodoro config set [--work <min>] [--break <min>] [--long-break <min>] [--bell|--no-bell]",
		"  pomodoro config get [--json]",
		"",
		"Defaults:",
		"  work=25 break=5 long-break=15 bell=true",
	}
	fmt.Fprintln(out, strings.Join(msg, "\n"))
}

func printConfigHelp(out io.Writer) {
	msg := []string{
		"Usage:",
		"  pomodoro config set [--work <min>] [--break <min>] [--long-break <min>] [--bell|--no-bell]",
		"  pomodoro config get [--json]",
	}
	fmt.Fprintln(out, strings.Join(msg, "\n"))
}
