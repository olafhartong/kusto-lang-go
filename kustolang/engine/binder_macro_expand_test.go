package kustolang

import (
	"strings"
	"testing"
)

func macroExpandTestGlobals() *GlobalState {
	t1 := NewTableSymbol("T1", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	t2 := NewTableSymbol("T2", NewColumnSymbol("A", ScalarTypes.Long.Type()), NewColumnSymbol("B", ScalarTypes.String.Type()))
	group := NewEntityGroupSymbol("Targets", t1, t2)
	return NewGlobalState(NewDatabaseSymbol("db", t1, t2, group))
}

func TestNativeTreeBinderMacroExpandInlineAndNamedGroupsUnifySchemas(t *testing.T) {
	for _, text := range []string{
		`macro-expand entity_group[T1, T2] as Scope (Scope)`,
		`macro-expand Targets as Scope (Scope)`,
	} {
		t.Run(text, func(t *testing.T) {
			code := ParseWithGlobals(text, macroExpandTestGlobals())
			query := code.Query
			binder := NewBinder(macroExpandTestGlobals())
			binder.TryBind(code.Syntax())
			macro := firstBoundNode[*MacroExpandOperator](t, query)
			if !macro.IsBound() {
				t.Fatalf("macro-expand unbound: group=%T groupBound=%t groupSymbol=%T", macro.EntityGroup(), simpleBinderSyntaxNode(macro.EntityGroup()).IsBound(), expressionReferencedSymbol(macro.EntityGroup()))
			}
			assertOrderedColumns(t, semanticTable(t, macro), "A", "B")
			assertDiagnosticCodes(t, macro)
			if got := len(macro.Alternates()); got != 0 {
				t.Fatalf("macro-expand alternates = %d, want 0", got)
			}
			statementList := macro.StatementList().SyntaxList
			if got := len(statementList.Alternates()); got != 1 {
				t.Fatalf("statement-list alternates = %d, want 1", got)
			}
			alternate := statementList.Alternates()[0]
			alternateTree := alternate.Tree()
			if alternateTree == nil || alternateTree.Original != code.Tree {
				t.Fatalf("alternate tree provenance = %#v", alternateTree)
			}
			if alternate.Kind() != SyntaxKindList || alternate.ChildCount() != 1 {
				t.Fatalf("alternate shape = %v with %d children, want full statement-list fragment", alternate.Kind(), alternate.ChildCount())
			}
			clonedList, ok := alternate.GetChild(0).(*SyntaxList)
			if !ok || clonedList == statementList || clonedList.Text(IncludeTriviaAll) != statementList.Text(IncludeTriviaAll) || clonedList.Tree() != alternateTree {
				t.Fatalf("alternate statement list = %#v, want owned full clone", alternate.GetChild(0))
			}
			before, _ := macro.GetSemanticInfo()
			binder.TryBind(code.Syntax())
			after, _ := macro.GetSemanticInfo()
			listAfter, _ := statementList.GetSemanticInfo()
			if before.ReferencedSymbol() != after.ReferencedSymbol() || len(after.Alternates) != 0 || len(listAfter.Alternates) != 1 || listAfter.Alternates[0] != alternate || listAfter.Alternates[0].Tree() != alternateTree {
				t.Fatalf("macro-expand rebinding changed semantics: before=%#v after=%#v", before, after)
			}
		})
	}
}

func TestNativeTreeBinderMacroExpandRelocatesAlternateDiagnostics(t *testing.T) {
	text := `macro-expand entity_group[T2, T1] as Scope (Scope | project B)`
	code := ParseAndAnalyze(text, macroExpandTestGlobals())
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS142" {
		t.Fatalf("macro-expand diagnostics = %#v, want one KS142", diagnostics)
	}
	wantStart := strings.LastIndex(text, "B")
	if diagnostics[0].Start != wantStart || diagnostics[0].Length != 1 {
		t.Fatalf("alternate diagnostic range = %d+%d, want %d+1", diagnostics[0].Start, diagnostics[0].Length, wantStart)
	}
}

func TestNativeTreeBinderMacroExpandProjectsDirectAlternateMemberDiagnostics(t *testing.T) {
	text := `macro-expand Targets as Scope (1)`
	code := ParseAndAnalyze(text, macroExpandTestGlobals())
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(code.Syntax(), DiagnosticsIncludeDefault)
	if len(diagnostics) != 2 {
		macro := firstBoundNode[*MacroExpandOperator](t, code.Query)
		t.Fatalf("macro-expand member diagnostics = %#v, semantic diagnostics = %#v, statement-list alternates = %d, want two", diagnostics, macro.SemanticDiagnostics(), len(macro.StatementList().Alternates()))
	}
	wantStart := strings.LastIndex(text, "1")
	for _, diagnostic := range diagnostics {
		if diagnostic.Code != "KS111" || diagnostic.Start != wantStart || diagnostic.Length != 1 {
			t.Fatalf("macro-expand member diagnostic = %#v, want KS111 at %d+1", diagnostic, wantStart)
		}
	}
}

func TestNativeTreeBinderMacroExpandEmptyResolvedGroup(t *testing.T) {
	empty := NewEntityGroupSymbol("Empty")
	globals := NewGlobalState(NewDatabaseSymbol("db", empty))
	code := ParseAndAnalyze(`macro-expand Empty as Scope (Scope)`, globals)
	macro := firstBoundNode[*MacroExpandOperator](t, code.Query)

	if !macro.IsBound() {
		t.Fatal("macro-expand over an empty resolved group is unbound")
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("empty-group diagnostics = %#v", diagnostics)
	}
	if columns := semanticTable(t, macro).Columns(); len(columns) != 0 {
		t.Fatalf("empty-group result columns = %#v", columns)
	}
	if alternates := macro.Alternates(); len(alternates) != 0 {
		t.Fatalf("empty-group alternates = %#v", alternates)
	}
	if alternates := macro.StatementList().Alternates(); len(alternates) != 0 {
		t.Fatalf("empty-group statement-list alternates = %#v", alternates)
	}
}

func TestNativeTreeBinderMacroExpandWithoutDeclaredScopeClonesEveryBody(t *testing.T) {
	code := ParseAndAnalyze(`macro-expand Targets (print x=1)`, macroExpandTestGlobals())
	macro := firstBoundNode[*MacroExpandOperator](t, code.Query)
	if alternates := macro.Alternates(); len(alternates) != 0 {
		t.Fatalf("macro-expand alternates = %#v, want none", alternates)
	}
	if got := len(macro.StatementList().Alternates()); got != 2 {
		t.Fatalf("statement-list alternates = %d, want one full clone per entity", got)
	}
	for _, alternate := range macro.StatementList().Alternates() {
		if alternate == nil || alternate.ChildCount() != 1 || alternate.Text(IncludeTriviaAll) != macro.StatementList().Text(IncludeTriviaAll) {
			t.Fatalf("alternate = %#v, want full statement-list clone", alternate)
		}
		if alternate.Tree() == nil || alternate.Tree().Original != code.Tree {
			t.Fatalf("alternate tree provenance = %#v", alternate.Tree())
		}
	}

	text := `macro-expand Targets (1)`
	nonTabular := ParseAndAnalyze(text, macroExpandTestGlobals())
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(nonTabular.Syntax(), DiagnosticsIncludeDefault)
	if len(diagnostics) != 2 {
		t.Fatalf("no-scope diagnostics = %#v, want two", diagnostics)
	}
	wantStart := strings.LastIndex(text, "1")
	for _, diagnostic := range diagnostics {
		if diagnostic.Code != "KS111" || diagnostic.Start != wantStart || diagnostic.Length != 1 {
			t.Fatalf("no-scope diagnostic = %#v, want KS111 at %d+1", diagnostic, wantStart)
		}
	}
}

func TestNativeTreeBinderMacroExpandSpecialMembersAndNonTabularBody(t *testing.T) {
	query := ParseQuery(`macro-expand Targets as Scope (print Database=Scope.$current_database, Cluster=Scope.$current_cluster_endpoint)`)
	NewBinder(macroExpandTestGlobals()).TryBind(query)
	macro := firstBoundNode[*MacroExpandOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, macro), "Database", "Cluster")
	assertDiagnosticCodes(t, macro)

	nonTabular := ParseQuery(`macro-expand Targets as Scope (1)`)
	NewBinder(macroExpandTestGlobals()).TryBind(nonTabular)
	assertDiagnosticCodes(t, firstBoundNode[*MacroExpandOperator](t, nonTabular), "KS111", "KS111")
}

func TestNativeTreeBinderMacroExpandFuzzyMember(t *testing.T) {
	query := ParseQuery(`macro-expand isfuzzy=true entity_group[T1, Missing] as Scope (Scope)`)
	NewBinder(macroExpandTestGlobals()).TryBind(query)
	macro := firstBoundNode[*MacroExpandOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, macro), "A")
	reference := firstNamedReference(t, macro.EntityGroup(), "Missing")
	assertDiagnosticCodes(t, reference, "KS142")
	assertDiagnosticCodes(t, simpleBinderSyntaxNode(macro.EntityGroup()), "KS251")
}

func assertMacroExpandDiagnosticCodes(t *testing.T, diagnostics []Diagnostic, want ...string) {
	t.Helper()
	if len(diagnostics) != len(want) {
		t.Fatalf("diagnostics = %#v, want codes %v", diagnostics, want)
	}
	for i, code := range want {
		if diagnostics[i].Code != code {
			t.Fatalf("diagnostic %d = %#v, want code %s", i, diagnostics[i], code)
		}
	}
}

func TestNativeTreeBinderMacroExpandParametersPlacementAndUnknownGroup(t *testing.T) {
	invalidParameter := ParseQuery(`macro-expand isfuzzy=1 Targets as Scope (Scope)`)
	NewBinder(macroExpandTestGlobals()).TryBind(invalidParameter)
	assertDiagnosticCodes(t, firstBoundNode[*MacroExpandOperator](t, invalidParameter))

	unknown := ParseQuery(`macro-expand Missing as Scope (print x=1)`)
	NewBinder(macroExpandTestGlobals()).TryBind(unknown)
	assertDiagnosticCodes(t, firstNamedReference(t, unknown, "Missing"), "KS142")
	assertDiagnosticCodes(t, firstBoundNode[*MacroExpandOperator](t, unknown))

	piped := ParseQuery(`T1 | macro-expand Targets as Scope (Scope)`)
	NewBinder(macroExpandTestGlobals()).TryBind(piped)
	assertDiagnosticCodes(t, firstBoundNode[*MacroExpandOperator](t, piped))
}
