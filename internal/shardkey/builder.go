package shardkey

import (
	"sql-sharding-v2/internal/schema"
)

// BuildShardKeyPlan runs the full shard key inference pipeline
// using schema-only analysis.
func BuildShardKeyPlan(s *schema.LogicalSchema) InferenceResult {

	result := InferenceResult{
		ProjectID: s.ProjectID,
	}

	candidates := ExtractCandidates(s)

	fanout := ComputeFanout(s, candidates)

	for tableName, localCandidates := range candidates {

		ranked := RankTableCandidates(
			tableName,
			localCandidates,
			fanout,
			s,
		)

		decision := selectBestCandidate(tableName, ranked)
		if decision == nil {
			continue
		}

		result.Decisions = append(result.Decisions, *decision)
	}

	return result
}

func selectBestCandidate(table string, ranked []RankedCandidate) *ShardKeyDecision {

	if len(ranked) == 0 {
		return nil
	}

	best := ranked[0]

	return &ShardKeyDecision{
		Table:   table,
		Column:  best.Column,
		Score:   best.Score,
		Reasons: best.Reasons,
	}
}
