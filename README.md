# MySQL MCP Server

A Model Context Protocol (MCP) server that provides AI agents with direct access to MySQL databases. This server enables AI models to query, search, and analyze MySQL database content through a set of well-defined tools.

## Features

- **Schema exploration** - List all schemas/databases with pagination
- **Table discovery** - Browse tables in any schema with detailed metadata
- **Structure inspection** - Get column definitions, types, and indexes
- **Safe querying** - Execute SELECT queries with automatic result limiting
- **Full-text search** - Search for values across all text columns in a table
- **DDL retrieval** - Get CREATE TABLE statements for any table

## Quick Start with Docker

### Running with Docker (Recommended)

The easiest way to run the MySQL MCP server is using Docker:

```bash
# Pull the latest image (stdio mode by default)
docker pull ghcr.io/sagenkoder/go-mysql-mcp-server:latest

# Or pull a specific mode
docker pull ghcr.io/sagenkoder/go-mysql-mcp-server:stdio
docker pull ghcr.io/sagenkoder/go-mysql-mcp-server:http
docker pull ghcr.io/sagenkoder/go-mysql-mcp-server:interactive
```

### Docker Usage Examples

#### Stdio mode (for Claude Desktop)
```bash
# Connect to MySQL on host machine
docker run -i --rm \
  --network host \
  -e MYSQL_HOST=localhost \
  -e MYSQL_USER=your_user \
  -e MYSQL_PASSWORD=your_password \
  -e MYSQL_DATABASE=your_database \
  ghcr.io/sagenkoder/go-mysql-mcp-server:stdio
```

#### HTTP server mode
```bash
# Run HTTP server on port 8080
docker run -d \
  --name mysql-mcp-http \
  --network host \
  -p 8080:8080 \
  -e MYSQL_HOST=localhost \
  -e MYSQL_USER=your_user \
  -e MYSQL_PASSWORD=your_password \
  ghcr.io/sagenkoder/go-mysql-mcp-server:http
```

#### Interactive mode (for testing)
```bash
# Run in interactive mode
docker run -it --rm \
  --network host \
  -e MYSQL_HOST=localhost \
  -e MYSQL_USER=your_user \
  -e MYSQL_PASSWORD=your_password \
  ghcr.io/sagenkoder/go-mysql-mcp-server:interactive
```

### Connecting to MySQL in Docker

If your MySQL is also running in Docker, use Docker networking:

```bash
# Create a network
docker network create myapp

# Run MySQL (example)
docker run -d \
  --name mysql \
  --network myapp \
  -e MYSQL_ROOT_PASSWORD=rootpass \
  -e MYSQL_DATABASE=mydb \
  mysql:8

# Run MCP server
docker run -i --rm \
  --network myapp \
  -e MYSQL_HOST=mysql \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=rootpass \
  -e MYSQL_DATABASE=mydb \
  ghcr.io/sagenkoder/go-mysql-mcp-server:stdio
```

## Claude Desktop Configuration

### Using Docker with Claude Desktop

Add this to your Claude Desktop configuration file:

```json
{
  "mcpServers": {
    "mysql": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--network", "host",
        "-e", "MYSQL_HOST=localhost",
        "-e", "MYSQL_USER=your_user",
        "-e", "MYSQL_PASSWORD=your_password",
        "-e", "MYSQL_DATABASE=your_database",
        "ghcr.io/sagenkoder/go-mysql-mcp-server:stdio"
      ]
    }
  }
}
```

### Using Binary with Claude Desktop

If you prefer to use the binary directly:

```json
{
  "mcpServers": {
    "mysql": {
      "command": "/path/to/mysql-mcp-stdio",
      "env": {
        "MYSQL_HOST": "localhost",
        "MYSQL_PORT": "3306",
        "MYSQL_USER": "your_user",
        "MYSQL_PASSWORD": "your_password",
        "MYSQL_DATABASE": "your_database"
      }
    }
  }
}
```

## Building from Source

### Prerequisites
- Go 1.23 or later
- Docker (optional, for building Docker images)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/sagenkoder/go-mysql-mcp-server.git
cd go-mysql-mcp-server

# Build all binaries
./build.sh

# Build Docker images
./build.sh docker
```

This will create:
- Binaries:
  - `mysql-mcp-stdio` - For stdio-based MCP communication
  - `mysql-mcp-http` - HTTP server mode
  - `mysql-mcp-interactive` - Interactive CLI mode for testing
- Docker images:
  - `ghcr.io/sagenkoder/go-mysql-mcp-server:stdio` (also tagged as `mysql-mcp:latest`)
  - `ghcr.io/sagenkoder/go-mysql-mcp-server:http`
  - `ghcr.io/sagenkoder/go-mysql-mcp-server:interactive`

## Configuration

The server connects to MySQL using these environment variables:

- `MYSQL_HOST` - MySQL server hostname (default: localhost)
- `MYSQL_PORT` - MySQL server port (default: 3306) 
- `MYSQL_USER` - MySQL username (default: root)
- `MYSQL_PASSWORD` - MySQL password (required)
- `MYSQL_DATABASE` - Default database (optional)

## Available Tools

### list_schemas
List all schemas/databases available in the MySQL server.

Parameters:
- `page` (optional): Page number for pagination (default: 1)
- `page_size` (optional): Number of items per page (default: 20, max: 100)

### list_tables
List all tables in a specific schema with metadata.

Parameters:
- `schema` (required): The schema/database name
- `page` (optional): Page number for pagination
- `page_size` (optional): Number of items per page

### get_table_structure
Get detailed column and index information for a table.

Parameters:
- `schema` (required): The schema/database name
- `table` (required): The table name

### get_table_create
Get the CREATE TABLE statement for a specific table.

Parameters:
- `schema` (required): The schema/database name
- `table` (required): The table name

### execute_query
Execute a SQL query (SELECT, SHOW, DESCRIBE, EXPLAIN only).

Parameters:
- `query` (required): The SQL query to execute
- `limit` (optional): Maximum rows to return (default: 100)

### search_table
Search for a value across all text columns in a table.

Parameters:
- `schema` (required): The schema/database name
- `table` (required): The table name
- `search_term` (required): The term to search for
- `limit` (optional): Maximum rows to return (default: 100)

## Testing the Connection

Use the interactive mode to test your connection:

```bash
# With Docker
docker run -it --rm \
  --network host \
  -e MYSQL_HOST=localhost \
  -e MYSQL_USER=test \
  -e MYSQL_PASSWORD=test \
  ghcr.io/sagenkoder/go-mysql-mcp-server:interactive

# With binary
MYSQL_USER=test MYSQL_PASSWORD=test ./mysql-mcp-interactive
```

## Security Considerations

- Only SELECT, SHOW, DESCRIBE, and EXPLAIN queries are allowed
- All queries are automatically limited to prevent large result sets
- Table searches only scan text-based columns
- Connection details should be stored securely as environment variables
- The Docker image runs as a non-root user for security

## Troubleshooting

### Connection Issues
- Ensure MySQL is running and accessible
- Check that the MySQL user has appropriate permissions
- When using Docker, verify network connectivity (`--network host` for local MySQL)
- Test with mysql CLI client first: `mysql -h localhost -u user -p`

### Docker Network Issues
- Use `--network host` to connect to MySQL on the host machine
- For MySQL in Docker, create a shared network and use container names as hostnames
- Check firewall rules if connecting to remote MySQL

## License

MIT License - see LICENSE file for details.