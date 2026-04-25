# mcp-assert demo: testing excel-mcp-server (3,750 stars)

Output assets/demo.gif
Set Shell "bash"
Set FontSize 15
Set FontFamily "Menlo"
Set Width 960
Set Height 520
Set Theme "Catppuccin Mocha"
Set Padding 25
Set TypingSpeed 35ms
Set WindowBar "Colorful"
Set WindowBarSize 40
Set CursorBlink false
Set Margin 10
Set MarginFill "#1e1e2e"
Set BorderRadius 12

# Setup: clean slate, suppress progress counter
Hide
Type "rm -rf /tmp/excel-demo && export EXCEL_FILES_PATH=/tmp && export NO_COLOR=1 && mkdir -p /tmp/excel-fixtures"
Enter
Sleep 500ms
Type "cd /Users/dayna.blackwell/code/mcp-assert"
Enter
Sleep 500ms
Show

# Step 1: Auto-generate assertions from a live server
Type "# Discover tools from a running MCP server"
Enter
Sleep 800ms
Type "mcp-assert generate --server 'excel-mcp-server stdio' --output /tmp/excel-demo/"
Sleep 200ms
Enter
Sleep 5s

# Pause to read
Sleep 1.5s

# Step 2: Run the hand-crafted suite (all pass)
Type "# Run assertions against the server"
Enter
Sleep 800ms
Type "mcp-assert run --suite examples/excel-mcp/ --fixture /tmp/excel-fixtures 2>/dev/null"
Sleep 200ms
Enter
Sleep 8s

# Hold on results
Sleep 3s
