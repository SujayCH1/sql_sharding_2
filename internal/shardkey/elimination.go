package shardkey

import (
	"strings"

	"sql-sharding-v2/internal/schema"
)

// ExtractCandidates performs HARD elimination only.
// This stage must remove only fundamentally invalid shard keys.
func ExtractCandidates(s *schema.LogicalSchema) CandidateSet {

	candidates := make(CandidateSet)

	for tableName, table := range s.Tables {

		var tableCandidates []ColumnRef

		for _, column := range table.Columns {

			eliminated, _ := isEliminated(column)
			if eliminated {
				continue
			}

			tableCandidates = append(tableCandidates, ColumnRef{
				Table:  tableName,
				Column: column.Name,
			})
		}

		if len(tableCandidates) > 0 {
			candidates[tableName] = tableCandidates
		}
	}

	return candidates
}

// isEliminated applies ONLY structural hard-elimination rules
func isEliminated(col *schema.Column) (bool, string) {

	if isNullable(col) {
		return true, "column is nullable"
	}

	if isTechnicalColumn(col) {
		return true, "technical metadata column"
	}

	if isLowCardinality(col) {
		return true, "low cardinality column"
	}

	return false, ""
}

func isNullable(col *schema.Column) bool {
	return col.Nullable
}

func isLowCardinality(col *schema.Column) bool {
	switch strings.ToLower(col.DataType) {
	case "bool", "boolean":
		return true
	}

	name := strings.ToLower(col.Name)
	if strings.HasPrefix(name, "is_") ||
		strings.Contains(name, "flag") ||
		strings.Contains(name, "status") {
		return true
	}

	return false
}

func isTechnicalColumn(col *schema.Column) bool {
	switch strings.ToLower(col.Name) {
	case "created_at", "updated_at", "deleted_at", "version":
		return true
	}
	return false
}
