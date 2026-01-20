package repository

import (
	"context"
	"database/sql"
	"time"
)

// ShardKeys represents a row in table_shard_keys (read model)
type ShardKeys struct {
	ProjectID        string    `json:"project_id"`
	TableName        string    `json:"table_name"`
	ShardKeyColumn   string    `json:"shard_key_column"`
	IsManualOverride bool      `json:"is_manual_override"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ShardKeyRecord is a write-only DTO used for inserts/replacements
// avoids importing inference/service packages.
type ShardKeyRecord struct {
	TableName      string
	ShardKeyColumn string
	IsManual       bool
}

// repository
type ShardKeysRepository struct {
	db *sql.DB
}

// constructor
func NewShardKeysRepository(db *sql.DB) *ShardKeysRepository {
	return &ShardKeysRepository{
		db: db,
	}
}

// fetches all shard keys for a project
func (s *ShardKeysRepository) FetchShardKeysByProjectID(
	ctx context.Context,
	projectID string,
) ([]ShardKeys, error) {

	query := `
		SELECT
			project_id,
			table_name,
			shard_key_column,
			is_manual_override,
			updated_at
		FROM table_shard_keys
		WHERE project_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []ShardKeys

	for rows.Next() {
		var key ShardKeys
		if err := rows.Scan(
			&key.ProjectID,
			&key.TableName,
			&key.ShardKeyColumn,
			&key.IsManualOverride,
			&key.UpdatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// inserts a single shard key row
func (s *ShardKeysRepository) AddShardKey(
	ctx context.Context,
	projectID string,
	record ShardKeyRecord,
) error {

	query := `
		INSERT INTO table_shard_keys
		(project_id, table_name, shard_key_column, is_manual_override, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		projectID,
		record.TableName,
		record.ShardKeyColumn,
		record.IsManual,
		time.Now(),
	)

	return err
}

// deletes all shard keys for a project
func (s *ShardKeysRepository) DeleteShardKeysByProjectID(
	ctx context.Context,
	projectID string,
) error {

	query := `
		DELETE FROM table_shard_keys
		WHERE project_id = $1
	`

	_, err := s.db.ExecContext(ctx, query, projectID)
	return err
}

// replaces inferred shard keys atomically.
func (s *ShardKeysRepository) ReplaceShardKeysForProject(
	ctx context.Context,
	projectID string,
	records []ShardKeyRecord,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery := `
		DELETE FROM table_shard_keys
		WHERE project_id = $1
		  AND is_manual_override = FALSE
	`
	if _, err := tx.ExecContext(ctx, deleteQuery, projectID); err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO table_shard_keys
		(project_id, table_name, shard_key_column, is_manual_override, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()

	for _, r := range records {
		if _, err := tx.ExecContext(
			ctx,
			insertQuery,
			projectID,
			r.TableName,
			r.ShardKeyColumn,
			r.IsManual,
			now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
