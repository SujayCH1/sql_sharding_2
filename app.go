package main

import (
	"context"
	"errors"
	"net/http"
	"sql-sharding-v2/internal/api"
	"sql-sharding-v2/internal/connections"
	"sql-sharding-v2/internal/executor"
	"sql-sharding-v2/internal/loader"
	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/internal/router"
	"sql-sharding-v2/internal/schema"
	"sql-sharding-v2/internal/shardkey"
	"sql-sharding-v2/pkg/logger"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	emitter *logger.LogEmitter

	// http
	httpServer *http.Server

	//config
	RouterConfig router.RouterConfig

	// repository
	ProjectRepo               *repository.ProjectRepository
	ShardRepo                 *repository.ShardRepository
	ShardConnectionRepo       *repository.ShardConnectionRepository
	ProjectSchemaRepo         *repository.ProjectSchemaRepository
	SchemaExecutionStatusRepo *repository.SchemaExecutionStatusRepository
	ColumnsRepo               *repository.ColumnRepository
	FKEdgesRepo               *repository.FKEdgesRepository
	ShardKeysRepo             *repository.ShardKeysRepository

	// conn layer
	ShardConnectionStore   *connections.ConnectionStore
	ShardConnectionManager *connections.ConnectionManager

	//services
	SchemaService    *schema.SchemaService
	InferenceService *shardkey.InferenceService
	RouterService    *router.RouterService
	ExecutorService  *executor.Executor
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.emitter = logger.NewLogEmitter(ctx)

	err := loader.LoadServices(ctx)
	if err != nil {
		panic(err)
	}

	db, err := loader.LoadAppilcationDatabase()
	if err != nil {
		panic(err)
	}

	//config
	a.RouterConfig = router.DefaultRouterConfig()

	// repos
	a.ProjectRepo = repository.NewProjectRepository(db)
	a.ShardRepo = repository.NewShardRepository(db)
	a.ShardConnectionRepo = repository.NewShardConnectionRepository(db)
	a.ProjectSchemaRepo = repository.NewProjectSchemaRepository(db)
	a.SchemaExecutionStatusRepo = repository.NewSchemaExecutionStatusRepository(db)
	a.ColumnsRepo = repository.NewColumnsRepository(db)
	a.FKEdgesRepo = repository.NewFKEdgesRepository(db)
	a.ShardKeysRepo = repository.NewShardKeysRepository(db)

	// stores
	a.ShardConnectionStore = connections.NewConnectionStore()

	// managers
	a.ShardConnectionManager = connections.NewConnectionManager(
		a.ShardConnectionStore,
		a.ProjectRepo,
		a.ShardRepo,
		a.ShardConnectionRepo,
	)

	// services
	a.SchemaService = schema.NewSchemaService(
		a.ColumnsRepo,
		a.FKEdgesRepo,
	)
	a.InferenceService = shardkey.NewInferenceService(
		a.ColumnsRepo,
		a.FKEdgesRepo,
		a.ShardKeysRepo,
	)
	a.RouterService = router.NewRouterService(
		a.ShardKeysRepo,
		a.ShardRepo,
		a.RouterConfig,
	)

	err = a.ShardConnectionManager.InititateConnectionsAll(ctx)
	if err != nil {
		logger.Logger.Error("Failed to initiate connection for active project", "error", err)
		// panic(err)
	}

	a.ExecutorService = executor.NewExecutor(
		a.ShardConnectionStore,
	)

	//api
	mux := http.NewServeMux()
	apiHandler := api.NewHandler(a)
	api.RegisterRoutes(mux, apiHandler)
	a.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		logger.Logger.Info("HTTP server started", "addr", ":8080")
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("HTTP server failed", "error", err)
		}
	}()

	// funcs
	a.MonitorShards(a.ctx)

	logger.Logger.Info("Application startup successful!")

}

// project repository - Call to add a new project
func (a *App) CreateProject(name string, description string) (*repository.Project, error) {

	result, err := a.ProjectRepo.ProjectAdd(a.ctx, name, description)
	if err != nil {
		logger.Logger.Error("Error while creating project", "project_name", name, "error", err)
		a.emitter.Error("Project creation failed", "application - CreateProject", map[string]string{
			"project_name": name,
			"error":        err.Error(),
		})
		return nil, err
	}

	logger.Logger.Info("Successfully created project", "project_name", name)
	a.emitter.Info("project creation successful", "application - CreateProject", map[string]string{
		"project_name": name,
	})

	return result, nil
}

// project repository - Call to list existing project
func (a *App) ListProjects() ([]repository.Project, error) {

	result, err := a.ProjectRepo.ProjectList(a.ctx)

	if err != nil {
		logger.Logger.Error("Error while fetching projects", "error", err)
		a.emitter.Error("Projects listing failed", "application - ListProjects", map[string]string{
			"error": err.Error(),
		})
		return nil, err
	}

	return result, nil
}

// project repository - Call to delete a project
func (a *App) DeleteProject(id string) error {

	err := a.ProjectRepo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while deleting project: ", "project _ id", id, "error", err)
		a.emitter.Error("Project deletion failed", "application - DeleteProject", map[string]string{
			"project_id": id,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully deleted project", "project_id", id)
	a.emitter.Info("Project deletion failed", "application - DeleteProject", map[string]string{
		"project_id": id,
	})

	return nil
}

// project repository - Call to fetch project by ID
func (a *App) FetchProjectByID(id string) (repository.Project, error) {

	result, err := a.ProjectRepo.GetProjectByID(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Error while fetching project", "project_id", id, "error", err)
		a.emitter.Error("Project fetching failed", "application - FetchProjectByID", map[string]string{
			"project_id": id,
			"error":      err.Error(),
		})
		return repository.Project{}, err
	}

	return result, err
}

// shard repository - Call to add a shard
func (a *App) AddShard(projectID string) (*repository.Shard, error) {

	result, err := a.ShardRepo.ShardAdd(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to add shard", "project_id", projectID, "error", err)
		a.emitter.Error("Shard addition failed", "application - AddShard", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	logger.Logger.Info("Successfully created shard", "shard_id", result.ID)
	a.emitter.Info("Shard addition successful", "application - AddShard", map[string]string{
		"project_id": projectID,
		"shard_id":   result.ID,
	})

	return result, nil
}

// shard repository - Call to get list of all shards
func (a *App) ListShards(projectID string) ([]repository.Shard, error) {

	result, err := a.ShardRepo.ShardList(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to list shards", "project_id", projectID, "error", err)
		a.emitter.Error("Shard listing failed", "application - ListShards", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return result, nil
}

// shard repository - Call to deactivate a shard
func (a *App) DeactivateShard(shardID string) error {

	if err := a.ShardRepo.ShardDeactivate(a.ctx, shardID); err != nil {
		logger.Logger.Error("Failed to deactivate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard deactivation failed", "application - ShardDeactivate", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return err
	}

	projectID, err := a.ShardRepo.FetchProjectID(a.ctx, shardID)

	shards, err := a.ListShards(projectID)
	if err != nil {
		a.emitter.Error("Project fetching failed for automated project deactivation", "application - ShardDeactivate", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      err.Error(),
		})
		return err
	}

	for _, shard := range shards {
		if shard.Status == "inactive" {

			if err := a.ProjectRepo.ProjectDeactivate(a.ctx, projectID); err != nil {
				logger.Logger.Error("Failed to deactivate project for automated project deactivation", "project_id", projectID, "shard_id", shardID, "error", err)
				a.emitter.Error("Failed to deactivate project for automated project deactivation", "application - ShardDeactivate", map[string]string{
					"project_id": projectID,
					"shard_id":   shardID,
					"error":      err.Error(),
				})
				return err
			}

			runtime.EventsEmit(a.ctx, "project:status_changed", map[string]string{
				"project_id": projectID,
				"status":     "inactive",
			})

			break
		}
	}

	logger.Logger.Info("Automatic project deactivation successful", "project_Id", projectID)
	a.emitter.Info("Automated project deactivation successful", "application - DeactivateProject", map[string]string{
		"project_id": projectID,
	})

	logger.Logger.Info("Successfully deactivated shard", "shard_id", shardID, "project_id", projectID)
	a.emitter.Info("Shard deactivation successful", "application - DeactivateProject", map[string]string{
		"project_id": projectID,
		"shardd_id":  shardID,
	})

	return nil
}

// shard repository - Call to delete all shards of a project
func (a *App) DeleteAllShards(projectID string) error {

	err := a.ShardRepo.ShardDeleteAll(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to delete shards for project", "project_id", projectID, "error", err)
		a.emitter.Error("Shards deletion failed", "application - DeleteAllShards", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully deleted all shards", "project_id", projectID)
	a.emitter.Info("Shards deletion successful", "application - DeleteAllShards", map[string]string{
		"project_id": projectID,
	})

	return nil
}

// shard repository - Call to delete a single shard
func (a *App) DeleteShard(shardID string) (string, error) {

	isInactive, err := a.checkIfShardInactive(shardID)
	if err != nil {
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return "", err
	}

	if !isInactive {
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    "cannot delete active shard",
		})
		return "CANNOT_DELETE_ACTIVE_SHARD", nil
	}

	err = a.DeleteConnection(shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard connection", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    "failed to delete shard connection",
		})
		return "", err
	}

	err = a.ShardRepo.ShardDelete(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return "", err
	}

	logger.Logger.Info("Successfully deleted shard", "shard_id", shardID)
	a.emitter.Info("Shard deletion successful", "application - DeleteShard", map[string]string{
		"shard_id": shardID,
		"error":    err.Error(),
	})

	return "DELETED", nil
}

// shard repository - Call to activate a shard
func (a *App) ActivateShard(shardID string) error {

	err := a.RetryShardConnections(a.ctx)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"shard_id": shardID,
			"error":    "retry mechanism failed",
		})
	}

	projectID, err := a.ShardRepo.FetchProjectID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"shard_id": shardID,
			"error":    "project fetching failed for health checking",
		})
		return err
	}

	isConnected, err := a.checkShardHealth(a.ctx, projectID, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "projectid", projectID, "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      "failed for checking shard health",
		})
		return err
	}

	if !isConnected {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", "shard connection not available")
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      "shard connection failed",
		})
		return err
	}

	err = a.ShardRepo.ShardActivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully activated shard", "shard_id", shardID)
	a.emitter.Info("Shard activation successfull", "application - ActivateShard", map[string]string{
		"shard_id": shardID,
	})

	return nil
}

// shard repository - Call to fetch status of a shard
func (a *App) FetchShardStatus(shardID string) (string, error) {

	status, err := a.ShardRepo.FetchShardStatus(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fetch shard shard", "shard_Id", shardID, "error", err)
		a.emitter.Error("Shard status fetching failed", "application - FetchShardStatus", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return "", err
	}

	return status, nil
}

// shard connection repository - add connection detail for one shard
func (a *App) AddConnection(connectionInfo *repository.ShardConnection) error {

	err := a.ShardConnectionRepo.ConnectionCreate(a.ctx, connectionInfo)
	if err != nil {
		logger.Logger.Error("Failed to add shard connection details", "shard_id", connectionInfo.ShardID, "error", err)
		a.emitter.Error("shard connection addition failed", "application - AddConnection", map[string]string{
			"shard_id": connectionInfo.ShardID,
			"error":    err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully added shard connection details", "shard_id", connectionInfo.ShardID)
	a.emitter.Info("shard connection addition successfull", "application - AddConnection", map[string]string{
		"shard_id": connectionInfo.ShardID,
	})

	return nil
}

// shard connection repo - remove connection detail of a shard
func (a *App) DeleteConnection(shardID string) error {

	err := a.ShardConnectionRepo.ConnectionRemove(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to remove shard connection details", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard connection deletion failed", "application - DeleteConnection", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
	}

	logger.Logger.Info("Successfully deleted shard connection details", "shard_id", shardID)
	a.emitter.Info("Shard connection deletion successfull", "application - DeleteConnection", map[string]string{
		"shard_id": shardID,
	})

	return nil
}

// shard connection repo - fetch connection details of a shard using shard id
func (a *App) FetchConnectionInfo(shardID string) (repository.ShardConnection, error) {

	conn, err := a.ShardConnectionRepo.FetchConnectionByShardID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fecth sahrd connection infomation", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard connection info fetching failed", "application - FetchConnectionInfo", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return repository.ShardConnection{}, err
	}

	return conn, nil
}

// shard connection repo - update existing connection details
func (a *App) UpdateConnection(connInfo repository.ShardConnection) error {

	err := a.ShardConnectionRepo.ConnectionUpdate(a.ctx, connInfo)
	if err != nil {
		logger.Logger.Error("Failed to update shard connection details", "shard_id", connInfo.ShardID, "error", err)
		a.emitter.Error("Shard connection updation failed", "application - UpdateConnection", map[string]string{
			"shard_id": connInfo.ShardID,
			"error":    err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully updated shard connection details", "shard_id", connInfo.ShardID)
	a.emitter.Info("Shard connection updation successfull", "application - UpdateConnection", map[string]string{
		"shard_id": connInfo.ShardID,
	})

	return nil
}

// project repository - set status of project to active
func (a *App) Activateproject(projectID string) error {

	otherProjectsInactive, err := a.checkAllProjectsInactive()
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "project_id", projectID, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectID,
			"error":      "unable to fetch projects for activation constratints",
		})
		return err
	}

	if otherProjectsInactive == false {
		logger.Logger.Error("Failed to activate project", "error", "project_id", projectID, "another project is already active")
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectID,
			"error":      "another project is already active",
		})
		return errors.New("another project is already active")
	}

	allShardsNotActive, err := a.checkAllShardsActive(projectID)
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "project_id", projectID, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectID,
			"error":      "all project shards are not active",
		})
		return err
	}

	if allShardsNotActive == false {
		logger.Logger.Error("Failed to activate project", "error", "project_id", projectID, "All shards are not active")
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectID,
			"error":      "another project is already active",
		})
		return errors.New("All shards are not active")
	}

	err = a.ProjectRepo.ProjectActivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to activate project", "project_id", projectID, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	err = a.ShardConnectionManager.InititateConnectionsAll(a.ctx)

	logger.Logger.Info("Successfully activated the project", "project_id", projectID)
	a.emitter.Info("Project activation successfull", "application - Activateproject", map[string]string{
		"project_id": projectID,
	})

	return nil
}

// project repository - set status of project to inactive
func (a *App) Deactivateproject(projectID string) error {

	err := a.ProjectRepo.ProjectDeactivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate project", "project_id", projectID, "error", err)
		a.emitter.Error("Project deletion failed", "application - Deactivateproject", map[string]string{
			"project_Id": projectID,
			"error":      err.Error(),
		})
	}

	logger.Logger.Info("Successfully deactivated the project", "project_id", projectID)
	a.emitter.Info("Project deletion successfull", "application - Deactivateproject", map[string]string{
		"project_Id": projectID,
	})

	return nil
}

// project repository - fetch status of a project
func (a *App) FetchProjectStatus(projectID string) (string, error) {

	status, err := a.ProjectRepo.FetchProjectStatus(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "project_id", projectID, "error", err)
		a.emitter.Error("Project status fetching failed", "application - FetchProjectStatus", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return "", err
	}

	return status, nil

}

// project schema repository - create new schema draft
func (a *App) CreateSchemaDraft(projectID string, ddlSQL string) (*repository.ProjectSchema, error) {

	schema, err := a.ProjectSchemaRepo.ProjectSchemaCreateDraft(a.ctx, projectID, ddlSQL)
	if err != nil {
		logger.Logger.Error("Failed to create schema draft", "project_id", projectID, "error", err)
		a.emitter.Error("Schema draft creation failed", "application - CreateSchemaDraft", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	logger.Logger.Info("Succesfully created schema draft of project", "project_id", projectID)
	a.emitter.Info("Schema draft creation successfull", "application - CreateSchemaDraft", map[string]string{
		"project_id": projectID,
	})
	return schema, nil

}

// project schema repository - commit existing schema draft
func (a *App) CommitSchemaDraft(projectID string, schemaID string) error {

	ok, err := a.checkIfProjectInactive(projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to check project status",
		})
		return err
	}

	if !ok {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "project is active",
		})
		return errors.New("project must be inactive to modify schema")
	}

	ok, err = a.checkIfSchemaDraft(schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch schema state", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to fetch schema state",
		})
		return err
	}
	if !ok {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "schema should be in draft before commit",
		})
		return errors.New("schema must be in draft state to commit")
	}

	inFlight, err := a.checkIfSchemaInFlight(projectID)
	if err != nil {
		logger.Logger.Error("Failed to check schema in-flight status", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to check schema in-flight status",
		})
		return err
	}
	if inFlight {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "another schema change is already in progress",
		})
		return errors.New("another schema change is already in progress")
	}

	projectSchema, err := a.ProjectSchemaRepo.ProjectSchemaFetchBySchemaID(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch schema by id", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to fetch schema infomation",
		})
		return err
	}

	destructive, err := a.checkIfDDLDestructive(projectID, projectSchema.DDL_SQL)
	if err != nil {
		logger.Logger.Error("Failed to validate destructive DDL", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to validate destructive DDL",
		})
		return err
	}
	if destructive {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "destructive ddl in query",
		})
		return errors.New("destructive DDL is not allowed after initial schema")
	}

	logger.Logger.Info("Applying committed schema to metadata", "project_id", projectID, "schema_id", schemaID)
	a.emitter.Info("applying committed schema to metadata", "application - CommitSchemaDraft", map[string]string{
		"project_id": projectID,
	})

	err = a.SchemaService.ApplyDDLAndRecomputeShardKeys(a.ctx, projectID, projectSchema.DDL_SQL)
	if err != nil {
		logger.Logger.Error("Failed to apply schema to metadata", "project_id", projectID, "schema_id", schemaID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "Failed to apply schema to metadata" + err.Error(),
		})
		return err
	}

	err = a.ProjectSchemaRepo.ProjectSchemaCommitDraft(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to commit project schema", "project_id", projectID, "schema_id", schemaID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "Failed to commit project schema" + err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully committed project schema", "project_id", projectID, "schema_id", schemaID)
	a.emitter.Info("Schema draft commit successfull", "application - CommitSchemaDraft", map[string]string{
		"project_id": projectID,
	})

	return nil
}

// project schema repository - get latest schema of project
func (a *App) GetCurrentSchema(projectID string) (*repository.ProjectSchema, error) {

	schema, err := a.ProjectSchemaRepo.ProjectSchemaGetLatest(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch latest schema of project", "project_id", projectID, "error", err)
		a.emitter.Error("Current schema fetching failed", "application - GetCurrentSchema", map[string]string{
			"projecy_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return schema, nil

}

// project schema repository - get history of a schema
func (a *App) GetSchemaHistory(projectID string) ([]repository.ProjectSchema, error) {

	history, err := a.ProjectSchemaRepo.ProjectSchemaFetchHistory(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project schema history", "project_id", projectID, "error", err)
		a.emitter.Error("Schema history fetching failed", "application - GetSchemaHistory", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return history, nil

}

// project schema repository - delete a draft of schema
func (a *App) DeleteSchemaDraft(schemaID string) error {

	err := a.ProjectSchemaRepo.ProjectSchemaDeleteDraft(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to delete project schema draft", "schema_id", schemaID, "error", err)
		a.emitter.Error("Schema deletion failed", "application - DeleteSchemaDraft", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully deleted project schema draft", "schema_id", schemaID)
	a.emitter.Info("Schema deletion successfull", "application - DeleteSchemaDraft", map[string]string{
		"schema_id": schemaID,
	})
	return nil

}

// schema execution status repo -get execution status of all shards
func (a *App) GetSchemaExecutionStatus(schemaID string) ([]repository.SchemaExecutionStatus, error) {

	statuAll, err := a.SchemaExecutionStatusRepo.ExecutionRecordsFetchStatusAll(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch execution status of all records of schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Schema execution status fetching failed", "application - GetSchemaExecutionStatus", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return nil, err
	}

	return statuAll, err
}

// schema execution status repo - get record of failed shard executions
func (a *App) GetFailedShardExecutions(schemaID string) ([]repository.SchemaExecutionStatus, error) {

	statuAll, err := a.SchemaExecutionStatusRepo.ExecutionRecordsFetchStatusFailed(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch execution status of all failed records of schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Failed shard execution status fetching failed", "application - GetFailedShardExecutions", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return nil, err
	}

	return statuAll, err
}

// project schema repository - get status of a projectschema
func (a *App) GetProjectSchemaStatus(schemaID string) (string, error) {
	status, err := a.ProjectSchemaRepo.ProjectSchemaGetState(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Fialed to fetch status of a schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Project schema status fetching failed", "application - GetProjectSchemaStatus", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return "", err
	}

	return status, nil
}

// DDL executor - execute pending schema
func (a *App) ExecuteProjectSchema(projectID string) error {

	caps, err := a.GetSchemaCapabilities(projectID)
	if err != nil {
		a.emitter.Error("Project schema execution failed", "application - ExecuteProjectSchema", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	if !caps.CanExecute {
		a.emitter.Error("Project schema execution failed", "application - ExecuteProjectSchema", map[string]string{
			"project_id": projectID,
			"error":      "schema execution not allowed",
		})
		return errors.New("schema execution not allowed")
	}

	err = schema.ExecuteProjectSchema(
		a.ctx,
		projectID,
		a.ProjectSchemaRepo,
		a.ShardRepo,
		a.SchemaExecutionStatusRepo,
		func(shardID string, ddl string) error {
			return a.execDDLonShard(projectID, shardID, ddl)
		},
	)

	if err != nil {
		logger.Logger.Error("failed to execute project schema", "project_id", projectID, "error", err)
		a.emitter.Error("Project schema execution failed", "application - ExecuteProjectSchema", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully executed projects schema", "project_id", projectID)
	a.emitter.Info("Project schema execution successfull", "application - ExecuteProjectSchema", map[string]string{
		"project_id": projectID,
	})

	return nil
}

// DDL executor - retry mechanism
func (a *App) RetrySchemaExecution(projectID string) error {

	return schema.RetryFailedSchema(
		a.ctx,
		projectID,
		a.ProjectSchemaRepo,
		a.SchemaExecutionStatusRepo,
	)
}

// shardkey service - run inference
func (a *App) RecomputeKeys(projectID string) error {

	err := a.InferenceService.ApplyShardKeyInference(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("shard key inference failed", "project_id", projectID, "error", err)
		a.emitter.Error("Shard key recomputation failed", "application -RecomputeKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("shard key inference completed successfully", "project_id", projectID)
	a.emitter.Info("Shard key recomputation successfull	", "application -RecomputeKeys", map[string]string{
		"project_id": projectID,
	})
	return nil

}

// shard keys repository - fetch keys
func (a *App) FetchShardKeys(projectID string) ([]repository.ShardKeys, error) {

	keys, err := a.ShardKeysRepo.FetchShardKeysByProjectID(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("failed to fetch shard keys", "project_id", projectID, "error", err)
		a.emitter.Error("Shard key fetching failed", "application - FetchShardKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return keys, nil

}

// shard keys repository - replace keys
func (a *App) ReplaceShardKeys(projectID string, keys []repository.ShardKeyRecord) error {

	err := a.ShardKeysRepo.ReplaceShardKeysForProject(a.ctx, projectID, keys)
	if err != nil {
		logger.Logger.Error("filed to replace shard keys", "projectID", projectID, "error", err)
		a.emitter.Error("Shard key replacing failed", "application - ReplaceShardKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("successfully replaced shard keys", "projectID", projectID)
	a.emitter.Info("Shard key replacing successfull", "application - ReplaceShardKeys", map[string]string{
		"project_id": projectID,
	})
	return nil
}

// func to execute DML quereis on repective schema
func (a *App) ExecuteSQL(projectID string, sqlText string) ([]executor.ExecutionResult, error) {

	plan, err := a.RouterService.RouteSQL(
		a.ctx,
		projectID,
		sqlText,
	)
	if err != nil {
		logger.Logger.Error("Filed to route and execute query", "project_id", projectID, "error", err)
		a.emitter.Error("Query execution failed", "application - ExecuteSQL", map[string]string{
			"project_id": projectID,
			"error":      "query router:" + err.Error(),
		})
		return nil, err
	}

	result, err := a.ExecutorService.Execute(
		a.ctx,
		projectID,
		sqlText,
		plan,
	)
	if err != nil {
		logger.Logger.Error("failed to execute query", "project_id", projectID, "error", err)
		logger.Logger.Error("Filed to route and execute query", "project_id", projectID, "error", err)
		a.emitter.Error("Query execution failed", "application - ExecuteSQL", map[string]string{
			"project_id": projectID,
			"error":      "query executor:" + err.Error(),
		})
		return nil, err
	}

	return result, nil
}

// func to continuously check  health of shards of projects
func (a *App) MonitorShards(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Shard monitor stopped")
			a.emitter.Info("Shard monitor stopper", "application - MonitorShards", map[string]string{})
			return

		case <-ticker.C:
			a.checkAllShards(ctx)
		}
	}
}

// func to retry all projects shard connections
func (a *App) RetryShardConnections(ctx context.Context) error {

	err := a.ShardConnectionManager.InititateConnectionsAll(a.ctx)
	if err != nil {
		logger.Logger.Error("shard connection retry mechanism failed", "error", err)
		a.emitter.Error("Shard connection retry mechanism failed", "application - RetryShardConnections", map[string]string{
			"error": err.Error(),
		})
		return err
	}

	a.emitter.Info("Shard connection retry mechanism successfull", "application - RetryShardConnections", map[string]string{})
	return nil

}

// allows flow of schema draft -> execute in fe
func (a *App) GetSchemaCapabilities(projectID string) (*SchemaCapabilities, error) {
	caps := &SchemaCapabilities{}

	projectStatus, err := a.FetchProjectStatus(projectID)
	if err != nil {
		return nil, err
	}

	schema, _ := a.GetCurrentSchema(projectID)

	if projectStatus == "active" {
		caps.Reason = "Project is active"
		return caps, nil
	}

	if schema == nil {
		caps.CanCreateDraft = true
		return caps, nil
	}

	switch schema.State {

	case "draft":
		caps.CanEditDraft = true
		caps.CanCommit = true

	case "pending":
		allActive, err := a.checkAllShardsActive(projectID)
		if err != nil {
			return nil, err
		}
		if allActive {
			caps.CanExecute = true
		} else {
			caps.Reason = "All shards must be active to execute"
		}

	case "failed":
		caps.CanRetry = true

	case "applied":
		caps.CanCreateDraft = true
	}

	return caps, nil
}

// helper to pass repos to DDL executor
func (a *App) execDDLonShard(projectID string, shardID string, ddl string) error {

	db, err := a.ShardConnectionStore.Get(projectID, shardID)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(a.ctx, ddl)
	return err
}
