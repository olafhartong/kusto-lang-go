package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodePartitionKeyScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeOperatorLocalProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			validText := `T | partition by Tenant { Other | extend K=Tenant }`
			valid := profile.parse(validText, kustoCodeOperatorLocalGlobals())
			if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid partition diagnostics = %#v", diagnostics)
			}
			position := strings.LastIndex(validText, "Tenant")
			if got := valid.GetSpeculativeReferencedSymbol(position, "Tenant", SymbolMatchColumn|SymbolMatchLocal); got == nil {
				t.Fatal("partition key is absent from the braced operand scope")
			}

			malformedText := `T | partition by Tenant { Other | extend K=`
			malformed := profile.parse(malformedText, kustoCodeOperatorLocalGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("malformed partition operand produced no diagnostics")
			}
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "Tenant")
		})
	}
}

func TestKustoCodeMacroExpandScope(t *testing.T) {
	for _, test := range []struct {
		text string
		name string
	}{
		{`macro-expand Targets as Scope (Scope | take 1)`, "Scope"},
		{`macro-expand Targets (Targets | take 1)`, "Targets"},
	} {
		valid := ParseAndAnalyze(test.text, kustoCodeOperatorLocalGlobals())
		if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("valid macro-expand diagnostics = %#v", diagnostics)
		}
		position := strings.LastIndex(test.text, test.name)
		if got := valid.GetSpeculativeReferencedSymbol(position, test.name, SymbolMatchLocal|SymbolMatchEntityGroupElement); got == nil {
			t.Fatalf("macro-expand scope %q is absent inside the body", test.name)
		}
	}

	malformedText := `macro-expand Targets as Scope (print X=)`
	malformed := ParseAndAnalyze(malformedText, kustoCodeOperatorLocalGlobals())
	if len(malformed.GetDiagnostics()) == 0 {
		t.Fatal("malformed macro-expand body produced no diagnostics")
	}
	completionPosition := len(malformedText) - 1
	if got := malformed.GetSpeculativeReferencedSymbol(completionPosition, "Scope", SymbolMatchLocal|SymbolMatchEntityGroupElement); got == nil {
		t.Fatal("malformed macro-expand body lost its scope symbol")
	}
	items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(completionPosition, CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "Scope")
}

func TestKustoCodeInlineExternalTableDeclarationScopes(t *testing.T) {
	validText := `inline_external_table(A:string, B:long) kind=storage partition by (P:string=A) pathformat=(P) dataformat=csv ("u")`
	valid := ParseAndAnalyze(validText, nil)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid inline external table diagnostics = %#v", diagnostics)
	}
	partitionPosition := strings.Index(validText, "=A)") + 1
	partitionSymbols := valid.GetSymbolsInScope(partitionPosition, SymbolMatchAny, IncludeFunctionKindAll)
	if !containsScopeSymbolName(partitionSymbols, "A") || containsScopeSymbolName(partitionSymbols, "P") {
		t.Fatalf("partition scope = %#v, want schema A without partition P", partitionSymbols)
	}
	pathPosition := strings.Index(validText, "pathformat=(P") + len("pathformat=(")
	pathSymbols := valid.GetSymbolsInScope(pathPosition, SymbolMatchAny, IncludeFunctionKindAll)
	if !containsScopeSymbolName(pathSymbols, "P") || containsScopeSymbolName(pathSymbols, "A") {
		t.Fatalf("pathformat scope = %#v, want partition P without schema A", pathSymbols)
	}

	malformedText := `inline_external_table(A:string) kind=storage partition by (P:string=A) pathformat=(`
	malformed := ParseAndAnalyze(malformedText, nil)
	if len(malformed.GetDiagnostics()) == 0 {
		t.Fatal("malformed pathformat produced no diagnostics")
	}
	items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "P")
	assertCompletionExcludes(t, items, "A")
}

func TestKustoCodeMaterializedViewCombineDeltaAggregationScope(t *testing.T) {
	validText := `materialized-view-combine("v") base(Base) delta(Delta) aggregations(summarize Total=sum(Value) by Key)`
	valid := ParseAndAnalyze(validText, kustoCodeOperatorLocalGlobals())
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid materialized-view-combine diagnostics = %#v", diagnostics)
	}
	assertRowColumnNames(t, valid.GetColumnsInScope(strings.Index(validText, "Value")), "Key", "Value")
	if table := valid.GetColumnsInScope(strings.Index(validText, "Value")); table.GetColumn("BaseOnly") != nil {
		t.Fatalf("aggregation row scope contains base-only column: %#v", table.Columns())
	}

	malformedText := `materialized-view-combine("v") base(Base) delta(Delta) aggregations(summarize Total=sum(`
	malformed := ParseAndAnalyze(malformedText, kustoCodeOperatorLocalGlobals())
	if len(malformed.GetDiagnostics()) == 0 {
		t.Fatal("malformed materialized-view-combine aggregation produced no diagnostics")
	}
	items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
	assertCompletionIncludes(t, items, "Key", "Value")
	assertCompletionExcludes(t, items, "BaseOnly")
}

type kustoCodeOperatorLocalProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeOperatorLocalProfiles() []kustoCodeOperatorLocalProfile {
	return []kustoCodeOperatorLocalProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeOperatorLocalGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("Tenant", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	base := NewTableSymbol("Base", NewColumnSymbol("BaseOnly", ScalarTypes.Bool.Type()))
	delta := NewTableSymbol("Delta",
		NewColumnSymbol("Key", ScalarTypes.String.Type()),
		NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	other := NewTableSymbol("Other", NewColumnSymbol("OtherValue", ScalarTypes.Long.Type()))
	group := NewEntityGroupSymbol("Targets", table)
	return NewGlobalState(NewDatabaseSymbol("db", table, base, delta, other, group))
}
