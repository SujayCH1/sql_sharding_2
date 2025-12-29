package environment

import (
	"os"
	"sql-sharding-v2/internal/config"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}

func LoadEnvVariables() {
	config.ApplicationDatabaseCredentials.DB_HOST = os.Getenv("DB_HOST")
	config.ApplicationDatabaseCredentials.DB_NAME = os.Getenv("DB_NAME")
	config.ApplicationDatabaseCredentials.DB_PASS = os.Getenv("DB_PASSWORD")
	config.ApplicationDatabaseCredentials.DB_PORT = os.Getenv("DB_PORT")
	config.ApplicationDatabaseCredentials.DB_USER = os.Getenv("DB_USER")
}
