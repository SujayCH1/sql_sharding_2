package main

import (
	"context"
	"sql-sharding-v2/internal/config"
	"sql-sharding-v2/internal/loader"
	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	loader.LoadServices(ctx)
}

// Call to add a new project
func (a *App) CreateProject(name string, description string) (*repository.Project, error) {
	repo := repository.NewProjectRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ProjectAdd(a.ctx, name, description)

	if err != nil {
		logger.Logger.Error("Error while creating Project: %w", err)
		return nil, err
	}

	return result, nil
}

// Call to list existing project
func (a *App) ListProjects() ([]repository.Project, error) {
	repo := repository.NewProjectRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ProjectList(a.ctx)

	if err != nil {
		logger.Logger.Error("Error while fetching Projects: %w", err)
		return nil, err
	}

	return result, nil
}

// Call to delete a project
func (a *App) DeleteProject(id string) error {
	repo := repository.NewProjectRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while deleting project: %w", err)
		return err
	}

	return nil
}
