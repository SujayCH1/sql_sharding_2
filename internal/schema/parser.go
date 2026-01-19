package schema

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// ParseDDL parses raw SQL text into a Postgres AST.
func parseDDL(sql string) (*pg_query.ParseResult, error) {

	tree, err := pg_query.Parse(sql)
	if err != nil {
		return nil, err
	}

	return tree, nil

}
