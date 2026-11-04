package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result is one step outcome.
type Result struct {
	Step   string `json:"step"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

// Report is the e2e summary written to JSON and Markdown.
type Report struct {
	Commit  string   `json:"commit"`
	Server  string   `json:"server"`
	Tools   []string `json:"tools"`
	Results []Result `json:"results"`
}

// Add records a step; why == "" is a pass.
func (r *Report) Add(step, why string) {
	r.Results = append(r.Results, Result{Step: step, Pass: why == "", Detail: why})
}

// Failed counts failing steps.
func (r *Report) Failed() int {
	n := 0
	for _, x := range r.Results {
		if !x.Pass {
			n++
		}
	}
	return n
}

// Write writes local-e2e.json and local-e2e.md under dir.
func (r *Report) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "local-e2e.json"), append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "local-e2e.md"), []byte(r.Markdown()), 0o644)
}

// Markdown renders the report table.
func (r *Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Local end-to-end (GMA-543)\n\nCommit: `%s`. Server: `%s`.\n\n", r.Commit, r.Server)
	fmt.Fprintf(&b, "Tools listed: %s.\n\n", strings.Join(r.Tools, ", "))
	fmt.Fprintf(&b, "Result: %d of %d steps pass.\n\n| Step | Result | Detail |\n|---|---|---|\n", len(r.Results)-r.Failed(), len(r.Results))
	for _, x := range r.Results {
		res := "pass"
		if !x.Pass {
			res = "FAIL"
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", x.Step, res, strings.ReplaceAll(x.Detail, "|", "\\|"))
	}
	return b.String()
}
