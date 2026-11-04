package shadow

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	LayaVersionExpected = "0.4.22"
	ModelPin            = "english"
	WeightsSHA256       = "891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"
	ModeShadow          = "shadow"
	EngramLogSkipped    = "skipped"
	gateGo              = 90
	gatePR              = 95
)

// Item is one fixture result. Err is set for a Laya miss. The state body is not stored.
type Item struct {
	ID               string  `json:"id"`
	Actual           string  `json:"actual"`
	Shadow           string  `json:"shadow"`
	Confidence       float64 `json:"confidence"`
	AnswerConfidence float64 `json:"answer_confidence"`
	Model            string  `json:"model"`
	StateHash        string  `json:"state_hash"`
	Agree            bool    `json:"agree"`
	Err              string  `json:"err,omitempty"`
}

// PathReport is one decision path. Nil threshold and pass marshal as JSON null.
type PathReport struct {
	Path         string  `json:"path"`
	Total        int     `json:"total"`
	Agree        int     `json:"agree"`
	Errors       int     `json:"errors"`
	DisplayPct   float64 `json:"display_pct"`
	ThresholdPct *int    `json:"threshold_pct"`
	Pass         *bool   `json:"pass"`
	Items        []Item  `json:"items"`
}

// Report is the agreement file. WeightsSHA256 is the expected pin, not a measurement.
type Report struct {
	GeneratedAt         string       `json:"generated_at"`
	LayaVersionExpected string       `json:"laya_version_expected"`
	ModelPin            string       `json:"model_pin"`
	WeightsSHA256       string       `json:"weights_sha256"`
	Mode                string       `json:"mode"`
	EngramLog           string       `json:"engram_log"`
	Paths               []PathReport `json:"paths"`
	OverallPass         bool         `json:"overall_pass"`
}

// DisplayPct rounds agree/total to one decimal. It is not the pass bit.
func DisplayPct(agree, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(agree)*1000/float64(total)) / 10
}

func gatePct(path string) (int, bool) {
	switch path {
	case PathGo:
		return gateGo, true
	case PathPR:
		return gatePR, true
	default:
		return 0, false
	}
}

func summarize(path string, items []Item) PathReport {
	if items == nil {
		items = []Item{}
	}
	agree := 0
	errors := 0
	for i := range items {
		if items[i].Err != "" {
			items[i].Agree = false
			errors++
			continue
		}
		if items[i].Agree {
			agree++
		}
	}
	total := len(items)
	out := PathReport{
		Path:       path,
		Total:      total,
		Agree:      agree,
		Errors:     errors,
		DisplayPct: DisplayPct(agree, total),
		Items:      items,
	}
	gate, gated := gatePct(path)
	if !gated {
		return out
	}
	pass := total > 0 && errors == 0 && agree*100 >= gate*total
	out.ThresholdPct = &gate
	out.Pass = &pass
	return out
}

func overallPass(paths []PathReport) bool {
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if p.Errors != 0 {
			return false
		}
		if _, gated := gatePct(p.Path); !gated {
			continue
		}
		if p.Pass == nil || !*p.Pass {
			return false
		}
	}
	return true
}

// BuildReport tallies the four paths. Ungated paths keep null threshold and pass.
func BuildReport(generated time.Time, byPath map[string][]Item) Report {
	order := []string{PathGo, PathPR, PathStop, PathAudit}
	paths := make([]PathReport, 0, len(order))
	for _, path := range order {
		paths = append(paths, summarize(path, byPath[path]))
	}
	return Report{
		GeneratedAt:         generated.UTC().Format(time.RFC3339Nano),
		LayaVersionExpected: LayaVersionExpected,
		ModelPin:            ModelPin,
		WeightsSHA256:       WeightsSHA256,
		Mode:                ModeShadow,
		EngramLog:           EngramLogSkipped,
		Paths:               paths,
		OverallPass:         overallPass(paths),
	}
}

func cleanCell(s string) string {
	return strings.NewReplacer("<", "", ">", "", "\n", " ", "\r", " ", "|", "/").Replace(s)
}

func thresholdCell(p *int) string {
	if p == nil {
		return "null"
	}
	return strconv.Itoa(*p)
}

func passCell(p *bool) string {
	if p == nil {
		return "null"
	}
	if *p {
		return "true"
	}
	return "false"
}

// Markdown is one table plus a disagreement list. It has no state bodies and no links.
func Markdown(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Laya shadow agreement\n\n")
	fmt.Fprintf(&b, "generated_at: %s\n", r.GeneratedAt)
	fmt.Fprintf(&b, "laya_version_expected: %s\n", r.LayaVersionExpected)
	fmt.Fprintf(&b, "model_pin: %s\n", r.ModelPin)
	fmt.Fprintf(&b, "weights_sha256: %s\n", r.WeightsSHA256)
	fmt.Fprintf(&b, "mode: %s\n", r.Mode)
	fmt.Fprintf(&b, "engram_log: %s\n", r.EngramLog)
	fmt.Fprintf(&b, "overall_pass: %t\n\n", r.OverallPass)
	b.WriteString("| path | total | agree | errors | display_pct | threshold_pct | pass |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, p := range r.Paths {
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %.1f | %s | %s |\n",
			p.Path, p.Total, p.Agree, p.Errors, p.DisplayPct, thresholdCell(p.ThresholdPct), passCell(p.Pass))
	}
	b.WriteString("\n## Disagreements\n\n")
	n := 0
	for _, p := range r.Paths {
		for _, it := range p.Items {
			if it.Err == "" && it.Agree {
				continue
			}
			n++
			fmt.Fprintf(&b, "- id=%s actual=%s shadow=%s err=%s\n",
				cleanCell(it.ID), cleanCell(it.Actual), cleanCell(it.Shadow), cleanCell(it.Err))
		}
	}
	if n == 0 {
		b.WriteString("None.\n")
	}
	return b.String()
}

// WriteReport writes the json and markdown files. A markdown failure removes the json file.
func WriteReport(dir string, r Report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	jsonPath := filepath.Join(dir, "laya-shadow-agreement.json")
	mdPath := filepath.Join(dir, "laya-shadow-agreement.md")
	if err := os.WriteFile(jsonPath, raw, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(mdPath, []byte(Markdown(r)), 0o644); err != nil {
		_ = os.Remove(jsonPath)
		return err
	}
	return nil
}

// ReportExit is 1 when the file was not written, 0 on pass, and 2 on a written failure.
func ReportExit(wrote bool, pass bool) int {
	if !wrote {
		return 1
	}
	if pass {
		return 0
	}
	return 2
}

func safeErr(err error, body string) string {
	if err == nil {
		return ""
	}
	return clip(err.Error(), body)
}

func clip(msg, body string) string {
	if len(body) >= 8 {
		msg = strings.ReplaceAll(msg, body, "redacted")
	}
	msg = cleanCell(msg)
	if len(msg) > 180 {
		msg = msg[:180]
	}
	return msg
}
