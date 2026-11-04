package shadow

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

func TestRulePrecedence(t *testing.T) {
	got, err := RulePR(PRFeatures{DocsOnly: true, Infra: true})
	if err != nil || got != "T3" {
		t.Fatalf("docs+infra %s %v", got, err)
	}
	got, err = RulePR(PRFeatures{ExistingModule: true, NewModule: true})
	if err != nil || got != "T2" {
		t.Fatalf("existing+new %s %v", got, err)
	}
	got, err = RulePR(PRFeatures{NewService: true, NewModule: true})
	if err != nil || got != "T3" {
		t.Fatalf("service+module %s %v", got, err)
	}
	if _, err := RulePR(PRFeatures{}); err == nil {
		t.Fatal("unmatched pr")
	}
	got, err = RuleStop(StopFeatures{SameTool: true, SameErrorClass: true, Count: 2})
	if err != nil || got != "continue" {
		t.Fatalf("count 2 %s %v", got, err)
	}
	got, err = RuleStop(StopFeatures{SameTool: false, SameErrorClass: true, Count: 3})
	if err != nil || got != "continue" {
		t.Fatalf("one flag %s %v", got, err)
	}
	got, err = RuleStop(StopFeatures{SameTool: true, SameErrorClass: true, Count: 3})
	if err != nil || got != "stop" {
		t.Fatalf("stop %s %v", got, err)
	}
	if _, err := RuleStop(StopFeatures{Count: -1}); err == nil {
		t.Fatal("negative count")
	}
	got, err = RuleAudit(AuditFeatures{Kind: "security"})
	if err != nil || got != "blocker" {
		t.Fatal(got, err)
	}
	got, err = RuleAudit(AuditFeatures{Kind: "behavior_bug"})
	if err != nil || got != "should_fix" {
		t.Fatal(got, err)
	}
	if _, err := RuleAudit(AuditFeatures{Kind: "other"}); err == nil {
		t.Fatal("unmatched audit")
	}
	if _, err := RuleGo(GoFeatures{NameOK: true, Action: "send", Blocked: false}); err == nil {
		t.Fatal("forbidden action reached a label")
	}
	got, err = RuleGo(GoFeatures{NameOK: false, Action: "write_file", Blocked: false})
	if err != nil || got != "no_go" {
		t.Fatal(got, err)
	}
	got, err = RuleGo(GoFeatures{NameOK: true, Action: "read_only", Blocked: false})
	if err != nil || got != "no_go" {
		t.Fatal(got, err)
	}
	got, err = RuleGo(GoFeatures{NameOK: true, Action: "retain_summary", Blocked: false})
	if err != nil || got != "go" {
		t.Fatal(got, err)
	}
}

func TestGoFeaturesHasNoFailureCount(t *testing.T) {
	typ := reflect.TypeOf(GoFeatures{})
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		tag := strings.ToLower(typ.Field(i).Tag.Get("json"))
		if strings.Contains(name, "fail") || strings.Contains(name, "count") || strings.Contains(tag, "fail") || strings.Contains(tag, "count") {
			t.Fatalf("field %s", typ.Field(i).Name)
		}
	}
	text := RenderGo(GoFeatures{NameOK: true, Action: "write_file", Blocked: false})
	if strings.Contains(text, "identical_failure") || strings.Contains(text, "failure") {
		t.Fatal(text)
	}
}

func TestCriteriaAndInstructions(t *testing.T) {
	if DeniedToken(laya.Instructions) != "" {
		t.Fatal(laya.Instructions)
	}
	for _, path := range []string{PathGo, PathPR, PathStop, PathAudit} {
		m, err := Criteria(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(m) < 2 || len(m) > 4 {
			t.Fatalf("%s len %d", path, len(m))
		}
		for key, gloss := range m {
			if !verdictOK(path, key) {
				t.Fatalf("%s key %s", path, key)
			}
			if DeniedToken(key) != "" || DeniedToken(gloss) != "" {
				t.Fatalf("%s %s %s", path, key, gloss)
			}
		}
	}
	pr, err := Criteria(PathPR)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pr["T0"], "typos") {
		t.Fatal(pr["T0"])
	}
	if !strings.Contains(pr["T3"], "major architectural move") {
		t.Fatal(pr["T3"])
	}
	rendered := RenderPR(PRFeatures{DocsOnly: true})
	if !strings.Contains(rendered, "typos") || !strings.Contains(rendered, "major architectural move") {
		t.Fatal(rendered)
	}
	if _, err := Criteria("nope"); err == nil || err.Error() != "refused: path" {
		t.Fatal(err)
	}
}
