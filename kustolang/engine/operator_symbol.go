package kustolang

// Ported from src/Kusto.Language/Symbols/OperatorKind.cs and OperatorSymbol.cs.

type OperatorKind string

const (
	OperatorKindNone OperatorKind = "None"

	OperatorKindUnaryMinus OperatorKind = "UnaryMinus"
	OperatorKindUnaryPlus  OperatorKind = "UnaryPlus"

	OperatorKindAdd      OperatorKind = "Add"
	OperatorKindSubtract OperatorKind = "Subtract"
	OperatorKindMultiply OperatorKind = "Multiply"
	OperatorKindDivide   OperatorKind = "Divide"
	OperatorKindModulo   OperatorKind = "Modulo"

	OperatorKindEqualTilde         OperatorKind = "EqualTilde"
	OperatorKindBangTilde          OperatorKind = "BangTilde"
	OperatorKindHas                OperatorKind = "Has"
	OperatorKindHasCs              OperatorKind = "HasCs"
	OperatorKindNotHas             OperatorKind = "NotHas"
	OperatorKindNotHasCs           OperatorKind = "NotHasCs"
	OperatorKindHasPrefix          OperatorKind = "HasPrefix"
	OperatorKindHasPrefixCs        OperatorKind = "HasPrefixCs"
	OperatorKindNotHasPrefix       OperatorKind = "NotHasPrefix"
	OperatorKindNotHasPrefixCs     OperatorKind = "NotHasPrefixCs"
	OperatorKindHasSuffix          OperatorKind = "HasSuffix"
	OperatorKindHasSuffixCs        OperatorKind = "HasSuffixCs"
	OperatorKindNotHasSuffix       OperatorKind = "NotHasSuffix"
	OperatorKindNotHasSuffixCs     OperatorKind = "NotHasSuffixCs"
	OperatorKindLike               OperatorKind = "Like"
	OperatorKindLikeCs             OperatorKind = "LikeCs"
	OperatorKindNotLike            OperatorKind = "NotLike"
	OperatorKindNotLikeCs          OperatorKind = "NotLikeCs"
	OperatorKindContains           OperatorKind = "Contains"
	OperatorKindContainsCs         OperatorKind = "ContainsCs"
	OperatorKindNotContains        OperatorKind = "NotContains"
	OperatorKindNotContainsCs      OperatorKind = "NotContainsCs"
	OperatorKindStartsWith         OperatorKind = "StartsWith"
	OperatorKindStartsWithCs       OperatorKind = "StartsWithCs"
	OperatorKindNotStartsWith      OperatorKind = "NotStartsWith"
	OperatorKindNotStartsWithCs    OperatorKind = "NotStartsWithCs"
	OperatorKindEndsWith           OperatorKind = "EndsWith"
	OperatorKindEndsWithCs         OperatorKind = "EndsWithCs"
	OperatorKindNotEndsWith        OperatorKind = "NotEndsWith"
	OperatorKindNotEndsWithCs      OperatorKind = "NotEndsWithCs"
	OperatorKindMatchRegex         OperatorKind = "MatchRegex"
	OperatorKindSearch             OperatorKind = "Search"
	OperatorKindLessThan           OperatorKind = "LessThan"
	OperatorKindLessThanOrEqual    OperatorKind = "LessThanOrEqual"
	OperatorKindGreaterThan        OperatorKind = "GreaterThan"
	OperatorKindGreaterThanOrEqual OperatorKind = "GreaterThanOrEqual"
	OperatorKindEqual              OperatorKind = "Equal"
	OperatorKindNotEqual           OperatorKind = "NotEqual"
	OperatorKindAnd                OperatorKind = "And"
	OperatorKindOr                 OperatorKind = "Or"
	OperatorKindIn                 OperatorKind = "In"
	OperatorKindInCs               OperatorKind = "InCs"
	OperatorKindNotIn              OperatorKind = "NotIn"
	OperatorKindNotInCs            OperatorKind = "NotInCs"
	OperatorKindBetween            OperatorKind = "Between"
	OperatorKindNotBetween         OperatorKind = "NotBetween"
	OperatorKindHasAny             OperatorKind = "HasAny"
	OperatorKindHasAll             OperatorKind = "HasAll"
)

type OperatorSymbol struct {
	BaseSymbol
	OperatorKind OperatorKind
	Result       *TypeSymbol
	signatures   []*Signature
}

func NewOperatorSymbol(kind OperatorKind, signatures ...*Signature) *OperatorSymbol {
	op := &OperatorSymbol{
		BaseSymbol:   NewBaseSymbol(string(kind)),
		OperatorKind: kind,
		signatures:   append([]*Signature{}, signatures...),
	}
	for _, signature := range op.signatures {
		if signature != nil {
			signature.Symbol = op
		}
	}
	if len(op.signatures) > 0 && op.signatures[0] != nil {
		op.Result = op.signatures[0].DeclaredReturnType
	}
	return op
}

func NewOperatorSymbolWithResult(kind OperatorKind, resultType *TypeSymbol) *OperatorSymbol {
	return NewOperatorSymbol(kind, NewSignature(resultType))
}

func (s *OperatorSymbol) SymbolKind() SymbolKind {
	return SymbolKindOperator
}

func (s *OperatorSymbol) SymbolTabularity() Tabularity {
	return TabularityScalar
}

func (s *OperatorSymbol) Signatures() []*Signature {
	if s == nil {
		return nil
	}
	return append([]*Signature{}, s.signatures...)
}
