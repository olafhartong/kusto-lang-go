package kustolang

import "testing"

func TestNativeDynamicBinderTypedValuesAndConstants(t *testing.T) {
	tests := []struct {
		text string
		typ  *TypeSymbol
	}{
		{`dynamic(1)`, ScalarTypes.DynamicLong.Type()},
		{`dynamic("x")`, ScalarTypes.DynamicString.Type()},
		{`dynamic([1, 2])`, ScalarTypes.DynamicArrayOfLong.Type()},
		{`dynamic([1, 2.0])`, ScalarTypes.DynamicArray.Type()},
		{`dynamic([[1.0], [2.0]])`, ScalarTypes.DynamicArrayOfArrayOfReal.Type()},
		{`dynamic(null)`, ScalarTypes.Dynamic.Type()},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			node := ParseLiteral(test.text).(*DynamicExpression)
			if !supportsDynamicExpression(node) {
				t.Fatalf("unsupported dynamic shape: children=%d expression=%T expression-children=%d", node.ChildCount(), node.Expression(), node.Expression().ChildCount())
			}
			NewBinder(nil).TryBind(node)
			info, ok := node.GetSemanticInfo()
			if !ok || info.ResultType != test.typ || !info.IsConstant {
				t.Fatalf("semantic info = %#v, want type %v and constant", info, test.typ)
			}
		})
	}
}

func TestNativeDynamicBinderBagOrderDuplicatesDecodedNamesAndNull(t *testing.T) {
	node := ParseLiteral(`dynamic({"a\n":1,"x":null,"a\n":"last"})`).(*DynamicExpression)
	NewBinder(nil).TryBind(node)
	object := node.Expression().(*JsonObjectExpression)
	bag, ok := DynamicBagForType(ExpressionResultType(object))
	if !ok {
		t.Fatalf("object type = %v, want dynamic bag", ExpressionResultType(object))
	}
	columns := bag.Properties()
	if len(columns) != 3 || columns[0].SymbolName() != "a\n" || columns[1].SymbolName() != "x" || columns[2].SymbolName() != "a\n" {
		t.Fatalf("bag columns = %#v", columns)
	}
	if columns[0].Type != ScalarTypes.Long.Type() || columns[1].Type != ScalarTypes.Null.Type() || columns[2].Type != ScalarTypes.String.Type() {
		t.Fatalf("bag column types = %v, %v, %v", columns[0].Type, columns[1].Type, columns[2].Type)
	}
	for _, separated := range object.Pairs().Elements() {
		if separated.Element().IsBound() {
			t.Fatal("JsonPair unexpectedly received semantic info")
		}
	}
}

func TestNativeDynamicMissingCommaPreservesExplicitEvaluateSchemaCurrentReferenceParity(t *testing.T) {
	const query = `datatable(Properties:dynamic)[dynamic({'Name':'alpha''Count':2})] | evaluate bag_unpack(Properties) : (*, Name:string, Count:long) | project Name, Count`
	code := ParseAndAnalyze(query)
	assertDiagnosticLocations(t, code.GetDiagnostics(),
		diagnosticLocation{"KS005", 60, 1},
		diagnosticLocation{"KS006", 60, 1},
	)
	result := terminalResultTable(code)
	if result == nil {
		t.Fatal("missing recovered result table")
	}
	assertOrderedColumns(t, result, "Name", "Count")
	if result.Columns()[0].Type != ScalarTypes.String.Type() || result.Columns()[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("recovered column types = %v, %v", result.Columns()[0].Type, result.Columns()[1].Type)
	}
}

func TestNativeDynamicBinderDecimalDepthRestored(t *testing.T) {
	dynamic := ParseLiteral(`dynamic([decimal(1), [decimal(2)]])`).(*DynamicExpression)
	binder := NewBinder(nil)
	binder.TryBind(dynamic)
	if binder.dynamicDepth != 0 {
		t.Fatalf("dynamic depth = %d after bind", binder.dynamicDepth)
	}
	diagnostics := GetContainedDiagnosticsPreserveDuplicates(dynamic, DiagnosticsIncludeSemantic)
	codes := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		codes[i] = diagnostic.Code
	}
	if len(codes) != 2 || codes[0] != "KS215" || codes[1] != "KS215" {
		t.Fatalf("semantic diagnostic codes = %v, want [KS215 KS215]", codes)
	}

	plain := ParseLiteral(`decimal(3)`).(*LiteralExpression)
	binder.TryBind(plain)
	if len(plain.SemanticDiagnostics()) != 0 {
		t.Fatalf("plain decimal inherited dynamic context: %#v", plain.SemanticDiagnostics())
	}
}

func TestNativeDynamicBinderMalformedRollbackAndRetry(t *testing.T) {
	node := ParseLiteral(`dynamic({"a":[1,2]})`).(*DynamicExpression)
	object := node.Expression().(*JsonObjectExpression)
	array := object.Pairs().At(0).Element().Value().(*JsonArrayExpression)
	saved := array.SyntaxNode.children[1]
	array.SyntaxNode.children[1] = nil
	binder := NewBinder(nil)
	binder.TryBind(node)
	if node.IsBound() || object.IsBound() || array.IsBound() || binder.dynamicDepth != 0 {
		t.Fatal("malformed dynamic bind committed semantics or leaked depth")
	}

	array.SyntaxNode.children[1] = saved
	binder.TryBind(node)
	first := ExpressionResultType(node)
	binder.TryBind(node)
	if !node.IsBound() || ExpressionResultType(node) != first || binder.dynamicDepth != 0 {
		t.Fatal("valid retry or idempotent rebind failed")
	}
}

func TestNativeDynamicBinderCompleteJSONValueFamily(t *testing.T) {
	node := ParseLiteral(`dynamic({"negative":-1,"compound":"a" "b","nested":dynamic([2])})`).(*DynamicExpression)
	if !supportsDynamicExpression(node) {
		t.Fatal("complete upstream JSON value family was rejected")
	}
	NewBinder(nil).TryBind(node)
	bag, ok := DynamicBagForType(ExpressionResultType(node.Expression()))
	if !ok {
		t.Fatalf("object type = %v, want typed dynamic bag", ExpressionResultType(node.Expression()))
	}
	if property, found := bag.TryGetProperty("negative"); !found || property.Type != ScalarTypes.Long.Type() {
		t.Fatalf("negative property = %#v/%t, want long", property, found)
	}
	if property, found := bag.TryGetProperty("compound"); !found || property.Type != ScalarTypes.String.Type() {
		t.Fatalf("compound property = %#v/%t, want string", property, found)
	}
	if property, found := bag.TryGetProperty("nested"); !found || property.Type != ScalarTypes.DynamicArrayOfLong.Type() {
		t.Fatalf("nested property = %#v/%t, want dynamic array of long", property, found)
	}

	tenant := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true).
		WithClientSymbols([]Symbol{tenant})
	parameterValue := ParseLiteral(`dynamic({tenant})`).(*DynamicExpression)
	if !supportsDynamicExpression(parameterValue) {
		t.Fatal("JSON client parameter was rejected")
	}
	NewBinder(globals).TryBind(parameterValue)
	reference := firstBoundNode[*NameReference](t, parameterValue)
	if reference.ReferencedSymbol() != tenant || ExpressionResultType(parameterValue) != ScalarTypes.DynamicString.Type() {
		t.Fatalf("JSON client parameter = symbol:%#v dynamic:%v", reference.ReferencedSymbol(), ExpressionResultType(parameterValue))
	}
}
