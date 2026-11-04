package rules

import "testing"

func TestDisplayPctAndGates(t *testing.T) {
	if DisplayPct(0, 0) != 0 || DisplayPct(2, 3) != 66.7 || DisplayPct(19, 20) != 95 {
		t.Fatal("display pct")
	}
	if g, ok := GatePct(PathGo); !ok || g != 90 {
		t.Fatal("go gate")
	}
	if g, ok := GatePct(PathPR); !ok || g != 95 {
		t.Fatal("pr gate")
	}
	if _, ok := GatePct(PathStop); ok {
		t.Fatal("stop has no gate")
	}
}
