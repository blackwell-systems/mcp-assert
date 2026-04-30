package runner

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/mcp"
)

// runNotificationAssertion calls a tool and captures all notifications
// emitted by the server during the call. It then asserts on the captured
// notifications using the NotificationExpect block.
func runNotificationAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	nb := a.AssertNotifications
	if nb.Tool == "" {
		return failResult(a.Name, start, "assert_notifications requires a tool name")
	}

	ctx, cancel, mcpClient, err := initializedClient(a, fixture, timeout, dockerImage)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}
	defer cancel()
	defer mcpClient.Close()

	// Register notification handler to capture all notifications.
	var mu sync.Mutex
	var captured []assertion.CapturedNotification
	mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
		data, _ := json.Marshal(n.Params.AdditionalFields)
		mu.Lock()
		captured = append(captured, assertion.CapturedNotification{
			Method: n.Method,
			Params: string(data),
		})
		mu.Unlock()
	})

	// Run setup steps if any.
	capturedVars, err := runSetupSteps(ctx, mcpClient, a.Setup, fixture)
	if err != nil {
		return failResult(a.Name, start, fmt.Sprintf("setup failed: %v", err))
	}

	// Call the tool.
	assertArgs := substituteAll(nb.Args, fixture, capturedVars)
	req := mcp.CallToolRequest{}
	req.Params.Name = nb.Tool
	req.Params.Arguments = assertArgs

	_, err = mcpClient.CallTool(ctx, req)
	if err != nil {
		return failResult(a.Name, start, fmt.Sprintf("tool call %s failed: %v", nb.Tool, err))
	}

	// Check notification expectations.
	mu.Lock()
	capturedCopy := make([]assertion.CapturedNotification, len(captured))
	copy(capturedCopy, captured)
	mu.Unlock()

	if err := assertion.CheckNotifications(nb.Expect, capturedCopy); err != nil {
		return failResult(a.Name, start, err.Error())
	}

	return passResult(a.Name, start)
}
