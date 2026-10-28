package content

import (
	"encoding/json"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
)

func TestPrepareHandoffJSONPrivate(t *testing.T) {
	raw, _ := json.Marshal(domain.Handoff{
		Goal:      "ship",
		OpenItems: []string{"<private>secret</private>"},
	})
	if err := RejectPrivateInHandoff(domain.Handoff{
		Goal:      "ship",
		OpenItems: []string{"<private>secret</private>"},
	}); err == nil {
		t.Fatalf("handoff private should fail before marshal, json was %s", raw)
	}
}
