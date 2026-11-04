package rules

import (
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
