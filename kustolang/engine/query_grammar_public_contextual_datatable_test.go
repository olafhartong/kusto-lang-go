package kustolang

import "testing"

func TestQueryGrammarPublicContextualDataTableValidASTAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.Expression,
		`__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (Account:string, Count:long) | where Count > 0 | project Account`)

	projectPipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindProjectOperator)
	wherePipe, _ := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	contextual, ok := wherePipe.Expression().(*ContextualDataTableExpression)
	if !ok {
		t.Fatalf("contextual datatable pipeline source = %T, want *ContextualDataTableExpression", wherePipe.Expression())
	}
	if contextual.Kind() != SyntaxKindContextualDataTableExpression ||
		contextual.ContextualDataTableKeyword().Kind() != SyntaxKindContextualDataTableKeyword ||
		contextual.ContextualDataTableKeyword().Width() == 0 {
		t.Fatalf("contextual datatable kind/keyword = %v/%#v", contextual.Kind(), contextual.ContextualDataTableKeyword())
	}
	if contextual.Id().Kind() != SyntaxKindGuidLiteralExpression {
		t.Fatalf("contextual datatable id kind = %v, want %v", contextual.Id().Kind(), SyntaxKindGuidLiteralExpression)
	}
	if contextual.Schema().Columns().Count() != 2 {
		t.Fatalf("contextual datatable schema columns = %d, want 2", contextual.Schema().Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, contextual.Schema().Columns().At(0).Element(), "Account", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, contextual.Schema().Columns().At(1).Element(), "Count", SyntaxKindLongKeyword)
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid contextual datatable diagnostics = %#v", diagnostics)
	}

	root := parseQueryGrammarProduction(t, grammar.Expression, `__contextual_datatable "ContextId" ()`)
	rootContextual, ok := root.(*ContextualDataTableExpression)
	if !ok {
		t.Fatalf("contextual datatable root = %T, want *ContextualDataTableExpression", root)
	}
	if rootContextual.Id().Kind() != SyntaxKindStringLiteralExpression || rootContextual.Schema().Columns().Count() != 0 ||
		len(rootContextual.ContainedDiagnostics()) != 0 {
		t.Fatalf("non-guid contextual datatable = id:%#v schema:%#v diagnostics:%#v",
			rootContextual.Id(), rootContextual.Schema(), rootContextual.ContainedDiagnostics())
	}

	arbitraryID := parseQueryGrammarProduction(t, grammar.UnnamedExpression,
		`__contextual_datatable 1 + 2 (Value:long)`).(*ContextualDataTableExpression)
	if arbitraryID.Id().Kind() != SyntaxKindAddExpression || arbitraryID.Schema().Columns().Count() != 1 ||
		len(arbitraryID.ContainedDiagnostics()) != 0 {
		t.Fatalf("expression-id contextual datatable = id:%#v schema:%#v diagnostics:%#v",
			arbitraryID.Id(), arbitraryID.Schema(), arbitraryID.ContainedDiagnostics())
	}

	assertQueryGrammarProductionRejects(t, grammar.QueryOperator,
		`__contextual_datatable guid(00000000-0000-0000-0000-000000000001) ()`)
}

func TestQueryGrammarPublicContextualDataTableSchemas(t *testing.T) {
	contextual := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`__contextual_datatable "context" (, where:string, ['event id']:long, Nested:(Value:dynamic, ['inner value']:guid),)`).(*ContextualDataTableExpression)
	schema := contextual.Schema()
	if schema.OpenParen().Kind() != SyntaxKindOpenParenToken || schema.CloseParen().Kind() != SyntaxKindCloseParenToken ||
		schema.LeadingComma() == nil || schema.LeadingComma().Kind() != SyntaxKindCommaToken || schema.Columns().Count() != 3 {
		t.Fatalf("contextual datatable schema = parens:%v/%v leading:%#v columns:%d",
			schema.OpenParen().Kind(), schema.CloseParen().Kind(), schema.LeadingComma(), schema.Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(0).Element(), "where", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(1).Element(), "event id", SyntaxKindLongKeyword)

	nested := schema.Columns().At(2).Element()
	if NameDeclarationSimpleName(nested.Name()) != "Nested" {
		t.Fatalf("nested contextual schema column name = %q, want Nested", NameDeclarationSimpleName(nested.Name()))
	}
	nestedType, ok := nested.Type().(*SchemaTypeExpression)
	if !ok || nestedType.Columns().Count() != 2 {
		t.Fatalf("nested contextual schema type = %#v, want two columns", nested.Type())
	}
	firstNested, ok := nestedType.Columns().At(0).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("first nested contextual schema element = %T, want *NameAndTypeDeclaration", nestedType.Columns().At(0).Element())
	}
	secondNested, ok := nestedType.Columns().At(1).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("second nested contextual schema element = %T, want *NameAndTypeDeclaration", nestedType.Columns().At(1).Element())
	}
	assertQueryGrammarRowSchemaColumn(t, firstNested, "Value", SyntaxKindDynamicKeyword)
	assertQueryGrammarRowSchemaColumn(t, secondNested, "inner value", SyntaxKindGuidKeyword)
	if schema.Columns().At(2).Separator() == nil || schema.Columns().At(2).Separator().Kind() != SyntaxKindCommaToken ||
		len(contextual.ContainedDiagnostics()) != 0 {
		t.Fatalf("contextual schema trailing comma/diagnostics = %#v/%#v",
			schema.Columns().At(2).Separator(), contextual.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicContextualDataTableMissingIDAndSchemaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	expression := parseQueryGrammarProduction(t, grammar.Expression, `__contextual_datatable | take 1`)
	pipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindTakeOperator)
	contextual, ok := pipe.Expression().(*ContextualDataTableExpression)
	if !ok {
		t.Fatalf("recovered contextual datatable source = %T, want *ContextualDataTableExpression", pipe.Expression())
	}
	if contextual.Id().Width() != 0 || contextual.Id().Kind() != SyntaxKindNameReference ||
		!diagnosticsContainMessage(contextual.Id().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing contextual datatable id = %#v diagnostics:%#v", contextual.Id(), contextual.Id().ContainedDiagnostics())
	}
	if contextual.Schema().OpenParen().Width() != 0 || contextual.Schema().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(contextual.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing contextual datatable schema = %#v diagnostics:%#v",
			contextual.Schema(), contextual.Schema().ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicContextualDataTableMalformedSchemaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	const prefix = `__contextual_datatable guid(00000000-0000-0000-0000-000000000001) `
	tests := []struct {
		name    string
		schema  string
		columns int
		code    string
		message string
	}{
		{name: "missing schema", columns: 0, code: "KS006", message: "Missing schema declaration"},
		{name: "missing close paren", schema: "(A:string", columns: 1, code: "KS005", message: "Expected: )"},
		{name: "missing name", schema: "(:string)", columns: 1, code: "KS006", message: "Missing name"},
		{name: "missing type", schema: "(A:)", columns: 1, code: "KS006", message: "Missing type name"},
		{name: "missing comma", schema: "(A:string B:long)", columns: 2, code: "KS005", message: "Expected: ,"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contextual := parseQueryGrammarProduction(t, grammar.Expression, prefix+test.schema).(*ContextualDataTableExpression)
			if contextual.Schema().Columns().Count() != test.columns ||
				!diagnosticsContainMessage(contextual.Schema().ContainedDiagnostics(), test.code, test.message) {
				t.Fatalf("contextual datatable schema recovery %q = columns:%d diagnostics:%#v",
					test.schema, contextual.Schema().Columns().Count(), contextual.Schema().ContainedDiagnostics())
			}
		})
	}
}
