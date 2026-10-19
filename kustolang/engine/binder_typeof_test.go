package kustolang

import "testing"

func TestNativeBinderBindsTypeOfPrimitiveAsConstantTypeLiteral(t *testing.T) {
	expression := ParseExpression(`typeof(long)`)
	literal, ok := expression.(*TypeOfLiteralExpression)
	if !ok {
		t.Fatalf("expression = %T, want *TypeOfLiteralExpression", expression)
	}

	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(literal)
	info, ok := literal.GetSemanticInfo()
	if !ok {
		t.Fatal("typeof literal was not bound")
	}
	if info.ReferencedSymbol() != ScalarTypes.Long.Type() || info.ResultType != ScalarTypes.TypeLit.Type() || !info.IsConstant {
		t.Fatalf("semantic info = %#v", info)
	}
	if len(info.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", info.Diagnostics)
	}
}

func TestNativeBinderBindsTypeOfSchemaAsConstantTableLiteral(t *testing.T) {
	expression := ParseExpression(`typeof(A:string, B:long)`)
	literal := expression.(*TypeOfLiteralExpression)

	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(literal)
	info, ok := literal.GetSemanticInfo()
	if !ok {
		t.Fatal("typeof literal was not bound")
	}
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("referenced symbol = %T, want *TableSymbol", info.ReferencedSymbol())
	}
	columns := table.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "A" || columns[0].Type != ScalarTypes.String.Type() || columns[1].SymbolName() != "B" || columns[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("columns = %#v", columns)
	}
	if info.ResultType != ScalarTypes.TypeLit.Type() || !info.IsConstant || len(info.Diagnostics) != 0 {
		t.Fatalf("semantic info = %#v", info)
	}
}

func TestNativeBinderTypeOfStarUsesRowScope(t *testing.T) {
	rowScope := NewTableSymbol("T",
		NewColumnSymbol("Declared", ScalarTypes.String.Type()),
	).WithIsOpen(true)
	inferred := NewColumnSymbol("Inferred", ScalarTypes.Long.Type())
	expression := ParseExpression(`typeof(*)`)
	literal := expression.(*TypeOfLiteralExpression)

	binder := NewBinder(NewGlobalState(nil))
	binder.rowScope = rowScope
	binder.openTableInferredColumns = map[*TableSymbol]map[string]*ColumnSymbol{}
	binder.openTableInferredOrder = map[*TableSymbol][]*ColumnSymbol{}
	binder.openTableInferredColumns[rowScope] = map[string]*ColumnSymbol{"Inferred": inferred}
	binder.openTableInferredOrder[rowScope] = []*ColumnSymbol{inferred}
	binder.TryBind(literal)
	info, _ := literal.GetSemanticInfo()
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("referenced symbol = %T", info.ReferencedSymbol())
	}
	columns := table.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "Declared" || columns[1].SymbolName() != "Inferred" {
		t.Fatalf("columns = %#v", columns)
	}
	if len(info.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", info.Diagnostics)
	}
}

func TestNativeBinderTypeOfStarWithoutRowScopeReportsDiagnostic(t *testing.T) {
	expression := ParseExpression(`typeof(*)`)
	literal := expression.(*TypeOfLiteralExpression)

	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(literal)
	info, _ := literal.GetSemanticInfo()
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok || len(table.Columns()) != 0 {
		t.Fatalf("referenced symbol = %#v", info.ReferencedSymbol())
	}
	if !diagnosticsContainMessage(info.Diagnostics, "KS199", "No columns are currently in scope.") {
		t.Fatalf("diagnostics = %#v", info.Diagnostics)
	}
	if info.Diagnostics[0].Start != 7 || info.Diagnostics[0].Length != 1 {
		t.Fatalf("diagnostic location = %d+%d", info.Diagnostics[0].Start, info.Diagnostics[0].Length)
	}
}

func TestNativeBinderTypeOfMalformedDeclarationReportsKS170(t *testing.T) {
	expression := ParseExpression(`typeof(*, long)`)
	literal, ok := expression.(*TypeOfLiteralExpression)
	if !ok {
		t.Fatalf("expression = %T, want *TypeOfLiteralExpression", expression)
	}

	binder := NewBinder(NewGlobalState(nil))
	binder.rowScope = NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	binder.TryBind(literal)
	info, _ := literal.GetSemanticInfo()
	if info.ResultType != ScalarTypes.TypeLit.Type() || !info.IsConstant {
		t.Fatalf("semantic info = %#v", info)
	}
	if len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS170" || info.Diagnostics[0].Message != "The syntax is not a valid column declaration." {
		t.Fatalf("semantic diagnostics = %#v", info.Diagnostics)
	}
}

func TestNativeBinderBindsCompoundStringLiteral(t *testing.T) {
	expression := ParseExpression(`'a' 'b'`)
	compound, ok := expression.(*CompoundStringLiteralExpression)
	if !ok {
		t.Fatalf("expression = %T, want *CompoundStringLiteralExpression", expression)
	}

	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(compound)
	info, ok := compound.GetSemanticInfo()
	if !ok || info.ResultType != ScalarTypes.String.Type() || !info.IsConstant || len(info.Diagnostics) != 0 {
		t.Fatalf("semantic info = %#v, bound = %t", info, ok)
	}
}
