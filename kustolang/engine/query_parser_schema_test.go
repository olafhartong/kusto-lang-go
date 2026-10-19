package kustolang

import "testing"

func TestParseRowSchemaBuildsGeneratedSyntaxShape(t *testing.T) {
	schema := ParseRowSchema("(Account: string, ['bad name']: long)")
	if schema == nil {
		t.Fatalf("ParseRowSchema returned nil")
	}
	if schema.OpenParen().Kind() != SyntaxKindOpenParenToken || schema.CloseParen().Kind() != SyntaxKindCloseParenToken {
		t.Fatalf("schema parens not preserved")
	}
	columns := schema.Columns()
	if columns.Count() != 2 {
		t.Fatalf("columns count = %d", columns.Count())
	}
	first := columns.At(0).Element()
	if tokenName, ok := first.Name().Name().(*TokenName); !ok || tokenName.Name().ValueText() != "Account" {
		t.Fatalf("first name = %#v", first.Name().Name())
	}
	if typ := first.Type().(*PrimitiveTypeExpression).Type(); typ.Kind() != SyntaxKindStringKeyword {
		t.Fatalf("first type = %v", typ.Kind())
	}
	second := columns.At(1).Element()
	if _, ok := second.Name().Name().(*BracketedName); !ok {
		t.Fatalf("second name = %#v", second.Name().Name())
	}
	if got := schema.Text(IncludeTriviaAll); got != "(Account: string, ['bad name']: long)" {
		t.Fatalf("schema text = %q", got)
	}
}

func TestParseRowSchemaAllowsEmptySchema(t *testing.T) {
	schema := ParseRowSchema("()")
	if schema == nil {
		t.Fatalf("ParseRowSchema returned nil")
	}
	if diagnostics := schema.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics = %#v", diagnostics)
	}
	if schema.Columns() == nil || schema.Columns().Count() != 0 {
		t.Fatalf("columns = %#v", schema.Columns())
	}
	if got := schema.Text(IncludeTriviaAll); got != "()" {
		t.Fatalf("schema text = %q", got)
	}
}

func TestParseRowSchemaAllowsUpstreamExtendedKeywordsAsColumnNames(t *testing.T) {
	schema := ParseRowSchema("(access:string, alias:string, database:string, evaluate:string, from:string, where:string, by:long, print:string, sample-distinct:string, toscalar:string)")
	if schema == nil {
		t.Fatalf("ParseRowSchema returned nil")
	}
	columns := schema.Columns()
	want := []string{"access", "alias", "database", "evaluate", "from", "where", "by", "print", "sample-distinct", "toscalar"}
	if columns.Count() != len(want) {
		t.Fatalf("columns count = %d, want %d", columns.Count(), len(want))
	}
	for i, name := range want {
		column := columns.At(i).Element()
		tokenName, ok := column.Name().Name().(*TokenName)
		if !ok {
			t.Fatalf("column %d name = %#v", i, column.Name().Name())
		}
		if tokenName.Name().ValueText() != name {
			t.Fatalf("column %d name = %q, want %q", i, tokenName.Name().ValueText(), name)
		}
	}
}

func TestParseRowSchemaProducesUpstreamStyleDiagnostics(t *testing.T) {
	missingType := ParseRowSchema("(Account:)")
	if missingType == nil {
		t.Fatalf("missing type schema returned nil")
	}
	if diagnostics := missingType.ContainedDiagnostics(); len(diagnostics) == 0 || diagnostics[0].Code != "KS006" {
		t.Fatalf("missing type diagnostics = %#v", diagnostics)
	} else if diagnosticsContain(diagnostics, "KS001") {
		t.Fatalf("missing type should not also synthesize a missing-token diagnostic: %#v", diagnostics)
	}

	invalidType := ParseRowSchema("(Account: float)")
	if invalidType == nil {
		t.Fatalf("invalid type schema returned nil")
	}
	if diagnostics := invalidType.ContainedDiagnostics(); len(diagnostics) == 0 || diagnostics[0].Code != "KS160" {
		t.Fatalf("invalid type diagnostics = %#v", diagnostics)
	}
}
