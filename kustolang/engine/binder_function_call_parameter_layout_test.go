package kustolang

import "testing"

func binderFunctionArgumentLayoutProfiles() []binderFunctionArgumentLayoutProfile {
	return []binderFunctionArgumentLayoutProfile{{
		name: "current",
		parse: func(text string, globals *GlobalState) *KustoCode {
			return ParseAndAnalyze(text, globals)
		},
	}}
}

func TestTreeBinderFunctionCallUsesCustomArgumentParameterLayout(t *testing.T) {
	for _, profile := range binderFunctionArgumentLayoutProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals, sourceOnly := binderFunctionArgumentLayoutGlobals()

			valid := profile.parse(`Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, true)`, globals)
			if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid ipv4_lookup diagnostics = %#v", diagnostics)
			}

			malformed := profile.parse(`Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, return_unmatched=SourceOnly)`, globals)
			reference := findBoundNameReference(malformed.Syntax(), "SourceOnly")
			if reference == nil || reference.ReferencedSymbol() != sourceOnly {
				t.Fatalf("return_unmatched expression symbol = %#v, want source column %#v", reference, sourceOnly)
			}
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("non-literal return_unmatched produced no diagnostics")
			}
		})
	}
}

func TestTreeBinderFunctionCallCustomLayoutMalformedRecovery(t *testing.T) {
	for _, profile := range binderFunctionArgumentLayoutProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, query := range []string{
				`Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, true, )`,
				`Source | evaluate ipv4_lookup(Lookup, SourceIP, LookupIP, true`,
			} {
				code := profile.parse(query, binderFunctionArgumentLayoutGlobalsOnly())
				if len(code.GetDiagnostics()) == 0 {
					t.Fatalf("malformed query %q produced no diagnostics", query)
				}
				if reference := findBoundNameReference(code.Syntax(), "LookupIP"); reference == nil {
					t.Fatalf("lookup key was not bound after recovery for %q", query)
				}
			}
		})
	}
}

func TestTreeBinderDependentColumnWithoutParameterTableUsesAggregateScope(t *testing.T) {
	for _, profile := range binderFunctionArgumentLayoutProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			code := profile.parse(`Source | evaluate ipv4_lookup(MissingLookup, SourceIP, count())`, binderFunctionArgumentLayoutGlobalsOnly())
			var call *FunctionCallExpression
			for _, candidate := range GetDescendants[*FunctionCallExpression](code.Syntax()) {
				if NameReferenceSimpleName(candidate.Name()) == "count" {
					call = candidate
					break
				}
			}
			if call == nil || call.ReferencedSymbol() != Aggregates.Count {
				t.Fatalf("fallback aggregate call = %#v symbol %#v", call, func() Symbol {
					if call == nil {
						return nil
					}
					return call.ReferencedSymbol()
				}())
			}
			if len(code.GetDiagnostics()) == 0 {
				t.Fatal("invalid dependent-column call produced no diagnostics")
			}
		})
	}
}

func TestTreeBinderFunctionCallUsesTupleElementArgumentScope(t *testing.T) {
	globals := kustoCodeFunctionArgumentGraphGlobals()
	for _, query := range []string{
		graphScopePrefix() + `graph-match (a)-[e]->(b) where all(a, Enabled)`,
		graphScopePrefix() + `graph-match (a)-[e]->(b) where map(a, Label) != dynamic([])`,
	} {
		code := ParseAndAnalyze(query, globals)
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("graph function query %q diagnostics = %#v", query, diagnostics)
		}
		name := "Enabled"
		if findBoundNameReference(code.Syntax(), name) == nil {
			name = "Label"
		}
		if reference := findBoundNameReference(code.Syntax(), name); reference == nil {
			t.Fatalf("graph tuple element %q was not bound", name)
		}
	}

	for _, query := range []string{
		graphScopePrefix() + `graph-match (a)-[e]->(b) where all(a, Enabled, )`,
		graphScopePrefix() + `graph-match (a)-[e]->(b) where map(a, Label`,
	} {
		code := ParseAndAnalyze(query, globals)
		if len(code.GetDiagnostics()) == 0 {
			t.Fatalf("malformed graph function query %q produced no diagnostics", query)
		}
	}
}

type binderFunctionArgumentLayoutProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func binderFunctionArgumentLayoutGlobals() (*GlobalState, *ColumnSymbol) {
	sourceOnly := NewColumnSymbol("SourceOnly", ScalarTypes.Long.Type())
	source := NewTableSymbol("Source",
		NewColumnSymbol("SourceIP", ScalarTypes.String.Type()),
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		sourceOnly)
	lookup := NewTableSymbol("Lookup",
		NewColumnSymbol("LookupIP", ScalarTypes.String.Type()),
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("LookupOnly", ScalarTypes.Bool.Type()))
	return NewGlobalState(NewDatabaseSymbol("db", source, lookup)), sourceOnly
}

func binderFunctionArgumentLayoutGlobalsOnly() *GlobalState {
	globals, _ := binderFunctionArgumentLayoutGlobals()
	return globals
}
