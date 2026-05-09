package runner

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestParseServerSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantCmd string
		wantN   int // number of args
		wantErr bool
	}{
		{
			name:    "single command",
			spec:    "my-server",
			wantCmd: "my-server",
			wantN:   0,
		},
		{
			name:    "command with args",
			spec:    "agent-lsp go:gopls",
			wantCmd: "agent-lsp",
			wantN:   1,
		},
		{
			name:    "multiple args",
			spec:    "npx @modelcontextprotocol/server-filesystem /tmp",
			wantCmd: "npx",
			wantN:   2,
		},
		{
			name:    "empty string",
			spec:    "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			spec:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseServerSpec(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Command != tt.wantCmd {
				t.Errorf("command = %q, want %q", cfg.Command, tt.wantCmd)
			}
			if len(cfg.Args) != tt.wantN {
				t.Errorf("args count = %d, want %d", len(cfg.Args), tt.wantN)
			}
		})
	}
}

func TestExtractText_SingleContent(t *testing.T) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Text: "hello world"},
		},
	}
	got := extractText(result)
	if got != "hello world" {
		t.Errorf("extractText = %q, want %q", got, "hello world")
	}
}

func TestExtractText_MultipleContent(t *testing.T) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Text: "line one"},
			mcp.TextContent{Text: "line two"},
		},
	}
	got := extractText(result)
	if got != "line one\nline two" {
		t.Errorf("extractText = %q, want %q", got, "line one\nline two")
	}
}

func TestExtractText_EmptyContent(t *testing.T) {
	result := &mcp.CallToolResult{}
	got := extractText(result)
	if got != "" {
		t.Errorf("extractText empty = %q, want empty", got)
	}
}

func TestEffectiveTransport(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "stdio"},
		{"sse", "sse"},
		{"http", "http"},
		{"stdio", "stdio"},
	}

	for _, tt := range tests {
		got := effectiveTransport(tt.input)
		if got != tt.want {
			t.Errorf("effectiveTransport(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
