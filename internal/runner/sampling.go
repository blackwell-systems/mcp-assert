package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/mcp"
)

// runSamplingAssertion tests MCP sampling as a first-class test subject.
// It configures client_capabilities.sampling from the assertion's MockText/MockModel,
// calls the specified tool (which triggers the server's sampling request),
// and asserts on the final tool result.
func runSamplingAssertion(a assertion.Assertion, fixture string, timeout time.Duration, dockerImage string, start time.Time) assertion.Result {
	sb := a.AssertSampling

	// Build a modified server config that includes sampling capability.
	server := a.Server
	model := sb.MockModel
	if model == "" {
		model = "mock"
	}
	if server.ClientCapabilities.Sampling == nil {
		server.ClientCapabilities.Sampling = &assertion.SamplingConfig{
			Text:       sb.MockText,
			Model:      model,
			StopReason: "end_turn",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpClient, err := createMCPClient(server, fixture, dockerImage)
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

	// Run setup steps with variable capture.
	captured := make(map[string]string)
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
				Duration: assertion.DurationMS(time.Since(start)),
			}
		}

		if len(step.Capture) > 0 && stepResult != nil {
			responseText := extractText(stepResult)
			for varName, jsonPath := range step.Capture {
				val, err := extractJSONPath(responseText, jsonPath)
				if err != nil {
					return assertion.Result{
						Name:     a.Name,
						Status:   assertion.StatusFail,
						Detail:   fmt.Sprintf("setup step %s: capture %q from %q failed: %v", step.Tool, varName, jsonPath, err),
						Duration: assertion.DurationMS(time.Since(start)),
					}
				}
				captured[varName] = val
			}
		}
	}

	// Call the tool that triggers the server's sampling request.
	req := mcp.CallToolRequest{}
	req.Params.Name = sb.Tool
	req.Params.Arguments = substituteAll(sb.Args, fixture, captured)
	result, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   fmt.Sprintf("tool call %s failed: %v", sb.Tool, err),
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}

	resultText := extractText(result)
	if err := assertion.Check(sb.Expect, resultText, result.IsError); err != nil {
		detail := err.Error()
		if result.IsError && resultText != "" {
			detail += "\n      server response: " + resultText
		}
		return assertion.Result{
			Name:     a.Name,
			Status:   assertion.StatusFail,
			Detail:   detail,
			Duration: assertion.DurationMS(time.Since(start)),
		}
	}

	return assertion.Result{
		Name:     a.Name,
		Status:   assertion.StatusPass,
		Duration: assertion.DurationMS(time.Since(start)),
	}
}
