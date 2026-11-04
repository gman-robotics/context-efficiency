package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Caller calls one MCP tool and returns its text and whether it was an error.
type Caller interface {
	Call(ctx context.Context, name string, args map[string]any) (string, bool, error)
}

type caller struct{ c *client.Client }

func (k caller) Call(ctx context.Context, name string, args map[string]any) (string, bool, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := k.c.CallTool(ctx, req)
	if err != nil {
		return "", false, err
	}
	var b strings.Builder
	for _, part := range res.Content {
		if t, ok := part.(mcp.TextContent); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String(), res.IsError, nil
}

// Step is one end-to-end check. Run returns "" on pass, else the reason.
type Step struct {
	Name string
	Run  func(ctx context.Context, c Caller) string
}

const scope = "bot"
const fact = "ce-e2e fact: the plain decide layer has effect none"

// Steps is the ordered list. Recall depends on retain; handoff_read on handoff_write.
func Steps() []Step {
	steps := []Step{
		{"retain", stepRetain},
		{"recall", stepRecall},
		{"handoff_write", stepHandoffWrite},
		{"handoff_read", stepHandoffRead},
	}
	for _, d := range decideCases() {
		steps = append(steps, Step{"decide " + d.name, d.run})
	}
	return steps
}

func ok(ctx context.Context, c Caller, name string, args map[string]any) (string, string) {
	text, isErr, err := c.Call(ctx, name, args)
	if err != nil {
		return "", "transport: " + err.Error()
	}
	if isErr {
		return "", "tool error: " + text
	}
	return text, ""
}

func stepRetain(ctx context.Context, c Caller) string {
	id, why := ok(ctx, c, "retain", map[string]any{"text": fact, "importance": 3, "tags": []string{"e2e"}, "scope": scope})
	if why == "" && strings.TrimSpace(id) == "" {
		return "empty id"
	}
	return why
}

func stepRecall(ctx context.Context, c Caller) string {
	text, why := ok(ctx, c, "recall", map[string]any{"query": "plain decide layer", "scope": scope, "limit": 5})
	if why == "" && !strings.Contains(text, fact) {
		return "retained fact not recalled: " + text
	}
	return why
}

var handoff = map[string]any{
	"goal":      "GMA-543 local e2e",
	"openItems": []string{"none"},
	"decisions": []string{"plain decide, no Laya"},
	"filePaths": []string{"scripts/e2e-local.sh"},
	"nextSteps": []string{"Kit runs CE-6"},
}

func stepHandoffWrite(ctx context.Context, c Caller) string {
	_, why := ok(ctx, c, "handoff_write", map[string]any{"handoff": handoff, "scope": scope})
	return why
}

func stepHandoffRead(ctx context.Context, c Caller) string {
	text, why := ok(ctx, c, "handoff_read", map[string]any{"scope": scope})
	if why != "" {
		return why
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		return "handoff not JSON: " + text
	}
	for k := range handoff {
		g, _ := json.Marshal(got[k])
		w, _ := json.Marshal(handoff[k])
		if string(g) != string(w) {
			return fmt.Sprintf("handoff %s = %s, want %s", k, g, w)
		}
	}
	return ""
}

type decideCase struct {
	name     string
	path     string
	features map[string]any
	label    string // "" means the call must be refused
}

func decideCases() []decideCase {
	cases := []decideCase{
		{"routine_go_nogo go", "routine_go_nogo", map[string]any{"name_ok": true, "action": "write_file", "blocked": false}, "go"},
		{"routine_go_nogo no_go", "routine_go_nogo", map[string]any{"name_ok": true, "action": "read_only"}, "no_go"},
		{"pr_class T3", "pr_class", map[string]any{"infra": true}, "T3"},
		{"pr_class T0", "pr_class", map[string]any{"docs_only": true}, "T0"},
		{"tool_failure_stop stop", "tool_failure_stop", map[string]any{"same_tool": true, "same_error_class": true, "count": 3}, "stop"},
		{"tool_failure_stop continue", "tool_failure_stop", map[string]any{"same_tool": true, "same_error_class": false, "count": 5}, "continue"},
		{"unknown path refused", "audit_sort", map[string]any{"kind": "x"}, ""},
	}
	// Sends, merges, deploys and money are not decision-layer actions.
	for _, a := range []string{"send_email", "merge", "deploy", "pay"} {
		cases = append(cases, decideCase{a + " refused", "routine_go_nogo", map[string]any{"name_ok": true, "action": a}, ""})
	}
	return cases
}

func (d decideCase) run(ctx context.Context, c Caller) string {
	text, isErr, err := c.Call(ctx, "decide", map[string]any{"path": d.path, "features": d.features})
	if err != nil {
		return "transport: " + err.Error()
	}
	if d.label == "" {
		if isErr && strings.HasPrefix(text, "refused:") {
			return ""
		}
		return "want refused:, got " + text
	}
	if isErr {
		return "tool error: " + text
	}
	var r struct{ Label, Effect string }
	if err := json.Unmarshal([]byte(text), &r); err != nil {
		return "decide not JSON: " + text
	}
	if r.Label != d.label || r.Effect != "none" {
		return fmt.Sprintf("got label %q effect %q, want %q none", r.Label, r.Effect, d.label)
	}
	return ""
}
