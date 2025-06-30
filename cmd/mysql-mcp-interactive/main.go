package main

import (
	"log"

	"go_mysql_mcp/internal"
)

func main() {
	// Create MySQL server
	ms, err := internal.NewMySQLServer()
	if err != nil {
		log.Fatalf("Failed to create MySQL server: %v", err)
	}
	defer ms.Close()

	// Run interactive mode
	internal.RunInteractiveMode(ms)
}