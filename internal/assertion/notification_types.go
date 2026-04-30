package assertion

// NotificationAssertBlock tests arbitrary MCP server notifications.
// It calls a tool (which triggers the server to emit notifications),
// captures all notifications during the call, and asserts on them.
//
// Example YAML:
//
//	assert_notifications:
//	  tool: long_running_task
//	  args:
//	    input: "test"
//	  expect:
//	    min_count: 3
//	    methods: ["notifications/progress"]
//	    contains_data: ["processing"]
type NotificationAssertBlock struct {
	Tool   string         `yaml:"tool"`
	Args   map[string]any `yaml:"args,omitempty"`
	Expect NotificationExpect `yaml:"expect"`
}

// NotificationExpect defines assertions on captured notifications.
type NotificationExpect struct {
	// MinCount requires at least N notifications to arrive during the tool call.
	MinCount *int `yaml:"min_count,omitempty"`
	// MaxCount requires at most N notifications.
	MaxCount *int `yaml:"max_count,omitempty"`
	// Methods requires all listed notification methods to appear at least once.
	Methods []string `yaml:"methods,omitempty"`
	// NotMethods requires none of the listed methods to appear.
	NotMethods []string `yaml:"not_methods,omitempty"`
	// ContainsData requires at least one notification's params to contain each substring.
	ContainsData []string `yaml:"contains_data,omitempty"`
	// NotContainsData requires no notification's params to contain any of these substrings.
	NotContainsData []string `yaml:"not_contains_data,omitempty"`
}

// CapturedNotification represents a notification received during a tool call.
type CapturedNotification struct {
	Method string
	Params string // JSON-serialized params
}
