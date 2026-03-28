package shardkey

import (
	"fmt"
)

func buildPrompt(schemaJson []byte) string {
	return fmt.Sprintf(`
You are an expert in distributed database sharding and query optimization.

Analyze the following database schema and select the optimal shard key for each table.

SCHEMA:
%s

RULES (follow in priority order):
1. If a table has a foreign key referencing another table's shard key, use that same column (co-location)
2. Prefer high-cardinality columns (user_id, order_id, tenant_id) over low-cardinality ones (status, type, bool)
3. Prefer columns that appear most in JOIN conditions
4. For junction/mapping tables, use the foreign key that points to the largest parent table
5. Never pick nullable columns as shard keys
6. Never pick timestamp or created_at columns unless no better option exists

OUTPUT FORMAT:
Return ONLY a valid JSON object. No explanation, no markdown, no extra text.

{
  "decisions": [
    {"table": "table_name", "column": "column_name"}
  ]
}
`, string(schemaJson))
}
