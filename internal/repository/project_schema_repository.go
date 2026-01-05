package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// represents project_schema table
type ProjectSchema struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Version     int     `json:"version"`
	State       string  `json:"state"`
	DDL_SQL     string  `json:"ddl_sql"`
	ErrMsg      *string `json:"error_message"`
	CreatedAt   string  `json:"created _at"`
	CommittedAt *string `json:"commited_at"`
	AppliedAt   *string `json:"applied_at"`
}

// project_schema as a db
type ProjectSchemaRepository struct {
	projSchm *sql.DB
}

// constructor for ProjectSchema
func NewProjectSchema(projSchm *sql.DB) *ProjectSchemaRepository {
	return &ProjectSchemaRepository{
		projSchm: projSchm,
	}
}

// Creats a draft of a version of a schema
func (p *ProjectSchemaRepository) ProjectSchemaCreateDraft(
	ctx context.Context,
	projectID string,
	ddlSQL string,
) (*ProjectSchema, error) {

	versions, err := p.fetchProjectSchemaVersions(ctx, projectID)
	if err != nil {
		return nil, err
	}

	nextVersion := findMaxVer(versions) + 1
	schemaID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO project_schemas
		(id, project_id, version, state, ddl_sql, created_at)
		VALUES
		($1, $2, $3, $4, $5, $6)
	`

	_, err = p.projSchm.ExecContext(
		ctx,
		query,
		schemaID,
		projectID,
		nextVersion,
		"draft",
		ddlSQL,
		now,
	)
	if err != nil {
		return nil, err
	}

	return &ProjectSchema{
		ID:        schemaID,
		ProjectID: projectID,
		Version:   nextVersion,
		State:     "draft",
		DDL_SQL:   ddlSQL,
		CreatedAt: now.String(),
	}, nil
}

// func to make schema state pending
func (p *ProjectSchemaRepository) ProjectSchemaCommitDraft(
	ctx context.Context,
	schemaID string,
) error {

	query := `
		UPDATE project_schemas
		SET state = 'pending', committed_at = $1
		WHERE id = $2
	`

	result, err := p.projSchm.ExecContext(
		ctx,
		query,
		time.Now(),
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

// func to fetch latest schema of a project
func (p *ProjectSchemaRepository) ProjectSchemaGetLatest(ctx context.Context, projectID string) (*ProjectSchema, error) {

	versions, err := p.fetchProjectSchemaVersions(ctx, projectID)
	if err != nil {
		return nil, err
	}

	maxVerr := findMaxVer(versions)

	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE project_id = $1 AND version = $2
	`

	row := p.projSchm.QueryRowContext(
		ctx,
		query,
		projectID,
		maxVerr,
	)

	var latestSchema ProjectSchema

	err = row.Scan(
		&latestSchema.ID,
		&latestSchema.ProjectID,
		&latestSchema.Version,
		&latestSchema.State,
		&latestSchema.DDL_SQL,
		&latestSchema.ErrMsg,
		&latestSchema.CreatedAt,
		&latestSchema.CommittedAt,
		&latestSchema.AppliedAt,
	)

	if err != nil {
		return nil, err
	}

	return &latestSchema, nil

}

// func to fetch schema history pf a project
func (p *ProjectSchemaRepository) ProjectSchemaFetchHistory(ctx context.Context, projectID string) ([]ProjectSchema, error) {

	query := `
		SELECT 
			id, project_id, version, state, ddl_sql,
			error_message, created_at, committed_at, applied_at
		FROM project_schemas
		WHERE project_id = $1
		ORDER BY version ASC
	`

	rows, err := p.projSchm.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var history []ProjectSchema

	for rows.Next() {
		var schema ProjectSchema

		err := rows.Scan(
			&schema.ID,
			&schema.ProjectID,
			&schema.Version,
			&schema.State,
			&schema.DDL_SQL,
			&schema.ErrMsg,
			&schema.CreatedAt,
			&schema.CommittedAt,
			&schema.AppliedAt,
		)
		if err != nil {
			return nil, err
		}

		history = append(history, schema)
	}

	return history, nil

}

// func to fetch schema using schemaID id
func (p *ProjectSchemaRepository) ProjectSchemaFetchBySchemaID(ctx context.Context, schemaID string) (*ProjectSchema, error) {

	query := `
	SELECT 
		id, project_id, version, state, ddl_sql,
		error_message, created_at, committed_at, applied_at
	FROM project_schemas
	WHERE id = $1
	`

	row := p.projSchm.QueryRowContext(
		ctx,
		query,
		schemaID,
	)

	var schema ProjectSchema

	err := row.Scan(
		&schema.ID,
		&schema.ProjectID,
		&schema.Version,
		&schema.State,
		&schema.DDL_SQL,
		&schema.ErrMsg,
		&schema.CreatedAt,
		&schema.CommittedAt,
		&schema.AppliedAt,
	)

	if err != nil {
		return nil, err
	}

	return &schema, nil

}

// func to update schema state
func (p *ProjectSchemaRepository) ProjectSchemaUpdateSchemaState(
	ctx context.Context,
	schemaID string,
	state string,
	errorMessage *string,
) error {

	query := `
		UPDATE project_schemas
		SET state = $1, error_message = $2
		WHERE id = $3
	`

	_, err := p.projSchm.ExecContext(
		ctx,
		query,
		state,
		errorMessage,
		schemaID,
	)

	if err != nil {
		return err
	}

	return nil
}

// func to get state of a schema
func (p *ProjectSchemaRepository) ProjectSchemaGetState(ctx context.Context, schemaID string) (string, error) {

	query := `
		SELECT state
		FROM project_schemas
		WHERE id = $1
	`

	row := p.projSchm.QueryRowContext(ctx, query, schemaID)

	var state string
	if err := row.Scan(&state); err != nil {
		return "", err
	}

	return state, nil
}

// func to delete drafts
func (p *ProjectSchemaRepository) ProjectSchemaDeleteDraft(ctx context.Context, schemaID string) error {

	query := `
		DELETE FROM project_schemas
		WHERE id = $1 AND state = 'draft'
	`

	result, err := p.projSchm.ExecContext(ctx, query, schemaID)
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

// func to fecth latest applied schema
func (p *ProjectSchemaRepository) ProjectSchemaGetApplied(ctx context.Context, projectID string) (*ProjectSchema, error) {

	query := `
	SELECT 
		id, project_id, version, state, ddl_sql,
		error_message, created_at, committed_at, applied_at
	FROM project_schemas
	WHERE project_id = $1 AND state = 'applied'
	ORDER BY version DESC
	LIMIT 1
	`

	row := p.projSchm.QueryRowContext(
		ctx,
		query,
		projectID,
	)

	var schema ProjectSchema

	err := row.Scan(
		&schema.ID,
		&schema.ProjectID,
		&schema.Version,
		&schema.State,
		&schema.DDL_SQL,
		&schema.ErrMsg,
		&schema.CreatedAt,
		&schema.CommittedAt,
		&schema.AppliedAt,
	)

	if err != nil {
		return nil, err
	}

	return &schema, nil

}

// helper to decide correct version fo schema
func (p *ProjectSchemaRepository) fetchProjectSchemaVersions(ctx context.Context, projectID string) ([]int, error) {

	query := `
		SELECT version FROM project_schemas
		WHERE project_id = $1 
	`

	rows, err := p.projSchm.QueryContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	var versions []int
	var version int

	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil

}

// helper to decide coorect version fo schema
func findMaxVer(versions []int) int {

	if len(versions) == 0 {
		return 0
	}

	maxVer := versions[0]

	for _, ver := range versions {
		if ver > maxVer {
			maxVer = ver
		}
	}

	return maxVer
}
