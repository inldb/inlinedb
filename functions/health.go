package functions

import dbclass "github.com/MultiX0/db-test/db"

func HealthCheck() (bool, error) {
	db := dbclass.DB
	var health string
	err := db.QueryRow("PRAGMA integrity_check;").Scan(&health)
	if err != nil {
		return false, err
	}

	if health == "ok" {
		return true, nil
	}

	return false, nil
}
