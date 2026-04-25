# mcp-assert demo: testing excel-mcp-server
# Records the zero-effort testing flow

Output assets/demo.gif
Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Dracula"
Set Padding 20

# Title
Type "# Test any MCP server in 60 seconds"
Enter
Sleep 1s

# Step 1: Generate assertions
Type "mcp-assert generate --server 'excel-mcp-server stdio' --output /tmp/excel-demo/"
Sleep 500ms
Enter
Sleep 3s

# Step 2: Run the stubs
Type "mcp-assert run --suite /tmp/excel-demo/"
Sleep 500ms
Enter
Sleep 5s

# Step 3: Coverage
Type "mcp-assert coverage --suite /tmp/excel-demo/ --server 'excel-mcp-server stdio'"
Sleep 500ms
Enter
Sleep 3s

# Done
Type "# 25 tools discovered, 25 assertions generated, 100% coverage"
Enter
Sleep 3s
