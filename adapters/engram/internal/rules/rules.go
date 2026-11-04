package rules

import (
	"fmt"
)

const (
	PathGo    = "routine_go_nogo"
	PathPR    = "pr_class"
	PathStop  = "tool_failure_stop"
	PathAudit = "audit_sort"
)

// GoFeatures is the go/no-go rubric. It has no failure count.
type GoFeatures struct {
	NameOK  bool   `json:"name_ok"`
	Action  string `json:"action"`
	Blocked bool   `json:"blocked"`
}

// PRFeatures is the standards tier rubric. First match wins.
type PRFeatures struct {
	Infra          bool `json:"infra"`
	Migration      bool `json:"migration"`
	NewService     bool `json:"new_service"`
	NewModule      bool `json:"new_module"`
	CrossCutting   bool `json:"cross_cutting"`
	ExistingModule bool `json:"existing_module"`
	DocsOnly       bool `json:"docs_only"`
	CheckOnly      bool `json:"check_only"`
	ConfigTweak    bool `json:"config_tweak"`
}

// StopFeatures is the tool-retry rubric.
type StopFeatures struct {
	SameTool       bool   `json:"same_tool"`
	SameErrorClass bool   `json:"same_error_class"`
	Count          int    `json:"count"`
	Tool           string `json:"tool"`
	OtherTool      string `json:"other_tool"`
	ErrorClass     string `json:"error_class"`
	OtherError     string `json:"other_error"`
}

// AuditFeatures is the audit-bucket rubric.
type AuditFeatures struct {
	Kind string `json:"kind"`
}

// Case is one embedded fixture row plus its rendered state.
type Case struct {
	ID     string
	Path   string
	Actual string
	Action string
	Labels []string
	Body   string
	Go     GoFeatures
	PR     PRFeatures
	Stop   StopFeatures
	Audit  AuditFeatures
}

// Rule applies the rule for c.Path. The fixture census test uses it.
func Rule(c Case) (string, error) {
	switch c.Path {
	case PathGo:
		return RuleGo(c.Go)
	case PathPR:
		return RulePR(c.PR)
	case PathStop:
		return RuleStop(c.Stop)
	case PathAudit:
		return RuleAudit(c.Audit)
	default:
		return "", fmt.Errorf("path")
	}
}

// RuleGo returns go only when the name is ok, the action is write_file or retain_summary, and blocked is false.
func RuleGo(f GoFeatures) (string, error) {
	switch f.Action {
	case "write_file", "retain_summary", "read_only":
	default:
		return "", fmt.Errorf("action not allowed")
	}
	if f.NameOK && !f.Blocked && (f.Action == "write_file" || f.Action == "retain_summary") {
		return "go", nil
	}
	return "no_go", nil
}

// RulePR applies the standards scale. An unmatched shape is an error.
func RulePR(f PRFeatures) (string, error) {
	if f.Infra || f.Migration || f.NewService {
		return "T3", nil
	}
	if f.NewModule || f.CrossCutting {
		return "T2", nil
	}
	if f.ExistingModule {
		return "T1", nil
	}
	if f.DocsOnly || f.CheckOnly || f.ConfigTweak {
		return "T0", nil
	}
	return "", fmt.Errorf("pr tier unmatched")
}

// RuleStop returns stop only when both flags are set and count is at least 3.
func RuleStop(f StopFeatures) (string, error) {
	if f.Count < 0 {
		return "", fmt.Errorf("count")
	}
	if f.SameTool && f.SameErrorClass && f.Count >= 3 {
		return "stop", nil
	}
	return "continue", nil
}

// RuleAudit maps a finding kind to a bucket.
func RuleAudit(f AuditFeatures) (string, error) {
	switch f.Kind {
	case "security", "data_loss", "test_failure":
		return "blocker", nil
	case "behavior_bug":
		return "should_fix", nil
	case "naming":
		return "nit", nil
	case "typo_question":
		return "note", nil
	default:
		return "", fmt.Errorf("audit kind unmatched")
	}
}

// Criteria returns the fixed choice glosses for a path. Each map has 2 to 4 entries.
func Criteria(path string) (map[string]string, error) {
	switch path {
	case PathGo:
		return map[string]string{
			"go":    "Name ok, action is write_file or retain_summary, and blocked is false.",
			"no_go": "Name is not ok, action is read_only, or blocked is true.",
		}, nil
	case PathPR:
		return map[string]string{
			"T0": "Docs only, check only, or config tweak, including typos.",
			"T1": "Existing module work.",
			"T2": "New module or cross cutting work.",
			"T3": "Infra, migration, or new service, a major architectural move.",
		}, nil
	case PathStop:
		return map[string]string{
			"continue": "Count under 3, or the tool differs, or the error class differs.",
			"stop":     "Same tool, same error class, and count is 3 or more.",
		}, nil
	case PathAudit:
		return map[string]string{
			"blocker":    "Security, data loss, or test failure.",
			"should_fix": "Behavior bug.",
			"nit":        "Naming.",
			"note":       "Typo question.",
		}, nil
	default:
		return nil, fmt.Errorf("refused: path")
	}
}

// RenderGo is the fixed state paragraph for path routine_go_nogo.
func RenderGo(f GoFeatures) string {
	return fmt.Sprintf("Routine gate. Name ok is %t. Action is %s. Blocked is %t. Label go only when the name is ok, the action is write_file or retain_summary, and blocked is false. Otherwise label no_go.", f.NameOK, f.Action, f.Blocked)
}

// RenderPR is the fixed state paragraph for path pr_class.
func RenderPR(f PRFeatures) string {
	return fmt.Sprintf("PR tier. Infra is %t. Migration is %t. New service is %t. New module is %t. Cross cutting is %t. Existing module is %t. Docs only is %t. Check only is %t. Config tweak is %t. First match: infra or migration or new service is T3, a major architectural move. Else new module or cross cutting is T2. Else existing module is T1. Else docs only or check only or config tweak is T0, including typos.",
		f.Infra, f.Migration, f.NewService, f.NewModule, f.CrossCutting, f.ExistingModule, f.DocsOnly, f.CheckOnly, f.ConfigTweak)
}

// RenderStop is the fixed state paragraph for path tool_failure_stop.
func RenderStop(f StopFeatures) string {
	return fmt.Sprintf("Tool retry gate. Tool is %s. Other tool is %s. Error class is %s. Other error class is %s. Same tool is %t. Same error class is %t. Count is %d. Label stop when the same tool and the same error class and the count is 3 or more. Otherwise label continue.",
		f.Tool, f.OtherTool, f.ErrorClass, f.OtherError, f.SameTool, f.SameErrorClass, f.Count)
}

// RenderAudit is the fixed state paragraph for path audit_sort.
func RenderAudit(f AuditFeatures) string {
	return fmt.Sprintf("Audit bucket. Kind is %s. Label blocker for security, data loss, or test failure. Label should_fix for a behavior bug. Label nit for naming. Label note for a typo question.", f.Kind)
}
