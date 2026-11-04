// Package decide is the plain deterministic decision layer. It calls the shared
// rules in package shadow and never calls Laya.
package decide

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
)

// Result is one plain decision. Effect is always none: the caller acts on Label.
type Result struct {
	Path   string `json:"path"`
	Label  string `json:"label"`
	Rule   string `json:"rule"`
	Effect string `json:"effect"`
}

// Paths lists the paths the plain layer decides. audit_sort is not one of them.
var Paths = []string{shadow.PathGo, shadow.PathPR, shadow.PathStop}

// Decide parses features for path and applies the shared rule.
// Unknown paths, unknown fields, and unmatched shapes are errors starting with "refused:".
func Decide(path string, features []byte) (Result, error) {
	label, err := apply(path, features)
	if err != nil {
		return Result{}, fmt.Errorf("refused: %s", err)
	}
	crit, err := shadow.Criteria(path)
	if err != nil {
		return Result{}, err
	}
	return Result{Path: path, Label: label, Rule: crit[label], Effect: "none"}, nil
}

func apply(path string, raw []byte) (string, error) {
	switch path {
	case shadow.PathGo:
		var f shadow.GoFeatures
		if err := strict(raw, &f); err != nil {
			return "", err
		}
		return shadow.RuleGo(f)
	case shadow.PathPR:
		var f shadow.PRFeatures
		if err := strict(raw, &f); err != nil {
			return "", err
		}
		return shadow.RulePR(f)
	case shadow.PathStop:
		var f shadow.StopFeatures
		if err := strict(raw, &f); err != nil {
			return "", err
		}
		return shadow.RuleStop(f)
	default:
		return "", fmt.Errorf("path")
	}
}

func strict(raw []byte, v any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("features")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("features")
	}
	if dec.More() {
		return fmt.Errorf("features")
	}
	return nil
}
