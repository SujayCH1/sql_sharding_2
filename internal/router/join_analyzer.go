package router

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type JoinInfo struct {
	LeftTable   string
	RightTable  string
	LeftColumn  string
	RightColumn string
}

func ExtractJoinInfo(stmt *pg_query.SelectStmt) (*JoinInfo, bool) {

	if len(stmt.FromClause) != 1 {
		return nil, false
	}

	joinNode, ok := stmt.FromClause[0].Node.(*pg_query.Node_JoinExpr)
	if !ok {
		return nil, false
	}

	join := joinNode.JoinExpr

	left := join.Larg.Node.(*pg_query.Node_RangeVar).RangeVar.Relname
	right := join.Rarg.Node.(*pg_query.Node_RangeVar).RangeVar.Relname

	expr := join.Quals.Node.(*pg_query.Node_AExpr).AExpr

	fmt.Printf("Expression: %s", expr)

	leftCol, ok1 := extractColumn(expr.Lexpr)
	rightCol, ok2 := extractColumn(expr.Rexpr)

	if !ok1 || !ok2 {
		fmt.Println("Failed to extract columns from join condition")
		return nil, false
	}

	// fmt.Println("left Table", left, "Right Table: ", right, "left column: ", leftCol, "Right column: ", rightCol)

	return &JoinInfo{
		LeftTable:   left,
		RightTable:  right,
		LeftColumn:  leftCol,
		RightColumn: rightCol,
	}, true
}
