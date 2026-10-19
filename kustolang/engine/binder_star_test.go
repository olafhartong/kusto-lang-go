package kustolang

import "testing"

func TestNativeBinderFunctionArgumentStarBindsGroupAndTuple(t *testing.T) {
	declared := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbol("T", declared).WithIsOpen(true)
	inferred := NewColumnSymbol("Count", ScalarTypes.Long.Type())
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.openTableInferredColumns = map[*TableSymbol]map[string]*ColumnSymbol{table: {"Count": inferred}}
	binder.openTableInferredOrder = map[*TableSymbol][]*ColumnSymbol{table: {inferred}}
	query := ParseQuery(`T | project Packed=pack_array(*)`)
	binder.TryBind(query)

	star := firstBoundNode[*StarExpression](t, query)
	group, ok := star.ReferencedSymbol().(*GroupSymbol)
	if !ok {
		t.Fatalf("referenced symbol = %T, want *GroupSymbol", star.ReferencedSymbol())
	}
	members := group.Members()
	if len(members) != 2 || members[0] != declared || members[1] != inferred {
		t.Fatalf("group members = %#v", members)
	}
	tuple, ok := TupleForType(ExpressionResultType(star))
	if !ok {
		t.Fatalf("result type = %#v, want tuple", ExpressionResultType(star))
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0] != declared || columns[1] != inferred {
		t.Fatalf("tuple columns = %#v", columns)
	}
}

func TestNativeBinderBinaryLeftStarBindsDynamic(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	query := ParseQuery(`T | where * has "alice"`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	star := firstBoundNode[*StarExpression](t, query)
	if star.ReferencedSymbol() != nil || ExpressionResultType(star) != ScalarTypes.Dynamic.Type() {
		t.Fatalf("star semantics = ref:%#v type:%#v", star.ReferencedSymbol(), ExpressionResultType(star))
	}
}

func TestNativeBinderStandaloneStarRemainsUnbound(t *testing.T) {
	star := MustNewStarExpressionTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindAsteriskToken, "", "*")))
	NewBinder(nil).TryBind(star)
	if star.IsBound() {
		t.Fatalf("standalone star semantic info = %#v", star.ReferencedSymbol())
	}
}

func TestNativeBinderFunctionArgumentStarWithEmptyScopeBindsEmptyTuple(t *testing.T) {
	query := ParseQuery(`print Packed=pack_array(*)`)
	NewBinder(nil).TryBind(query)
	star := firstBoundNode[*StarExpression](t, query)
	group, groupOK := star.ReferencedSymbol().(*GroupSymbol)
	tuple, tupleOK := TupleForType(ExpressionResultType(star))
	if !groupOK || !tupleOK || len(group.Members()) != 0 || len(tuple.Columns()) != 0 {
		t.Fatalf("empty-scope star = group:%#v tuple:%#v", group, tuple)
	}
}
