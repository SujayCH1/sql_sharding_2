package executor

import (
	"context"
	"sql-sharding-v2/internal/connections"
	"sql-sharding-v2/internal/router"
)

// Executor is responsible for executing routed SQL on shards.
type Executor struct {
	connStore *connections.ConnectionStore
}

// NewExecutor creates a new executor service.
func NewExecutor(store *connections.ConnectionStore) *Executor {
	return &Executor{
		connStore: store,
	}
}

// Execute executes a single SQL statement on routed shards.
func (e *Executor) Execute(
	ctx context.Context,
	projectID string,
	sqlText string,
	plan *router.RoutingPlan,
) ([]ExecutionResult, error) {

	// Router already validated this
	if plan.Mode == router.RoutingModeRejected {
		return nil, plan.RejectError
	}

	results := make([]ExecutionResult, 0, len(plan.Targets))

	for _, target := range plan.Targets {

		db, err := e.connStore.Get(projectID, string(target.ShardID))
		if err != nil {
			results = append(results, ExecutionResult{
				ShardID: string(target.ShardID),
				Err:     err,
			})
			continue
		}

		result := executeOnShard(ctx, db, string(target.ShardID), sqlText)
		results = append(results, result)
	}

	return results, nil
}
