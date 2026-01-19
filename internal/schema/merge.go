package schema

func cloneLogicalSchema(base *LogicalSchema) *LogicalSchema {
	if base == nil {
		return NewLogicalSchema()
	}

	newSchema := NewLogicalSchema()
	newSchema.ProjectID = base.ProjectID

	for tableName, table := range base.Tables {
		newSchema.Tables[tableName] = cloneTable(table)
	}

	return newSchema
}

func cloneTable(t *Table) *Table {
	if t == nil {
		return &Table{
			Columns: make(map[string]*Column),
			FKs:     make(map[FKKey]*FK),
		}
	}

	newTable := &Table{
		Columns: make(map[string]*Column),
		FKs:     make(map[FKKey]*FK),
	}

	for colName, col := range t.Columns {
		newTable.Columns[colName] = &Column{
			Name:         col.Name,
			DataType:     col.DataType,
			Nullable:     col.Nullable,
			IsPrimaryKey: col.IsPrimaryKey,
		}
	}

	for fkKey, fk := range t.FKs {
		newTable.FKs[fkKey] = &FK{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}

	return newTable
}

func mergeTable(base *Table, delta *Table) *Table {
	merged := cloneTable(base)

	for colName, col := range delta.Columns {
		merged.Columns[colName] = &Column{
			Name:         col.Name,
			DataType:     col.DataType,
			Nullable:     col.Nullable,
			IsPrimaryKey: col.IsPrimaryKey,
		}
	}

	for fkKey, fk := range delta.FKs {
		merged.FKs[fkKey] = &FK{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}

	return merged
}
