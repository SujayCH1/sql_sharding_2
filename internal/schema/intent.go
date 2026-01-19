package schema

// interfaces for converion from AST to intent
type AddColumnIntent struct {
	TableName string
	Column    Column
}

type RemoveColumnIntent struct {
	TableName  string
	ColumnName string
}

type AddFKIntent struct {
	TableName string // child table
	FK        FK
}

type RemoveFKIntent struct {
	TableName string // child table
	FKKey     FKKey
}
