package kustolang

import "testing"

func serializeScalarParityGlobals() *GlobalState {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
	)
	return NewGlobalState(NewDatabaseSymbol("db", table))
}

func TestNativeSerializeGeneralBoundScalarsAndDownstreamBinding(t *testing.T) {
	query := ParseQuery("T | serialize F=strlen(A), B=N + 1, P=D.k, E=D[0], U=-N | project F, B, P, E, U")
	NewBinder(serializeScalarParityGlobals()).TryBind(query)

	serialize := firstBoundNode[*SerializeOperator](t, query)
	if !serialize.IsBound() {
		t.Fatal("serialize with supported general scalar expressions was not bound")
	}
	result := semanticTable(t, serialize)
	assertOrderedColumns(t, result, "A", "N", "D", "F", "B", "P", "E", "U")
	for _, name := range []string{"F", "B", "U"} {
		if column := result.GetColumn(name); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
			t.Fatalf("serialize %s = %#v, want long", name, column)
		}
	}
	for _, name := range []string{"P", "E"} {
		if column := result.GetColumn(name); column == nil || !IsScalar(column.Type) || column.Type == &Error.TypeSymbol {
			t.Fatalf("serialize %s = %#v, want a supported scalar", name, column)
		}
	}

	project := firstBoundNode[*ProjectOperator](t, query)
	if !project.IsBound() {
		t.Fatal("project downstream of general-scalar serialize was not bound")
	}
	assertOrderedColumns(t, semanticTable(t, project), "F", "B", "P", "E", "U")
}

func TestNativeSerializeRecoveredFunctionClosePreservesDiagnostic(t *testing.T) {
	query := ParseQuery("T | serialize F=strlen(A")
	NewBinder(serializeScalarParityGlobals()).TryBind(query)

	serialize := firstBoundNode[*SerializeOperator](t, query)
	if !serialize.IsBound() {
		t.Fatal("serialize with a recovered function close was not bound")
	}
	result := semanticTable(t, serialize)
	assertOrderedColumns(t, result, "A", "N", "D", "F")
	if column := result.GetColumn("F"); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("recovered F = %#v, want long", column)
	}

	diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS005" || diagnostics[0].Message != "Expected: )" {
		t.Fatalf("recovered diagnostics = %#v, want one KS005 Expected: )", diagnostics)
	}
}

func TestNativeSerializeGeneralScalarPreflightIsTransactional(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "serialize X=strlen(A)").(*SerializeOperator)
	binder := NewBinder(serializeScalarParityGlobals())
	binder.rowScope = binder.Globals.Database.GetTable("T")

	if info := newSimpleNodeBinder(binder).bind(operator.SyntaxNode); info != nil {
		t.Fatalf("unbound general-scalar preflight returned semantic info %#v", info)
	}
	named := operator.Expressions().At(0).Element().(*SimpleNamedExpression)
	if operator.IsBound() || named.Name().ReferencedSymbol() != nil {
		t.Fatal("rejected unbound serialize expression mutated operator or declaration semantics")
	}

	query := ParseQuery("T | serialize X=missing_function(A) | project X")
	NewBinder(serializeScalarParityGlobals()).TryBind(query)
	serialize := firstBoundNode[*SerializeOperator](t, query)
	if !serialize.IsBound() {
		t.Fatal("serialize with an error-valued scalar was not bound")
	}
	column := semanticTable(t, serialize).GetColumn("X")
	if column == nil || column.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("error-valued X = %#v, want current unknown", column)
	}
	declaration := firstBoundNode[*SimpleNamedExpression](t, query)
	if declaration.Name().ReferencedSymbol() != column {
		t.Fatal("serialize declaration did not reference committed X")
	}
}
