package kustolang

import (
	"strings"
	"testing"
)

func TestNativeFunctionCallBindsArgumentsSignatureAndName(t *testing.T) {
	value := NewColumnSymbol("Value", ScalarTypes.String.Type())
	query := ParseQuery("T | project N=strlen(Value)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", value)))).TryBind(query)

	call := firstBoundNode[*FunctionCallExpression](t, query)
	if call.ReferencedSignature() == nil || call.ReferencedSignature().Symbol != Functions.Strlen {
		t.Fatalf("call signature = %#v, want strlen", call.ReferencedSignature())
	}
	if call.Name().ReferencedSignature() != call.ReferencedSignature() || call.Name().RawResultType() != call.RawResultType() {
		t.Fatalf("call/name semantics differ: call=%#v name=%#v", call.SyntaxNode, call.Name().SyntaxNode)
	}
	argument := call.ArgumentList().Expressions().At(0).Element().(*NameReference)
	if argument.ReferencedSymbol() != value {
		t.Fatalf("argument symbol = %#v, want Value", argument.ReferencedSymbol())
	}
}

func TestNativeFunctionCallOverloadAndArgumentKindDiagnostics(t *testing.T) {
	for _, test := range []struct{ text, code string }{
		{`strlen()`, "KS119"},
		{`format_datetime(datetime(2024-01-01), strcat("y", "y"))`, "KS131"},
	} {
		expression := ParseExpression(test.text)
		NewBinder(nil).TryBind(expression)
		call := firstBoundNode[*FunctionCallExpression](t, expression)
		if !diagnosticsContain(call.SemanticDiagnostics(), test.code) {
			t.Fatalf("%s diagnostics = %#v, want %s", test.text, call.SemanticDiagnostics(), test.code)
		}
	}
}

func TestNativeUserFunctionOverloadsAllowImplicitCoercion(t *testing.T) {
	newOverloads := func(name string) (*FunctionSymbol, *Signature, *Signature) {
		intSignature := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.Int.Type()))
		boolSignature := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.Bool.Type()))
		return NewFunctionSymbol(name, intSignature, boolSignature), intSignature, boolSignature
	}
	assertCall := func(t *testing.T, expressionText string, globals *GlobalState, wantSignature *Signature, wantDiagnostic string) {
		t.Helper()
		expression := ParseExpression(expressionText)
		NewBinder(globals).TryBind(expression)
		call := firstBoundNode[*FunctionCallExpression](t, expression)
		if call.ReferencedSignature() != wantSignature {
			t.Fatalf("%s signature = %p, want %p; diagnostics = %#v", expressionText, call.ReferencedSignature(), wantSignature, call.SemanticDiagnostics())
		}
		if wantDiagnostic == "" {
			if diagnostics := call.SemanticDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("%s diagnostics = %#v, want none", expressionText, diagnostics)
			}
		} else if !diagnosticsContain(call.SemanticDiagnostics(), wantDiagnostic) {
			t.Fatalf("%s diagnostics = %#v, want %s", expressionText, call.SemanticDiagnostics(), wantDiagnostic)
		}
	}

	databaseFunction, databaseInt, databaseBool := newOverloads("database_pick")
	databaseGlobals := NewGlobalState(NewDatabaseSymbol("db", databaseFunction))
	assertCall(t, `database_pick(long(1))`, databaseGlobals, databaseInt, "")
	assertCall(t, `database_pick(true)`, databaseGlobals, databaseBool, "")
	assertCall(t, `database_pick("no match")`, databaseGlobals, nil, "KS127")
	summarize := ParseQuery(`datatable(value:long)[1] | summarize x=database_pick(count())`)
	NewBinder(databaseGlobals).TryBind(summarize)
	summarizeCall := namedFunctionCall(t, summarize, "database_pick")
	if summarizeCall.ReferencedSignature() != databaseInt || len(summarizeCall.SemanticDiagnostics()) != 0 {
		t.Fatalf("summarize database_pick signature = %p, want %p; diagnostics = %#v", summarizeCall.ReferencedSignature(), databaseInt, summarizeCall.SemanticDiagnostics())
	}

	declaration := firstBoundNode[*FunctionDeclaration](t, ParseQuery(`let user_pick=(value:int) { "int" };`)).Body()
	userInt := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.Int.Type()))
	userInt.Declaration = declaration
	userBool := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.Bool.Type()))
	userBool.Declaration = declaration
	userFunction := NewFunctionSymbol("user_pick", userInt, userBool)
	userGlobals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{userFunction})
	assertCall(t, `user_pick(long(1))`, userGlobals, userInt, "")

	localBool := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.Bool.Type()))
	localBool.Declaration = declaration
	localInt := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.Int.Type()))
	localFunction := NewFunctionSymbol("local_pick", localBool, localInt)
	localGlobals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{localFunction})
	assertCall(t, `local_pick(long(1))`, localGlobals, localInt, "")

	builtInFunction, _, _ := newOverloads("builtin_pick")
	builtInGlobals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{builtInFunction})
	assertCall(t, `builtin_pick(long(1))`, builtInGlobals, nil, "KS127")
}

func TestNativeSourceReferenceFunctionInvalidOverloadDiagnosticsAndRecovery(t *testing.T) {
	tests := []struct {
		expression    string
		code          string
		resultTabular bool
	}{
		{`external_table()`, "KS120", true},
		{`external_table(1)`, "KS107", true},
		{`external_table(1,2)`, "KS127", true},
		{`graph()`, "KS120", false},
		{`graph(1)`, "KS107", false},
		{`graph(1,2)`, "KS127", false},
		{`stored_query_result()`, "KS120", true},
		{`stored_query_result(1)`, "KS107", true},
		{`stored_query_result(1,2)`, "KS107", true},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			expression := ParseExpression(test.expression)
			NewBinder(nil).TryBind(expression)
			call := firstBoundNode[*FunctionCallExpression](t, expression)
			if !diagnosticsContain(call.SemanticDiagnostics(), test.code) {
				t.Fatalf("diagnostics = %#v, want %s", call.SemanticDiagnostics(), test.code)
			}
			result := call.RawResultType()
			if result == nil {
				t.Fatal("missing recovery result type")
			}
			if test.resultTabular && result.SymbolTabularity() != TabularityTabular {
				t.Fatalf("result = %#v, want tabular recovery type", result)
			}
			if strings.HasPrefix(test.expression, "external_table") {
				table, ok := TableSymbolForType(result)
				if !ok || table == nil || !table.IsOpen() || table.SymbolKind() != SymbolKindTable {
					t.Fatalf("external-table recovery = %#v, want open ordinary table", table)
				}
			}
			if strings.HasPrefix(test.expression, "graph") && result != &EmptyGraphSymbol.TypeSymbol {
				t.Fatalf("graph recovery = %#v, want GraphSymbol.Empty type", result)
			}
		})
	}
}

func TestNativeAmbiguousFunctionSuppressesOverloadDiagnosticForUnknownArgument(t *testing.T) {
	expression := ParseExpression(`external_table(Missing)`)
	NewBinder(nil).TryBind(expression)
	call := firstBoundNode[*FunctionCallExpression](t, expression)
	if diagnosticsContain(call.SemanticDiagnostics(), "KS127") {
		t.Fatalf("call diagnostics = %#v, want no KS127 for an error or unknown argument", call.SemanticDiagnostics())
	}
}

func TestNativeAmbiguousFunctionEvaluatesEveryCustomReturnAtCallSite(t *testing.T) {
	calls := 0
	custom := func(context CustomReturnTypeContext) Symbol {
		calls++
		return NewTupleSymbolOf(NewColumnSymbol("value", context.ArgumentTypes[0]))
	}
	fn := NewFunctionSymbol("ambiguous_custom",
		NewCustomSignature(custom, TabularityScalar, NewParameter("value", ScalarTypes.Long.Type())),
		NewCustomSignature(custom, TabularityScalar, NewParameter("value", ScalarTypes.Long.Type())),
	)
	expression := ParseExpression(`ambiguous_custom(1)`)
	NewBinder(NewGlobalState(nil).WithFunctions([]*FunctionSymbol{fn})).TryBind(expression)
	call := firstBoundNode[*FunctionCallExpression](t, expression)
	if calls != 2 {
		t.Fatalf("custom return calls = %d, want one call for each ambiguous signature", calls)
	}
	if !diagnosticsContain(call.SemanticDiagnostics(), "KS127") {
		t.Fatalf("call diagnostics = %#v, want KS127", call.SemanticDiagnostics())
	}
}

func TestNativeAmbiguousFunctionEvaluatesComputedReturnsAtCallSite(t *testing.T) {
	parameter1 := NewParameter("name", ScalarTypes.String.Type(), WithArgumentKind(ArgumentKindConstant))
	parameter2 := NewParameter("name", ScalarTypes.String.Type(), WithArgumentKind(ArgumentKindConstant))
	rows := NewFunctionSymbol("Rows",
		NewComputedSignature(`table(name)`, TabularityTabular, parameter1),
		NewComputedSignature(`table(name)`, TabularityTabular, parameter2),
	)
	column := NewColumnSymbol("A", ScalarTypes.Long.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", column), rows))
	query := ParseQuery(`Rows("T") | project A`)
	NewBinder(globals).TryBind(query)
	call := namedFunctionCall(t, query, "Rows")
	if !diagnosticsContain(call.SemanticDiagnostics(), "KS127") {
		t.Fatalf("call diagnostics = %#v, want KS127", call.SemanticDiagnostics())
	}
	callResult, _ := TableSymbolForType(call.RawResultType())
	if callResult == nil {
		t.Fatalf("computed ambiguous call type = %s, want table", typeName(call.RawResultType()))
	}
	if got := callResult.GetColumn("A"); got == nil || !sameType(got.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("computed ambiguous call result = %#v, want A:long", callResult.Columns())
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	result := semanticTable(t, project)
	if got := result.GetColumn("A"); got == nil || !sameType(got.Type, ScalarTypes.Long.Type()) {
		if got == nil {
			t.Fatalf("computed ambiguous result = %#v, want A:long", result.Columns())
		}
		t.Fatalf("computed ambiguous result A type = %s, want long", typeName(got.Type))
	}
}

func TestNativeFunctionCallDynamicRetryPreservesMalformedRecovery(t *testing.T) {
	fn := NewFunctionSymbol("pick",
		NewSignature(ScalarTypes.Dynamic.Type(), NewParameter("value", ScalarTypes.Dynamic.Type())),
		NewSignature(ScalarTypes.Long.Type(), NewParameter("value", ScalarTypes.Long.Type())),
	)
	globals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{fn})
	code := ParseWithGlobals(`print x=pick(dynamic(1)`, globals).Analyze()
	call := namedFunctionCall(t, code.Syntax(), "pick")
	if call.RawResultType() != ScalarTypes.DynamicLong.Type() {
		t.Fatalf("malformed pick result = %s, want dynamic(long)", typeName(call.RawResultType()))
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS005" {
		t.Fatalf("malformed pick diagnostics = %#v, want KS005", diagnostics)
	}
}

func TestNativeIifAcceptsDifferentDynamicShapes(t *testing.T) {
	queries := []string{
		`print x=iif(true, split("a:b", ":")[0], parse_json("{}").value)`,
		`print x=iff(true, "literal", split("a-b", "-", 1)[0])`,
	}
	for _, query := range queries {
		code := ParseWithGlobals(query, DefaultGlobalState).Analyze()
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("dynamic common arguments produced diagnostics for %q: %#v", query, diagnostics)
		}
	}
}

func TestNativeSourceReferenceNonLiteralFallbackSymbolsMatchUpstream(t *testing.T) {
	externalExpression := ParseExpression(`external_table(strcat("x", "y"))`)
	NewBinder(nil).TryBind(externalExpression)
	externalCall := firstBoundNode[*FunctionCallExpression](t, externalExpression)
	externalResult, _ := externalCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	externalTable, ok := externalResult.(*TableSymbol)
	if !ok || !externalTable.IsOpen() || externalTable.SymbolKind() != SymbolKindTable {
		t.Fatalf("external-table fallback = %#v, want open ordinary table", externalResult)
	}

	graphExpression := ParseExpression(`graph(strcat("x", "y"))`)
	NewBinder(nil).TryBind(graphExpression)
	graphCall := firstBoundNode[*FunctionCallExpression](t, graphExpression)
	graphResult, _ := graphCall.SyntaxNode.Payload(functionCallResultSymbolPayload)
	if graphResult != EmptyGraphSymbol {
		t.Fatalf("graph fallback = %#v, want GraphSymbol.Empty", graphResult)
	}
}

func TestNativeFunctionCallScopeRestored(t *testing.T) {
	binder := NewBinder(nil)
	binder.scopeKind = ScopeKindAggregate
	call := ParseExpression("count()").(*FunctionCallExpression)
	newTreeBinder(binder).visitFunctionCall(call)
	if binder.scopeKind != ScopeKindAggregate || binder.pathScope != nil || binder.rowScope != nil {
		t.Fatalf("binder state not restored: scope=%v path=%#v row=%#v", binder.scopeKind, binder.pathScope, binder.rowScope)
	}
}

func TestNativeFunctionCallCompoundNamedArgumentReferenceParity(t *testing.T) {
	tests := []struct {
		name, query, column string
		typ                 *TypeSymbol
	}{
		{"scalar", `print Value=strlen((A,B)="x")`, "Value", ScalarTypes.Long.Type()},
		{"aggregate", `datatable(x:long)[1] | summarize Total=sum((A,B)=x)`, "Total", ScalarTypes.Long.Type()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query, nil)
			start := strings.Index(test.query, "(A,B)")
			assertDiagnosticLocations(t, code.GetDiagnostics(), diagnosticLocation{"KS194", start, len("(A,B)")})
			table := compoundNamedResultTable(t, code)
			if column := table.GetColumn(test.column); column == nil || !sameType(column.Type, test.typ) {
				t.Fatalf("%s result column = %#v, want %s", test.name, column, typeName(test.typ))
			}
		})
	}

	validNamedLiteral := ParseExpression(`format_datetime(datetime(2024-01-01), format="yyyy")`)
	NewBinder(nil).TryBind(validNamedLiteral)
	if diagnostics := GetContainedDiagnostics(validNamedLiteral, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
		t.Fatalf("simple named literal argument diagnostics = %#v", diagnostics)
	}

	malformed := ParseAndAnalyze(`print Value=strlen((A,B)=)`, nil)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("missing compound argument RHS produced no diagnostics")
	}

	t.Run("wrong count suppresses compound diagnostic", func(t *testing.T) {
		expression := ParseExpression(`strlen((A,B)="x", "extra")`)
		NewBinder(nil).TryBind(expression)
		call := firstBoundNode[*FunctionCallExpression](t, expression)
		if diagnostics := call.SemanticDiagnostics(); len(diagnostics) == 0 || diagnosticsContain(diagnostics, "KS194") {
			t.Fatalf("wrong-count diagnostics = %#v, want a signature diagnostic without KS194", diagnostics)
		}
	})
}

func TestNativeFunctionCallNamedArgumentSignatureDiagnosticsReferenceParity(t *testing.T) {
	a := NewParameter("A", ScalarTypes.String.Type())
	b := NewParameter("B", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1))
	function := NewFunctionSymbol("F", NewSignature(ScalarTypes.String.Type(), a, b))
	globals := NewGlobalState(NewDatabaseSymbol("db", function))
	tests := []struct {
		name, expression, code, target string
	}{
		{"valid reordered", `F(B="b", A="a")`, "", ""},
		{"unknown name", `F(C="c", A="a")`, "KS196", "C"},
		{"missing required", `F(B="b")`, "KS197", "F"},
		{"unnamed after reordered", `F(B="b", "a")`, "KS195", `"a"`},
		{"compound remains positional", `F((A)="a")`, "KS194", "(A)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression := ParseExpression(test.expression)
			NewBinder(globals).TryBind(expression)
			call := firstBoundNode[*FunctionCallExpression](t, expression)
			if !sameType(call.RawResultType(), ScalarTypes.String.Type()) {
				t.Fatalf("call result = %s, want string", typeName(call.RawResultType()))
			}
			if test.code == "" {
				assertDiagnosticCodes(t, call)
				return
			}
			start := strings.LastIndex(test.expression, test.target)
			assertDiagnosticLocations(t, call.SemanticDiagnostics(), diagnosticLocation{test.code, start, len(test.target)})
		})
	}

	t.Run("out-of-order named suppresses compound diagnostic", func(t *testing.T) {
		expressionText := `F(B="b", (A)="a")`
		expression := ParseExpression(expressionText)
		NewBinder(globals).TryBind(expression)
		call := firstBoundNode[*FunctionCallExpression](t, expression)
		start := strings.Index(expressionText, `(A)="a"`)
		assertDiagnosticLocations(t, call.SemanticDiagnostics(), diagnosticLocation{"KS195", start, len(`(A)="a"`)})
	})
}

func TestNativeFunctionCallCompoundNamedArgumentTypeChecksReferenceParity(t *testing.T) {
	tests := []struct {
		expression, code, target string
	}{
		{`T | project y=strlen((A,B)=1)`, "KS141", `(A,B)=1`},
		{`T | project y=format_datetime(datetime(2024-01-01), (A,B)="yyyy")`, "KS131", `(A,B)="yyyy"`},
		{`T | project y=prev((A,B)=x)`, "KS109", `(A,B)=x`},
	}
	row := NewTableSymbol("T", NewColumnSymbol("x", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", row))
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			code := ParseAndAnalyze(test.expression, globals)
			call := firstBoundNode[*FunctionCallExpression](t, code.Syntax())
			diagnostics := call.SemanticDiagnostics()
			if !diagnosticsContain(diagnostics, "KS194") || !diagnosticsContain(diagnostics, test.code) {
				t.Fatalf("diagnostics = %#v, want KS194 and %s", diagnostics, test.code)
			}
			start := strings.Index(test.expression, test.target)
			found := false
			for _, diagnostic := range diagnostics {
				if diagnostic.Code == test.code && diagnostic.Start == start && diagnostic.Length == len(test.target) {
					found = true
				}
			}
			if !found {
				t.Fatalf("diagnostics = %#v, want %s at %d/%d", diagnostics, test.code, start, len(test.target))
			}
		})
	}
}

func TestNativeFunctionCallCustomReturnContextUsesBinderNamesAndConstantAliases(t *testing.T) {
	t.Run("bin auto query parameters", func(t *testing.T) {
		query := ParseQuery(`declare query_parameters(query_bin_auto_size:real, query_bin_auto_at:real); print x=bin_auto(int(1))`)
		NewBinder(nil).TryBind(query)
		printOperator := firstBoundNode[*PrintOperator](t, query)
		assertDiagnosticCodes(t, printOperator)
		column := semanticTable(t, printOperator).GetColumn("x")
		if column == nil || !sameType(column.Type, ScalarTypes.Real.Type()) {
			t.Fatalf("bin_auto result = %#v, want real", column)
		}
	})

	t.Run("bag set key constant path alias", func(t *testing.T) {
		query := ParseQuery(`let Path = "$.added"; print B=bag_set_key(bag_pack("base", 1), Path, "value")`)
		NewBinder(nil).TryBind(query)
		call := namedFunctionCall(t, query, "bag_set_key")
		assertDiagnosticCodes(t, call)
		result, _ := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
		bag, ok := result.(*DynamicBagSymbol)
		if !ok {
			t.Fatalf("bag_set_key result = %T, want dynamic bag", result)
		}
		if base, found := bag.TryGetProperty("base"); !found || !sameType(base.Type, ScalarTypes.Long.Type()) {
			t.Fatalf("bag_set_key base property = %#v", base)
		}
		if added, found := bag.TryGetProperty("added"); !found || !sameType(added.Type, ScalarTypes.String.Type()) {
			t.Fatalf("bag_set_key aliased path property = %#v", added)
		}
	})
}

func TestNativeFunctionCallCustomReturnContextMatchesUpstreamBinderState(t *testing.T) {
	var captured CustomReturnTypeContext
	var referenced Symbol
	capture := NewFunctionSymbol("capture", NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
		captured = context
		referenced = context.GetReferencedSymbol("Alias")
		return ScalarTypes.Long.Type()
	}, TabularityScalar))
	database := NewDatabaseSymbol("db", capture)
	cluster := NewClusterSymbol("cluster", database)
	globals := NewGlobalState(database).WithCluster(cluster).WithDatabase(database)
	query := ParseQuery(`let Alias = 1; print x=capture()`)
	binder := NewBinder(globals)
	binder.TryBind(query)
	call := namedFunctionCall(t, query, "capture")
	if captured.Location != call.SyntaxNode {
		t.Fatalf("context location = %#v, want call node", captured.Location)
	}
	if captured.RowScope != EmptyTableSymbol || captured.CurrentCluster != cluster || captured.CurrentDatabase != database || captured.Globals != binder.Globals {
		t.Fatalf("context state = row:%#v cluster:%#v database:%#v globals:%#v", captured.RowScope, captured.CurrentCluster, captured.CurrentDatabase, captured.Globals)
	}
	variable, ok := referenced.(*VariableSymbol)
	if !ok || variable.SymbolName() != "Alias" || !variable.IsConstant || ExpressionConstantValue(variable.Source) != int64(1) {
		t.Fatalf("context name lookup = %#v", referenced)
	}
	if captured.GetResultType("Alias") != ScalarTypes.Long.Type() || captured.CurrentFunction != nil {
		t.Fatalf("context result/current function = %s/%#v", typeName(captured.GetResultType("Alias")), captured.CurrentFunction)
	}
}
