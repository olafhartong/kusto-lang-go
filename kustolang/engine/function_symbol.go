package kustolang

// Ported from src/Kusto.Language/Symbols/FunctionSymbol.cs. This is the
// signature container and metadata shape used by GlobalState and local data.

type ResultNameKind string

const (
	ResultNameNone                       ResultNameKind = "None"
	ResultNameNameAndFirstArgument       ResultNameKind = "NameAndFirstArgument"
	ResultNameNameAndOnlyArgument        ResultNameKind = "NameAndOnlyArgument"
	ResultNamePrefixAndFirstArgument     ResultNameKind = "PrefixAndFirstArgument"
	ResultNamePrefixAndOnlyArgument      ResultNameKind = "PrefixAndOnlyArgument"
	ResultNameFirstArgumentValueIfColumn ResultNameKind = "FirstArgumentValueIfColumn"
	ResultNamePrefixOnly                 ResultNameKind = "PrefixOnly"
	ResultNameFirstArgument              ResultNameKind = "FirstArgument"
	ResultNameOnlyArgument               ResultNameKind = "OnlyArgument"

	ResultNameDefault = ResultNameNone
	// ResultNamePrefix is retained as a compatibility alias for the earlier Go port.
	ResultNamePrefix = ResultNamePrefixOnly
)

// CustomAvailability mirrors Kusto.Language.Symbols.CustomAvailability. The
// current parser does not maintain parent pointers yet, so callers can supply
// the graph-scope facts computed by a future binder through the context.
type CustomAvailability func(CustomAvailabilityContext) bool

type CustomAvailabilityContext struct {
	Location                      SyntaxElement
	InGraphWhereOrProjectClause   bool
	InGraphMatchOrShortestPaths   bool
	InGraphAllAnyMapFirstArgument bool
}

type FunctionSymbol struct {
	TypeSymbol
	signatures           []*Signature
	description          string
	resultNameKind       ResultNameKind
	resultNamePrefix     string
	alternative          string
	optimizedAlternative string
	customAvailability   CustomAvailability
	isHidden             bool
	isConstantFoldable   bool
	isView               bool
	isObsolete           bool
}

func NewFunctionSymbol(name string, signatures ...*Signature) *FunctionSymbol {
	fn := &FunctionSymbol{
		TypeSymbol:     TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		signatures:     append([]*Signature{}, signatures...),
		resultNameKind: ResultNameDefault,
	}
	for _, signature := range fn.signatures {
		if signature != nil {
			signature.Symbol = fn
		}
	}
	return fn
}

func NewScalarFunctionSymbol(name string, returnType *TypeSymbol, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewSignature(returnType, parameters...))
}

func NewComputedFunctionSymbol(name, body string, tabularity Tabularity, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewComputedSignature(body, tabularity, parameters...))
}

// NewReturnKindFunctionSymbol constructs a function whose return type is
// determined from its arguments according to returnKind.
func NewReturnKindFunctionSymbol(name string, returnKind ReturnTypeKind, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewReturnKindSignature(returnKind, parameters...))
}

// NewCustomFunctionSymbol constructs a function whose return symbol
// is computed by customReturnType.
func NewCustomFunctionSymbol(name string, customReturnType CustomReturnType, tabularity Tabularity, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewCustomSignature(customReturnType, tabularity, parameters...))
}

// NewFunctionSymbolFromParameterList constructs a computed function from a
// source parameter list and leaves its tabularity to be inferred from the body.
func NewFunctionSymbolFromParameterList(name, parameterList, body string, description ...string) *FunctionSymbol {
	return newFunctionSymbolWithOptionalDescription(
		name,
		NewComputedSignatureInferred(body, ParseParameterList(parameterList)...),
		description,
	)
}

// NewFunctionSymbolFromParameterListWithTabularity constructs a computed
// function from a source parameter list and an explicit tabularity.
func NewFunctionSymbolFromParameterListWithTabularity(name, parameterList, body string, tabularity Tabularity, description ...string) *FunctionSymbol {
	return newFunctionSymbolWithOptionalDescription(
		name,
		NewComputedSignature(body, tabularity, ParseParameterList(parameterList)...),
		description,
	)
}

// NewDeclarationFunctionSymbol constructs a computed function from an
// already parsed function body.
func NewDeclarationFunctionSymbol(name string, declaration *FunctionBody, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewDeclarationSignature(declaration, parameters...))
}

func newFunctionSymbolWithOptionalDescription(name string, signature *Signature, description []string) *FunctionSymbol {
	function := NewFunctionSymbol(name, signature)
	if len(description) > 0 {
		function = function.WithDescription(description[0])
	}
	return function
}

func (s *FunctionSymbol) SymbolKind() SymbolKind {
	return SymbolKindFunction
}

func (s *FunctionSymbol) Signatures() []*Signature {
	return append([]*Signature{}, s.signatures...)
}

func (s *FunctionSymbol) SymbolTabularity() Tabularity {
	if len(s.signatures) == 0 || s.signatures[0] == nil {
		return TabularityUnknown
	}
	return s.signatures[0].Tabularity
}

func (s *FunctionSymbol) Description() string {
	return s.description
}

func (s *FunctionSymbol) ResultNameKind() ResultNameKind {
	return s.resultNameKind
}

func (s *FunctionSymbol) ResultNamePrefix() string {
	return s.resultNamePrefix
}

func (s *FunctionSymbol) Alternative() string {
	return s.alternative
}

func (s *FunctionSymbol) IsObsolete() bool {
	return s != nil && s.isObsolete
}

func (s *FunctionSymbol) OptimizedAlternative() string {
	return s.optimizedAlternative
}

func (s *FunctionSymbol) CustomAvailability() CustomAvailability {
	return s.customAvailability
}

func (s *FunctionSymbol) IsAvailable(context CustomAvailabilityContext) bool {
	return s == nil || s.customAvailability == nil || s.customAvailability(context)
}

func (s *FunctionSymbol) IsHidden() bool {
	return s.isHidden || s.TypeSymbol.IsHidden()
}

func (s *FunctionSymbol) IsConstantFoldable() bool {
	return s.isConstantFoldable
}

func (s *FunctionSymbol) IsView() bool {
	return s.isView
}

func (s *FunctionSymbol) WithIsView(isView bool) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.isView == isView {
		return s
	}
	clone := s.clone()
	clone.isView = isView
	return clone
}

func (s *FunctionSymbol) WithIsHidden(isHidden bool) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.isHidden == isHidden {
		return s
	}
	clone := s.clone()
	clone.isHidden = isHidden
	return clone
}

func (s *FunctionSymbol) Hide() *FunctionSymbol {
	return s.WithIsHidden(true)
}

func (s *FunctionSymbol) WithIsConstantFoldable(isConstantFoldable bool) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.isConstantFoldable == isConstantFoldable {
		return s
	}
	clone := s.clone()
	clone.isConstantFoldable = isConstantFoldable
	return clone
}

func (s *FunctionSymbol) ConstantFoldable() *FunctionSymbol {
	return s.WithIsConstantFoldable(true)
}

func (s *FunctionSymbol) WithResultNamePrefix(resultNamePrefix string) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.resultNamePrefix == resultNamePrefix {
		return s
	}
	clone := s.clone()
	clone.resultNamePrefix = resultNamePrefix
	return clone
}

func (s *FunctionSymbol) WithResultNameKind(kind ResultNameKind) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.resultNameKind == kind {
		return s
	}
	clone := s.clone()
	clone.resultNameKind = kind
	return clone
}

func (s *FunctionSymbol) WithDescription(description string) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.description == description {
		return s
	}
	clone := s.clone()
	clone.description = description
	return clone
}

func (s *FunctionSymbol) WithIsObsolete(isObsolete bool, alternative ...string) *FunctionSymbol {
	if s == nil {
		return nil
	}
	alt := ""
	if len(alternative) > 0 {
		alt = alternative[0]
	}
	if s.isObsolete == isObsolete && s.alternative == alt {
		return s
	}
	clone := s.clone()
	clone.isObsolete = isObsolete
	clone.alternative = alt
	if !isObsolete {
		clone.alternative = ""
	}
	return clone
}

func (s *FunctionSymbol) Obsolete(alternative string) *FunctionSymbol {
	return s.WithIsObsolete(true, alternative).Hide()
}

func (s *FunctionSymbol) WithOptimizedAlternative(alternative string) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.optimizedAlternative == alternative {
		return s
	}
	clone := s.clone()
	clone.optimizedAlternative = alternative
	return clone
}

func (s *FunctionSymbol) WithCustomAvailability(availability CustomAvailability) *FunctionSymbol {
	if s == nil {
		return nil
	}
	if s.customAvailability == nil && availability == nil {
		return s
	}
	clone := s.clone()
	clone.customAvailability = availability
	return clone
}

func (s *FunctionSymbol) GetReturnType(globals *GlobalState) *TypeSymbol {
	if s == nil || len(s.signatures) == 0 || s.signatures[0] == nil {
		return ScalarTypes.Unknown.Type()
	}
	return s.signatures[0].GetReturnType(globals)
}

func (s *FunctionSymbol) MinArgumentCount() int {
	if s == nil || len(s.signatures) == 0 {
		return -1
	}
	min := -1
	for _, signature := range s.signatures {
		if signature == nil {
			continue
		}
		if min == -1 || signature.MinArgumentCount < min {
			min = signature.MinArgumentCount
		}
	}
	return min
}

func (s *FunctionSymbol) MaxArgumentCount() int {
	if s == nil || len(s.signatures) == 0 {
		return -1
	}
	max := -1
	for _, signature := range s.signatures {
		if signature == nil {
			continue
		}
		if max == -1 || signature.MaxArgumentCount > max {
			max = signature.MaxArgumentCount
		}
	}
	return max
}

func (s *FunctionSymbol) clone() *FunctionSymbol {
	clone := *s
	clone.signatures = make([]*Signature, len(s.signatures))
	for i, signature := range s.signatures {
		if signature == nil {
			continue
		}
		signatureClone := *signature
		signatureClone.Parameters = append([]*Parameter{}, signature.Parameters...)
		signatureClone.Symbol = &clone
		clone.signatures[i] = &signatureClone
	}
	return &clone
}
