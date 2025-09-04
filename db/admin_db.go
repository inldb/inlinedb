package dbclass

import (
	"fmt"

	"github.com/MultiX0/db-test/models"
)

func SetupAdminSchema() error {
	err := CreateTableSchema()
	if err != nil {
		return fmt.Errorf("%s", "create table schema failed: "+err.Error())
	}

	err = CreateColumnsSchema()
	if err != nil {
		return fmt.Errorf("%s", "create columns schema failed: "+err.Error())
	}

	return nil

}

func CreateTableSchema() error {
	sqlstmt := "CREATE TABLE IF NOT EXISTS tables ( name TEXT PRIMARY KEY, description TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP );"
	_, err := AdminDB.Exec(sqlstmt)
	return err
}
func CreateColumnsSchema() error {
	sqlstmt := "CREATE TABLE IF NOT EXISTS columns ( id INTEGER PRIMARY KEY AUTOINCREMENT, table_name TEXT NOT NULL, is_pk BOOLEAN NOT NULL DEFAULT 0, null_able BOOLEAN NOT NULL DEFAULT 1, date_type TEXT NOT NULL, original_type TEXT NOT NULL, FOREIGN KEY (table_name) REFERENCES tables(name) ON DELETE CASCADE ON UPDATE CASCADE );"
	_, err := AdminDB.Exec(sqlstmt)
	return err

}

func InsertTable(table models.TableModel) {

}

func InsertColumns(columns []models.ColumnModel) {}
