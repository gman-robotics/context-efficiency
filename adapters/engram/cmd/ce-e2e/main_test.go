package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/decide"
)

// fakeCaller serves decide from the real rules and keeps memory in a map.
type fakeCaller struct {
	facts   []string
	handoff map[string]any
	broken  string
}

func (f *fakeCaller) Call(_ context.Context, name string, args map[string]any) (string, bool, error) {
	if name == f.broken {
		return "boom", true, nil
	}
	switch name {
	case "retain":
		f.facts = append(f.facts, args["text"].(string))
		return "id-1", false, nil
	case "recall":
		b, _ := json.Marshal(f.facts)
		return string(b), false, nil
	case "handoff_write":
		f.handoff = args["handoff"].(map[string]any)
		return "ok", false, nil
	case "handoff_read":
		b, _ := json.Marshal(f.handoff)
		return string(b), false, nil
	}
	raw, _ := json.Marshal(args["features"])
	r, err := decide.Decide(args["path"].(string), raw)
	if err != nil {
		return err.Error(), true, nil
	}
	b, _ := json.Marshal(r)
	return string(b), false, nil
}

func runAll(c Caller) *Report {
	rep := &Report{Commit: "test"}
	for _, s := range Steps() {
		rep.Add(s.Name, s.Run(context.Background(), c))
	}
	return rep
}

func TestStepsPassAgainstRealRules(t *testing.T) {
	rep := runAll(&fakeCaller{})
	if rep.Failed() != 0 {
		t.Fatalf("%+v", rep.Results)
	}
	if len(rep.Results) != 15 {
		t.Fatalf("steps %d", len(rep.Results))
	}
}

func TestEachToolFailureIsReported(t *testing.T) {
	for _, tool := range []string{"retain", "recall", "handoff_write", "handoff_read", "decide"} {
		rep := runAll(&fakeCaller{broken: tool})
		if rep.Failed() == 0 {
			t.Fatalf("%s failure not reported", tool)
		}
	}
}

func TestHandoffFieldDropIsReported(t *testing.T) {
	f := &fakeCaller{handoff: map[string]any{"goal": handoff["goal"]}}
	if stepHandoffRead(context.Background(), f) == "" {
		t.Fatal("dropped handoff fields passed")
	}
}

func TestRefusedCaseFailsWhenAllowed(t *testing.T) {
	d := decideCase{"x", "routine_go_nogo", map[string]any{"name_ok": true, "action": "write_file"}, ""}
	if d.run(context.Background(), &fakeCaller{}) == "" {
		t.Fatal("allowed action passed a refused case")
	}
	d.label = "no_go"
	if !strings.Contains(d.run(context.Background(), &fakeCaller{}), "want \"no_go\"") {
		t.Fatal("label mismatch not reported")
	}
}

func TestCheckTools(t *testing.T) {
	if err := checkTools([]string{"retain", "recall", "handoff_write", "handoff_read", "decide"}); err != nil {
		t.Fatal(err)
	}
	if checkTools(append([]string{"decide_shadow"}, WantTools...)) == nil {
		t.Fatal("decide_shadow accepted")
	}
}

func TestReportWrite(t *testing.T) {
	dir := t.TempDir()
	rep := &Report{Commit: "abc", Server: "s 1", Tools: []string{"recall"}}
	rep.Add("a", "")
	rep.Add("b", "x|y")
	if err := rep.Write(dir); err != nil {
		t.Fatal(err)
	}
	md, _ := os.ReadFile(filepath.Join(dir, "local-e2e.md"))
	if !strings.Contains(string(md), "1 of 2 steps pass") || !strings.Contains(string(md), "x\\|y") {
		t.Fatalf("%s", md)
	}
	if _, err := os.Stat(filepath.Join(dir, "local-e2e.json")); err != nil {
		t.Fatal(err)
	}
}
