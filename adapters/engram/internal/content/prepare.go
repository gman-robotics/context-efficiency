package content

import (
	"fmt"
	"regexp"
	"strings"
)

const maxStoreBytes = 50000

var privateTag = regexp.MustCompile(`(?i)<private>.*?</private>`)

// StripPrivateTags replaces private blocks like Engram prepareStoredContent.
func StripPrivateTags(s string) string {
	return privateTag.ReplaceAllString(s, "[REDACTED]")
}

// TruncateAtMax applies Engram's 50000-byte truncation marker.
func TruncateAtMax(s string) string {
	b := []byte(s)
	if len(b) <= maxStoreBytes {
		return s
	}
	cut := b[:maxStoreBytes]
	return string(cut) + "... [truncated]"
}

// PrepareForStore returns content that would be stored, or error if mutation is forbidden.
func PrepareForStore(raw string) (string, error) {
	afterPrivate := StripPrivateTags(raw)
	afterTrunc := TruncateAtMax(afterPrivate)
	if afterPrivate != raw || afterTrunc != afterPrivate {
		return "", fmt.Errorf("content would be redacted or truncated by storage rules")
	}
	return raw, nil
}

// PrepareTitle applies the same rules to titles (handoff uses literal title "handoff").
func PrepareTitle(raw string) (string, error) {
	afterPrivate := StripPrivateTags(raw)
	afterTrunc := TruncateAtMax(afterPrivate)
	if afterPrivate != raw || afterTrunc != afterPrivate {
		return "", fmt.Errorf("title would be redacted or truncated by storage rules")
	}
	return raw, nil
}

// RetainTitle is token plus a short clip of text.
func RetainTitle(token, text string) string {
	clip := strings.TrimSpace(text)
	if len(clip) > 40 {
		clip = clip[:40]
	}
	return token + " " + clip
}

// HandoffTitle is always the literal required by GMA-429 review.
const HandoffTitle = "handoff"
