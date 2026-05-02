package runner

import (
	"flag"
	"fmt"
	"sort"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

// Coverage queries the MCP server's tools/list and reports which server tools
// have direct tool assertions in the suite. Non-tool assertion types (resources,
// prompts, logging) are not counted; this is specifically tool coverage.
func Coverage(args []string) error {
	fs := flag.NewFlagSet("coverage", flag.ExitOnError)
	suiteDir := fs.String("suite", "", "Directory containing assertion YAML files")
	serverSpec := fs.String("server", "", "Server command (e.g. 'agent-lsp go:gopls')")
	timeout := fs.Duration("timeout", 15*time.Second, "Timeout for tools/list call")
	coverageJSON := fs.String("coverage-json", "", "Write coverage data as JSON to path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *suiteDir == "" || *serverSpec == "" {
		return fmt.Errorf("--suite and --server are required")
	}

	// Load assertion suite to find which tools are tested.
	suite, err := assertion.LoadSuite(*suiteDir)
	if err != nil {
		return err
	}

	// Count which tools have direct assertions (both standard and sampling).
	// Setup tools are intentionally excluded: they are test infrastructure,
	// not coverage targets.
	testedTools := make(map[string]int) // tool name -> assertion count
	for _, a := range suite.Assertions {
		if a.Assert.Tool != "" {
			testedTools[a.Assert.Tool]++
		}
		if a.AssertSampling != nil && a.AssertSampling.Tool != "" {
			testedTools[a.AssertSampling.Tool]++
		}
	}

	// Start the server and query tools/list to get the full set of
	// server-advertised tools for the denominator of the coverage ratio.
	serverConfig, err := parseServerSpec(*serverSpec)
	if err != nil {
		return err
	}

	ctx, cancel, mcpClient, err := initializedClientFromConfig(serverConfig, "", *timeout, "")
	if err != nil {
		return err
	}
	defer cancel()
	defer mcpClient.Close()

	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	// Build the coverage report.
	var serverTools []string
	for _, t := range toolsResult.Tools {
		serverTools = append(serverTools, t.Name)
	}
	sort.Strings(serverTools)

	var covered, uncovered []string
	for _, tool := range serverTools {
		if testedTools[tool] > 0 {
			covered = append(covered, tool)
		} else {
			uncovered = append(uncovered, tool)
		}
	}

	total := len(serverTools)
	covCount := len(covered)
	pct := 0
	if total > 0 {
		pct = (covCount * 100) / total
	}

	// Print report.
	useColor := report.ColorEnabled()

	fmt.Printf("\nServer exposes %d tools, %d have assertions (%d%% coverage)\n\n", total, covCount, pct)

	if len(covered) > 0 {
		label := "Covered"
		if useColor {
			label = "\033[32m" + label + "\033[0m"
		}
		fmt.Printf("%s (%d):\n", label, len(covered))
		for _, t := range covered {
			count := testedTools[t]
			icon := "✓"
			if !useColor {
				icon = "+"
			} else {
				icon = "\033[32m✓\033[0m"
			}
			fmt.Printf("  %s %s (%d assertion%s)\n", icon, t, count, plural(count))
		}
	}

	if len(uncovered) > 0 {
		fmt.Println()
		label := "Not covered"
		if useColor {
			label = "\033[31m" + label + "\033[0m"
		}
		fmt.Printf("%s (%d):\n", label, len(uncovered))
		for _, t := range uncovered {
			icon := "○"
			if !useColor {
				icon = "-"
			} else {
				icon = "\033[31m○\033[0m"
			}
			fmt.Printf("  %s %s\n", icon, t)
		}
	}

	fmt.Println()

	if *coverageJSON != "" {
		if err := report.WriteToolCoverageJSON(serverTools, testedTools, *coverageJSON); err != nil {
			return fmt.Errorf("writing coverage JSON: %w", err)
		}
	}

	return nil
}

// plural returns "s" for counts other than 1 (for English pluralization).
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
