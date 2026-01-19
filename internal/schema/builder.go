package schema

import (
	"context"
	"errors"
	"sql-sharding-v2/internal/repository"
)

// func to convert raw sql query to logical schema
func BuildLogicalSchemaFromDDL(ctx context.Context, sql string) (*LogicalSchema, error) {

	logicalSchemaInst := NewLogicalSchema()

	ast, err := parseDDL(sql)
	if err != nil {
		return nil, err
	}

	err = extractSchemaFromAST(ast, logicalSchemaInst)
	if err != nil {
		return nil, err
	}

	return logicalSchemaInst, nil
}

// func to convert columns and fk repository data to logical schema
func BuildLogicalSchemaFromMetadata(
	projectID string,
	columns []repository.Columns,
	fkEdges []repository.FKEdges,
) (*LogicalSchema, error) {

	schema := NewLogicalSchema()
	schema.ProjectID = projectID

	addColsToLogicalSchema(projectID, schema, columns)
	addFKsToLogicalSchema(schema, fkEdges)

	return schema, nil
}

// func to merge logical schema from metadata and sql query
func MergeLogicalSchema(baseSchema *LogicalSchema, changes *LogicalSchema) (*LogicalSchema, error) {

	mergedSchema := cloneLogicalSchema(baseSchema)

	for tableName, deltaTable := range changes.Tables {

		baseTable, exists := mergedSchema.Tables[tableName]
		if !exists {
			mergedSchema.Tables[tableName] = cloneTable(deltaTable)
			continue
		}

		mergedSchema.Tables[tableName] = mergeTable(baseTable, deltaTable)
	}

	return mergedSchema, nil

}

// FlattenLogicalSchema converts a LogicalSchema into metadata tables
func FlattenLogicalSchema(schema *LogicalSchema) ([]repository.Columns, []repository.FKEdges, error) {

	if schema == nil {
		return nil, nil, errors.New("nil columns or fk_edges")
	}

	columns := FlattenColumns(schema)
	fkEdges := FlattenFKEdges(schema)

	return columns, fkEdges, nil
}
