package front

import (
	"context"
	"encoding/json"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/adaptermap"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RunHost serves domain MCP tools on stdio.
func RunHost(ctx context.Context, svc *adaptermap.Service, dec *shadow.Decider) error {
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

	s.AddTool(mcp.NewTool("decide_shadow",
		mcp.WithDescription("Record a shadow label beside the caller-supplied actual. Follow actual. Effect is none. A missing key, transport error, non-2xx, or truncated window is not a no_go."),
		mcp.WithString("path", mcp.Required()),
		mcp.WithString("state_body", mcp.Required()),
		mcp.WithString("actual", mcp.Required()),
		mcp.WithString("action"),
		mcp.WithArray("labels", mcp.WithStringItems()),
		mcp.WithString("request_id"),
	), decideShadowHandler(dec))

	return server.ServeStdio(s)
}

func decideShadowHandler(dec *shadow.Decider) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if dec == nil {
			return mcp.NewToolResultError("laya decider not set"), nil
		}
		args := req.GetArguments()
		path, _ := args["path"].(string)
		body, _ := args["state_body"].(string)
		actual, _ := args["actual"].(string)
		action, _ := args["action"].(string)
		requestID, _ := args["request_id"].(string)
		var labels []string
		if raw, ok := args["labels"].([]interface{}); ok {
			for _, item := range raw {
				if s, ok := item.(string); ok {
					labels = append(labels, s)
				}
			}
		}
		out, err := dec.Decide(ctx, shadow.Input{
			Path:      path,
			StateBody: body,
			Actual:    actual,
			Action:    action,
			Labels:    labels,
			RequestID: requestID,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, err := json.Marshal(out)
		if err != nil {
			return mcp.NewToolResultError("laya response unreadable"), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	}
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
