package router

import (
	"context"
	"fmt"
	"sort"

	pg_query "github.com/pganalyze/pg_query_go/v5"

	"sql-sharding-v2/internal/repository"
	"sql-sharding-v2/pkg/logger"
)

type RouterService struct {
	shardKeysRepo *repository.ShardKeysRepository
	shardRepo     *repository.ShardRepository
	cfg           RouterConfig
}

func NewRouterService(
	shardKeysRepo *repository.ShardKeysRepository,
	shardRepo *repository.ShardRepository,
	cfg RouterConfig,
) *RouterService {
	return &RouterService{
		shardKeysRepo: shardKeysRepo,
		shardRepo:     shardRepo,
		cfg:           cfg,
	}
}

// RouteSQL is the router service entry point
func (s *RouterService) RouteSQL(
	ctx context.Context,
	projectID string,
	sql string,
) (*RoutingPlan, error) {

	logger.Logger.Info("router entry reached")

	// 1. Parse SQL
	parseResult, err := pg_query.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("sql parse error: %w", err)
	}

	if len(parseResult.Stmts) != 1 {
		return nil, fmt.Errorf("only single-statement queries supported")
	}

	rawStmt := parseResult.Stmts[0]
	node := rawStmt.Stmt

	// detect joins
	var joinInfo *JoinInfo
	var isJoin bool

	if selectNode, ok := node.Node.(*pg_query.Node_SelectStmt); ok {
		joinInfo, isJoin = ExtractJoinInfo(selectNode.SelectStmt)
	}

	// 2. Fetch shard keys for project
	shardKeys, err := s.shardKeysRepo.FetchShardKeysByProjectID(
		ctx,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	// Build table → shard key map
	shardKeyMap := make(map[string]string)
	for _, k := range shardKeys {
		shardKeyMap[k.TableName] = k.ShardKeyColumn
	}

	// handle join queries
	if isJoin {

		leftKey, ok1 := shardKeyMap[joinInfo.LeftTable]
		rightKey, ok2 := shardKeyMap[joinInfo.RightTable]

		// fmt.Println("left key: ", leftKey, "Right key:", rightKey)

		if !ok1 || !ok2 {
			return &RoutingPlan{
				Mode:   RoutingModeRejected,
				Reason: "shard key missing for join tables",
				RejectError: &RoutingError{
					Code:    ErrNoShardKey,
					Message: "missing shard key for join tables",
				},
			}, nil
		}

		// Check colocated join condition
		if joinInfo.LeftColumn == leftKey && joinInfo.RightColumn == rightKey {

			// Fetch shards
			shards, err := s.shardRepo.ShardList(ctx, projectID)
			if err != nil {
				return nil, err
			}

			targets := make([]ShardTarget, 0)

			for _, sh := range shards {
				if sh.Status == "active" {
					targets = append(targets, ShardTarget{
						ShardID: ShardID(sh.ID),
					})
				}
			}

			return &RoutingPlan{
				Mode:    RoutingModeBroadcast,
				Targets: targets,
				Reason:  "colocated join detected",
			}, nil
		}

		// Non-colocated joins not supported yet
		return &RoutingPlan{
			Mode:   RoutingModeRejected,
			Reason: "non-colocated joins not supported",
			RejectError: &RoutingError{
				Code:    ErrUnsupportedPredicate,
				Message: "non colocated joins not supported",
			},
		}, nil
	}

	// non join query flow
	tableName, _, err := extractTableAndNode(rawStmt)
	if err != nil {
		return nil, err
	}

	shardKeyColumn, ok := shardKeyMap[tableName]
	if !ok {
		return &RoutingPlan{
			Mode: RoutingModeRejected,
			Reason: fmt.Sprintf(
				"no shard key defined for table %s",
				tableName,
			),
			RejectError: &RoutingError{
				Code:    ErrNoShardKey,
				Message: "shard key not found",
			},
		}, nil
	}

	// Fetch shards
	shards, err := s.shardRepo.ShardList(ctx, projectID)
	if err != nil {
		return nil, err
	}

	activeShards := make([]repository.Shard, 0)
	for _, sh := range shards {
		if sh.Status == "active" {
			activeShards = append(activeShards, sh)
		}
	}

	if len(activeShards) == 0 {
		return nil, fmt.Errorf("no active shards for project")
	}

	sort.Slice(activeShards, func(i, j int) bool {
		return activeShards[i].ShardIndex < activeShards[j].ShardIndex
	})

	shardIDs := make([]ShardID, 0, len(activeShards))
	for _, sh := range activeShards {
		shardIDs = append(shardIDs, ShardID(sh.ID))
	}

	ring := NewRing(shardIDs)
	hasher := NewHasher()

	planner := NewPlanner(
		s.cfg,
		hasher,
		ring,
	)

	plan := planner.Plan(
		*node,
		tableName,
		shardKeyColumn,
	)

	return plan, nil
}
func extractTableAndNode(
	stmt *pg_query.RawStmt,
) (string, *pg_query.Node, error) {

	node := stmt.Stmt

	switch n := node.Node.(type) {

	case *pg_query.Node_SelectStmt:
		from := n.SelectStmt.FromClause
		if len(from) != 1 {
			return "", nil, fmt.Errorf("joins not supported in v1")
		}
		rv := from[0].Node.(*pg_query.Node_RangeVar)
		return rv.RangeVar.Relname, node, nil

	case *pg_query.Node_InsertStmt:
		return n.InsertStmt.Relation.Relname, node, nil

	case *pg_query.Node_UpdateStmt:
		return n.UpdateStmt.Relation.Relname, node, nil

	case *pg_query.Node_DeleteStmt:
		return n.DeleteStmt.Relation.Relname, node, nil

	default:
		return "", nil, fmt.Errorf("unsupported statement type")
	}
}
