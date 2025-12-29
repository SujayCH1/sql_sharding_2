package database

import (
	"database/sql"
	"fmt"

	"sql-sharding-v2/internal/config"

	_ "github.com/lib/pq"
)

func EnsureApplicationDatabaseExists() error {
	// 1. Connect to system DB (postgres)
	sysConnStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		config.ApplicationDatabaseCredentials.DB_USER,
		config.ApplicationDatabaseCredentials.DB_PASS,
		config.ApplicationDatabaseCredentials.DB_HOST,
		config.ApplicationDatabaseCredentials.DB_PORT,
	)

	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		return err
	}
	defer sysDB.Close()

	// 2. Check if app DB exists
	var exists bool
	err = sysDB.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		config.ApplicationDatabaseCredentials.DB_NAME,
	).Scan(&exists)

	if err != nil {
		return err
	}

	// 3. Create DB if it does not exist
	if !exists {
		_, err = sysDB.Exec(
			fmt.Sprintf(`CREATE DATABASE "%s"`, config.ApplicationDatabaseCredentials.DB_NAME),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
