// Command ce-e2e drives the plugin MCP server over stdio, through the plugin
// launcher, and writes a pass/fail report (GMA-543). It needs ENGRAM_BIN,
// CONTEXT_EFFICIENCY_MCP_BIN and a throwaway ENGRAM_DATA_DIR. It never sets
// LAYA_API_KEY. Run it through scripts/e2e-local.sh.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	launcher := flag.String("launcher", "plugin/bin/run-context-efficiency-mcp.sh", "plugin launcher")
	out := flag.String("out", "reports", "report directory")
	commit := flag.String("commit", "unknown", "commit under test")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	os.Exit(run(ctx, *launcher, *out, *commit))
}

func run(ctx context.Context, launcher, out, commit string) int {
	c, err := client.NewStdioMCPClient("bash", nil, launcher)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ce-e2e: start:", err)
		return 1
	}
	defer c.Close()
	rep := &Report{Commit: commit}
	if err := initialize(ctx, c, rep); err != nil {
		fmt.Fprintln(os.Stderr, "ce-e2e: initialize:", err)
		return 1
	}
	for _, s := range Steps() {
		rep.Add(s.Name, s.Run(ctx, caller{c}))
	}
	if err := rep.Write(out); err != nil {
		fmt.Fprintln(os.Stderr, "ce-e2e: report:", err)
		return 1
	}
	if rep.Failed() > 0 {
		return 2
	}
	return 0
}

func initialize(ctx context.Context, c *client.Client, rep *Report) error {
	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: "ce-e2e", Version: "0.1.0"}
	res, err := c.Initialize(ctx, req)
	if err != nil {
		return err
	}
	rep.Server = res.ServerInfo.Name + " " + res.ServerInfo.Version
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return err
	}
	for _, t := range tools.Tools {
		rep.Tools = append(rep.Tools, t.Name)
	}
	return nil
}
