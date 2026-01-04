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
		config.ApplicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ProjectAdd(a.ctx, name, description)

	if err != nil {
		logger.Logger.Error("Error while creating project", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully created project", "project_name", name)

	return result, nil
}

// project repository - Call to list existing project
func (a *App) ListProjects() ([]repository.Project, error) {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ProjectList(a.ctx)

	if err != nil {
		logger.Logger.Error("Error while fetching projects", "error", err)
		return nil, err
	}

	logger.Logger.Info("Sucessfully fetched all projects")

	return result, nil
}

// project repository - Call to delete a project
func (a *App) DeleteProject(id string) error {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while deleting project: ", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deleted project", "project_id", id)

	return nil
}

// project repository - Call to fetch project by ID
func (a *App) FetchProjectByID(id string) (repository.Project, error) {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	result, err := repo.GetProjectByID(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while fetching project", "error", err)
		return repository.Project{}, err
	}

	logger.Logger.Info("Successfully fetched project", "project_name", result.Name, "project_id", result.ID)

	return result, err
}

// shard repository - Call to add a shard
func (a *App) AddShard(projectID string) (*repository.Shard, error) {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ShardAdd(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to add shard", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully created shard", "shard_id", result.ID)

	return result, nil
}

// shard repository - Call to get list of all shards
func (a *App) ListShards(projectID string) ([]repository.Shard, error) {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	result, err := repo.ShardList(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to list shards", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully fected all shards")

	return result, nil
}

// shard repository - Call to deactivate a shard
func (a *App) DeactivateShard(shardID string) error {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardDeactivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate shard", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deactivated shard", "shard_id", shardID)

	return nil
}

// shard repository - Call to delete all shards of a project
func (a *App) DeleteAllShards(projectID string) error {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardDeleteAll(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to delete shards for project", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deleted all shards")

	return nil
}

// shard repository - Call to delete a single shard
func (a *App) DeleteShard(shardID string) (string, error) {

	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	isInactive, err := a.checkIfShardInactive(shardID)
	if err != nil {
		return "", err
	}

	if !isInactive {
		return "CANNOT_DELETE_ACTIVE_SHARD", nil
	}

	err = a.DeleteConnection(shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard connection", "error", err)
		return "", err
	}

	err = repo.ShardDelete(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard", "error", err)
		return "", err
	}

	logger.Logger.Info("Successfully deleted shard", "shard_id", shardID)
	return "DELETED", nil
}

// shard repository - Call to activate a shard
func (a *App) ActivateShard(shardID string) error {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ShardActivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "error", err)
		return err
	}

	logger.Logger.Info("Successfully activated shard", "shard_id", shardID)

	return nil
}

// shard repository - Call to fetch status of a shard
func (a *App) FetchShardStatus(shardID string) (string, error) {
	repo := repository.NewShardRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	status, err := repo.FetchShardStatus(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fetch shard shard", "error", err)
		return "", err
	}

	logger.Logger.Info("Successfully fetched shard status", "shard_id", shardID)

	return status, nil
}

// shard connection repository - add connection detail for one shard
func (a *App) AddConnection(connectionInfo *repository.ShardConnection) error {
	repo := repository.NewShardConnectionRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ConnectionCreate(a.ctx, connectionInfo)
	if err != nil {
		logger.Logger.Error("Failed to add shard connection details", "error", err)
		return err
	}

	logger.Logger.Info("Successfully added shard connection details", "shard_id", connectionInfo.ShardID)

	return nil
}

// shard connection repo - remove connection detail of a shard
func (a *App) DeleteConnection(shardID string) error {
	repo := repository.NewShardConnectionRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ConnectionRemove(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to remove shard connection details", "error", err)
	}

	logger.Logger.Info("Successfully deleted shard connection details", "shard_id", shardID)

	return nil
}

// shard connection repo - fetch connection details of a shard using shard id
func (a *App) FetchConnectionInfo(shardID string) (repository.ShardConnection, error) {
	repo := repository.NewShardConnectionRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	conn, err := repo.FetchConnectionByShardID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fecth sahrd connection infomation", "error", err)
		return repository.ShardConnection{}, err
	}

	logger.Logger.Info("Successfully fetched shard connection details", "shard_id", shardID)

	return conn, nil
}

// shard connection repo - update existing connection details
func (a *App) UpdateConnection(connInfo repository.ShardConnection) error {
	repo := repository.NewShardConnectionRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ConnectionUpdate(a.ctx, connInfo)
	if err != nil {
		logger.Logger.Error("Failed to update shard connection details", "error", err)
		return err
	}

	logger.Logger.Info("Successfully updated shard connection details", "shard_id", connInfo.ShardID)

	return nil
}

// project repository - set status of project to active
func (a *App) Activateproject(projectID string) error {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	otherProjectsInactive, err := a.checkAllProjectsInactive()
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "error", err)
		return err
	}

	if otherProjectsInactive == false {
		logger.Logger.Error("Failed to activate project", "error", "another project is already active")
		return errors.New("another project is already active")
	}

	allShardsNotActive, err := a.checkAllShardsActive(projectID)
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "error", err)
		return err
	}

	if allShardsNotActive == false {
		logger.Logger.Error("Failed to activate project", "error", "All shards are not active")
		return errors.New("All shards are not active")
	}

	err = repo.ProjectActivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to activate project", "error", err)
		return err
	}

	logger.Logger.Info("Successfully activated the project", "project_id", projectID)

	return nil
}

// project repository - set status of project to inactive
func (a *App) Deactivateproject(projectID string) error {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	err := repo.ProjectDeactivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate project", "error", err)
	}

	logger.Logger.Info("Successfully deactivated the project", "project_id", projectID)

	return nil
}

// project repository - fetch status of a project
func (a *App) FetchProjectStatus(projectID string) (string, error) {
	repo := repository.NewProjectRepository(
		config.ApplicationDatabaseConnection.ConnInst,
	)

	status, err := repo.FetchProjectStatus(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "error", err)
		return "", err
	}

	logger.Logger.Info("Succesfully fetched status of project", "project_id", projectID)
	return status, nil

}
