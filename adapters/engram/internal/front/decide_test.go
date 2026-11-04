package front

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func callDecide(t *testing.T, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	res, err := decideHandler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestDecideToolLabels(t *testing.T) {
	res := callDecide(t, map[string]interface{}{
		"path":     "tool_failure_stop",
		"features": map[string]interface{}{"same_tool": true, "same_error_class": true, "count": 3},
	})
	if res.IsError {
		t.Fatal(text(t, res))
	}
	var out struct{ Label, Effect string }
	if err := json.Unmarshal([]byte(text(t, res)), &out); err != nil {
		t.Fatal(err)
	}
	if out.Label != "stop" || out.Effect != "none" {
		t.Fatalf("%+v", out)
	}
}

func TestDecideToolRefuses(t *testing.T) {
	for _, args := range []map[string]interface{}{
		{"path": "routine_go_nogo"},
		{"path": "audit_sort", "features": map[string]interface{}{"kind": "naming"}},
		{"path": "routine_go_nogo", "features": map[string]interface{}{"name_ok": true, "action": "rm"}},
	} {
		res := callDecide(t, args)
		if !res.IsError || !strings.HasPrefix(text(t, res), "refused:") {
			t.Fatalf("%v: %s", args, text(t, res))
		}
	}
}
