package shardkey

import (
	"fmt"
	"sort"
	"strings"

	"sql-sharding-v2/internal/schema"
)

// RankTableCandidates ranks shard key candidates for a single table
// using fanout + ownership + root affinity + identity + content signals.
func RankTableCandidates(
	tableName string,
	local []ColumnRef,
	fanout map[ColumnRef]FanoutStats,
	s *schema.LogicalSchema,
) []RankedCandidate {

	var ranked []RankedCandidate

	table := s.Tables[tableName]

	for _, col := range local {

		stats, ok := fanout[col]
		if !ok {
			stats = FanoutStats{}
		}

		column := table.Columns[col.Column]

		score, reasons := scoreColumn(col, column, stats, table, fanout)

		ranked = append(ranked, RankedCandidate{
			Column:  col,
			Score:   score,
			Reasons: reasons,
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return tieBreak(ranked[i].Column, ranked[j].Column)
	})

	return ranked
}

// scoreColumn computes the final ranking score for a column.
func scoreColumn(
	col ColumnRef,
	column *schema.Column,
	stats FanoutStats,
	table *schema.Table,
	fanout map[ColumnRef]FanoutStats,
) (int, []string) {

	score := 0
	var reasons []string

	if stats.IncomingFKs > 0 {
		value := stats.IncomingFKs * 10
		score += value
		reasons = append(reasons,
			fmt.Sprintf("referenced by %d foreign keys", stats.IncomingFKs),
		)
	}

	if stats.ReferencingTables > 0 {
		value := stats.ReferencingTables * 5
		score += value
		reasons = append(reasons,
			fmt.Sprintf("shared across %d tables", stats.ReferencingTables),
		)
	}

	if isForeignKey(col.Column, table) {
		score += 20
		reasons = append(reasons, "foreign key (ownership column)")

		if bonus, reason := rootAffinityBonus(col.Column, table, fanout); bonus > 0 {
			score += bonus
			reasons = append(reasons, reason)
		}
	}

	if column.IsPrimaryKey {
		score += 10
		reasons = append(reasons, "primary key (identity column)")
	}

	if isTextualColumn(column) {
		score -= 15
		reasons = append(reasons, "textual/content column")
	}

	score += 1
	reasons = append(reasons, "local column")

	return score, reasons
}

// rootAffinityBonus prefers FKs that point to root tables
// (tables with high incoming fanout).
func rootAffinityBonus(
	columnName string,
	table *schema.Table,
	fanout map[ColumnRef]FanoutStats,
) (int, string) {

	for _, fk := range table.FKs {
		if fk.ChildColumn != columnName {
			continue
		}

		parent := ColumnRef{
			Table:  fk.ParentTable,
			Column: fk.ParentColumn,
		}

		stats, ok := fanout[parent]
		if !ok {
			return 0, ""
		}

		if stats.IncomingFKs > 0 {
			bonus := stats.IncomingFKs * 5
			return bonus, fmt.Sprintf(
				"points to root table (%d incoming references)",
				stats.IncomingFKs,
			)
		}
	}

	return 0, ""
}

// isForeignKey checks if a column participates as a FK in its table.
func isForeignKey(columnName string, table *schema.Table) bool {
	for _, fk := range table.FKs {
		if fk.ChildColumn == columnName {
			return true
		}
	}
	return false
}

// isTextualColumn penalizes free-text / descriptive fields.
func isTextualColumn(col *schema.Column) bool {
	switch strings.ToLower(col.DataType) {
	case "text", "varchar", "char", "character varying":
		return true
	}
	return false
}

// tieBreak ensures deterministic ordering when scores are equal.
func tieBreak(a, b ColumnRef) bool {
	if a.Table != b.Table {
		return a.Table < b.Table
	}
	return a.Column < b.Column
}
