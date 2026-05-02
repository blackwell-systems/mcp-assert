package runner

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// Intercept starts a stdio proxy between stdin/stdout and the given server command.
// It captures every tools/call JSON-RPC request, then validates trajectory assertions
// against the captured trace when the connection ends.
func Intercept(args []string) error {
	fs := flag.NewFlagSet("intercept", flag.ContinueOnError)
	serverSpec := fs.String("server", "", "Server command to proxy (required)")
	trajectoryPath := fs.String("trajectory", "", "Path to YAML file with trajectory assertions (required)")
	timeout := fs.Duration("timeout", 0, "Timeout (0 = no timeout, default)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *serverSpec == "" {
		return fmt.Errorf("--server is required")
	}
	if *trajectoryPath == "" {
		return fmt.Errorf("--trajectory is required")
	}

	// Parse server command and args.
	parts := strings.Fields(*serverSpec)
	if len(parts) == 0 {
		return fmt.Errorf("--server must not be empty")
	}
	serverCmd := parts[0]
	serverArgs := parts[1:]

	// Load trajectory assertions from the YAML file.
	suite, err := assertion.LoadSuite(*trajectoryPath)
	if err != nil {
		return fmt.Errorf("loading trajectory file: %w", err)
	}
	if len(suite.Assertions) == 0 {
		return fmt.Errorf("no assertions found in %s", *trajectoryPath)
	}
	trajectory := suite.Assertions[0].Trajectory

	// Run the proxy loop, capturing tool calls.
	// The mutex protects trace from concurrent access: onToolCall appends
	// from the proxy goroutine while the main goroutine reads after completion.
	var mu sync.Mutex
	var trace []assertion.TraceEntry
	onToolCall := func(entry assertion.TraceEntry) {
		mu.Lock()
		trace = append(trace, entry)
		mu.Unlock()
	}

	// proxyStdio returns the exec.Cmd so we can kill the server on timeout.
	type proxyResult struct {
		cmd *exec.Cmd
		err error
	}
	done := make(chan proxyResult, 1)
	go func() {
		cmd, err := proxyStdio(serverCmd, serverArgs, nil, onToolCall)
		done <- proxyResult{cmd: cmd, err: err}
	}()
	if *timeout > 0 {
		timer := time.NewTimer(*timeout)
		defer timer.Stop()
		select {
		case res := <-done:
			if res.err != nil {
				fmt.Fprintf(os.Stderr, "server exited: %v\n", res.err)
			}
		case <-timer.C:
			fmt.Fprintf(os.Stderr, "intercept timed out after %s\n", *timeout)
			// Kill the server process to stop the proxy goroutine.
			// Receive cmd from the channel if available; otherwise the
			// goroutine hasn't returned yet and we wait for it below.
			select {
			case res := <-done:
				if res.cmd != nil && res.cmd.Process != nil {
					_ = res.cmd.Process.Kill()
				}
			default:
				// Goroutine still running; wait for it, then kill.
				res := <-done
				if res.cmd != nil && res.cmd.Process != nil {
					_ = res.cmd.Process.Kill()
				}
			}
		}
	} else {
		if res := <-done; res.err != nil {
			fmt.Fprintf(os.Stderr, "server exited: %v\n", res.err)
		}
	}

	// Safe to read trace now: proxy goroutine has exited.
	mu.Lock()
	traceCopy := make([]assertion.TraceEntry, len(trace))
	copy(traceCopy, trace)
	mu.Unlock()
	trace = traceCopy

	// Print captured trace summary.
	fmt.Printf("\nCaptured %d tool call(s):\n", len(trace))
	for i, entry := range trace {
		argsJSON, _ := json.Marshal(entry.Args)
		fmt.Printf("  [%d] %s %s\n", i+1, entry.Tool, string(argsJSON))
	}

	// Validate trajectory assertions.
	if len(trajectory) == 0 {
		fmt.Println("\nNo trajectory assertions defined.")
		return nil
	}

	fmt.Printf("\nValidating %d trajectory assertion(s):\n", len(trajectory))
	if err := assertion.CheckTrajectory(trajectory, trace); err != nil {
		fmt.Printf("  FAIL: %v\n", err)
		return fmt.Errorf("trajectory validation failed: %w", err)
	}

	fmt.Println("  PASS: all trajectory assertions passed")
	return nil
}

// proxyStdio runs the stdio proxy loop between os.Stdin/os.Stdout and the
// given server process. onToolCall is invoked for every tools/call request seen
// in the agent-to-server direction.
func proxyStdio(
	serverCmd string,
	serverArgs []string,
	serverEnv []string,
	onToolCall func(entry assertion.TraceEntry),
) (*exec.Cmd, error) {
	cmd := exec.Command(serverCmd, serverArgs...)
	if len(serverEnv) > 0 {
		cmd.Env = append(os.Environ(), serverEnv...)
	}

	serverIn, err := cmd.StdinPipe()
	if err != nil {
		return cmd, fmt.Errorf("creating server stdin pipe: %w", err)
	}
	serverOut, err := cmd.StdoutPipe()
	if err != nil {
		return cmd, fmt.Errorf("creating server stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return cmd, fmt.Errorf("starting server %q: %w", serverCmd, err)
	}

	// Channel to signal goroutine completion.
	done := make(chan error, 2)

	// Agent-to-server: read JSON-RPC lines from os.Stdin, inspect for tool calls,
	// forward each line to the server's stdin.
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		// Increase buffer from default 64KB to 1MB to handle large JSON-RPC
		// messages (embedded file content, base64 blobs, etc.).
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Bytes()
			if entry := extractToolCall(line); entry != nil {
				onToolCall(*entry)
			}
			// Write the original line plus newline to server stdin.
			if _, err := fmt.Fprintf(serverIn, "%s\n", line); err != nil {
				break
			}
		}
		serverIn.Close()
		done <- scanner.Err()
	}()

	// Server-to-agent: read from server stdout, write to os.Stdout.
	go func() {
		_, err := io.Copy(os.Stdout, serverOut)
		done <- err
	}()

	// Wait for both goroutines to finish.
	<-done
	<-done

	return cmd, cmd.Wait()
}

// extractToolCall parses a single JSON-RPC line and returns a TraceEntry if
// the message is a tools/call request. Returns nil for non-tool-call messages,
// invalid JSON, or notifications (messages without an id).
func extractToolCall(line []byte) *assertion.TraceEntry {
	if len(line) == 0 {
		return nil
	}

	var msg map[string]any
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil
	}

	// Must have method == "tools/call".
	method, _ := msg["method"].(string)
	if method != "tools/call" {
		return nil
	}

	// Extract params.name and params.arguments.
	params, _ := msg["params"].(map[string]any)
	if params == nil {
		return nil
	}
	toolName, _ := params["name"].(string)
	if toolName == "" {
		return nil
	}

	entry := &assertion.TraceEntry{Tool: toolName}
	if arguments, ok := params["arguments"].(map[string]any); ok {
		entry.Args = arguments
	}
	return entry
}
