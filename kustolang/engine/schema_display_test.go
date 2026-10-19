package kustolang

import "testing"

func TestCanBeIdentifierNameUsesLexerAndKeywordFacts(t *testing.T) {
	tests := map[string]bool{
		"Account":                 true,
		"Delta_Δ":                 false,
		"$table":                  true,
		"access":                  true,
		"application":             true,
		"where":                   false,
		"true":                    false,
		"1bad":                    false,
		"bad name":                false,
		"bad-hyphen":              false,
		string([]byte{0xff, 'x'}): false,
	}
	for name, want := range tests {
		if got := CanBeIdentifier(name); got != want {
			t.Fatalf("CanBeIdentifier(%q) = %t, want %t", name, got, want)
		}
		if got := CanBeIdentifierName(name); got != want {
			t.Fatalf("compatibility CanBeIdentifierName(%q) = %t, want %t", name, got, want)
		}
	}
}

func TestCanBeIdentifierNameUsesDialectSpecificKeywordFacts(t *testing.T) {
	if !CanBeIdentifierName("application") {
		t.Fatal("application should be a query identifier")
	}
	if CanBeIdentifierName("application", KustoDialectEngineCommand) {
		t.Fatal("application should need brackets in engine commands")
	}
	if CanBeIdentifierName("callouts", KustoDialectEngineCommand) {
		t.Fatal("callouts should need brackets in engine commands")
	}
	if CanBeIdentifierName("graph_query", KustoDialectEngineCommand) || CanBeIdentifierName("snapshot_indexing", KustoDialectEngineCommand) {
		t.Fatal("new upstream engine command keywords should need brackets")
	}
	if !CanBeIdentifierName("access", KustoDialectEngineCommand) {
		t.Fatal("access should remain an engine command identifier")
	}
	if !KustoFactsEngineCommandKeywordNeedsBrackets("GB") || KustoFactsEngineCommandKeywordNeedsBrackets("gb") {
		t.Fatal("engine command keyword set should preserve upstream case sensitivity")
	}
	if CanBeIdentifierName("Account", KustoDialectClusterManagerCommand) {
		t.Fatal("cluster manager command names should always need brackets")
	}
	if CanBeIdentifierName("Account", KustoDialectDataManagerCommand) {
		t.Fatal("data manager command names should always need brackets")
	}
	if !CanBeIdentifier("where", KustoDialectUnknown) {
		t.Fatal("unknown dialect should apply only lexical identifier restrictions")
	}
}

func TestBracketNameIfNecessaryUsesDialect(t *testing.T) {
	if got, want := BracketNameIfNecessary("application"), "application"; got != want {
		t.Fatalf("query bracket = %q, want %q", got, want)
	}
	if got, want := BracketNameIfNecessary("application", KustoDialectEngineCommand), "['application']"; got != want {
		t.Fatalf("engine bracket = %q, want %q", got, want)
	}
	if got, want := BracketNameIfNecessary("Account", KustoDialectDataManagerCommand), "['Account']"; got != want {
		t.Fatalf("data manager bracket = %q, want %q", got, want)
	}
	if got, want := BracketNameIfNecessary("Delta_Δ"), "['Delta_Δ']"; got != want {
		t.Fatalf("Unicode query bracket = %q, want %q", got, want)
	}
}

func TestSchemaDisplayBracketsNamesAndTypes(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("bad name", ScalarTypes.Long.Type()),
		NewColumnSymbol("where", ScalarTypes.DynamicArray.Type()),
	)
	got := SchemaDisplayText(table)
	want := "(Account: string, ['bad name']: long, ['where']: dynamic)"
	if got != want {
		t.Fatalf("SchemaDisplayText(table) = %q, want %q", got, want)
	}
}

func TestSchemaDisplayFunctionParameters(t *testing.T) {
	fn := NewScalarFunctionSymbol("f", ScalarTypes.Bool.Type(),
		NewParameter("value", ScalarTypes.String.Type()),
		NewParameterKind("rows", ParameterTypeTabular),
		NewParameterTypes("choice", []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()}),
	)
	got := SchemaDisplayText(fn)
	want := "(value: string, rows: (*), choice: dynamic)"
	if got != want {
		t.Fatalf("SchemaDisplayText(fn) = %q, want %q", got, want)
	}
}

func TestStringLiteralAndBracketedNameEscaping(t *testing.T) {
	if got, want := GetStringLiteral("plain"), "'plain'"; got != want {
		t.Fatalf("GetStringLiteral plain = %q, want %q", got, want)
	}
	if got, want := GetStringLiteral("can't"), `"can't"`; got != want {
		t.Fatalf("GetStringLiteral quote = %q, want %q", got, want)
	}
	if got, want := GetStringLiteral("line\nnext\\slash"), "'line\\nnext\\\\slash'"; got != want {
		t.Fatalf("GetStringLiteral escapes = %q, want %q", got, want)
	}
	if got, want := BracketNameIfNecessary("bad name"), "['bad name']"; got != want {
		t.Fatalf("BracketNameIfNecessary = %q, want %q", got, want)
	}
	if got, want := GetBracketedName("can't"), `["can't"]`; got != want {
		t.Fatalf("GetBracketedName quote = %q, want %q", got, want)
	}
	if got, want := GetSingleQuotedStringLiteral("'\\\a\b\f\n\r\t\v"), "'\\'\\\\\\a\\b\\f\\n\\r\\t\v'"; got != want {
		t.Fatalf("GetSingleQuotedStringLiteral escapes = %q, want %q", got, want)
	}
	if got, want := GetDoubleQuotedStringLiteral("\"'"), `"\"'"`; got != want {
		t.Fatalf("GetDoubleQuotedStringLiteral escapes = %q, want %q", got, want)
	}
}

func TestKustoFactsASCIIIdentifierRequirementAffectsQueryDiagnostics(t *testing.T) {
	valid := Parse("datatable(['Delta_Δ']:long)[1] | project ['Delta_Δ']")
	if diagnostics := valid.GetSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("bracketed non-ASCII name diagnostics = %#v", diagnostics)
	}

	invalid := Parse("datatable(Delta_Δ:long)[1] | project Delta_Δ")
	if diagnostics := invalid.GetSyntaxDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("unbracketed non-ASCII name should produce syntax diagnostics")
	}
}

func TestTableSymbolFromSchemaRoundTripsSchemaDisplay(t *testing.T) {
	table, err := TableSymbolFromSchema("Account: string, ['bad name']: long, [\"can't\"]: dynamic")
	if err != nil {
		t.Fatalf("TableSymbolFromSchema returned error: %v", err)
	}
	if got := SchemaDisplayText(table); got != "(Account: string, ['bad name']: long, [\"can't\"]: dynamic)" {
		t.Fatalf("round-trip schema = %q", got)
	}
	columns := table.Columns()
	if len(columns) != 3 || columns[2].SymbolName() != "can't" || columns[2].Type != ScalarTypes.Dynamic.Type() {
		t.Fatalf("columns = %#v", columns)
	}
}

func TestTableSymbolFromSchemaPreservesRecoveredRowsLikeUpstream(t *testing.T) {
	table, err := TableSymbolFromSchema("(:string, Account:float, Account:long)")
	if err != nil {
		t.Fatalf("recovered schema returned error: %v", err)
	}
	columns := table.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "" || columns[0].Type != ScalarTypes.String.Type() ||
		columns[1].SymbolName() != "Account" || columns[1].Type != ScalarTypes.Real.Type() {
		t.Fatalf("recovered schema columns = %#v", columns)
	}
}
