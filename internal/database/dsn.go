package database

import (
	"fmt"
	"sql-sharding-v2/internal/config"
)

func BuildDSN(dbName string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.ApplicationDatabaseCredentials.DB_USER,
		config.ApplicationDatabaseCredentials.DB_PASS,
		config.ApplicationDatabaseCredentials.DB_HOST,
		config.ApplicationDatabaseCredentials.DB_PORT,
		config.ApplicationDatabaseCredentials.DB_NAME,
	)
}
