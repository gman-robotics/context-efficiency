package shadow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

func TestStateHashABC(t *testing.T) {
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if StateHash("abc") != want {
		t.Fatalf("%s", StateHash("abc"))
	}
}

func TestReportExit(t *testing.T) {
	if ReportExit(false, true) != 1 || ReportExit(true, true) != 0 || ReportExit(true, false) != 2 {
		t.Fatal("exit")
	}
}

func TestThresholdMath(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	pr19 := append(filled(19, "T3", "T3", true), Item{ID: "miss", Actual: "T3", Shadow: "T0", Agree: false})
	rep := BuildReport(now, map[string][]Item{
		PathGo:    filled(20, "go", "go", true),
		PathPR:    pr19,
		PathStop:  filled(1, "continue", "continue", true),
		PathAudit: filled(1, "note", "note", true),
	})
	pr := findPath(t, rep, PathPR)
	if pr.Pass == nil || !*pr.Pass || fmt.Sprintf("%.1f", pr.DisplayPct) != "95.0" {
		t.Fatalf("19/20 %#v", pr)
	}
	if !rep.OverallPass {
		t.Fatal("19/20 overall")
	}

	pr18 := append(filled(18, "T0", "T0", true), filled(2, "T0", "T1", false)...)
	rep = BuildReport(now, map[string][]Item{
		PathGo:    filled(20, "go", "go", true),
		PathPR:    pr18,
		PathStop:  filled(1, "continue", "continue", true),
		PathAudit: filled(1, "note", "note", true),
	})
	pr = findPath(t, rep, PathPR)
	if pr.Pass == nil || *pr.Pass || fmt.Sprintf("%.1f", pr.DisplayPct) != "90.0" || rep.OverallPass {
		t.Fatalf("18/20 pr pass %v display %v overall %v", pr.Pass, pr.DisplayPct, rep.OverallPass)
	}

	rep = BuildReport(now, map[string][]Item{
		PathGo:    append(filled(18, "go", "go", true), filled(2, "go", "no_go", false)...),
		PathPR:    filled(20, "T1", "T1", true),
		PathStop:  filled(1, "stop", "stop", true),
		PathAudit: filled(1, "nit", "nit", true),
	})
	gng := findPath(t, rep, PathGo)
	if gng.Pass == nil || !*gng.Pass {
		t.Fatal("18/20 go")
	}
	rep = BuildReport(now, map[string][]Item{
		PathGo:    append(filled(17, "go", "go", true), filled(3, "go", "no_go", false)...),
		PathPR:    filled(20, "T1", "T1", true),
		PathStop:  filled(1, "stop", "stop", true),
		PathAudit: filled(1, "nit", "nit", true),
	})
	gng = findPath(t, rep, PathGo)
	if gng.Pass == nil || *gng.Pass {
		t.Fatal("17/20 go")
	}

	rep = BuildReport(now, map[string][]Item{
		PathGo:    {},
		PathPR:    filled(20, "T1", "T1", true),
		PathStop:  {},
		PathAudit: {},
	})
	gng = findPath(t, rep, PathGo)
	if gng.Pass == nil || *gng.Pass || gng.Total != 0 {
		t.Fatal("empty go")
	}
	stop := findPath(t, rep, PathStop)
	if stop.ThresholdPct != nil || stop.Pass != nil {
		t.Fatal("ungated should be null")
	}

	rep = BuildReport(now, map[string][]Item{
		PathGo:    []Item{{ID: "e", Actual: "go", Shadow: "go", Agree: true, Err: "laya http 500: x"}},
		PathPR:    filled(20, "T1", "T1", true),
		PathStop:  filled(1, "continue", "continue", true),
		PathAudit: filled(1, "note", "note", true),
	})
	gng = findPath(t, rep, PathGo)
	if gng.Errors != 1 || gng.Agree != 0 || gng.Total != 1 || gng.Items[0].Agree || gng.Pass == nil || *gng.Pass {
		t.Fatalf("error item %#v", gng)
	}

	rep = BuildReport(now, map[string][]Item{
		PathGo:    filled(20, "go", "go", true),
		PathPR:    filled(20, "T2", "T2", true),
		PathStop:  []Item{{ID: "s", Actual: "stop", Shadow: "continue", Agree: false}},
		PathAudit: filled(1, "blocker", "blocker", true),
	})
	if !rep.OverallPass {
		t.Fatal("path 3 disagreement cleared overall")
	}
	stop = findPath(t, rep, PathStop)
	if stop.Pass != nil || stop.ThresholdPct != nil {
		t.Fatal("path 3 gate set")
	}

	rep = BuildReport(now, map[string][]Item{
		PathGo:    filled(20, "go", "go", true),
		PathPR:    filled(20, "T2", "T2", true),
		PathStop:  []Item{{ID: "s", Actual: "stop", Err: "laya http 500: x", Agree: true}},
		PathAudit: filled(1, "blocker", "blocker", true),
	})
	if rep.OverallPass || findPath(t, rep, PathStop).Errors != 1 {
		t.Fatal("path 3 error should fail overall")
	}
	if fmt.Sprintf("%.1f", DisplayPct(1, 3)) != "33.3" {
		t.Fatalf("display %v", DisplayPct(1, 3))
	}
}

func TestReportOmitsStateAndNullThresholds(t *testing.T) {
	const sentinel = "SENTINELBODYTOKEN"
	now := time.Date(2026, 5, 2, 1, 2, 3, 4, time.UTC)
	caller := &countCaller{choice: "go"}
	cases := []Case{{
		ID: "one", Path: PathGo, Actual: "go", Action: "write_file", Body: sentinel + " please merge later",
		Go: GoFeatures{NameOK: true, Action: "write_file"},
	}}
	rep, down, err := RunCases(context.Background(), caller, cases, now)
	if err != nil || down {
		t.Fatal(err, down)
	}
	if findPath(t, rep, PathGo).Agree != 1 {
		t.Fatal("sentinel row missing")
	}
	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	md := Markdown(rep)
	if strings.Contains(string(raw), sentinel) || strings.Contains(md, sentinel) || strings.Contains(md, "merge") {
		t.Fatal("state leaked")
	}
	if strings.Contains(md, "<") || strings.Contains(md, ">") || strings.Contains(md, "](") {
		t.Fatal(md)
	}
	dir := t.TempDir()
	gated := BuildReport(now, map[string][]Item{
		PathGo:    filled(1, "go", "go", true),
		PathPR:    filled(1, "T0", "T0", true),
		PathStop:  filled(1, "continue", "stop", false),
		PathAudit: filled(1, "note", "nit", false),
	})
	if err := WriteReport(dir, gated); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dir + "/laya-shadow-agreement.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"threshold_pct": null`) || !strings.Contains(string(got), `"pass": null`) {
		t.Fatal(string(got))
	}
	var back Report
	if err := json.Unmarshal(got, &back); err != nil {
		t.Fatal(err)
	}
	stop := findPath(t, back, PathStop)
	if stop.ThresholdPct != nil || stop.Pass != nil {
		t.Fatal("round trip null")
	}
	mdBytes, err := os.ReadFile(dir + "/laya-shadow-agreement.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mdBytes), "| null | null |") {
		t.Fatal(string(mdBytes))
	}
	if back.EngramLog != EngramLogSkipped || back.LayaVersionExpected != LayaVersionExpected || back.WeightsSHA256 != WeightsSHA256 || back.ModelPin != ModelPin {
		t.Fatalf("%#v", back)
	}
}

func TestRunCasesEchoLoadFixtures(t *testing.T) {
	cases, err := LoadFixtures()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]string{}
	for _, c := range cases {
		by[c.Body] = c.Actual
	}
	caller := &mapCaller{by: by}
	now := time.Unix(10, 0).UTC()
	rep, down, err := RunCases(context.Background(), caller, cases, now)
	if err != nil || down || !rep.OverallPass {
		t.Fatal(err, down, rep.OverallPass)
	}
	if caller.n != 68 {
		t.Fatalf("calls %d", caller.n)
	}
	raw, _ := json.Marshal(rep)
	md := Markdown(rep)
	for _, c := range cases {
		if strings.Contains(string(raw), c.Body) || strings.Contains(md, c.Body) {
			t.Fatalf("body leaked %s", c.ID)
		}
	}
	if strings.Contains(md, "<") || strings.Contains(md, ">") {
		t.Fatal("brackets")
	}
}

func TestRunCasesDownDriftAndItems(t *testing.T) {
	cases, err := LoadFixtures()
	if err != nil {
		t.Fatal(err)
	}
	caller := &stubCaller{err: fmt.Errorf("%w: dial", laya.ErrTransport)}
	_, down, err := RunCases(context.Background(), caller, cases, time.Unix(0, 0))
	if err != nil || !down || caller.n != 1 {
		t.Fatalf("down %v err %v n %d", down, err, caller.n)
	}
	caller = &stubCaller{err: laya.ErrKey}
	_, down, err = RunCases(context.Background(), caller, cases, time.Unix(0, 0))
	if !down || !errors.Is(err, laya.ErrKey) || caller.n != 1 {
		t.Fatal(down, err, caller.n)
	}
	drift := append([]Case(nil), cases...)
	drift[0].Actual = "no_go"
	caller = &stubCaller{}
	_, down, err = RunCases(context.Background(), caller, drift, time.Unix(0, 0))
	if down || err == nil || !strings.Contains(err.Error(), "fixture actual drift:") || caller.n != 0 {
		t.Fatal(down, err, caller.n)
	}
	refused := append([]Case(nil), cases...)
	refused[0].Body = "  "
	caller = &stubCaller{}
	_, _, err = RunCases(context.Background(), caller, refused, time.Unix(0, 0))
	if err == nil || !strings.Contains(err.Error(), "fixture refused:") || caller.n != 0 {
		t.Fatal(err, caller.n)
	}
	_, _, err = RunCases(context.Background(), caller, []Case{{ID: "bad", Path: "nope", Actual: "go"}}, time.Unix(0, 0))
	if err == nil || !strings.Contains(err.Error(), "fixture rule: bad") || caller.n != 0 {
		t.Fatal(err, caller.n)
	}
	caller = &stubCaller{err: errors.New("laya http 500: inference failed")}
	rep, down, err := RunCases(context.Background(), caller, cases, time.Unix(0, 0))
	if err != nil || down || rep.OverallPass || caller.n != 68 {
		t.Fatal(err, down, rep.OverallPass, caller.n)
	}
	total := 0
	for _, p := range rep.Paths {
		total += p.Total
		if p.Errors != p.Total {
			t.Fatalf("path %s errors %d total %d", p.Path, p.Errors, p.Total)
		}
	}
	if total != 68 {
		t.Fatal(total)
	}
	seq := &seqCaller{}
	rep, down, err = RunCases(context.Background(), seq, cases, time.Unix(0, 0))
	if err != nil || down || seq.n != 68 || rep.OverallPass {
		t.Fatal(err, down, seq.n, rep.OverallPass)
	}
}

func TestSafeErrStripsBody(t *testing.T) {
	const body = "hello-body-xx"
	msg := safeErr(errors.New("boom <x> | "+body), body)
	if strings.ContainsAny(msg, "<>|") || strings.Contains(msg, body) {
		t.Fatal(msg)
	}
}

type mapCaller struct {
	by map[string]string
	n  int
}

func (m *mapCaller) SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error) {
	m.n++
	choice, ok := m.by[body]
	if !ok {
		return laya.Parsed{}, errors.New("missing fixture")
	}
	return laya.Parsed{Choice: choice, Confidence: 0.2, AnswerConfidence: 0.1, Model: "english"}, nil
}

type stubCaller struct {
	n   int
	err error
}

func (s *stubCaller) SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error) {
	s.n++
	if s.err != nil {
		return laya.Parsed{}, s.err
	}
	return laya.Parsed{Choice: "go", Model: "english"}, nil
}

type seqCaller struct {
	n int
}

func (s *seqCaller) SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error) {
	s.n++
	if s.n == 1 {
		return laya.Parsed{Choice: "go", Model: "english"}, nil
	}
	return laya.Parsed{}, fmt.Errorf("%w: later", laya.ErrTransport)
}

func filled(n int, actual, shadow string, agree bool) []Item {
	items := make([]Item, n)
	for i := range items {
		items[i] = Item{ID: fmt.Sprintf("%s-%d", actual, i), Actual: actual, Shadow: shadow, Agree: agree, Model: "english"}
	}
	return items
}

func findPath(t *testing.T, rep Report, path string) PathReport {
	t.Helper()
	for _, p := range rep.Paths {
		if p.Path == path {
			return p
		}
	}
	t.Fatalf("missing %s", path)
	return PathReport{}
}
