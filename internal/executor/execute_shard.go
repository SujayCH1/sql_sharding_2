package executor

import (
	"context"
	"database/sql"
)

func executeOnShard(
	ctx context.Context,
	db *sql.DB,
	shardID string,
	sqlText string,
) ExecutionResult {

	// Attempt SELECT first
	rows, err := db.QueryContext(ctx, sqlText)
	if err == nil {
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return ExecutionResult{
				ShardID: shardID,
				Err:     err,
			}
		}

		data := make([][]any, 0)

		for rows.Next() {
			values := make([]any, len(cols))
			ptrs := make([]any, len(cols))

			for i := range values {
				ptrs[i] = &values[i]
			}

			if err := rows.Scan(ptrs...); err != nil {
				return ExecutionResult{
					ShardID: shardID,
					Err:     err,
				}
			}

			data = append(data, values)
		}

		return ExecutionResult{
			ShardID: shardID,
			Columns: cols,
			Rows:    data,
		}
	}

	// Fallback to DML / TX / others
	res, execErr := db.ExecContext(ctx, sqlText)
	if execErr != nil {
		return ExecutionResult{
			ShardID: shardID,
			Err:     execErr,
		}
	}

	affected, _ := res.RowsAffected()

	return ExecutionResult{
		ShardID:      shardID,
		RowsAffected: affected,
	}
}
