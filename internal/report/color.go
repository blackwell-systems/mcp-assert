package report

import (
	"fmt"
	"os"
	"sync"
)

// ANSI color codes.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	gray   = "\033[90m"
	cyan   = "\033[36m"
)

var (
	colorOnce   sync.Once
	colorCached bool // true when stdout is a TTY and NO_COLOR is unset
)

// ColorEnabled returns true if stdout is a terminal and NO_COLOR is not set.
// The result is computed once and cached for the lifetime of the process.
func ColorEnabled() bool {
	colorOnce.Do(func() {
		if os.Getenv("NO_COLOR") != "" {
			return
		}
		if os.Getenv("TERM") == "dumb" {
			return
		}
		fi, err := os.Stdout.Stat()
		if err != nil {
			return
		}
		// ModeCharDevice is set when stdout is a character device (TTY),
		// absent when piped to a file or another process.
		colorCached = fi.Mode()&os.ModeCharDevice != 0
	})
	return colorCached
}

// colorize wraps text in ANSI color codes when color output is enabled.
// Returns plain text when color is disabled (pipe, NO_COLOR, dumb terminal).
func colorize(color, text string) string {
	if !ColorEnabled() {
		return text
	}
	return color + text + reset
}

func passIcon() string {
	if ColorEnabled() {
		return green + "✓" + reset
	}
	return "PASS"
}

func failIcon() string {
	if ColorEnabled() {
		return red + "✗" + reset
	}
	return "FAIL"
}

func skipIcon() string {
	if ColorEnabled() {
		return yellow + "○" + reset
	}
	return "SKIP"
}

// ProgressLine prints a live progress indicator. Uses \r to overwrite.
func ProgressLine(current, total int, name string) {
	if !ColorEnabled() {
		return
	}
	maxName := 50
	if len(name) > maxName {
		name = name[:maxName-1] + "…"
	}
	fmt.Fprintf(os.Stderr, "\r%s[%d/%d]%s %s", gray, current, total, reset, name)
}

// ClearProgress clears the progress line.
func ClearProgress() {
	if !ColorEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "\r\033[K")
}
