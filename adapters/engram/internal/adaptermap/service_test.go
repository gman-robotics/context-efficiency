package adaptermap

import (
	"context"
	"strings"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/child"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/content"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/scope"
)

func TestHandoffWriteTitleLiteral(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	_, err := svc.HandoffWrite(context.Background(), domain.Handoff{Goal: "g"}, "bot")
	if err != nil {
		t.Fatal(err)
	}
	if fake.LastSaveTitle() != content.HandoffTitle {
		t.Fatalf("title %q want handoff", fake.LastSaveTitle())
	}
}

func TestRetainRejectsPrivateBeforeSave(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	_, err := svc.Retain(context.Background(), `hello <private>x</private>`, nil, 1, "bot")
	if err == nil {
		t.Fatal("expected error")
	}
	items, _ := svc.Recall(context.Background(), "hello", "bot", 5)
	if len(items) != 0 {
		t.Fatal("recall should be empty")
	}
}

func TestDefaultScopeIsBot(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	tokenBot, _ := scope.Token("bot")
	tokenShared, _ := scope.Token("shared")
	_, _ = svc.Retain(context.Background(), "shared-only", nil, 1, "shared")
	_, err := svc.Retain(context.Background(), "bot-only", nil, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	items, err := svc.Recall(context.Background(), "bot-only", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.Text == "shared-only" {
			t.Fatal("scopeless recall must not return shared row")
		}
	}
	if len(items) < 1 || items[0].Text != "bot-only" {
		t.Fatalf("expected bot-only item, got %+v", items)
	}
	_ = tokenBot
	_ = tokenShared
}

func TestRecallSkipsHandoffRows(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	_, _ = svc.HandoffWrite(context.Background(), domain.Handoff{Goal: "g"}, "bot")
	_, _ = svc.Retain(context.Background(), "visible fact", nil, 1, "bot")
	items, err := svc.Recall(context.Background(), "visible", "bot", 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if strings.Contains(it.Text, `"goal"`) {
			t.Fatal("handoff row leaked into recall")
		}
	}
}

func TestUnknownProjectEmptyRecall(t *testing.T) {
	fake := &child.FakeEngram{UnknownProject: true}
	svc := &Service{Client: fake}
	items, err := svc.Recall(context.Background(), "q", "bot", 5)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("want empty slice, got %v", items)
	}
	h, err := svc.HandoffRead(context.Background(), "bot")
	if err != nil || h != nil {
		t.Fatalf("handoff_read want nil, got %v err %v", h, err)
	}
}

func TestScopeTokenFixedLength(t *testing.T) {
	tok, err := scope.Token("bot")
	if err != nil || len(tok) != 40 {
		t.Fatalf("token len %d", len(tok))
	}
}

func TestRetainJSONRoundTrip(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	id, err := svc.Retain(context.Background(), "alpha", []string{"t"}, 2, "bot")
	if err != nil || id == "" {
		t.Fatal(err)
	}
	items, err := svc.Recall(context.Background(), "alpha", "bot", 5)
	if err != nil || len(items) != 1 {
		t.Fatalf("recall %+v err %v", items, err)
	}
	if items[0].Text != "alpha" || items[0].CreatedAt == "" {
		t.Fatalf("item %+v", items[0])
	}
}

func TestHandoffBodyMutationError(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	_, err := svc.HandoffWrite(context.Background(), domain.Handoff{
		Goal:      "ship",
		OpenItems: []string{"<private>secret</private>"},
	}, "bot")
	if err == nil {
		t.Fatal("expected handoff private error")
	}
}

// expose saveRec for tests in same package - fake uses unexported type; use LastSaveTitle only

func TestSearchNoResultsArray(t *testing.T) {
	fake := &child.FakeEngram{}
	svc := &Service{Client: fake}
	// After first retain, empty JSON search still ok
	_, _ = svc.Retain(context.Background(), "seed", nil, 1, "bot")
	items, err := svc.Recall(context.Background(), "seed", "bot", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 1 {
		t.Fatal("expected seed")
	}
}
