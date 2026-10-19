package kustolang

// Ported from src/Kusto.Language/Symbols/ArgumentKind.cs,
// ParameterTypeKind.cs, ReturnTypeKind.cs, Parameter.cs, ParameterLayout.cs,
// ParameterLayouts.cs and Signature.cs.
// This file implements the constructor and argument-count behavior needed for
// database/local function symbols. Full parameter layouts and expression-based
// return computation remain future binder work.

type ArgumentKind string

const (
	ArgumentKindExpression                  ArgumentKind = "Expression"
	ArgumentKindExpressionParameter0Element ArgumentKind = "Expression_Parameter0_Element"
	ArgumentKindStarOnly                    ArgumentKind = "StarOnly"
	ArgumentKindStarAllowed                 ArgumentKind = "StarAllowed"
	ArgumentKindAggregate                   ArgumentKind = "Aggregate"
	ArgumentKindColumn                      ArgumentKind = "Column"
	ArgumentKindColumnOrLiteral             ArgumentKind = "Column|Literal"
	ArgumentKindColumnParameter0            ArgumentKind = "Column_Parameter0"
	ArgumentKindColumnParameter0Common      ArgumentKind = "Column_Parameter0_Common"
	ArgumentKindLiteral                     ArgumentKind = "Literal"
	ArgumentKindLiteralNotEmpty             ArgumentKind = "LiteralNotEmpty"
	ArgumentKindConstant                    ArgumentKind = "Constant"

	// ArgumentKindStar is kept as a compatibility alias for the earlier Go port.
	ArgumentKindStar = ArgumentKindStarOnly
)

type ParameterTypeKind string

const (
	ParameterTypeAny                   ParameterTypeKind = "Any"
	ParameterTypeDeclared              ParameterTypeKind = "Declared"
	ParameterTypeScalar                ParameterTypeKind = "Scalar"
	ParameterTypeTabular               ParameterTypeKind = "Tabular"
	ParameterTypeDatabase              ParameterTypeKind = "Database"
	ParameterTypeCluster               ParameterTypeKind = "Cluster"
	ParameterTypeInteger               ParameterTypeKind = "Integer"
	ParameterTypeRealOrDecimal         ParameterTypeKind = "RealOrDecimal"
	ParameterTypeStringOrDynamic       ParameterTypeKind = "StringOrDynamic"
	ParameterTypeStringOrArray         ParameterTypeKind = "StringOrArray"
	ParameterTypeIntegerOrArray        ParameterTypeKind = "IntegerOrArray"
	ParameterTypeDynamicArray          ParameterTypeKind = "DynamicArray"
	ParameterTypeDynamicBag            ParameterTypeKind = "DynamicBag"
	ParameterTypeNumber                ParameterTypeKind = "Number"
	ParameterTypeNumberOrBool          ParameterTypeKind = "NumberOrBool"
	ParameterTypeSummable              ParameterTypeKind = "Summable"
	ParameterTypeOrderable             ParameterTypeKind = "Orderable"
	ParameterTypeNotRealOrBool         ParameterTypeKind = "NotRealOrBool"
	ParameterTypeNotDynamic            ParameterTypeKind = "NotDynamic"
	ParameterTypeNotBool               ParameterTypeKind = "NotBool"
	ParameterTypeParameter0            ParameterTypeKind = "Parameter0"
	ParameterTypeParameter1            ParameterTypeKind = "Parameter1"
	ParameterTypeParameter2            ParameterTypeKind = "Parameter2"
	ParameterTypeCommonScalar          ParameterTypeKind = "CommonScalar"
	ParameterTypeCommonNumber          ParameterTypeKind = "CommonNumber"
	ParameterTypeCommonSummable        ParameterTypeKind = "CommonSummable"
	ParameterTypeCommonOrderable       ParameterTypeKind = "CommonOrderable"
	ParameterTypeCommonScalarOrDynamic ParameterTypeKind = "CommonScalarOrDynamic"
)

type ReturnTypeKind string

const (
	ReturnTypeDeclared                    ReturnTypeKind = "Declared"
	ReturnTypeComputed                    ReturnTypeKind = "Computed"
	ReturnTypeCustom                      ReturnTypeKind = "Custom"
	ReturnTypeParameter0                  ReturnTypeKind = "Parameter0"
	ReturnTypeParameter1                  ReturnTypeKind = "Parameter1"
	ReturnTypeParameter2                  ReturnTypeKind = "Parameter2"
	ReturnTypeParameterN                  ReturnTypeKind = "ParameterN"
	ReturnTypeParameter0Literal           ReturnTypeKind = "Parameter0Literal"
	ReturnTypeParameter1Literal           ReturnTypeKind = "Parameter1Literal"
	ReturnTypeParameterNLiteral           ReturnTypeKind = "ParameterNLiteral"
	ReturnTypeParameter0Promoted          ReturnTypeKind = "Parameter0Promoted"
	ReturnTypeParameter0Cluster           ReturnTypeKind = "Parameter0Cluster"
	ReturnTypeParameter0Database          ReturnTypeKind = "Parameter0Database"
	ReturnTypeParameter0Table             ReturnTypeKind = "Parameter0Table"
	ReturnTypeParameter0ExternalTable     ReturnTypeKind = "Parameter0ExternalTable"
	ReturnTypeParameter0Array             ReturnTypeKind = "Parameter0Array"
	ReturnTypeCommon                      ReturnTypeKind = "Common"
	ReturnTypeCommonNonDynamic            ReturnTypeKind = "CommonNonDynamic"
	ReturnTypeWidest                      ReturnTypeKind = "Widest"
	ReturnTypeParameter0MaterializedView  ReturnTypeKind = "Parameter0MaterializedView"
	ReturnTypeParameter0EntityGroup       ReturnTypeKind = "Parameter0EntityGroup"
	ReturnTypeParameter0StoredQueryResult ReturnTypeKind = "Parameter0StoredQueryResult"
	ReturnTypeParameter0Graph             ReturnTypeKind = "Parameter0Graph"
)

const MaxRepeat = 32767

type ParameterLayout string

const (
	ParameterLayoutFixed          ParameterLayout = "Fixed"
	ParameterLayoutRepeating      ParameterLayout = "Repeating"
	ParameterLayoutRepeatingSkip  ParameterLayout = "RepeatingSkipping"
	ParameterLayoutBlockRepeating ParameterLayout = "BlockRepeating"
	ParameterLayoutCustom         ParameterLayout = "Custom"
)

type CustomReturnType func(CustomReturnTypeContext) Symbol

type ParameterLayoutBuilder func(signature *Signature, arguments []Expression, argumentTypes []Symbol) []*Parameter

type CustomReturnTypeContext struct {
	Location                *SyntaxNode
	Arguments               []Expression
	ArgumentTypes           []*TypeSymbol
	ArgumentResultSymbols   []Symbol
	ArgumentParameters      []*Parameter
	ArgumentIsConstant      []bool
	ArgumentSemantics       []SemanticInfo
	ArgumentResultNames     []string
	ArgumentConstantValues  []any
	RowScope                *TableSymbol
	SummarizeByColumns      []*ColumnSymbol
	Signature               *Signature
	Globals                 *GlobalState
	CurrentCluster          *ClusterSymbol
	CurrentDatabase         *DatabaseSymbol
	CurrentFunction         *FunctionSymbol
	ReferencedSymbolByName  func(name string) Symbol
	ResultTypeByName        func(name string) *TypeSymbol
	ResultNameForExpression func(expression Expression, defaultName string) string
}

func (c CustomReturnTypeContext) GetArgument(parameterName string) Expression {
	parameter := c.getParameter(parameterName)
	return c.GetArgumentForParameter(parameter)
}

func (c CustomReturnTypeContext) GetArgumentForParameter(parameter *Parameter) Expression {
	for i, candidate := range c.ArgumentParameters {
		if candidate == parameter && i < len(c.Arguments) {
			return c.Arguments[i]
		}
	}
	return nil
}

func (c CustomReturnTypeContext) GetArguments(parameterName string) []Expression {
	parameter := c.getParameter(parameterName)
	if parameter == nil {
		return []Expression{}
	}
	return c.GetArgumentsForParameter(parameter)
}

func (c CustomReturnTypeContext) GetArgumentsForParameter(parameter *Parameter) []Expression {
	out := []Expression{}
	for i, candidate := range c.ArgumentParameters {
		if candidate == parameter && i < len(c.Arguments) {
			out = append(out, c.Arguments[i])
		}
	}
	return out
}

func (c CustomReturnTypeContext) GetArgumentReferencedSymbol(parameterName string) Symbol {
	parameter := c.getParameter(parameterName)
	if parameter == nil {
		return nil
	}
	for i, candidate := range c.ArgumentParameters {
		if candidate == parameter && i < len(c.ArgumentSemantics) {
			return c.ArgumentSemantics[i].ReferencedSymbol()
		}
	}
	return nil
}

func (c CustomReturnTypeContext) GetArgumentReferencedSymbols(parameterName string) []Symbol {
	parameter := c.getParameter(parameterName)
	if parameter == nil {
		return nil
	}
	out := []Symbol{}
	for i, candidate := range c.ArgumentParameters {
		if candidate == parameter && i < len(c.ArgumentSemantics) {
			if symbol := c.ArgumentSemantics[i].ReferencedSymbol(); symbol != nil {
				out = append(out, symbol)
			}
		}
	}
	return out
}

func (c CustomReturnTypeContext) GetReferencedSymbol(name string) Symbol {
	if c.ReferencedSymbolByName == nil {
		return nil
	}
	return c.ReferencedSymbolByName(name)
}

func (c CustomReturnTypeContext) GetResultType(name string) *TypeSymbol {
	if c.ResultTypeByName == nil {
		return nil
	}
	return c.ResultTypeByName(name)
}

func (c CustomReturnTypeContext) GetResultName(argument Expression, defaultName ...string) string {
	fallback := ""
	if len(defaultName) > 0 {
		fallback = defaultName[0]
	}
	if c.ResultNameForExpression != nil {
		return c.ResultNameForExpression(argument, fallback)
	}
	for i, candidate := range c.Arguments {
		if candidate == argument {
			if i < len(c.ArgumentResultNames) && c.ArgumentResultNames[i] != "" {
				return c.ArgumentResultNames[i]
			}
			if i < len(c.ArgumentSemantics) {
				if column, ok := c.ArgumentSemantics[i].ReferencedSymbol().(*ColumnSymbol); ok {
					return column.SymbolName()
				}
			}
			return fallback
		}
	}
	return fallback
}

func (c CustomReturnTypeContext) GetConstantValue(argument Expression) any {
	for i, candidate := range c.Arguments {
		if candidate == argument && i < len(c.ArgumentConstantValues) {
			return c.ArgumentConstantValues[i]
		}
	}
	return nil
}

func (c CustomReturnTypeContext) getParameter(parameterName string) *Parameter {
	if c.Signature == nil {
		return nil
	}
	return c.Signature.GetParameter(parameterName)
}

type Parameter struct {
	Name                  string
	TypeKind              ParameterTypeKind
	DeclaredTypes         []*TypeSymbol
	DeclaredSymbols       []Symbol
	ArgumentKind          ArgumentKind
	MinOccurring          int
	MaxOccurring          int
	Values                []any
	Examples              []string
	IsCaseSensitive       bool
	DefaultValueIndicator string
	DefaultValue          Expression
	Description           string
}

func NewParameter(name string, typ *TypeSymbol, options ...ParameterOption) *Parameter {
	if typ == nil {
		typ = ScalarTypes.Unknown.Type()
	}
	p := &Parameter{
		Name:            name,
		TypeKind:        ParameterTypeDeclared,
		DeclaredTypes:   []*TypeSymbol{typ},
		DeclaredSymbols: []Symbol{typ},
		ArgumentKind:    ArgumentKindExpression,
		MinOccurring:    1,
		MaxOccurring:    1,
	}
	return applyParameterOptions(p, options)
}

func NewParameterSymbolType(name string, symbol Symbol, options ...ParameterOption) *Parameter {
	if symbol == nil {
		symbol = ScalarTypes.Unknown.Type()
	}
	typ := returnTypeFromSymbol(symbol)
	if typ == nil {
		typ = ScalarTypes.Unknown.Type()
	}
	p := &Parameter{
		Name:            name,
		TypeKind:        ParameterTypeDeclared,
		DeclaredTypes:   []*TypeSymbol{typ},
		DeclaredSymbols: []Symbol{symbol},
		ArgumentKind:    ArgumentKindExpression,
		MinOccurring:    1,
		MaxOccurring:    1,
	}
	return applyParameterOptions(p, options)
}

func NewParameterKind(name string, typeKind ParameterTypeKind, options ...ParameterOption) *Parameter {
	p := &Parameter{
		Name:         name,
		TypeKind:     typeKind,
		ArgumentKind: ArgumentKindExpression,
		MinOccurring: 1,
		MaxOccurring: 1,
	}
	return applyParameterOptions(p, options)
}

func NewParameterTypes(name string, types []*TypeSymbol, options ...ParameterOption) *Parameter {
	if len(types) == 0 {
		types = []*TypeSymbol{ScalarTypes.Unknown.Type()}
	}
	p := &Parameter{
		Name:            name,
		TypeKind:        ParameterTypeDeclared,
		DeclaredTypes:   append([]*TypeSymbol{}, types...),
		DeclaredSymbols: declaredSymbolsFromTypes(types),
		ArgumentKind:    ArgumentKindExpression,
		MinOccurring:    1,
		MaxOccurring:    1,
	}
	return applyParameterOptions(p, options)
}

func NewParameterFromSymbol(parameter *ParameterSymbol, isOptional bool, defaultValue Expression) *Parameter {
	if parameter == nil {
		parameter = NewParameterSymbol("", ScalarTypes.Unknown.Type(), "")
	}
	options := []ParameterOption{WithParameterDescription(parameter.Description())}
	if isOptional {
		options = append(options, WithParameterOccurrence(0, 1))
	}
	if defaultValue != nil {
		options = append(options, WithParameterDefaultValue(defaultValue))
	}
	return NewParameterSymbolType(parameter.SymbolName(), parameter.Type, options...)
}

type ParameterOption func(*Parameter)

func applyParameterOptions(parameter *Parameter, options []ParameterOption) *Parameter {
	for _, option := range options {
		if option != nil {
			option(parameter)
		}
	}
	normalizeParameterDefaultOccurrence(parameter)
	return parameter
}

func normalizeParameterDefaultOccurrence(parameter *Parameter) {
	if parameter != nil && parameter.DefaultValue != nil {
		parameter.MinOccurring = 0
		parameter.MaxOccurring = 1
	}
}

func WithArgumentKind(kind ArgumentKind) ParameterOption {
	return func(p *Parameter) {
		p.ArgumentKind = kind
	}
}

func WithParameterOccurrence(min, max int) ParameterOption {
	return func(p *Parameter) {
		p.MinOccurring = min
		p.MaxOccurring = max
	}
}

func WithParameterDescription(description string) ParameterOption {
	return func(p *Parameter) {
		p.Description = description
	}
}

func WithParameterValues(values ...any) ParameterOption {
	return func(p *Parameter) {
		p.Values = append([]any{}, values...)
	}
}

func WithParameterExamples(examples ...string) ParameterOption {
	return func(p *Parameter) {
		p.Examples = append([]string{}, examples...)
	}
}

func WithParameterCaseSensitive(isCaseSensitive bool) ParameterOption {
	return func(p *Parameter) {
		p.IsCaseSensitive = isCaseSensitive
	}
}

func WithParameterDefaultValueIndicator(indicator string) ParameterOption {
	return func(p *Parameter) {
		p.DefaultValueIndicator = indicator
	}
}

func WithParameterDefaultValue(defaultValue Expression) ParameterOption {
	return func(p *Parameter) {
		p.DefaultValue = defaultValue
		if defaultValue != nil {
			p.MinOccurring = 0
			p.MaxOccurring = 1
		}
	}
}

func (p *Parameter) IsOptional() bool {
	return p != nil && p.MinOccurring == 0
}

func (p *Parameter) IsRepeatable() bool {
	return p != nil && p.MaxOccurring > 1
}

func (p *Parameter) Tabularity() Tabularity {
	if p == nil {
		return TabularityUnknown
	}
	switch p.TypeKind {
	case ParameterTypeDeclared:
		if symbol := p.DeclaredSymbol(0); symbol != nil {
			return symbol.SymbolTabularity()
		}
		if len(p.DeclaredTypes) == 0 || p.DeclaredTypes[0] == nil {
			return TabularityUnknown
		}
		return p.DeclaredTypes[0].SymbolTabularity()
	case ParameterTypeTabular, ParameterTypeDatabase, ParameterTypeCluster:
		return TabularityTabular
	default:
		return TabularityScalar
	}
}

func (p *Parameter) DeclaredSymbol(index int) Symbol {
	if p == nil || index < 0 {
		return nil
	}
	if index < len(p.DeclaredSymbols) && p.DeclaredSymbols[index] != nil {
		return p.DeclaredSymbols[index]
	}
	if index < len(p.DeclaredTypes) && p.DeclaredTypes[index] != nil {
		return p.DeclaredTypes[index]
	}
	return nil
}

func (p *Parameter) TypeDependsOnArguments() bool {
	if p == nil {
		return false
	}
	if p.TypeKind != ParameterTypeDeclared {
		return true
	}
	return p.DeclaredSymbol(0) != nil && p.DeclaredSymbol(0).SymbolTabularity() == TabularityTabular
}

func (p *Parameter) IsScalar() bool {
	if p == nil {
		return false
	}
	switch p.Tabularity() {
	case TabularityScalar, TabularityUnknown:
		return true
	default:
		return false
	}
}

func (p *Parameter) IsTabular() bool {
	if p == nil {
		return false
	}
	switch p.Tabularity() {
	case TabularityTabular, TabularityUnknown:
		return true
	default:
		return false
	}
}

func declaredSymbolsFromTypes(types []*TypeSymbol) []Symbol {
	symbols := make([]Symbol, len(types))
	for i, typ := range types {
		if typ != nil {
			symbols[i] = typ
		}
	}
	return symbols
}

type Signature struct {
	Symbol                  Symbol
	ReturnKind              ReturnTypeKind
	Parameters              []*Parameter
	MinArgumentCount        int
	MaxArgumentCount        int
	DeclaredReturnType      *TypeSymbol
	DeclaredReturnSymbol    Symbol
	Declaration             *FunctionBody
	CustomReturnType        CustomReturnType
	Body                    string
	Tabularity              Tabularity
	Layout                  ParameterLayout
	CustomLayoutBuilder     ParameterLayoutBuilder
	layoutStrategy          ParameterLayoutStrategy
	IsHidden                bool
	Alternative             string
	isObsolete              bool
	HasRepeatableParameters bool
	HasOptionalParameters   bool
	HasAggregateParameters  bool
	tabularityUnspecified   bool
}

func NewSignature(returnType *TypeSymbol, parameters ...*Parameter) *Signature {
	if returnType == nil {
		returnType = ScalarTypes.Unknown.Type()
	}
	return newSignature(ReturnTypeDeclared, returnType, returnType, nil, "", TabularityUnspecified, parameters)
}

func NewSymbolSignature(returnSymbol Symbol, parameters ...*Parameter) *Signature {
	if returnSymbol == nil {
		returnSymbol = ScalarTypes.Unknown.Type()
	}
	return newSignature(ReturnTypeDeclared, returnTypeFromSymbol(returnSymbol), returnSymbol, nil, "", TabularityUnspecified, parameters)
}

func returnTypeFromSymbol(symbol Symbol) *TypeSymbol {
	switch s := symbol.(type) {
	case *TypeSymbol:
		return s
	case *ScalarSymbol:
		return s.Type()
	case *ColumnSymbol:
		return s.Type
	case *TableSymbol:
		return &s.TypeSymbol
	case *MaterializedViewSymbol:
		return &s.TypeSymbol
	case *ExternalTableSymbol:
		return &s.TypeSymbol
	case *StoredQueryResultSymbol:
		return &s.TypeSymbol
	case *DatabaseSymbol:
		return &s.TypeSymbol
	case *ClusterSymbol:
		return &s.TypeSymbol
	case *EntityGroupSymbol:
		return &s.TypeSymbol
	case *EntityGroupElementSymbol:
		return &s.TypeSymbol
	case *PatternSymbol:
		return &s.TypeSymbol
	case *GraphSymbol:
		return &s.TypeSymbol
	case *GraphModelSymbol:
		return &s.TypeSymbol
	case *ErrorSymbol:
		return &s.TypeSymbol
	case *VoidSymbol:
		return &s.TypeSymbol
	case *DynamicAnySymbol:
		return s.Type()
	case *DynamicPrimitiveSymbol:
		return s.Type()
	case *DynamicArraySymbol:
		return s.Type()
	case *DynamicBagSymbol:
		return s.Type()
	case *TupleSymbol:
		return s.Type()
	default:
		if symbol != nil && symbol.SymbolTabularity() == TabularityScalar {
			return NewTypeSymbol(symbol.SymbolName())
		}
		return ScalarTypes.Unknown.Type()
	}
}

func (s *Signature) GetReturnType(globals *GlobalState) *TypeSymbol {
	if s == nil {
		return ScalarTypes.Unknown.Type()
	}
	switch s.ReturnKind {
	case ReturnTypeDeclared:
		if s.DeclaredReturnType != nil {
			return s.DeclaredReturnType
		}
		if s.DeclaredReturnSymbol != nil {
			return returnTypeFromSymbol(s.DeclaredReturnSymbol)
		}
	case ReturnTypeComputed:
		return GetComputedReturnType(s, globals, nil)
	case ReturnTypeParameter0Cluster:
		return &UnknownClusterSymbol.TypeSymbol
	case ReturnTypeParameter0Database:
		return &UnknownDatabaseSymbol.TypeSymbol
	case ReturnTypeParameter0Table, ReturnTypeParameter0ExternalTable, ReturnTypeParameter0MaterializedView:
		table := EmptyTableSymbol.WithIsOpen(true)
		return &table.TypeSymbol
	case ReturnTypeParameter0EntityGroup:
		return &EmptyEntityGroupSymbol.TypeSymbol
	case ReturnTypeParameter0StoredQueryResult:
		stored := NewStoredQueryResultSymbol("")
		return &stored.TypeSymbol
	}
	if s.Tabularity == TabularityTabular {
		table := EmptyTableSymbol.WithIsOpen(true)
		return &table.TypeSymbol
	}
	return ScalarTypes.Unknown.Type()
}

func (s *Signature) GetReturnTypeForArgumentTypes(globals *GlobalState, argumentTypes []*TypeSymbol) *TypeSymbol {
	if s == nil {
		return ScalarTypes.Unknown.Type()
	}
	argumentParameters := s.ArgumentParametersForTypes(argumentTypes)
	return s.GetReturnTypeForArgumentParameters(globals, argumentTypes, argumentParameters)
}

func (s *Signature) GetReturnTypeForArguments(globals *GlobalState, arguments []Expression, argumentTypes []*TypeSymbol) *TypeSymbol {
	if s == nil {
		return ScalarTypes.Unknown.Type()
	}
	argumentParameters := s.ArgumentParametersForExpressions(arguments, symbolsFromTypeSymbols(argumentTypes), globals)
	return s.GetReturnTypeForArgumentExpressions(globals, arguments, argumentTypes, argumentParameters)
}

func (s *Signature) GetReturnTypeForArgumentExpressions(globals *GlobalState, arguments []Expression, argumentTypes []*TypeSymbol, argumentParameters []*Parameter) *TypeSymbol {
	if s == nil {
		return ScalarTypes.Unknown.Type()
	}
	switch s.ReturnKind {
	case ReturnTypeParameter0Literal:
		return s.argumentTypeLiteralForParameter(0, arguments, argumentParameters)
	case ReturnTypeParameter1Literal:
		return s.argumentTypeLiteralForParameter(1, arguments, argumentParameters)
	case ReturnTypeParameterNLiteral:
		return s.argumentTypeLiteralForParameter(len(s.Parameters)-1, arguments, argumentParameters)
	default:
		return s.GetReturnTypeForArgumentParameters(globals, argumentTypes, argumentParameters)
	}
}

func (s *Signature) GetReturnTypeForArgumentParameters(globals *GlobalState, argumentTypes []*TypeSymbol, argumentParameters []*Parameter) *TypeSymbol {
	if s == nil {
		return ScalarTypes.Unknown.Type()
	}
	switch s.ReturnKind {
	case ReturnTypeDeclared:
		if s.DeclaredReturnType != nil {
			return s.DeclaredReturnType
		}
		if s.DeclaredReturnSymbol != nil {
			return returnTypeFromSymbol(s.DeclaredReturnSymbol)
		}
	case ReturnTypeComputed:
		return GetComputedReturnType(s, globals, argumentTypes)
	case ReturnTypeParameter0:
		return s.argumentTypeForParameter(0, argumentTypes, argumentParameters)
	case ReturnTypeParameter1:
		return s.argumentTypeForParameter(1, argumentTypes, argumentParameters)
	case ReturnTypeParameter2:
		return s.argumentTypeForParameter(2, argumentTypes, argumentParameters)
	case ReturnTypeParameterN:
		return s.argumentTypeForParameter(len(s.Parameters)-1, argumentTypes, argumentParameters)
	case ReturnTypeParameter0Array:
		elementType := s.argumentTypeForParameter(0, argumentTypes, argumentParameters)
		if isErrorTypeSymbol(elementType) {
			return &Error.TypeSymbol
		}
		return ScalarTypes.GetDynamicArray(elementType).Type()
	case ReturnTypeParameter0Promoted:
		return PromoteToLong(s.argumentTypeForParameter(0, argumentTypes, argumentParameters))
	case ReturnTypeCommon:
		if common := GetCommonArgumentType(argumentParameters, argumentTypes); common != nil {
			return common
		}
		return &Error.TypeSymbol
	case ReturnTypeCommonNonDynamic:
		if common := GetCommonArgumentTypeIgnoreDynamic(argumentParameters, argumentTypes); common != nil {
			return common
		}
		return &Error.TypeSymbol
	case ReturnTypeWidest:
		if widest := GetWidestScalarType(argumentTypes); widest != nil {
			return PromoteToLong(widest.Type())
		}
		return &Error.TypeSymbol
	case ReturnTypeParameter0Cluster,
		ReturnTypeParameter0Database,
		ReturnTypeParameter0Table,
		ReturnTypeParameter0ExternalTable,
		ReturnTypeParameter0MaterializedView,
		ReturnTypeParameter0EntityGroup,
		ReturnTypeParameter0StoredQueryResult:
		return s.GetReturnType(globals)
	}
	if s.Tabularity == TabularityTabular {
		table := EmptyTableSymbol.WithIsOpen(true)
		return &table.TypeSymbol
	}
	return ScalarTypes.Unknown.Type()
}

func (s *Signature) argumentTypeLiteralForParameter(parameterIndex int, arguments []Expression, argumentParameters []*Parameter) *TypeSymbol {
	argumentIndex := s.argumentIndexForParameter(parameterIndex, argumentParameters)
	if argumentIndex < 0 || argumentIndex >= len(arguments) {
		return &Error.TypeSymbol
	}
	if typ, ok := TypeOfLiteralResultType(arguments[argumentIndex]); ok {
		return typ
	}
	return &Error.TypeSymbol
}

func (s *Signature) argumentTypeForParameter(parameterIndex int, argumentTypes []*TypeSymbol, argumentParameters []*Parameter) *TypeSymbol {
	argumentIndex := s.argumentIndexForParameter(parameterIndex, argumentParameters)
	if argumentIndex < 0 || argumentIndex >= len(argumentTypes) || argumentTypes[argumentIndex] == nil {
		return &Error.TypeSymbol
	}
	return argumentTypes[argumentIndex]
}

func (s *Signature) argumentIndexForParameter(parameterIndex int, argumentParameters []*Parameter) int {
	if s == nil || parameterIndex < 0 || parameterIndex >= len(s.Parameters) {
		return -1
	}
	target := s.Parameters[parameterIndex]
	for i, parameter := range argumentParameters {
		if parameter == target {
			return i
		}
	}
	return -1
}

func TypeOfLiteralResultType(argument Expression) (*TypeSymbol, bool) {
	switch expr := argument.(type) {
	case *ParenthesizedExpression:
		return TypeOfLiteralResultType(expr.Expression())
	case *SimpleNamedExpression:
		return TypeOfLiteralResultType(expr.Expression())
	case *CompoundNamedExpression:
		return TypeOfLiteralResultType(expr.Expression())
	case *TypeOfLiteralExpression:
		if symbol := expr.ReferencedSymbol(); symbol != nil {
			if typ := returnTypeFromSymbol(symbol); typ != nil && !isErrorTypeSymbol(typ) {
				return typ, true
			}
		}
		if expr.Types() == nil || expr.Types().Count() != 1 {
			return &Error.TypeSymbol, false
		}
		primitive, ok := expr.Types().At(0).Element().(*PrimitiveTypeExpression)
		if !ok {
			return &Error.TypeSymbol, false
		}
		symbol := declaredSymbolFromPrimitiveTypeExpression(primitive)
		if typ := returnTypeFromSymbol(symbol); typ != nil {
			return typ, true
		}
		return &Error.TypeSymbol, false
	default:
		return &Error.TypeSymbol, false
	}
}

func symbolsFromTypeSymbols(types []*TypeSymbol) []Symbol {
	symbols := make([]Symbol, len(types))
	for i, typ := range types {
		if typ != nil {
			symbols[i] = typ
		}
	}
	return symbols
}

func NewComputedSignature(body string, tabularity Tabularity, parameters ...*Parameter) *Signature {
	return newSignature(ReturnTypeComputed, nil, nil, nil, body, tabularity, parameters)
}

func NewDeclarationSignature(declaration *FunctionBody, parameters ...*Parameter) *Signature {
	body := ""
	if declaration != nil {
		body = declaration.Text(IncludeTriviaInterior)
	}
	signature := newSignature(ReturnTypeComputed, nil, nil, nil, body, TabularityUnspecified, parameters)
	signature.Declaration = declaration
	return signature
}

func NewReturnKindSignature(returnKind ReturnTypeKind, parameters ...*Parameter) *Signature {
	return newSignature(returnKind, nil, nil, nil, "", TabularityUnspecified, parameters)
}

func NewCustomSignature(customReturnType CustomReturnType, tabularity Tabularity, parameters ...*Parameter) *Signature {
	return newSignature(ReturnTypeCustom, nil, nil, customReturnType, "", tabularity, parameters)
}

func newSignature(returnKind ReturnTypeKind, returnType *TypeSymbol, returnSymbol Symbol, customReturnType CustomReturnType, body string, tabularity Tabularity, parameters []*Parameter) *Signature {
	s := &Signature{
		ReturnKind:            returnKind,
		Parameters:            append([]*Parameter{}, parameters...),
		DeclaredReturnType:    returnType,
		DeclaredReturnSymbol:  returnSymbol,
		CustomReturnType:      customReturnType,
		Body:                  body,
		Tabularity:            tabularity,
		tabularityUnspecified: tabularity == TabularityUnspecified,
	}
	for _, p := range s.Parameters {
		if p == nil {
			continue
		}
		if p.IsRepeatable() {
			s.HasRepeatableParameters = true
		}
		if p.IsOptional() {
			s.HasOptionalParameters = true
		}
		if p.ArgumentKind == ArgumentKindAggregate {
			s.HasAggregateParameters = true
		}
		s.MinArgumentCount += p.MinOccurring
		s.MaxArgumentCount += p.MaxOccurring
	}
	if s.Tabularity == TabularityUnspecified {
		s.Tabularity = inferSignatureTabularity(s)
	}
	if s.Layout == "" {
		if s.HasRepeatableParameters {
			s.Layout = ParameterLayoutRepeating
		} else {
			s.Layout = ParameterLayoutFixed
		}
	}
	return s
}

func inferSignatureTabularity(s *Signature) Tabularity {
	if s == nil {
		return TabularityUnknown
	}
	switch s.ReturnKind {
	case ReturnTypeDeclared:
		if s.DeclaredReturnSymbol != nil {
			return s.DeclaredReturnSymbol.SymbolTabularity()
		}
		if s.DeclaredReturnType != nil {
			return s.DeclaredReturnType.SymbolTabularity()
		}
	case ReturnTypeComputed, ReturnTypeCustom:
		return TabularityUnknown
	case ReturnTypeParameter0:
		if len(s.Parameters) > 0 {
			return s.Parameters[0].Tabularity()
		}
		return TabularityUnknown
	case ReturnTypeParameter1:
		if len(s.Parameters) > 1 {
			return s.Parameters[1].Tabularity()
		}
		return TabularityUnknown
	case ReturnTypeParameter2:
		if len(s.Parameters) > 2 {
			return s.Parameters[2].Tabularity()
		}
		return TabularityUnknown
	case ReturnTypeParameterN:
		if len(s.Parameters) > 0 {
			return s.Parameters[len(s.Parameters)-1].Tabularity()
		}
		return TabularityUnknown
	case ReturnTypeParameter0Table,
		ReturnTypeParameter0ExternalTable,
		ReturnTypeParameter0MaterializedView,
		ReturnTypeParameter0Database,
		ReturnTypeParameter0Cluster,
		ReturnTypeParameter0EntityGroup,
		ReturnTypeParameter0StoredQueryResult:
		return TabularityTabular
	case ReturnTypeParameter0Graph:
		return TabularityOther
	}
	return TabularityScalar
}

func (s *Signature) IsValidArgumentCount(argumentCount int) bool {
	if s == nil {
		return false
	}
	if s.layoutStrategy != nil {
		return s.layoutStrategy.IsValidArgumentCount(s, argumentCount)
	}
	if argumentCount < s.MinArgumentCount || argumentCount > s.MaxArgumentCount {
		return false
	}
	if s.Layout != ParameterLayoutBlockRepeating || !s.HasRepeatableParameters {
		return true
	}
	first, last := -1, -1
	for i, parameter := range s.Parameters {
		if parameter != nil && parameter.IsRepeatable() {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 || last < first {
		return true
	}
	variable := last - first + 1
	before := first
	after := len(s.Parameters) - last - 1
	fixed := before + after
	return (argumentCount-fixed)%variable == 0
}

func (s *Signature) IsObsolete() bool {
	return s != nil && s.isObsolete
}

func (s *Signature) IsScalar() bool {
	if s == nil {
		return false
	}
	switch s.Tabularity {
	case TabularityScalar, TabularityUnknown:
		return true
	default:
		return false
	}
}

func (s *Signature) IsTabular() bool {
	if s == nil {
		return false
	}
	switch s.Tabularity {
	case TabularityTabular, TabularityUnknown:
		return true
	default:
		return false
	}
}

func (s *Signature) GetParameter(name string) *Parameter {
	if s == nil {
		return nil
	}
	for i := len(s.Parameters) - 1; i >= 0; i-- {
		parameter := s.Parameters[i]
		if parameter != nil && parameter.Name == name {
			return parameter
		}
	}
	return nil
}

// ArgumentParameterAt returns the parameter associated with an argument index
// for the ported fixed, repeating and block-repeating layouts. It intentionally
// does not model expression-specific named arguments or optional-parameter
// skipping decisions that require the full binder expression list.
func (s *Signature) ArgumentParameterAt(index int, argumentCount int) *Parameter {
	if s == nil || index < 0 || len(s.Parameters) == 0 {
		return nil
	}
	if argumentCount < index+1 {
		argumentCount = index + 1
	}
	switch s.Layout {
	case ParameterLayoutBlockRepeating:
		return s.blockRepeatingArgumentParameterAt(index, argumentCount)
	case ParameterLayoutRepeating, ParameterLayoutRepeatingSkip, ParameterLayoutCustom:
		return s.repeatingArgumentParameterAt(index)
	default:
		if index < len(s.Parameters) {
			return s.Parameters[index]
		}
		return nil
	}
}

func (s *Signature) repeatingArgumentParameterAt(index int) *Parameter {
	currentParameter := 0
	currentParameterCount := 0
	for i := 0; i <= index; i++ {
		if i >= s.MaxArgumentCount {
			return nil
		}
		if currentParameter >= len(s.Parameters) {
			return nil
		}
		parameter := s.Parameters[currentParameter]
		if currentParameter < len(s.Parameters)-1 && parameter != nil && currentParameterCount >= parameter.MaxOccurring {
			currentParameter++
			currentParameterCount = 0
			parameter = s.Parameters[currentParameter]
		}
		if parameter == nil {
			return nil
		}
		if currentParameterCount < parameter.MinOccurring ||
			(currentParameterCount == 0 && parameter.MinOccurring == 0 && s.Layout != ParameterLayoutRepeatingSkip) ||
			currentParameter == len(s.Parameters)-1 {
			currentParameterCount++
			if i == index {
				if currentParameterCount <= parameter.MaxOccurring {
					return parameter
				}
				return nil
			}
			continue
		}
		currentParameter++
		currentParameterCount = 1
		if currentParameter >= len(s.Parameters) {
			return nil
		}
		if i == index {
			return s.Parameters[currentParameter]
		}
	}
	return nil
}

func (s *Signature) blockRepeatingArgumentParameterAt(index int, argumentCount int) *Parameter {
	if !s.HasRepeatableParameters {
		if index < len(s.Parameters) {
			return s.Parameters[index]
		}
		return nil
	}
	firstRepeatable := -1
	lastRepeatable := -1
	for i, parameter := range s.Parameters {
		if parameter != nil && parameter.IsRepeatable() {
			if firstRepeatable < 0 {
				firstRepeatable = i
			}
			lastRepeatable = i
		}
	}
	if firstRepeatable < 0 || lastRepeatable < firstRepeatable {
		if index < len(s.Parameters) {
			return s.Parameters[index]
		}
		return nil
	}
	firstRepeatingArgument := firstRepeatable
	repeatingParameterCount := lastRepeatable - firstRepeatable + 1
	minRepeats := s.Parameters[firstRepeatable].MinOccurring
	maxRepeats := s.Parameters[firstRepeatable].MaxOccurring
	minRepeatingArguments := repeatingParameterCount * minRepeats
	maxRepeatingArguments := repeatingParameterCount * maxRepeats
	parametersAfterLastRepeating := len(s.Parameters) - lastRepeatable - 1
	possibleRepeatingArguments := (argumentCount - firstRepeatingArgument) - parametersAfterLastRepeating
	expectedRepeatingArguments := minInt(maxRepeatingArguments, maxInt(minRepeatingArguments, possibleRepeatingArguments))
	repeatingArgumentGroups := (expectedRepeatingArguments + repeatingParameterCount - 1) / repeatingParameterCount
	totalRepeatingArguments := repeatingArgumentGroups * repeatingParameterCount
	lastRepeatingArgument := firstRepeatable + totalRepeatingArguments - 1

	switch {
	case index < firstRepeatingArgument:
		if index < len(s.Parameters) {
			return s.Parameters[index]
		}
	case index <= lastRepeatingArgument:
		return s.Parameters[firstRepeatable+((index-firstRepeatingArgument)%repeatingParameterCount)]
	default:
		parameterIndex := lastRepeatable + (index - lastRepeatingArgument)
		if parameterIndex < len(s.Parameters) {
			return s.Parameters[parameterIndex]
		}
	}
	return nil
}

func (s *Signature) WithLayout(layout ParameterLayout) *Signature {
	if s == nil {
		return nil
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.Layout = layout
	clone.layoutStrategy = nil
	if layout != ParameterLayoutCustom {
		clone.CustomLayoutBuilder = nil
	}
	return &clone
}

func (s *Signature) WithCustomLayout(builder ParameterLayoutBuilder) *Signature {
	if s == nil {
		return nil
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.Layout = ParameterLayoutCustom
	clone.CustomLayoutBuilder = builder
	clone.layoutStrategy = nil
	return &clone
}

func (s *Signature) WithAlternative(alternative string) *Signature {
	if s == nil {
		return nil
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.Alternative = alternative
	clone.isObsolete = true
	return &clone
}

func (s *Signature) WithIsHidden(isHidden bool) *Signature {
	if s == nil {
		return nil
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.IsHidden = isHidden
	return &clone
}

func (s *Signature) Hide() *Signature {
	return s.WithIsHidden(true)
}

func (s *Signature) WithIsObsolete(isObsolete bool, alternative string) *Signature {
	if s == nil {
		return nil
	}
	if isObsolete {
		return s.WithAlternative(alternative)
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.Alternative = ""
	clone.isObsolete = false
	return &clone
}

func (s *Signature) Obsolete(alternative string) *Signature {
	return s.WithIsObsolete(true, alternative)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
