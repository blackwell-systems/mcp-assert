package assertion

// LoggingAssertBlock tests MCP logging (logging/setLevel + notifications/message).
type LoggingAssertBlock struct {
	SetLevel string         `yaml:"set_level"`
	Tool     string         `yaml:"tool,omitempty"`
	Args     map[string]any `yaml:"args,omitempty"`
	Expect   LoggingExpect  `yaml:"expect"`
}

// LoggingExpect defines assertions on captured log messages.
type LoggingExpect struct {
	MinMessages   *int     `yaml:"min_messages,omitempty"`
	ContainsLevel []string `yaml:"contains_level,omitempty"`
	ContainsData  []string `yaml:"contains_data,omitempty"`
}

// LogMessage represents a captured notifications/message event.
type LogMessage struct {
	Level  string
	Logger string
	Data   string
}
