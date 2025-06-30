package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (ms *MySQLServer) listSchemasHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	page := getIntFromArgs(args, "page", 1)
	pageSize := getIntFromArgs(args, "page_size", DefaultPageSize)

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM information_schema.SCHEMATA"
	if err := ms.db.QueryRow(countQuery).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get schema count: %w", err)
	}

	// Get schemas with pagination
	query := "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA ORDER BY SCHEMA_NAME LIMIT ? OFFSET ?"
	rows, err := ms.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	result := map[string]interface{}{
		"schemas":     schemas,
		"page":        page,
		"page_size":   pageSize,
		"total_count": totalCount,
		"total_pages": (totalCount + pageSize - 1) / pageSize,
	}

	return jsonResult(result)
}

func (ms *MySQLServer) listTablesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	schema, ok := args["schema"].(string)
	if !ok || schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}

	page := getIntFromArgs(args, "page", 1)
	pageSize := getIntFromArgs(args, "page_size", DefaultPageSize)

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?"
	if err := ms.db.QueryRow(countQuery, schema).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	// Get tables with pagination
	query := `
		SELECT TABLE_NAME, TABLE_TYPE, ENGINE, TABLE_ROWS, 
		       DATA_LENGTH, INDEX_LENGTH, CREATE_TIME, UPDATE_TIME
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME 
		LIMIT ? OFFSET ?
	`
	rows, err := ms.db.Query(query, schema, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []map[string]interface{}
	for rows.Next() {
		var tableName, tableType string
		var engine, createTime, updateTime sql.NullString
		var tableRows, dataLength, indexLength sql.NullInt64

		if err := rows.Scan(&tableName, &tableType, &engine, &tableRows, 
			&dataLength, &indexLength, &createTime, &updateTime); err != nil {
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}

		table := map[string]interface{}{
			"name": tableName,
			"type": tableType,
		}

		if engine.Valid {
			table["engine"] = engine.String
		}
		if tableRows.Valid {
			table["rows"] = tableRows.Int64
		}
		if dataLength.Valid {
			table["data_size"] = dataLength.Int64
		}
		if indexLength.Valid {
			table["index_size"] = indexLength.Int64
		}
		if createTime.Valid {
			table["created_at"] = createTime.String
		}
		if updateTime.Valid {
			table["updated_at"] = updateTime.String
		}

		tables = append(tables, table)
	}

	result := map[string]interface{}{
		"schema":      schema,
		"tables":      tables,
		"page":        page,
		"page_size":   pageSize,
		"total_count": totalCount,
		"total_pages": (totalCount + pageSize - 1) / pageSize,
	}

	return jsonResult(result)
}

func (ms *MySQLServer) getTableCreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	schema, ok := args["schema"].(string)
	if !ok || schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}

	table, ok := args["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table parameter is required")
	}

	// Get CREATE TABLE statement
	var tableName, createStmt string
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", schema, table)
	err := ms.db.QueryRow(query).Scan(&tableName, &createStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get create statement: %w", err)
	}

	result := map[string]interface{}{
		"schema":          schema,
		"table":           table,
		"create_statement": createStmt,
	}

	return jsonResult(result)
}

func (ms *MySQLServer) executeQueryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Basic safety check - only allow SELECT statements
	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") && !strings.HasPrefix(trimmedQuery, "SHOW") && 
	   !strings.HasPrefix(trimmedQuery, "DESCRIBE") && !strings.HasPrefix(trimmedQuery, "EXPLAIN") {
		return nil, fmt.Errorf("only SELECT, SHOW, DESCRIBE, and EXPLAIN statements are allowed")
	}

	limit := getIntFromArgs(args, "limit", 100)

	// Add LIMIT if not present and it's a SELECT query
	if strings.HasPrefix(trimmedQuery, "SELECT") && !strings.Contains(trimmedQuery, "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	rows, err := ms.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result
	var results []map[string]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	result := map[string]interface{}{
		"columns": columns,
		"rows":    results,
		"count":   len(results),
	}

	return jsonResult(result)
}

func (ms *MySQLServer) searchTableHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	schema, ok := args["schema"].(string)
	if !ok || schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}

	table, ok := args["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table parameter is required")
	}

	searchTerm, ok := args["search_term"].(string)
	if !ok || searchTerm == "" {
		return nil, fmt.Errorf("search_term parameter is required")
	}

	limit := getIntFromArgs(args, "limit", 100)

	// First, get all columns for the table
	colQuery := `
		SELECT COLUMN_NAME, DATA_TYPE 
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`
	colRows, err := ms.db.Query(colQuery, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer colRows.Close()

	var columns []string
	var searchableColumns []string
	for colRows.Next() {
		var colName, dataType string
		if err := colRows.Scan(&colName, &dataType); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, colName)
		
		// Only search in text-like columns
		if isSearchableType(dataType) {
			searchableColumns = append(searchableColumns, fmt.Sprintf("`%s` LIKE ?", colName))
		}
	}

	if len(searchableColumns) == 0 {
		return nil, fmt.Errorf("no searchable columns found in table")
	}

	// Build search query
	whereClause := strings.Join(searchableColumns, " OR ")
	searchQuery := fmt.Sprintf("SELECT * FROM `%s`.`%s` WHERE %s LIMIT %d", schema, table, whereClause, limit)

	// Prepare search parameters
	searchPattern := fmt.Sprintf("%%%s%%", searchTerm)
	params := make([]interface{}, len(searchableColumns))
	for i := range params {
		params[i] = searchPattern
	}

	rows, err := ms.db.Query(searchQuery, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to search table: %w", err)
	}
	defer rows.Close()

	// Prepare result
	var results []map[string]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	result := map[string]interface{}{
		"schema":      schema,
		"table":       table,
		"search_term": searchTerm,
		"columns":     columns,
		"rows":        results,
		"count":       len(results),
	}

	return jsonResult(result)
}

func (ms *MySQLServer) getTableStructureHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	schema, ok := args["schema"].(string)
	if !ok || schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}

	table, ok := args["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table parameter is required")
	}

	// Get column information
	colQuery := `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, 
		       COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`
	rows, err := ms.db.Query(colQuery, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []map[string]interface{}
	for rows.Next() {
		var colName, colType, isNullable, colKey, extra, comment string
		var colDefault sql.NullString

		if err := rows.Scan(&colName, &colType, &isNullable, &colKey, 
			&colDefault, &extra, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		column := map[string]interface{}{
			"name":     colName,
			"type":     colType,
			"nullable": isNullable == "YES",
			"key":      colKey,
			"extra":    extra,
		}

		if colDefault.Valid {
			column["default"] = colDefault.String
		}
		if comment != "" {
			column["comment"] = comment
		}

		columns = append(columns, column)
	}

	// Get indexes
	indexQuery := `
		SELECT INDEX_NAME, NON_UNIQUE, GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as COLUMNS
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		GROUP BY INDEX_NAME, NON_UNIQUE
	`
	indexRows, err := ms.db.Query(indexQuery, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	defer indexRows.Close()

	var indexes []map[string]interface{}
	for indexRows.Next() {
		var indexName, columns string
		var nonUnique int

		if err := indexRows.Scan(&indexName, &nonUnique, &columns); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		index := map[string]interface{}{
			"name":    indexName,
			"unique":  nonUnique == 0,
			"columns": strings.Split(columns, ","),
		}
		indexes = append(indexes, index)
	}

	result := map[string]interface{}{
		"schema":  schema,
		"table":   table,
		"columns": columns,
		"indexes": indexes,
	}

	return jsonResult(result)
}

// Helper functions
func getIntFromArgs(args map[string]any, key string, defaultValue int) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}
	return defaultValue
}

func jsonResult(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

func isSearchableType(dataType string) bool {
	searchableTypes := []string{
		"char", "varchar", "text", "tinytext", "mediumtext", "longtext",
		"enum", "set",
	}
	
	dataType = strings.ToLower(dataType)
	for _, t := range searchableTypes {
		if strings.Contains(dataType, t) {
			return true
		}
	}
	return false
}