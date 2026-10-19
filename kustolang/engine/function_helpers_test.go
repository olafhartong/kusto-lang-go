package kustolang

import "testing"

func TestMakeColumnNameAndValidNameFragment(t *testing.T) {
	if got := MakeColumnName("series", "value", "avg"); got != "series_value_avg" {
		t.Fatalf("MakeColumnName = %q", got)
	}
	if got := MakeColumnName("value"); got != "value" {
		t.Fatalf("single-part MakeColumnName = %q", got)
	}
	if got := MakeColumnName("left", "", "right"); got != "left__right" {
		t.Fatalf("MakeColumnName should retain empty string parts, got %q", got)
	}
	if got := MakeValidNameFragment("a-b c"); got != "a_b_c" {
		t.Fatalf("MakeValidNameFragment = %q", got)
	}
	if got := MakeValidNameFragment("abc123"); got != "abc123" {
		t.Fatalf("MakeValidNameFragment changed valid text to %q", got)
	}
}

func TestCustomReturnTypeContextArgumentHelpers(t *testing.T) {
	valueParam := NewParameter("value", ScalarTypes.String.Type())
	dimensionParam := NewParameterKind("dimension", ParameterTypeNotDynamic, WithParameterOccurrence(0, MaxRepeat))
	signature := NewSignature(ScalarTypes.Dynamic.Type(), valueParam, dimensionParam)
	valueArg := MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindStringLiteralToken, Text: "'a'"}))
	dimArg := testNameReference("Account")
	valueColumn := NewColumnSymbol("Value", ScalarTypes.String.Type())
	dimColumn := NewColumnSymbol("Account", ScalarTypes.String.Type())
	context := CustomReturnTypeContext{
		Arguments:              []Expression{valueArg, dimArg},
		ArgumentParameters:     []*Parameter{valueParam, dimensionParam},
		ArgumentTypes:          []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()},
		ArgumentSemantics:      []SemanticInfo{NewSemanticInfoForSymbol(valueColumn, valueColumn.Type), NewSemanticInfoForSymbol(dimColumn, dimColumn.Type)},
		ArgumentResultNames:    []string{"ValueExpr", "Account"},
		ArgumentConstantValues: []any{"constant-value"},
		Signature:              signature,
	}

	if context.GetArgument("value") != valueArg {
		t.Fatal("GetArgument returned wrong value argument")
	}
	if context.GetArgumentForParameter(valueParam) != valueArg {
		t.Fatal("GetArgumentForParameter returned wrong value argument")
	}
	if got := context.GetArguments("dimension"); len(got) != 1 || got[0] != dimArg {
		t.Fatalf("GetArguments(dimension) = %#v", got)
	}
	if got := context.GetArgumentsForParameter(dimensionParam); len(got) != 1 || got[0] != dimArg {
		t.Fatalf("GetArgumentsForParameter(dimension) = %#v", got)
	}
	if got := context.GetArguments("missing"); got == nil || len(got) != 0 {
		t.Fatalf("GetArguments(missing) = %#v, want non-nil empty", got)
	}
	if context.GetArgumentReferencedSymbol("dimension") != dimColumn {
		t.Fatal("GetReferencedSymbol returned wrong column")
	}
	if got := context.GetResultName(valueArg); got != "ValueExpr" {
		t.Fatalf("GetResultName = %q", got)
	}
	if got := GetConstantValue(valueArg, context); got != "constant-value" {
		t.Fatalf("GetConstantValue = %q", got)
	}
	if !IsBoolean(valueArg, CustomReturnTypeContext{Arguments: []Expression{valueArg}, ArgumentTypes: []*TypeSymbol{ScalarTypes.Bool.Type()}}) {
		t.Fatal("IsBoolean should use context argument type")
	}
}

func TestFunctionHelpersConstantListsAndSummarizeColumns(t *testing.T) {
	dynamic, ok := ParseExpression(`dynamic(["A", 2, true, Missing])`).(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic expression = %T", ParseExpression(`dynamic(["A", 2, true, Missing])`))
	}
	array, ok := dynamic.Expression().(*JsonArrayExpression)
	if !ok {
		t.Fatalf("dynamic value = %T", dynamic.Expression())
	}
	if got := GetConstantValues(array.Values()); len(got) != 4 ||
		got[0] != "A" || got[1] != "2" || got[2] != "True" || got[3] != "" {
		t.Fatalf("constant values = %#v", got)
	}
	if got := GetConstantValues(nil); got == nil || len(got) != 0 {
		t.Fatalf("nil constant list = %#v, want non-nil empty", got)
	}

	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	payload := NewColumnSymbol("Payload", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", account, score, payload)
	query := ParseQuery(`T | summarize arg_max(Score, *) by Alias=Account`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	call := firstBoundNode[*FunctionCallExpression](t, query)
	columns := GetSummarizeByColumns(functionCallArguments(call))
	if len(columns) != 1 || columns[0] != account {
		t.Fatalf("summarize by columns = %#v, want Account identity", columns)
	}
	summarize := firstBoundNode[*SummarizeOperator](t, query)
	result := expressionTableResult(summarize)
	if result == nil || result.GetColumn("Alias") == nil || result.GetColumn("Score") == nil ||
		result.GetColumn("Payload") == nil || result.GetColumn("Account") != nil {
		t.Fatalf("named summarize/arg_max result = %#v", result)
	}
}

func TestFunctionHelpersMalformedSummarizeAndConversionLineage(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	payload := NewColumnSymbol("Payload", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", account, payload)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	malformed := ParseQuery(`T | summarize take_any(*) by Alias=`)
	NewBinder(globals).TryBind(malformed)
	if diagnostics := GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault); len(diagnostics) == 0 {
		t.Fatal("malformed summarize produced no diagnostics")
	}
	call := firstBoundNode[*FunctionCallExpression](t, malformed)
	if columns := GetSummarizeByColumns(functionCallArguments(call)); columns == nil {
		t.Fatal("malformed summarize helper returned nil")
	}

	projected := ParseQuery(`T | project todynamic(Payload)`)
	NewBinder(globals).TryBind(projected)
	project := firstBoundNode[*ProjectOperator](t, projected)
	result := expressionTableResult(project)
	if result == nil || result.GetColumn("Payload") != payload {
		t.Fatalf("todynamic projection lineage = %#v, want original Payload column", result)
	}

	boolean := ParseQuery(`T | project toboolean(Account)`)
	NewBinder(globals).TryBind(boolean)
	booleanProject := firstBoundNode[*ProjectOperator](t, boolean)
	booleanResult := expressionTableResult(booleanProject)
	if booleanResult == nil || booleanResult.GetColumn("Account") == account {
		t.Fatalf("toboolean must not use source conversion-column identity: %#v", booleanResult)
	}
}

func TestFunctionHelpersBooleanAndProjectionPublicSurface(t *testing.T) {
	enabled := NewColumnSymbol("Enabled", ScalarTypes.Bool.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	payload := NewColumnSymbol("Payload", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", enabled, score, payload)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	query := ParseQuery(`T | where Enabled and Score > 0 | project Payload.user["id"], todynamic(Payload), toboolean(Enabled)`)
	NewBinder(globals).TryBind(query)

	filter := firstBoundNode[*FilterOperator](t, query)
	if !IsBoolean(filter.Condition()) {
		t.Fatalf("bound filter condition type = %s, want bool", typeName(ExpressionResultType(filter.Condition())))
	}
	if IsBoolean(firstBoundNode[*FunctionCallExpression](t, query)) {
		t.Fatal("dynamic conversion call reported boolean")
	}

	project := firstBoundNode[*ProjectOperator](t, query)
	path := project.Expressions().At(0).Element()
	if got := GetGraphExpressionResultName(path); got != "Payload_user_id" {
		t.Fatalf("graph expression result name = %q", got)
	}
	dynamicCall, _ := project.Expressions().At(1).Element().(*FunctionCallExpression)
	booleanCall, _ := project.Expressions().At(2).Element().(*FunctionCallExpression)
	if !IsConversionFunction(Functions.ToDynamic) || !IsConversionFunctionExpression(dynamicCall) {
		t.Fatal("todynamic is not classified as a conversion function")
	}
	if IsConversionFunction(Functions.ToBoolean) || IsConversionFunctionExpression(booleanCall) {
		t.Fatal("toboolean must preserve the pinned source conversion-classification omission")
	}

	summarizeQuery := ParseQuery(`T | summarize n=count() by Alias=Score`)
	NewBinder(globals).TryBind(summarizeQuery)
	summarize := firstBoundNode[*SummarizeOperator](t, summarizeQuery)
	named := summarize.ByClause().Expressions().At(0).Element()
	if got := GetExpressionDeclaredName(named); got != "Alias" || GetUnderlyingExpression(named) == named {
		t.Fatalf("declared/underlying expression = %q/%T", got, GetUnderlyingExpression(named))
	}
}

func TestFunctionHelpersMalformedBooleanAndGraphNames(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Payload", ScalarTypes.Dynamic.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	whereQuery := ParseQuery(`T | where (`)
	NewBinder(globals).TryBind(whereQuery)
	if diagnostics := GetContainedDiagnostics(whereQuery, DiagnosticsIncludeDefault); len(diagnostics) == 0 {
		t.Fatal("malformed boolean query produced no diagnostics")
	}
	filter := firstBoundNode[*FilterOperator](t, whereQuery)
	if IsBoolean(filter.Condition()) && isErrorTypeSymbol(ExpressionResultType(filter.Condition())) {
		t.Fatal("missing error expression reported boolean")
	}

	projectQuery := ParseQuery(`T | project Payload.user[`)
	NewBinder(globals).TryBind(projectQuery)
	if diagnostics := GetContainedDiagnostics(projectQuery, DiagnosticsIncludeDefault); len(diagnostics) == 0 {
		t.Fatal("malformed graph-path query produced no diagnostics")
	}
	project := firstBoundNode[*ProjectOperator](t, projectQuery)
	if got := GetGraphExpressionResultName(project.Expressions().At(0).Element()); got != "Payload_user_" {
		t.Fatalf("malformed graph result name = %q", got)
	}
}

func TestAddReferencedColumnsAndPrefixedTuple(t *testing.T) {
	valueParam := NewParameter("value", ScalarTypes.String.Type())
	dimensionParam := NewParameterKind("dimension", ParameterTypeNotDynamic, WithParameterOccurrence(0, MaxRepeat))
	signature := NewSignature(ScalarTypes.Dynamic.Type(), valueParam, dimensionParam)
	fn := NewFunctionSymbol("series_stats", signature).WithResultNamePrefix("stats")
	signature.Symbol = fn

	valueArg := testNameReference("Score")
	dimArg := testNameReference("Account")
	valueColumn := NewColumnSymbol("Score", ScalarTypes.Real.Type())
	dimColumn := NewColumnSymbol("Account", ScalarTypes.String.Type())
	context := CustomReturnTypeContext{
		Arguments:           []Expression{valueArg, dimArg},
		ArgumentParameters:  []*Parameter{valueParam, dimensionParam},
		ArgumentSemantics:   []SemanticInfo{NewSemanticInfoForSymbol(valueColumn, valueColumn.Type), NewSemanticInfoForSymbol(dimColumn, dimColumn.Type)},
		ArgumentResultNames: []string{"Score", "Account"},
		Signature:           signature,
	}

	var columns []*ColumnSymbol
	columns = AddReferencedColumn(columns, context, "value")
	columns = AddReferencedColumns(columns, context, "dimension")
	if len(columns) != 2 || columns[0] != valueColumn || columns[1] != dimColumn {
		t.Fatalf("referenced columns = %#v", columns)
	}

	base := NewTupleSymbolOf(NewColumnSymbol("avg", ScalarTypes.Real.Type()), NewColumnSymbol("stdev", ScalarTypes.Real.Type()))
	prefixed := MakePrefixedTuple(context, "value", base)
	if got := prefixed.Columns(); len(got) != 2 || got[0].SymbolName() != "stats_Score_avg" || got[1].SymbolName() != "stats_Score_stdev" {
		t.Fatalf("prefixed tuple columns = %#v", got)
	}
	if got := prefixed.Columns()[0].OriginalColumns; len(got) != 1 || got[0].SymbolName() != "avg" {
		t.Fatalf("prefixed tuple original columns = %#v", got)
	}
}

func testNameReference(name string) *NameReference {
	return MustNewNameReferenceTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindIdentifierToken, Text: name})),
		SymbolMatchDefault,
	)
}

func TestColumnAndFunctionCloneHelpers(t *testing.T) {
	source := SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindIdentifierToken, Text: "source"})
	original := NewColumnSymbol("Original", ScalarTypes.String.Type())
	renamed := original.WithName("Renamed").WithType(ScalarTypes.Long.Type()).WithDescription("desc").WithOriginalColumns(original).WithSource(source).WithExamples([]string{"1"})
	if renamed.SymbolName() != "Renamed" || renamed.Type != ScalarTypes.Long.Type() || renamed.Description != "desc" {
		t.Fatalf("renamed column metadata not preserved")
	}
	if len(renamed.OriginalColumns) != 1 || renamed.OriginalColumns[0] != original || renamed.Source != source || renamed.Examples[0] != "1" {
		t.Fatalf("renamed column lineage/source/examples not preserved")
	}

	fn := NewScalarFunctionSymbol("f", ScalarTypes.String.Type()).
		WithDescription("function").
		WithResultNamePrefix("prefix").
		WithResultNameKind(ResultNamePrefix).
		WithIsConstantFoldable(true).
		WithCustomAvailability(func(context CustomAvailabilityContext) bool { return context.InGraphWhereOrProjectClause }).
		WithOptimizedAlternative("g").
		Obsolete("h")
	if !fn.IsHidden() || !fn.IsConstantFoldable() || fn.Description() != "function" || fn.ResultNamePrefix() != "prefix" || fn.ResultNameKind() != ResultNamePrefix || fn.OptimizedAlternative() != "g" || fn.Alternative() != "h" || fn.CustomAvailability() == nil || !fn.IsAvailable(CustomAvailabilityContext{InGraphWhereOrProjectClause: true}) {
		t.Fatalf("function clone helpers did not preserve metadata")
	}
	if !fn.IsObsolete() {
		t.Fatal("obsolete function should report IsObsolete")
	}
	if same := fn.WithDescription("function").WithResultNamePrefix("prefix").WithResultNameKind(ResultNamePrefix).WithIsConstantFoldable(true).WithIsHidden(true).WithIsView(false).WithOptimizedAlternative("g").WithIsObsolete(true, "h"); same != fn {
		t.Fatal("unchanged FunctionSymbol With* helpers should return the same instance")
	}
	if cleared := fn.WithIsObsolete(false); cleared.IsObsolete() || cleared.Alternative() != "" || !fn.IsObsolete() {
		t.Fatalf("WithIsObsolete(false) should clear obsolete state without mutating original")
	}
	emptyAlternative := NewScalarFunctionSymbol("old", ScalarTypes.String.Type()).WithIsObsolete(true, "")
	if !emptyAlternative.IsObsolete() || emptyAlternative.Alternative() != "" {
		t.Fatalf("empty alternative still represents upstream obsolete state")
	}
}

func TestFunctionArgumentBoundsAndReturnType(t *testing.T) {
	one := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.String.Type()))
	rangeSig := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("required", ScalarTypes.String.Type()),
		NewParameter("optional", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1)),
		NewParameterKind("terms", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat)),
	)
	fn := NewFunctionSymbol("mixed", one, rangeSig)

	if fn.MinArgumentCount() != 1 || fn.MaxArgumentCount() != MaxRepeat+2 {
		t.Fatalf("function argument bounds = %d..%d", fn.MinArgumentCount(), fn.MaxArgumentCount())
	}
	if got := fn.GetReturnType(NewGlobalState(nil)); got != ScalarTypes.Bool.Type() {
		t.Fatalf("function return type = %#v", got)
	}

	table := NewTableSymbol("Rows", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	tableFn := NewFunctionSymbol("rows", NewSymbolSignature(table))
	if got := tableFn.GetReturnType(NewGlobalState(nil)); got != &table.TypeSymbol {
		t.Fatalf("table-valued function return type = %#v", got)
	}
	if got := NewFunctionSymbol("computed", NewComputedSignature("T | take 1", TabularityTabular)).GetReturnType(NewGlobalState(nil)); got == nil || got.SymbolName() != "" {
		t.Fatalf("computed function return type = %#v", got)
	}
}

func TestSignatureMetadataHelpersAndBestEffortReturnType(t *testing.T) {
	scalar := NewSignature(ScalarTypes.String.Type())
	if !scalar.IsScalar() || scalar.IsTabular() || scalar.GetReturnType(NewGlobalState(nil)) != ScalarTypes.String.Type() {
		t.Fatalf("declared scalar signature metadata failed")
	}

	tabular := NewReturnKindSignature(ReturnTypeParameter0Table, NewParameter("name", ScalarTypes.String.Type()))
	if tabular.IsScalar() || !tabular.IsTabular() {
		t.Fatalf("source-reference table signature tabularity = scalar:%t tabular:%t", tabular.IsScalar(), tabular.IsTabular())
	}
	if got := tabular.GetReturnType(NewGlobalState(nil)); got == nil || got.SymbolName() != "" {
		t.Fatalf("source-reference table fallback return type = %#v", got)
	}

	graph := NewReturnKindSignature(ReturnTypeParameter0Graph, NewParameter("name", ScalarTypes.String.Type()))
	if graph.IsScalar() || graph.IsTabular() || graph.GetReturnType(NewGlobalState(nil)) != ScalarTypes.Unknown.Type() {
		t.Fatalf("graph signature metadata = scalar:%t tabular:%t return:%#v", graph.IsScalar(), graph.IsTabular(), graph.GetReturnType(NewGlobalState(nil)))
	}

	hidden := scalar.Hide()
	if !hidden.IsHidden || scalar.IsHidden {
		t.Fatal("Hide should clone signature hidden state")
	}
	alternative := scalar.WithAlternative("")
	if !alternative.IsObsolete() || alternative.Alternative != "" {
		t.Fatalf("empty alternative should still mark signature obsolete")
	}
	cleared := alternative.WithIsObsolete(false, "")
	if cleared.IsObsolete() || cleared.Alternative != "" || !alternative.IsObsolete() {
		t.Fatalf("WithIsObsolete(false) should clear obsolete state without mutating original")
	}
}

func TestSignatureArgumentAwareReturnType(t *testing.T) {
	globals := NewGlobalState(nil)

	sum := Aggregates.Sum.Signatures()[0]
	if got := sum.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Int.Type()}); got != ScalarTypes.Long.Type() {
		t.Fatalf("sum(int) return type = %s, want long", typeName(got))
	}

	add := Operators.Add.Signatures()[0]
	if got := add.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Real.Type()}); got != ScalarTypes.Real.Type() {
		t.Fatalf("int + real return type = %s, want real", typeName(got))
	}

	iif := Functions.Iif.Signatures()[0]
	if got := iif.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Bool.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.String.Type()}); got != ScalarTypes.String.Type() {
		t.Fatalf("iif(bool,dynamic,string) return type = %s, want string", typeName(got))
	}

	coalesce := Functions.Coalesce.Signatures()[0]
	if got := coalesce.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Long.Type()}); got != ScalarTypes.Long.Type() {
		t.Fatalf("coalesce(int,long) return type = %s, want long", typeName(got))
	}

	first := NewParameterKind("first", ParameterTypeScalar)
	second := NewParameterKind("second", ParameterTypeScalar)
	fromSecond := NewReturnKindSignature(ReturnTypeParameter1, first, second)
	if got := fromSecond.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.DateTime.Type()}); got != ScalarTypes.DateTime.Type() {
		t.Fatalf("Parameter1 return type = %s, want datetime", typeName(got))
	}
	if got := fromSecond.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.String.Type()}); got != &Error.TypeSymbol {
		t.Fatalf("missing Parameter1 return type = %s, want error", typeName(got))
	}

	makeList := Aggregates.MakeList.Signatures()[0]
	if got := makeList.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.String.Type()}); got != ScalarTypes.DynamicArrayOfString.Type() {
		t.Fatalf("make_list(string) return type = %s, want dynamic array of string", typeName(got))
	}
	if got := makeList.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Int.Type()}); got != ScalarTypes.DynamicArrayOfLong.Type() {
		t.Fatalf("make_list(int) return type = %s, want dynamic array of long", typeName(got))
	}
	if got := makeList.GetReturnTypeForArgumentTypes(globals, []*TypeSymbol{ScalarTypes.Decimal.Type()}); got != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("make_list(decimal) return type = %s, want generic dynamic array", typeName(got))
	}
}

func TestSignatureArgumentExpressionReturnType(t *testing.T) {
	globals := NewGlobalState(nil)

	cast := Functions.Cast.Signatures()[0]
	castArgs := []Expression{
		functionCallArgument(t, `__cast(typeof(long), true)`, 0),
		functionCallArgument(t, `__cast(typeof(long), true)`, 1),
	}
	if got := cast.GetReturnTypeForArguments(globals, castArgs, []*TypeSymbol{ScalarTypes.TypeLit.Type(), ScalarTypes.Bool.Type()}); got != ScalarTypes.Long.Type() {
		t.Fatalf("__cast(typeof(long), bool) return type = %s, want long", typeName(got))
	}

	invoke := Functions.Invoke.Signatures()[0]
	invokeArgs := []Expression{
		functionCallArgument(t, `__invoke("fn", typeof(datetime), 1)`, 0),
		functionCallArgument(t, `__invoke("fn", typeof(datetime), 1)`, 1),
		functionCallArgument(t, `__invoke("fn", typeof(datetime), 1)`, 2),
	}
	if got := invoke.GetReturnTypeForArguments(globals, invokeArgs, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.TypeLit.Type(), ScalarTypes.Long.Type()}); got != ScalarTypes.DateTime.Type() {
		t.Fatalf("__invoke type-literal return type = %s, want datetime", typeName(got))
	}

	last := NewReturnKindSignature(
		ReturnTypeParameterNLiteral,
		NewParameter("value", ScalarTypes.String.Type()),
		NewParameter("typ", ScalarTypes.TypeLit.Type()),
	)
	lastArgs := []Expression{
		functionCallArgument(t, `f("x", typeof(real))`, 0),
		functionCallArgument(t, `f("x", typeof(real))`, 1),
	}
	if got := last.GetReturnTypeForArguments(globals, lastArgs, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.TypeLit.Type()}); got != ScalarTypes.Real.Type() {
		t.Fatalf("ParameterNLiteral return type = %s, want real", typeName(got))
	}

	if got, ok := TypeOfLiteralResultType(ParseExpression("typeof(string)")); !ok || got != ScalarTypes.String.Type() {
		t.Fatalf("typeof(string) result type = %s/%t, want string/true", typeName(got), ok)
	}
	if got := cast.GetReturnTypeForArguments(globals, []Expression{ParseExpression("TypeName"), ParseExpression("1")}, []*TypeSymbol{ScalarTypes.TypeLit.Type(), ScalarTypes.Long.Type()}); got != &Error.TypeSymbol {
		t.Fatalf("non-type-literal cast return type = %s, want error", typeName(got))
	}
}
