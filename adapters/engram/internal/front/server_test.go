package front

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
	"github.com/mark3labs/mcp-go/mcp"
)

type stubCaller struct {
	choice string
	err    error
	body   string
}

func (s *stubCaller) SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error) {
	s.body = body
	if s.err != nil {
		return laya.Parsed{}, s.err
	}
	return laya.Parsed{Choice: s.choice, Model: "english"}, nil
}

func call(t *testing.T, dec *shadow.Decider, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	res, err := decideShadowHandler(dec)(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func text(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("content %d", len(res.Content))
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("%T", res.Content[0])
	}
	return tc.Text
}

func goArgs() map[string]interface{} {
	return map[string]interface{}{
		"path":       shadow.PathGo,
		"state_body": "STATEBODYMARKER routine gate",
		"actual":     "go",
		"action":     "write_file",
		"labels":     []interface{}{"a", 7, "b"},
		"request_id": "req-1",
	}
}

func TestDecideShadowNilDecider(t *testing.T) {
	res := call(t, nil, goArgs())
	if !res.IsError || text(t, res) != "laya decider not set" {
		t.Fatalf("%#v", res)
	}
}

func TestDecideShadowErrorIsToolError(t *testing.T) {
	dec := shadow.NewDecider(&stubCaller{err: errors.New("laya transport: down")}, nil)
	res := call(t, dec, goArgs())
	if !res.IsError || text(t, res) != "laya transport: down" {
		t.Fatalf("%#v", res)
	}
}

// The tool result never replaces actual: effect is none and actual is echoed as given.
func TestDecideShadowDisagreementKeepsActual(t *testing.T) {
	caller := &stubCaller{choice: "no_go"}
	res := call(t, shadow.NewDecider(caller, nil), goArgs())
	if res.IsError {
		t.Fatal(text(t, res))
	}
	var out shadow.Outcome
	if err := json.Unmarshal([]byte(text(t, res)), &out); err != nil {
		t.Fatal(err)
	}
	if out.Effect != "none" || out.Actual != "go" || out.Shadow != "no_go" || out.Agree || out.RequestID != "req-1" {
		t.Fatalf("%#v", out)
	}
	if caller.body != "STATEBODYMARKER routine gate" {
		t.Fatal(caller.body)
	}
}

func TestLayaDecider(t *testing.T) {
	env := map[string]string{"LAYA_API_KEY": "  ", "LAYA_BASE_URL": "http://example.com:1"}
	get := func(k string) string { return env[k] }
	if d, err := LayaDecider(nil, get); err != nil || d == nil {
		t.Fatalf("empty key must still start: %v", err)
	}
	env["LAYA_API_KEY"] = "k-test"
	if _, err := LayaDecider(nil, get); err == nil {
		t.Fatal("non-loopback url accepted with key set")
	}
	env["LAYA_BASE_URL"] = ""
	if d, err := LayaDecider(nil, get); err != nil || d == nil {
		t.Fatal(err)
	}
}
