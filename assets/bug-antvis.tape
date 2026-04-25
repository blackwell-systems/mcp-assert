# Bug demo: antvis/mcp-server-chart (2K stars) — tools crash with stack traces
# Real command, real output

Output assets/bug-antvis.gif
Set Shell "bash"
Set FontSize 14
Set FontFamily "Menlo"
Set Width 1060
Set Height 580
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 30ms
Set WindowBar "Colorful"
Set WindowBarSize 40
Set CursorBlink false
Set Margin 10
Set MarginFill "#1e1e2e"
Set BorderRadius 12

Hide
Type `export PS1="$ "`
Enter
Sleep 300ms
Type "cd /Users/dayna.blackwell/code/mcp-assert"
Enter
Sleep 300ms
Show

Sleep 500ms

Type "mcp-assert run --suite /tmp/antvis-demo-short/ --timeout 15s 2>/dev/null"
Sleep 200ms
Enter
Sleep 14s

Sleep 3s
