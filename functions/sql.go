package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dbclass "github.com/MultiX0/db-test/db"
)

// TYPE WHEN WE RETURN SOMETHING
// SELECT * FROM <TABLE_NAME>;

// TYPE WHEN WE RETURN NOTHING
// INSERT INTO <TABLE_NAME>(ID, NAME) VALUES(<VAL_1>,<VAL_2>);

func RawSQL(sqlStmt string) (any, error) {

	db := dbclass.DB

	querySlice := strings.Split(sqlStmt, " ")
	queryKeyword := strings.ToLower(strings.TrimSpace(querySlice[0]))
	fmt.Println(queryKeyword)

	if queryKeyword == "select" {
		data, err := QueryAsJson(sqlStmt)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	successResult := map[string]any{"message": "success"}
	return successResult, nil

}

func QueryAsJson(sqlStmt string) ([]byte, error) {

	rows, err := dbclass.DB.Query(sqlStmt)
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		return nil, err
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {

		scanArgs := make([]interface{}, count)

		for i, v := range columnTypes {

			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
			case "NUMERIC", "INTEGER", "REAL":
				scanArgs[i] = new(sql.NullInt64)
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			return nil, err
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}

	marshalData, err := json.Marshal(finalRows)

	if err != nil {
		return nil, err
	}

	return marshalData, err

}
