package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func RunInteractiveMode(ms *MySQLServer) {
	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		displayMenu()

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		if choice == "7" {
			fmt.Println("Exiting...")
			return
		}

		handleUserChoice(ctx, choice, ms, scanner)
	}
}

func displayMenu() {
	fmt.Println("\n=== MySQL MCP Tools ===")
	fmt.Println("1. List Schemas")
	fmt.Println("2. List Tables")
	fmt.Println("3. Get Table Structure")
	fmt.Println("4. Get Table CREATE Statement")
	fmt.Println("5. Execute Query")
	fmt.Println("6. Search in Table")
	fmt.Println("7. Exit")
	fmt.Print("\nSelect a tool (1-7): ")
}

func handleUserChoice(ctx context.Context, choice string, ms *MySQLServer, scanner *bufio.Scanner) {
	switch choice {
	case "1":
		// List schemas
		fmt.Print("Enter page number (default 1): ")
		scanner.Scan()
		pageStr := strings.TrimSpace(scanner.Text())
		page := 1
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_schemas",
				Arguments: map[string]interface{}{
					"page": page,
				},
			},
		}

		result, err := ms.listSchemasHandler(ctx, req)
		printResult(result, err)

	case "2":
		// List tables
		fmt.Print("Enter schema name: ")
		scanner.Scan()
		schema := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter page number (default 1): ")
		scanner.Scan()
		pageStr := strings.TrimSpace(scanner.Text())
		page := 1
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil {
				page = p
			}
		}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_tables",
				Arguments: map[string]interface{}{
					"schema": schema,
					"page":   page,
				},
			},
		}

		result, err := ms.listTablesHandler(ctx, req)
		printResult(result, err)

	case "3":
		// Get table structure
		fmt.Print("Enter schema name: ")
		scanner.Scan()
		schema := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter table name: ")
		scanner.Scan()
		table := strings.TrimSpace(scanner.Text())

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_table_structure",
				Arguments: map[string]interface{}{
					"schema": schema,
					"table":  table,
				},
			},
		}

		result, err := ms.getTableStructureHandler(ctx, req)
		printResult(result, err)

	case "4":
		// Get table CREATE statement
		fmt.Print("Enter schema name: ")
		scanner.Scan()
		schema := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter table name: ")
		scanner.Scan()
		table := strings.TrimSpace(scanner.Text())

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_table_create",
				Arguments: map[string]interface{}{
					"schema": schema,
					"table":  table,
				},
			},
		}

		result, err := ms.getTableCreateHandler(ctx, req)
		printResult(result, err)

	case "5":
		// Execute query
		fmt.Print("Enter SQL query (SELECT/SHOW/DESCRIBE/EXPLAIN only): ")
		scanner.Scan()
		query := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter result limit (default 100): ")
		scanner.Scan()
		limitStr := strings.TrimSpace(scanner.Text())
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "execute_query",
				Arguments: map[string]interface{}{
					"query": query,
					"limit": limit,
				},
			},
		}

		result, err := ms.executeQueryHandler(ctx, req)
		printResult(result, err)

	case "6":
		// Search in table
		fmt.Print("Enter schema name: ")
		scanner.Scan()
		schema := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter table name: ")
		scanner.Scan()
		table := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter search term: ")
		scanner.Scan()
		searchTerm := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter result limit (default 100): ")
		scanner.Scan()
		limitStr := strings.TrimSpace(scanner.Text())
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_table",
				Arguments: map[string]interface{}{
					"schema":      schema,
					"table":       table,
					"search_term": searchTerm,
					"limit":       limit,
				},
			},
		}

		result, err := ms.searchTableHandler(ctx, req)
		printResult(result, err)

	default:
		fmt.Println("Invalid choice. Please select 1-7.")
	}
}

func printResult(result *mcp.CallToolResult, err error) {
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	if result == nil {
		fmt.Println("\nNo result returned")
		return
	}

	for _, content := range result.Content {
		if textContent, ok := mcp.AsTextContent(content); ok {
			// Try to pretty print JSON
			var data interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &data); err == nil {
				prettyJSON, _ := json.MarshalIndent(data, "", "  ")
				fmt.Printf("\n%s\n", string(prettyJSON))
			} else {
				fmt.Printf("\n%s\n", textContent.Text)
			}
		} else {
			fmt.Printf("\nUnknown content type\n")
		}
	}
}