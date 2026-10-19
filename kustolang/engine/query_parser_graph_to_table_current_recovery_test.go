package kustolang

import "testing"

func TestCurrentGraphToTableOutputListRequiresRealComma(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		statements int
	}{
		{"output clause", `T | graph-to-table nodes as N edges as E`, 3},
		{"node parameter", `T | graph-to-table nodes with_node_id NodeKey`, 3},
		{"edge parameters", `T | graph-to-table edges with_source_id SrcId with_target_id DstId`, 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseQuery(test.query)
			statements, ok := block.GetChild(1).(*SyntaxList)
			if !ok || statements.Count() != test.statements {
				t.Fatalf("statements = %#v, want %d", block.GetChild(1), test.statements)
			}
			pipe, ok := firstParsedQueryExpression(t, test.query).(*PipeExpression)
			if !ok {
				t.Fatalf("first expression = %#v", firstParsedQueryExpression(t, test.query))
			}
			op, ok := pipe.Operator().(*GraphToTableOperator)
			if !ok || op.OutputClause().Count() != 1 || op.OutputClause().At(0).Element().Parameters().Count() != 0 {
				t.Fatalf("graph-to-table operator = %#v", pipe.Operator())
			}
		})
	}
}

func TestCurrentGraphClauseParametersDoNotRecoverAcrossClauseBoundaries(t *testing.T) {
	tests := []string{
		`T | graph-to-table nodes with_source_id=Src`,
		`T | graph-to-table nodes with_source_id='Src' 'Hash'`,
		`T | graph-mark-components with_component_id Component | graph-to-table nodes`,
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			pipe, ok := firstParsedQueryExpression(t, query).(*PipeExpression)
			if !ok {
				t.Fatalf("first expression = %#v", firstParsedQueryExpression(t, query))
			}
			switch op := pipe.Operator().(type) {
			case *GraphToTableOperator:
				if op.OutputClause().At(0).Element().Parameters().Count() != 0 {
					t.Fatalf("wrong-clause parameter was attached to graph-to-table: %#v", op)
				}
			case *GraphMarkComponentsOperator:
				if op.Parameters().Count() != 0 {
					t.Fatalf("missing-equals parameter was attached to graph-mark-components: %#v", op)
				}
			default:
				t.Fatalf("operator = %T", op)
			}
		})
	}
}
