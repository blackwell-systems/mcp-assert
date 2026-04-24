package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/mcp"
)

func runAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) assertion.Result {
	start := time.Now()

	// Trajectory assertions check a tool call sequence without calling the server.
	if len(a.Trajectory) > 0 {
		return runTrajectoryAssertion(a, fixture, start)
	}

	// Resource assertions call resources/list or resources/read instead of tools/call.
	if a.AssertResources != nil {
		return runResourceAssertion(a, fixture, timeout, dockerImage, start)
	}

	// Prompt assertions call prompts/list or prompts/get instead of tools/call.
	if a.AssertPrompts != nil {
		return runPromptAssertion(a, fixture, timeout, dockerImage, start)
	}

	// Completion assertions call completion/complete.
	if a.AssertCompletion != nil {
		return runCompletionAssertion(a, fixture, timeout, dockerImage, start)
	}

	// Sampling assertions call a tool that triggers server-side sampling.
	if a.AssertSampling != nil {
		return runSamplingAssertion(a, fixture, timeout, dockerImage, start)
	}

	// Logging assertions test logging/setLevel and notifications/message.
	if a.AssertLogging != nil {
		return runLoggingAssertion(a, fixture, timeout, dockerImage, start)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
		}
	}
	defer mcpClient.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil { //nolint:errcheck
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("MCP initialize failed: %v", err),
			Duration: time.Since(start),
		}
	}

	// Register progress notification handler before setup so it's active for the full lifetime.
	var progressCount int32
	if a.Assert.CaptureProgress {
		mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
			if n.Method == "notifications/progress" {
				atomic.AddInt32(&progressCount, 1)
			}
		})
	}

	// Run setup steps with variable capture.
	captured := make(map[string]string) // variable_name -> captured value
	for _, step := range a.Setup {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		stepResult, err := mcpClient.CallTool(ctx, req)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("setup step %s failed: %v", step.Tool, err),
				Duration: time.Since(start),
			}
		}

		// Capture variables from the response.
		if len(step.Capture) > 0 && stepResult != nil {
			responseText := extractText(stepResult)
			for varName, jsonPath := range step.Capture {
				val, err := extractJSONPath(responseText, jsonPath)
				if err != nil {
					return assertion.Result{
						Name:     a.Name,
						Status:   assertion.StatusFail,
						Detail:   fmt.Sprintf("setup step %s: capture %q from %q failed: %v", step.Tool, varName, jsonPath, err),
						Duration: time.Since(start),
					}
				}
				captured[varName] = val
			}
		}
	}

	// Snapshot files for file_unchanged assertions.
	snapshots := make(map[string]string)
	for _, path := range a.Assert.Expect.FileUnchanged {
		p := strings.ReplaceAll(path, "{{fixture}}", fixture)
		if data, err := os.ReadFile(p); err == nil {
			snapshots[p] = string(data)
		}
	}

	// Run the assertion tool call.
	assertArgs := substituteAll(a.Assert.Args, fixture, captured)
	req := mcp.CallToolRequest{}
	req.Params.Name = a.Assert.Tool
	req.Params.Arguments = assertArgs
	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("tool call %s failed: %v", a.Assert.Tool, err),
			Duration: time.Since(start),
		}
	}

	// Extract text from result.
	resultText := extractText(result)
	isError := result.IsError

	// Check assertions (with file snapshots for file_unchanged).
	if err := assertion.CheckWithSnapshots(a.Assert.Expect, resultText, isError, snapshots); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   detail,
			Duration: time.Since(start),
		}
	}

	// Check progress notification count if capture_progress was requested.
	if a.Assert.CaptureProgress {
		if err := assertion.CheckProgress(a.Assert.Expect, int(atomic.LoadInt32(&progressCount))); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   err.Error(),
				Duration: time.Since(start),
			}
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runResourceAssertion tests MCP resources (resources/list or resources/read).
func runResourceAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	// Validate up front — avoids starting the server for a malformed assertion.
	rb := a.AssertResources
	if rb.List == nil && rb.Read == "" && rb.Subscribe == "" {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_resources requires 'list', 'read', or 'subscribe'",
			Duration: time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
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
			Duration: time.Since(start),
		}
	}

	// Run setup steps.
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
				Duration: time.Since(start),
			}
		}
	}

	// Handle resource subscriptions.
	var notificationCount int32
	if rb.Subscribe != "" {
		mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
			if n.Method == "notifications/resources/updated" {
				atomic.AddInt32(&notificationCount, 1)
			}
		})
		subReq := mcp.SubscribeRequest{}
		subReq.Params.URI = rb.Subscribe
		if err := mcpClient.Subscribe(ctx, subReq); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/subscribe failed for %s: %v", rb.Subscribe, err),
				Duration: time.Since(start),
			}
		}
	}

	var resultText string
	var isError bool

	if rb.List != nil {
		// resources/list
		listReq := mcp.ListResourcesRequest{}
		if rb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(rb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListResources(ctx, listReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/list failed: %v", err),
				Duration: time.Since(start),
			}
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if rb.Read != "" {
		// resources/read
		uri := strings.ReplaceAll(rb.Read, "{{fixture}}", fixture)
		readReq := mcp.ReadResourceRequest{}
		readReq.Params.URI = uri
		result, err := mcpClient.ReadResource(ctx, readReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/read failed for %s: %v", uri, err),
				Duration: time.Since(start),
			}
		}
		// Combine all content items into result text.
		var parts []string
		for _, c := range result.Contents {
			switch v := c.(type) {
			case mcp.TextResourceContents:
				parts = append(parts, v.Text)
			case mcp.BlobResourceContents:
				parts = append(parts, fmt.Sprintf("<blob mimeType=%q len=%d>", v.MIMEType, len(v.Blob)))
			default:
				data, _ := json.Marshal(v)
				parts = append(parts, string(data))
			}
		}
		resultText = strings.Join(parts, "\n")
	} else {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_resources requires either 'list' or 'read'",
			Duration: time.Since(start),
		}
	}

	if err := assertion.Check(rb.Expect, resultText, isError); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   detail,
			Duration: time.Since(start),
		}
	}

	// Check notification expectation if set.
	if rb.ExpectNotification != nil && *rb.ExpectNotification {
		if atomic.LoadInt32(&notificationCount) == 0 {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   "expected resource update notification but received none",
				Duration: time.Since(start),
			}
		}
	}

	// Unsubscribe if requested.
	if rb.Unsubscribe != "" {
		unsubReq := mcp.UnsubscribeRequest{}
		unsubReq.Params.URI = rb.Unsubscribe
		if err := mcpClient.Unsubscribe(ctx, unsubReq); err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("resources/unsubscribe failed for %s: %v", rb.Unsubscribe, err),
				Duration: time.Since(start),
			}
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runPromptAssertion tests MCP prompts (prompts/list or prompts/get).
func runPromptAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	// Validate up front — avoids starting the server for a malformed assertion.
	pb := a.AssertPrompts
	if pb.List == nil && pb.Get == nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   "assert_prompts requires either 'list' or 'get'",
			Duration: time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
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
			Duration: time.Since(start),
		}
	}

	// Run setup steps.
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
				Duration: time.Since(start),
			}
		}
	}

	var resultText string

	if pb.List != nil {
		// prompts/list
		listReq := mcp.ListPromptsRequest{}
		if pb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(pb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListPrompts(ctx, listReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("prompts/list failed: %v", err),
				Duration: time.Since(start),
			}
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if pb.Get != nil {
		// prompts/get
		name := strings.ReplaceAll(pb.Get.Name, "{{fixture}}", fixture)
		for k, v := range captured {
			name = strings.ReplaceAll(name, "{{"+k+"}}", v)
		}
		// Substitute captured variables in arguments too.
		args := make(map[string]string, len(pb.Get.Arguments))
		for k, v := range pb.Get.Arguments {
			v = strings.ReplaceAll(v, "{{fixture}}", fixture)
			for varName, varVal := range captured {
				v = strings.ReplaceAll(v, "{{"+varName+"}}", varVal)
			}
			args[k] = v
		}
		getReq := mcp.GetPromptRequest{}
		getReq.Params.Name = name
		getReq.Params.Arguments = args
		result, err := mcpClient.GetPrompt(ctx, getReq)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("prompts/get failed for %q: %v", name, err),
				Duration: time.Since(start),
			}
		}
		// Build result text from messages.
		var parts []string
		if result.Description != "" {
			parts = append(parts, result.Description)
		}
		for _, msg := range result.Messages {
			switch c := msg.Content.(type) {
			case mcp.TextContent:
				parts = append(parts, c.Text)
			default:
				data, _ := json.Marshal(msg.Content)
				parts = append(parts, string(data))
			}
		}
		resultText = strings.Join(parts, "\n")
	}

	if err := assertion.Check(pb.Expect, resultText, false); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runCompletionAssertion tests MCP completion/complete.
func runCompletionAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	cb := a.AssertCompletion

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(a.Server, fixture, dockerImage)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("failed to start MCP server: %v", err),
			Duration: time.Since(start),
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
			Duration: time.Since(start),
		}
	}

	// Build the completion request with the appropriate reference type.
	completeReq := mcp.CompleteRequest{}
	completeReq.Params.Argument = mcp.CompleteArgument{
		Name:  cb.Argument.Name,
		Value: cb.Argument.Value,
	}
	switch cb.Ref.Type {
	case "ref/prompt":
		completeReq.Params.Ref = mcp.PromptReference{
			Type: "ref/prompt",
			Name: cb.Ref.Name,
		}
	case "ref/resource":
		completeReq.Params.Ref = mcp.ResourceReference{
			Type: "ref/resource",
			URI:  cb.Ref.Name,
		}
	default:
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("unsupported completion ref type: %q (use \"ref/prompt\" or \"ref/resource\")", cb.Ref.Type),
			Duration: time.Since(start),
		}
	}

	result, err := mcpClient.Complete(ctx, completeReq)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("completion/complete failed: %v", err),
			Duration: time.Since(start),
		}
	}

	data, _ := json.Marshal(result)
	resultText := string(data)

	if err := assertion.Check(cb.Expect, resultText, false); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// runTrajectoryAssertion checks a tool call sequence without starting an MCP server.
// The trace comes from the assertion's inline Trace field or an AuditLog file.
func runTrajectoryAssertion(a assertion.Assertion, fixture string, start time.Time) assertion.Result {
	var trace []assertion.TraceEntry

	if a.AuditLog != "" {
		// Load trace from agent-lsp JSONL audit log.
		path := strings.ReplaceAll(a.AuditLog, "{{fixture}}", fixture)
		loaded, err := assertion.LoadAuditLog(path)
		if err != nil {
			return assertion.Result{
				Name:     a.Name,
				Status:   assertion.StatusFail,
				Detail:   fmt.Sprintf("audit_log: %v", err),
				Duration: time.Since(start),
			}
		}
		trace = loaded
	} else {
		// Use inline trace from YAML.
		trace = a.Trace
	}

	if err := assertion.CheckTrajectory(a.Trajectory, trace); err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   err.Error(),
			Duration: time.Since(start),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}
