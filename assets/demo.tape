# mcp-assert demo: testing excel-mcp-server (3,750 stars)

Output assets/demo.gif
Set Shell "bash"
Set FontSize 13
Set Width 880
Set Height 500
Set Theme "Dracula"
Set Padding 15
Set TypingSpeed 40ms

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
Type "# Step 1: Auto-discover tools from a running MCP server"
Enter
Sleep 1s
Type "mcp-assert generate --server 'excel-mcp-server stdio' --output /tmp/excel-demo/"
Sleep 300ms
Enter
Sleep 5s

# Step 2: Run the hand-crafted suite (all pass)
Type "# Step 2: Run assertions"
Enter
Sleep 1s
Type "mcp-assert run --suite examples/excel-mcp/ --fixture /tmp/excel-fixtures 2>/dev/null"
Sleep 300ms
Enter
Sleep 8s

Sleep 2s
