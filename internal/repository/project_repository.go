package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// represents the project table in the database
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ShardCount  int    `json:"shard_count"`
	Status      bool   `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// a 'database' with a 'connection'
type ProjectRepository struct {
	db *sql.DB
}

// return an instance of 'database' binded with connection infomation received in args
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// execute query to add a poject row
func (r *ProjectRepository) ProjectAdd(ctx context.Context, name string,
	descriptrion string) (*Project, error) {

	newID := uuid.New()

	convertedID := newID.String()

	project := &Project{
		ID:          convertedID,
		Name:        name,
		Description: descriptrion,
		ShardCount:  0,
		Status:      false,
		CreatedAt:   time.Now().String(),
	}

	query := `
		INSERT INTO projects (id, name, description, status,  shard_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		project.ID,
		project.Name,
		project.Description,
		project.Status,
		project.ShardCount,
		time.Now(),
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

// return row of all project in the porjects table
func (r *ProjectRepository) ProjectList(ctx context.Context) ([]Project, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, description, shard_count, status, created_at
		 FROM projects
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]Project, 0)

	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.ShardCount,
			&p.Status,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// remove a project from the projects table
func (r *ProjectRepository) ProjectRemove(ctx context.Context, id string) error {

	projectID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	query := `
		DELETE FROM projects WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		projectID,
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

// reterive asingle project
func (r *ProjectRepository) GetProjectByID(ctx context.Context, id string) (Project, error) {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return Project{}, err
	}

	query := `
		SELECT id, name, description, shard_count, status, created_at
		FROM projects
		WHERE id = $1
	`

	var p Project

	err = r.db.QueryRowContext(ctx, query, projectID).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.ShardCount,
		&p.Status,
		&p.CreatedAt,
	)

	if err != nil {
		return Project{}, err
	}

	return p, nil
}
