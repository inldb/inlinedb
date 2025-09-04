package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/MultiX0/db-test/constants"
	dbclass "github.com/MultiX0/db-test/db"
	"github.com/MultiX0/db-test/models"
	"github.com/google/uuid"
)

func GetAllTables() (*models.TablesModel, error) {

	var tables models.TablesModel

	sqlStmt := "SELECT name FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	rows, err := dbclass.DB.Query(sqlStmt)
	if err != nil {
		fmt.Println("Here 1")
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)

		if err != nil {
			fmt.Println("Here 2")
			return nil, err
		}

		table, err := GetTableData(name)
		if err != nil {
			fmt.Println("Here 3")

			return nil, err
		}

		tables.Tables = append(tables.Tables, *table)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &tables, nil

}

func GetTableData(tableName string) (*models.TableModel, error) {
	if len(tableName) == 0 {
		return nil, fmt.Errorf("you need to enter the table name to get the info")
	}

	stmt, err := dbclass.DB.Prepare("SELECT sql FROM sqlite_schema WHERE name = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// sql.NullString to handle NULL values
	var sqlValue sql.NullString
	err = stmt.QueryRow(tableName).Scan(&sqlValue)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf("this table does not exist")
		}
		return nil, err
	}

	// Convert to regular string ,if NULL, use empty string
	var sqlString string
	if sqlValue.Valid {
		sqlString = sqlValue.String
	} else {
		sqlString = ""
	}

	columns, err := GetTableColumns(tableName)
	if err != nil {
		return nil, err
	}

	recordsCount, err := GetTableCount(tableName)
	if err != nil {
		return nil, err
	}

	return &models.TableModel{
		Name:         tableName,
		Sql:          sqlString,
		Columns:      *columns,
		RecordsCount: *recordsCount,
	}, nil
}

func GetTableColumns(tableName string) (*[]models.ColumnModel, error) {

	if len(tableName) == 0 {
		return nil, fmt.Errorf("you need to enter the table name to get the info")
	}

	var columns []models.ColumnModel
	sqlStmt := fmt.Sprintf("SELECT name, type, pk, \"notnull\", dflt_value FROM pragma_table_info('%s')", tableName)
	rows, err := dbclass.DB.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var _type string
		var pk int
		var notnull int
		var default_value *string

		err = rows.Scan(&name, &_type, &pk, &notnull, &default_value)
		if err != nil {
			return nil, err
		}

		// notnull = 1 means NOT NULL (nullable = false)
		// notnull = 0 means NULL allowed (nullable = true)
		columns = append(columns, models.ColumnModel{
			Name:          name,
			DataType:      _type,
			IsPrimaryKey:  (pk == 1),
			Nullable:      (notnull == 0), // notnull=0 means nullable=true
			Default_Value: default_value,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &columns, nil
}

func GetTableCount(tableName string) (*int, error) {
	if len(tableName) == 0 {
		return nil, fmt.Errorf("you need to enter the table name to get the info")
	}

	var exists int
	err := dbclass.DB.QueryRow("SELECT COUNT(*) FROM sqlite_schema WHERE name = ? AND type='table'", tableName).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("table does not exist")
	}

	// vulnerable to SQL injection edit this later on...
	sqlStmt := "SELECT COUNT(*) FROM " + tableName
	var count int
	err = dbclass.DB.QueryRow(sqlStmt).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

func GetNumberOfTables() (*int, error) {
	sqlStmt := "SELECT COUNT(name) FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	var count int
	err := dbclass.DB.QueryRow(sqlStmt).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// create table test(id integer not null primary key, name text)

// on the dashboard make new datatype called uuid, this datatype is not part of sqlite by default but here is what we gonna do
// make the column datatype to text
// make it notnull and pk
// add constraints that ensure the column length should be exact 36 character (uuid-length)
// for any (insert - update - upsert) check the length of the id manually if it is implemented on the query , check if it is valid uuid or not
// if the user dose not implement any id value that is fine just make it default uuid.V4

// Fixed CreateTable function with TEXT primary key and explicit index creation
func CreateTable(table models.TableModel) error {
	if table.Name == "" || len(table.Columns) == 0 {
		return fmt.Errorf("table name and columns are required")
	}

	// Start a transaction for atomic operations
	tx, err := dbclass.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var columns []string
	var primaryKeys []string
	var indexesToCreate []string

	for _, column := range table.Columns {
		var parts []string

		// Get the SQLite data type from constants
		sqlType := constants.DataTypes[column.DataType]

		parts = append(parts, column.Name, sqlType)

		if column.Default_Value != nil && len(strings.TrimSpace(*column.Default_Value)) != 0 {
			parts = append(parts, "DEFAULT", string(*column.Default_Value))

		}

		if !column.Nullable {
			parts = append(parts, "NOT NULL")
		}

		// Handle primary key
		if column.IsPrimaryKey {
			primaryKeys = append(primaryKeys, column.Name)

			// Create explicit index for TEXT primary keys to avoid auto-index issues
			// SQLite data types are case-insensitive, so check for common TEXT variations
			sqlTypeUpper := strings.ToUpper(sqlType)
			if sqlTypeUpper == "TEXT" || sqlTypeUpper == "VARCHAR" || sqlTypeUpper == "CHAR" {
				indexName := fmt.Sprintf("idx_%s_%s", table.Name, column.Name)
				indexSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
					indexName, table.Name, column.Name)
				indexesToCreate = append(indexesToCreate, indexSQL)
			}
		}

		columns = append(columns, strings.Join(parts, " "))
	}

	// Add primary key constraint
	var sqlStmt string
	if len(primaryKeys) > 0 {
		pkConstraint := fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
		sqlStmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s, %s)",
			table.Name,
			strings.Join(columns, ", "),
			pkConstraint)
	} else {
		sqlStmt = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
			table.Name,
			strings.Join(columns, ", "))
	}

	fmt.Printf("Executing SQL: %s\n", sqlStmt)

	// Create the table
	_, err = tx.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Create explicit indexes for primary key columns
	for _, indexSQL := range indexesToCreate {
		fmt.Printf("Creating index: %s\n", indexSQL)
		_, err = tx.Exec(indexSQL)
		if err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func InsertIntoTable(insertModel models.InsertModel) (string, error) {
	for _, column := range insertModel.Columns {
		if strings.TrimSpace(column) == "id" {
			return "", fmt.Errorf("insert request should not contains the id, id is auto generated by the system and will be returned in the response")
		}
	}

	id := uuid.New()

	// Build the SQL with placeholders
	placeholders := make([]string, len(insertModel.Values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sqlStmt := fmt.Sprintf("INSERT INTO %s (id, %s) VALUES (?, %s)",
		insertModel.TableName,
		strings.Join(insertModel.Columns, ", "),
		strings.Join(placeholders, ", "))

	// Prepare the statement
	stmt, err := dbclass.DB.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Prepare the arguments slice
	args := make([]interface{}, len(insertModel.Values)+1)
	args[0] = id.String() // First argument is the ID (UUID-V4)
	for i, value := range insertModel.Values {
		args[i+1] = value
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute statement: %v", err)
	}

	return id.String(), nil
}

// ValidateColumnName checks if a column name is safe (no SQL injection)
func ValidateColumnName(columnName string) error {
	columnName = strings.TrimSpace(columnName)

	if columnName == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	if columnName == "*" {
		return nil
	}

	validIdentifier := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validIdentifier.MatchString(columnName) {
		return fmt.Errorf("invalid column name format: %s", columnName)
	}

	dangerousKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter",
		"union", "exec", "execute", "script", "declare", "cast", "convert",
		"information_schema", "sys", "master", "msdb", "tempdb",
	}

	lowerColumn := strings.ToLower(columnName)
	for _, keyword := range dangerousKeywords {
		if lowerColumn == keyword {
			return fmt.Errorf("column name contains dangerous keyword: %s", columnName)
		}
	}

	injectionPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_", "@@",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerColumn, pattern) {
			return fmt.Errorf("column name contains potentially dangerous characters: %s", columnName)
		}
	}

	return nil
}

func ValidateColumns(tableName string, columns []string) error {

	if len(columns) == 1 && columns[0] == "*" {
		return nil
	}

	columnsPtr, err := GetTableColumns(tableName)
	if err != nil {
		return fmt.Errorf("failed to get table columns: %w", err)
	}

	if columnsPtr == nil {
		return fmt.Errorf("no column information returned for table: %s", tableName)
	}

	actualColumnSet := make(map[string]bool)
	for _, col := range *columnsPtr {
		actualColumnSet[col.Name] = true
	}

	for _, requestedCol := range columns {
		if err := ValidateColumnName(requestedCol); err != nil {
			return err
		}

		if !actualColumnSet[requestedCol] {
			return fmt.Errorf("column '%s' does not exist in table '%s'", requestedCol, tableName)
		}
	}

	return nil
}

func ValidateOperator(operator string) error {
	validOps := map[string]bool{
		"eq": true, "ne": true, "gt": true, "lt": true,
		"gte": true, "lte": true, "like": true, "in": true, "not_in": true,
		"is_null": true, "is_not_null": true,
	}

	if !validOps[operator] {
		return fmt.Errorf("invalid operator: %s", operator)
	}
	return nil
}

func BuildWhereClause(tableName string, filters []models.FilterGroup) (string, []any, error) {
	if len(filters) == 0 {
		return "", []any{}, nil
	}

	columnsPtr, err := GetTableColumns(tableName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get table columns: %w", err)
	}

	actualColumnSet := make(map[string]bool)
	for _, col := range *columnsPtr {
		actualColumnSet[col.Name] = true
	}

	var whereParts []string
	var params []any

	for _, group := range filters {
		if len(group.Conditions) == 0 {
			continue
		}

		logic := strings.ToUpper(group.Logic)
		if logic != "AND" && logic != "OR" {
			logic = "AND"
		}

		var conditionParts []string

		for _, condition := range group.Conditions {
			if err := ValidateColumnName(condition.Column); err != nil {
				return "", nil, err
			}

			if !actualColumnSet[condition.Column] {
				return "", nil, fmt.Errorf("filter column '%s' does not exist in table '%s'", condition.Column, tableName)
			}

			if err := ValidateOperator(condition.Operator); err != nil {
				return "", nil, err
			}

			conditionSQL, conditionParams, err := buildCondition(condition)
			if err != nil {
				return "", nil, err
			}

			conditionParts = append(conditionParts, conditionSQL)
			params = append(params, conditionParams...)
		}

		if len(conditionParts) > 0 {
			groupClause := "(" + strings.Join(conditionParts, " "+logic+" ") + ")"
			whereParts = append(whereParts, groupClause)
		}
	}

	whereClause := strings.Join(whereParts, " AND ")
	return whereClause, params, nil
}

func buildCondition(condition models.FilterCondition) (string, []any, error) {
	column := condition.Column
	operator := condition.Operator
	value := condition.Value

	switch operator {
	case "eq":
		return fmt.Sprintf("%s = ?", column), []any{value}, nil
	case "ne":
		return fmt.Sprintf("%s != ?", column), []any{value}, nil
	case "gt":
		return fmt.Sprintf("%s > ?", column), []any{value}, nil
	case "lt":
		return fmt.Sprintf("%s < ?", column), []any{value}, nil
	case "gte":
		return fmt.Sprintf("%s >= ?", column), []any{value}, nil
	case "lte":
		return fmt.Sprintf("%s <= ?", column), []any{value}, nil
	case "like":
		return fmt.Sprintf("%s LIKE ?", column), []any{value}, nil
	case "is_null":
		return fmt.Sprintf("%s IS NULL", column), []any{}, nil
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", column), []any{}, nil
	case "in":
		values, ok := value.([]any)
		if !ok {
			return "", nil, fmt.Errorf("IN operator requires array of values")
		}
		if len(values) == 0 {
			return "", nil, fmt.Errorf("IN operator requires at least one value")
		}

		placeholders := strings.Repeat("?,", len(values))
		placeholders = placeholders[:len(placeholders)-1]

		return fmt.Sprintf("%s IN (%s)", column, placeholders), values, nil
	case "not_in":
		values, ok := value.([]any)
		if !ok {
			return "", nil, fmt.Errorf("NOT IN operator requires array of values")
		}
		if len(values) == 0 {
			return "", nil, fmt.Errorf("NOT IN operator requires at least one value")
		}

		placeholders := strings.Repeat("?,", len(values))
		placeholders = placeholders[:len(placeholders)-1]

		return fmt.Sprintf("%s NOT IN (%s)", column, placeholders), values, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", operator)
	}
}

func BuildSelectQuery(selectModel models.SelectModel) (string, []any, error) {
	if len(selectModel.SelectedColumns) == 0 {
		return "", nil, fmt.Errorf("no columns specified")
	}

	if err := ValidateColumns(selectModel.TableName, selectModel.SelectedColumns); err != nil {
		return "", nil, err
	}

	whereClause, params, err := BuildWhereClause(selectModel.TableName, selectModel.Filters)
	if err != nil {
		return "", nil, err
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectModel.SelectedColumns, ", "), selectModel.TableName)

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query, params, nil
}

func SelectFromTable(selectModel models.SelectModel) ([]byte, error) {
	if len(strings.TrimSpace(selectModel.TableName)) == 0 {
		return nil, fmt.Errorf("you should enter the table name first to select")
	}

	query, params, err := BuildSelectQuery(selectModel)
	if err != nil {
		return nil, err
	}

	stmt, err := dbclass.DB.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	jsonResult, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	return jsonResult, nil
}
