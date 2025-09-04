package dbclass

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB      *sql.DB
	AdminDB *sql.DB
)

func InitDB() error {
	db, err := sql.Open("sqlite3", "./inline.db")
	if err != nil {
		return err
	}

	admin, err := sql.Open("sqlite3", "./admin.db")
	if err != nil {
		return err
	}

	DB = db
	AdminDB = admin

	err = SetupAdminSchema()
	if err != nil {
		return err
	}

	return nil
}
