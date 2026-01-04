package loader

import (
	"context"
	"database/sql"
	"sql-sharding-v2/internal/config"
	"sql-sharding-v2/internal/database"
	"sql-sharding-v2/pkg/environment"
	"sql-sharding-v2/pkg/logger"
)

// Funcion to load all application services
func LoadServices(ctx context.Context) error {

	//load Environment and environment variables
	err := environment.LoadEnv()
	if err != nil {
		logger.Logger.Error("Failed to load application environment", "error", err)
		return err
	}
	environment.LoadEnvVariables()
	logger.Logger.Info("Successfully loaded environment variables")

	// ensure application databse
	if err := database.EnsureApplicationDatabaseExists(); err != nil {
		logger.Logger.Error("failed to connect application database", "error", err)
		return err
	}

	// run migrations
	dsn := database.BuildDSN(config.ApplicationDatabaseCredentials.DB_NAME)
	err = database.RunMigrations(dsn)
	if err != nil {
		logger.Logger.Error("failed to run database migrations", "error", err)
		return err
	}
	logger.Logger.Info("Successfuly migrated application database")

	return nil
}

func LoadAppilcationDatabase() (*sql.DB, error) {

	db, err := database.LoadAppilcationDatabase()
	if err != nil {
		logger.Logger.Error("failed to load application database", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully connected to application database")
	return db, nil
}
