package front

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

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
