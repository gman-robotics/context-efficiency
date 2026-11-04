package shadow

import (
	"strings"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/content"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

func TestFixtureCensus(t *testing.T) {
	cases, err := LoadFixtures()
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 68 {
		t.Fatalf("count %d", len(cases))
	}
	by := map[string][]Case{}
	seen := map[string]struct{}{}
	var goCount int
	tiers := map[string]int{}
	kinds := map[string]int{}
	for _, c := range cases {
		if _, ok := seen[c.ID]; ok {
			t.Fatalf("dup %s", c.ID)
		}
		seen[c.ID] = struct{}{}
		by[c.Path] = append(by[c.Path], c)
		got, err := Rule(c)
		if err != nil {
			t.Fatalf("%s %v", c.ID, err)
		}
		if got != c.Actual {
			t.Fatalf("%s rule %s actual %s", c.ID, got, c.Actual)
		}
		if DeniedToken(c.Body) != "" || strings.Contains(c.Body, "@") || digitRun.MatchString(c.Body) {
			t.Fatalf("body %s", c.ID)
		}
		if err := content.RejectPrivateInString(c.Body); err != nil {
			t.Fatal(c.ID, err)
		}
		if strings.Contains(c.Body, "identical_failure") {
			t.Fatal(c.ID)
		}
		switch c.Path {
		case PathGo:
			switch c.Action {
			case "write_file", "retain_summary", "read_only":
			default:
				t.Fatalf("action %s %s", c.ID, c.Action)
			}
			if c.Actual == "go" {
				goCount++
			}
		case PathPR:
			if c.Action != "" {
				t.Fatal(c.ID)
			}
			tiers[c.Actual]++
		case PathStop:
			if c.Action != "" {
				t.Fatal(c.ID)
			}
			checkStop(t, c)
		case PathAudit:
			if c.Action != "" {
				t.Fatal(c.ID)
			}
			kinds[c.Audit.Kind]++
		default:
			t.Fatal(c.Path)
		}
	}
	if goCount < 8 {
		t.Fatalf("go %d", goCount)
	}
	for _, tier := range []string{"T0", "T1", "T2", "T3"} {
		if tiers[tier] < 4 {
			t.Fatalf("%s %d", tier, tiers[tier])
		}
	}
	wantKinds := map[string]int{
		"security": 3, "data_loss": 3, "test_failure": 2,
		"behavior_bug": 3, "naming": 3, "typo_question": 2,
	}
	for kind, n := range wantKinds {
		if kinds[kind] != n {
			t.Fatalf("%s %d", kind, kinds[kind])
		}
	}
	if len(by[PathGo]) != 20 || len(by[PathPR]) != 20 || len(by[PathStop]) != 12 || len(by[PathAudit]) != 16 {
		t.Fatalf("sizes %d %d %d %d", len(by[PathGo]), len(by[PathPR]), len(by[PathStop]), len(by[PathAudit]))
	}
	assertIDs(t, by[PathGo], []string{
		"gng-go-write-01", "gng-go-write-02", "gng-go-write-03", "gng-go-write-04",
		"gng-go-retain-01", "gng-go-retain-02", "gng-go-retain-03", "gng-go-retain-04",
		"gng-nogo-name-01", "gng-nogo-name-02", "gng-nogo-name-03", "gng-nogo-name-04",
		"gng-nogo-read-01", "gng-nogo-read-02", "gng-nogo-read-03", "gng-nogo-read-04",
		"gng-nogo-blocked-01", "gng-nogo-blocked-02", "gng-nogo-blocked-03", "gng-nogo-blocked-04",
	})
	assertIDs(t, by[PathPR], []string{
		"pr-t3-infra", "pr-t3-migration", "pr-t3-new-service", "pr-t3-docs-and-infra", "pr-t3-infra-and-module",
		"pr-t2-new-module", "pr-t2-cross-cutting", "pr-t2-existing-and-new-module", "pr-t2-cross-and-docs", "pr-t2-new-module-b",
		"pr-t1-existing-01", "pr-t1-existing-02", "pr-t1-existing-03", "pr-t1-existing-04", "pr-t1-existing-05",
		"pr-t0-docs", "pr-t0-check", "pr-t0-config", "pr-t0-docs-and-check", "pr-t0-config-b",
	})
	assertIDs(t, by[PathStop], []string{
		"stop-both-count-0", "stop-both-count-1", "stop-both-count-2",
		"stop-both-count-3", "stop-both-count-4", "stop-both-count-5",
		"stop-tool-diff-count-3", "stop-class-diff-count-3", "stop-both-diff-count-4",
		"stop-tool-diff-count-4", "stop-class-diff-count-5", "stop-both-diff-count-5",
	})
	assertIDs(t, by[PathAudit], []string{
		"audit-security-01", "audit-security-02", "audit-security-03",
		"audit-data-loss-01", "audit-data-loss-02", "audit-data-loss-03",
		"audit-test-failure-01", "audit-test-failure-02",
		"audit-behavior-bug-01", "audit-behavior-bug-02", "audit-behavior-bug-03",
		"audit-naming-01", "audit-naming-02", "audit-naming-03",
		"audit-typo-01", "audit-typo-02",
	})
	if DeniedToken(laya.Instructions) != "" {
		t.Fatal("instructions")
	}
	for _, c := range by[PathPR] {
		if !strings.Contains(c.Body, "typos") || !strings.Contains(c.Body, "major architectural move") {
			t.Fatal(c.ID)
		}
	}
	for _, c := range by[PathGo] {
		if strings.Contains(c.Body, "identical_failure") {
			t.Fatal(c.ID)
		}
	}
}

func checkStop(t *testing.T, c Case) {
	t.Helper()
	f := c.Stop
	if f.SameTool && f.Tool != f.OtherTool {
		t.Fatalf("%s tools", c.ID)
	}
	if !f.SameTool && f.Tool == f.OtherTool {
		t.Fatalf("%s tools differ", c.ID)
	}
	if f.SameErrorClass && f.ErrorClass != f.OtherError {
		t.Fatalf("%s class", c.ID)
	}
	if !f.SameErrorClass && f.ErrorClass == f.OtherError {
		t.Fatalf("%s class differ", c.ID)
	}
	for _, tool := range []string{f.Tool, f.OtherTool} {
		if tool != "tool_a" && tool != "tool_b" {
			t.Fatalf("%s tool %s", c.ID, tool)
		}
	}
	for _, class := range []string{f.ErrorClass, f.OtherError} {
		if class != "err_x" && class != "err_y" {
			t.Fatalf("%s class %s", c.ID, class)
		}
	}
}

func assertIDs(t *testing.T, cases []Case, want []string) {
	t.Helper()
	if len(cases) != len(want) {
		t.Fatalf("len %d want %d", len(cases), len(want))
	}
	for i, id := range want {
		if cases[i].ID != id {
			t.Fatalf("id %d %s want %s", i, cases[i].ID, id)
		}
	}
}
