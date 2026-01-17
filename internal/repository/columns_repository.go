package repository

import (
	"context"
	"database/sql"
)

// represents columns table
type Columns struct {
	ProjectID    string `json:"project_id"`
	TableName    string `json:"table_name"`
	ColumnName   string `json:"column_name"`
	DataType     string `json:"data_type"`
	Nullable     bool   `json:"nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
}

// columns as db
type ColumnRepository struct {
	cols *sql.DB
}

// constructor for columns repository
func NewColumnsRepository(cols *sql.DB) *ColumnRepository {
	return &ColumnRepository{
		cols: cols,
	}
}

// func to fetch all columns of a project
func (c *ColumnRepository) GetColumnsByProjectID(ctx context.Context, projectID string) ([]Columns, error) {

	query := `
		SELECT 
			project_id, table_name, column_name, data_type, nullable, is_primary_key
		FROM columns
		WHERE project_id = $1
	`
	rows, err := c.cols.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var result []Columns
	var temp Columns

	for rows.Next() {

		err := rows.Scan(
			&temp.ProjectID,
			&temp.TableName,
			&temp.ColumnName,
			&temp.DataType,
			&temp.Nullable,
			&temp.IsPrimaryKey,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, temp)

	}

	defer rows.Close()

	return result, nil

}

// func to add new column of a project
func (c *ColumnRepository) AddProjectColumn(
	ctx context.Context, projectID string, tableName string,
	columnName string, dataType string, nullable bool, isPK bool,
) error {

	query := `
		INSERT INTO columns 
		(project_id, table_name, column_name, data_type, nullable, is_primary_key)
		VALUES
		($1, $2, $3, $4, $5, $6)
	`

	_, err := c.cols.ExecContext(
		ctx,
		query,
		projectID,
		tableName,
		columnName,
		dataType,
		nullable,
		isPK,
	)
	if err != nil {
		return err
	}

	return nil

}

// func to update existing columns for a project
func (c *ColumnRepository) ReplaceExistingColumns(
	ctx context.Context,
	projectID string,
	cols []Columns,
) error {

	tx, err := c.cols.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM columns WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO columns
			(project_id, table_name, column_name, data_type, nullable, is_primary_key)
		VALUES
			($1, $2, $3, $4, $5, $6)
	`

	for _, col := range cols {
		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			col.ProjectID,
			col.TableName,
			col.ColumnName,
			col.DataType,
			col.Nullable,
			col.IsPrimaryKey,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
