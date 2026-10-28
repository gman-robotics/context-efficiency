package adaptermap

import "context"

// EngramCaller is the engram child MCP surface the adapter drives (implemented by child.Client and test fakes).
type EngramCaller interface {
	CallTool(ctx context.Context, name string, args map[string]interface{}) (text string, isError bool, err error)
}
