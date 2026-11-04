package front

import (
	"context"
	"encoding/json"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/adaptermap"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/decide"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RunHost serves domain MCP tools on stdio.
func RunHost(ctx context.Context, svc *adaptermap.Service) error {
	s := server.NewMCPServer("context-efficiency-mcp", "0.1.0")

	s.AddTool(mcp.NewTool("recall",
		mcp.WithDescription("Search durable memory"),
		mcp.WithString("query", mcp.Required()),
		mcp.WithString("scope"),
		mcp.WithNumber("limit"),
	), recallHandler(svc))

	s.AddTool(mcp.NewTool("retain",
		mcp.WithDescription("Store a distilled fact"),
		mcp.WithString("text", mcp.Required()),
		mcp.WithArray("tags", mcp.WithStringItems()),
		mcp.WithNumber("importance", mcp.Required()),
		mcp.WithString("scope"),
	), retainHandler(svc))

	s.AddTool(mcp.NewTool("handoff_write",
		mcp.WithDescription("Persist a structured handoff"),
		mcp.WithObject("handoff", mcp.Required()),
		mcp.WithString("scope"),
	), handoffWriteHandler(svc))

	s.AddTool(mcp.NewTool("handoff_read",
		mcp.WithDescription("Load handoff for scope"),
		mcp.WithString("scope"),
	), handoffReadHandler(svc))

	s.AddTool(mcp.NewTool("decide",
		mcp.WithDescription("Plain deterministic decision for routine_go_nogo, pr_class, or tool_failure_stop. Fixed rules, no model. Effect is none; act on label. An error starting with refused: means stop."),
		mcp.WithString("path", mcp.Required()),
		mcp.WithObject("features", mcp.Required()),
	), decideHandler)

	return server.ServeStdio(s)
}

func decideHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	path, _ := args["path"].(string)
	features, ok := args["features"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("refused: features"), nil
	}
	raw, err := json.Marshal(features)
	if err != nil {
		return mcp.NewToolResultError("refused: features"), nil
	}
	out, err := decide.Decide(path, raw)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return mcp.NewToolResultError("decide response unreadable"), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func recallHandler(svc *adaptermap.Service) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q, _ := req.GetArguments()["query"].(string)
		scopeStr, _ := req.GetArguments()["scope"].(string)
		limit := 5
		if l, ok := req.GetArguments()["limit"].(float64); ok {
			limit = int(l)
		}
		items, err := svc.Recall(ctx, q, scopeStr, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.Marshal(items)
		return mcp.NewToolResultText(string(b)), nil
	}
}

func retainHandler(svc *adaptermap.Service) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		text, _ := args["text"].(string)
		scopeStr, _ := args["scope"].(string)
		imp := 1
		if v, ok := args["importance"].(float64); ok {
			imp = int(v)
		}
		var tags []string
		if raw, ok := args["tags"].([]interface{}); ok {
			for _, t := range raw {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		}
		id, err := svc.Retain(ctx, text, tags, imp, scopeStr)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(id), nil
	}
}

func handoffWriteHandler(svc *adaptermap.Service) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		scopeStr, _ := args["scope"].(string)
		raw, ok := args["handoff"].(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("handoff required"), nil
		}
		b, _ := json.Marshal(raw)
		var h domain.Handoff
		if err := json.Unmarshal(b, &h); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := svc.HandoffWrite(ctx, h, scopeStr)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(id), nil
	}
}

func handoffReadHandler(svc *adaptermap.Service) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		scopeStr, _ := req.GetArguments()["scope"].(string)
		h, err := svc.HandoffRead(ctx, scopeStr)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if h == nil {
			return mcp.NewToolResultText("null"), nil
		}
		b, _ := json.Marshal(h)
		return mcp.NewToolResultText(string(b)), nil
	}
}
