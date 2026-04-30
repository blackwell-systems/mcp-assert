package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/mcp"
)

// runLoggingAssertion tests MCP logging (logging/setLevel + notifications/message).
// It sets the log level, optionally calls a tool to trigger log output,
// and asserts on the captured log messages.
func runLoggingAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	lb := a.AssertLogging

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("MCP initialize failed: %v", err),
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}

	// Register notification handler to capture log messages.
	var mu sync.Mutex
	var logMessages []assertion.LogMessage
	mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
		if n.Method == "notifications/message" {
			// Log message params are in AdditionalFields (level, logger, data).
			data, _ := json.Marshal(n.Params.AdditionalFields)
			var params struct {
				Level  string `json:"level"`
				Logger string `json:"logger"`
				Data   any    `json:"data"`
			}
			json.Unmarshal(data, &params) //nolint:errcheck
			mu.Lock()
			logMessages = append(logMessages, assertion.LogMessage{
				Level:  params.Level,
				Logger: params.Logger,
				Data:   fmt.Sprintf("%v", params.Data),
			})
			mu.Unlock()
		}
	})

	// Set the logging level.
	levelReq := mcp.SetLevelRequest{}
	levelReq.Params.Level = mcp.LoggingLevel(lb.SetLevel)
	if err := mcpClient.SetLevel(ctx, levelReq); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("logging/setLevel failed: %v", err),
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}

	// Run setup steps with variable capture.
	captured := make(map[string]string)
	for _, step := range a.Setup {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("setup step %s failed: %v", step.Tool, err),
				Duration: assertion.DurationMS(time.Since(start)),
			}
		}
	}

	// If a tool is specified, call it to trigger log output.
	if lb.Tool != "" {
		req := mcp.CallToolRequest{}
		req.Params.Name = lb.Tool
		req.Params.Arguments = substituteAll(lb.Args, fixture, captured)
		if _, err := mcpClient.CallTool(ctx, req); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("tool call %s failed: %v", lb.Tool, err),
				Duration: assertion.DurationMS(time.Since(start)),
			}
		}
	}

	// Allow a brief window for any remaining notifications to arrive.
	time.Sleep(100 * time.Millisecond)

	// Check logging assertions.
	mu.Lock()
	msgs := make([]assertion.LogMessage, len(logMessages))
	copy(msgs, logMessages)
	mu.Unlock()

	if err := assertion.CheckLogging(lb.Expect, msgs); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: assertion.DurationMS(time.Since(start)),
	}
}
