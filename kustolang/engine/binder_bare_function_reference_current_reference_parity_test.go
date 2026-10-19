package kustolang

import "testing"

func TestNativeCurrentBareFunctionReferenceReferenceParity(t *testing.T) {
	tests := []struct {
		name, function, code, typ string
	}{
		{name: "parameter return", function: "range", code: "KS117", typ: "unknown"},
		{name: "custom dynamic return", function: "pack_array", code: "KS117", typ: "dynamic"},
		{name: "common return", function: "case", code: "KS117", typ: "unknown"},
		{name: "declared return", function: "now", code: "KS117", typ: "datetime"},
		{name: "aggregate outside aggregate scope", function: "sum", code: "KS142", typ: "unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := "print X=" + test.function
			code := ParseAndAnalyze(query, DefaultGlobalState)
			diagnostics := code.GetDiagnostics()
			if len(diagnostics) != 1 || diagnostics[0].Code != test.code || diagnostics[0].Start != 8 || diagnostics[0].Length != len(test.function) {
				t.Fatalf("diagnostics = %#v, want %s at 8+%d", diagnostics, test.code, len(test.function))
			}
			print := firstBoundNode[*PrintOperator](t, code.Query)
			table, _ := TableSymbolForType(ExpressionRawResultType(print))
			if table == nil || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "X" || table.Columns()[0].Type.SymbolName() != test.typ {
				t.Fatalf("print result = %#v, want X:%s", table, test.typ)
			}
		})
	}
}

func TestNativeFunctionCatalogRespectsScopeKind(t *testing.T) {
	binder := NewBinder(DefaultGlobalState)
	if info := binder.ResolveName("sum", SymbolMatchFunction, false); info.ReferencedSymbol() != Error {
		t.Fatalf("normal sum = %#v, want unresolved", info)
	}
	binder.scopeKind = ScopeKindAggregate
	if info := binder.ResolveName("sum", SymbolMatchFunction, false); info.ReferencedSymbol() != Aggregates.Sum {
		t.Fatalf("aggregate sum = %#v, want aggregate symbol", info)
	}
	if info := binder.ResolveName("strlen", SymbolMatchFunction, false); info.ReferencedSymbol() != Functions.Strlen {
		t.Fatalf("aggregate strlen fallback = %#v, want scalar function", info)
	}
}
