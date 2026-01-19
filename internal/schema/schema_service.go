package schema

import (
	"context"
	"sql-sharding-v2/internal/repository"
)

type SchemaService struct {
	columnRepo *repository.ColumnRepository
	fkRepo     *repository.FKEdgesRepository
}

func NewSchemaService(
	colRepo *repository.ColumnRepository,
	fkRepo *repository.FKEdgesRepository,
) *SchemaService {
	return &SchemaService{
		columnRepo: colRepo,
		fkRepo:     fkRepo,
	}
}

func (s *SchemaService) ApplyDDLAndRecomputeShardKeys(
	ctx context.Context,
	projectID string,
	ddl string,
) error {

	// STEP 1 — DDL → delta schema
	deltaSchema, err := BuildLogicalSchemaFromDDL(ctx, ddl)
	if err != nil {
		return err
	}

	// STEP 2 — fetch existing metadata
	columns, err := s.columnRepo.GetColumnsByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	fkEdges, err := s.fkRepo.GetEdgesByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// STEP 2 — metadata → base schema
	baseSchema, err := BuildLogicalSchemaFromMetadata(
		projectID,
		columns,
		fkEdges,
	)
	if err != nil {
		return err
	}

	// STEP 3 — merge schemas
	mergedSchema, err := MergeLogicalSchema(baseSchema, deltaSchema)
	if err != nil {
		return err
	}

	// STEP 4 — flatten schema
	newColumns, newFKEdges, err := FlattenLogicalSchema(mergedSchema)
	if err != nil {
		return err
	}

	// STEP 5a — replace metadata atomically
	if err := s.columnRepo.ReplaceExistingColumns(ctx, projectID, newColumns); err != nil {
		return err
	}

	if err := s.fkRepo.ReplaceFKEdgesForProject(ctx, projectID, newFKEdges); err != nil {
		return err
	}

	return nil
}
