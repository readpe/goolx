// cmd/mcp-server provides an MCP (Model Context Protocol) server for interacting
// with ASPEN Oneliner power system cases via goolx. It exposes read-only tools
// for querying equipment, running fault analysis, and inspecting results.
//
// Usage:
//
//	mcp-server
//
// The server communicates over stdio using the MCP protocol.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/readpe/goolx"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := goolx.NewClient()
	defer client.Release()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "goolx-mcp-server",
		Version: "0.1.0",
	}, nil)

	registerTools(server, client)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
