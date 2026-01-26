package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// represents the shard table in database
type Shard struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	ShardIndex int       `json:"shard_index"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// a 'shard' with a 'connection'
type ShardRepository struct {
	shd *sql.DB
}

// unique error case to handle delete shard when active case
var ErrShardDeleteBlocked = errors.New("shard_delete_blocked")

// return an instance of 'shard' binded with connection infomation received in args
func NewShardRepository(shd *sql.DB) *ShardRepository {
	return &ShardRepository{shd: shd}
}

func (s *ShardRepository) ShardAdd(ctx context.Context, projectID string) (*Shard, error) {
	var shard Shard

	indexes, err := s.FetchShardIndexes(ctx, projectID)
	if err != nil {
		return nil, err
	}

	newIndex := getIndex(indexes)

	shard.ID = uuid.New().String()
	shard.ProjectID = projectID
	shard.ShardIndex = newIndex
	shard.Status = "inactive"

	query := `
		INSERT INTO shards (id, project_id, shard_index, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err = s.shd.QueryRowContext(
		ctx,
		query,
		shard.ID,
		shard.ProjectID,
		shard.ShardIndex,
		shard.Status,
	).Scan(&shard.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &shard, nil
}

// function to fetch all shards from a project
func (s *ShardRepository) ShardList(ctx context.Context, projectID string) ([]Shard, error) {
	query := `
		SELECT id, project_id, shard_index, status, created_at FROM shards
		WHERE project_id = $1
		ORDER BY shard_index
	`

	rows, err := s.shd.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shards := make([]Shard, 0)

	for rows.Next() {
		var shard Shard
		err := rows.Scan(
			&shard.ID,
			&shard.ProjectID,
			&shard.ShardIndex,
			&shard.Status,
			&shard.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		shards = append(shards, shard)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shards, nil
}

// fuction to delete all shards of a project
func (s *ShardRepository) ShardDeleteAll(ctx context.Context, projectID string) error {
	query := `
		DELETE FROM shards WHERE project_id = $1
	`
	result, err := s.shd.ExecContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil

}

// funcion to delete a single shard of a project
func (s *ShardRepository) ShardDelete(ctx context.Context, shardID string) error {
	var status string

	err := s.shd.QueryRowContext(
		ctx,
		`SELECT status FROM shards WHERE id = $1`,
		shardID,
	).Scan(&status)
	if err != nil {
		return err
	}

	if status == "active" {
		return ErrShardDeleteBlocked
	}

	_, err = s.shd.ExecContext(
		ctx,
		`DELETE FROM shards WHERE id = $1`,
		shardID,
	)

	return err
}

// func to deactivate a single shard
func (s *ShardRepository) ShardDeactivate(ctx context.Context, shardID string) error {
	query := `
		UPDATE shards SET status = 'inactive' WHERE id = $1
	`

	_, err := s.shd.ExecContext(
		ctx,
		query,
		shardID,
	)

	if err != nil {
		return err
	}

	return nil
}

// func to activate a single shard
func (s *ShardRepository) ShardActivate(ctx context.Context, shardID string) error {
	query := `
		UPDATE shards SET status = 'active' WHERE id = $1
	`
	_, err := s.shd.ExecContext(
		ctx,
		query,
		shardID,
	)

	if err != nil {
		return err
	}

	return nil

}

// func to fecth status of a shard using its id
func (s *ShardRepository) FetchShardStatus(ctx context.Context, shardID string) (string, error) {
	query := `
		SELECT status FROM shards WHERE id = $1
	`

	rows := s.shd.QueryRowContext(
		ctx,
		query,
		shardID,
	)

	var status string

	err := rows.Scan(&status)

	if err != nil {
		return "", err
	}

	return status, nil
}

// helper to fetch indexes from database (PROJECT SCOPED)
func (s *ShardRepository) FetchShardIndexes(ctx context.Context, projectID string) ([]int, error) {
	query := `
		SELECT shard_index FROM shards WHERE project_id = $1
	`

	rows, err := s.shd.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexList []int

	for rows.Next() {
		var index int
		err := rows.Scan(&index)
		if err != nil {
			return nil, err
		}

		indexList = append(indexList, index)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexList, nil
}

// func to fetch project ID using shard
func (s *ShardRepository) FetchProjectID(ctx context.Context, shardID string) (string, error) {

	query := `
		SELECT project_id
		FROM shards
		WHERE id = $1
	`

	row := s.shd.QueryRowContext(
		ctx,
		query,
		shardID,
	)

	var projectID string

	err := row.Scan(
		&projectID,
	)
	if err != nil {
		return "", err
	}

	return projectID, nil

}

// helper to calculate shard index
func getIndex(arr []int) int {
	if len(arr) == 0 {
		return 0
	}

	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max + 1
}
