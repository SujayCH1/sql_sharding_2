package shardkey

import (
	"sql-sharding-v2/internal/schema"
)

// ComputeFanout computes incoming fanout statistics for each candidate column.
// Fanout is based ONLY on direct foreign key references.
func ComputeFanout(s *schema.LogicalSchema, candidates CandidateSet) map[ColumnRef]FanoutStats {

	candidateIndex := indexCandidates(candidates)

	fanout := make(map[ColumnRef]FanoutStats)

	seenTables := make(map[ColumnRef]map[string]struct{})

	for _, table := range s.Tables {
		for _, fk := range table.FKs {

			parentCol := ColumnRef{
				Table:  fk.ParentTable,
				Column: fk.ParentColumn,
			}

			if _, ok := candidateIndex[parentCol]; !ok {
				continue
			}

			stats := fanout[parentCol]
			stats.IncomingFKs++

			if _, ok := seenTables[parentCol]; !ok {
				seenTables[parentCol] = make(map[string]struct{})
			}

			if _, seen := seenTables[parentCol][fk.ChildTable]; !seen {
				seenTables[parentCol][fk.ChildTable] = struct{}{}
				stats.ReferencingTables++
			}

			fanout[parentCol] = stats
		}
	}

	return fanout
}

// indexCandidates flattens CandidateSet into a lookup map
func indexCandidates(candidates CandidateSet) map[ColumnRef]struct{} {

	index := make(map[ColumnRef]struct{})

	for tableName, cols := range candidates {
		for _, col := range cols {
			index[ColumnRef{
				Table:  tableName,
				Column: col.Column,
			}] = struct{}{}
		}
	}

	return index
}
