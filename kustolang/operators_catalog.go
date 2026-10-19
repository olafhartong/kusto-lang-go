package kustolang

// Ported from src/Kusto.Language/Operators.cs. See third_party/kusto/LICENSE
// and third_party/kusto/UPSTREAM.json.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// OperatorCatalog contains the built-in scalar operator symbols used by the
// binder. All returns a defensive copy of the pinned upstream ordering.
type OperatorCatalog struct {
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
	all             []*OperatorSymbol
}

var Operators = newOperatorCatalog()

func newOperatorCatalog() *OperatorCatalog {
	source := internal.Operators
	return &OperatorCatalog{
		UnaryMinus:         source.UnaryMinus,
		UnaryPlus:          source.UnaryPlus,
		And:                source.And,
		Or:                 source.Or,
		Add:                source.Add,
		Subtract:           source.Subtract,
		Multiply:           source.Multiply,
		Divide:             source.Divide,
		Modulo:             source.Modulo,
		LessThan:           source.LessThan,
		LessThanOrEqual:    source.LessThanOrEqual,
		GreaterThan:        source.GreaterThan,
		GreaterThanOrEqual: source.GreaterThanOrEqual,
		Equal:              source.Equal,
		NotEqual:           source.NotEqual,
		EqualTilde:         source.EqualTilde,
		BangTilde:          source.BangTilde,
		Has:                source.Has,
		HasCs:              source.HasCs,
		NotHas:             source.NotHas,
		NotHasCs:           source.NotHasCs,
		HasPrefix:          source.HasPrefix,
		HasPrefixCs:        source.HasPrefixCs,
		NotHasPrefix:       source.NotHasPrefix,
		NotHasPrefixCs:     source.NotHasPrefixCs,
		HasSuffix:          source.HasSuffix,
		HasSuffixCs:        source.HasSuffixCs,
		NotHasSuffix:       source.NotHasSuffix,
		NotHasSuffixCs:     source.NotHasSuffixCs,
		Like:               source.Like,
		LikeCs:             source.LikeCs,
		NotLike:            source.NotLike,
		NotLikeCs:          source.NotLikeCs,
		Contains:           source.Contains,
		ContainsCs:         source.ContainsCs,
		NotContains:        source.NotContains,
		NotContainsCs:      source.NotContainsCs,
		StartsWith:         source.StartsWith,
		StartsWithCs:       source.StartsWithCs,
		NotStartsWith:      source.NotStartsWith,
		NotStartsWithCs:    source.NotStartsWithCs,
		EndsWith:           source.EndsWith,
		EndsWithCs:         source.EndsWithCs,
		NotEndsWith:        source.NotEndsWith,
		NotEndsWithCs:      source.NotEndsWithCs,
		MatchRegex:         source.MatchRegex,
		Search:             source.Search,
		In:                 source.In,
		HasAny:             source.HasAny,
		HasAll:             source.HasAll,
		InCs:               source.InCs,
		NotIn:              source.NotIn,
		NotInCs:            source.NotInCs,
		Between:            source.Between,
		NotBetween:         source.NotBetween,
		all:                append([]*OperatorSymbol{}, source.All...),
	}
}

func (c *OperatorCatalog) All() []*OperatorSymbol {
	if c == nil {
		return nil
	}
	return append([]*OperatorSymbol{}, c.all...)
}
