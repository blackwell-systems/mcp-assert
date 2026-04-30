// Package runner
// execute.go contains the per-assertion execution logic.
//
// runAssertion is the central dispatcher. It handles skip conditions, then
// routes to the appropriate sub-executor based on which assertion block is set:
//
//   - Trajectory: offline sequence check (no server needed)
//   - AssertResources: resources/list or resources/read
//   - AssertPrompts: prompts/list or prompts/get
//   - AssertCompletion: completion/complete
//   - AssertSampling: tool call that triggers server-side sampling
//   - AssertLogging: logging/setLevel + notification capture
//   - Assert (default): tools/call with Expect checks
//
// Each sub-executor follows the same lifecycle:
//  1. Create an MCP client and perform the initialize handshake
//  2. Run setup steps (with variable capture)
//  3. Execute the primary operation
//  4. Evaluate the Expect block
//  5. Return a Result with PASS, FAIL, or SKIP
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
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// initializedClient creates an MCP client from an assertion's server config,
// connects to the server, and performs the initialize handshake. On success,
// the caller is responsible for calling cancel() and mcpClient.Close(). On
// error, all resources are cleaned up before returning.
func initializedClient(
	a assertion.Assertion,
	fixture string,
	timeout time.Duration,
	dockerImage string,
) (context.Context, context.CancelFunc, client.MCPClient, error) {
	return initializedClientFromConfig(a.Server, fixture, timeout, dockerImage)
}

// initializedClientFromConfig creates an MCP client from a ServerConfig,
// connects, and performs the initialize handshake. Use this when you have a
// bare ServerConfig (e.g., coverage command) rather than a full Assertion.
func initializedClientFromConfig(
	server assertion.ServerConfig,
	fixture string,
	timeout time.Duration,
	dockerImage string,
) (context.Context, context.CancelFunc, client.MCPClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	mcpClient, err := createMCPClient(server, fixture, dockerImage)
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ClientInfo = mcp.Implementation{Name: "mcp-assert", Version: "1.0"}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION

	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		_ = mcpClient.Close()
		cancel()
		return nil, nil, nil, fmt.Errorf("MCP initialize failed: %w", err)
	}

	return ctx, cancel, mcpClient, nil
}

// passResult creates a PASS result with the given name and timing.
func passResult(name string, start time.Time) assertion.Result {
	return assertion.Result{
		Name:     name,
		Status:   assertion.StatusPass,
		Duration: time.Since(start),
	}
}

// failResult creates a FAIL result with the given name, timing, and detail message.
func failResult(name string, start time.Time, detail string) assertion.Result {
	return assertion.Result{
		Name:     name,
		Status:   assertion.StatusFail,
		Detail:   detail,
		Duration: time.Since(start),
	}
}

// skipResult creates a SKIP result with the given name, timing, and reason.
func skipResult(name string, start time.Time, detail string) assertion.Result {
	return assertion.Result{
		Name:     name,
		Status:   assertion.StatusSkip,
		Detail:   detail,
		Duration: time.Since(start),
	}
}

// runSetupSteps executes setup tool calls sequentially, capturing values from
// responses via JSONPath. Returns the captured variables map for substitution
// into subsequent calls.
func runSetupSteps(
	ctx context.Context,
	mcpClient client.MCPClient,
	steps []assertion.ToolCall,
	fixture string,
) (map[string]string, error) {
	captured := make(map[string]string)
	for _, step := range steps {
		stepArgs := substituteAll(step.Args, fixture, captured)
		req := mcp.CallToolRequest{}
		req.Params.Name = step.Tool
		req.Params.Arguments = stepArgs
		stepResult, err := mcpClient.CallTool(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("setup step %s failed: %w", step.Tool, err)
		}
		if len(step.Capture) > 0 && stepResult != nil {
			responseText := extractText(stepResult)
			for varName, jsonPath := range step.Capture {
				val, err := extractJSONPath(responseText, jsonPath)
				if err != nil {
					return nil, fmt.Errorf("setup step %s: capture %q from %q failed: %w", step.Tool, varName, jsonPath, err)
				}
				captured[varName] = val
			}
		}
	}
	return captured, nil
}

// resolveTimeout returns the per-assertion timeout if set in YAML,
// otherwise falls back to the CLI-provided default.
func resolveTimeout(a assertion.Assertion, cliTimeout time.Duration) time.Duration {
	if a.Timeout != "" {
		if d, err := time.ParseDuration(a.Timeout); err == nil {
			return d
		}
	}
	return cliTimeout
}

// runAssertion dispatches a single assertion to the appropriate executor
// based on which block is set (assert, assert_resources, trajectory, etc.).
// Returns a Result capturing pass/fail status, timing, and failure detail.
func runAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string) assertion.Result {
	start := time.Now()
	timeout = resolveTimeout(a, timeout)

	if a.Skip {
		return skipResult(a.Name, start, "")
	}

	if a.SkipUnlessEnv != "" && os.Getenv(a.SkipUnlessEnv) == "" {
		return skipResult(a.Name, start, fmt.Sprintf("skipped: env var %s not set", a.SkipUnlessEnv))
	}

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

	ctx, cancel, mcpClient, err := initializedClient(a, fixture, timeout, dockerImage)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}
	defer cancel()
	defer mcpClient.Close()

	// Register progress notification handler before setup so it captures
	// notifications from both setup steps and the main tool call.
	var progressCount int32
	if a.Assert.CaptureProgress {
		mcpClient.OnNotification(func(n mcp.JSONRPCNotification) {
			if n.Method == "notifications/progress" {
				atomic.AddInt32(&progressCount, 1)
			}
		})
	}

	captured, err := runSetupSteps(ctx, mcpClient, a.Setup, fixture)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}

	// Snapshot files before the tool call so file_unchanged can compare
	// before vs. after content to verify the tool did not modify them.
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
		return failResult(a.Name, start, fmt.Sprintf("tool call %s failed: %v", a.Assert.Tool, err))
	}

	resultText := extractText(result)
	isError := result.IsError

	if err := assertion.CheckWithSnapshots(a.Assert.Expect, resultText, isError, snapshots); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return failResult(a.Name, start, detail)
	}

	if a.Assert.CaptureProgress {
		if err := assertion.CheckProgress(a.Assert.Expect, int(atomic.LoadInt32(&progressCount))); err != nil {
			return failResult(a.Name, start, err.Error())
		}
	}

	return passResult(a.Name, start)
}

// runResourceAssertion tests MCP resources (resources/list or resources/read).
func runResourceAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	rb := a.AssertResources
	if rb.List == nil && rb.Read == "" && rb.Subscribe == "" {
		return failResult(a.Name, start, "assert_resources requires 'list', 'read', or 'subscribe'")
	}

	ctx, cancel, mcpClient, err := initializedClient(a, fixture, timeout, dockerImage)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}
	defer cancel()
	defer mcpClient.Close()

	if _, err := runSetupSteps(ctx, mcpClient, a.Setup, fixture); err != nil {
		return failResult(a.Name, start, err.Error())
	}

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
			return failResult(a.Name, start, fmt.Sprintf("resources/subscribe failed for %s: %v", rb.Subscribe, err))
		}
	}

	var resultText string
	var isError bool

	if rb.List != nil {
		listReq := mcp.ListResourcesRequest{}
		if rb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(rb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListResources(ctx, listReq)
		if err != nil {
			return failResult(a.Name, start, fmt.Sprintf("resources/list failed: %v", err))
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if rb.Read != "" {
		uri := strings.ReplaceAll(rb.Read, "{{fixture}}", fixture)
		readReq := mcp.ReadResourceRequest{}
		readReq.Params.URI = uri
		result, err := mcpClient.ReadResource(ctx, readReq)
		if err != nil {
			return failResult(a.Name, start, fmt.Sprintf("resources/read failed for %s: %v", uri, err))
		}
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
		return failResult(a.Name, start, "assert_resources requires either 'list' or 'read'")
	}

	if err := assertion.Check(rb.Expect, resultText, isError); err != nil {
		detail := err.Error()
		if isError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return failResult(a.Name, start, detail)
	}

	if rb.ExpectNotification != nil && *rb.ExpectNotification {
		if atomic.LoadInt32(&notificationCount) == 0 {
			return failResult(a.Name, start, "expected resource update notification but received none")
		}
	}

	if rb.Unsubscribe != "" {
		unsubReq := mcp.UnsubscribeRequest{}
		unsubReq.Params.URI = rb.Unsubscribe
		if err := mcpClient.Unsubscribe(ctx, unsubReq); err != nil {
			return failResult(a.Name, start, fmt.Sprintf("resources/unsubscribe failed for %s: %v", rb.Unsubscribe, err))
		}
	}

	return passResult(a.Name, start)
}

// runPromptAssertion tests MCP prompts (prompts/list or prompts/get).
func runPromptAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	pb := a.AssertPrompts
	if pb.List == nil && pb.Get == nil {
		return failResult(a.Name, start, "assert_prompts requires either 'list' or 'get'")
	}

	ctx, cancel, mcpClient, err := initializedClient(a, fixture, timeout, dockerImage)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}
	defer cancel()
	defer mcpClient.Close()

	captured, err := runSetupSteps(ctx, mcpClient, a.Setup, fixture)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}

	var resultText string

	if pb.List != nil {
		listReq := mcp.ListPromptsRequest{}
		if pb.List.Cursor != "" {
			listReq.Params.Cursor = mcp.Cursor(strings.ReplaceAll(pb.List.Cursor, "{{fixture}}", fixture))
		}
		result, err := mcpClient.ListPrompts(ctx, listReq)
		if err != nil {
			return failResult(a.Name, start, fmt.Sprintf("prompts/list failed: %v", err))
		}
		data, _ := json.Marshal(result)
		resultText = string(data)
	} else if pb.Get != nil {
		name := strings.ReplaceAll(pb.Get.Name, "{{fixture}}", fixture)
		for k, v := range captured {
			name = strings.ReplaceAll(name, "{{"+k+"}}", v)
		}
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
			return failResult(a.Name, start, fmt.Sprintf("prompts/get failed for %q: %v", name, err))
		}
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
		return failResult(a.Name, start, err.Error())
	}

	return passResult(a.Name, start)
}

// runCompletionAssertion tests MCP completion/complete.
func runCompletionAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	cb := a.AssertCompletion

	ctx, cancel, mcpClient, err := initializedClient(a, fixture, timeout, dockerImage)
	if err != nil {
		return failResult(a.Name, start, err.Error())
	}
	defer cancel()
	defer mcpClient.Close()

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
		return failResult(a.Name, start, fmt.Sprintf("unsupported completion ref type: %q (use \"ref/prompt\" or \"ref/resource\")", cb.Ref.Type))
	}

	result, err := mcpClient.Complete(ctx, completeReq)
	if err != nil {
		return failResult(a.Name, start, fmt.Sprintf("completion/complete failed: %v", err))
	}

	data, _ := json.Marshal(result)
	resultText := string(data)

	if err := assertion.Check(cb.Expect, resultText, false); err != nil {
		return failResult(a.Name, start, err.Error())
	}

	return passResult(a.Name, start)
}

// runTrajectoryAssertion checks a tool call sequence without starting an MCP server.
// The trace comes from the assertion's inline Trace field or an AuditLog file.
func runTrajectoryAssertion(a assertion.Assertion, fixture string, start time.Time) assertion.Result {
	var trace []assertion.TraceEntry

	if a.AuditLog != "" {
		path := strings.ReplaceAll(a.AuditLog, "{{fixture}}", fixture)
		loaded, err := assertion.LoadAuditLog(path)
		if err != nil {
			return failResult(a.Name, start, fmt.Sprintf("audit_log: %v", err))
		}
		trace = loaded
	} else {
		trace = a.Trace
	}

	if err := assertion.CheckTrajectory(a.Trajectory, trace); err != nil {
		return failResult(a.Name, start, err.Error())
	}

	return passResult(a.Name, start)
}
