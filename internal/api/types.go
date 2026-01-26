package api

type ExecuteQueryRequest struct {
	ProjectID string `json:"project_id"`
	SQL       string `json:"sql"`
}

type ShardResultResponse struct {
	ShardID      string   `json:"shard_id"`
	Columns      []string `json:"columns,omitempty"`
	Rows         [][]any  `json:"rows,omitempty"`
	RowsAffected int64    `json:"rows_affected,omitempty"`
	Error        string   `json:"error,omitempty"`
}

type ExecuteQueryResponse struct {
	Results []ShardResultResponse `json:"results"`
}
