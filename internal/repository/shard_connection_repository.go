package repository

import (
	"context"
	"database/sql"
)

// reporesents the shard_connection table in database
type ShardConnection struct {
	ShardID      string `json:"shard_id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// instance of a shard connection repo
type ShardConnectionRepository struct {
	shdConn *sql.DB
}

// returns an instance of sharConnRepo binded with incoming *sql.db
func NewShardConnectionRepository(shdConn *sql.DB) *ShardConnectionRepository {
	return &ShardConnectionRepository{shdConn: shdConn}
}

// used to add connection infomation fo a shard
func (c *ShardConnectionRepository) ConnectionCreate(ctx context.Context, conn *ShardConnection) error {

	query := `
        INSERT INTO shard_connections (shard_id, host, port, database_name, username, password)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := c.shdConn.ExecContext(
		ctx,
		query,
		conn.ShardID,
		conn.Host,
		conn.Port,
		conn.DatabaseName,
		conn.Username,
		conn.Password,
	)

	return err
}

// func to remove connection details of a single shard
func (c *ShardConnectionRepository) ConnectionRemove(ctx context.Context, shardID string) error {

	query := `
		DELETE FROM shard_connections WHERE shard_id = $1
	`

	_, err := c.shdConn.ExecContext(
		ctx,
		query,
		shardID,
	)

	if err != nil {
		return err
	}

	return nil

}

// Fetch connection info of a shard by shard_id
func (c *ShardConnectionRepository) FetchConnectionByShardID(ctx context.Context, shardID string) (ShardConnection, error) {

	query := `
        SELECT
            shard_id,
            host,
            port,
            database_name,
            username,
            password,
            created_at,
            updated_at
        FROM shard_connections
        WHERE shard_id = $1
    `

	var conn ShardConnection

	err := c.shdConn.QueryRowContext(
		ctx,
		query,
		shardID,
	).Scan(
		&conn.ShardID,
		&conn.Host,
		&conn.Port,
		&conn.DatabaseName,
		&conn.Username,
		&conn.Password,
		&conn.CreatedAt,
		&conn.UpdatedAt,
	)

	if err != nil {
		return ShardConnection{}, err
	}

	return conn, nil
}

// func to update existing connection details
func (c *ShardConnectionRepository) ConnectionUpdate(ctx context.Context, connInfo ShardConnection) error {

	query := `
		UPDATE shard_connections
		SET 
		host = $1,
		port = $2,
		database_name = $3,
		username = $4,
		password = $5,
		updated_at = NOW()
		WHERE shard_id = $6
	`

	_, err := c.shdConn.ExecContext(
		ctx,
		query,
		connInfo.Host,
		connInfo.Port,
		connInfo.DatabaseName,
		connInfo.Username,
		connInfo.Password,
		connInfo.ShardID,
	)

	if err != nil {
		return err
	}

	return nil

}
