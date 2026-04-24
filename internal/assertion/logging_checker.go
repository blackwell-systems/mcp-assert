package assertion

import (
	"fmt"
	"strings"
)

// CheckLogging evaluates logging expectations against captured log messages.
func CheckLogging(expect LoggingExpect, messages []LogMessage) error {
	if expect.MinMessages != nil && len(messages) < *expect.MinMessages {
		return fmt.Errorf("expected at least %d log message(s), got %d",
			*expect.MinMessages, len(messages))
	}
	for _, level := range expect.ContainsLevel {
		found := false
		for _, msg := range messages {
			if msg.Level == level {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected log message at level %q, not found", level)
		}
	}
	for _, substr := range expect.ContainsData {
		found := false
		for _, msg := range messages {
			if strings.Contains(msg.Data, substr) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected log data containing %q, not found", substr)
		}
	}
	return nil
}
