package shardkey

// Identifies a column uniquely across schema
type ColumnRef struct {
	Table  string
	Column string
}

// Hard-elimination output
type CandidateSet map[string][]ColumnRef

// table_name -> candidate columns

// Fanout statistics for a column
type FanoutStats struct {
	IncomingFKs       int
	ReferencingTables int
}

// Ranked candidate for a table
type RankedCandidate struct {
	Column  ColumnRef
	Score   int
	Reasons []string
}

// Final inference result
type ShardKeyDecision struct {
	Table   string
	Column  ColumnRef
	Score   int
	Reasons []string
}

// Whole-project output
type InferenceResult struct {
	ProjectID string
	Decisions []ShardKeyDecision
}
