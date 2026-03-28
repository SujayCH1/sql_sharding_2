package shardkey

import (
	"context"
	"fmt"

	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

type InferenceService struct {
	columnRepo   *repository.ColumnRepository
	fkRepo       *repository.FKEdgesRepository
	shardKeyRepo *repository.ShardKeysRepository
	aiConfigRepo *repository.AIConfigRepository
}

func NewInferenceService(
	columnRepo *repository.ColumnRepository,
	fkRepo *repository.FKEdgesRepository,
	shardKeyRepo *repository.ShardKeysRepository,
	aiConfigRepo *repository.AIConfigRepository,
) *InferenceService {
	return &InferenceService{
		columnRepo:   columnRepo,
		fkRepo:       fkRepo,
		shardKeyRepo: shardKeyRepo,
		aiConfigRepo: aiConfigRepo,
	}
}

func (s *InferenceService) ApplyShardKeyInference(
	ctx context.Context,
	projectID string,
) error {

	config, err := s.aiConfigRepo.GetConfigByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// No config at all → fallback
	if config == nil {
		logger.Logger.Info("No AI config found → using heuristic inference")
		return s.RunHeuristicInference(ctx, projectID)
	}

	fmt.Println("AI Config:", config)

	// Ollama (no API key needed)
	if config.Provider == "ollama" {
		logger.Logger.Info("Using Ollama for inference")

		err := s.RunLLMInference(ctx, projectID, config)
		if err != nil {
			logger.Logger.Warn("Ollama failed → fallback to heuristic", "error", err)
			return s.RunHeuristicInference(ctx, projectID)
		}

		return nil
	}

	// Other providers (require API key)
	if config.APIKey == "" {
		logger.Logger.Info("No API key → fallback to heuristic")
		return s.RunHeuristicInference(ctx, projectID)
	}

	// Ping external LLM
	if err := PingLLM(ctx, config.APIKey, config.Model); err != nil {
		logger.Logger.Warn("LLM unavailable → fallback to heuristic", "error", err)
		return s.RunHeuristicInference(ctx, projectID)
	}

	// LLM inference
	return s.RunLLMInference(ctx, projectID, config)
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
