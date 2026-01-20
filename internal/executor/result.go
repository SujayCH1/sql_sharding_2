package executor

type ExecutionResult struct {
	ShardID      string
	Columns      []string
	Rows         [][]any
	RowsAffected int64
	Err          error
}
