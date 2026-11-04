package decide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/rules"
)

// Circularity is printed in every report. The fixtures are checked against the
// same rules this layer calls, so agreement is 100% by construction.
const Circularity = "Agreement is 100% by construction: the fixtures' actual labels are validated against the same rules.Rule* functions this layer calls. This is not an independent result; it only shows the plain layer wires those rules correctly."

// PathScore is one path's agreement.
type PathScore struct {
	Path         string   `json:"path"`
	Total        int      `json:"total"`
	Agree        int      `json:"agree"`
	DisplayPct   float64  `json:"display_pct"`
	ThresholdPct *int     `json:"threshold_pct"`
	Pass         *bool    `json:"pass"`
	Misses       []string `json:"misses"`
}

// Report is the plain agreement file.
type Report struct {
	GeneratedAt string      `json:"generated_at"`
	Mode        string      `json:"mode"`
	Note        string      `json:"note"`
	Paths       []PathScore `json:"paths"`
	OverallPass bool        `json:"overall_pass"`
}

// Score runs every case on a plain path through Decide. Cases on other paths are skipped.
func Score(cases []rules.Case, now time.Time) (Report, error) {
	by := map[string]*PathScore{}
	for _, p := range Paths {
		by[p] = &PathScore{Path: p, Misses: []string{}}
	}
	for _, c := range cases {
		if ps, ok := by[c.Path]; ok {
			if err := scoreCase(ps, c); err != nil {
				return Report{}, err
			}
		}
	}
	rep := Report{GeneratedAt: now.UTC().Format(time.RFC3339), Mode: "plain", Note: Circularity, OverallPass: true}
	for _, p := range Paths {
		ps := gate(*by[p])
		if ps.Pass != nil && !*ps.Pass {
			rep.OverallPass = false
		}
		rep.Paths = append(rep.Paths, ps)
	}
	return rep, nil
}

func scoreCase(ps *PathScore, c rules.Case) error {
	raw, err := features(c)
	if err != nil {
		return err
	}
	ps.Total++
	res, err := Decide(c.Path, raw)
	if err == nil && res.Label == c.Actual {
		ps.Agree++
		return nil
	}
	ps.Misses = append(ps.Misses, c.ID)
	return nil
}

// gate fills the display pct and, for gated paths, the threshold and pass bit.
func gate(ps PathScore) PathScore {
	ps.DisplayPct = rules.DisplayPct(ps.Agree, ps.Total)
	g, ok := rules.GatePct(ps.Path)
	if !ok {
		return ps
	}
	pass := ps.Total > 0 && ps.Agree*100 >= g*ps.Total
	ps.ThresholdPct, ps.Pass = &g, &pass
	return ps
}

func features(c rules.Case) ([]byte, error) {
	switch c.Path {
	case rules.PathGo:
		return json.Marshal(c.Go)
	case rules.PathPR:
		return json.Marshal(c.PR)
	case rules.PathStop:
		return json.Marshal(c.Stop)
	}
	return nil, fmt.Errorf("path")
}

// Markdown renders the report.
func Markdown(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Plain decide agreement\n\nGenerated %s. Mode %s.\n\n> %s\n\n", r.GeneratedAt, r.Mode, r.Note)
	b.WriteString("| Path | Agree | Total | Pct | Gate | Pass |\n|---|---|---|---|---|---|\n")
	for _, p := range r.Paths {
		gate, pass := "n/a", "n/a"
		if p.ThresholdPct != nil {
			gate = fmt.Sprintf(">= %d%%", *p.ThresholdPct)
			pass = fmt.Sprintf("%t", *p.Pass)
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %.1f%% | %s | %s |\n", p.Path, p.Agree, p.Total, p.DisplayPct, gate, pass)
	}
	fmt.Fprintf(&b, "\nOverall pass: %t\n", r.OverallPass)
	return b.String()
}

// Write saves plain-decide-agreement.json and .md under dir.
func Write(dir string, r Report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "plain-decide-agreement.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "plain-decide-agreement.md"), []byte(Markdown(r)), 0o644)
}
