package schema

import (
	"sql-sharding-v2/internal/repository"
)

func addColsToLogicalSchema(
	projectID string,
	schema *LogicalSchema,
	cols []repository.Columns,
) {

	schema.ProjectID = projectID

	for _, col := range cols {
		ensureTable(schema, col.TableName)

		schema.Tables[col.TableName].Columns[col.ColumnName] = &Column{
			Name:         col.ColumnName,
			DataType:     col.DataType,
			Nullable:     col.Nullable,
			IsPrimaryKey: col.IsPrimaryKey,
		}
	}
}

func addFKsToLogicalSchema(
	schema *LogicalSchema,
	fks []repository.FKEdges,
) {

	for _, fk := range fks {
		ensureTable(schema, fk.ChildTable)

		key := FKKey{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}

		schema.Tables[fk.ChildTable].FKs[key] = &FK{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}
}

func ensureTable(schema *LogicalSchema, tableName string) {
	if _, ok := schema.Tables[tableName]; !ok {
		schema.Tables[tableName] = &Table{
			Columns: make(map[string]*Column),
			FKs:     make(map[FKKey]*FK),
		}
	}
}
