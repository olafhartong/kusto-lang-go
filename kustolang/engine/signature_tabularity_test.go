package kustolang

import "testing"

func TestSignatureComputeTabularityUsesSyntaxThenNativeBinding(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	databaseFunction := NewComputedFunctionSymbol("DatabaseRows", "T", TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, databaseFunction))

	tests := []struct {
		name string
		body string
		want Tabularity
	}{
		{name: "literal syntax", body: "1 + 2", want: TabularityScalar},
		{name: "pipe syntax", body: "T | project Value", want: TabularityTabular},
		{name: "database function syntax", body: "DatabaseRows()", want: TabularityTabular},
		{name: "name binding scalar", body: "toscalar(T | count)", want: TabularityScalar},
		{name: "name binding tabular", body: "T", want: TabularityTabular},
		{name: "braced body", body: "{ T }", want: TabularityTabular},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signature := NewComputedSignatureInferred(test.body)
			if signature.Tabularity != TabularityUnknown {
				t.Fatalf("initial tabularity = %s, want existing public Unknown behavior", signature.Tabularity)
			}
			signature.ComputeTabularity(globals)
			if signature.Tabularity != test.want {
				t.Fatalf("tabularity = %s, want %s", signature.Tabularity, test.want)
			}
		})
	}
}

func TestSignatureComputeTabularityPreservesExplicitAndUsesRepresentativeParameters(t *testing.T) {
	explicitUnknown := NewComputedSignature("print Value=1", TabularityUnknown)
	explicitUnknown.ComputeTabularity(DefaultGlobalState)
	if explicitUnknown.Tabularity != TabularityUnknown {
		t.Fatalf("explicit Unknown changed to %s", explicitUnknown.Tabularity)
	}

	input := NewParameterKind("Input", ParameterTypeTabular)
	parameterized := NewComputedSignatureInferred("Input | take 1", input)
	parameterized.ComputeTabularity(DefaultGlobalState)
	if parameterized.Tabularity != TabularityTabular {
		t.Fatalf("parameterized tabularity = %s, want Tabular", parameterized.Tabularity)
	}

	malformed := NewComputedSignatureInferred("unknown_function(")
	malformed.ComputeTabularity(DefaultGlobalState)
	if malformed.Tabularity != TabularityUnknown {
		t.Fatalf("malformed tabularity = %s, want Unknown", malformed.Tabularity)
	}
}

func TestNewComputedFunctionSymbolInferredAuthorsUnspecifiedSignature(t *testing.T) {
	function := NewComputedFunctionSymbolInferred("Rows", "print Value=1")
	if function.SymbolTabularity() != TabularityUnknown {
		t.Fatalf("initial function tabularity = %s", function.SymbolTabularity())
	}
	function.Signatures()[0].ComputeTabularity(DefaultGlobalState)
	if function.SymbolTabularity() != TabularityTabular {
		t.Fatalf("computed function tabularity = %s, want Tabular", function.SymbolTabularity())
	}
}
