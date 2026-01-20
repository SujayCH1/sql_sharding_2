package router

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// Planner builds routing plans from parsed statements.
type Planner struct {
	cfg    RouterConfig
	hasher *Hasher
	ring   *Ring
}

func NewPlanner(
	cfg RouterConfig,
	hasher *Hasher,
	ring *Ring,
) *Planner {
	return &Planner{
		cfg:    cfg,
		hasher: hasher,
		ring:   ring,
	}
}

// Plan builds a RoutingPlan for a single SQL statement.
func (p *Planner) Plan(
	node pg_query.Node,
	table string,
	shardKey string,
) *RoutingPlan {

	// 1. Extract shard-key predicate
	pred, err := ExtractShardPredicate(node, table, shardKey)
	if err != nil {
		return &RoutingPlan{
			Mode:        RoutingModeRejected,
			Reason:      err.Message,
			RejectError: err,
		}
	}

	// 2. Hash shard-key values
	hashes := make([]HashValue, 0, len(pred.Values))
	for _, v := range pred.Values {
		hashes = append(hashes, p.hasher.Hash(v))
	}

	// 3. Resolve shards via ring
	shards := p.ring.LocateShards(hashes)

	if len(shards) == 0 {
		return &RoutingPlan{
			Mode:   RoutingModeRejected,
			Reason: "no shards resolved for shard key",
			RejectError: &RoutingError{
				Code:    ErrInvalid,
				Message: "no shards resolved",
			},
		}
	}

	// 4. Enforce fanout limits
	if len(shards) > 1 && len(shards) > p.cfg.MaxShardFanout {
		return &RoutingPlan{
			Mode:   RoutingModeRejected,
			Reason: "shard fanout exceeded",
			RejectError: &RoutingError{
				Code:    ErrFanoutExceeded,
				Message: "query touches too many shards",
			},
		}
	}

	// 5. Build routing plan
	targets := make([]ShardTarget, 0, len(shards))
	for _, sid := range shards {
		targets = append(targets, ShardTarget{
			ShardID: sid,
		})
	}

	mode := RoutingModeSingle
	if len(targets) > 1 {
		mode = RoutingModeMulti
	}

	return &RoutingPlan{
		Mode:    mode,
		Targets: targets,
		Reason:  "shard key resolved successfully",
	}
}
