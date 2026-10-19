package kustolang

// Ported from src/Kusto.Language/Binder/ScopeKind.cs and the argument-scope
// helpers in src/Kusto.Language/Binder/Binder_Misc.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ScopeKind int

const (
	ScopeKindNormal ScopeKind = iota
	ScopeKindAggregate
	ScopeKindPlugIn
	ScopeKindOption
)

func (k ScopeKind) String() string {
	switch k {
	case ScopeKindNormal:
		return "Normal"
	case ScopeKindAggregate:
		return "Aggregate"
	case ScopeKindPlugIn:
		return "PlugIn"
	case ScopeKindOption:
		return "Option"
	default:
		return "Unknown"
	}
}

func FunctionHasAggregateParameters(fn *FunctionSymbol) bool {
	if fn == nil {
		return false
	}
	for _, signature := range fn.Signatures() {
		if signature != nil && signature.HasAggregateParameters {
			return true
		}
	}
	return false
}

// ArgumentScopeForFunction mirrors Binder.GetArgumentScope(fc, outerScope) for
// the part that depends on the referenced function and global aggregate catalog.
func ArgumentScopeForFunction(globals *GlobalState, fn *FunctionSymbol, outerScope ScopeKind) ScopeKind {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if globals.IsAggregateFunction(fn) {
		return ScopeKindNormal
	}
	if outerScope == ScopeKindAggregate {
		return ScopeKindAggregate
	}
	return ScopeKindNormal
}

// ArgumentScopeForFunctionArgument mirrors Binder.GetArgumentScope(fc, position,
// outerScope) for the signature metadata already ported to Go. argumentCount is
// used for block-repeating parameter layout decisions.
func ArgumentScopeForFunctionArgument(globals *GlobalState, fn *FunctionSymbol, signature *Signature, argumentIndex int, argumentCount int, outerScope ScopeKind) ScopeKind {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if globals.IsAggregateFunction(fn) {
		return ScopeKindNormal
	}
	if functionArgumentMayBeAggregate(fn, signature, argumentIndex, argumentCount) {
		return ScopeKindAggregate
	}
	if outerScope == ScopeKindAggregate {
		return ScopeKindAggregate
	}
	return ScopeKindNormal
}

// ArgumentScopeForFunctionArgumentExpressions is the source-shaped layout
// path used when the syntax arguments are available. It shares the same
// possible-parameter algorithm as completion, including named, missing,
// custom-layout, and unresolved-overload cases.
func ArgumentScopeForFunctionArgumentExpressions(globals *GlobalState, fn *FunctionSymbol, signature *Signature, arguments []Expression, argumentIndex int, outerScope ScopeKind) ScopeKind {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if globals.IsAggregateFunction(fn) {
		return ScopeKindNormal
	}
	if functionArgumentExpressionsMayBeAggregate(globals, fn, signature, arguments, argumentIndex) {
		return ScopeKindAggregate
	}
	if outerScope == ScopeKindAggregate {
		return ScopeKindAggregate
	}
	return ScopeKindNormal
}

// ArgumentScopeForFunctionCallPosition preserves the upstream distinction
// between a caret inside an existing argument and one positioned after the
// last argument, including whitespace and zero-width recovery nodes.
func ArgumentScopeForFunctionCallPosition(globals *GlobalState, fn *FunctionSymbol, signature *Signature, call *FunctionCallExpression, position int, outerScope ScopeKind) ScopeKind {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if globals.IsAggregateFunction(fn) {
		return ScopeKindNormal
	}
	if functionCallPositionMayBeAggregate(globals, fn, signature, call, position) {
		return ScopeKindAggregate
	}
	if outerScope == ScopeKindAggregate {
		return ScopeKindAggregate
	}
	return ScopeKindNormal
}

func functionCallPositionMayBeAggregate(globals *GlobalState, fn *FunctionSymbol, signature *Signature, call *FunctionCallExpression, position int) bool {
	if fn == nil || call == nil || call.ArgumentList() == nil {
		return false
	}
	signatures := fn.Signatures()
	if signature != nil {
		signatures = []*Signature{signature}
	}
	arguments := functionCallCompletionArguments(call)
	argumentIndex := functionCallCompletionArgumentIndex(call, position)
	for _, candidate := range signatures {
		if candidate == nil || !candidate.HasAggregateParameters {
			continue
		}
		for _, parameter := range possibleArgumentParametersAt(candidate, arguments, argumentIndex, position, globals) {
			if parameter != nil && parameter.ArgumentKind == ArgumentKindAggregate {
				return true
			}
		}
	}
	return false
}

func possibleArgumentParametersAt(signature *Signature, arguments []Expression, argumentIndex int, position int, globals *GlobalState) []*Parameter {
	if signature == nil {
		return []*Parameter{}
	}
	if signature.AllowsNamedArguments(globals) && argumentIndex >= 0 && argumentIndex < len(arguments) {
		if named, ok := arguments[argumentIndex].(*SimpleNamedExpression); ok {
			if parameter := signature.GetParameter(NameDeclarationSimpleName(named.Name())); parameter != nil {
				return []*Parameter{parameter}
			}
			return []*Parameter{}
		}
	}
	afterArguments := len(arguments) == 0
	if !afterArguments {
		last := arguments[len(arguments)-1]
		afterArguments = last != nil && position > last.End()
	}
	if afterArguments {
		prior := append([]Expression{}, arguments...)
		for len(prior) > 0 && SyntaxIsMissing(prior[len(prior)-1]) {
			prior = prior[:len(prior)-1]
		}
		return signature.NextPossibleParametersForExpressions(prior, functionArgumentSymbols(prior))
	}
	mapped := signature.ArgumentParametersForExpressions(arguments, functionArgumentSymbols(arguments), globals)
	if argumentIndex >= 0 && argumentIndex < len(mapped) && mapped[argumentIndex] != nil && mapped[argumentIndex] != UnknownParameter {
		return []*Parameter{mapped[argumentIndex]}
	}
	return []*Parameter{}
}

func functionArgumentExpressionsMayBeAggregate(globals *GlobalState, fn *FunctionSymbol, signature *Signature, arguments []Expression, argumentIndex int) bool {
	if fn == nil || argumentIndex < 0 {
		return false
	}
	signatures := fn.Signatures()
	if signature != nil {
		signatures = []*Signature{signature}
	}
	hasAggregate := false
	for _, candidate := range signatures {
		if candidate != nil && candidate.HasAggregateParameters {
			hasAggregate = true
			break
		}
	}
	if !hasAggregate {
		return false
	}
	possible := functionPossibleParameters(functionArgumentCompletionContext{
		function:      fn,
		signatures:    signatures,
		arguments:     arguments,
		argumentIndex: argumentIndex,
	}, globals)
	for _, candidate := range possible {
		if candidate.parameter != nil && candidate.parameter.ArgumentKind == ArgumentKindAggregate {
			return true
		}
	}
	return false
}

func functionArgumentMayBeAggregate(fn *FunctionSymbol, signature *Signature, argumentIndex int, argumentCount int) bool {
	if signature != nil {
		parameter := signature.ArgumentParameterAt(argumentIndex, argumentCount)
		return parameter != nil && parameter.ArgumentKind == ArgumentKindAggregate
	}
	if !FunctionHasAggregateParameters(fn) {
		return false
	}
	for _, candidate := range fn.Signatures() {
		parameter := candidate.ArgumentParameterAt(argumentIndex, argumentCount)
		if parameter != nil && parameter.ArgumentKind == ArgumentKindAggregate {
			return true
		}
	}
	return false
}
