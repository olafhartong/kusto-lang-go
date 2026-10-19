package kustolang

import "testing"

func TestScopeKindValuesMatchUpstream(t *testing.T) {
	tests := []struct {
		kind ScopeKind
		want int
		text string
	}{
		{ScopeKindNormal, 0, "Normal"},
		{ScopeKindAggregate, 1, "Aggregate"},
		{ScopeKindPlugIn, 2, "PlugIn"},
		{ScopeKindOption, 3, "Option"},
	}
	for _, tt := range tests {
		if int(tt.kind) != tt.want || tt.kind.String() != tt.text {
			t.Fatalf("scope kind %s = %d/%q, want %d/%q", tt.text, int(tt.kind), tt.kind.String(), tt.want, tt.text)
		}
	}
}

func TestGlobalStateAggregateLookupAndScope(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	if globals.GetAggregate("sum") != Aggregates.Sum {
		t.Fatalf("sum aggregate lookup = %#v", globals.GetAggregate("sum"))
	}
	if !globals.IsAggregateFunction(Aggregates.Sum) {
		t.Fatalf("sum should be an aggregate")
	}
	if globals.IsAggregateFunction(Functions.Strlen) {
		t.Fatalf("strlen should not be an aggregate")
	}
	localSum := NewScalarFunctionSymbol("sum", ScalarTypes.String.Type())
	if globals.IsAggregateFunction(localSum) {
		t.Fatalf("same-name local function should not be the built-in aggregate")
	}
	if got := ArgumentScopeForFunction(globals, Aggregates.Sum, ScopeKindAggregate); got != ScopeKindNormal {
		t.Fatalf("aggregate function argument scope = %s", got)
	}
	if got := ArgumentScopeForFunction(globals, Functions.Strlen, ScopeKindAggregate); got != ScopeKindAggregate {
		t.Fatalf("non-aggregate nested in aggregate scope = %s", got)
	}
	if got := ArgumentScopeForFunction(globals, Functions.Strlen, ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("normal function argument scope = %s", got)
	}
}

func TestArgumentScopeForFunctionArgumentUsesAggregateParameters(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	value := NewParameterKind("value", ParameterTypeScalar)
	aggregate := NewParameterKind("aggregate", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate))
	signature := NewSignature(ScalarTypes.Long.Type(), value, aggregate)
	fn := NewFunctionSymbol("has_aggregate_arg", signature)

	if !FunctionHasAggregateParameters(fn) {
		t.Fatalf("function should report aggregate parameters")
	}
	if got := ArgumentScopeForFunctionArgument(globals, fn, signature, 0, 2, ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("argument 0 scope = %s", got)
	}
	if got := ArgumentScopeForFunctionArgument(globals, fn, signature, 1, 2, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("argument 1 scope = %s", got)
	}
	if got := ArgumentScopeForFunctionArgument(globals, fn, nil, 1, 2, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("argument 1 scope across signatures = %s", got)
	}
	if got := ArgumentScopeForFunctionArgument(globals, Aggregates.Sum, Aggregates.Sum.Signatures()[0], 0, 1, ScopeKindAggregate); got != ScopeKindNormal {
		t.Fatalf("aggregate catalog function argument scope = %s", got)
	}
	if got := ArgumentScopeForFunctionArgument(globals, Functions.Strlen, Functions.Strlen.Signatures()[0], 0, 1, ScopeKindAggregate); got != ScopeKindAggregate {
		t.Fatalf("non-aggregate under aggregate outer scope = %s", got)
	}
}

func TestExpressionAwareArgumentScopeUsesCustomAndNamedLayouts(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	normal := NewParameterKind("normal", ParameterTypeScalar)
	aggregate := NewParameterKind("aggregate", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate), WithParameterOccurrence(0, 1))
	custom := NewSignature(ScalarTypes.Long.Type(), normal, aggregate).
		WithCustomLayout(func(_ *Signature, arguments []Expression, _ []Symbol) []*Parameter {
			result := make([]*Parameter, len(arguments))
			for i := range result {
				result[i] = aggregate
			}
			return result
		})
	fixed := NewSignature(ScalarTypes.Long.Type(), normal)
	function := NewFunctionSymbol("custom_scope", fixed, custom)
	argument := FakeExpressionCreate(ScalarTypes.Long.Type())

	if got := ArgumentScopeForFunctionArgument(globals, function, nil, 0, 1, ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("compatibility positional scope = %s, want Normal", got)
	}
	if got := ArgumentScopeForFunctionArgumentExpressions(globals, function, nil, []Expression{argument}, 0, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("custom expression-aware scope = %s, want Aggregate", got)
	}

	first := NewParameterKind("first", ParameterTypeScalar)
	second := NewParameterKind("second", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate))
	block := NewSignature(ScalarTypes.Long.Type(), first, second).WithLayout(ParameterLayoutBlockRepeating)
	namedFunction := NewFunctionSymbol("named_scope", block)
	named := functionCallArgument(t, `named_scope(second=sum(X))`, 0)
	if got := ArgumentScopeForFunctionArgumentExpressions(globals, namedFunction, block, []Expression{named}, 0, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("named block scope = %s, want Aggregate", got)
	}
}

func TestExpressionAwareArgumentScopeUsesNextParameterForMalformedTail(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	value := NewParameterKind("value", ParameterTypeScalar)
	aggregate := NewParameterKind("aggregate", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate))
	signature := NewSignature(ScalarTypes.Long.Type(), value, aggregate)
	function := NewFunctionSymbol("missing_scope", signature)
	arguments := []Expression{
		FakeExpressionCreate(ScalarTypes.Long.Type()),
		FakeExpressionCreate(ScalarTypes.Unknown.Type()),
	}

	if got := ArgumentScopeForFunctionArgumentExpressions(globals, function, signature, arguments, 1, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("missing-tail scope = %s, want Aggregate", got)
	}
}

func TestPositionAwareArgumentScopeDistinguishesCurrentAndNextParameter(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	value := NewParameterKind("value", ParameterTypeScalar)
	aggregate := NewParameterKind("aggregate", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate), WithParameterOccurrence(0, 1))
	signature := NewSignature(ScalarTypes.Long.Type(), value, aggregate)
	function := NewFunctionSymbol("position_scope", signature)
	call, ok := ParseExpression(`position_scope(x )`).(*FunctionCallExpression)
	if !ok {
		t.Fatal("position fixture did not parse as a function call")
	}
	argument := call.ArgumentList().Expressions().At(0).Element()

	if got := ArgumentScopeForFunctionCallPosition(globals, function, signature, call, argument.TextStart(), ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("inside argument scope = %s, want Normal", got)
	}
	if got := ArgumentScopeForFunctionCallPosition(globals, function, signature, call, argument.End(), ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("at argument end scope = %s, want Normal", got)
	}
	if got := ArgumentScopeForFunctionCallPosition(globals, function, signature, call, argument.End()+1, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("after argument scope = %s, want Aggregate", got)
	}

	unknownNamed, ok := ParseExpression(`position_scope(unknown=1)`).(*FunctionCallExpression)
	if !ok {
		t.Fatal("named fixture did not parse as a function call")
	}
	named := unknownNamed.ArgumentList().Expressions().At(0).Element()
	if got := ArgumentScopeForFunctionCallPosition(globals, function, signature, unknownNamed, named.TextStart(), ScopeKindNormal); got != ScopeKindNormal {
		t.Fatalf("unknown named scope = %s, want Normal", got)
	}
}

func TestBinderMultiSignatureCustomLayoutKeepsBroadScope(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	normal := NewParameterKind("normal", ParameterTypeScalar)
	aggregate := NewParameterKind("aggregate", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate), WithParameterOccurrence(0, 1))
	fixed := NewSignature(ScalarTypes.Long.Type(), normal)
	custom := NewSignature(ScalarTypes.Long.Type(), normal, aggregate).
		WithCustomLayout(func(_ *Signature, arguments []Expression, _ []Symbol) []*Parameter {
			result := make([]*Parameter, len(arguments))
			for i := range result {
				result[i] = aggregate
			}
			return result
		})
	function := NewFunctionSymbol("custom_scope", fixed, custom)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, function))

	code := ParseAndAnalyze(`T | project Value=custom_scope(sum(X))`, globals)
	found := false
	for _, diagnostic := range code.GetDiagnostics() {
		if diagnostic.Code == "KS144" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unresolved overload accepted signature-specific aggregate scope: %#v", code.GetDiagnostics())
	}
}

func TestSignatureArgumentParameterAtLayouts(t *testing.T) {
	first := NewParameterKind("first", ParameterTypeScalar)
	repeating := NewParameterKind("repeating", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))
	signature := NewSignature(ScalarTypes.Long.Type(), first, repeating)
	if got := signature.ArgumentParameterAt(0, 3); got != first {
		t.Fatalf("fixed first parameter = %#v", got)
	}
	if got := signature.ArgumentParameterAt(2, 3); got != repeating {
		t.Fatalf("repeating parameter = %#v", got)
	}

	prefix := NewParameterKind("prefix", ParameterTypeScalar)
	repeatingA := NewParameterKind("a", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))
	repeatingB := NewParameterKind("b", ParameterTypeScalar, WithArgumentKind(ArgumentKindAggregate), WithParameterOccurrence(1, MaxRepeat))
	suffix := NewParameterKind("suffix", ParameterTypeScalar)
	block := NewSignature(ScalarTypes.Long.Type(), prefix, repeatingA, repeatingB, suffix).WithLayout(ParameterLayoutBlockRepeating)
	if got := block.ArgumentParameterAt(0, 6); got != prefix {
		t.Fatalf("block prefix = %#v", got)
	}
	if got := block.ArgumentParameterAt(1, 6); got != repeatingA {
		t.Fatalf("block repeating a = %#v", got)
	}
	if got := block.ArgumentParameterAt(2, 6); got != repeatingB {
		t.Fatalf("block repeating b = %#v", got)
	}
	if got := block.ArgumentParameterAt(3, 6); got != repeatingA {
		t.Fatalf("block repeating a second group = %#v", got)
	}
	if got := block.ArgumentParameterAt(5, 6); got != suffix {
		t.Fatalf("block suffix = %#v", got)
	}
	if got := ArgumentScopeForFunctionArgument(NewGlobalState(NewDatabaseSymbol("db")), NewFunctionSymbol("block", block), block, 2, 6, ScopeKindNormal); got != ScopeKindAggregate {
		t.Fatalf("block aggregate argument scope = %s", got)
	}
}
