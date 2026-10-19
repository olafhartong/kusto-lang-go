package kustolang

import "testing"

func TestNativeInvokeImplicitTableArgumentAndExactCustomResult(t *testing.T) {
	input := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	output := NewTableSymbol("Output", NewColumnSymbol("B", ScalarTypes.String.Type()))
	var context CustomReturnTypeContext
	transform := NewFunctionSymbol("Transform", NewCustomSignature(func(c CustomReturnTypeContext) Symbol {
		context = c
		return output
	}, TabularityTabular,
		NewParameterKind("source", ParameterTypeTabular),
		NewParameter("limit", ScalarTypes.Long.Type()),
	))
	globals := NewGlobalState(NewDatabaseSymbol("db", input, transform))
	query := ParseQuery("T | invoke Transform(10)")
	binder := NewBinder(globals)
	binder.TryBind(query)

	invoke := firstBoundNode[*InvokeOperator](t, query)
	call, _ := invokeFunctionCall(invoke.Function())
	if call.ReferencedSignature() == nil || call.ReferencedSignature().Symbol != transform {
		t.Fatalf("invoke signature = %#v, want Transform", call.ReferencedSignature())
	}
	if len(context.Arguments) != 2 || context.ArgumentTypes[0] != &input.TypeSymbol || context.ArgumentTypes[1] != ScalarTypes.Long.Type() {
		t.Fatalf("custom return arguments/types = %#v / %#v", context.Arguments, context.ArgumentTypes)
	}
	info, _ := invoke.GetSemanticInfo()
	if info.ReferencedSymbol() != output || info.ResultType != &output.TypeSymbol {
		t.Fatalf("invoke result = %#v / %#v, want exact output table", info.ReferencedSymbol(), info.ResultType)
	}
	if table := expressionTableResult(firstBoundNode[*PipeExpression](t, query)); table != output {
		t.Fatalf("pipe result = %#v, want exact output table", table)
	}
}

func TestNativeInvokeQualifiedDatabaseMemberBindsTerminalCallOnly(t *testing.T) {
	input := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	output := NewTableSymbol("Output", NewColumnSymbol("B", ScalarTypes.String.Type()))
	var context CustomReturnTypeContext
	transform := NewFunctionSymbol("Transform", NewCustomSignature(func(c CustomReturnTypeContext) Symbol {
		context = c
		return output
	}, TabularityTabular, NewParameterKind("source", ParameterTypeTabular)))
	database := NewDatabaseSymbol("db", input, transform)
	getDatabase := NewFunctionSymbol("GetDatabase", NewSymbolSignature(database))
	globals := NewGlobalState(database).WithFunctions([]*FunctionSymbol{getDatabase})
	query := ParseQuery("T | invoke GetDatabase().Transform()")
	NewBinder(globals).TryBind(query)

	invoke := firstBoundNode[*InvokeOperator](t, query)
	path, ok := invoke.Function().(*PathExpression)
	if !ok || !path.IsBound() {
		t.Fatalf("qualified invoke function = %T, bound=%v", invoke.Function(), expressionIsBound(invoke.Function()))
	}
	receiver := path.Expression().(*FunctionCallExpression)
	if diagnosticsContain(receiver.SemanticDiagnostics(), "KS119") || diagnosticsContain(receiver.SemanticDiagnostics(), "KS123") {
		t.Fatalf("implicit argument leaked into database() receiver: %#v", receiver.SemanticDiagnostics())
	}
	selector := path.Selector().(*FunctionCallExpression)
	if selector.ReferencedSignature() == nil || selector.ReferencedSignature().Symbol != transform {
		t.Fatalf("qualified selector signature = %#v, symbol=%#v diagnostics=%#v, want Transform", selector.ReferencedSignature(), selector.ReferencedSymbol(), selector.SemanticDiagnostics())
	}
	if len(context.Arguments) != 1 || context.ArgumentTypes[0] != &input.TypeSymbol {
		t.Fatalf("custom return arguments/types = %#v / %#v", context.Arguments, context.ArgumentTypes)
	}
	info, _ := invoke.GetSemanticInfo()
	if info.ReferencedSymbol() != output || info.ResultType != &output.TypeSymbol {
		t.Fatalf("invoke result = %#v / %#v, want exact output table", info.ReferencedSymbol(), info.ResultType)
	}
}

func TestNativeInvokeExplicitArgumentCannotSeeIncomingRowScope(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.Long.Type())
	input := NewTableSymbol("T", column)
	transform := NewFunctionSymbol("Transform", NewSymbolSignature(input,
		NewParameterKind("source", ParameterTypeTabular),
		NewParameter("value", ScalarTypes.Long.Type()),
	))
	query := ParseQuery("T | invoke Transform(A)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", input, transform))).TryBind(query)

	invoke := firstBoundNode[*InvokeOperator](t, query)
	call, _ := invokeFunctionCall(invoke.Function())
	argument := call.ArgumentList().Expressions().Elements()[0].Element()
	if expressionReferencedSymbol(argument) == column {
		t.Fatalf("explicit argument resolved against incoming row scope")
	}
	diagnostics := argument.(interface{ SemanticDiagnostics() []Diagnostic }).SemanticDiagnostics()
	if len(diagnostics) == 0 {
		t.Fatalf("explicit argument diagnostics = %#v, want unresolved-name diagnostic", diagnostics)
	}
}

func TestNativeInvokeMissingArgumentListAndRestoresState(t *testing.T) {
	input := NewTableSymbol("T")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", input)))
	prior := &implicitFunctionArgument{typ: ScalarTypes.String.Type()}
	binder.implicitFunctionArgument = prior
	binder.scopeKind = ScopeKindAggregate
	query := ParseQuery("T | invoke T")
	binder.TryBind(query)

	invoke := firstBoundNode[*InvokeOperator](t, query)
	if diagnosticsContain(invoke.SemanticDiagnostics(), "KS105") {
		t.Fatalf("invoke diagnostics = %#v, do not want Go-only KS105", invoke.SemanticDiagnostics())
	}
	if binder.implicitFunctionArgument != prior || binder.scopeKind != ScopeKindAggregate || binder.rowScope != nil || binder.pathScope != nil {
		t.Fatalf("binder state not restored: implicit=%#v scope=%v row=%#v path=%#v", binder.implicitFunctionArgument, binder.scopeKind, binder.rowScope, binder.pathScope)
	}
}

func TestNativeInvokeDoesNotRepublishFunctionCallDiagnostics(t *testing.T) {
	const query = `let Transform=(T:(Key:long,Value:string), threshold:long) { T | where Key > threshold | project Value }; datatable(Key:long,Valuestring)[1,'a'] | invoke Transform(0)`
	code := ParseAndAnalyze(query)
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS005", 135, 1},
		diagnosticLocation{"KS006", 135, 0},
		diagnosticLocation{"KS249", 105, 38},
	)
	result := terminalResultTable(code)
	assertOrderedColumns(t, result, "Value")
	if result.Columns()[0].Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("recovered Value type = %v, want unknown", result.Columns()[0].Type)
	}
}

func TestNativeStandaloneInvokeUsesEmptyImplicitTableAndOwnsOnlyOperatorDiagnostic(t *testing.T) {
	const declaration = `let Transform=(T:(Key:long,Value:string), threshold:long) { T | where Key > threshold | project Value }; `

	standalone := ParseAndAnalyze(declaration + `datatable(Key:long,Value:string)[1,'a']  invoke Transform(0)`)
	assertDiagnosticLocations(t, standalone.GetDiagnostics(),
		diagnosticLocation{"KS005", 144, 0},
		diagnosticLocation{"KS174", 146, 6},
		diagnosticLocation{"KS249", 153, 9},
	)
	standaloneResult := terminalResultTable(standalone)
	assertOrderedColumns(t, standaloneResult, "Value")
	if standaloneResult.Columns()[0].Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("standalone Value type = %v, want unknown", standaloneResult.Columns()[0].Type)
	}

	piped := ParseAndAnalyze(declaration + `datatable(Key:long,Value:string)[1,'a'] | invoke Transform(0)`)
	if diagnostics := piped.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("piped invoke diagnostics = %#v, want none", diagnostics)
	}
	pipedResult := terminalResultTable(piped)
	assertOrderedColumns(t, pipedResult, "Value")
	if pipedResult.Columns()[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("piped Value type = %v, want string", pipedResult.Columns()[0].Type)
	}
}
