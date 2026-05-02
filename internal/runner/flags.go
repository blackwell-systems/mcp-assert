package runner

import (
	"flag"
	"strings"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// serverFlags holds the common flags shared across commands that connect to
// an MCP server (audit, fuzz, generate, coverage). Registering these in one
// place prevents flag name/default/description drift between commands.
type serverFlags struct {
	server    string
	transport string
	headers   string
	timeout   time.Duration
	jsonOut   bool
}

// register adds the common server flags to a FlagSet.
func (f *serverFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&f.server, "server", "", "Server command (stdio) or URL (http/sse)")
	fs.StringVar(&f.transport, "transport", "stdio", "Transport type: stdio (default), http, sse")
	fs.StringVar(&f.headers, "headers", "", "Custom headers as key=value pairs, comma-separated")
	fs.DurationVar(&f.timeout, "timeout", 15*time.Second, "Per-call timeout")
	fs.BoolVar(&f.jsonOut, "json", false, "Output results as JSON")
}

// serverConfig builds an assertion.ServerConfig from the parsed flags.
func (f *serverFlags) serverConfig() (assertion.ServerConfig, error) {
	transportLower := strings.ToLower(f.transport)
	headers := parseHeadersFlag(f.headers)

	if transportLower == "http" || transportLower == "sse" {
		return assertion.ServerConfig{
			Transport: transportLower,
			URL:       f.server,
			Headers:   headers,
		}, nil
	}

	parts := strings.Fields(f.server)
	if len(parts) == 0 {
		return assertion.ServerConfig{}, nil
	}
	return assertion.ServerConfig{
		Command: parts[0],
		Args:    parts[1:],
	}, nil
}
