package assertion

// SamplingAssertBlock tests MCP sampling as a first-class test subject.
// Combines mock sampling config + tool call + assertions into one block.
// The runner configures client_capabilities.sampling automatically from
// MockText/MockModel, calls the specified tool (which triggers the
// server's sampling request), and asserts on the final tool result.
type SamplingAssertBlock struct {
	Tool      string         `yaml:"tool"`
	Args      map[string]any `yaml:"args,omitempty"`
	MockText  string         `yaml:"mock_text"`
	MockModel string         `yaml:"mock_model,omitempty"`
	Expect    Expect         `yaml:"expect"`
}
