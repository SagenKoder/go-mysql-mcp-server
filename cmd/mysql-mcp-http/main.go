package main

import (
	"flag"
	"log"

	"go_mysql_mcp/internal"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "HTTP server address")
	flag.Parse()

	// Create MySQL server
	ms, err := internal.NewMySQLServer()
	if err != nil {
		log.Fatalf("Failed to create MySQL server: %v", err)
	}
	defer ms.Close()

	// Start HTTP server
	if err := internal.StartHTTPServer(ms, addr); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}