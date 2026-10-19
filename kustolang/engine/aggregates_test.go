package kustolang

import "testing"

func TestAggregatesCatalogMetadata(t *testing.T) {
	if len(Aggregates.All) != 61 {
		t.Fatalf("aggregate count = %d, want 61", len(Aggregates.All))
	}

	if Aggregates.Count.SymbolName() != "count" || Aggregates.Count.ResultNameKind() != ResultNamePrefixAndFirstArgument || Aggregates.Count.ResultNamePrefix() != "count" {
		t.Fatalf("count metadata = %s %s %s", Aggregates.Count.SymbolName(), Aggregates.Count.ResultNameKind(), Aggregates.Count.ResultNamePrefix())
	}
	countSignatures := Aggregates.Count.Signatures()
	if len(countSignatures) != 2 || countSignatures[0].MinArgumentCount != 0 || countSignatures[1].Alternative != "countif" || !countSignatures[1].IsHidden {
		t.Fatalf("count signatures = %#v", countSignatures)
	}

	if !Aggregates.MakeListDeprecated.IsHidden() || Aggregates.MakeListDeprecated.Alternative() != "make_list" {
		t.Fatalf("makelist obsolete metadata = hidden:%t alt:%q", Aggregates.MakeListDeprecated.IsHidden(), Aggregates.MakeListDeprecated.Alternative())
	}
	if Aggregates.CountDistinct.OptimizedAlternative() != "dcount" || Aggregates.CountDistinctIf.OptimizedAlternative() != "dcountif" {
		t.Fatalf("count_distinct alternatives = %q %q", Aggregates.CountDistinct.OptimizedAlternative(), Aggregates.CountDistinctIf.OptimizedAlternative())
	}

	avgSignatures := Aggregates.Avg.Signatures()
	if len(avgSignatures) != 5 || avgSignatures[0].Parameters[0].TypeKind != ParameterTypeInteger || avgSignatures[2].DeclaredReturnType != ScalarTypes.Decimal.Type() {
		t.Fatalf("avg signatures = %#v", avgSignatures)
	}
	if !Aggregates.Percentiles.Signatures()[0].HasRepeatableParameters {
		t.Fatal("percentiles should have repeatable percentile parameter")
	}
	if Aggregates.Passthrough.ResultNameKind() != ResultNameFirstArgument || !Aggregates.Passthrough.IsHidden() {
		t.Fatalf("passthrough metadata = %s hidden:%t", Aggregates.Passthrough.ResultNameKind(), Aggregates.Passthrough.IsHidden())
	}
	if !Aggregates.Any.IsHidden() || Aggregates.Any.Alternative() != "take_any" || len(Aggregates.Any.Signatures()) != 3 {
		t.Fatalf("any metadata = hidden:%t alt:%q signatures:%d", Aggregates.Any.IsHidden(), Aggregates.Any.Alternative(), len(Aggregates.Any.Signatures()))
	}
	takeAnySignature := Aggregates.TakeAny.Signatures()[0]
	if takeAnySignature.Parameters[0].ArgumentKind != ArgumentKindStarAllowed || !takeAnySignature.Parameters[0].IsRepeatable() {
		t.Fatalf("take_any parameter = %#v", takeAnySignature.Parameters[0])
	}
	if !Aggregates.ArgMaxDeprecated.IsHidden() || Aggregates.ArgMaxDeprecated.Alternative() != "arg_max" || Aggregates.ArgMaxDeprecated.ResultNamePrefix() != "max" {
		t.Fatalf("argmax metadata = hidden:%t alt:%q prefix:%q", Aggregates.ArgMaxDeprecated.IsHidden(), Aggregates.ArgMaxDeprecated.Alternative(), Aggregates.ArgMaxDeprecated.ResultNamePrefix())
	}
	if Aggregates.All[len(Aggregates.All)-2] != Aggregates.CountDistinct || Aggregates.All[len(Aggregates.All)-1] != Aggregates.CountDistinctIf {
		t.Fatalf("aggregate All tail = %#v", Aggregates.All[len(Aggregates.All)-2:])
	}
}

func TestAggregatesAreAvailableThroughGlobalState(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	if len(globals.Aggregates) != len(Aggregates.All) {
		t.Fatalf("global aggregate count = %d", len(globals.Aggregates))
	}
	if globals.GetAggregate("sum") != Aggregates.Sum {
		t.Fatal("sum aggregate was not found through GlobalState")
	}

	localCount := NewScalarFunctionSymbol("count", ScalarTypes.String.Type())
	globals = NewGlobalState(NewDatabaseSymbol("db", localCount))
	if globals.GetAnyFunction("count") != localCount {
		t.Fatal("database function should take precedence over built-in aggregate")
	}

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db")))
	info := binder.ResolveFunction("countif")
	if info.ReferencedSignature() == nil || info.ReferencedSignature().MinArgumentCount != 1 {
		t.Fatalf("countif binding info = %#v", info)
	}
	if diags := binder.ValidateFunctionArgumentCount("countif", 0); len(diags) != 1 || diags[0].Code != "KS119" {
		t.Fatalf("countif argument diagnostics = %#v", diags)
	}
}

func TestPercentileCustomReturnColumns(t *testing.T) {
	signature := Aggregates.Percentiles.Signatures()[0]
	valueArg := testNameReference("DurationMs")
	p50 := MustNewLiteralExpressionTyped(SyntaxKindIntLiteralExpression, SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindIntLiteralToken, Text: "50"}))
	p95 := MustNewLiteralExpressionTyped(SyntaxKindRealLiteralExpression, SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindRealLiteralToken, Text: "95.5"}))

	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:              []Expression{valueArg, p50, p95},
		ArgumentParameters:     []*Parameter{signature.GetParameter("expr"), signature.GetParameter("percentile"), signature.GetParameter("percentile")},
		ArgumentTypes:          []*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Int.Type(), ScalarTypes.Real.Type()},
		ArgumentResultNames:    []string{"DurationMs", "", ""},
		ArgumentConstantValues: []any{nil, 50, "95.5"},
		Signature:              signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("percentiles result = %T", result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 {
		t.Fatalf("percentile columns = %#v", columns)
	}
	if columns[0].SymbolName() != "percentile_DurationMs_50" || columns[0].Type != ScalarTypes.Long.Type() {
		t.Fatalf("first percentile column = %s %s", columns[0].SymbolName(), columns[0].Type.SymbolName())
	}
	if columns[1].SymbolName() != "percentile_DurationMs_95_5" || columns[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("second percentile column = %s %s", columns[1].SymbolName(), columns[1].Type.SymbolName())
	}

	arraySignature := Aggregates.PercentilesArray.Signatures()[0]
	arrayResult := arraySignature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{valueArg, p50},
		ArgumentParameters:  []*Parameter{arraySignature.GetParameter("expr"), arraySignature.GetParameter("percentile")},
		ArgumentResultNames: []string{"DurationMs", ""},
		Signature:           arraySignature,
	})
	arrayTuple, ok := arrayResult.(*TupleSymbol)
	if !ok || len(arrayTuple.Columns()) != 1 {
		t.Fatalf("percentiles_array result = %#v", arrayResult)
	}
	if column := arrayTuple.Columns()[0]; column.SymbolName() != "percentiles_DurationMs" || column.Type != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("percentiles_array column = %s %s", column.SymbolName(), column.Type.SymbolName())
	}
}

func TestAnyCustomReturnUsesRowScopeAndSummarizeByColumns(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	app := NewColumnSymbol("AppDisplayName", ScalarTypes.String.Type())
	count := NewColumnSymbol("Count", ScalarTypes.Long.Type())
	rowScope := NewTableSymbol("SigninLogs", account, app, count)
	star := testStarExpression()
	signature := Aggregates.TakeAny.Signatures()[0]

	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:          []Expression{star},
		ArgumentParameters: []*Parameter{signature.GetParameter("expr")},
		RowScope:           rowScope,
		SummarizeByColumns: []*ColumnSymbol{app},
		Signature:          signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("take_any result = %T", result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0] != account || columns[1] != count {
		t.Fatalf("take_any columns = %#v", columns)
	}
}

func TestAnyCustomReturnDuplicatesExplicitColumns(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	first := testNameReference("Account")
	second := testNameReference("Account")
	signature := Aggregates.Any.Signatures()[1]

	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:          []Expression{first, second},
		ArgumentParameters: []*Parameter{signature.GetParameter("expr"), signature.GetParameter("expr")},
		ArgumentSemantics: []SemanticInfo{
			NewSemanticInfoForSymbol(account, account.Type),
			NewSemanticInfoForSymbol(account, account.Type),
		},
		Signature: signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("any result = %T", result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0] != account || columns[1] == account || columns[1].SymbolName() != "Account" {
		t.Fatalf("any duplicate columns = %#v", columns)
	}
	if len(columns[1].OriginalColumns) != 1 || columns[1].OriginalColumns[0] != account {
		t.Fatalf("any duplicate original columns = %#v", columns[1].OriginalColumns)
	}
}

func TestArgMinMaxCustomReturns(t *testing.T) {
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	app := NewColumnSymbol("AppDisplayName", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	rowScope := NewTableSymbol("SigninLogs", account, app, score)
	scoreArg := testNameReference("Score")
	star := testStarExpression()

	signature := Aggregates.ArgMax.Signatures()[0]
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:          []Expression{scoreArg, star},
		ArgumentParameters: []*Parameter{signature.GetParameter("maximized"), signature.GetParameter("returned")},
		ArgumentSemantics:  []SemanticInfo{NewSemanticInfoForSymbol(score, score.Type), SemanticInfo{}},
		RowScope:           rowScope,
		SummarizeByColumns: []*ColumnSymbol{app},
		Signature:          signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("arg_max result = %T", result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0] != score || columns[1] != account {
		t.Fatalf("arg_max columns = %#v", columns)
	}

	deprecated := Aggregates.ArgMaxDeprecated.Signatures()[0]
	deprecatedResult := deprecated.CustomReturnType(CustomReturnTypeContext{
		Arguments:          []Expression{scoreArg, star},
		ArgumentParameters: []*Parameter{deprecated.GetParameter("maximized"), deprecated.GetParameter("returned")},
		ArgumentSemantics:  []SemanticInfo{NewSemanticInfoForSymbol(score, score.Type), SemanticInfo{}},
		RowScope:           rowScope,
		SummarizeByColumns: []*ColumnSymbol{app},
		Signature:          deprecated,
	})
	deprecatedTuple, ok := deprecatedResult.(*TupleSymbol)
	if !ok {
		t.Fatalf("argmax result = %T", deprecatedResult)
	}
	deprecatedColumns := deprecatedTuple.Columns()
	if len(deprecatedColumns) != 2 || deprecatedColumns[0] != score || deprecatedColumns[1].SymbolName() != "Score_Account" {
		t.Fatalf("argmax columns = %#v", deprecatedColumns)
	}
	if len(deprecatedColumns[1].OriginalColumns) != 1 || deprecatedColumns[1].OriginalColumns[0] != account {
		t.Fatalf("argmax original columns = %#v", deprecatedColumns[1].OriginalColumns)
	}
}

func testStarExpression() *StarExpression {
	return MustNewStarExpressionTyped(SyntaxTokenFromLexical(LexicalToken{Kind: SyntaxKindAsteriskToken, Text: "*"}))
}
