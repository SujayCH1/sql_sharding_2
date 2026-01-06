package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SchemaExecutionStatus struct {
	ID         string `json:"id"`
	SchemaID   string `json:"schema_id"`
	ShardID    string `json:"shard_id"`
	State      string `json:"state"`
	ErrMsg     string `json:"error_message"`
	ExecutedAt string `json:"executed_at"`
}

type SchemaExecutionStatusRepository struct {
	schmExeSt *sql.DB
}

func NewSchemaExecutionStatusRepository(schmExeSt *sql.DB) *SchemaExecutionStatusRepository {
	return &SchemaExecutionStatusRepository{
		schmExeSt: schmExeSt,
	}
}

// func to add record of a execution
func (r *SchemaExecutionStatusRepository) ExecutionRecordsCreateRecord(
	ctx context.Context,
	record SchemaExecutionStatus,
) error {

	query := `
		INSERT INTO schema_execution_status
		(id, schema_id, shard_id, state, error_message, executed_at)
		VALUES
		($1, $2, $3, $4, $5, NOW())
	`

	_, err := r.schmExeSt.ExecContext(
		ctx,
		query,
		uuid.New().String(),
		record.SchemaID,
		record.ShardID,
		record.State,
		record.ErrMsg,
	)
	if err != nil {
		return err
	}

	return nil
}

// func to update state of a execution record
func (r *SchemaExecutionStatusRepository) ExxecutionRecordsUpdateState(
	ctx context.Context,
	schemaID string,
	shardID string,
	state string,
	errorMessage *string,
) error {

	query := `
		UPDATE schema_execution_status
		SET state = $1, error_message = $2, executed_at = $3
		WHERE shard_id = $4 AND schema_id = $5
	`

	result, err := r.schmExeSt.ExecContext(
		ctx,
		query,
		state,
		errorMessage,
		time.Now(),
		shardID,
		schemaID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// func to fetch execution status of all shards
func (r *SchemaExecutionStatusRepository) ExecutionRecordsFetchStatusAll(
	ctx context.Context,
	schemaID string,
) ([]SchemaExecutionStatus, error) {

	query := `
		SELECT id, schema_id, shard_id, state, error_message, executed_at
		FROM schema_execution_status
		WHERE schema_id = $1
	`

	rows, err := r.schmExeSt.QueryContext(
		ctx,
		query,
		schemaID,
	)
	if err != nil {
		return nil, err
	}

	var records []SchemaExecutionStatus

	for rows.Next() {
		var record SchemaExecutionStatus

		err = rows.Scan(
			&record.ID,
			&record.SchemaID,
			&record.ShardID,
			&record.State,
			&record.ErrMsg,
			&record.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil

}

// func to fetch failed shard execution
func (r *SchemaExecutionStatusRepository) ExecutionRecordsFetchStatusFailed(
	ctx context.Context,
	schemaID string,
) ([]SchemaExecutionStatus, error) {

	query := `
		SELECT id, schema_id, shard_id, state, error_message, executed_at
		FROM schema_execution_status
		WHERE schema_id = $1 AND state = 'failed'
	`

	rows, err := r.schmExeSt.QueryContext(
		ctx,
		query,
		schemaID,
	)
	if err != nil {
		return nil, err
	}

	var records []SchemaExecutionStatus

	for rows.Next() {
		var record SchemaExecutionStatus

		err = rows.Scan(
			&record.ID,
			&record.SchemaID,
			&record.ShardID,
			&record.State,
			&record.ErrMsg,
			&record.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil

}

// func to reset execution state for retry
func (r *SchemaExecutionStatusRepository) ExecutionRecordsResetState(
	ctx context.Context,
	schemaID string,
	shardID string,
) error {

	query := `
		UPDATE schema_execution_status
		SET state = 'pending', error_message = NULL, executed_at = NULL
		WHERE schema_id = $1 AND shard_id = $2
	`

	result, err := r.schmExeSt.ExecContext(ctx, query, schemaID, shardID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// func to check if all shards are applied correctly
func (r *SchemaExecutionStatusRepository) ExecutionRecordsCheckAppliedAll(
	ctx context.Context,
	schemaID string,
) (bool, error) {

	query := `
		SELECT COUNT(*)
		FROM schema_execution_status
		WHERE schema_id = $1 AND state != 'applied'
	`

	var count int
	err := r.schmExeSt.QueryRowContext(ctx, query, schemaID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
