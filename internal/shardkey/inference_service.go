package shardkey

import (
	"context"

	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/internal/schema"
	"sql-sharding-v2/pkg/logger"
)

type InferenceService struct {
	columnRepo   *repository.ColumnRepository
	fkRepo       *repository.FKEdgesRepository
	shardKeyRepo *repository.ShardKeysRepository
}

func NewInferenceService(
	columnRepo *repository.ColumnRepository,
	fkRepo *repository.FKEdgesRepository,
	shardKeyRepo *repository.ShardKeysRepository,
) *InferenceService {
	return &InferenceService{
		columnRepo:   columnRepo,
		fkRepo:       fkRepo,
		shardKeyRepo: shardKeyRepo,
	}
}

func (s *InferenceService) ApplyShardKeyInference(
	ctx context.Context,
	projectID string,
) error {

	logger.Logger.Info("inference entry reached")

	columns, err := s.columnRepo.GetColumnsByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	fkEdges, err := s.fkRepo.GetEdgesByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	logicalSchema, err := schema.BuildLogicalSchemaFromMetadata(
		projectID,
		columns,
		fkEdges,
	)
	if err != nil {
		return err
	}

	inferenceResult := BuildShardKeyPlan(logicalSchema)
	inferred := convertDecisionsToShardKeyRecords(inferenceResult.Decisions)

	return s.shardKeyRepo.ReplaceShardKeysForProject(ctx, projectID, inferred)
}

func convertDecisionsToShardKeyRecords(
	decisions []ShardKeyDecision,
) []repository.ShardKeyRecord {

	records := make([]repository.ShardKeyRecord, 0, len(decisions))

	for _, d := range decisions {
		records = append(records, repository.ShardKeyRecord{
			TableName:      d.Table,
			ShardKeyColumn: d.Column.Column,
			IsManual:       false,
		})
	}

	return records
}
