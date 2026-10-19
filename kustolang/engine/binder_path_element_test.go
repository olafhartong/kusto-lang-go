package kustolang

import "testing"

func TestNativePathElementDynamicAndTupleSelectors(t *testing.T) {
	known := NewColumnSymbol("known", ScalarTypes.Long.Type())
	tupleColumn := NewColumnSymbol("second", ScalarTypes.String.Type())
	binder := NewBinder(nil)
	binder.LocalScope.AddSymbol(NewVariableSymbol("bag", ScalarTypes.GetDynamicBag(known), false, nil, nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("array", ScalarTypes.GetDynamicArray(ScalarTypes.Real.Type()), false, nil, nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("tuple", NewTupleSymbol([]*ColumnSymbol{
		NewColumnSymbol("first", ScalarTypes.Long.Type()), tupleColumn,
	}), false, nil, nil))

	tests := []struct {
		text string
		typ  *TypeSymbol
	}{
		{"bag.known", ScalarTypes.DynamicLong.Type()},
		{`bag["known"]`, ScalarTypes.DynamicLong.Type()},
		{`bag["missing"]`, ScalarTypes.Dynamic.Type()},
		{"array[0]", ScalarTypes.DynamicReal.Type()},
		{"tuple[9]", ScalarTypes.Unknown.Type()},
	}
	for _, test := range tests {
		expression := ParseExpression(test.text)
		binder.TryBind(expression)
		if got := ExpressionRawResultType(expression); got != test.typ {
			t.Fatalf("%s result = %v, want %v", test.text, got, test.typ)
		}
	}

	tuple := ParseExpression("tuple[1]").(*ElementExpression)
	binder.TryBind(tuple)
	selector := tuple.Selector().(*BracketedExpression)
	if selector.ReferencedSymbol() != tupleColumn || tuple.ReferencedSymbol() != tupleColumn || ExpressionRawResultType(tuple) != ScalarTypes.String.Type() {
		t.Fatalf("tuple selector identity = selector:%v wrapper:%v type:%v", selector.ReferencedSymbol(), tuple.ReferencedSymbol(), ExpressionRawResultType(tuple))
	}
}

func TestNativePathElementDiagnosticsAndSuppression(t *testing.T) {
	binder := NewBinder(nil)
	binder.LocalScope.AddSymbol(NewVariableSymbol("array", ScalarTypes.DynamicArrayOfLong, false, nil, nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("scalar", ScalarTypes.Long.Type(), false, nil, nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("unknown", ScalarTypes.Unknown.Type(), false, nil, nil))

	tests := []struct {
		text string
		code string
	}{
		{`array["x"]`, "KS141"},
		{"scalar[0]", "KS147"},
		{"[1]", "KS141"},
		{`[("a")]`, "KS131"},
	}
	for _, test := range tests {
		expression := ParseExpression(test.text)
		binder.TryBind(expression)
		diagnostics := GetContainedDiagnosticsPreserveDuplicates(expression, DiagnosticsIncludeSemantic)
		if len(diagnostics) != 1 || diagnostics[0].Code != test.code {
			t.Fatalf("%s diagnostics = %#v, want only %s", test.text, diagnostics, test.code)
		}
	}

	for _, text := range []string{"unknown[false]", "Missing[false]"} {
		expression := ParseExpression(text)
		binder.TryBind(expression)
		for _, diagnostic := range GetContainedDiagnosticsPreserveDuplicates(expression, DiagnosticsIncludeSemantic) {
			if diagnostic.Code == "KS141" || diagnostic.Code == "KS147" {
				t.Fatalf("%s produced cascading selector diagnostic %s", text, diagnostic.Code)
			}
		}
	}
}

func TestNativePathScopeIsolationRestorationAndMalformedRetry(t *testing.T) {
	binder := NewBinder(nil)
	binder.LocalScope.AddSymbol(NewVariableSymbol("bag", ScalarTypes.GetDynamicBag(NewColumnSymbol("member", ScalarTypes.Long.Type())), false, nil, nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("member", ScalarTypes.String.Type(), true, "member", nil))

	bracketed := ParseExpression("bag.[member]").(*PathExpression)
	binder.TryBind(bracketed)
	index := bracketed.Selector().(*BracketedExpression).Expression().(*NameReference)
	if index.ReferencedSymbol() == nil || index.ReferencedSymbol().SymbolName() != "member" || ExpressionRawResultType(bracketed) != ScalarTypes.DynamicLong.Type() {
		t.Fatalf("bracket selector did not use ordinary scope: symbol=%v type=%v", index.ReferencedSymbol(), ExpressionRawResultType(bracketed))
	}
	if binder.pathScope != nil {
		t.Fatalf("path scope leaked after bracket selector: %v", binder.pathScope)
	}

	malformed := ParseExpression("bag.member").(*PathExpression)
	saved := malformed.SyntaxNode.children[2]
	malformed.SyntaxNode.children[2] = nil
	binder.TryBind(malformed)
	if malformed.IsBound() || malformed.Expression().(*NameReference).IsBound() || binder.pathScope != nil {
		t.Fatal("malformed path committed subtree semantics or leaked path scope")
	}
	malformed.SyntaxNode.children[2] = saved
	binder.TryBind(malformed)
	firstType := ExpressionRawResultType(malformed)
	binder.TryBind(malformed)
	if !malformed.IsBound() || firstType != ScalarTypes.DynamicLong.Type() || ExpressionRawResultType(malformed) != firstType || binder.pathScope != nil {
		t.Fatal("path retry or idempotent rebind failed")
	}
}
