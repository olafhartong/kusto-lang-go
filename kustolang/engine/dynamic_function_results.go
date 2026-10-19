package kustolang

// Ported from src/Kusto.Language/Binder/Binder_Misc.cs and
// src/Kusto.Language/Binder/Binder_FunctionCalls.cs at the revision pinned in
// third_party/kusto/UPSTREAM.json.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// HasDynamicPrimitives reports whether types contains a typed dynamic
// primitive. Generic dynamic values, dynamic arrays, and dynamic bags are not
// dynamic primitives for the binder's rebinding rule.
func HasDynamicPrimitives(types []*TypeSymbol) bool {
	for _, typ := range types {
		if _, ok := DynamicPrimitiveForType(typ); ok {
			return true
		}
	}
	return false
}

// GetUnwrappedDynamicPrimitives returns a positional copy of types with each
// typed dynamic primitive replaced by its underlying scalar type. All other
// values, including nil, errors, arrays, and bags, are preserved unchanged.
func GetUnwrappedDynamicPrimitives(types []*TypeSymbol) []*TypeSymbol {
	if types == nil {
		return nil
	}

	unwrapped := make([]*TypeSymbol, len(types))
	for i, typ := range types {
		if primitive, ok := DynamicPrimitiveForType(typ); ok {
			if primitive.UnderlyingType != nil {
				unwrapped[i] = primitive.UnderlyingType.Type()
			}
		} else {
			unwrapped[i] = typ
		}
	}
	return unwrapped
}

// RefineDynamicFunctionResult applies the binder's dynamic primitive
// rebinding rule. A generic dynamic initial result is refined only when the
// rebound result is a known, non-dynamic scalar other than Unknown.
func RefineDynamicFunctionResult(initialResult, reboundUnwrappedResult *TypeSymbol) *TypeSymbol {
	if initialResult != ScalarTypes.Dynamic.Type() ||
		reboundUnwrappedResult == nil ||
		reboundUnwrappedResult == ScalarTypes.Dynamic.Type() ||
		reboundUnwrappedResult == ScalarTypes.Unknown.Type() ||
		scalarFromType(reboundUnwrappedResult) == nil {
		return initialResult
	}

	return ScalarTypes.GetDynamic(reboundUnwrappedResult)
}
