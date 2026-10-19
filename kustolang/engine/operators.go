package kustolang

// Ported from src/Kusto.Language/Operators.cs.

type operatorSet struct {
	UnaryMinus *OperatorSymbol
	UnaryPlus  *OperatorSymbol

	And                *OperatorSymbol
	Or                 *OperatorSymbol
	Add                *OperatorSymbol
	Subtract           *OperatorSymbol
	Multiply           *OperatorSymbol
	Divide             *OperatorSymbol
	Modulo             *OperatorSymbol
	LessThan           *OperatorSymbol
	LessThanOrEqual    *OperatorSymbol
	GreaterThan        *OperatorSymbol
	GreaterThanOrEqual *OperatorSymbol
	Equal              *OperatorSymbol
	NotEqual           *OperatorSymbol

	EqualTilde      *OperatorSymbol
	BangTilde       *OperatorSymbol
	Has             *OperatorSymbol
	HasCs           *OperatorSymbol
	NotHas          *OperatorSymbol
	NotHasCs        *OperatorSymbol
	HasPrefix       *OperatorSymbol
	HasPrefixCs     *OperatorSymbol
	NotHasPrefix    *OperatorSymbol
	NotHasPrefixCs  *OperatorSymbol
	HasSuffix       *OperatorSymbol
	HasSuffixCs     *OperatorSymbol
	NotHasSuffix    *OperatorSymbol
	NotHasSuffixCs  *OperatorSymbol
	Like            *OperatorSymbol
	LikeCs          *OperatorSymbol
	NotLike         *OperatorSymbol
	NotLikeCs       *OperatorSymbol
	Contains        *OperatorSymbol
	ContainsCs      *OperatorSymbol
	NotContains     *OperatorSymbol
	NotContainsCs   *OperatorSymbol
	StartsWith      *OperatorSymbol
	StartsWithCs    *OperatorSymbol
	NotStartsWith   *OperatorSymbol
	NotStartsWithCs *OperatorSymbol
	EndsWith        *OperatorSymbol
	EndsWithCs      *OperatorSymbol
	NotEndsWith     *OperatorSymbol
	NotEndsWithCs   *OperatorSymbol
	MatchRegex      *OperatorSymbol
	Search          *OperatorSymbol
	In              *OperatorSymbol
	HasAny          *OperatorSymbol
	HasAll          *OperatorSymbol
	InCs            *OperatorSymbol
	NotIn           *OperatorSymbol
	NotInCs         *OperatorSymbol
	Between         *OperatorSymbol
	NotBetween      *OperatorSymbol

	All []*OperatorSymbol
}

var Operators = newOperatorSet()

func newOperatorSet() *operatorSet {
	ops := &operatorSet{}
	dateAndTimespan := typeSymbols(ScalarTypes.DateTime, ScalarTypes.TimeSpan)
	stringOrDynamic := typeSymbols(ScalarTypes.String, ScalarTypes.Dynamic)
	dynamicAddable := typeSymbols(
		ScalarTypes.Int,
		ScalarTypes.Long,
		ScalarTypes.Real,
		ScalarTypes.Decimal,
		ScalarTypes.TimeSpan,
		ScalarTypes.DateTime,
	)

	ops.UnaryMinus = operator(OperatorKindUnaryMinus,
		NewCustomSignature(unaryOperatorReturnType, TabularityScalar, pk("operand", ParameterTypeSummable)),
		sig(ScalarTypes.Dynamic, p("operand", ScalarTypes.Dynamic)),
	)
	ops.UnaryPlus = operator(OperatorKindUnaryPlus,
		NewCustomSignature(unaryOperatorReturnType, TabularityScalar, pk("operand", ParameterTypeSummable)),
		sig(ScalarTypes.Dynamic, p("operand", ScalarTypes.Dynamic)),
	)

	ops.And = operator(OperatorKindAnd,
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), p("right", ScalarTypes.Bool)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Bool)).Hide(),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), p("right", ScalarTypes.Dynamic)).Hide(),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)).Hide(),
	)
	ops.Or = operator(OperatorKindOr,
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), p("right", ScalarTypes.Bool)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Bool)).Hide(),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), p("right", ScalarTypes.Dynamic)).Hide(),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)).Hide(),
	)
	ops.Add = operator(OperatorKindAdd,
		returnSig(ReturnTypeWidest, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.DateTime, pt("left", dateAndTimespan), pt("right", dateAndTimespan)),
		sig(ScalarTypes.Dynamic, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Int)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Int), p("right", ScalarTypes.Dynamic)),
		returnSig(ReturnTypeParameter1, p("left", ScalarTypes.Dynamic), pt("right", dynamicAddable)),
		returnSig(ReturnTypeParameter0, pt("left", dynamicAddable), p("right", ScalarTypes.Dynamic)),
	)
	ops.Subtract = operator(OperatorKindSubtract,
		returnSig(ReturnTypeWidest, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.DateTime, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.DateTime, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.Dynamic, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Int)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Int), p("right", ScalarTypes.Dynamic)),
		returnSig(ReturnTypeParameter1, p("left", ScalarTypes.Dynamic), pt("right", dynamicAddable)),
		returnSig(ReturnTypeParameter0, pt("left", dynamicAddable), p("right", ScalarTypes.Dynamic)),
	)
	ops.Multiply = operator(OperatorKindMultiply,
		returnSig(ReturnTypeWidest, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, pk("left", ParameterTypeNumber), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.DateTime, p("left", ScalarTypes.DateTime), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.DateTime, pk("left", ParameterTypeNumber), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.Dynamic, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Int)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Int), p("right", ScalarTypes.Dynamic)),
		returnSig(ReturnTypeParameter1, p("left", ScalarTypes.Dynamic), pk("right", ParameterTypeNumber)),
		returnSig(ReturnTypeParameter0, pk("left", ParameterTypeNumber), p("right", ScalarTypes.Dynamic)),
	)
	ops.Divide = operator(OperatorKindDivide,
		returnSig(ReturnTypeWidest, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.Real, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.Real, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.Real, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.DateTime, p("left", ScalarTypes.DateTime), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.Dynamic, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Int)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Int), p("right", ScalarTypes.Dynamic)),
		returnSig(ReturnTypeParameter1, p("left", ScalarTypes.Dynamic), pk("right", ParameterTypeNumber)),
		returnSig(ReturnTypeParameter0, pk("left", ParameterTypeNumber), p("right", ScalarTypes.Dynamic)),
	)
	ops.Modulo = operator(OperatorKindModulo,
		returnSig(ReturnTypeWidest, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.TimeSpan), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.TimeSpan, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.DateTime, p("left", ScalarTypes.DateTime), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.Dynamic, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Int)),
		sig(ScalarTypes.Long, p("left", ScalarTypes.Int), p("right", ScalarTypes.Dynamic)),
		returnSig(ReturnTypeParameter1, p("left", ScalarTypes.Dynamic), pk("right", ParameterTypeNumber)),
		returnSig(ReturnTypeParameter0, pk("left", ParameterTypeNumber), p("right", ScalarTypes.Dynamic)),
	)

	ops.LessThan = comparisonOperator(OperatorKindLessThan)
	ops.LessThanOrEqual = comparisonOperator(OperatorKindLessThanOrEqual)
	ops.GreaterThan = comparisonOperator(OperatorKindGreaterThan)
	ops.GreaterThanOrEqual = comparisonOperator(OperatorKindGreaterThanOrEqual)
	ops.Equal = operator(OperatorKindEqual,
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), pk("right", ParameterTypeScalar)).Hide(),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeNotBool), pk("right", ParameterTypeScalar)),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeNotBool, WithArgumentKind(ArgumentKindStarOnly)), pk("right", ParameterTypeScalar)),
	)
	ops.NotEqual = operator(OperatorKindNotEqual,
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), pk("right", ParameterTypeScalar)).Hide(),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeNotBool), pk("right", ParameterTypeScalar)),
	)

	ops.EqualTilde = stringBinary(OperatorKindEqualTilde, true)
	ops.BangTilde = stringBinary(OperatorKindBangTilde, true)
	ops.Has = stringBinary(OperatorKindHas, false)
	ops.HasCs = stringBinary(OperatorKindHasCs, false)
	ops.NotHas = stringBinary(OperatorKindNotHas, false)
	ops.NotHasCs = stringBinary(OperatorKindNotHasCs, false)
	ops.HasPrefix = stringBinary(OperatorKindHasPrefix, false)
	ops.HasPrefixCs = stringBinary(OperatorKindHasPrefixCs, false)
	ops.NotHasPrefix = stringBinary(OperatorKindNotHasPrefix, false)
	ops.NotHasPrefixCs = stringBinary(OperatorKindNotHasPrefixCs, false)
	ops.HasSuffix = stringBinary(OperatorKindHasSuffix, false)
	ops.HasSuffixCs = stringBinary(OperatorKindHasSuffixCs, false)
	ops.NotHasSuffix = stringBinary(OperatorKindNotHasSuffix, false)
	ops.NotHasSuffixCs = stringBinary(OperatorKindNotHasSuffixCs, false)
	ops.Like = stringBinary(OperatorKindLike, true)
	ops.LikeCs = stringBinary(OperatorKindLikeCs, true)
	ops.NotLike = stringBinary(OperatorKindNotLike, true)
	ops.NotLikeCs = stringBinary(OperatorKindNotLikeCs, true)
	ops.Contains = stringBinary(OperatorKindContains, true)
	ops.ContainsCs = stringBinary(OperatorKindContainsCs, true)
	ops.NotContains = stringBinary(OperatorKindNotContains, true)
	ops.NotContainsCs = stringBinary(OperatorKindNotContainsCs, true)
	ops.StartsWith = stringBinary(OperatorKindStartsWith, true)
	ops.StartsWithCs = stringBinary(OperatorKindStartsWithCs, true)
	ops.NotStartsWith = stringBinary(OperatorKindNotStartsWith, true)
	ops.NotStartsWithCs = stringBinary(OperatorKindNotStartsWithCs, true)
	ops.EndsWith = stringBinary(OperatorKindEndsWith, true)
	ops.EndsWithCs = stringBinary(OperatorKindEndsWithCs, true)
	ops.NotEndsWith = stringBinary(OperatorKindNotEndsWith, true)
	ops.NotEndsWithCs = stringBinary(OperatorKindNotEndsWithCs, true)
	ops.MatchRegex = stringBinary(OperatorKindMatchRegex, true)
	ops.Search = stringBinary(OperatorKindSearch, true)

	ops.In = operator(OperatorKindIn,
		sig(ScalarTypes.Bool, p("value", ScalarTypes.Bool), pk("table", ParameterTypeTabular)).Hide(),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNotBool), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, p("value", ScalarTypes.Bool), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))).Hide(),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNotBool), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.HasAny = operator(OperatorKindHasAny,
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.HasAll = operator(OperatorKindHasAll,
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.InCs = operator(OperatorKindInCs,
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.NotIn = operator(OperatorKindNotIn,
		sig(ScalarTypes.Bool, p("value", ScalarTypes.Bool), pk("table", ParameterTypeTabular)).Hide(),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNotBool), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, p("value", ScalarTypes.Bool), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))).Hide(),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNotBool), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.NotInCs = operator(OperatorKindNotInCs,
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("table", ParameterTypeTabular)),
		sig(ScalarTypes.Bool, pt("value", stringOrDynamic), pk("value", ParameterTypeScalar, WithParameterOccurrence(1, MaxRepeat))),
	)
	ops.Between = betweenOperator(OperatorKindBetween)
	ops.NotBetween = betweenOperator(OperatorKindNotBetween)

	ops.All = []*OperatorSymbol{
		ops.UnaryMinus,
		ops.UnaryPlus,
		ops.And,
		ops.Or,
		ops.Add,
		ops.Subtract,
		ops.Multiply,
		ops.Divide,
		ops.Modulo,
		ops.LessThan,
		ops.LessThanOrEqual,
		ops.GreaterThan,
		ops.GreaterThanOrEqual,
		ops.Equal,
		ops.NotEqual,
		ops.EqualTilde,
		ops.BangTilde,
		ops.Has,
		ops.HasCs,
		ops.NotHas,
		ops.NotHasCs,
		ops.HasPrefix,
		ops.HasPrefixCs,
		ops.NotHasPrefix,
		ops.NotHasPrefixCs,
		ops.HasSuffix,
		ops.HasSuffixCs,
		ops.NotHasSuffix,
		ops.NotHasSuffixCs,
		ops.Like,
		ops.LikeCs,
		ops.NotLike,
		ops.NotLikeCs,
		ops.Contains,
		ops.ContainsCs,
		ops.NotContains,
		ops.NotContainsCs,
		ops.StartsWith,
		ops.StartsWithCs,
		ops.NotStartsWith,
		ops.NotStartsWithCs,
		ops.EndsWith,
		ops.EndsWithCs,
		ops.NotEndsWith,
		ops.NotEndsWithCs,
		ops.MatchRegex,
		ops.Search,
		ops.In,
		ops.InCs,
		ops.NotIn,
		ops.NotInCs,
		ops.Between,
		ops.NotBetween,
		ops.HasAny,
		ops.HasAll,
	}
	return ops
}

func stringBinary(kind OperatorKind, dynamicRHS bool) *OperatorSymbol {
	if dynamicRHS {
		return operator(kind,
			sig(ScalarTypes.Bool, pk("left", ParameterTypeStringOrDynamic), pk("right", ParameterTypeStringOrDynamic)),
			sig(ScalarTypes.Bool, pk("left", ParameterTypeStringOrDynamic, WithArgumentKind(ArgumentKindStarOnly)), pk("right", ParameterTypeStringOrDynamic)),
		)
	}
	return operator(kind,
		sig(ScalarTypes.Bool, pk("left", ParameterTypeStringOrDynamic), p("right", ScalarTypes.String)),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeStringOrDynamic, WithArgumentKind(ArgumentKindStarOnly)), p("right", ScalarTypes.String)),
	)
}

func comparisonOperator(kind OperatorKind) *OperatorSymbol {
	return operator(kind,
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Bool), pk("right", ParameterTypeNumberOrBool)).Hide(),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeNumber), pk("right", ParameterTypeNumberOrBool)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.TimeSpan), p("right", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.DateTime), p("right", ScalarTypes.DateTime)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), p("right", ScalarTypes.Dynamic)),
		sig(ScalarTypes.Bool, p("left", ScalarTypes.Dynamic), pk("right", ParameterTypeNumber)),
		sig(ScalarTypes.Bool, pk("left", ParameterTypeNumber), p("right", ScalarTypes.Dynamic)),
	)
}

func betweenOperator(kind OperatorKind) *OperatorSymbol {
	return operator(kind,
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNumber), pk("start", ParameterTypeNumber), pk("end", ParameterTypeNumber)),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeSummable), pk("start", ParameterTypeParameter0), pk("end", ParameterTypeParameter0)),
		sig(ScalarTypes.Bool, p("value", ScalarTypes.DateTime), p("start", ScalarTypes.DateTime), p("end", ScalarTypes.TimeSpan)),
		sig(ScalarTypes.Bool, pk("value", ParameterTypeNumber), p("start", ScalarTypes.Dynamic), pk("end", ParameterTypeParameter1)),
		sig(ScalarTypes.Bool, p("value", ScalarTypes.Dynamic), pk("start", ParameterTypeNumber), pk("end", ParameterTypeParameter1)),
	)
}

func unaryOperatorReturnType(context CustomReturnTypeContext) Symbol {
	if len(context.ArgumentTypes) != 1 || context.ArgumentTypes[0] == nil {
		return ScalarTypes.Unknown.Type()
	}
	if context.ArgumentTypes[0] == ScalarTypes.Int.Type() && !argumentIsConstant(context, 0) {
		return ScalarTypes.Long.Type()
	}
	return context.ArgumentTypes[0]
}

func argumentIsConstant(context CustomReturnTypeContext, index int) bool {
	return index >= 0 && index < len(context.ArgumentIsConstant) && context.ArgumentIsConstant[index]
}

func operator(kind OperatorKind, signatures ...*Signature) *OperatorSymbol {
	return NewOperatorSymbol(kind, signatures...)
}

func sig(returnType *ScalarSymbol, parameters ...*Parameter) *Signature {
	return NewSignature(returnType.Type(), parameters...)
}

func returnSig(returnKind ReturnTypeKind, parameters ...*Parameter) *Signature {
	return NewReturnKindSignature(returnKind, parameters...)
}

func p(name string, typ *ScalarSymbol, options ...ParameterOption) *Parameter {
	return NewParameter(name, typ.Type(), options...)
}

func pk(name string, typeKind ParameterTypeKind, options ...ParameterOption) *Parameter {
	return NewParameterKind(name, typeKind, options...)
}

func pt(name string, types []*TypeSymbol, options ...ParameterOption) *Parameter {
	return NewParameterTypes(name, types, options...)
}

func typeSymbols(types ...*ScalarSymbol) []*TypeSymbol {
	symbols := make([]*TypeSymbol, 0, len(types))
	for _, typ := range types {
		if typ == nil {
			symbols = append(symbols, ScalarTypes.Unknown.Type())
			continue
		}
		symbols = append(symbols, typ.Type())
	}
	return symbols
}
