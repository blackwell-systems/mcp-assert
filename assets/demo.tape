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

# Setup
Hide
Type "rm -rf /tmp/excel-demo && export EXCEL_FILES_PATH=/tmp && mkdir -p /tmp/excel-fixtures"
Enter
Sleep 500ms
Type "cd /Users/dayna.blackwell/code/mcp-assert"
Enter
Sleep 500ms
Show

Sleep 500ms

# Generate
Type "mcp-assert generate --server 'excel-mcp-server stdio' --output /tmp/excel-demo/"
Sleep 200ms
Enter
Sleep 3.5s

Sleep 500ms

# Run
Type "mcp-assert run --suite examples/excel-mcp/ --fixture /tmp/excel-fixtures 2>/dev/null"
Sleep 200ms
Enter
Sleep 7s

Sleep 2s
