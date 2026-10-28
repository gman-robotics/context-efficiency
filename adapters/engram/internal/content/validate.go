package content

import (
	"fmt"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
)

// RejectPrivateInString errors when a field would be redacted.
func RejectPrivateInString(s string) error {
	if StripPrivateTags(s) != s {
		return fmt.Errorf("content would be redacted by storage rules")
	}
	return nil
}

// RejectPrivateInHandoff scans handoff string fields before JSON marshal.
func RejectPrivateInHandoff(h domain.Handoff) error {
	if err := RejectPrivateInString(h.Goal); err != nil {
		return err
	}
	for _, s := range h.OpenItems {
		if err := RejectPrivateInString(s); err != nil {
			return err
		}
	}
	for _, s := range h.Decisions {
		if err := RejectPrivateInString(s); err != nil {
			return err
		}
	}
	for _, s := range h.FilePaths {
		if err := RejectPrivateInString(s); err != nil {
			return err
		}
	}
	for _, s := range h.NextSteps {
		if err := RejectPrivateInString(s); err != nil {
			return err
		}
	}
	return nil
}

// RejectOversizeBytes rejects content that would exceed Engram storage limits.
func RejectOversizeBytes(b []byte, max int) error {
	if len(b) > max {
		return fmt.Errorf("content too large")
	}
	prepared := TruncateAtMax(string(b))
	if prepared != string(b) {
		return fmt.Errorf("content would be truncated by storage rules")
	}
	return nil
}
