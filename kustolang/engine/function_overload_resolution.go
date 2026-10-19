package kustolang

// Ported from src/Kusto.Language/Binder/Binder_FunctionCalls.cs at the
// revision pinned in third_party/kusto/UPSTREAM.json.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// GetBestMatchingSignatures returns the signatures that best match the
// supplied arguments. Multiple signatures are returned when no candidate is
// unambiguously better than all others.
//
// arguments and argumentTypes describe the same argument list and should have
// equal lengths. A sole signature is returned unconditionally, even when its
// argument count does not match, to preserve binder recovery behavior.
func GetBestMatchingSignatures(signatures []*Signature, arguments []Expression, argumentTypes []Symbol, requireAllArgumentsMatch bool) []*Signature {
	return getBestMatchingSignatures(signatures, arguments, argumentTypes, requireAllArgumentsMatch, nil)
}

func getBestMatchingSignatures(signatures []*Signature, arguments []Expression, argumentTypes []Symbol, requireAllArgumentsMatch bool, allowLooseParameterMatching func(*Signature) bool) []*Signature {
	if len(signatures) == 0 {
		return nil
	}
	if len(signatures) == 1 {
		return []*Signature{signatures[0]}
	}

	argCount := len(argumentTypes)
	result := make([]*Signature, 0, len(signatures))
	closestCount := 0
	maxCount := 0
	for _, signature := range signatures {
		if argCount >= signature.MinArgumentCount && argCount <= signature.MaxArgumentCount {
			result = append(result, signature)
		} else if argCount < signature.MinArgumentCount && closestCount > signature.MinArgumentCount {
			// Preserve the pinned upstream condition, including its initial-zero behavior.
			closestCount = signature.MinArgumentCount
		}
		if signature.MaxArgumentCount > maxCount {
			maxCount = signature.MaxArgumentCount
		}
	}

	if len(result) == 0 {
		if closestCount == 0 {
			closestCount = maxCount
		}
		for _, signature := range signatures {
			if closestCount >= signature.MinArgumentCount && closestCount <= signature.MaxArgumentCount {
				result = append(result, signature)
			}
		}
	}

	if len(result) <= 1 {
		return result
	}

	mostMatchingParameterCount := 0
	for _, signature := range result {
		count := getParameterMatchCount(signature, arguments, argumentTypes, allowLooseParameterMatching)
		if count > mostMatchingParameterCount {
			mostMatchingParameterCount = count
		}
	}
	if requireAllArgumentsMatch && mostMatchingParameterCount < len(arguments) {
		mostMatchingParameterCount = len(arguments)
	}

	matching := result[:0]
	for _, signature := range result {
		if getParameterMatchCount(signature, arguments, argumentTypes, allowLooseParameterMatching) == mostMatchingParameterCount {
			matching = append(matching, signature)
		}
	}
	result = matching

	if len(result) <= 1 {
		return result
	}

	best := result[0]
	for _, signature := range result[1:] {
		comparison := compareSignatureMatch(signature, best, arguments, argumentTypes, allowLooseParameterMatching)
		if comparison > 0 {
			best = signature
		} else if comparison == 0 {
			return result
		}
	}

	for _, signature := range result {
		if signature != best && compareSignatureMatch(best, signature, arguments, argumentTypes, allowLooseParameterMatching) <= 0 {
			return result
		}
	}

	return []*Signature{best}
}

// GetCommonSignatureReturnType returns the common result type for a set of
// matching signatures. The resolver supplies call-site-aware return types for
// custom and computed signatures that cannot be derived from static metadata.
func GetCommonSignatureReturnType(signatures []*Signature, argumentTypes []*TypeSymbol, resolve func(*Signature) *TypeSymbol) *TypeSymbol {
	if len(signatures) == 0 || resolve == nil {
		return &Error.TypeSymbol
	}
	first := resolve(signatures[0])
	if first == nil {
		first = &Error.TypeSymbol
	}
	for _, signature := range signatures[1:] {
		next := resolve(signature)
		if next == nil || !SymbolAssignableTo(next, first, ConversionNone) {
			if argumentTypesHaveErrorsOrUnknown(argumentTypes) {
				return ScalarTypes.Unknown.Type()
			}
			return &Error.TypeSymbol
		}
	}
	return first
}

func argumentTypesHaveErrorsOrUnknown(argumentTypes []*TypeSymbol) bool {
	for _, argumentType := range argumentTypes {
		if isErrorTypeSymbol(argumentType) || argumentType == ScalarTypes.Unknown.Type() {
			return true
		}
	}
	return false
}

func compareSignatureMatch(signature1, signature2 *Signature, arguments []Expression, argumentTypes []Symbol, allowLooseParameterMatching func(*Signature) bool) int {
	argCount := len(argumentTypes)
	matchCount1 := getParameterMatchCount(signature1, arguments, argumentTypes, allowLooseParameterMatching)
	matchCount2 := getParameterMatchCount(signature2, arguments, argumentTypes, allowLooseParameterMatching)

	if matchCount1 == argCount && matchCount2 < argCount {
		return matchCount1 - matchCount2
	}

	worstMatch1 := getWorstParameterMatch(signature1, arguments, argumentTypes, allowLooseParameterMatching)
	worstMatch2 := getWorstParameterMatch(signature2, arguments, argumentTypes, allowLooseParameterMatching)
	if worstMatch1 != worstMatch2 {
		return int(worstMatch1 - worstMatch2)
	}

	bestMatch1 := getBestParameterMatch(signature1, arguments, argumentTypes, allowLooseParameterMatching)
	bestMatch2 := getBestParameterMatch(signature2, arguments, argumentTypes, allowLooseParameterMatching)
	if bestMatch1 != bestMatch2 {
		return int(bestMatch1 - bestMatch2)
	}

	return matchCount1 - matchCount2
}

func getParameterMatchCount(signature *Signature, arguments []Expression, argumentTypes []Symbol, allowLooseParameterMatching func(*Signature) bool) int {
	argumentParameters := signature.ArgumentParametersForExpressions(arguments, argumentTypes)
	matches := 0
	for i := range argumentTypes {
		if parameterMatchAt(signature, argumentParameters, arguments, argumentTypes, i, allowLooseParameterMatching) != ParameterMatchNone {
			matches++
		}
	}
	return matches
}

func getWorstParameterMatch(signature *Signature, arguments []Expression, argumentTypes []Symbol, allowLooseParameterMatching func(*Signature) bool) ParameterMatchKind {
	argumentParameters := signature.ArgumentParametersForExpressions(arguments, argumentTypes)
	worstMatchKind := ParameterMatchExact
	for i := range arguments {
		matchKind := parameterMatchAt(signature, argumentParameters, arguments, argumentTypes, i, allowLooseParameterMatching)
		if matchKind < worstMatchKind {
			worstMatchKind = matchKind
		}
	}
	return worstMatchKind
}

func getBestParameterMatch(signature *Signature, arguments []Expression, argumentTypes []Symbol, allowLooseParameterMatching func(*Signature) bool) ParameterMatchKind {
	argumentParameters := signature.ArgumentParametersForExpressions(arguments, argumentTypes)
	bestMatchKind := ParameterMatchNone
	for i := range arguments {
		matchKind := parameterMatchAt(signature, argumentParameters, arguments, argumentTypes, i, allowLooseParameterMatching)
		if matchKind > bestMatchKind {
			bestMatchKind = matchKind
		}
	}
	return bestMatchKind
}

func parameterMatchAt(signature *Signature, argumentParameters []*Parameter, arguments []Expression, argumentTypes []Symbol, index int, allowLooseParameterMatching func(*Signature) bool) ParameterMatchKind {
	var parameter *Parameter
	if index >= 0 && index < len(argumentParameters) {
		parameter = argumentParameters[index]
	}
	return signature.ParameterMatchKind(
		argumentParameters,
		argumentTypes,
		parameter,
		expressionAt(arguments, index),
		symbolAt(argumentTypes, index),
		allowLooseParameterMatching != nil && allowLooseParameterMatching(signature),
	)
}
