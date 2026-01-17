package repository

import (
	"context"
	"database/sql"
)

// represents fk_edges table
type FKEdges struct {
	ProjectID    string `json:"project_id"`
	ParentTable  string `json:"parent_table"`
	ParentColumn string `json:"parent_column"`
	ChildTable   string `json:"child_table"`
	ChildColumn  string `json:"child_column"`
}

// FKEdges as db
type FKEdgesRepository struct {
	edg *sql.DB
}

// constructor for fkedges repository
func NewFKEdgesRepository(edg *sql.DB) *FKEdgesRepository {
	return &FKEdgesRepository{
		edg: edg,
	}
}

// func to fetch all edges of a project
func (e *FKEdgesRepository) GetEdgesByProjectID(
	ctx context.Context, projectID string,
) ([]FKEdges, error) {

	query := `
		SELECT
			project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1
	`

	rows, err := e.edg.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var result []FKEdges
	var temp FKEdges

	for rows.Next() {

		err = rows.Scan(
			&temp.ProjectID,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, temp)

	}

	defer rows.Close()

	return result, err

}

// func to fetch all edges where table is child
func (e *FKEdgesRepository) GetEdgesByChildTable(
	ctx context.Context, projectID string, tableName string,
) ([]FKEdges, error) {

	query := `
		SELECT 
			project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1 AND child_table = $2
	`

	rows, err := e.edg.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var result []FKEdges
	var temp FKEdges

	for rows.Next() {

		err = rows.Scan(
			&temp.ProjectID,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, temp)

	}

	defer rows.Close()

	return result, err

}

// func to fetch all edges where table is parent
func (e *FKEdgesRepository) GetEdgesByParentTable(
	ctx context.Context, projectID string, tableName string,
) ([]FKEdges, error) {

	query := `
		SELECT 
			project_id, parent_table, parent_column, child_table, child_column
		FROM fk_edges
		WHERE project_id = $1 AND parent_table = $2
	`

	rows, err := e.edg.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var result []FKEdges
	var temp FKEdges

	for rows.Next() {

		err = rows.Scan(
			&temp.ProjectID,
			&temp.ParentTable,
			&temp.ParentColumn,
			&temp.ChildTable,
			&temp.ChildColumn,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, temp)

	}
	defer rows.Close()

	return result, err

}

// func to add a new edge to database
func (e *FKEdgesRepository) AddFKEdge(
	ctx context.Context, projectID string, parentTable string,
	parentColumn string, childTable string, childColumn string,
) error {

	query := `
		INSERT INTO fk_edges
		(project_id, parent_table, parent_column, child_table, child_column)
		VALUES
		($1, $2, $3, $4, $5)
	`

	_, err := e.edg.ExecContext(
		ctx,
		query,
		projectID,
		parentTable,
		parentColumn,
		childTable,
		childColumn,
	)

	if err != nil {
		return err
	}

	return nil
}

// func to replace exitins edges of a project
func (e *FKEdgesRepository) ReplaceFKEdgesForProject(
	ctx context.Context,
	projectID string,
	edges []FKEdges,
) error {

	tx, err := e.edg.BeginTx(ctx, nil)
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
		`DELETE FROM fk_edges WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO fk_edges
			(project_id, parent_table, parent_column, child_table, child_column)
		VALUES
			($1, $2, $3, $4, $5)
	`

	for _, edge := range edges {
		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			edge.ProjectID,
			edge.ParentTable,
			edge.ParentColumn,
			edge.ChildTable,
			edge.ChildColumn,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
