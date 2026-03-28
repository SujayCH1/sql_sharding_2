package repository

import (
	"context"
	"database/sql"
)

type AIConfig struct {
	ProjectID string
	Provider  string
	APIKey    string
	Model     string
}

type AIConfigRepository struct {
	db *sql.DB
}

func NewAIConfigRepository(db *sql.DB) *AIConfigRepository {
	return &AIConfigRepository{db: db}
}

func (r *AIConfigRepository) UpsertConfig(ctx context.Context, config AIConfig) error {

	query := `
    INSERT INTO ai_config (project_id, provider, api_key, model)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (project_id)
    DO UPDATE SET
        provider = EXCLUDED.provider,
        api_key = EXCLUDED.api_key,
        model = EXCLUDED.model,
        updated_at = NOW()
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		config.ProjectID,
		config.Provider,
		config.APIKey,
		config.Model,
	)

	return err
}

func (r *AIConfigRepository) GetConfigByProjectID(ctx context.Context, projectID string) (*AIConfig, error) {

	query := `
    SELECT project_id, provider, api_key, model
    FROM ai_config
    WHERE project_id = $1
    `

	row := r.db.QueryRowContext(ctx, query, projectID)

	var config AIConfig
	err := row.Scan(
		&config.ProjectID,
		&config.Provider,
		&config.APIKey,
		&config.Model,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}
