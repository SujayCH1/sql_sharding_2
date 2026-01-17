package repository

import (
	"context"
	"database/sql"
)

// represents table_shard_keys table
type TableShardKey struct {
	ProjectID        string `json:"project_id"`
	TableName        string `json:"table_name"`
	ShardKeyColumn   string `json:"shard_key_column"`
	IsManualOverride bool   `json:"is_manual_override"`
	UpdatedAt        string `json:"updated_at"`
}

type TableShardKeyRepository struct {
	db *sql.DB
}

func NewTableShardKeyRepository(db *sql.DB) *TableShardKeyRepository {
	return &TableShardKeyRepository{db: db}
}

func (r *TableShardKeyRepository) GetShardKey(
	ctx context.Context,
	projectID string,
	tableName string,
) (*TableShardKey, error) {

	query := `
		SELECT project_id, table_name, shard_key_column, is_manual_override, updated_at
		FROM table_shard_keys
		WHERE project_id = $1 AND table_name = $2
	`

	row := r.db.QueryRowContext(ctx, query, projectID, tableName)

	var sk TableShardKey
	err := row.Scan(
		&sk.ProjectID,
		&sk.TableName,
		&sk.ShardKeyColumn,
		&sk.IsManualOverride,
		&sk.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &sk, nil
}

func (r *TableShardKeyRepository) UpsertShardKey(
	ctx context.Context,
	projectID string,
	tableName string,
	shardKeyColumn string,
	isManualOverride bool,
) error {

	query := `
		INSERT INTO table_shard_keys
			(project_id, table_name, shard_key_column, is_manual_override)
		VALUES
			($1, $2, $3, $4)
		ON CONFLICT (project_id, table_name)
		DO UPDATE SET
			shard_key_column = EXCLUDED.shard_key_column,
			is_manual_override = EXCLUDED.is_manual_override,
			updated_at = now()
		WHERE table_shard_keys.is_manual_override = false
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		projectID,
		tableName,
		shardKeyColumn,
		isManualOverride,
	)

	return err
}

func (r *TableShardKeyRepository) GetShardKeysByProject(
	ctx context.Context,
	projectID string,
) ([]TableShardKey, error) {

	query := `
		SELECT project_id, table_name, shard_key_column, is_manual_override, updated_at
		FROM table_shard_keys
		WHERE project_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TableShardKey

	for rows.Next() {
		var sk TableShardKey
		if err := rows.Scan(
			&sk.ProjectID,
			&sk.TableName,
			&sk.ShardKeyColumn,
			&sk.IsManualOverride,
			&sk.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, sk)
	}

	return result, nil
}

func (r *TableShardKeyRepository) DeleteShardKey(
	ctx context.Context,
	projectID string,
	tableName string,
) error {

	query := `
		DELETE FROM table_shard_keys
		WHERE project_id = $1 AND table_name = $2
	`

	_, err := r.db.ExecContext(ctx, query, projectID, tableName)
	return err
}
