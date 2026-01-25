package connections

import (
	"context"
	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

type ConnectionManager struct {
	store         *ConnectionStore
	projectRepo   *repository.ProjectRepository
	shardRepo     *repository.ShardRepository
	shardConnRepo *repository.ShardConnectionRepository
}

func NewConnectionManager(
	store *ConnectionStore,
	projectRepo *repository.ProjectRepository,
	shardRepo *repository.ShardRepository,
	shardConnRepo *repository.ShardConnectionRepository,
) *ConnectionManager {
	return &ConnectionManager{
		store:         store,
		projectRepo:   projectRepo,
		shardRepo:     shardRepo,
		shardConnRepo: shardConnRepo,
	}
}

// func to get and store connection for all active shards
func (m *ConnectionManager) InitiateActiveConnections(ctx context.Context) error {

	activeProj, err := m.projectRepo.FetchActiveProject(ctx)
	if err != nil {
		return err
	}

	if activeProj == "" {
		logger.Logger.Error("No active projects")
		return nil
	}

	shards, err := m.shardRepo.ShardList(ctx, activeProj)
	if err != nil {
		return err
	}

	for _, shard := range shards {

		connInfo, err := m.shardConnRepo.FetchConnectionByShardID(ctx, shard.ID)
		if err != nil {
			logger.Logger.Warn("Failed to connect shards", "error", err)
			continue
		}

		dsn := buildDSN(connInfo)

		db, err := NewConnection(ctx, dsn)
		if err != nil {
			logger.Logger.Warn("Failed to connect shards", "error", err)
			continue
		}

		m.store.Set(activeProj, shard.ID, db)
	}

	logger.Logger.Info("Sucessfully initiated shard connections for active project")
	return nil
}

// func to ping a shardto check its connection status
func (m *ConnectionManager) CheckConnectionHealth(ctx context.Context, projectID string, shardID string) (bool, error) {

	conn, err := m.store.Get(projectID, shardID)
	if err != nil {
		return false, err
	}

	err = conn.Ping()
	if err != nil {
		return false, err
	}

	return true, nil

}
