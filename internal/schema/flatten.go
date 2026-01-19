package schema

import "sql-sharding-v2/internal/repository"

func FlattenColumns(schema *LogicalSchema) []repository.Columns {

	var result []repository.Columns

	for tableName, table := range schema.Tables {
		for _, column := range table.Columns {

			result = append(result, repository.Columns{
				ProjectID:    schema.ProjectID,
				TableName:    tableName,
				ColumnName:   column.Name,
				DataType:     column.DataType,
				Nullable:     column.Nullable,
				IsPrimaryKey: column.IsPrimaryKey,
			})
		}
	}

	return result
}

func FlattenFKEdges(schema *LogicalSchema) []repository.FKEdges {

	var result []repository.FKEdges

	for tableName, table := range schema.Tables {
		for _, fk := range table.FKs {

			result = append(result, repository.FKEdges{
				ProjectID:    schema.ProjectID,
				ParentTable:  fk.ParentTable,
				ParentColumn: fk.ParentColumn,
				ChildTable:   tableName,
				ChildColumn:  fk.ChildColumn,
			})
		}
	}

	return result
}
