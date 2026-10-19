package kustolang

import "testing"

func TestNativeLetBindsScalarAndTabularValuesInStatementOrder(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	query := ParseQuery(`let Threshold=10; let Rows=T | where Value > Threshold; Rows | project Value`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	if !binder.TryBind(query) {
		t.Fatal("let query did not bind")
	}

	threshold := localVariable(t, binder.LocalScope, "Threshold")
	if threshold.Type != ScalarTypes.Long.Type() || !threshold.IsConstant || ExpressionConstantValue(threshold.Source) != int64(10) {
		t.Fatalf("Threshold = type:%#v constant:%t value:%#v", threshold.Type, threshold.IsConstant, ExpressionConstantValue(threshold.Source))
	}
	rows := localVariable(t, binder.LocalScope, "Rows")
	rowsTable, ok := rows.Type.(*TableSymbol)
	if !ok || rowsTable.GetColumn("Value") == nil {
		t.Fatalf("Rows type = %#v, want table with Value", rows.Type)
	}
	for _, reference := range namedReferences(query, "Threshold") {
		if reference.ReferencedSymbol() != threshold {
			t.Fatalf("Threshold reference resolved to %#v", reference.ReferencedSymbol())
		}
	}
	if reference := firstNamedReference(t, query, "Rows"); reference.ReferencedSymbol() != rows {
		t.Fatalf("Rows reference resolved to %#v", reference.ReferencedSymbol())
	}
}

func TestNativeLetFunctionBindsParametersAndCapturedScope(t *testing.T) {
	query := ParseQuery(`let Offset=2; let Add=(x:long) { x + Offset }; print Result=Add(3)`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	var function *FunctionSymbol
	var symbols []Symbol
	binder.LocalScope.GetSymbols("Add", SymbolMatchFunction, &symbols)
	if len(symbols) == 1 {
		function, _ = symbols[0].(*FunctionSymbol)
	}
	if function == nil || len(function.Signatures()) != 1 || function.Signatures()[0].Declaration == nil {
		t.Fatalf("Add local function = %#v", symbols)
	}
	call := namedFunctionCall(t, query, "Add")
	if call.ReferencedSymbol() != function || call.GetCalledFunctionBody() == nil {
		t.Fatalf("Add call = symbol:%#v name-symbol:%#v expansion:%#v diagnostics:%#v", call.ReferencedSymbol(), call.Name().ReferencedSymbol(), call.GetCalledFunctionBody(), GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion))
	}
	if got := ExpressionResultType(call); got != ScalarTypes.Long.Type() {
		t.Fatalf("Add call result = %#v, want long", got)
	}

	parameterReference := firstNamedReference(t, function.Signatures()[0].Declaration, "x")
	if _, ok := parameterReference.ReferencedSymbol().(*ParameterSymbol); !ok {
		t.Fatalf("function-body x = %#v, want ParameterSymbol", parameterReference.ReferencedSymbol())
	}
	if binder.LocalScope.ContainsSymbol("x") {
		t.Fatal("function parameter leaked into the outer query scope")
	}
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion); len(diagnostics) != 0 {
		t.Fatalf("function query diagnostics = %#v", diagnostics)
	}
}

func TestNativeLetFunctionCallPublishesReturnedValue(t *testing.T) {
	query := ParseQuery(`let Length=strlen("abc"); print Result=Length`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)
	length := localVariable(t, binder.LocalScope, "Length")
	if length.Type != ScalarTypes.Long.Type() {
		t.Fatalf("Length type = %#v, want long", length.Type)
	}
	if _, ok := any(length.Type).(*FunctionSymbol); ok {
		t.Fatal("function-call result was published as a function")
	}
}

func TestNativeLetFunctionUsesOptionalDefaultValues(t *testing.T) {
	query := ParseQuery(`let Pick=(value:long=1) { value }; print Result=Pick()`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	call := namedFunctionCall(t, query, "Pick")
	if result := ExpressionResultType(call); result != ScalarTypes.Long.Type() {
		t.Fatalf("optional default function result = %#v", result)
	}
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion); len(diagnostics) != 0 {
		t.Fatalf("optional default diagnostics = %#v", diagnostics)
	}
}

func TestNativeLetFunctionUsesDeclaredScalarAndActualTabularParameterTypes(t *testing.T) {
	query := ParseQuery(`let Scalar=(x:long) { x }; let Tabular=(T:(A:long)) { T | project A }; print X=Scalar(int(1)); Tabular(datatable(A:long, B:string)[1, "b"])`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion); len(diagnostics) != 0 {
		t.Fatalf("typed local function diagnostics = %#v", diagnostics)
	}
	if got := ExpressionResultType(namedFunctionCall(t, query, "Scalar")); got != ScalarTypes.Long.Type() {
		t.Fatalf("Scalar result = %#v, want declared long", got)
	}
	tabular := namedFunctionCall(t, query, "Tabular")
	result := expressionTableResult(tabular)
	if result == nil || result.GetColumn("A") == nil || result.GetColumn("B") != nil {
		t.Fatalf("Tabular result = %#v, want projected A only", result)
	}
}

func TestNativeLetFunctionCapturesOnlyEarlierDeclarations(t *testing.T) {
	query := ParseQuery(`let X=1; let F=() { X }; let X="later"; print Result=F()`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	call := namedFunctionCall(t, query, "F")
	if got := ExpressionResultType(call); got != ScalarTypes.Long.Type() {
		t.Fatalf("captured F result = %#v, want long", got)
	}
	lets := boundNodes[*LetStatement](query)
	if len(lets) != 3 || !diagnosticsContain(lets[2].SemanticDiagnostics(), "KS201") {
		t.Fatalf("captured-scope duplicate diagnostics = %#v", GetContainedDiagnostics(query, DiagnosticsIncludeSemantic))
	}
}

func TestNativeLetTabularFunctionCanBeUsedAsImplicitSource(t *testing.T) {
	for _, declaration := range []string{
		`let Rows=() { datatable(A:long)[1] }; Rows | project A`,
		`let Rows=view() { datatable(A:long)[1] }; Rows | project A`,
	} {
		query := ParseQuery(declaration)
		NewBinder(NewGlobalState(nil)).TryBind(query)
		if diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion); len(diagnostics) != 0 {
			t.Fatalf("implicit source diagnostics for %q = %#v", declaration, diagnostics)
		}
		project := firstBoundNode[*ProjectOperator](t, query)
		if table := semanticTable(t, project); table.GetColumn("A") == nil {
			t.Fatalf("implicit source project = %#v", table)
		}
	}
}

func TestNativeLetReportsDuplicateAndFunctionParameterOrder(t *testing.T) {
	duplicate := ParseQuery(`let x=1; let x=2; print x`)
	NewBinder(NewGlobalState(nil)).TryBind(duplicate)
	lets := boundNodes[*LetStatement](duplicate)
	if len(lets) != 2 || !diagnosticsContain(lets[1].SemanticDiagnostics(), "KS201") {
		t.Fatalf("duplicate let diagnostics = %#v", GetContainedDiagnostics(duplicate, DiagnosticsIncludeSemantic))
	}

	badOrder := ParseQuery(`let F=(x:long, T:(*)) { T | take x }; print y=1`)
	NewBinder(NewGlobalState(nil)).TryBind(badOrder)
	declaration := firstBoundNode[*FunctionDeclaration](t, badOrder)
	if !diagnosticsContain(declaration.SemanticDiagnostics(), "KS226") {
		t.Fatalf("function parameter diagnostics = %#v", declaration.SemanticDiagnostics())
	}
}

func TestNativeLetMalformedFunctionStillProducesDiagnostics(t *testing.T) {
	query := ParseQuery(`let Broken=(x:long { x + }; print Broken(1)`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(query, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion)
	if len(diagnostics) == 0 {
		t.Fatal("malformed let function produced no diagnostics")
	}
}

func namedReferences(root SyntaxElement, name string) []*NameReference {
	var references []*NameReference
	walkSyntaxWithTriviaStartAndAncestors(root, 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(reference) == name {
			references = append(references, reference)
		}
	})
	return references
}
