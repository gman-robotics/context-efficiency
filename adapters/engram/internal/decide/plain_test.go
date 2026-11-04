package decide

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
)

func TestDecideGo(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"name_ok":true,"action":"write_file","blocked":false}`, "go"},
		{`{"name_ok":true,"action":"retain_summary"}`, "go"},
		{`{"name_ok":true,"action":"read_only"}`, "no_go"},
		{`{"name_ok":false,"action":"write_file"}`, "no_go"},
		{`{"name_ok":true,"action":"write_file","blocked":true}`, "no_go"},
	}
	for _, c := range cases {
		got, err := Decide(shadow.PathGo, []byte(c.in))
		if err != nil || got.Label != c.want || got.Effect != "none" || got.Rule == "" {
			t.Fatalf("%s: got %+v err %v want %s", c.in, got, err, c.want)
		}
	}
}

func TestDecidePRFirstMatch(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"docs_only":true,"infra":true}`, "T3"},
		{`{"existing_module":true,"new_module":true}`, "T2"},
		{`{"existing_module":true,"docs_only":true}`, "T1"},
		{`{"config_tweak":true}`, "T0"},
	}
	for _, c := range cases {
		got, err := Decide(shadow.PathPR, []byte(c.in))
		if err != nil || got.Label != c.want {
			t.Fatalf("%s: got %+v err %v want %s", c.in, got, err, c.want)
		}
	}
}

func TestDecideStop(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"same_tool":true,"same_error_class":true,"count":3}`, "stop"},
		{`{"same_tool":true,"same_error_class":true,"count":2}`, "continue"},
		{`{"same_tool":false,"same_error_class":true,"count":9}`, "continue"},
		{`{"same_tool":true,"same_error_class":false,"count":9}`, "continue"},
	}
	for _, c := range cases {
		got, err := Decide(shadow.PathStop, []byte(c.in))
		if err != nil || got.Label != c.want {
			t.Fatalf("%s: got %+v err %v want %s", c.in, got, err, c.want)
		}
	}
}

func TestDecideRefuses(t *testing.T) {
	bad := []struct{ path, in string }{
		{shadow.PathAudit, `{"kind":"naming"}`},
		{"nope", `{}`},
		{shadow.PathGo, `{"name_ok":true,"action":"delete_all"}`},
		{shadow.PathGo, `{"name_ok":true,"action":"write_file","extra":1}`},
		{shadow.PathGo, ``},
		{shadow.PathGo, `{"name_ok":true,"action":"write_file"} {}`},
		{shadow.PathPR, `{}`},
		{shadow.PathStop, `{"count":-1}`},
	}
	for _, b := range bad {
		_, err := Decide(b.path, []byte(b.in))
		if err == nil || !strings.HasPrefix(err.Error(), "refused:") {
			t.Fatalf("%s %s: want refused, got %v", b.path, b.in, err)
		}
	}
}

func TestScoreFixtures(t *testing.T) {
	cases, err := shadow.LoadFixtures()
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Score(cases, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !rep.OverallPass || len(rep.Paths) != 3 || rep.Note != Circularity {
		t.Fatalf("report %+v", rep)
	}
	for _, p := range rep.Paths {
		if p.Total == 0 || p.Agree != p.Total || len(p.Misses) != 0 {
			t.Fatalf("path %+v", p)
		}
	}
	if !strings.Contains(Markdown(rep), "by construction") {
		t.Fatal("markdown missing circularity note")
	}
}

func TestScoreCountsMiss(t *testing.T) {
	c := shadow.Case{ID: "x", Path: shadow.PathGo, Actual: "go", Go: shadow.GoFeatures{NameOK: false, Action: "write_file"}}
	rep, err := Score([]shadow.Case{c}, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if rep.OverallPass || rep.Paths[0].Misses[0] != "x" || *rep.Paths[0].Pass {
		t.Fatalf("report %+v", rep)
	}
}

func TestGateBoundary(t *testing.T) {
	cases := []struct {
		agree, total int
		want         bool
	}{{18, 20, true}, {17, 20, false}, {0, 0, false}}
	for _, c := range cases {
		got := gate(PathScore{Path: shadow.PathGo, Agree: c.agree, Total: c.total})
		if got.Pass == nil || *got.Pass != c.want || *got.ThresholdPct != 90 {
			t.Fatalf("%+v: got %+v", c, got)
		}
	}
	if got := gate(PathScore{Path: shadow.PathStop, Agree: 1, Total: 1}); got.Pass != nil || got.DisplayPct != 100 {
		t.Fatalf("stop gated: %+v", got)
	}
}

func TestWriteFiles(t *testing.T) {
	dir := t.TempDir()
	rep := Report{Note: Circularity, Paths: []PathScore{gate(PathScore{Path: shadow.PathPR, Agree: 19, Total: 20})}}
	if err := Write(dir, rep); err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(filepath.Join(dir, "plain-decide-agreement.md"))
	if err != nil || !strings.Contains(string(md), "| pr_class | 19 | 20 | 95.0% | >= 95% | true |") {
		t.Fatalf("md %q err %v", md, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "plain-decide-agreement.json"))
	if err != nil || !strings.Contains(string(raw), `"threshold_pct": 95`) {
		t.Fatalf("json %s err %v", raw, err)
	}
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(filepath.Join(blocker, "sub"), rep); err == nil {
		t.Fatal("want error under a file")
	}
}
