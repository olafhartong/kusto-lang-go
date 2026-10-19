package kustolang

import (
	"reflect"
	"testing"
)

func TestKustoCodeAnalyzePublishesResultTypeAndPreservesIdentity(t *testing.T) {
	parsed := Parse("print x=1")
	if parsed.HasSemantics() || parsed.ResultType != nil {
		t.Fatalf("parse-only state = semantics:%t result:%#v", parsed.HasSemantics(), parsed.ResultType)
	}

	analyzed := parsed.Analyze()
	if analyzed == parsed || !analyzed.HasSemantics() || analyzed.ResultType == nil {
		t.Fatalf("analyzed state = same:%t semantics:%t result:%#v", analyzed == parsed, analyzed.HasSemantics(), analyzed.ResultType)
	}
	table := terminalResultTable(analyzed)
	if table == nil || len(table.Columns()) != 1 || table.Columns()[0].SymbolName() != "x" || table.Columns()[0].Type != ScalarTypes.Long.Type() {
		t.Fatalf("result table = %#v", table)
	}
	if analyzed.Analyze() != analyzed {
		t.Fatal("already analyzed code should preserve identity for the same globals")
	}
	if parsed.HasSemantics() || parsed.ResultType != nil {
		t.Fatal("Analyze mutated the parse-only instance")
	}
}

func TestKustoCodeWithGlobalsPreservesRequestedAnalysisState(t *testing.T) {
	firstTable := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	secondTable := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	first := NewGlobalState(NewDatabaseSymbol("first", firstTable))
	second := NewGlobalState(NewDatabaseSymbol("second", secondTable))

	parsed := ParseWithGlobals("T | project A", first)
	if parsed.WithGlobals(parsed.Globals) != parsed {
		t.Fatal("parse-only code should preserve identity for the same globals")
	}
	reparsed := parsed.WithGlobals(second)
	if reparsed == parsed || reparsed.HasSemantics() || reparsed.ResultType != nil || reparsed.Globals == parsed.Globals {
		t.Fatalf("reparsed state = same:%t semantics:%t result:%#v globalsSame:%t", reparsed == parsed, reparsed.HasSemantics(), reparsed.ResultType, reparsed.Globals == parsed.Globals)
	}

	analyzed := parsed.Analyze()
	rebound := analyzed.WithGlobals(second)
	if rebound == analyzed || !rebound.HasSemantics() {
		t.Fatalf("rebound state = same:%t semantics:%t", rebound == analyzed, rebound.HasSemantics())
	}
	result := terminalResultTable(rebound)
	if result == nil || len(result.Columns()) != 1 || result.Columns()[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("rebound result = %#v", result)
	}
	original := terminalResultTable(analyzed)
	if original == nil || original.Columns()[0].Type != ScalarTypes.Long.Type() {
		t.Fatalf("original result changed = %#v", original)
	}
}

func TestKustoCodeAnalyzeMalformedSyntaxRetainsDiagnostics(t *testing.T) {
	parsed := Parse("T | where (")
	analyzed := parsed.Analyze()
	if !analyzed.HasSemantics() || len(analyzed.GetSyntaxDiagnostics()) == 0 || len(analyzed.GetDiagnostics()) == 0 {
		t.Fatalf("malformed analysis = semantics:%t syntax:%#v all:%#v", analyzed.HasSemantics(), analyzed.GetSyntaxDiagnostics(), analyzed.GetDiagnostics())
	}
	if parsed.HasSemantics() {
		t.Fatal("malformed Analyze mutated the source instance")
	}
}

func TestKustoCodeSeparatesSyntaxAndSemanticDiagnosticsForMalformedQuery(t *testing.T) {
	const text = `datatable(A:long)[1] | project Missing, B=`
	code := ParseAndAnalyze(text)

	syntax := code.GetSyntaxDiagnostics()
	if len(syntax) != 1 || syntax[0].Code != "KS006" || syntax[0].Start != 42 || syntax[0].Length != 0 {
		t.Fatalf("syntax diagnostics = %#v", syntax)
	}
	all := code.GetDiagnostics()
	if len(all) != 2 || all[0].Code != "KS142" || all[0].Start != 31 || all[0].Length != 7 ||
		all[1].Code != "KS006" || all[1].Start != 42 || all[1].Length != 0 {
		t.Fatalf("all diagnostics = %#v", all)
	}
	if !reflect.DeepEqual(code.GetSyntaxDiagnostics(), syntax) || !reflect.DeepEqual(code.GetDiagnostics(), all) {
		t.Fatal("repeated diagnostic reads changed membership or ordering")
	}
}

func TestKustoCodeDiagnosticCachesBypassLaterCancellationAndOwnResults(t *testing.T) {
	code := ParseAndAnalyze(`datatable(A:long)[1] | project Missing, B=`)
	all := code.GetDiagnostics()
	syntax := code.GetSyntaxDiagnostics()
	canceled := NewCancellationTokenFunc(func() bool { return true })
	if got := code.GetDiagnosticsWithCancellation(canceled); !reflect.DeepEqual(got, all) {
		t.Fatalf("cached diagnostics = %#v, want %#v", got, all)
	}
	if got := code.GetSyntaxDiagnosticsWithCancellation(canceled); !reflect.DeepEqual(got, syntax) {
		t.Fatalf("cached syntax diagnostics = %#v, want %#v", got, syntax)
	}
	all[0] = NewDiagnostic("TEST", "mutated")
	syntax[0] = NewDiagnostic("TEST", "mutated")
	if code.GetDiagnostics()[0].Code == "TEST" || code.GetSyntaxDiagnostics()[0].Code == "TEST" {
		t.Fatal("diagnostic cache exposed mutable storage")
	}
}

func TestKustoCodeLifecycleHandlesCommandsAndDeclarationOnlyQueries(t *testing.T) {
	command := ParseAndAnalyze(".show version", NewGlobalState(nil))
	if !command.HasSemantics() || command.ResultType == nil || len(command.GetDiagnostics()) != 0 {
		t.Fatalf("command lifecycle = semantics:%t result:%#v diagnostics:%#v", command.HasSemantics(), command.ResultType, command.GetDiagnostics())
	}

	declaration := ParseAndAnalyze("let x=1;")
	if !declaration.HasSemantics() || declaration.ResultType != nil {
		t.Fatalf("declaration lifecycle = semantics:%t result:%#v", declaration.HasSemantics(), declaration.ResultType)
	}
}

func TestKustoCodeLifecycleReusesDefensiveLexicalTokens(t *testing.T) {
	first := NewGlobalState(NewDatabaseSymbol("first", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))))
	second := NewGlobalState(NewDatabaseSymbol("second", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))))
	parsed := ParseWithGlobals("T | project A", first)
	want := parsed.GetLexicalTokens()

	analyzed := parsed.Analyze(second)
	if !reflect.DeepEqual(analyzed.GetLexicalTokens(), want) {
		t.Fatalf("Analyze tokens changed:\ngot:  %#v\nwant: %#v", analyzed.GetLexicalTokens(), want)
	}
	rebound := analyzed.WithGlobals(first)
	if !reflect.DeepEqual(rebound.GetLexicalTokens(), want) {
		t.Fatalf("WithGlobals tokens changed:\ngot:  %#v\nwant: %#v", rebound.GetLexicalTokens(), want)
	}

	copy := parsed.GetLexicalTokens()
	copy[0].Text = "Changed"
	if reflect.DeepEqual(parsed.GetLexicalTokens(), copy) {
		t.Fatal("GetLexicalTokens exposed mutable stored tokens")
	}
}

func TestKustoCodeUnsafeAnalysisReportsKS245OutsideSyntaxDiagnostics(t *testing.T) {
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.MaxAnalysisDepth, 1)
	code := ParseAndAnalyze("print x=1", globals)
	if code.HasSemantics() || code.ResultType != nil {
		t.Fatalf("unsafe state = semantics:%t result:%#v", code.HasSemantics(), code.ResultType)
	}
	if diagnostics := code.GetSyntaxDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("syntax diagnostics = %#v, want none", diagnostics)
	}
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS245" {
		t.Fatalf("diagnostics = %#v, want KS245", diagnostics)
	}
}

func TestKustoCodeGetKindMatchesCodeKindDetection(t *testing.T) {
	for _, text := range []string{"print x=1", ".show version", "#database db\nT"} {
		if GetKind(text) != CodeKindOf(text) {
			t.Fatalf("GetKind(%q) = %q, want %q", text, GetKind(text), CodeKindOf(text))
		}
	}
}

func terminalResultTable(code *KustoCode) *TableSymbol {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	var statements *SeparatedSyntaxListOf[Statement]
	if code.Query != nil {
		statements = code.Query.Statements()
	} else if root, ok := WrapGeneratedSyntaxElement(code.Syntax()).(*CommandBlock); ok {
		statements = root.Statements()
	}
	if statements == nil || len(statements.Elements()) == 0 {
		return nil
	}
	last := statements.At(len(statements.Elements()) - 1)
	if last == nil {
		return nil
	}
	statement, ok := WrapGeneratedSyntaxElement(last.Element()).(*ExpressionStatement)
	if !ok || statement.Expression() == nil {
		return nil
	}
	table, _ := tableSymbolFromSymbol(expressionReferencedSymbol(statement.Expression()))
	return table
}
