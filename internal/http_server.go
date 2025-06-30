package internal

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

// StartHTTPServer starts the HTTP server using the built-in StreamableHTTPServer
func StartHTTPServer(ms *MySQLServer, addr string) error {
	// Create MCP server instance with the same tools as stdio
	mcpServer := CreateMCPServerWithTools(ms)

	// Create HTTP server with StreamableHTTPServer
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	log.Printf("MySQL MCP HTTP server starting on %s", addr)

	return httpServer.Start(addr)
}