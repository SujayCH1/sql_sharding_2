package loader

import (
	"context"
	"sql-sharding-v2/internal/config"
	"sql-sharding-v2/internal/database"
	"sql-sharding-v2/pkg/environment"
	"sql-sharding-v2/pkg/logger"
)

// Funcion to load all application services
func LoadServices(ctx context.Context) {

	//load Environment and environment variables
	err := environment.LoadEnv()
	if err != nil {
		logger.Logger.Error("Failed to load application environment: %s", err)
		panic(err)
	}

	environment.LoadEnvVariables()
	logger.Logger.Info("Successfully loaded environment variables")

	// ensure application databse
	if err := database.EnsureApplicationDatabaseExists(); err != nil {
		logger.Logger.Error("failed to connect application database: ", err)
		panic(err)
	}

	// Load application database
	err = database.LoadAppilcationDatabase()
	if err != nil {
		logger.Logger.Error("failed to load application database: ", err)
		panic(err)
	}

	logger.Logger.Info("Successfully connected to application database")

	// run migrations
	dsn := database.BuildDSN(config.ApplicationDatabaseCredentials.DB_NAME)

	err = database.RunMigrations(dsn)
	if err != nil {
		logger.Logger.Error("failed to run database migrations: ", err)
		panic(err)
	}

	logger.Logger.Info("Successfuly migrated application database")
}
