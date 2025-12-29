package database

import (
	"database/sql"
	"fmt"
	"sql-sharding-v2/internal/config"

	_ "github.com/lib/pq"
)

func LoadAppilcationDatabase() error {

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.ApplicationDatabaseCredentials.DB_USER,
		config.ApplicationDatabaseCredentials.DB_PASS,
		config.ApplicationDatabaseCredentials.DB_HOST,
		config.ApplicationDatabaseCredentials.DB_PORT,
		config.ApplicationDatabaseCredentials.DB_NAME,
	)

	conn, err := sql.Open("postgres", connStr)

	if err != nil {
		return err
	}

	err = conn.Ping()

	if err != nil {
		return err
	}

	config.AppicationDatabaseConnection.ConnInst = conn

	return nil
}
