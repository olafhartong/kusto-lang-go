package kustolang

// Ported from src/Kusto.Language/Symbols/ParameterLayout.cs,
// src/Kusto.Language/Symbols/ParameterLayouts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var UnknownParameter = NewParameter("", ScalarTypes.Unknown.Type())

func (s *Signature) ArgumentParametersForTypes(argumentTypes []*TypeSymbol) []*Parameter {
	if s != nil && s.layoutStrategy != nil {
		return cloneParameterList(s.layoutStrategy.GetArgumentParametersForTypes(s, argumentTypes))
	}
	symbols := make([]Symbol, len(argumentTypes))
	arguments := make([]Expression, len(argumentTypes))
	for i, typ := range argumentTypes {
		if typ != nil {
			symbols[i] = typ
			arguments[i] = FakeExpressionCreate(typ)
		} else {
			symbols[i] = ScalarTypes.Unknown.Type()
			arguments[i] = FakeExpressionCreate(ScalarTypes.Unknown.Type())
		}
	}
	if s != nil && s.Layout == ParameterLayoutCustom && s.CustomLayoutBuilder != nil {
		return s.customArgumentParameters(arguments, symbols)
	}
	return s.ArgumentParametersForSymbols(symbols)
}

func (s *Signature) ArgumentParametersForExpressions(arguments []Expression, argumentTypes []Symbol, globals ...*GlobalState) []*Parameter {
	if s != nil && s.layoutStrategy != nil {
		return cloneParameterList(s.layoutStrategy.GetArgumentParameters(s, arguments, globals...))
	}
	types := normalizeExpressionArgumentSymbols(arguments, argumentTypes)
	if s == nil {
		return unknownParameterList(len(arguments))
	}
	switch s.Layout {
	case ParameterLayoutBlockRepeating:
		if !s.HasRepeatableParameters {
			return s.fixedArgumentParametersForExpressions(arguments, globals...)
		}
		return s.blockRepeatingArgumentParameters(len(arguments))
	case ParameterLayoutRepeatingSkip:
		return s.repeatingArgumentParameters(types, arguments, true)
	case ParameterLayoutRepeating:
		return s.repeatingArgumentParameters(types, arguments, false)
	case ParameterLayoutCustom:
		if s.CustomLayoutBuilder != nil {
			return s.customArgumentParameters(arguments, types)
		}
		return s.repeatingArgumentParameters(types, arguments, false)
	default:
		return s.fixedArgumentParametersForExpressions(arguments, globals...)
	}
}

func (s *Signature) ArgumentParametersForSymbols(argumentTypes []Symbol) []*Parameter {
	if s != nil && s.layoutStrategy != nil {
		return cloneParameterList(s.layoutStrategy.GetArgumentParametersForTypes(s, typeSymbolsFromSymbols(argumentTypes)))
	}
	if s == nil {
		return unknownParameterList(len(argumentTypes))
	}
	switch s.Layout {
	case ParameterLayoutBlockRepeating:
		return s.blockRepeatingArgumentParameters(len(argumentTypes))
	case ParameterLayoutRepeatingSkip:
		return s.repeatingArgumentParameters(argumentTypes, nil, true)
	case ParameterLayoutRepeating:
		return s.repeatingArgumentParameters(argumentTypes, nil, false)
	case ParameterLayoutCustom:
		if s.CustomLayoutBuilder != nil {
			return s.customArgumentParameters(nil, argumentTypes)
		}
		return s.repeatingArgumentParameters(argumentTypes, nil, false)
	default:
		return s.fixedArgumentParameters(len(argumentTypes))
	}
}

func (s *Signature) NextPossibleParametersForTypes(argumentTypes []*TypeSymbol) []*Parameter {
	symbols := make([]Symbol, len(argumentTypes))
	for i, typ := range argumentTypes {
		if typ != nil {
			symbols[i] = typ
		} else {
			symbols[i] = ScalarTypes.Unknown.Type()
		}
	}
	return s.NextPossibleParametersForSymbols(symbols)
}

func (s *Signature) NextPossibleParametersForExpressions(arguments []Expression, argumentTypes []Symbol) []*Parameter {
	if s != nil && s.layoutStrategy != nil {
		return cloneParameterList(s.layoutStrategy.GetNextPossibleParameters(s, arguments))
	}
	types := normalizeExpressionArgumentSymbols(arguments, argumentTypes)
	if s == nil || len(s.Parameters) == 0 || len(arguments) >= s.MaxArgumentCount {
		return []*Parameter{}
	}
	switch s.Layout {
	case ParameterLayoutBlockRepeating:
		return s.blockRepeatingNextPossibleParameters(len(arguments))
	case ParameterLayoutRepeatingSkip:
		return s.repeatingNextPossibleParameters(types, arguments, true)
	case ParameterLayoutRepeating:
		return s.repeatingNextPossibleParameters(types, arguments, false)
	case ParameterLayoutCustom:
		if s.CustomLayoutBuilder != nil {
			return s.customNextPossibleParameters(arguments, types)
		}
		return s.repeatingNextPossibleParameters(types, arguments, false)
	default:
		if len(arguments) < len(s.Parameters) {
			return []*Parameter{s.Parameters[len(arguments)]}
		}
		return []*Parameter{}
	}
}

func (s *Signature) NextPossibleParametersForSymbols(argumentTypes []Symbol) []*Parameter {
	if s != nil && s.layoutStrategy != nil {
		arguments := make([]Expression, len(argumentTypes))
		for i, symbol := range argumentTypes {
			arguments[i] = FakeExpressionCreate(returnTypeFromSymbol(symbol))
		}
		return cloneParameterList(s.layoutStrategy.GetNextPossibleParameters(s, arguments))
	}
	if s == nil || len(s.Parameters) == 0 || len(argumentTypes) >= s.MaxArgumentCount {
		return []*Parameter{}
	}
	switch s.Layout {
	case ParameterLayoutBlockRepeating:
		return s.blockRepeatingNextPossibleParameters(len(argumentTypes))
	case ParameterLayoutRepeatingSkip:
		return s.repeatingNextPossibleParameters(argumentTypes, nil, true)
	case ParameterLayoutRepeating:
		return s.repeatingNextPossibleParameters(argumentTypes, nil, false)
	case ParameterLayoutCustom:
		if s.CustomLayoutBuilder != nil {
			return s.customNextPossibleParameters(nil, argumentTypes)
		}
		return s.repeatingNextPossibleParameters(argumentTypes, nil, false)
	default:
		if len(argumentTypes) < len(s.Parameters) {
			return []*Parameter{s.Parameters[len(argumentTypes)]}
		}
		return []*Parameter{}
	}
}

func (s *Signature) AllowsNamedArguments(globals ...*GlobalState) bool {
	if s == nil {
		return false
	}
	fn, ok := s.Symbol.(*FunctionSymbol)
	if !ok || fn == nil {
		return true
	}
	g := DefaultGlobalState
	if len(globals) > 0 && globals[0] != nil {
		g = globals[0]
	}
	if g.IsBuiltInFunction(fn) {
		return false
	}
	return true
}

func (s *Signature) customArgumentParameters(arguments []Expression, argumentTypes []Symbol) []*Parameter {
	count := maxInt(len(arguments), len(argumentTypes))
	out := s.CustomLayoutBuilder(s, arguments, normalizeArgumentSymbols(count, argumentTypes))
	if out == nil {
		out = []*Parameter{}
	}
	result := make([]*Parameter, maxInt(count, len(out)))
	for i := range result {
		if i < len(out) && out[i] != nil {
			result[i] = out[i]
		} else {
			result[i] = UnknownParameter
		}
	}
	return result
}

func (s *Signature) customNextPossibleParameters(arguments []Expression, argumentTypes []Symbol) []*Parameter {
	count := maxInt(len(arguments), len(argumentTypes))
	extendedArguments := append([]Expression{}, arguments...)
	if len(extendedArguments) < count+1 {
		for len(extendedArguments) < count+1 {
			extendedArguments = append(extendedArguments, nil)
		}
	} else {
		extendedArguments = append(extendedArguments, nil)
	}
	extendedTypes := normalizeArgumentSymbols(count+1, argumentTypes)
	parameters := s.customArgumentParameters(extendedArguments, extendedTypes)
	if count >= len(parameters) || parameters[count] == nil {
		return []*Parameter{}
	}
	possible := []*Parameter{parameters[count]}
	if parameters[count].MinOccurring == 0 {
		argumentIndex := count
		for argumentIndex > 0 && parameters[argumentIndex-1] == parameters[count] {
			argumentIndex--
		}
		parameterIndex := parameterIndexOf(s.Parameters, parameters[count])
		if parameterIndex == argumentIndex && parameterIndex < len(s.Parameters)-1 {
			for parameterIndex < len(s.Parameters)-1 {
				parameterIndex++
				next := s.Parameters[parameterIndex]
				if next == nil {
					continue
				}
				possible = append(possible, next)
				if next.MinOccurring != 0 {
					break
				}
			}
		}
	}
	return possible
}

func (s *Signature) fixedArgumentParameters(argumentCount int) []*Parameter {
	out := make([]*Parameter, 0, argumentCount)
	for i := 0; i < argumentCount; i++ {
		if i < len(s.Parameters) && s.Parameters[i] != nil {
			out = append(out, s.Parameters[i])
		} else {
			out = append(out, UnknownParameter)
		}
	}
	return out
}

func (s *Signature) fixedArgumentParametersForExpressions(arguments []Expression, globals ...*GlobalState) []*Parameter {
	out := make([]*Parameter, 0, len(arguments))
	for i, argument := range arguments {
		if s.AllowsNamedArguments(globals...) {
			if named, ok := argument.(*SimpleNamedExpression); ok {
				if parameter := s.GetParameter(parameterNameFromDeclaration(named.Name())); parameter != nil {
					out = append(out, parameter)
				} else {
					out = append(out, UnknownParameter)
				}
				continue
			}
		}
		if i < len(s.Parameters) && s.Parameters[i] != nil {
			out = append(out, s.Parameters[i])
		} else {
			out = append(out, UnknownParameter)
		}
	}
	return out
}

func (s *Signature) repeatingArgumentParameters(argumentTypes []Symbol, arguments []Expression, allowSkippingOptionalParameters bool) []*Parameter {
	out := make([]*Parameter, 0, len(argumentTypes))
	currentParameter := 0
	currentParameterCount := 0
	for i, argumentType := range argumentTypes {
		argument := expressionAt(arguments, i)
		if i >= s.MaxArgumentCount || len(s.Parameters) == 0 || currentParameter >= len(s.Parameters) {
			out = append(out, UnknownParameter)
			continue
		}
		currentParam := s.Parameters[currentParameter]
		if currentParameter < len(s.Parameters)-1 && currentParam != nil && currentParameterCount >= currentParam.MaxOccurring {
			currentParameter++
			currentParameterCount = 0
			currentParam = s.Parameters[currentParameter]
		}
		if currentParam == nil {
			out = append(out, UnknownParameter)
			continue
		}

		if currentParameterCount < currentParam.MinOccurring ||
			(currentParameterCount == 0 && currentParam.MinOccurring == 0 && !allowSkippingOptionalParameters) {
			out = append(out, currentParam)
			currentParameterCount++
			continue
		}

		if currentParameter == len(s.Parameters)-1 {
			if currentParameterCount < currentParam.MaxOccurring {
				out = append(out, currentParam)
				currentParameterCount++
			} else {
				out = append(out, UnknownParameter)
			}
			continue
		}

		currentMatch := s.ParameterMatchKind(out, argumentTypes, currentParam, argument, argumentType, false)
		matched := false
		nextParameter := currentParameter + 1
		for nextParameter < len(s.Parameters) {
			nextParam := s.Parameters[nextParameter]
			nextMatch := s.ParameterMatchKind(out, argumentTypes, nextParam, argument, argumentType, false)
			switch {
			case currentMatch >= nextMatch && currentMatch != ParameterMatchNone:
				out = append(out, currentParam)
				currentParameterCount++
				matched = true
			case nextMatch != ParameterMatchNone:
				out = append(out, nextParam)
				currentParameter = nextParameter
				currentParameterCount = 1
				matched = true
			case nextParam != nil && nextParam.MinOccurring == 0 && allowSkippingOptionalParameters && nextParameter < len(s.Parameters)-1:
				nextParameter++
				continue
			default:
				out = append(out, currentParam)
				currentParameterCount++
				matched = true
			}
			break
		}
		if !matched {
			out = append(out, currentParam)
			currentParameterCount++
		}
	}
	return out
}

func (s *Signature) repeatingNextPossibleParameters(argumentTypes []Symbol, arguments []Expression, allowSkippingOptionalParameters bool) []*Parameter {
	existing, currentParameter, currentParameterCount := s.repeatingArgumentParametersWithState(argumentTypes, arguments, allowSkippingOptionalParameters)
	_ = existing
	if currentParameter >= len(s.Parameters) {
		return []*Parameter{}
	}
	currentParam := s.Parameters[currentParameter]
	if currentParameter < len(s.Parameters)-1 && currentParam != nil && currentParameterCount >= currentParam.MaxOccurring {
		currentParameter++
		currentParameterCount = 0
		currentParam = s.Parameters[currentParameter]
	}
	if currentParam == nil {
		return []*Parameter{}
	}
	if currentParameterCount < currentParam.MinOccurring {
		return []*Parameter{currentParam}
	}
	if currentParameter == len(s.Parameters)-1 {
		if currentParameterCount < currentParam.MaxOccurring {
			return []*Parameter{currentParam}
		}
		return []*Parameter{}
	}
	out := []*Parameter{currentParam}
	if currentParameterCount > 0 || allowSkippingOptionalParameters {
		for nextParameter := currentParameter + 1; nextParameter < len(s.Parameters); nextParameter++ {
			nextParam := s.Parameters[nextParameter]
			if nextParam == nil {
				continue
			}
			out = append(out, nextParam)
			if nextParam.MinOccurring == 0 && allowSkippingOptionalParameters && nextParameter < len(s.Parameters)-1 {
				continue
			}
			break
		}
	}
	return out
}

func (s *Signature) repeatingArgumentParametersWithState(argumentTypes []Symbol, arguments []Expression, allowSkippingOptionalParameters bool) ([]*Parameter, int, int) {
	out := make([]*Parameter, 0, len(argumentTypes))
	currentParameter := 0
	currentParameterCount := 0
	for i, argumentType := range argumentTypes {
		argument := expressionAt(arguments, i)
		if i >= s.MaxArgumentCount || len(s.Parameters) == 0 || currentParameter >= len(s.Parameters) {
			out = append(out, UnknownParameter)
			continue
		}
		currentParam := s.Parameters[currentParameter]
		if currentParameter < len(s.Parameters)-1 && currentParam != nil && currentParameterCount >= currentParam.MaxOccurring {
			currentParameter++
			currentParameterCount = 0
			currentParam = s.Parameters[currentParameter]
		}
		if currentParam == nil {
			out = append(out, UnknownParameter)
			continue
		}
		if currentParameterCount < currentParam.MinOccurring ||
			(currentParameterCount == 0 && currentParam.MinOccurring == 0 && !allowSkippingOptionalParameters) {
			out = append(out, currentParam)
			currentParameterCount++
			continue
		}
		if currentParameter == len(s.Parameters)-1 {
			if currentParameterCount < currentParam.MaxOccurring {
				out = append(out, currentParam)
				currentParameterCount++
			} else {
				out = append(out, UnknownParameter)
			}
			continue
		}
		currentMatch := s.ParameterMatchKind(out, argumentTypes, currentParam, argument, argumentType, false)
		matched := false
		nextParameter := currentParameter + 1
		for nextParameter < len(s.Parameters) {
			nextParam := s.Parameters[nextParameter]
			nextMatch := s.ParameterMatchKind(out, argumentTypes, nextParam, argument, argumentType, false)
			switch {
			case currentMatch >= nextMatch && currentMatch != ParameterMatchNone:
				out = append(out, currentParam)
				currentParameterCount++
				matched = true
			case nextMatch != ParameterMatchNone:
				out = append(out, nextParam)
				currentParameter = nextParameter
				currentParameterCount = 1
				matched = true
			case nextParam != nil && nextParam.MinOccurring == 0 && allowSkippingOptionalParameters && nextParameter < len(s.Parameters)-1:
				nextParameter++
				continue
			default:
				out = append(out, currentParam)
				currentParameterCount++
				matched = true
			}
			break
		}
		if !matched {
			out = append(out, currentParam)
			currentParameterCount++
		}
	}
	return out, currentParameter, currentParameterCount
}

func (s *Signature) blockRepeatingArgumentParameters(argumentCount int) []*Parameter {
	out := make([]*Parameter, 0, argumentCount)
	for i := 0; i < argumentCount; i++ {
		parameter := s.blockRepeatingArgumentParameterAt(i, argumentCount)
		if parameter == nil {
			parameter = UnknownParameter
		}
		out = append(out, parameter)
	}
	return out
}

func (s *Signature) blockRepeatingNextPossibleParameters(argumentCount int) []*Parameter {
	if !s.HasRepeatableParameters {
		if argumentCount < len(s.Parameters) {
			return []*Parameter{s.Parameters[argumentCount]}
		}
		return []*Parameter{}
	}
	firstRepeatable, lastRepeatable := s.repeatableParameterRange()
	if firstRepeatable < 0 {
		if argumentCount < len(s.Parameters) {
			return []*Parameter{s.Parameters[argumentCount]}
		}
		return []*Parameter{}
	}
	if argumentCount < firstRepeatable {
		return []*Parameter{s.Parameters[argumentCount]}
	}
	repeatableCount := lastRepeatable - firstRepeatable + 1
	minOccurring := s.Parameters[firstRepeatable].MinOccurring
	maxOccurring := s.Parameters[firstRepeatable].MaxOccurring
	repeatGroup := (argumentCount - firstRepeatable) / repeatableCount
	repeatOffset := (argumentCount - firstRepeatable) % repeatableCount
	if repeatGroup < maxOccurring {
		out := []*Parameter{s.Parameters[firstRepeatable+repeatOffset]}
		if repeatGroup >= minOccurring {
			endParam := lastRepeatable + repeatOffset + 1
			if endParam < len(s.Parameters) {
				out = append(out, s.Parameters[endParam])
			}
		}
		return out
	}
	lastRepeatingArgument := firstRepeatable + repeatableCount*maxOccurring
	endParam := (argumentCount - lastRepeatingArgument) + lastRepeatable + 1
	if endParam < len(s.Parameters) {
		return []*Parameter{s.Parameters[endParam]}
	}
	return []*Parameter{}
}

func (s *Signature) repeatableParameterRange() (int, int) {
	first, last := -1, -1
	if s == nil {
		return first, last
	}
	for i, parameter := range s.Parameters {
		if parameter != nil && parameter.IsRepeatable() {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	return first, last
}

func (s *Signature) ParameterMatchKind(argumentParameters []*Parameter, argumentTypes []Symbol, parameter *Parameter, argument Expression, argumentType Symbol, allowImplicitArgumentCoercion bool) ParameterMatchKind {
	if parameter == nil {
		return ParameterMatchNone
	}
	if argumentType == nil || sameType(returnTypeFromSymbol(argumentType), ScalarTypes.Unknown.Type()) {
		return ParameterMatchUnknown
	}
	if isDefaultValueIndicator(parameter, argument, argumentType) {
		return ParameterMatchExact
	}
	if _, ok := argument.(*StarExpression); ok {
		if parameter.ArgumentKind == ArgumentKindStarOnly || parameter.ArgumentKind == ArgumentKindStarAllowed {
			return ParameterMatchExact
		}
		return ParameterMatchNone
	}
	if parameter.ArgumentKind == ArgumentKindStarOnly {
		return ParameterMatchNone
	}
	switch parameter.TypeKind {
	case ParameterTypeAny:
		return ParameterMatchUnknown
	case ParameterTypeDeclared:
		targets := parameter.DeclaredSymbolsList()
		if symbolAssignableToAny(argumentType, targets, ConversionNone) {
			if len(targets) == 1 {
				return ParameterMatchExact
			}
			return ParameterMatchOneOfMany
		}
		if symbolAssignableToAny(argumentType, targets, ConversionPromotable) {
			return ParameterMatchPromoted
		}
		if allowImplicitArgumentCoercion && symbolAssignableToAny(argumentType, targets, ConversionDynamic) {
			return ParameterMatchDynamic
		}
		if allowImplicitArgumentCoercion && symbolAssignableToAny(argumentType, targets, ConversionCompatible) {
			return ParameterMatchCompatible
		}
	case ParameterTypeScalar:
		if IsScalar(argumentType) {
			return ParameterMatchScalar
		}
	case ParameterTypeInteger:
		if isIntegerSymbol(argumentType) {
			return ParameterMatchInteger
		}
	case ParameterTypeRealOrDecimal:
		if isRealOrDecimalSymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeStringOrDynamic:
		if isStringOrDynamicSymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeStringOrArray:
		if isStringOrArraySymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeIntegerOrArray:
		if isIntegerOrArraySymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeDynamicArray:
		if isDynamicArraySymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeDynamicBag:
		if isDynamicBagSymbol(argumentType) {
			return ParameterMatchOneOfMany
		}
	case ParameterTypeNumber:
		if isNumericSymbol(argumentType) {
			return ParameterMatchNumber
		}
	case ParameterTypeNumberOrBool:
		if isNumericSymbol(argumentType) || sameType(returnTypeFromSymbol(argumentType), ScalarTypes.Bool.Type()) {
			return ParameterMatchNumber
		}
	case ParameterTypeSummable:
		if scalarHasFlag(argumentType, ScalarFlagSummable) {
			return ParameterMatchSummable
		}
	case ParameterTypeOrderable:
		if scalarHasFlag(argumentType, ScalarFlagOrderable) {
			return ParameterMatchOrderable
		}
	case ParameterTypeTabular:
		if IsTabular(argumentType) {
			return ParameterMatchTabular
		}
	case ParameterTypeDatabase:
		if argumentType.SymbolKind() == SymbolKindDatabase {
			return ParameterMatchDatabase
		}
	case ParameterTypeCluster:
		if argumentType.SymbolKind() == SymbolKindCluster {
			return ParameterMatchCluster
		}
	case ParameterTypeNotBool:
		if !SymbolAssignableTo(argumentType, ScalarTypes.Bool.Type()) {
			return ParameterMatchNotType
		}
	case ParameterTypeNotRealOrBool:
		if !SymbolAssignableTo(argumentType, ScalarTypes.Real.Type()) && !SymbolAssignableTo(argumentType, ScalarTypes.Bool.Type()) {
			return ParameterMatchNotType
		}
	case ParameterTypeNotDynamic:
		if scalar := scalarSymbolFromSymbol(argumentType); scalar != nil && !isDynamicTypedSymbol(argumentType) {
			return ParameterMatchNotType
		}
	case ParameterTypeParameter0:
		return s.matchDependentParameter(argumentParameters, argumentTypes, 0, argument, argumentType, allowImplicitArgumentCoercion)
	case ParameterTypeParameter1:
		return s.matchDependentParameter(argumentParameters, argumentTypes, 1, argument, argumentType, allowImplicitArgumentCoercion)
	case ParameterTypeParameter2:
		return s.matchDependentParameter(argumentParameters, argumentTypes, 2, argument, argumentType, allowImplicitArgumentCoercion)
	case ParameterTypeCommonScalar, ParameterTypeCommonNumber, ParameterTypeCommonSummable, ParameterTypeCommonOrderable, ParameterTypeCommonScalarOrDynamic:
		if common := commonArgumentType(argumentParameters, argumentTypes); common != nil {
			if SymbolAssignableTo(argumentType, common) {
				return ParameterMatchExact
			}
			if SymbolAssignableTo(argumentType, common, ConversionPromotable) {
				return ParameterMatchPromoted
			}
			if SymbolAssignableTo(argumentType, common, ConversionDynamic) {
				return ParameterMatchDynamic
			}
			if allowImplicitArgumentCoercion && SymbolAssignableTo(argumentType, common, ConversionCompatible) {
				return ParameterMatchCompatible
			}
			if parameter.TypeKind == ParameterTypeCommonScalarOrDynamic && SymbolAssignableTo(argumentType, ScalarTypes.Dynamic.Type()) {
				return ParameterMatchExact
			}
		}
	}
	return ParameterMatchNone
}

func (p *Parameter) DeclaredSymbolsList() []Symbol {
	if p == nil {
		return nil
	}
	count := maxInt(len(p.DeclaredSymbols), len(p.DeclaredTypes))
	out := make([]Symbol, 0, count)
	for i := 0; i < count; i++ {
		if symbol := p.DeclaredSymbol(i); symbol != nil {
			out = append(out, symbol)
		}
	}
	return out
}

func (s *Signature) matchDependentParameter(argumentParameters []*Parameter, argumentTypes []Symbol, index int, argument Expression, argumentType Symbol, allowImplicitArgumentCoercion bool) ParameterMatchKind {
	if index < 0 || index >= len(argumentParameters) {
		return ParameterMatchNone
	}
	return s.ParameterMatchKind(argumentParameters, argumentTypes, argumentParameters[index], argument, argumentType, allowImplicitArgumentCoercion)
}

func symbolAssignableToAny(argumentType Symbol, targets []Symbol, conversion Conversion) bool {
	for _, target := range targets {
		if SymbolAssignableTo(argumentType, target, conversion) {
			return true
		}
	}
	return false
}

func commonArgumentType(argumentParameters []*Parameter, argumentTypes []Symbol) Symbol {
	return GetCommonArgumentType(argumentParameters, typeSymbolsFromSymbols(argumentTypes))
}

func unknownParameterList(count int) []*Parameter {
	out := make([]*Parameter, count)
	for i := range out {
		out[i] = UnknownParameter
	}
	return out
}

func cloneParameterList(parameters []*Parameter) []*Parameter {
	if parameters == nil {
		return []*Parameter{}
	}
	return append([]*Parameter{}, parameters...)
}

func parameterIndexOf(parameters []*Parameter, parameter *Parameter) int {
	for i, candidate := range parameters {
		if candidate == parameter {
			return i
		}
	}
	return -1
}

func normalizeArgumentSymbols(count int, symbols []Symbol) []Symbol {
	out := make([]Symbol, count)
	for i := range out {
		if i < len(symbols) && symbols[i] != nil {
			out[i] = symbols[i]
		} else {
			out[i] = ScalarTypes.Unknown.Type()
		}
	}
	return out
}

func normalizeExpressionArgumentSymbols(arguments []Expression, symbols []Symbol) []Symbol {
	out := normalizeArgumentSymbols(len(arguments), symbols)
	for i, argument := range arguments {
		if i < len(symbols) && symbols[i] != nil {
			continue
		}
		if argument != nil && argument.ResultType() != nil {
			out[i] = argument.ResultType()
		}
	}
	return out
}

func symbolAt(symbols []Symbol, index int) Symbol {
	if index < 0 || index >= len(symbols) || symbols[index] == nil {
		return ScalarTypes.Unknown.Type()
	}
	return symbols[index]
}

func typeSymbolsFromSymbols(symbols []Symbol) []*TypeSymbol {
	types := make([]*TypeSymbol, len(symbols))
	for i, symbol := range symbols {
		types[i] = returnTypeFromSymbol(symbol)
	}
	return types
}

func expressionAt(arguments []Expression, index int) Expression {
	if index < 0 || index >= len(arguments) {
		return nil
	}
	return arguments[index]
}

func isDefaultValueIndicator(parameter *Parameter, argument Expression, argumentType Symbol) bool {
	if parameter == nil || parameter.DefaultValueIndicator == "" || argument == nil {
		return false
	}
	if !sameType(returnTypeFromSymbol(argumentType), ScalarTypes.String.Type()) {
		return false
	}
	literal, ok := argument.(*LiteralExpression)
	if !ok || literal.Token() == nil {
		return false
	}
	return literal.Token().ValueText() == parameter.DefaultValueIndicator
}

func isIntegerSymbol(symbol Symbol) bool {
	return scalarHasFlag(symbol, ScalarFlagInteger)
}

func isNumericSymbol(symbol Symbol) bool {
	return scalarHasFlag(symbol, ScalarFlagNumeric)
}

func isRealOrDecimalSymbol(symbol Symbol) bool {
	typ := returnTypeFromSymbol(symbol)
	return sameType(typ, ScalarTypes.Real.Type()) || sameType(typ, ScalarTypes.Decimal.Type())
}

func isStringOrDynamicSymbol(symbol Symbol) bool {
	typ := returnTypeFromSymbol(symbol)
	return sameType(typ, ScalarTypes.String.Type()) || sameType(typ, ScalarTypes.Dynamic.Type()) || isDynamicTypedSymbol(symbol)
}

func isStringOrArraySymbol(symbol Symbol) bool {
	typ := returnTypeFromSymbol(symbol)
	return sameType(typ, ScalarTypes.String.Type()) || isDynamicArraySymbol(symbol)
}

func isIntegerOrArraySymbol(symbol Symbol) bool {
	return isIntegerSymbol(symbol) || isDynamicArraySymbol(symbol)
}

func isDynamicArraySymbol(symbol Symbol) bool {
	return dynamicArraySymbolFromSymbol(symbol) != nil || sameType(returnTypeFromSymbol(symbol), ScalarTypes.Dynamic.Type())
}

func isDynamicBagSymbol(symbol Symbol) bool {
	return dynamicBagSymbolFromSymbol(symbol) != nil || sameType(returnTypeFromSymbol(symbol), ScalarTypes.Dynamic.Type())
}

func scalarHasFlag(symbol Symbol, flag ScalarFlag) bool {
	return scalarHasTypeFactFlag(symbol, flag)
}
