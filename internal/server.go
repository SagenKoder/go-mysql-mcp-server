package internal

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type MySQLServer struct {
	db *sql.DB
}

func NewMySQLServer() (*MySQLServer, error) {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLServer{db: db}, nil
}

func (ms *MySQLServer) Close() error {
	if ms.db != nil {
		return ms.db.Close()
	}
	return nil
}

// CreateMCPServerWithTools creates an MCP server instance with all tools registered
func CreateMCPServerWithTools(ms *MySQLServer) *server.MCPServer {
	s := server.NewMCPServer(
		"MySQL MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// List schemas tool
	listSchemasTool := mcp.NewTool("list_schemas",
		mcp.WithDescription("List all schemas/databases available in the MySQL server with pagination"),
		mcp.WithNumber("page",
			mcp.Description("Page number (1-based)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of items per page (default: 20, max: 100)"),
		),
	)
	s.AddTool(listSchemasTool, ms.listSchemasHandler)

	// List tables tool
	listTablesTool := mcp.NewTool("list_tables",
		mcp.WithDescription("List all tables in a specific schema with pagination"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("The schema/database name"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (1-based)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of items per page (default: 20, max: 100)"),
		),
	)
	s.AddTool(listTablesTool, ms.listTablesHandler)

	// Get table create statement tool
	getTableCreateTool := mcp.NewTool("get_table_create",
		mcp.WithDescription("Get the CREATE TABLE statement for a specific table"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("The schema/database name"),
		),
		mcp.WithString("table",
			mcp.Required(),
			mcp.Description("The table name"),
		),
	)
	s.AddTool(getTableCreateTool, ms.getTableCreateHandler)

	// Execute query tool
	executeQueryTool := mcp.NewTool("execute_query",
		mcp.WithDescription("Execute a SQL query (SELECT only for safety)"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute (SELECT statements only)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rows to return (default: 100)"),
		),
	)
	s.AddTool(executeQueryTool, ms.executeQueryHandler)

	// Search in table tool
	searchTableTool := mcp.NewTool("search_table",
		mcp.WithDescription("Search for a value across all columns in a table"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("The schema/database name"),
		),
		mcp.WithString("table",
			mcp.Required(),
			mcp.Description("The table name"),
		),
		mcp.WithString("search_term",
			mcp.Required(),
			mcp.Description("The term to search for"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rows to return (default: 100)"),
		),
	)
	s.AddTool(searchTableTool, ms.searchTableHandler)

	// Get table structure tool
	getTableStructureTool := mcp.NewTool("get_table_structure",
		mcp.WithDescription("Get the structure (columns, types, constraints) of a table"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("The schema/database name"),
		),
		mcp.WithString("table",
			mcp.Required(),
			mcp.Description("The table name"),
		),
	)
	s.AddTool(getTableStructureTool, ms.getTableStructureHandler)

	return s
}