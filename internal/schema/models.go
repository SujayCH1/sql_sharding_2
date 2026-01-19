package schema

// all table in a project
type LogicalSchema struct {
	ProjectID string
	Tables    map[string]*Table
}

// all columns in a table
type Table struct {
	Columns map[string]*Column
	FKs     map[FKKey]*FK
}

// column infomation
type Column struct {
	Name         string
	DataType     string
	Nullable     bool
	IsPrimaryKey bool
}

// describes what the relationship is.
type FK struct {
	ChildTable   string
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}

// describes how to uniquely identify it inside a table.
type FKKey struct {
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}

func NewLogicalSchema() *LogicalSchema {
	return &LogicalSchema{
		Tables: make(map[string]*Table),
	}
}
