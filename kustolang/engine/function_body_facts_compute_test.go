package kustolang

import "testing"

func TestComputeFunctionBodyFactsDetectsUnqualifiedTableDependencies(t *testing.T) {
	tableName := NewParameter("tableName", ScalarTypes.String.Type())
	signature := NewComputedSignature("table(tableName)", TabularityTabular, tableName)

	expr := ParseExpression(`table(tableName)`)
	call := functionBodyFactsTestCall(t, expr, "table")
	resultTable := NewTableSymbol("Result", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	call.SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Table, &resultTable.TypeSymbol))
	functionBodyFactsTestName(t, expr, "tableName").
		SetSemanticInfo(NewSemanticInfoForSymbol(NewParameterSymbol("tableName", ScalarTypes.String.Type(), ""), ScalarTypes.String.Type()))

	facts := ComputeFunctionBodyFacts(signature, testFunctionBody(expr))
	if !facts.HasUnqualifiedTableCall() || facts.HasQualifiedTableCall() {
		t.Fatalf("table-call flags = unqualified:%t qualified:%t", facts.HasUnqualifiedTableCall(), facts.HasQualifiedTableCall())
	}
	if len(facts.DependentParameters()) != 1 || facts.DependentParameters()[0] != tableName {
		t.Fatalf("dependent parameters = %#v", facts.DependentParameters())
	}
	if facts.NonVariableReturnType() != nil {
		t.Fatalf("variable table call should not have non-variable return type: %#v", facts.NonVariableReturnType())
	}

	constant := ParseExpression(`table("SigninLogs")`)
	functionBodyFactsTestCall(t, constant, "table").
		SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Table, &resultTable.TypeSymbol))
	constantFacts := ComputeFunctionBodyFacts(NewComputedSignature(`table("SigninLogs")`, TabularityTabular), testFunctionBody(constant))
	if len(constantFacts.UnqualifiedTableNames()) != 1 || constantFacts.UnqualifiedTableNames()[0] != "SigninLogs" {
		t.Fatalf("unqualified table names = %#v", constantFacts.UnqualifiedTableNames())
	}
}

func TestComputeFunctionBodyFactsDistinguishesQualifiedTableCalls(t *testing.T) {
	expr := ParseExpression(`database("db").table("T")`)
	resultTable := NewTableSymbol("Result", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	functionBodyFactsTestCall(t, expr, "database").
		SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Database, &UnknownDatabaseSymbol.TypeSymbol))
	functionBodyFactsTestCall(t, expr, "table").
		SetSemanticInfo(NewSemanticInfoForSymbol(Functions.Table, &resultTable.TypeSymbol))

	facts := ComputeFunctionBodyFacts(NewComputedSignature(`database("db").table("T")`, TabularityTabular), testFunctionBody(expr))
	if !facts.HasDatabaseCall() || !facts.HasQualifiedTableCall() || facts.HasUnqualifiedTableCall() {
		t.Fatalf("qualified flags = database:%t qualified:%t unqualified:%t", facts.HasDatabaseCall(), facts.HasQualifiedTableCall(), facts.HasUnqualifiedTableCall())
	}
	if len(facts.UnqualifiedTableNames()) != 0 {
		t.Fatalf("qualified table call should not add unqualified names: %#v", facts.UnqualifiedTableNames())
	}
}

func TestComputeFunctionBodyFactsTranslatesCalledFunctionDependentParameters(t *testing.T) {
	outerSource := NewParameter("source", ScalarTypes.String.Type())
	calledSource := NewParameter("innerSource", ScalarTypes.String.Type())
	calledSignature := NewComputedSignature("table(innerSource)", TabularityTabular, calledSource)
	calledFunction := NewFunctionSymbol("Inner", calledSignature)
	outerSignature := NewComputedSignature("Inner(source)", TabularityTabular, outerSource)

	expr := ParseExpression(`Inner(source)`)
	resultTable := NewTableSymbol("Result", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	call := functionBodyFactsTestCall(t, expr, "Inner")
	call.SetSemanticInfo(
		NewSemanticInfoForSignature(calledSignature, &resultTable.TypeSymbol).
			WithReferencedSymbol(calledFunction).
			WithCalledFunctionInfo(NewFunctionCallInfo(nil, FunctionBodyFactsDefault.
				AddDependentParameter(calledSource).
				WithHasClusterCall(true))),
	)
	functionBodyFactsTestName(t, expr, "source").
		SetSemanticInfo(NewSemanticInfoForSymbol(NewParameterSymbol("source", ScalarTypes.String.Type(), ""), ScalarTypes.String.Type()))

	facts := ComputeFunctionBodyFacts(outerSignature, testFunctionBody(expr))
	if !facts.HasClusterCall() {
		t.Fatal("called function facts should be combined")
	}
	if len(facts.DependentParameters()) != 1 || facts.DependentParameters()[0] != outerSource {
		t.Fatalf("outer dependent parameters = %#v", facts.DependentParameters())
	}
}

func TestGetReferencedParametersTreatsCrossTreeExpansionVariablesAsParameters(t *testing.T) {
	tableName := NewParameter("tableName", ScalarTypes.String.Type())
	signature := NewComputedSignature("table(tableName)", TabularityTabular, tableName)
	reference := ParseExpression("tableName")
	source := ParseExpression(`"T"`)
	NewSyntaxTree(syntaxNodeForBinderLocation(reference), nil, 0)
	NewSyntaxTree(syntaxNodeForBinderLocation(source), nil, 0)
	reference.(interface{ SetSemanticInfo(SemanticInfo) }).SetSemanticInfo(
		NewSemanticInfoForReferencedSymbol(NewVariableSymbol("tableName", ScalarTypes.String.Type(), true, "T", source)),
	)

	parameters := GetReferencedParameters(signature, reference)
	if len(parameters) != 1 || parameters[0] != tableName {
		t.Fatalf("cross-tree expansion parameters = %#v", parameters)
	}
}

func TestComputeFunctionBodyFactsRecordsNonVariableReturnAndSyntaxErrors(t *testing.T) {
	expr := ParseExpression("42")
	expr.(interface{ SetSemanticInfo(SemanticInfo) }).SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()))
	facts := ComputeFunctionBodyFacts(NewComputedSignature("42", TabularityScalar), testFunctionBody(expr))
	if facts.NonVariableReturnType() != ScalarTypes.Long.Type() || facts.HasVariableReturnType() || facts.HasSyntaxErrors() {
		t.Fatalf("constant facts = type:%#v variable:%t syntax:%t", facts.NonVariableReturnType(), facts.HasVariableReturnType(), facts.HasSyntaxErrors())
	}

	malformed := ParseExpression("fn(")
	malformedFacts := ComputeFunctionBodyFacts(NewComputedSignature("fn(", TabularityScalar), testFunctionBody(malformed))
	if !malformedFacts.HasSyntaxErrors() {
		t.Fatalf("expected syntax error facts for malformed body, diagnostics=%#v", malformed.ContainedDiagnostics())
	}
}

func TestGetOrComputeFunctionBodyFactsCachesBySignatureScope(t *testing.T) {
	dbFunction := NewComputedFunctionSymbol("DbFn", "42", TabularityScalar)
	dbSignature := dbFunction.Signatures()[0]
	globals := NewGlobalState(NewDatabaseSymbol("db", dbFunction))
	global := NewGlobalBindingCache()
	local := NewLocalBindingCache()
	expr := ParseExpression("42")
	expr.(interface{ SetSemanticInfo(SemanticInfo) }).SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()))

	facts := GetOrComputeFunctionBodyFacts(globals, global, local, dbSignature, testFunctionBody(expr))
	if facts.NonVariableReturnType() != ScalarTypes.Long.Type() {
		t.Fatalf("computed facts = %#v", facts)
	}
	if _, ok := global.DatabaseFunctionBodyFacts[dbSignature]; !ok || len(local.NonDatabaseFunctionBodyFacts) != 0 {
		t.Fatalf("database facts cache = global:%d local:%d", len(global.DatabaseFunctionBodyFacts), len(local.NonDatabaseFunctionBodyFacts))
	}
}

func functionBodyFactsTestCall(t *testing.T, root SyntaxElement, name string) *FunctionCallExpression {
	t.Helper()
	var call *FunctionCallExpression
	WalkSyntax(root, func(element SyntaxElement) {
		if call != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression); ok && NameReferenceSimpleName(found.Name()) == name {
			call = found
		}
	})
	if call == nil {
		t.Fatalf("function call %q not found in %q", name, root.Text(IncludeTriviaMinimal))
	}
	return call
}

func functionBodyFactsTestName(t *testing.T, root SyntaxElement, name string) *NameReference {
	t.Helper()
	var reference *NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		if reference != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(found) == name {
			reference = found
		}
	})
	if reference == nil {
		t.Fatalf("name reference %q not found in %q", name, root.Text(IncludeTriviaMinimal))
	}
	return reference
}
