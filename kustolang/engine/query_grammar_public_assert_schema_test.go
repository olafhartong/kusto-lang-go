package kustolang

import "testing"

func TestQueryGrammarPublicAssertSchemaValidRowSchema(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`assert-schema (User:string, ['event id']:long, Nested:(Value:dynamic, ['inner value']:guid),)`)
	assertSchema, ok := operator.(*AssertSchemaOperator)
	if !ok {
		t.Fatalf("assert-schema operator = %T, want *AssertSchemaOperator", operator)
	}
	if assertSchema.AssertSchemaKeyword().Kind() != SyntaxKindAssertSchemaKeyword {
		t.Fatalf("assert-schema keyword kind = %v", assertSchema.AssertSchemaKeyword().Kind())
	}

	schema := assertSchema.Schema()
	if schema.OpenParen().Kind() != SyntaxKindOpenParenToken || schema.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("assert-schema parens = %v/%v", schema.OpenParen().Kind(), schema.CloseParen().Kind())
	}
	if schema.LeadingComma() != nil || schema.Columns().Count() != 3 {
		t.Fatalf("assert-schema leading comma/columns = %#v/%d, want nil/3", schema.LeadingComma(), schema.Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(0).Element(), "User", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, schema.Columns().At(1).Element(), "event id", SyntaxKindLongKeyword)

	nested := schema.Columns().At(2).Element()
	if NameDeclarationSimpleName(nested.Name()) != "Nested" {
		t.Fatalf("nested column name = %q, want Nested", NameDeclarationSimpleName(nested.Name()))
	}
	nestedType, ok := nested.Type().(*SchemaTypeExpression)
	if !ok || nestedType.Columns().Count() != 2 {
		t.Fatalf("nested column type = %#v, want two-column schema", nested.Type())
	}
	firstNested, ok := nestedType.Columns().At(0).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("first nested schema element = %T, want *NameAndTypeDeclaration", nestedType.Columns().At(0).Element())
	}
	secondNested, ok := nestedType.Columns().At(1).Element().(*NameAndTypeDeclaration)
	if !ok {
		t.Fatalf("second nested schema element = %T, want *NameAndTypeDeclaration", nestedType.Columns().At(1).Element())
	}
	assertQueryGrammarRowSchemaColumn(t, firstNested, "Value", SyntaxKindDynamicKeyword)
	assertQueryGrammarRowSchemaColumn(t, secondNested, "inner value", SyntaxKindGuidKeyword)
	if diagnostics := assertSchema.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid assert-schema diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicAssertSchemaRowSchemaForms(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	empty := parseQueryGrammarProduction(t, grammar.QueryOperator, "assert-schema ()").(*AssertSchemaOperator).Schema()
	if empty.LeadingComma() != nil || empty.Columns().Count() != 0 || len(empty.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty assert-schema = %#v diagnostics:%#v", empty, empty.ContainedDiagnostics())
	}

	open := parseQueryGrammarProduction(t, grammar.QueryOperator, "assert-schema (, A:string)").(*AssertSchemaOperator).Schema()
	if open.LeadingComma() == nil || open.LeadingComma().Kind() != SyntaxKindCommaToken || open.Columns().Count() != 1 {
		t.Fatalf("open assert-schema leading comma/columns = %#v/%d", open.LeadingComma(), open.Columns().Count())
	}
	if diagnostics := open.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("open assert-schema diagnostics = %#v", diagnostics)
	}

	trailing := parseQueryGrammarProduction(t, grammar.QueryOperator, "assert-schema (A:string,)").(*AssertSchemaOperator).Schema()
	if trailing.Columns().Count() != 1 || trailing.Columns().At(0).Separator() == nil ||
		trailing.Columns().At(0).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("trailing-comma assert-schema columns = %#v", trailing.Columns())
	}
	if diagnostics := trailing.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("legal trailing-comma assert-schema diagnostics = %#v", diagnostics)
	}

	starTokens := Lex("assert-schema (*)")
	star := grammar.QueryOperator.ParseValue(NewArraySource(starTokens), 0)
	if star.Length != 2 {
		t.Fatalf("assert-schema star prefix length = %d, want 2: %#v", star.Length, star)
	}
	starSchema := star.Value.(*AssertSchemaOperator).Schema()
	if starSchema.Columns().Count() != 0 || starSchema.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(starSchema.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("assert-schema star recovery = %#v diagnostics:%#v", starSchema, starSchema.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicAssertSchemaPipeline(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		"T | project User, Count | assert-schema (User:string, Count:long) | where Count > 0 | take 5")

	takePipe, _ := queryGrammarPipeStage(t, expression, SyntaxKindTakeOperator)
	wherePipe, _ := queryGrammarPipeStage(t, takePipe.Expression(), SyntaxKindFilterOperator)
	assertPipe, assertNode := queryGrammarPipeStage(t, wherePipe.Expression(), SyntaxKindAssertSchemaOperator)
	projectPipe, _ := queryGrammarPipeStage(t, assertPipe.Expression(), SyntaxKindProjectOperator)
	if source, ok := projectPipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("assert-schema pipeline source = %#v, want T", projectPipe.Expression())
	}
	if assertNode.(*AssertSchemaOperator).Schema().Columns().Count() != 2 {
		t.Fatalf("assert-schema pipeline schema = %#v", assertNode.(*AssertSchemaOperator).Schema())
	}
	if diagnostics := expression.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("assert-schema pipeline diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicAssertSchemaExtendedNamesAndInvalidTypes(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	extended := parseQueryGrammarProduction(t, grammar.QueryOperator,
		"assert-schema (where:string, evaluate:long)").(*AssertSchemaOperator).Schema()
	if extended.Columns().Count() != 2 || NameDeclarationSimpleName(extended.Columns().At(0).Element().Name()) != "where" ||
		NameDeclarationSimpleName(extended.Columns().At(1).Element().Name()) != "evaluate" || len(extended.ContainedDiagnostics()) != 0 {
		t.Fatalf("extended keyword schema = %#v diagnostics:%#v", extended.Columns(), extended.ContainedDiagnostics())
	}

	for _, typeName := range []string{"float", "int16", "uint64"} {
		t.Run(typeName, func(t *testing.T) {
			schema := parseQueryGrammarProduction(t, grammar.QueryOperator,
				"assert-schema (A:"+typeName+")").(*AssertSchemaOperator).Schema()
			if schema.Columns().Count() != 1 || !diagnosticsContain(schema.ContainedDiagnostics(), "KS160") {
				t.Fatalf("invalid type %q schema diagnostics = %#v", typeName, schema.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicAssertSchemaRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	missingSchema := parseQueryGrammarProduction(t, grammar.QueryOperator, "assert-schema").(*AssertSchemaOperator).Schema()
	if missingSchema.OpenParen().Width() != 0 || missingSchema.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingSchema.ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing assert-schema schema = %#v diagnostics:%#v", missingSchema, missingSchema.ContainedDiagnostics())
	}

	missingParenTokens := Lex("assert-schema A:string")
	missingParen := grammar.QueryOperator.ParseValue(NewArraySource(missingParenTokens), 0)
	if missingParen.Length != 1 {
		t.Fatalf("missing-open-paren prefix length = %d, want 1: %#v", missingParen.Length, missingParen)
	}
	if schema := missingParen.Value.(*AssertSchemaOperator).Schema(); !diagnosticsContainMessage(schema.ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing-open-paren schema diagnostics = %#v", schema.ContainedDiagnostics())
	}

	tests := []struct {
		name    string
		text    string
		columns int
		code    string
		message string
	}{
		{name: "close paren", text: "assert-schema (A:string", columns: 1, code: "KS005", message: "Expected: )"},
		{name: "name", text: "assert-schema (:string)", columns: 1, code: "KS006", message: "Missing name"},
		{name: "type", text: "assert-schema (A:)", columns: 1, code: "KS006", message: "Missing type name"},
		{name: "comma", text: "assert-schema (A:string B:long)", columns: 2, code: "KS005", message: "Expected: ,"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text).(*AssertSchemaOperator).Schema()
			if schema.Columns().Count() != test.columns ||
				!diagnosticsContainMessage(schema.ContainedDiagnostics(), test.code, test.message) {
				t.Fatalf("assert-schema recovery %q = columns:%d diagnostics:%#v", test.text, schema.Columns().Count(), schema.ContainedDiagnostics())
			}
		})
	}

	doubleTrailingTokens := Lex("assert-schema (A:string,,)")
	doubleTrailing := grammar.QueryOperator.ParseValue(NewArraySource(doubleTrailingTokens), 0)
	if doubleTrailing.Length != 6 {
		t.Fatalf("double-trailing-comma prefix length = %d, want 6: %#v", doubleTrailing.Length, doubleTrailing)
	}
	doubleTrailingSchema := doubleTrailing.Value.(*AssertSchemaOperator).Schema()
	if doubleTrailingSchema.Columns().Count() != 1 || doubleTrailingSchema.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(doubleTrailingSchema.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("double-trailing-comma recovery = %#v diagnostics:%#v", doubleTrailingSchema, doubleTrailingSchema.ContainedDiagnostics())
	}
}

func assertQueryGrammarRowSchemaColumn(t *testing.T, column *NameAndTypeDeclaration, wantName string, wantType SyntaxKind) {
	t.Helper()
	if NameDeclarationSimpleName(column.Name()) != wantName {
		t.Fatalf("row-schema column name = %q, want %q", NameDeclarationSimpleName(column.Name()), wantName)
	}
	primitive, ok := column.Type().(*PrimitiveTypeExpression)
	if !ok || primitive.Type().Kind() != wantType {
		t.Fatalf("row-schema column %q type = %#v, want primitive %v", wantName, column.Type(), wantType)
	}
}
