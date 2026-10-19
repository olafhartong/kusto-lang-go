package kustolang

import "testing"

func TestNativeTreeBinderPrintGeneralScalarParityAndDownstream(t *testing.T) {
	query := ParseQuery(`print Arithmetic=1 + 2 * 3, Comparison=1 < 2, Logical=true and false, Unary=-1, Element=dynamic({"values":[10, 20]}).values[1], Nested=tolong(strcat("4", "2")) | project Arithmetic, Comparison, Logical, Unary, Element, Nested`)
	NewBinder(NewGlobalState(nil)).TryBind(query)

	print := firstBoundNode[*PrintOperator](t, query)
	if !print.IsBound() {
		t.Fatal("print with general scalar expressions was not bound")
	}
	assertDiagnosticCodes(t, print)
	result := semanticTable(t, print)
	assertOrderedColumns(t, result, "Arithmetic", "Comparison", "Logical", "Unary", "Element", "Nested")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Arithmetic": ScalarTypes.Long.Type(),
		"Comparison": ScalarTypes.Bool.Type(),
		"Logical":    ScalarTypes.Bool.Type(),
		"Unary":      ScalarTypes.Long.Type(),
		"Element":    result.GetColumn("Element").Type,
		"Nested":     ScalarTypes.Long.Type(),
	})
	if element := result.GetColumn("Element"); element == nil || !IsScalar(element.Type) {
		t.Fatalf("element access column = %#v, want scalar", element)
	}
	for _, node := range []SyntaxElement{
		firstBoundNode[*BinaryExpression](t, print),
		firstBoundNode[*PrefixUnaryExpression](t, print),
		firstBoundNode[*DynamicExpression](t, print),
		firstBoundNode[*ElementExpression](t, print),
	} {
		if !simpleBinderSyntaxNode(node).IsBound() {
			t.Fatalf("general scalar expression %T was not bound", node)
		}
	}
	for _, call := range boundNodes[*FunctionCallExpression](print) {
		if !call.IsBound() {
			t.Fatalf("nested function call %q was not bound", NameReferenceSimpleName(call.Name()))
		}
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Arithmetic", "Comparison", "Logical", "Unary", "Element", "Nested")
}

func TestNativeTreeBinderPrintNamedCompoundFunctionParity(t *testing.T) {
	pair := NewFunctionSymbol("pair", NewSymbolSignature(NewTupleSymbolOf(
		NewColumnSymbol("First", ScalarTypes.Long.Type()),
		NewColumnSymbol("Second", ScalarTypes.String.Type()),
	)))
	query := ParseQuery(`print (Left, Right)=pair(), Sum=tolong(tostring(40)) + 2 | project Left, Right, Sum`)
	NewBinder(NewGlobalState(nil).WithFunctions([]*FunctionSymbol{pair})).TryBind(query)

	print := firstBoundNode[*PrintOperator](t, query)
	assertDiagnosticCodes(t, print)
	result := semanticTable(t, print)
	assertOrderedColumns(t, result, "Left", "Right", "Sum")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Left":  ScalarTypes.Long.Type(),
		"Right": ScalarTypes.String.Type(),
		"Sum":   ScalarTypes.Long.Type(),
	})
	compound := print.Expressions().At(0).Element().(*CompoundNamedExpression)
	if compound.Names().Names().At(0).Element().ReferencedSymbol() != result.GetColumn("Left") ||
		compound.Names().Names().At(1).Element().ReferencedSymbol() != result.GetColumn("Right") {
		t.Fatal("compound print declarations were not bound to projected columns")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Left", "Right", "Sum")
}

func TestNativeTreeBinderPrintGeneralScalarDiagnosticsAndTabularValidation(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery(`print Error=1 + Missing, Tabular=T, Nested=tolong(strcat("x", Missing))`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	print := firstBoundNode[*PrintOperator](t, query)
	if !print.IsBound() {
		t.Fatal("print containing scalar errors and a tabular value was not bound")
	}
	for _, reference := range boundNodes[*NameReference](query) {
		if NameReferenceSimpleName(reference) == "Missing" {
			assertSemanticDiagnostic(t, reference, "KS142")
		}
	}
	assertDiagnosticCodes(t, print, "KS108")
	result := semanticTable(t, print)
	assertOrderedColumns(t, result, "Error", "Tabular", "Nested")
	if result.GetColumn("Tabular").Type.SymbolTabularity() != TabularityTabular {
		t.Fatalf("tabular print result type = %v, want tabular", result.GetColumn("Tabular").Type)
	}
	if binary := firstBoundNode[*BinaryExpression](t, print); !expressionIsBound(binary.Left()) || !expressionIsBound(binary.Right()) {
		t.Fatal("binary operands were not bound before the parent expression")
	}
}

func TestNativeTreeBinderPrintMalformedRecoveryAndTransactionalRollback(t *testing.T) {
	recovered := ParseQuery(`print Value=`)
	NewBinder(NewGlobalState(nil)).TryBind(recovered)
	recoveredPrint := firstBoundNode[*PrintOperator](t, recovered)
	if !recoveredPrint.IsBound() || !sameType(semanticTable(t, recoveredPrint).GetColumn("Value").Type, ScalarTypes.Unknown.Type()) {
		t.Fatal("top-level missing print value did not retain parser recovery semantics")
	}

	recoveredDynamic := ParseQuery(`print First=dynamic({"ok":[1, 2]}), Later=dynamic({"broken":[1, 2]}`)
	NewBinder(NewGlobalState(nil)).TryBind(recoveredDynamic)
	if dynamicPrint := firstBoundNode[*PrintOperator](t, recoveredDynamic); !dynamicPrint.IsBound() {
		t.Fatal("recoverable dynamic print expression was not bound")
	}
	if len(recoveredDynamic.ContainedDiagnostics()) == 0 {
		t.Fatal("recoverable dynamic print expression has no parser diagnostic")
	}

	recoveredCall := ParseQuery(`print First=1 + 2, Later=tolong(3`)
	NewBinder(NewGlobalState(nil)).TryBind(recoveredCall)
	if print := firstBoundNode[*PrintOperator](t, recoveredCall); !print.IsBound() ||
		!diagnosticsContainMessage(print.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("recovered function call = bound:%t diagnostics:%#v", print.IsBound(), print.ContainedDiagnostics())
	}

	query := ParseQuery(`print First=1 + 2, Later=tolong(3)`)
	print := firstBoundNode[*PrintOperator](t, query)
	call := firstBoundNode[*FunctionCallExpression](t, print)
	call.ArgumentList().OpenParen().text = ""
	NewBinder(NewGlobalState(nil)).TryBind(query)
	if print.IsBound() {
		t.Fatal("injected invalid print operator was bound")
	}
	for _, expression := range print.Expressions().Elements() {
		WalkSyntax(expression.Element(), func(element SyntaxElement) {
			if node := simpleBinderSyntaxNode(element); node != nil && node.IsBound() {
				t.Fatalf("failed print committed semantic info to %T", WrapGeneratedSyntaxElement(element))
			}
		})
	}
}
