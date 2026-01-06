package schema

import (
	"context"
	"errors"

	"sql-sharding-v2/internal/repository"
)

func ExecuteProjectSchema(
	ctx context.Context,
	projectID string,
	schemaRepo *repository.ProjectSchemaRepository,
	shardRepo *repository.ShardRepository,
	execRepo *repository.SchemaExecutionStatusRepository,
	execDDL func(shardID string, ddl string) error,
) error {

	schema, err := schemaRepo.ProjectSchemaGetPending(ctx, projectID)
	if err != nil {
		return err
	}

	if err := schemaRepo.ProjectSchemaSetApplying(ctx, schema.ID); err != nil {
		return err
	}

	shards, err := shardRepo.ShardList(ctx, projectID)
	if err != nil {
		return err
	}

	for _, shard := range shards {

		if err := execRepo.ExecutionRecordsCreateRecord(
			ctx,
			repository.SchemaExecutionStatus{
				SchemaID: schema.ID,
				ShardID:  shard.ID,
				State:    "pending",
			},
		); err != nil {
			return err
		}

		if shard.Status != "active" {
			msg := "shard inactive"
			_ = execRepo.ExxecutionRecordsUpdateState(
				ctx,
				schema.ID,
				shard.ID,
				"failed",
				&msg,
			)

			_ = schemaRepo.ProjectSchemaUpdateSchemaState(
				ctx,
				schema.ID,
				"failed",
				&msg,
			)

			return errors.New(msg)
		}

		if err := execDDL(shard.ID, schema.DDL_SQL); err != nil {
			msg := err.Error()

			_ = execRepo.ExxecutionRecordsUpdateState(
				ctx,
				schema.ID,
				shard.ID,
				"failed",
				&msg,
			)

			_ = schemaRepo.ProjectSchemaUpdateSchemaState(
				ctx,
				schema.ID,
				"failed",
				&msg,
			)

			return err
		}

		_ = execRepo.ExxecutionRecordsUpdateState(
			ctx,
			schema.ID,
			shard.ID,
			"applied",
			nil,
		)
	}

	return schemaRepo.ProjectSchemaUpdateSchemaState(
		ctx,
		schema.ID,
		"applied",
		nil,
	)
}

func RetryFailedSchema(
	ctx context.Context,
	projectID string,
	schemaRepo *repository.ProjectSchemaRepository,
	execRepo *repository.SchemaExecutionStatusRepository,
) error {

	schema, err := schemaRepo.ProjectSchemaGetLatest(ctx, projectID)
	if err != nil {
		return err
	}

	if schema.State != "failed" {
		return errors.New("schema is not in failed state")
	}

	failedRecords, err :=
		execRepo.ExecutionRecordsFetchStatusFailed(ctx, schema.ID)
	if err != nil {
		return err
	}

	for _, record := range failedRecords {
		if err := execRepo.ExecutionRecordsResetState(
			ctx,
			schema.ID,
			record.ShardID,
		); err != nil {
			return err
		}
	}

	return schemaRepo.ProjectSchemaUpdateSchemaState(
		ctx,
		schema.ID,
		"pending",
		nil,
	)
}
