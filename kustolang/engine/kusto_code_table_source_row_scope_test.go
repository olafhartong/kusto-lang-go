package kustolang

import (
	"strings"
	"testing"
)

func TestKustoCodeUnionAndInvokeClearIncomingRowProfiles(t *testing.T) {
	for _, profile := range kustoCodeTableSourceProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			unionText := `Left | union Right`
			union := profile.parse(unionText, kustoCodeTableSourceGlobals())
			if diagnostics := union.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid union diagnostics = %#v", diagnostics)
			}
			assertMetadataWithoutRowColumns(t, union, strings.Index(unionText, "Right"))

			malformedUnionText := `Left | union Right, `
			malformedUnion := profile.parse(malformedUnionText, kustoCodeTableSourceGlobals())
			if len(malformedUnion.GetDiagnostics()) == 0 {
				t.Fatal("trailing union operand produced no diagnostics")
			}
			assertMetadataWithoutRowColumns(t, malformedUnion, len(malformedUnionText))
			unionItems := NewKustoCodeServiceForCode(malformedUnion).GetCompletionItems(len(malformedUnionText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, unionItems, "Left", "Right")
			assertCompletionExcludes(t, unionItems, "Common", "LeftOnly", "RightOnly")

			invokeText := `Left | invoke Transform(1)`
			invoke := profile.parse(invokeText, kustoCodeTableSourceGlobals())
			if diagnostics := invoke.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid invoke diagnostics = %#v", diagnostics)
			}
			functionStart := strings.Index(invokeText, "Transform")
			if columns := invoke.GetColumnsInScope(functionStart); columns == nil || columns.SymbolName() != "Left" {
				t.Fatalf("invoke strict function boundary row = %#v, want Left", columns)
			}
			assertMetadataWithoutRowColumns(t, invoke, functionStart+1)

			malformedInvokeText := `Left | invoke Transform(`
			malformedInvoke := profile.parse(malformedInvokeText, kustoCodeTableSourceGlobals())
			if len(malformedInvoke.GetDiagnostics()) == 0 {
				t.Fatal("missing invoke argument produced no diagnostics")
			}
			assertMetadataWithoutRowColumns(t, malformedInvoke, len(malformedInvokeText))
			invokeItems := NewKustoCodeServiceForCode(malformedInvoke).GetCompletionItems(len(malformedInvokeText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, invokeItems, "Transform(source, value)")
			assertCompletionExcludes(t, invokeItems, "Common", "LeftOnly", "RightOnly")
		})
	}
}

func TestKustoCodeSearchRowScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeTableSourceProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			pipedText := `Left | search Common == "x" and LeftOnly > 0`
			piped := profile.parse(pipedText, kustoCodeTableSourceGlobals())
			if diagnostics := piped.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid piped search diagnostics = %#v", diagnostics)
			}
			assertRowColumnNames(t, piped.GetColumnsInScope(strings.Index(pipedText, "Common")), "Common", "LeftOnly")

			rootText := `search LeftOnly > 0 or RightOnly`
			root := profile.parse(rootText, kustoCodeTableSourceGlobals())
			if diagnostics := root.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid root search diagnostics = %#v", diagnostics)
			}
			assertRowColumnNames(t, root.GetColumnsInScope(strings.Index(rootText, "RightOnly")), "Common", "LeftOnly", "RightOnly")

			explicitText := `search in (Left, Right) Common == "x"`
			explicit := profile.parse(explicitText, kustoCodeTableSourceGlobals())
			if diagnostics := explicit.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid explicit search diagnostics = %#v", diagnostics)
			}
			if columns := explicit.GetColumnsInScope(strings.Index(explicitText, "Right")); columns != nil {
				t.Fatalf("search input row scope = %#v, want nil", columns)
			}
			if columns := explicit.GetColumnsInScope(strings.Index(explicitText, "Common")); columns != nil {
				t.Fatalf("explicit search condition row scope = %#v, want nil per ContextBuilder", columns)
			}

			malformedText := `search in (Left, Right) `
			malformed := profile.parse(malformedText, kustoCodeTableSourceGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("missing search predicate produced no diagnostics")
			}
			if columns := malformed.GetColumnsInScope(len(malformedText)); columns != nil {
				t.Fatalf("malformed explicit search row scope = %#v, want nil", columns)
			}
		})
	}
}

func TestKustoCodeFindCombinedPredicateScopeProfiles(t *testing.T) {
	for _, profile := range kustoCodeTableSourceProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			text := `find in (Left, Right) where Common == "x" and RightOnly`
			code := profile.parse(text, kustoCodeTableSourceGlobals())
			if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid find diagnostics = %#v", diagnostics)
			}
			assertRowColumnNames(t, code.GetColumnsInScope(strings.Index(text, "Common")), "Common", "LeftOnly", "RightOnly")

			impliedText := `find where LeftOnly > 0 and RightOnly`
			implied := profile.parse(impliedText, kustoCodeTableSourceGlobals())
			if diagnostics := implied.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid implied find diagnostics = %#v", diagnostics)
			}
			assertRowColumnNames(t, implied.GetColumnsInScope(strings.Index(impliedText, "RightOnly")), "Common", "LeftOnly", "RightOnly")

			malformedText := `find in (Left, Right) where `
			malformed := profile.parse(malformedText, kustoCodeTableSourceGlobals())
			if len(malformed.GetDiagnostics()) == 0 {
				t.Fatal("missing find predicate produced no diagnostics")
			}
			assertRowColumnNames(t, malformed.GetColumnsInScope(len(malformedText)), "Common", "LeftOnly", "RightOnly")
			items := NewKustoCodeServiceForCode(malformed).GetCompletionItems(len(malformedText), CompletionOptionsDefault, CancellationToken{}).Items
			assertCompletionIncludes(t, items, "Common", "LeftOnly", "RightOnly")

			viewText := `let V = view () { datatable(ViewOnly:string)["x"] }; find where `
			viewCode := profile.parse(viewText, kustoCodeTableSourceGlobals())
			if len(viewCode.GetDiagnostics()) == 0 {
				t.Fatal("missing find predicate after local view produced no diagnostics")
			}
			if columns := viewCode.GetColumnsInScope(len(viewText)); columns == nil || columns.GetColumn("ViewOnly") == nil {
				t.Fatalf("implied local-view row scope = %#v", columns)
			}
		})
	}
}

func TestKustoCodeToScalarAndToTableClearOuterRowProfiles(t *testing.T) {
	for _, profile := range kustoCodeTableSourceProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			for _, function := range []string{"toscalar", "totable"} {
				text := `Left | where isnotnull(` + function + `(`
				code := profile.parse(text, kustoCodeTableSourceGlobals())
				if len(code.GetDiagnostics()) == 0 {
					t.Fatalf("missing %s body produced no diagnostics", function)
				}
				assertMetadataWithoutRowColumns(t, code, len(text))
				items := NewKustoCodeServiceForCode(code).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
				assertCompletionIncludes(t, items, "Left", "Right")
				assertCompletionExcludes(t, items, "Common", "LeftOnly", "RightOnly")
			}

			nestedText := `Left | where isnotnull(toscalar(Right | where RightOnly | count))`
			nested := profile.parse(nestedText, kustoCodeTableSourceGlobals())
			if diagnostics := nested.GetDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("valid nested toscalar diagnostics = %#v", diagnostics)
			}
			assertRowColumnNames(t, nested.GetColumnsInScope(strings.Index(nestedText, "RightOnly")), "Common", "RightOnly")
		})
	}
}

type kustoCodeTableSourceProfile struct {
	name  string
	parse func(string, *GlobalState) *KustoCode
}

func kustoCodeTableSourceProfiles() []kustoCodeTableSourceProfile {
	return []kustoCodeTableSourceProfile{{"current", func(text string, globals *GlobalState) *KustoCode {
		return ParseAndAnalyze(text, globals)
	}}}
}

func kustoCodeTableSourceGlobals() *GlobalState {
	left := NewTableSymbol("Left",
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("LeftOnly", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right",
		NewColumnSymbol("Common", ScalarTypes.String.Type()),
		NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type()))
	output := NewTableSymbol("Output", NewColumnSymbol("Result", ScalarTypes.Long.Type()))
	transform := NewFunctionSymbol("Transform", NewSymbolSignature(output,
		NewParameterKind("source", ParameterTypeTabular),
		NewParameter("value", ScalarTypes.Long.Type())))
	return NewGlobalState(NewDatabaseSymbol("db", left, right, transform))
}

func assertMetadataWithoutRowColumns(t *testing.T, code *KustoCode, position int) {
	t.Helper()
	if columns := code.GetColumnsInScope(position); columns != nil {
		t.Fatalf("row scope = %#v, want nil", columns)
	}
	symbols := code.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll)
	for _, name := range []string{"Left", "Right", "Transform", "strlen"} {
		if !containsScopeSymbolName(symbols, name) {
			t.Fatalf("metadata scope missing %q", name)
		}
	}
	for _, name := range []string{"Common", "LeftOnly", "RightOnly"} {
		if containsScopeSymbolName(symbols, name) {
			t.Fatalf("metadata scope unexpectedly contains row column %q", name)
		}
	}
}

func assertRowColumnNames(t *testing.T, table *TableSymbol, names ...string) {
	t.Helper()
	if table == nil {
		t.Fatalf("row scope is nil, want %v", names)
	}
	for _, name := range names {
		if table.GetColumn(name) == nil {
			t.Fatalf("row scope missing %q: %#v", name, table.Columns())
		}
	}
}
