package content

import "testing"

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
