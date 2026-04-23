package runner

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Watch reruns assertions when YAML files in the suite directory change.
func Watch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	fixture := fs.String("fixture", "", "Fixture directory (substituted for {{fixture}})")
	server := fs.String("server", "", "Override server command (e.g. 'agent-lsp go:gopls')")
	interval := fs.Duration("interval", 2*time.Second, "Polling interval for file changes")
	timeout := fs.Duration("timeout", 30*time.Second, "Per-assertion timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" {
		return fmt.Errorf("--suite is required")
	}

	// Build the run args to pass through to Run on each iteration.
	buildRunArgs := func() []string {
		runArgs := []string{"--suite", *suiteDir}
		if *fixture != "" {
			runArgs = append(runArgs, "--fixture", *fixture)
		}
		if *server != "" {
			runArgs = append(runArgs, "--server", *server)
		}
		if *timeout != 30*time.Second {
			runArgs = append(runArgs, "--timeout", timeout.String())
		}
		return runArgs
	}

	// Snapshot mtimes of all YAML files in the suite directory.
	snapshot := func() (map[string]time.Time, error) {
		mtimes := make(map[string]time.Time)
		entries, err := os.ReadDir(*suiteDir)
		if err != nil {
			return nil, fmt.Errorf("reading suite dir: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := filepath.Ext(e.Name())
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			mtimes[e.Name()] = info.ModTime()
		}
		return mtimes, nil
	}

	changed := func(prev, curr map[string]time.Time) bool {
		if len(prev) != len(curr) {
			return true
		}
		for name, mt := range curr {
			if prev[name] != mt {
				return true
			}
		}
		return false
	}

	// Initial run.
	clearScreen()
	fmt.Printf("[watch] Running assertions from %s (polling every %s)\n\n", *suiteDir, *interval)
	_ = Run(buildRunArgs()) // errors are printed by Run

	lastMtimes, err := snapshot()
	if err != nil {
		return err
	}

	// Poll loop.
	for {
		time.Sleep(*interval)

		currentMtimes, err := snapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}

		if !changed(lastMtimes, currentMtimes) {
			continue
		}

		lastMtimes = currentMtimes
		clearScreen()
		fmt.Printf("[watch] Change detected, rerunning at %s\n\n", time.Now().Format("15:04:05"))
		_ = Run(buildRunArgs())
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
