package content

import (
	"strings"
	"testing"
)

func TestPrepareForStoreRejectsPrivate(t *testing.T) {
	_, err := PrepareForStore(`hello <private>x</private>`)
	if err == nil {
		t.Fatal("expected error for private tag")
	}
}

func TestHandoffTitleLiteral(t *testing.T) {
	if HandoffTitle != "handoff" {
		t.Fatalf("handoff title must be literal handoff")
	}
}

func TestPrepareForStoreRejectsOverMaxBytes(t *testing.T) {
	raw := strings.Repeat("a", 50001)
	_, err := PrepareForStore(raw)
	if err == nil || err.Error() != "content would be redacted or truncated by storage rules" {
		t.Fatalf("err %v", err)
	}
}
