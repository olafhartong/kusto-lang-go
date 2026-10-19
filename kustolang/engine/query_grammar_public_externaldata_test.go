package kustolang

import "testing"

func TestQueryGrammarPublicExternalDataValidAST(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`externaldata hint.materialized=true (Uri:string, Count:long) ["https://one", strcat("https://", Host), BaseUri + "/tail",] with (format="csv", records=2, ratio=1.5, compressed=true, since=datetime(2024-01-02), schema=typeof(string), mode=identifier)`)
	external, ok := expression.(*ExternalDataExpression)
	if !ok {
		t.Fatalf("externaldata expression = %T, want *ExternalDataExpression", expression)
	}
	if external.Kind() != SyntaxKindExternalDataExpression ||
		external.ExternalDataKeyword().Kind() != SyntaxKindExternalDataKeyword ||
		external.ExternalDataKeyword().Width() == 0 {
		t.Fatalf("externaldata kind/keyword = %v/%#v", external.Kind(), external.ExternalDataKeyword())
	}
	if external.Parameters().Count() != 1 {
		t.Fatalf("externaldata parameters = %d, want 1", external.Parameters().Count())
	}
	parameter := external.Parameters().At(0)
	if NameDeclarationSimpleName(parameter.Name()) != "hint.materialized" ||
		parameter.EqualToken().Kind() != SyntaxKindEqualToken || parameter.EqualToken().Width() == 0 ||
		parameter.Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("externaldata parameter = %#v", parameter)
	}

	if external.Schema().Columns().Count() != 2 {
		t.Fatalf("externaldata schema columns = %d, want 2", external.Schema().Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, external.Schema().Columns().At(0).Element(), "Uri", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, external.Schema().Columns().At(1).Element(), "Count", SyntaxKindLongKeyword)
	if external.OpenBracket().Kind() != SyntaxKindOpenBracketToken || external.CloseBracket().Kind() != SyntaxKindCloseBracketToken {
		t.Fatalf("externaldata brackets = %v/%v", external.OpenBracket().Kind(), external.CloseBracket().Kind())
	}
	wantURIKinds := []SyntaxKind{
		SyntaxKindStringLiteralExpression,
		SyntaxKindFunctionCallExpression,
		SyntaxKindAddExpression,
	}
	if external.URIs().Count() != len(wantURIKinds) {
		t.Fatalf("externaldata URI count = %d, want %d", external.URIs().Count(), len(wantURIKinds))
	}
	for i, want := range wantURIKinds {
		if got := external.URIs().At(i).Element().Kind(); got != want {
			t.Fatalf("externaldata URI %d kind = %v, want %v", i, got, want)
		}
	}
	if separator := external.URIs().At(2).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken {
		t.Fatalf("externaldata trailing comma = %#v", separator)
	}

	with := external.WithClause()
	if with == nil || with.Kind() != SyntaxKindExternalDataWithClause ||
		with.WithKeyword().Kind() != SyntaxKindWithKeyword ||
		with.OpenParen().Kind() != SyntaxKindOpenParenToken || with.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("externaldata with clause = %#v", with)
	}
	wantPropertyKinds := []SyntaxKind{
		SyntaxKindStringLiteralExpression,
		SyntaxKindLongLiteralExpression,
		SyntaxKindRealLiteralExpression,
		SyntaxKindBooleanLiteralExpression,
		SyntaxKindDateTimeLiteralExpression,
		SyntaxKindTypeOfLiteralExpression,
		SyntaxKindNameDeclaration,
	}
	if with.Properties().Count() != len(wantPropertyKinds) {
		t.Fatalf("externaldata property count = %d, want %d", with.Properties().Count(), len(wantPropertyKinds))
	}
	for i, want := range wantPropertyKinds {
		property := with.Properties().At(i).Element()
		if property.EqualToken().Kind() != SyntaxKindEqualToken || property.EqualToken().Width() == 0 ||
			property.Expression().Kind() != want {
			t.Fatalf("externaldata property %d = %#v, want value kind %v", i, property, want)
		}
	}
	if diagnostics := external.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid externaldata diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicExternalDataHiddenSpellingAndPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	hidden := parseQueryGrammarProduction(t, grammar.Expression, `external_data (Value:dynamic) [Payload]`)
	external, ok := hidden.(*ExternalDataExpression)
	if !ok || external.ExternalDataKeyword().Kind() != SyntaxKindExternal_DataKeyword ||
		external.URIs().Count() != 1 || external.URIs().At(0).Element().Kind() != SyntaxKindNameReference ||
		len(external.ContainedDiagnostics()) != 0 {
		t.Fatalf("hidden external_data AST = %#v diagnostics:%#v", hidden, hidden.ContainedDiagnostics())
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		`externaldata (Value:long) ["https://one"] | where Value > 0 | project Value`)
	projectPipe, project := queryGrammarPipeStage(t, pipeline, SyntaxKindProjectOperator)
	wherePipe, where := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	if _, ok := wherePipe.Expression().(*ExternalDataExpression); !ok {
		t.Fatalf("externaldata pipeline source = %T, want *ExternalDataExpression", wherePipe.Expression())
	}
	if _, ok := project.(*ProjectOperator); !ok {
		t.Fatalf("externaldata final operator = %T, want *ProjectOperator", project)
	}
	if _, ok := where.(*FilterOperator); !ok {
		t.Fatalf("externaldata first operator = %T, want *FilterOperator", where)
	}
	if diagnostics := pipeline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("externaldata pipeline diagnostics = %#v", diagnostics)
	}
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, `externaldata (Value:long) ["https://one"]`)
}

func TestQueryGrammarPublicExternalDataPropertyValueAlternatives(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name     string
		value    string
		wantKind SyntaxKind
	}{
		{name: "string", value: `"csv"`, wantKind: SyntaxKindStringLiteralExpression},
		{name: "long", value: `42`, wantKind: SyntaxKindLongLiteralExpression},
		{name: "real", value: `1.25`, wantKind: SyntaxKindRealLiteralExpression},
		{name: "boolean", value: `false`, wantKind: SyntaxKindBooleanLiteralExpression},
		{name: "datetime", value: `datetime(2024-01-02)`, wantKind: SyntaxKindDateTimeLiteralExpression},
		{name: "typeof", value: `typeof(A:string, B:long)`, wantKind: SyntaxKindTypeOfLiteralExpression},
		{name: "raw guid", value: `00000000-0000-0000-0000-000000000001`, wantKind: SyntaxKindGuidLiteralExpression},
		{name: "name", value: `native`, wantKind: SyntaxKindNameDeclaration},
		{name: "extended name", value: `["native mode"]`, wantKind: SyntaxKindNameDeclaration},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			external := parseQueryGrammarProduction(t, grammar.Expression,
				`externaldata (Value:string) ["https://one"] with (option=`+test.value+`)`).(*ExternalDataExpression)
			properties := external.WithClause().Properties()
			if properties.Count() != 1 || properties.At(0).Element().Expression().Kind() != test.wantKind {
				t.Fatalf("externaldata property value %q = %#v, want kind %v", test.value, properties, test.wantKind)
			}
			if len(external.ContainedDiagnostics()) != 0 {
				t.Fatalf("externaldata property value %q diagnostics = %#v", test.value, external.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicExternalDataRejectsGeneralPropertyExpressions(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, test := range []struct {
		value            string
		wantMissingValue bool
		wantPrefixKind   SyntaxKind
	}{
		{value: `strcat("c", "sv")`, wantPrefixKind: SyntaxKindNameDeclaration},
		{value: `"c" "sv"`, wantPrefixKind: SyntaxKindStringLiteralExpression},
		{value: `dynamic({"format":"csv"})`, wantMissingValue: true},
		{value: `1.25d`, wantMissingValue: true},
		{value: `1h`, wantMissingValue: true},
		{value: `guid(00000000-0000-0000-0000-000000000001)`, wantMissingValue: true},
	} {
		t.Run(test.value, func(t *testing.T) {
			text := `externaldata (Value:string) ["https://one"] with (option=` + test.value + `)`
			tokens := Lex(text)
			result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
			if result.Length < 0 || result.Length == len(tokens)-1 {
				t.Fatalf("restricted externaldata property value %q parse = %#v, want unconsumed tail", test.value, result)
			}
			external, ok := result.Value.(*ExternalDataExpression)
			if !ok || external.WithClause() == nil || external.WithClause().Properties().Count() != 1 {
				t.Fatalf("restricted externaldata property value %q result = %#v", test.value, result)
			}
			property := external.WithClause().Properties().At(0).Element()
			if test.wantMissingValue {
				if property.Expression().Width() != 0 ||
					!diagnosticsContainMessage(property.Expression().ContainedDiagnostics(), "KS006", "Missing value") {
					t.Fatalf("restricted externaldata property value %q recovery = %#v diagnostics:%#v",
						test.value, property.Expression(), property.Expression().ContainedDiagnostics())
				}
			} else if property.Expression().Kind() != test.wantPrefixKind || property.Expression().Width() == 0 {
				t.Fatalf("restricted property value %q = %#v, want prefix kind %v", test.value, property.Expression(), test.wantPrefixKind)
			}
		})
	}
}

func TestQueryGrammarPublicExternalDataMissingSchemaBracketsAndURI(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	missingSchema := parseQueryGrammarProduction(t, grammar.Expression, `externaldata ["https://one"]`).(*ExternalDataExpression)
	if missingSchema.Schema().OpenParen().Width() != 0 || missingSchema.Schema().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingSchema.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("missing externaldata schema = %#v diagnostics:%#v", missingSchema.Schema(), missingSchema.Schema().ContainedDiagnostics())
	}

	tests := []struct {
		name         string
		text         string
		missingOpen  bool
		missingClose bool
	}{
		{name: "both brackets", text: `externaldata (Value:string)`, missingOpen: true, missingClose: true},
		{name: "open bracket", text: `externaldata (Value:string) "https://one"]`, missingOpen: true},
		{name: "close bracket", text: `externaldata (Value:string) ["https://one"`, missingClose: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			external := parseQueryGrammarProduction(t, grammar.Expression, test.text).(*ExternalDataExpression)
			if (external.OpenBracket().Width() == 0) != test.missingOpen ||
				(external.CloseBracket().Width() == 0) != test.missingClose {
				t.Fatalf("externaldata bracket widths = %d/%d", external.OpenBracket().Width(), external.CloseBracket().Width())
			}
			if test.missingOpen && !diagnosticsContainMessage(external.OpenBracket().Diagnostics(), "KS005", "Expected: [") {
				t.Fatalf("missing externaldata open bracket diagnostics = %#v", external.OpenBracket().Diagnostics())
			}
			if test.missingClose && !diagnosticsContainMessage(external.CloseBracket().Diagnostics(), "KS005", "Expected: ]") {
				t.Fatalf("missing externaldata close bracket diagnostics = %#v", external.CloseBracket().Diagnostics())
			}
		})
	}

	missingURI := parseQueryGrammarProduction(t, grammar.Expression, `externaldata (Value:string) []`).(*ExternalDataExpression)
	if missingURI.URIs().Count() != 1 || missingURI.URIs().At(0).Element().Width() != 0 ||
		!diagnosticsContainMessage(missingURI.URIs().At(0).Element().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing externaldata URI = %#v diagnostics:%#v", missingURI.URIs(), missingURI.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicExternalDataMissingURIComma(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := `externaldata (Value:string) [strcat("https://", One) strcat("https://", Two)]`
	tokens := Lex(text)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	if result.Length < 0 || result.Length != len(tokens)-1 {
		t.Fatalf("missing-comma externaldata URI parse = %#v, want full recovery", result)
	}
	external, ok := result.Value.(*ExternalDataExpression)
	if !ok || external.URIs().Count() != 2 {
		t.Fatalf("missing-comma externaldata URI recovery = %#v", result)
	}
	separator := external.URIs().At(0).Separator()
	if separator == nil || separator.Width() != 0 ||
		!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing externaldata URI comma = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}
}

func TestQueryGrammarPublicExternalDataWithClauseRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	const prefix = `externaldata (Value:string) ["https://one"] with `

	missingOpen := parseQueryGrammarProduction(t, grammar.Expression, prefix+`format="csv")`).(*ExternalDataExpression).WithClause()
	if missingOpen.OpenParen().Width() != 0 ||
		!diagnosticsContainMessage(missingOpen.OpenParen().Diagnostics(), "KS005", "Expected: (") {
		t.Fatalf("missing externaldata property open paren = %#v diagnostics:%#v", missingOpen.OpenParen(), missingOpen.ContainedDiagnostics())
	}

	missingValue := parseQueryGrammarProduction(t, grammar.Expression, prefix+`(format=)`).(*ExternalDataExpression).WithClause()
	if missingValue.Properties().Count() != 1 || missingValue.Properties().At(0).Element().Expression().Width() != 0 ||
		!diagnosticsContainMessage(missingValue.Properties().At(0).Element().Expression().ContainedDiagnostics(), "KS006", "Missing value") {
		t.Fatalf("missing externaldata property value = %#v diagnostics:%#v", missingValue.Properties(), missingValue.ContainedDiagnostics())
	}

	missingComma := parseQueryGrammarProduction(t, grammar.Expression, prefix+`(format="csv" compressed=true)`).(*ExternalDataExpression).WithClause()
	if missingComma.Properties().Count() != 2 {
		t.Fatalf("missing-comma externaldata properties = %d, want 2", missingComma.Properties().Count())
	}
	separator := missingComma.Properties().At(0).Separator()
	if separator == nil || separator.Width() != 0 ||
		!diagnosticsContainMessage(separator.Diagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing externaldata property comma = %#v diagnostics:%#v", separator, separator.Diagnostics())
	}

	missingClose := parseQueryGrammarProduction(t, grammar.Expression, prefix+`(format="csv"`).(*ExternalDataExpression).WithClause()
	if missingClose.CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(missingClose.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing externaldata property close paren = %#v diagnostics:%#v", missingClose.CloseParen(), missingClose.ContainedDiagnostics())
	}

	missingEqualsText := prefix + `(format "csv")`
	tokens := Lex(missingEqualsText)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	if result.Length < 0 || result.Length == len(tokens)-1 {
		t.Fatalf("missing externaldata property equals parse = %#v, want unconsumed tail", result)
	}
	external, ok := result.Value.(*ExternalDataExpression)
	if !ok || external.WithClause() == nil || external.WithClause().Properties().Count() != 0 ||
		external.WithClause().CloseParen().Width() != 0 ||
		!diagnosticsContainMessage(external.WithClause().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing externaldata property equals recovery = %#v", result)
	}
}
