package report

import (
	"fmt"
	"os"
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

// ColorEnabled returns true if stdout is a terminal and NO_COLOR is not set.
func ColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

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
