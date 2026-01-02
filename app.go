package main

import (
	"context"
	"errors"
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

// project repository - Call to add a new project
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

// project repository - Call to list existing project
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

// project repository - Call to delete a project
func (a *App) DeleteProject(id string) error {
	repo := repository.NewProjectRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while deleting project: ", err)
		return err
	}

	return nil
}

// project repository - Call to fetch project by ID
func (a *App) FetchProjectByID(id string) (repository.Project, error) {
	repo := repository.NewProjectRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	resullt, err := repo.GetProjectByID(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while fetching project: ", err)
		return repository.Project{}, err
	}

	return resullt, err
}

// shard repository - Call to add a shard
func (a *App) AddShard(projectID string) (*repository.Shard, error) {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ShardAdd(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to add shard: ", err)
		return nil, err
	}

	return result, nil
}

// shard repository - Call to get list of all shards
func (a *App) ListShards(projectID string) ([]repository.Shard, error) {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ShardList(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to list shards: ", err)
		return nil, err
	}

	return result, nil
}

// shard repository - Call to deactivate a shard
func (a *App) DeactivateShard(shardID string) error {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardDeactivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate shard: ", err)
		return err
	}

	return nil
}

// shard repository - Call to delete all shards of a project
func (a *App) DeleteAllShards(projectID string) error {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardDeleteAll(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to delete shards for project: ", err)
		return err
	}

	return nil
}

// shard repository - Call to delete a single shard
func (a *App) DeleteShard(shardID string) (string, error) {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardDelete(a.ctx, shardID)
	if err == nil {
		return "DELETED", nil
	}

	if errors.Is(err, repository.ErrShardDeleteBlocked) {
		return "CANNOT_DELETE_ACTIVE_SHARD", nil
	}

	logger.Logger.Error("Failed to delete shard: ", err)
	return "", err
}

// shard repository - Call to activate a shard
func (a *App) ActivateShard(shardID string) error {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardActivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard: ", err)
		return err
	}

	return nil
}

// shard repository - Call to fetch a shard of status
func (a *App) FetchShardStatus(shardID string) (string, error) {
	repo := repository.NewShardRepository(
		config.AppicationDatabaseConnection.ConnInst,
	)

	status, err := repo.FetchShardStatus(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fetch shard shard: ", err)
		return "", err
	}

	return status, nil
}
