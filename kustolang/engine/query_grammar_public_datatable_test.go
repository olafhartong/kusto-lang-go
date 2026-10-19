package kustolang

import (
	"testing"
)

func TestQueryGrammarPublicDataTableValidAST(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`datatable (Account:string, Count:long, Active:bool, Payload:dynamic) ["alpha", 2, true, dynamic({"x":1}), "beta", 3.5, false, dynamic(null)]`)
	dataTable, ok := expression.(*DataTableExpression)
	if !ok {
		t.Fatalf("datatable expression = %T, want *DataTableExpression", expression)
	}
	if dataTable.DataTableKeyword().Kind() != SyntaxKindDataTableKeyword || dataTable.Parameters().Count() != 0 {
		t.Fatalf("datatable keyword/parameters = %v/%d", dataTable.DataTableKeyword().Kind(), dataTable.Parameters().Count())
	}
	if dataTable.OpenBracket().Kind() != SyntaxKindOpenBracketToken || dataTable.CloseBracket().Kind() != SyntaxKindCloseBracketToken ||
		dataTable.LeadingComma() != nil {
		t.Fatalf("datatable brackets/leading comma = %v/%v/%#v", dataTable.OpenBracket().Kind(), dataTable.CloseBracket().Kind(), dataTable.LeadingComma())
	}

	schema := dataTable.Schema()
	if schema.Columns().Count() != 4 {
		t.Fatalf("datatable schema columns = %d, want 4", schema.Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(0).Element(), "Account", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(1).Element(), "Count", SyntaxKindLongKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(2).Element(), "Active", SyntaxKindBoolKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(3).Element(), "Payload", SyntaxKindDynamicKeyword)

	wantValueKinds := []SyntaxKind{
		SyntaxKindStringLiteralExpression,
		SyntaxKindLongLiteralExpression,
		SyntaxKindBooleanLiteralExpression,
		SyntaxKindDynamicExpression,
		SyntaxKindStringLiteralExpression,
		SyntaxKindRealLiteralExpression,
		SyntaxKindBooleanLiteralExpression,
		SyntaxKindDynamicExpression,
	}
	if dataTable.Values().Count() != len(wantValueKinds) {
		t.Fatalf("datatable values = %d, want %d", dataTable.Values().Count(), len(wantValueKinds))
	}
	for i, want := range wantValueKinds {
		if got := dataTable.Values().At(i).Element().Kind(); got != want {
			t.Fatalf("datatable value %d kind = %v, want %v", i, got, want)
		}
	}
	if diagnostics := dataTable.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid datatable diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicDataTableEmptyLeadingAndTrailingValues(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.Expression, "datatable () []").(*DataTableExpression)
	if empty.Schema().Columns().Count() != 0 || empty.Values().Count() != 0 || empty.LeadingComma() != nil ||
		len(empty.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty datatable = schema:%d values:%d leading:%#v diagnostics:%#v",
			empty.Schema().Columns().Count(), empty.Values().Count(), empty.LeadingComma(), empty.ContainedDiagnostics())
	}

	leading := parseQueryGrammarProduction(t, grammar.Expression, "datatable (x:long) [, 1]").(*DataTableExpression)
	if leading.LeadingComma() == nil || leading.LeadingComma().Kind() != SyntaxKindCommaToken || leading.Values().Count() != 1 ||
		len(leading.ContainedDiagnostics()) != 0 {
		t.Fatalf("leading-comma datatable = leading:%#v values:%d diagnostics:%#v",
			leading.LeadingComma(), leading.Values().Count(), leading.ContainedDiagnostics())
	}

	trailing := parseQueryGrammarProduction(t, grammar.Expression, "datatable (x:long) [1,]").(*DataTableExpression)
	if trailing.Values().Count() != 1 || trailing.Values().At(0).Separator() == nil ||
		trailing.Values().At(0).Separator().Kind() != SyntaxKindCommaToken || len(trailing.ContainedDiagnostics()) != 0 {
		t.Fatalf("trailing-comma datatable values = %#v diagnostics:%#v", trailing.Values(), trailing.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicDataTableRowSchemaForms(t *testing.T) {
	dataTable := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`datatable (, where:string, ['event id']:long, Nested:(Value:dynamic),) []`).(*DataTableExpression)
	schema := dataTable.Schema()
	if schema.LeadingComma() == nil || schema.LeadingComma().Kind() != SyntaxKindCommaToken || schema.Columns().Count() != 3 {
		t.Fatalf("datatable row schema leading comma/columns = %#v/%d", schema.LeadingComma(), schema.Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(0).Element(), "where", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(1).Element(), "event id", SyntaxKindLongKeyword)
	nested := schema.Columns().At(2).Element()
	if NameDeclarationSimpleName(nested.Name()) != "Nested" {
		t.Fatalf("nested datatable column name = %q, want Nested", NameDeclarationSimpleName(nested.Name()))
	}
	nestedType, ok := nested.Type().(*SchemaTypeExpression)
	if !ok || nestedType.Columns().Count() != 1 {
		t.Fatalf("nested datatable column type = %#v", nested.Type())
	}
	nestedColumn, ok := nestedType.Columns().At(0).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("nested datatable schema element = %T, want *NameAndTypeDeclaration", nestedType.Columns().At(0).Element())
	}
	assertQueryGrammarRowSchemaColumn(t, nestedColumn, "Value", SyntaxKindDynamicKeyword)
	if schema.Columns().At(2).Separator() == nil || schema.Columns().At(2).Separator().Kind() != SyntaxKindCommaToken ||
		len(dataTable.ContainedDiagnostics()) != 0 {
		t.Fatalf("datatable row schema trailing comma/diagnostics = %#v/%#v", schema.Columns().At(2).Separator(), dataTable.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicDataTableAcceptsGlobalParameters(t *testing.T) {
	dataTable := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		"datatable hint.materialized=true (x:long) [1]").(*DataTableExpression)
	if dataTable.Parameters().Count() != 1 {
		t.Fatalf("datatable parameters = %d, want 1", dataTable.Parameters().Count())
	}
	parameter := dataTable.Parameters().At(0)
	if NameDeclarationSimpleName(parameter.Name()) != "hint.materialized" ||
		parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 ||
		parameter.Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("datatable global parameter = %#v", parameter)
	}
	if diagnostics := dataTable.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("datatable global parameter diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicDataTableRootAndPipeline(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	root := parseQueryGrammarProduction(t, grammar.Expression, "datatable (x:long) [1, 2]")
	if _, ok := root.(*DataTableExpression); !ok {
		t.Fatalf("datatable root = %T, want *DataTableExpression", root)
	}

	expression := parseQueryGrammarProduction(t, grammar.Expression,
		"datatable (x:long) [1, 2] | where x > 1 | project x")
	projectPipe, project := queryGrammarPipeStage(t, expression, SyntaxKindProjectOperator)
	wherePipe, where := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	if _, ok := wherePipe.Expression().(*DataTableExpression); !ok {
		t.Fatalf("datatable pipeline source = %T, want *DataTableExpression", wherePipe.Expression())
	}
	if _, ok := project.(*ProjectOperator); !ok {
		t.Fatalf("datatable final operator = %T, want *ProjectOperator", project)
	}
	if _, ok := where.(*FilterOperator); !ok {
		t.Fatalf("datatable first operator = %T, want *FilterOperator", where)
	}
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("datatable pipeline diagnostics = %#v", diagnostics)
	}
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, "datatable (x:long) [1]")
}

func TestQueryGrammarPublicDataTableMissingSchemaAndBrackets(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingSchema := parseQueryGrammarProduction(t, grammar.Expression, "datatable [1]").(*DataTableExpression)
	if missingSchema.Schema().OpenParen().Width() != 0 || missingSchema.Schema().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingSchema.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing datatable schema = %#v diagnostics:%#v", missingSchema.Schema(), missingSchema.Schema().ContainedDiagnostics())
	}

	tests := []struct {
		name         string
		text         string
		missingOpen  bool
		missingClose bool
	}{
		{name: "both brackets", text: "datatable (x:long)", missingOpen: true, missingClose: true},
		{name: "open bracket", text: "datatable (x:long) 1]", missingOpen: true},
		{name: "close bracket", text: "datatable (x:long) [1", missingClose: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataTable := parseQueryGrammarProduction(t, grammar.Expression, test.text).(*DataTableExpression)
			if (dataTable.OpenBracket().Width() == 0) != test.missingOpen ||
				(dataTable.CloseBracket().Width() == 0) != test.missingClose {
				t.Fatalf("datatable bracket widths = %d/%d", dataTable.OpenBracket().Width(), dataTable.CloseBracket().Width())
			}
			if test.missingOpen && !diagnosticsContainMessage(dataTable.OpenBracket().Diagnostics(), "KS005", "Expected: [") {
				t.Fatalf("missing open bracket diagnostics = %#v", dataTable.OpenBracket().Diagnostics())
			}
			if test.missingClose && !diagnosticsContainMessage(dataTable.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
				t.Fatalf("missing close bracket diagnostics = %#v", dataTable.CloseBracket().Diagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicDataTableRejectsNonliteralCells(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, text := range []string{
		"datatable (x:long) [x]",
		`datatable (x:string) [strcat("a", "b")]`,
	} {
		t.Run(text, func(t *testing.T) {
			tokens := Lex(text)
			result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
			if result.Length < 0 {
				t.Fatalf("datatable nonliteral prefix failed: %#v", result)
			}
			if result.Length == len(tokens)-1 {
				t.Fatalf("datatable nonliteral unexpectedly consumed all input: %#v", result)
			}
			dataTable, ok := result.Value.(*DataTableExpression)
			if !ok || dataTable.Values().Count() != 0 || dataTable.CloseBracket().Width() != 0 ||
				!diagnosticsContainMessage(dataTable.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
				t.Fatalf("datatable nonliteral result = %#v", result)
			}
		})
	}
}
