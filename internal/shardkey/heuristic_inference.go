package shardkey

import (
	"context"
	"sql-sharding-v2/pkg/logger"
)

func (s *InferenceService) RunHeuristicInference(ctx context.Context, projectID string) error {
	logger.Logger.Info("heuristic inference reached")

	// columns, err := s.columnRepo.GetColumnsByProjectID(ctx, projectID)
	// if err != nil {
	// 	return err
	// }

	// fkEdges, err := s.fkRepo.GetEdgesByProjectID(ctx, projectID)
	// if err != nil {
	// 	return err
	// }

	// logicalSchema, err := schema.BuildLogicalSchemaFromMetadata(
	// 	projectID,
	// 	columns,
	// 	fkEdges,
	// )

	logicalSchema, err := s.buildSchema(ctx, projectID)
	if err != nil {
		return err
	}

	inferenceResult := BuildShardKeyPlan(&logicalSchema)
	inferred := convertDecisionsToShardKeyRecords(inferenceResult.Decisions)

	return s.shardKeyRepo.ReplaceShardKeysForProject(ctx, projectID, inferred)
}
