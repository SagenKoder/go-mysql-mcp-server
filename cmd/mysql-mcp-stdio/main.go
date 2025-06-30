package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"go_mysql_mcp/internal"
)

func main() {
	// Create MySQL server
	ms, err := internal.NewMySQLServer()
	if err != nil {
		log.Fatalf("Failed to create MySQL server: %v", err)
	}
	defer ms.Close()

	// Create and run MCP server
	s := internal.CreateMCPServerWithTools(ms)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}