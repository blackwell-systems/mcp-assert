// Package runner
// client.go handles MCP client creation for all three transport modes.
//
// Transport selection:
//   - "stdio" (default): launches the server as a child process. Most common
//     for local development and CI. Supports Docker isolation (wraps the
//     command in "docker run --rm -i").
//   - "sse": connects to a running server via Server-Sent Events at a URL.
//   - "http": connects via Streamable HTTP (the newer MCP transport).
//
// For stdio clients that need mock capabilities (roots, sampling, elicitation),
// createStdioClientWithCapabilities builds a client with static handlers that
// respond to server-initiated requests with pre-configured values from the
// YAML assertion file.
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/mark3labs/mcp-go/client"
	clienttransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// stdioServerProcess holds the resolved command, args, and env for launching
// an MCP server subprocess. Produced by buildStdioServerProcess, which handles
// fixture substitution, env var expansion, and Docker wrapping.
type stdioServerProcess struct {
	command string
	args    []string
	env     []string
}

// buildStdioServerProcess resolves the server config into a concrete command
// to execute. Handles fixture path substitution, environment variable expansion,
// and Docker container wrapping (if dockerImage is set or server.Docker is set).
func buildStdioServerProcess(server assertion.ServerConfig, fixture, dockerImage string) stdioServerProcess {
	serverCmd := server.Command

	serverArgs := make([]string, len(server.Args))
	copy(serverArgs, server.Args)

	if fixture != "" {
		for i, arg := range serverArgs {
			serverArgs[i] = strings.ReplaceAll(arg, "{{fixture}}", fixture)
		}
	}

	// Sort env var keys for deterministic behavior. Map iteration order is
	// random in Go; sorting ensures consistent Docker -e flag ordering and
	// reproducible process environments, which matters for a testing tool.
	keys := make([]string, 0, len(server.Env))
	for k := range server.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	envSlice := make([]string, 0, len(keys))
	for _, k := range keys {
		envSlice = append(envSlice, k+"="+expandEnvVars(server.Env[k]))
	}

	// Docker isolation: wrap the server command in "docker run --rm -i".
	// The per-assertion docker field takes precedence over the CLI --docker flag.
	effectiveDocker := server.Docker
	if effectiveDocker == "" {
		effectiveDocker = dockerImage
	}

	if effectiveDocker == "" {
		return stdioServerProcess{
			command: serverCmd,
			args:    serverArgs,
			env:     envSlice,
		}
	}

	dockerArgs := []string{"run", "--rm", "-i"}
	if fixture != "" {
		absFixture, err := filepath.Abs(fixture)
		if err == nil {
			fixture = absFixture
		}
		dockerArgs = append(dockerArgs, "-v", fixture+":"+fixture)
	}
	for _, e := range envSlice {
		dockerArgs = append(dockerArgs, "-e", e)
	}
	dockerArgs = append(dockerArgs, effectiveDocker, serverCmd)
	dockerArgs = append(dockerArgs, serverArgs...)

	return stdioServerProcess{
		command: "docker",
		args:    dockerArgs,
		env:     nil, // env is passed via -e flags
	}
}

// hasClientCapabilities reports whether any mock client capabilities are configured.
func hasClientCapabilities(caps assertion.ClientCapabilities) bool {
	return len(caps.Roots) > 0 || caps.Sampling != nil || len(caps.Elicitation) > 0
}

// expandEnvVars resolves ${VAR} and $VAR patterns in a string
// from the parent process environment. Unset variables are replaced
// with empty string, matching shell behavior.
func expandEnvVars(value string) string {
	return os.ExpandEnv(value)
}

// expandHeaderVars resolves ${VAR} patterns in header values from the environment.
func expandHeaderVars(headers map[string]string) map[string]string {
	expanded := make(map[string]string, len(headers))
	for k, v := range headers {
		expanded[k] = expandEnvVars(v)
	}
	return expanded
}

// createMCPClient creates the appropriate MCP client based on the server config's
// transport type. For stdio (default), it launches a subprocess. For sse/http, it
// connects to the specified URL. Docker isolation is only supported with stdio.
func createMCPClient(server assertion.ServerConfig, fixture string, dockerImage string) (client.MCPClient, error) {
	transport := strings.ToLower(server.Transport)

	switch transport {
	case "sse":
		if server.URL == "" {
			return nil, fmt.Errorf("transport %q requires a url field", transport)
		}
		var sseOpts []clienttransport.ClientOption
		if len(server.Headers) > 0 {
			expanded := expandHeaderVars(server.Headers)
			sseOpts = append(sseOpts, clienttransport.WithHeaders(expanded))
		}
		sseClient, err := client.NewSSEMCPClient(server.URL, sseOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSE client: %w", err)
		}
		if err := sseClient.Start(context.Background()); err != nil {
			_ = sseClient.Close()
			return nil, fmt.Errorf("failed to start SSE transport: %w", err)
		}
		return sseClient, nil
	case "http":
		if server.URL == "" {
			return nil, fmt.Errorf("transport %q requires a url field", transport)
		}
		var httpOpts []clienttransport.StreamableHTTPCOption
		if len(server.Headers) > 0 {
			expanded := expandHeaderVars(server.Headers)
			httpOpts = append(httpOpts, clienttransport.WithHTTPHeaders(expanded))
		}
		return client.NewStreamableHttpClient(server.URL, httpOpts...)
	case "stdio", "":
		process := buildStdioServerProcess(server, fixture, dockerImage)

		caps := server.ClientCapabilities
		if hasClientCapabilities(caps) {
			return createStdioClientWithCapabilities(
				process.command, process.env, process.args, fixture, caps,
			)
		}

		return client.NewStdioMCPClient(process.command, process.env, process.args...)
	default:
		return nil, fmt.Errorf("unknown transport %q (expected stdio, sse, or http)", transport)
	}
}

// createStdioClientWithCapabilities creates a stdio client with mock client capabilities.
func createStdioClientWithCapabilities(
	command string,
	env []string,
	args []string,
	fixture string,
	caps assertion.ClientCapabilities,
) (client.MCPClient, error) {
	stdioTransport := clienttransport.NewStdioWithOptions(command, env, args)
	if err := stdioTransport.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start stdio transport: %w", err)
	}

	var opts []client.ClientOption

	// Roots: respond to roots/list with configured paths.
	if len(caps.Roots) > 0 {
		roots := make([]mcp.Root, 0, len(caps.Roots))
		for _, path := range caps.Roots {
			path = strings.ReplaceAll(path, "{{fixture}}", fixture)
			roots = append(roots, mcp.Root{
				URI:  "file://" + path,
				Name: filepath.Base(path),
			})
		}
		opts = append(opts, client.WithRootsHandler(&staticRootsHandler{roots: roots}))
	}

	// Sampling: respond to sampling/createMessage with a mock LLM response.
	if caps.Sampling != nil {
		text := caps.Sampling.Text
		model := caps.Sampling.Model
		if model == "" {
			model = "mock"
		}
		stopReason := caps.Sampling.StopReason
		if stopReason == "" {
			stopReason = "end_turn"
		}
		opts = append(opts, client.WithSamplingHandler(&staticSamplingHandler{
			text:       text,
			model:      model,
			stopReason: stopReason,
		}))
	}

	// Elicitation: respond to server-initiated user prompts with preset values.
	// Servers use elicitation to ask the user for input; we return canned answers.
	if len(caps.Elicitation) > 0 {
		opts = append(opts, client.WithElicitationHandler(&staticElicitationHandler{values: caps.Elicitation}))
	}

	c := client.NewClient(stdioTransport, opts...)
	// Start the client to register bidirectional request handlers.
	// transport.Start() is idempotent (guarded by c.started mutex),
	// so calling it again via c.Start() is safe.
	if err := c.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}
	return c, nil
}

// staticRootsHandler returns a fixed list of roots.
type staticRootsHandler struct {
	roots []mcp.Root
}

func (h *staticRootsHandler) ListRoots(_ context.Context, _ mcp.ListRootsRequest) (*mcp.ListRootsResult, error) {
	return &mcp.ListRootsResult{Roots: h.roots}, nil
}

// staticSamplingHandler returns a fixed mock LLM response.
type staticSamplingHandler struct {
	text       string
	model      string
	stopReason string
}

func (h *staticSamplingHandler) CreateMessage(_ context.Context, _ mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	return &mcp.CreateMessageResult{
		SamplingMessage: mcp.SamplingMessage{
			Role:    mcp.RoleAssistant,
			Content: mcp.TextContent{Type: "text", Text: h.text},
		},
		Model:      h.model,
		StopReason: h.stopReason,
	}, nil
}

// staticElicitationHandler returns preset values for server-initiated prompts.
type staticElicitationHandler struct {
	values map[string]any
}

func (h *staticElicitationHandler) Elicit(_ context.Context, _ mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	// If the preset values include an "action" key, use it to control whether
	// the response is accept/decline/cancel. Otherwise default to accept.
	action := mcp.ElicitationResponseActionAccept
	if a, ok := h.values["action"]; ok {
		if actionStr, ok := a.(string); ok {
			switch actionStr {
			case "decline":
				action = mcp.ElicitationResponseActionDecline
			case "cancel":
				action = mcp.ElicitationResponseActionCancel
			}
		}
	}

	var content any
	if c, ok := h.values["content"]; ok {
		// Explicit "content" key overrides the entire response body.
		content = c
	} else {
		// No explicit content: use all non-"action" keys as the response body.
		filtered := make(map[string]any)
		for k, v := range h.values {
			if k != "action" {
				filtered[k] = v
			}
		}
		content = filtered
	}
	return &mcp.ElicitationResult{
		ElicitationResponse: mcp.ElicitationResponse{
			Action:  action,
			Content: content,
		},
	}, nil
}
