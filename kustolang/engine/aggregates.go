package kustolang

// Ported from src/Kusto.Language/Aggregates.cs. Native summarize binding fills
// RowScope and SummarizeByColumns for parsed custom aggregate calls.

type aggregateSet struct {
	Sum          *FunctionSymbol
	SumIf        *FunctionSymbol
	Cnt          *FunctionSymbol
	Count        *FunctionSymbol
	CountIf      *FunctionSymbol
	DCount       *FunctionSymbol
	DCountIf     *FunctionSymbol
	TDigest      *FunctionSymbol
	TDigestMerge *FunctionSymbol
	MergeTDigest *FunctionSymbol
	Hll          *FunctionSymbol
	HllIf        *FunctionSymbol
	HllMerge     *FunctionSymbol
	Min          *FunctionSymbol
	MinIf        *FunctionSymbol
	Max          *FunctionSymbol
	MaxIf        *FunctionSymbol
	Avg          *FunctionSymbol
	AvgIf        *FunctionSymbol

	MakeListDeprecated *FunctionSymbol
	MakeList           *FunctionSymbol
	MakeListIf         *FunctionSymbol
	MakeListWithNulls  *FunctionSymbol
	MakeSetDeprecated  *FunctionSymbol
	MakeSet            *FunctionSymbol
	MakeSetIf          *FunctionSymbol
	Passthrough        *FunctionSymbol
	MakeDictionary     *FunctionSymbol
	MakeBag            *FunctionSymbol
	MakeBagIf          *FunctionSymbol
	BuildSchema        *FunctionSymbol
	BinaryAllOr        *FunctionSymbol
	BinaryAllAnd       *FunctionSymbol
	BinaryAllXor       *FunctionSymbol
	CountDistinct      *FunctionSymbol
	CountDistinctIf    *FunctionSymbol

	Percentile        *FunctionSymbol
	Percentiles       *FunctionSymbol
	PercentilesArray  *FunctionSymbol
	PercentileW       *FunctionSymbol
	PercentilesW      *FunctionSymbol
	PercentilesWArray *FunctionSymbol
	Stdev             *FunctionSymbol
	StdevIf           *FunctionSymbol
	Stdevp            *FunctionSymbol
	Variance          *FunctionSymbol
	VarianceIf        *FunctionSymbol
	Variancep         *FunctionSymbol
	VariancepIf       *FunctionSymbol
	Covariance        *FunctionSymbol
	CovarianceIf      *FunctionSymbol
	Covariancep       *FunctionSymbol
	CovariancepIf     *FunctionSymbol
	Any               *FunctionSymbol
	TakeAny           *FunctionSymbol
	AnyIf             *FunctionSymbol
	TakeAnyIf         *FunctionSymbol
	ArgMin            *FunctionSymbol
	ArgMax            *FunctionSymbol
	ArgMinDeprecated  *FunctionSymbol
	ArgMaxDeprecated  *FunctionSymbol

	All []*FunctionSymbol
}

var Aggregates = newAggregateSet()

func newAggregateSet() *aggregateSet {
	a := &aggregateSet{}

	a.Sum = aggregateWithPrefix(aggregateReturnKind("sum", ReturnTypeParameter0Promoted, pk("expr", ParameterTypeSummable)), "sum")
	a.SumIf = aggregateWithPrefix(aggregateReturnKind("sumif", ReturnTypeParameter0Promoted, pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "sumif")
	a.Cnt = aggregateWithPrefix(aggregateScalar("cnt", ScalarTypes.Long.Type()), "cnt").Obsolete("count").Hide()
	a.Count = aggregateWithPrefix(NewFunctionSymbol("count",
		NewSignature(ScalarTypes.Long.Type()),
		NewSignature(ScalarTypes.Long.Type(), p("predicate", ScalarTypes.Bool)).Hide().Obsolete("countif"),
	), "count")
	a.CountIf = aggregateWithPrefix(aggregateScalar("countif", ScalarTypes.Long.Type(), p("predicate", ScalarTypes.Bool)), "countif")
	a.DCount = aggregateWithPrefix(aggregateScalar("dcount", ScalarTypes.Long.Type(), pk("expr", ParameterTypeScalar), pk("accuracy", ParameterTypeNotDynamic, WithParameterOccurrence(0, 1))), "dcount")
	a.DCountIf = aggregateWithPrefix(aggregateScalar("dcountif", ScalarTypes.Long.Type(), pk("expr", ParameterTypeScalar), p("predicate", ScalarTypes.Bool), pk("accuracy", ParameterTypeNotDynamic, WithParameterOccurrence(0, 1))), "dcountif")
	a.TDigest = aggregateWithPrefix(aggregateScalar("tdigest", ScalarTypes.DynamicArray.Type(), pk("expr", ParameterTypeSummable), pk("weight", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "tdigest")
	a.TDigestMerge = aggregateWithPrefix(aggregateScalar("tdigest_merge", ScalarTypes.DynamicArray.Type(), ap("tdigest", ScalarTypes.DynamicArray.Type())), "merge_tdigests")
	a.MergeTDigest = aggregateWithPrefix(aggregateScalar("merge_tdigest", ScalarTypes.DynamicArray.Type(), ap("tdigest", ScalarTypes.DynamicArray.Type())), "merge_tdigests")
	a.Hll = aggregateWithPrefix(aggregateScalar("hll", ScalarTypes.DynamicArray.Type(), pk("expr", ParameterTypeNotRealOrBool), pk("accuracy", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "hll")
	a.HllIf = aggregateWithPrefix(aggregateScalar("hll_if", ScalarTypes.DynamicArray.Type(), pk("expr", ParameterTypeNotRealOrBool), p("predicate", ScalarTypes.Bool), pk("accuracy", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "hll_if")
	a.HllMerge = aggregateWithPrefix(aggregateScalar("hll_merge", ScalarTypes.DynamicArray.Type(), ap("hll", ScalarTypes.DynamicArray.Type())), "hll_merge")
	a.Min = aggregateWithPrefix(aggregateReturnKind("min", ReturnTypeParameter0, pk("expr", ParameterTypeOrderable)), "min")
	a.MinIf = aggregateWithPrefix(aggregateReturnKind("minif", ReturnTypeParameter0, pk("expr", ParameterTypeOrderable), p("predicate", ScalarTypes.Bool)), "minif")
	a.Max = aggregateWithPrefix(aggregateReturnKind("max", ReturnTypeParameter0, pk("expr", ParameterTypeOrderable)), "max")
	a.MaxIf = aggregateWithPrefix(aggregateReturnKind("maxif", ReturnTypeParameter0, pk("expr", ParameterTypeOrderable), p("predicate", ScalarTypes.Bool)), "maxif")
	a.Avg = aggregateWithPrefix(NewFunctionSymbol("avg",
		NewSignature(ScalarTypes.Real.Type(), pk("expr", ParameterTypeInteger)),
		NewSignature(ScalarTypes.Real.Type(), p("expr", ScalarTypes.Real)),
		NewSignature(ScalarTypes.Decimal.Type(), p("expr", ScalarTypes.Decimal)),
		NewSignature(ScalarTypes.TimeSpan.Type(), p("expr", ScalarTypes.TimeSpan)),
		NewSignature(ScalarTypes.DateTime.Type(), p("expr", ScalarTypes.DateTime)),
	), "avg")
	a.AvgIf = aggregateWithPrefix(NewFunctionSymbol("avgif",
		NewSignature(ScalarTypes.Real.Type(), pk("expr", ParameterTypeInteger), p("predicate", ScalarTypes.Bool)),
		NewSignature(ScalarTypes.Real.Type(), p("expr", ScalarTypes.Real), p("predicate", ScalarTypes.Bool)),
		NewSignature(ScalarTypes.Decimal.Type(), p("expr", ScalarTypes.Decimal), p("predicate", ScalarTypes.Bool)),
		NewSignature(ScalarTypes.TimeSpan.Type(), p("expr", ScalarTypes.TimeSpan), p("predicate", ScalarTypes.Bool)),
		NewSignature(ScalarTypes.DateTime.Type(), p("expr", ScalarTypes.DateTime), p("predicate", ScalarTypes.Bool)),
	), "avgif")

	a.MakeListDeprecated = aggregateWithPrefix(aggregateReturnKind("makelist", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "list").Obsolete("make_list").Hide()
	a.MakeList = aggregateWithPrefix(aggregateReturnKind("make_list", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "list")
	a.MakeListIf = aggregateWithPrefix(aggregateReturnKind("make_list_if", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), p("predicate", ScalarTypes.Bool), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "list")
	a.MakeListWithNulls = aggregateWithPrefix(aggregateReturnKind("make_list_with_nulls", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar)), "list")
	a.MakeSetDeprecated = aggregateWithPrefix(aggregateReturnKind("makeset", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "set").Obsolete("make_set").Hide()
	a.MakeSet = aggregateWithPrefix(aggregateReturnKind("make_set", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "set")
	a.MakeSetIf = aggregateWithPrefix(aggregateReturnKind("make_set_if", ReturnTypeParameter0Array, pk("expr", ParameterTypeScalar), p("predicate", ScalarTypes.Bool), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "set")
	a.Passthrough = aggregateWithFirstArgument(aggregateReturnKind("passthrough", ReturnTypeParameter0, pk("expr", ParameterTypeScalar))).Hide()
	a.MakeDictionary = aggregateWithPrefix(aggregateScalar("make_dictionary", ScalarTypes.DynamicBag.Type(), ap("expr", ScalarTypes.DynamicBag.Type()), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "dictionary").Hide()
	a.MakeBag = aggregateWithPrefix(aggregateScalar("make_bag", ScalarTypes.DynamicBag.Type(), ap("expr", ScalarTypes.DynamicBag.Type()), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "bag")
	a.MakeBagIf = aggregateWithPrefix(aggregateScalar("make_bag_if", ScalarTypes.DynamicBag.Type(), ap("expr", ScalarTypes.DynamicBag.Type()), p("predicate", ScalarTypes.Bool), pk("maxSize", ParameterTypeInteger, WithParameterOccurrence(0, 1))), "bag")
	a.BuildSchema = aggregateWithPrefix(aggregateScalar("buildschema", ScalarTypes.DynamicBag.Type(), ap("expr", ScalarTypes.DynamicBag.Type())), "schema")
	a.BinaryAllOr = aggregateWithFirstArgument(NewFunctionSymbol("binary_all_or", NewReturnKindSignature(ReturnTypeParameter0, pk("expr", ParameterTypeSummable))))
	a.BinaryAllAnd = aggregateWithFirstArgument(NewFunctionSymbol("binary_all_and", NewReturnKindSignature(ReturnTypeParameter0, pk("expr", ParameterTypeSummable))))
	a.BinaryAllXor = aggregateWithFirstArgument(NewFunctionSymbol("binary_all_xor", NewReturnKindSignature(ReturnTypeParameter0, pk("expr", ParameterTypeSummable))))
	a.CountDistinct = aggregateWithPrefix(aggregateScalar("count_distinct", ScalarTypes.Long.Type(), pk("expr", ParameterTypeNotDynamic)), "count_distinct").WithOptimizedAlternative("dcount")
	a.CountDistinctIf = aggregateWithPrefix(aggregateScalar("count_distinctif", ScalarTypes.Long.Type(), pk("expr", ParameterTypeNotDynamic), p("predicate", ScalarTypes.Bool)), "count_distinctif").WithOptimizedAlternative("dcountif")

	a.Percentile = NewFunctionSymbol("percentile", NewCustomSignature(percentileReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("percentile", ParameterTypeNumber)))
	a.Percentiles = NewFunctionSymbol("percentiles", NewCustomSignature(percentileReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("percentile", ParameterTypeNumber, WithParameterOccurrence(1, MaxRepeat))))
	a.PercentilesArray = NewFunctionSymbol("percentiles_array",
		NewCustomSignature(percentileArrayReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("percentile", ParameterTypeNumber, WithParameterOccurrence(1, MaxRepeat))),
		NewCustomSignature(percentileArrayReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("percentiles", ParameterTypeDynamicArray)),
	)
	a.PercentileW = NewFunctionSymbol("percentilew", NewCustomSignature(percentileReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("weight", ParameterTypeInteger), pk("percentile", ParameterTypeNumber)))
	a.PercentilesW = NewFunctionSymbol("percentilesw", NewCustomSignature(percentileReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("weight", ParameterTypeInteger), pk("percentile", ParameterTypeNumber, WithParameterOccurrence(1, MaxRepeat))))
	a.PercentilesWArray = NewFunctionSymbol("percentilesw_array",
		NewCustomSignature(percentileArrayReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("weight", ParameterTypeInteger), pk("percentile", ParameterTypeNumber, WithParameterOccurrence(1, MaxRepeat))),
		NewCustomSignature(percentileArrayReturn, TabularityScalar, pk("expr", ParameterTypeScalar), pk("weight", ParameterTypeInteger), pk("percentiles", ParameterTypeDynamicArray)),
	)
	a.Stdev = aggregateWithPrefix(aggregateScalar("stdev", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable)), "stdev")
	a.StdevIf = aggregateWithPrefix(aggregateScalar("stdevif", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "stdevif")
	a.Stdevp = aggregateWithPrefix(aggregateScalar("stdevp", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable)), "stdevp")
	a.Variance = aggregateWithPrefix(aggregateScalar("variance", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable)), "variance")
	a.VarianceIf = aggregateWithPrefix(aggregateScalar("varianceif", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "varianceif")
	a.Variancep = aggregateWithPrefix(aggregateScalar("variancep", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable)), "variancep")
	a.VariancepIf = aggregateWithPrefix(aggregateScalar("variancepif", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "variancepif")
	a.Covariance = aggregateWithPrefix(aggregateScalar("covariance", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), pk("expr", ParameterTypeSummable)), "covariance")
	a.CovarianceIf = aggregateWithPrefix(aggregateScalar("covarianceif", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "covarianceif")
	a.Covariancep = aggregateWithPrefix(aggregateScalar("covariancep", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), pk("expr", ParameterTypeSummable)), "covariancep")
	a.CovariancepIf = aggregateWithPrefix(aggregateScalar("covariancepif", ScalarTypes.Real.Type(), pk("expr", ParameterTypeSummable), pk("expr", ParameterTypeSummable), p("predicate", ScalarTypes.Bool)), "covariancepif")
	a.Any = aggregateWithPrefix(NewFunctionSymbol("any",
		NewReturnKindSignature(ReturnTypeParameter0, pk("expr", ParameterTypeScalar)),
		NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
			return getAnyResult(context, "")
		}, TabularityScalar, pk("expr", ParameterTypeScalar, WithParameterOccurrence(2, MaxRepeat))),
		NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
			return getAnyResult(context, "")
		}, TabularityScalar, pk("expr", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarOnly))),
	), "any").Obsolete("take_any")
	a.TakeAny = NewFunctionSymbol("take_any",
		NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
			return getAnyResult(context, "any_")
		}, TabularityScalar, pk("expr", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarAllowed), WithParameterOccurrence(1, MaxRepeat))),
	)
	a.AnyIf = aggregateWithPrefix(aggregateReturnKind("anyif", ReturnTypeParameter0, pk("expr", ParameterTypeScalar), p("predicate", ScalarTypes.Bool)), "anyif").Obsolete("take_anyif")
	a.TakeAnyIf = aggregateWithFirstArgument(aggregateReturnKind("take_anyif", ReturnTypeParameter0, pk("expr", ParameterTypeScalar), p("predicate", ScalarTypes.Bool)))
	a.ArgMin = NewFunctionSymbol("arg_min",
		NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
			return getArgMinMaxResult(context, "min")
		}, TabularityScalar, pk("minimized", ParameterTypeOrderable), pk("returned", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarAllowed), WithParameterOccurrence(0, MaxRepeat))),
	)
	a.ArgMax = NewFunctionSymbol("arg_max",
		NewCustomSignature(func(context CustomReturnTypeContext) Symbol {
			return getArgMinMaxResult(context, "max")
		}, TabularityScalar, pk("maximized", ParameterTypeOrderable), pk("returned", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarAllowed), WithParameterOccurrence(0, MaxRepeat))),
	)
	a.ArgMinDeprecated = NewFunctionSymbol("argmin",
		NewCustomSignature(getArgMinMaxDeprecatedResult, TabularityScalar, pk("minimized", ParameterTypeOrderable), pk("returned", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarAllowed), WithParameterOccurrence(0, MaxRepeat))),
	).WithResultNamePrefix("min").Obsolete("arg_min").Hide()
	a.ArgMaxDeprecated = NewFunctionSymbol("argmax",
		NewCustomSignature(getArgMinMaxDeprecatedResult, TabularityScalar, pk("maximized", ParameterTypeOrderable), pk("returned", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarAllowed), WithParameterOccurrence(0, MaxRepeat))),
	).WithResultNamePrefix("max").Obsolete("arg_max").Hide()

	a.All = []*FunctionSymbol{
		a.Sum, a.SumIf, a.Cnt, a.Count, a.CountIf, a.DCount, a.DCountIf,
		a.TDigest, a.TDigestMerge, a.MergeTDigest, a.Hll, a.HllIf, a.HllMerge,
		a.Min, a.MinIf, a.Max, a.MaxIf, a.Avg, a.AvgIf,
		a.MakeListDeprecated, a.MakeList, a.MakeListIf, a.MakeListWithNulls,
		a.MakeSetDeprecated, a.MakeSet, a.MakeSetIf, a.MakeDictionary,
		a.MakeBag, a.MakeBagIf, a.BuildSchema, a.Passthrough,
		a.Percentile, a.Percentiles, a.PercentilesArray, a.PercentileW,
		a.PercentilesW, a.PercentilesWArray,
		a.Stdev, a.StdevIf, a.Stdevp, a.Variance, a.VarianceIf, a.Variancep,
		a.VariancepIf, a.Covariance, a.CovarianceIf, a.Covariancep, a.CovariancepIf,
		a.Any, a.TakeAny, a.AnyIf, a.TakeAnyIf, a.ArgMin, a.ArgMax,
		a.ArgMinDeprecated, a.ArgMaxDeprecated, a.BinaryAllOr, a.BinaryAllAnd,
		a.BinaryAllXor, a.CountDistinct, a.CountDistinctIf,
	}
	return a
}

func aggregateScalar(name string, returnType *TypeSymbol, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewSignature(returnType, parameters...))
}

func aggregateReturnKind(name string, returnKind ReturnTypeKind, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewReturnKindSignature(returnKind, parameters...))
}

func aggregateWithPrefix(fn *FunctionSymbol, prefix string) *FunctionSymbol {
	return fn.WithResultNameKind(ResultNamePrefixAndFirstArgument).WithResultNamePrefix(prefix)
}

func aggregateWithFirstArgument(fn *FunctionSymbol) *FunctionSymbol {
	return fn.WithResultNameKind(ResultNameFirstArgument)
}

func ap(name string, typ *TypeSymbol, options ...ParameterOption) *Parameter {
	return NewParameter(name, typ, options...)
}

func percentileReturn(context CustomReturnTypeContext) Symbol {
	columns := addPercentileColumns(nil, context, "expr", "percentile")
	return NewTupleSymbol(columns)
}

func percentileArrayReturn(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	if valueArg := context.GetArgument("expr"); valueArg != nil {
		valueName := context.GetResultName(valueArg)
		columns = append(columns, NewColumnSymbolWithSource("percentiles_"+valueName, ScalarTypes.DynamicArray.Type(), valueArg))
	}
	return NewTupleSymbol(columns)
}

func addPercentileColumns(columns []*ColumnSymbol, context CustomReturnTypeContext, valueParameterName string, percentileParameterName string) []*ColumnSymbol {
	valueArg := context.GetArgument(valueParameterName)
	if valueArg == nil {
		return columns
	}

	resultType := promotedPercentileType(argumentType(context, valueArg))
	valueName := context.GetResultName(valueArg)
	for _, percentileArg := range context.GetArguments(percentileParameterName) {
		fragment := MakeValidNameFragment(GetConstantValue(percentileArg, context))
		name := percentileParameterName + "_" + valueName + "_" + fragment
		columns = append(columns, customAggregateResultColumn(context, name, resultType, valueArg, nil))
	}
	return columns
}

func argumentType(context CustomReturnTypeContext, argument Expression) *TypeSymbol {
	for i, candidate := range context.Arguments {
		if candidate == argument && i < len(context.ArgumentTypes) && context.ArgumentTypes[i] != nil {
			return context.ArgumentTypes[i]
		}
	}
	return ScalarTypes.Unknown.Type()
}

func promotedPercentileType(typ *TypeSymbol) *TypeSymbol {
	switch {
	case sameType(typ, ScalarTypes.Int.Type()):
		return ScalarTypes.Long.Type()
	case sameType(typ, ScalarTypes.Decimal.Type()):
		return ScalarTypes.Real.Type()
	case typ == nil:
		return ScalarTypes.Unknown.Type()
	default:
		return typ
	}
}

func getAnyResult(context CustomReturnTypeContext, unnamedExpressionPrefix string) Symbol {
	columns := []*ColumnSymbol{}
	doNotRepeat := columnSet(context.SummarizeByColumns)
	anyStar := containsStarExpression(context.Arguments)

	for i, arg := range context.Arguments {
		if SyntaxIsMissing(arg) {
			typ := ScalarTypes.Unknown.Type()
			columns = append(columns, customAggregateResultColumn(context, unnamedExpressionPrefix+"arg"+decimalString(i), typ, arg, nil))
			continue
		}
		if _, ok := arg.(*StarExpression); ok {
			for _, column := range rowScopeColumns(context) {
				if canAddAnyResultColumn(column, doNotRepeat, anyStar) {
					doNotRepeat[column] = true
					columns = append(columns, column)
				}
			}
			continue
		}

		if column := resultColumn(context, arg); column != nil {
			resultName := context.GetResultName(arg)
			if resultName != "" && resultName != column.SymbolName() {
				if canAddAnyResultColumn(column, doNotRepeat, anyStar) {
					doNotRepeat[column] = true
					columns = append(columns, clonedAggregateResultColumn(context, resultName, column.Type, arg, column))
				}
				continue
			}

			if canAddAnyResultColumn(column, doNotRepeat, anyStar) {
				if doNotRepeat[column] {
					columns = append(columns, clonedAggregateResultColumn(context, column.SymbolName(), column.Type, arg, column))
				} else {
					doNotRepeat[column] = true
					columns = append(columns, column)
				}
			}
			continue
		}

		expName := context.GetResultName(arg)
		if expName == "" {
			expName = unnamedExpressionPrefix + "arg" + decimalString(i)
		}
		columns = append(columns, customAggregateResultColumn(context, expName, argumentType(context, arg), arg, nil))
	}

	return NewTupleSymbol(columns)
}

// GetAnyResult exposes the source-level Aggregates.GetAnyResult helper.
func GetAnyResult(context CustomReturnTypeContext, unnamedExpressionPrefix string) Symbol {
	return getAnyResult(context, unnamedExpressionPrefix)
}

func canAddAnyResultColumn(column *ColumnSymbol, doNotRepeat map[*ColumnSymbol]bool, anyStar bool) bool {
	if !anyStar {
		return true
	}
	return column != nil && !doNotRepeat[column]
}

func getArgMinMaxResult(context CustomReturnTypeContext, prefix string) Symbol {
	columns := []*ColumnSymbol{}
	if len(context.Arguments) == 0 {
		return NewTupleSymbol(columns)
	}

	byClauseColumns := columnSet(context.SummarizeByColumns)
	doNotRepeat := map[*ColumnSymbol]bool{}
	primaryArg := context.Arguments[0]
	primaryColName := context.GetResultName(primaryArg)
	anyStar := containsStarExpression(context.Arguments)

	for i, arg := range context.Arguments {
		if SyntaxIsMissing(arg) {
			continue
		}
		if _, ok := arg.(*StarExpression); ok {
			for _, column := range rowScopeColumns(context) {
				if canAddArgMinMaxResultColumn(i, column, byClauseColumns, doNotRepeat, anyStar) {
					doNotRepeat[column] = true
					columns = append(columns, column)
				}
			}
			continue
		}

		if column := resultColumn(context, arg); column != nil {
			resultName := context.GetResultName(arg)
			if resultName != "" && resultName != column.SymbolName() {
				if canAddArgMinMaxResultColumn(i, column, byClauseColumns, doNotRepeat, anyStar) {
					doNotRepeat[column] = true
					columns = append(columns, clonedAggregateResultColumn(context, resultName, column.Type, arg, column))
				}
				continue
			}

			if canAddArgMinMaxResultColumn(i, column, byClauseColumns, doNotRepeat, anyStar) {
				if doNotRepeat[column] || byClauseColumns[column] {
					columns = append(columns, clonedAggregateResultColumn(context, column.SymbolName(), column.Type, arg, column))
				} else {
					doNotRepeat[column] = true
					columns = append(columns, column)
				}
			}
			continue
		}

		expName := context.GetResultName(arg)
		if expName == "" {
			if i == 0 {
				expName = prefix + "_"
			} else {
				expName = prefix + "_" + primaryColName + "_arg" + decimalString(i)
			}
		}
		columns = append(columns, customAggregateResultColumn(context, expName, argumentType(context, arg), arg, nil))
	}

	return NewTupleSymbol(columns)
}

func getArgMinMaxDeprecatedResult(context CustomReturnTypeContext) Symbol {
	columns := []*ColumnSymbol{}
	if len(context.Arguments) == 0 {
		return NewTupleSymbol(columns)
	}

	byClauseColumns := columnSet(context.SummarizeByColumns)
	doNotRepeat := map[*ColumnSymbol]bool{}
	anyStar := containsStarExpression(context.Arguments)
	primaryArg := context.Arguments[0]
	primaryColName := context.GetResultName(primaryArg)
	var primaryColumn *ColumnSymbol

	if column := resultColumn(context, primaryArg); column != nil {
		primaryColumn = column
		doNotRepeat[column] = true
		columns = append(columns, column)
		if primaryColName == "" {
			primaryColName = column.SymbolName()
		}
	} else {
		if primaryColName == "" {
			primaryColName = "arg0"
		}
		columns = append(columns, NewColumnSymbolWithSource(primaryColName, argumentType(context, primaryArg), primaryArg))
	}

	for i := 1; i < len(context.Arguments); i++ {
		arg := context.Arguments[i]
		if SyntaxIsMissing(arg) {
			continue
		}
		if _, ok := arg.(*StarExpression); ok {
			for _, column := range rowScopeColumns(context) {
				if column != primaryColumn && canAddArgMinMaxResultColumn(i, column, byClauseColumns, doNotRepeat, anyStar) {
					doNotRepeat[column] = true
					columns = append(columns, clonedAggregateResultColumn(context, primaryColName+"_"+column.SymbolName(), column.Type, arg, column))
				}
			}
			continue
		}

		if column := resultColumn(context, arg); column != nil {
			resultName := context.GetResultName(arg)
			if resultName == "" {
				resultName = column.SymbolName()
			}
			if canAddArgMinMaxResultColumn(i, column, byClauseColumns, doNotRepeat, anyStar) {
				doNotRepeat[column] = true
				columns = append(columns, clonedAggregateResultColumn(context, primaryColName+"_"+resultName, column.Type, arg, column))
			}
			continue
		}

		expName := context.GetResultName(arg)
		if expName == "" {
			expName = "arg" + decimalString(i)
		}
		columns = append(columns, customAggregateResultColumn(context, primaryColName+"_"+expName, argumentType(context, arg), arg, nil))
	}

	return NewTupleSymbol(columns)
}

func canAddArgMinMaxResultColumn(argIndex int, column *ColumnSymbol, byClauseColumns map[*ColumnSymbol]bool, doNotRepeat map[*ColumnSymbol]bool, anyStar bool) bool {
	if column == nil {
		return false
	}
	if argIndex == 0 {
		return true
	}
	if !anyStar {
		return true
	}
	return !doNotRepeat[column] && !byClauseColumns[column]
}

func resultColumn(context CustomReturnTypeContext, argument Expression) *ColumnSymbol {
	for i, candidate := range context.Arguments {
		if candidate == argument && i < len(context.ArgumentSemantics) {
			if column, ok := context.ArgumentSemantics[i].ReferencedSymbol().(*ColumnSymbol); ok {
				return column
			}
		}
	}
	return nil
}

func rowScopeColumns(context CustomReturnTypeContext) []*ColumnSymbol {
	if context.RowScope == nil {
		return nil
	}
	return context.RowScope.Columns()
}

func columnSet(columns []*ColumnSymbol) map[*ColumnSymbol]bool {
	set := map[*ColumnSymbol]bool{}
	for _, column := range columns {
		if column != nil {
			set[column] = true
		}
	}
	return set
}

func containsStarExpression(arguments []Expression) bool {
	for _, argument := range arguments {
		if _, ok := argument.(*StarExpression); ok {
			return true
		}
	}
	return false
}

func clonedAggregateResultColumn(context CustomReturnTypeContext, name string, typ *TypeSymbol, source SyntaxElement, original *ColumnSymbol) *ColumnSymbol {
	return customAggregateResultColumn(context, name, typ, source, []*ColumnSymbol{original})
}

func customAggregateResultColumn(context CustomReturnTypeContext, name string, typ *TypeSymbol, source SyntaxElement, original []*ColumnSymbol) *ColumnSymbol {
	return NewColumnSymbolFull(name, typ, "", original, source, nil)
}

func decimalString(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
