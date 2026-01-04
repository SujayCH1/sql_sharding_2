package main

import (
	"context"
	"errors"
	"sql-sharding-v2/internal/connections"
	"sql-sharding-v2/internal/loader"
	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

// App struct
type App struct {
	ctx context.Context

	// repository
	ProjectRepo         *repository.ProjectRepository
	ShardRepo           *repository.ShardRepository
	ShardConnectionRepo *repository.ShardConnectionRepository

	// conn layer
	ShardConnectionStore   *connections.ConnectionStore
	ShardConnectionManager *connections.ConnectionManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	err := loader.LoadServices(ctx)
	if err != nil {
		panic(err)
	}

	db, err := loader.LoadAppilcationDatabase()
	if err != nil {
		panic(err)
	}

	a.ProjectRepo = repository.NewProjectRepository(db)
	a.ShardRepo = repository.NewShardRepository(db)
	a.ShardConnectionRepo = repository.NewShardConnectionRepository(db)

	a.ShardConnectionStore = connections.NewConnectionStore()

	a.ShardConnectionManager = connections.NewConnectionManager(
		a.ShardConnectionStore,
		a.ProjectRepo,
		a.ShardRepo,
		a.ShardConnectionRepo,
	)

	err = a.ShardConnectionManager.InitiateActiveConnections(ctx)
	if err != nil {
		logger.Logger.Error("Failed to initiate connection for active project", "error", err)
		// panic(err)
	}

	logger.Logger.Info("Application startup successful!")

}

// var ProjectRepo repository.ProjectRepository = *repository.NewProjectRepository(config.ApplicationDatabaseConnection.ConnInst)

// var ShardRepo repository.ShardRepository = *repository.NewShardRepository(config.ApplicationDatabaseConnection.ConnInst)

// var ShardConnectionRepo repository.ShardConnectionRepository = *repository.NewShardConnectionRepository(config.ApplicationDatabaseConnection.ConnInst)

// var ShardConnectionStore connections.ConnectionStore = *connections.NewConnectionStore()

// var ShardConnectionManager connections.ConnectionManager = *connections.NewConnectionManager(
// 	&ShardConnectionStore,
// 	&ProjectRepo,
// 	&ShardRepo,
// 	&ShardConnectionRepo,
// )

// project repository - Call to add a new project
func (a *App) CreateProject(name string, description string) (*repository.Project, error) {

	result, err := a.ProjectRepo.ProjectAdd(a.ctx, name, description)

	if err != nil {
		logger.Logger.Error("Error while creating project", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully created project", "project_name", name)

	return result, nil
}

// project repository - Call to list existing project
func (a *App) ListProjects() ([]repository.Project, error) {

	result, err := a.ProjectRepo.ProjectList(a.ctx)

	if err != nil {
		logger.Logger.Error("Error while fetching projects", "error", err)
		return nil, err
	}

	logger.Logger.Info("Sucessfully fetched all projects")

	return result, nil
}

// project repository - Call to delete a project
func (a *App) DeleteProject(id string) error {

	err := a.ProjectRepo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while deleting project: ", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deleted project", "project_id", id)

	return nil
}

// project repository - Call to fetch project by ID
func (a *App) FetchProjectByID(id string) (repository.Project, error) {

	result, err := a.ProjectRepo.GetProjectByID(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while fetching project", "error", err)
		return repository.Project{}, err
	}

	logger.Logger.Info("Successfully fetched project", "project_name", result.Name, "project_id", result.ID)

	return result, err
}

// shard repository - Call to add a shard
func (a *App) AddShard(projectID string) (*repository.Shard, error) {

	result, err := a.ShardRepo.ShardAdd(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to add shard", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully created shard", "shard_id", result.ID)

	return result, nil
}

// shard repository - Call to get list of all shards
func (a *App) ListShards(projectID string) ([]repository.Shard, error) {

	result, err := a.ShardRepo.ShardList(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to list shards", "error", err)
		return nil, err
	}

	logger.Logger.Info("Successfully fetched all shards")

	return result, nil
}

// shard repository - Call to deactivate a shard
func (a *App) DeactivateShard(shardID string) error {

	err := a.ShardRepo.ShardDeactivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate shard", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deactivated shard", "shard_id", shardID)

	return nil
}

// shard repository - Call to delete all shards of a project
func (a *App) DeleteAllShards(projectID string) error {

	err := a.ShardRepo.ShardDeleteAll(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to delete shards for project", "error", err)
		return err
	}

	logger.Logger.Info("Successfully deleted all shards")

	return nil
}

// shard repository - Call to delete a single shard
func (a *App) DeleteShard(shardID string) (string, error) {

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

	err = a.ShardRepo.ShardDelete(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard", "error", err)
		return "", err
	}

	logger.Logger.Info("Successfully deleted shard", "shard_id", shardID)
	return "DELETED", nil
}

// shard repository - Call to activate a shard
func (a *App) ActivateShard(shardID string) error {

	err := a.ShardRepo.ShardActivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "error", err)
		return err
	}

	logger.Logger.Info("Successfully activated shard", "shard_id", shardID)

	return nil
}

// shard repository - Call to fetch status of a shard
func (a *App) FetchShardStatus(shardID string) (string, error) {

	status, err := a.ShardRepo.FetchShardStatus(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fetch shard shard", "error", err)
		return "", err
	}

	logger.Logger.Info("Successfully fetched shard status", "shard_id", shardID)

	return status, nil
}

// shard connection repository - add connection detail for one shard
func (a *App) AddConnection(connectionInfo *repository.ShardConnection) error {

	err := a.ShardConnectionRepo.ConnectionCreate(a.ctx, connectionInfo)
	if err != nil {
		logger.Logger.Error("Failed to add shard connection details", "error", err)
		return err
	}

	logger.Logger.Info("Successfully added shard connection details", "shard_id", connectionInfo.ShardID)

	return nil
}

// shard connection repo - remove connection detail of a shard
func (a *App) DeleteConnection(shardID string) error {

	err := a.ShardConnectionRepo.ConnectionRemove(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to remove shard connection details", "error", err)
	}

	logger.Logger.Info("Successfully deleted shard connection details", "shard_id", shardID)

	return nil
}

// shard connection repo - fetch connection details of a shard using shard id
func (a *App) FetchConnectionInfo(shardID string) (repository.ShardConnection, error) {

	conn, err := a.ShardConnectionRepo.FetchConnectionByShardID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fecth sahrd connection infomation", "error", err)
		return repository.ShardConnection{}, err
	}

	logger.Logger.Info("Successfully fetched shard connection details", "shard_id", shardID)

	return conn, nil
}

// shard connection repo - update existing connection details
func (a *App) UpdateConnection(connInfo repository.ShardConnection) error {

	err := a.ShardConnectionRepo.ConnectionUpdate(a.ctx, connInfo)
	if err != nil {
		logger.Logger.Error("Failed to update shard connection details", "error", err)
		return err
	}

	logger.Logger.Info("Successfully updated shard connection details", "shard_id", connInfo.ShardID)

	return nil
}

// project repository - set status of project to active
func (a *App) Activateproject(projectID string) error {

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

	err = a.ProjectRepo.ProjectActivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to activate project", "error", err)
		return err
	}

	logger.Logger.Info("Initiating shard connection for active projects")

	err = a.ShardConnectionManager.InitiateActiveConnections(a.ctx)

	logger.Logger.Info("Successfully activated the project", "project_id", projectID)

	return nil
}

// project repository - set status of project to inactive
func (a *App) Deactivateproject(projectID string) error {

	err := a.ProjectRepo.ProjectDeactivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate project", "error", err)
	}

	logger.Logger.Info("Successfully deactivated the project", "project_id", projectID)

	return nil
}

// project repository - fetch status of a project
func (a *App) FetchProjectStatus(projectID string) (string, error) {

	status, err := a.ProjectRepo.FetchProjectStatus(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "error", err)
		return "", err
	}

	logger.Logger.Info("Succesfully fetched status of project", "project_id", projectID)
	return status, nil

}
