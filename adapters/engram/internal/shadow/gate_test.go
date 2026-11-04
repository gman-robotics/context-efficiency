package shadow

import (
	"context"
	"errors"
	"testing"
)

func TestCaseForRejectsBadRows(t *testing.T) {
	cases := []struct {
		path string
		row  rawFixture
	}{
		{PathGo, rawFixture{}},
		{PathGo, rawFixture{ID: "g"}},
		{PathPR, rawFixture{ID: "p"}},
		{PathStop, rawFixture{ID: "s"}},
		{PathAudit, rawFixture{ID: "a"}},
		{"nope", rawFixture{ID: "x"}},
	}
	for _, tc := range cases {
		if _, err := caseFor(tc.path, tc.row); err == nil {
			t.Fatalf("accepted %s %#v", tc.path, tc.row)
		}
	}
	if pathByFile("fixtures/other.json") != "" {
		t.Fatal("unknown file mapped")
	}
}

func TestCaseForAcceptsEachPath(t *testing.T) {
	cases := []rawFixture{
		{ID: "g", Go: &GoFeatures{}},
		{ID: "p", PR: &PRFeatures{}},
		{ID: "s", Stop: &StopFeatures{}},
		{ID: "a", Audit: &AuditFeatures{}},
	}
	for i, path := range []string{PathGo, PathPR, PathStop, PathAudit} {
		c, err := caseFor(path, cases[i])
		if err != nil || c.Path != path || c.Body == "" {
			t.Fatalf("%s: %v %#v", path, err, c)
		}
	}
}

func TestDecideNilDeciderAndCaller(t *testing.T) {
	in := Input{Path: PathGo, StateBody: "STATEBODYMARKER", Actual: "go", Action: "write_file"}
	var d *Decider
	if _, err := d.Decide(context.Background(), in); err == nil || err.Error() != "laya decider not set" {
		t.Fatal(err)
	}
	if _, err := (&Decider{}).Decide(context.Background(), in); err == nil || err.Error() != "laya decider not set" {
		t.Fatal(err)
	}
}

func TestDecideCallerErrorPassesThrough(t *testing.T) {
	want := errors.New("laya transport: x")
	d := NewDecider(&countCaller{err: want}, nil)
	_, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: "STATEBODYMARKER", Actual: "go", Action: "write_file"})
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
}

func TestDecideNilStoreSkipsRetainAndHasNoEffect(t *testing.T) {
	d := NewDecider(&countCaller{choice: "no_go"}, nil)
	out, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: "STATEBODYMARKER", Actual: "go", Action: "write_file"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Effect != "none" || out.LogID != "" || out.Actual != "go" || out.Agree {
		t.Fatalf("%#v", out)
	}
}
