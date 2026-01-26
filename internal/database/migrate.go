package database

import (
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exeDir := filepath.Dir(exePath)

	projectRoot := filepath.Clean(filepath.Join(exeDir, "..", ".."))

	migrationsPath := filepath.Join(projectRoot, "migrations")

	sourceURL := "file://" + migrationsPath

	m, err := migrate.New(
		sourceURL,
		dsn,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
