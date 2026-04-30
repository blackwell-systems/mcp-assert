package assertion

import (
	"fmt"
	"strings"
)

// CheckNotifications evaluates notification expectations against captured notifications.
func CheckNotifications(expect NotificationExpect, notifications []CapturedNotification) error {
	count := len(notifications)

	// Check minimum count.
	if expect.MinCount != nil && count < *expect.MinCount {
		return fmt.Errorf("expected at least %d notification(s), got %d", *expect.MinCount, count)
	}

	// Check maximum count.
	if expect.MaxCount != nil && count > *expect.MaxCount {
		return fmt.Errorf("expected at most %d notification(s), got %d", *expect.MaxCount, count)
	}

	// Check required methods.
	for _, method := range expect.Methods {
		found := false
		for _, n := range notifications {
			if n.Method == method {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected notification method %q but it was not received", method)
		}
	}

	// Check excluded methods.
	for _, method := range expect.NotMethods {
		for _, n := range notifications {
			if n.Method == method {
				return fmt.Errorf("notification method %q was received but should not have been", method)
			}
		}
	}

	// Check that at least one notification's params contains each required substring.
	for _, substr := range expect.ContainsData {
		found := false
		for _, n := range notifications {
			if strings.Contains(n.Params, substr) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no notification params contain %q", substr)
		}
	}

	// Check that no notification's params contains any excluded substring.
	for _, substr := range expect.NotContainsData {
		for _, n := range notifications {
			if strings.Contains(n.Params, substr) {
				return fmt.Errorf("notification params contain %q but should not", substr)
			}
		}
	}

	return nil
}
