package kustolang

import "testing"

func TestGetBestMatchingSignaturesExactWinner(t *testing.T) {
	exact := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.String.Type()))
	category := NewSignature(ScalarTypes.String.Type(), NewParameterKind("value", ParameterTypeScalar))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.String.Type())

	got := GetBestMatchingSignatures([]*Signature{category, exact}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, exact)
}

func TestGetBestMatchingSignaturesCategoryBeatsPromoted(t *testing.T) {
	promoted := NewSignature(ScalarTypes.Long.Type(), NewParameter("value", ScalarTypes.Long.Type()))
	category := NewSignature(ScalarTypes.Long.Type(), NewParameterKind("value", ParameterTypeNumber))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Int.Type())

	got := GetBestMatchingSignatures([]*Signature{promoted, category}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, category)
}

func TestGetBestMatchingSignaturesKeepsAllNoneTie(t *testing.T) {
	longSignature := NewSignature(ScalarTypes.Long.Type(), NewParameter("value", ScalarTypes.Long.Type()))
	boolSignature := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.Bool.Type()))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.String.Type())

	got := GetBestMatchingSignatures([]*Signature{longSignature, boolSignature}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, longSignature, boolSignature)
}

func TestGetBestMatchingSignaturesUsesPerSignatureLooseMatching(t *testing.T) {
	intSignature := NewSignature(ScalarTypes.String.Type(), NewParameter("value", ScalarTypes.Int.Type()))
	boolSignature := NewSignature(ScalarTypes.Bool.Type(), NewParameter("value", ScalarTypes.Bool.Type()))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Long.Type())

	strict := GetBestMatchingSignatures([]*Signature{intSignature, boolSignature}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, strict, intSignature, boolSignature)

	loose := getBestMatchingSignatures(
		[]*Signature{intSignature, boolSignature},
		arguments,
		argumentTypes,
		false,
		func(signature *Signature) bool { return signature == intSignature },
	)
	assertMatchingSignatures(t, loose, intSignature)
}

func TestGetBestMatchingSignaturesKeepsPartialTie(t *testing.T) {
	first := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("left", ScalarTypes.Long.Type()),
		NewParameter("right", ScalarTypes.String.Type()),
	)
	second := NewSignature(
		ScalarTypes.String.Type(),
		NewParameter("left", ScalarTypes.String.Type()),
		NewParameter("right", ScalarTypes.Long.Type()),
	)
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Long.Type(), ScalarTypes.Long.Type())

	got := GetBestMatchingSignatures([]*Signature{first, second}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, first, second)
}

func TestGetBestMatchingSignaturesUsesPinnedArgumentCountFallback(t *testing.T) {
	twoArguments := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("one", ScalarTypes.Long.Type()),
		NewParameter("two", ScalarTypes.Long.Type()),
	)
	threeArguments := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("one", ScalarTypes.Long.Type()),
		NewParameter("two", ScalarTypes.Long.Type()),
		NewParameter("three", ScalarTypes.Long.Type()),
	)
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Long.Type())

	got := GetBestMatchingSignatures([]*Signature{twoArguments, threeArguments}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, threeArguments)
}

func TestGetBestMatchingSignaturesReturnsSingleInvalidCount(t *testing.T) {
	signature := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("one", ScalarTypes.Long.Type()),
		NewParameter("two", ScalarTypes.Long.Type()),
	)
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Long.Type())

	got := GetBestMatchingSignatures([]*Signature{signature}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, signature)
}

func TestGetBestMatchingSignaturesMatchesStar(t *testing.T) {
	star := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameterKind("value", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarOnly)),
	)
	scalar := NewSignature(ScalarTypes.Long.Type(), NewParameterKind("value", ParameterTypeScalar))
	argument := overloadResolutionCallArgument(t, "f(*)")

	got := GetBestMatchingSignatures(
		[]*Signature{scalar, star},
		[]Expression{argument},
		[]Symbol{ScalarTypes.String.Type()},
		false,
	)
	assertMatchingSignatures(t, got, star)
}

func TestGetBestMatchingSignaturesPrefersDynamicArrayOverNumber(t *testing.T) {
	array := NewSignature(ScalarTypes.DynamicArray.Type(), NewParameterKind("value", ParameterTypeDynamicArray))
	number := NewSignature(ScalarTypes.Long.Type(), NewParameterKind("value", ParameterTypeNumber))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.DynamicArray.Type())

	got := GetBestMatchingSignatures([]*Signature{number, array}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, got, array)
}

func TestGetBestMatchingSignaturesRejectsNotDynamicForTypedDynamic(t *testing.T) {
	notDynamic := NewSignature(ScalarTypes.String.Type(), NewParameterKind("value", ParameterTypeNotDynamic))
	array := NewSignature(ScalarTypes.DynamicArray.Type(), NewParameterKind("value", ParameterTypeDynamicArray))
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.DynamicArrayOfLong.Type())

	got := GetBestMatchingSignatures([]*Signature{notDynamic, array}, arguments, argumentTypes, true)
	assertMatchingSignatures(t, got, array)
}

func TestGetBestMatchingSignaturesRequireAllArgumentsMatch(t *testing.T) {
	partial := NewSignature(
		ScalarTypes.Long.Type(),
		NewParameter("one", ScalarTypes.Long.Type()),
		NewParameter("two", ScalarTypes.Long.Type()),
	)
	none := NewSignature(
		ScalarTypes.Bool.Type(),
		NewParameter("one", ScalarTypes.Bool.Type()),
		NewParameter("two", ScalarTypes.Bool.Type()),
	)
	arguments, argumentTypes := overloadResolutionInputs(ScalarTypes.Long.Type(), ScalarTypes.String.Type())

	withoutRequirement := GetBestMatchingSignatures([]*Signature{partial, none}, arguments, argumentTypes, false)
	assertMatchingSignatures(t, withoutRequirement, partial)

	withRequirement := GetBestMatchingSignatures([]*Signature{partial, none}, arguments, argumentTypes, true)
	assertMatchingSignatures(t, withRequirement)
}

func TestGetCommonSignatureReturnType(t *testing.T) {
	longSignature := NewSignature(ScalarTypes.Long.Type())
	secondLongSignature := NewSignature(ScalarTypes.Long.Type())
	stringSignature := NewSignature(ScalarTypes.String.Type())

	if got := GetCommonSignatureReturnType(nil, nil, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != &Error.TypeSymbol {
		t.Fatalf("empty common return = %#v, want error", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature}, nil, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != ScalarTypes.Long.Type() {
		t.Fatalf("single common return = %#v, want long", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature, secondLongSignature}, nil, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != ScalarTypes.Long.Type() {
		t.Fatalf("equal common return = %#v, want long", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature, stringSignature}, nil, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != &Error.TypeSymbol {
		t.Fatalf("incompatible common return = %#v, want error", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature, stringSignature}, []*TypeSymbol{&Error.TypeSymbol}, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != ScalarTypes.Unknown.Type() {
		t.Fatalf("error-argument common return = %#v, want unknown", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature, stringSignature}, []*TypeSymbol{ScalarTypes.Unknown.Type()}, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != ScalarTypes.Unknown.Type() {
		t.Fatalf("unknown-argument common return = %#v, want unknown", got)
	}
	if got := GetCommonSignatureReturnType([]*Signature{longSignature, secondLongSignature}, []*TypeSymbol{&Error.TypeSymbol}, func(signature *Signature) *TypeSymbol { return signature.DeclaredReturnType }); got != ScalarTypes.Long.Type() {
		t.Fatalf("matching returns with error argument = %#v, want long", got)
	}
}

func overloadResolutionInputs(types ...Symbol) ([]Expression, []Symbol) {
	return make([]Expression, len(types)), types
}

func overloadResolutionCallArgument(t *testing.T, text string) Expression {
	t.Helper()
	call, ok := ParseExpression(text).(*FunctionCallExpression)
	if !ok {
		t.Fatalf("expression %q did not parse as function call", text)
	}
	return call.ArgumentList().Expressions().At(0).Element()
}

func assertMatchingSignatures(t *testing.T, got []*Signature, want ...*Signature) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("matching signature count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("matching signature %d = %p, want %p", i, got[i], want[i])
		}
	}
}
