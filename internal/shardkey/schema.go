package shardkey

import (
	"context"
	"sql-sharding-v2/internal/schema"
)

type LLMSchema struct {
	Tables []LLMTable `json:"tables"`
}

type LLMTable struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	FKs     []LLMFK  `json:"fks"`
}

type LLMFK struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
}

func (s *InferenceService) buildSchema(ctx context.Context, projectID string) (schema.LogicalSchema, error) {

	columns, err := s.columnRepo.GetColumnsByProjectID(ctx, projectID)
	if err != nil {
		return schema.LogicalSchema{}, err
	}

	fkEdges, err := s.fkRepo.GetEdgesByProjectID(ctx, projectID)
	if err != nil {
		return schema.LogicalSchema{}, err
	}

	logicalSchema, err := schema.BuildLogicalSchemaFromMetadata(
		projectID,
		columns,
		fkEdges,
	)
	if err != nil {
		return schema.LogicalSchema{}, err
	}

	return *logicalSchema, nil

}

func buildLLMSchema(s schema.LogicalSchema) LLMSchema {

	var result LLMSchema

	for tableName, table := range s.Tables {

		t := LLMTable{
			Name: tableName,
		}

		for colName := range table.Columns {
			t.Columns = append(t.Columns, colName)
		}

		for _, fk := range table.FKs {
			t.FKs = append(t.FKs, LLMFK{
				FromTable:  fk.ChildTable,
				FromColumn: fk.ChildColumn,
				ToTable:    fk.ParentTable,
				ToColumn:   fk.ParentColumn,
			})
		}

		result.Tables = append(result.Tables, t)
	}

	return result
}
