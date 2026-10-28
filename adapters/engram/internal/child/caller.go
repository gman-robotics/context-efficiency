package child

import "context"

// ToolCaller is the engram child MCP surface used by the adapter.
type ToolCaller interface {
	CallTool(ctx context.Context, name string, args map[string]interface{}) (text string, isError bool, err error)
}
