package kustolang

import "testing"

func TestSignatureArgumentParametersForTypesFixedLayout(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second)

	got := signature.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.String.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.Real.Type(),
	})

	if len(got) != 3 || got[0] != first || got[1] != second || got[2] != UnknownParameter {
		t.Fatalf("fixed argument parameters = %#v", got)
	}
	next := signature.NextPossibleParametersForTypes([]*TypeSymbol{ScalarTypes.String.Type()})
	if len(next) != 1 || next[0] != second {
		t.Fatalf("fixed next possible parameters = %#v", next)
	}
}

func TestSignatureArgumentParametersForTypesRepeatingUsesTypeMatch(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type(), WithParameterOccurrence(1, MaxRepeat))
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), value, limit)

	got := signature.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.String.Type(),
		ScalarTypes.String.Type(),
		ScalarTypes.Long.Type(),
	})

	if len(got) != 3 || got[0] != value || got[1] != value || got[2] != limit {
		t.Fatalf("repeating argument parameters = %#v", got)
	}
	next := signature.NextPossibleParametersForTypes([]*TypeSymbol{ScalarTypes.String.Type()})
	if len(next) != 2 || next[0] != value || next[1] != limit {
		t.Fatalf("repeating next possible parameters = %#v", next)
	}
}

func TestSignatureArgumentParametersRepeatingSkippingOptional(t *testing.T) {
	optionalText := NewParameter("text", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1))
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	skipping := NewSignature(ScalarTypes.Bool.Type(), optionalText, limit).WithLayout(ParameterLayoutRepeatingSkip)
	nonSkipping := NewSignature(ScalarTypes.Bool.Type(), optionalText, limit).WithLayout(ParameterLayoutRepeating)

	skipped := skipping.ArgumentParametersForTypes([]*TypeSymbol{ScalarTypes.Long.Type()})
	if len(skipped) != 1 || skipped[0] != limit {
		t.Fatalf("repeating-skipping parameters = %#v", skipped)
	}
	notSkipped := nonSkipping.ArgumentParametersForTypes([]*TypeSymbol{ScalarTypes.Long.Type()})
	if len(notSkipped) != 1 || notSkipped[0] != optionalText {
		t.Fatalf("repeating non-skipping parameters = %#v", notSkipped)
	}
}

func TestSignatureArgumentParametersForSymbolsMatchesTabular(t *testing.T) {
	table := NewTableSymbol("Rows", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	rows := NewParameterSymbolType("rows", table)
	name := NewParameter("name", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), rows, name)

	got := signature.ArgumentParametersForSymbols([]Symbol{table, ScalarTypes.String.Type()})
	if len(got) != 2 || got[0] != rows || got[1] != name {
		t.Fatalf("symbol argument parameters = %#v", got)
	}
}

func TestSignatureArgumentParametersBlockRepeatingUnknownOverflow(t *testing.T) {
	prefix := NewParameter("prefix", ScalarTypes.String.Type())
	repeatingA := NewParameter("a", ScalarTypes.Long.Type(), WithParameterOccurrence(1, 2))
	repeatingB := NewParameter("b", ScalarTypes.Real.Type(), WithParameterOccurrence(1, 2))
	suffix := NewParameter("suffix", ScalarTypes.Guid.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), prefix, repeatingA, repeatingB, suffix).WithLayout(ParameterLayoutBlockRepeating)

	got := signature.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.String.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.Real.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.Real.Type(),
		ScalarTypes.Guid.Type(),
		ScalarTypes.String.Type(),
	})

	if len(got) != 7 ||
		got[0] != prefix ||
		got[1] != repeatingA ||
		got[2] != repeatingB ||
		got[3] != repeatingA ||
		got[4] != repeatingB ||
		got[5] != suffix ||
		got[6] != UnknownParameter {
		t.Fatalf("block-repeating argument parameters = %#v", got)
	}
	next := signature.NextPossibleParametersForTypes([]*TypeSymbol{
		ScalarTypes.String.Type(),
		ScalarTypes.Long.Type(),
		ScalarTypes.Real.Type(),
	})
	if len(next) != 2 || next[0] != repeatingA || next[1] != suffix {
		t.Fatalf("block-repeating next possible parameters = %#v", next)
	}
}

func TestSignatureArgumentParametersForExpressionsUsesNamedArgumentsWhenAllowed(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second)
	NewFunctionSymbol("local", signature)
	argument := functionCallArgument(t, `local(second=1)`, 0)

	got := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.Long.Type()})
	if len(got) != 1 || got[0] != second {
		t.Fatalf("named argument parameters = %#v", got)
	}
}

func TestBlockRepeatingWithoutRepeatsPreservesNamedArguments(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second).
		WithLayout(ParameterLayoutBlockRepeating)
	NewFunctionSymbol("local", signature)
	argument := functionCallArgument(t, `local(second=1)`, 0)

	got := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.Long.Type()})
	if len(got) != 1 || got[0] != second {
		t.Fatalf("block layout named argument parameters = %#v", got)
	}
}

func TestCustomLayoutPreservesSurplusBuilderParameters(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())
	third := NewParameter("third", ScalarTypes.Bool.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second, third).
		WithCustomLayout(func(_ *Signature, _ []Expression, _ []Symbol) []*Parameter {
			return []*Parameter{first, second, third}
		})

	got := signature.ArgumentParametersForTypes([]*TypeSymbol{ScalarTypes.String.Type()})
	if len(got) != 3 || got[0] != first || got[1] != second || got[2] != third {
		t.Fatalf("custom surplus parameters = %#v", got)
	}
}

func TestExpressionLayoutInfersTypesWithoutSideChannel(t *testing.T) {
	optionalText := NewParameter("text", ScalarTypes.String.Type(), WithParameterOccurrence(0, 1))
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), optionalText, limit).
		WithLayout(ParameterLayoutRepeatingSkip)
	argument := FakeExpressionCreate(ScalarTypes.Long.Type())

	got := signature.ArgumentParametersForExpressions([]Expression{argument}, nil)
	if len(got) != 1 || got[0] != limit {
		t.Fatalf("inferred expression argument parameters = %#v", got)
	}
}

func TestSignatureGetParameterUsesLastDuplicateName(t *testing.T) {
	first := NewParameter("value", ScalarTypes.String.Type())
	second := NewParameter("value", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second)
	NewFunctionSymbol("local", signature)

	if got := signature.GetParameter("value"); got != second {
		t.Fatalf("GetParameter duplicate = %s, want second declaration", parameterTestName(got))
	}

	argument := functionCallArgument(t, `local(value=1)`, 0)
	got := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.Long.Type()})
	if len(got) != 1 || got[0] != second {
		t.Fatalf("named duplicate argument parameters = %#v", got)
	}
}

func TestSignatureArgumentParametersForExpressionsRejectsBuiltInNamedArguments(t *testing.T) {
	signature := Functions.Strlen.Signatures()[0]
	argument := functionCallArgument(t, `strlen(not_a_parameter="abc")`, 0)

	got := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.String.Type()})
	if len(got) != 1 || got[0] != signature.Parameters[0] {
		t.Fatalf("built-in named argument parameters = %s, want %s; allowsNamed=%t", parameterTestName(got[0]), parameterTestName(signature.Parameters[0]), signature.AllowsNamedArguments())
	}
	if signature.AllowsNamedArguments() {
		t.Fatalf("built-in function signature should not allow named arguments")
	}
}

func TestSignatureArgumentParametersForExpressionsUsesActiveGlobalsForBuiltInOwnership(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), first, second)
	fn := NewFunctionSymbol("active_builtin", signature)
	argument := functionCallArgument(t, `active_builtin(second=1)`, 0)

	localMapping := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.Long.Type()})
	if len(localMapping) != 1 || localMapping[0] != second {
		t.Fatalf("default mapping = %#v, want named second parameter", localMapping)
	}

	globals := NewGlobalState(nil).WithFunctions([]*FunctionSymbol{fn})
	builtInMapping := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.Long.Type()}, globals)
	if len(builtInMapping) != 1 || builtInMapping[0] != first {
		t.Fatalf("active-global mapping = %#v, want positional first parameter", builtInMapping)
	}
	if signature.AllowsNamedArguments(globals) {
		t.Fatalf("active-global built-in signature should not allow named arguments")
	}
}

func TestSignatureArgumentParametersForExpressionsHandlesStarOnly(t *testing.T) {
	star := NewParameterKind("star", ParameterTypeScalar, WithArgumentKind(ArgumentKindStarOnly), WithParameterOccurrence(0, 1))
	text := NewParameter("text", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), star, text).WithLayout(ParameterLayoutRepeatingSkip)
	starArg := functionCallArgument(t, `f(*)`, 0)
	textArg := functionCallArgument(t, `f("abc")`, 0)

	gotStar := signature.ArgumentParametersForExpressions([]Expression{starArg}, []Symbol{ScalarTypes.String.Type()})
	if len(gotStar) != 1 || gotStar[0] != star {
		t.Fatalf("star argument parameters = %#v", gotStar)
	}
	gotText := signature.ArgumentParametersForExpressions([]Expression{textArg}, []Symbol{ScalarTypes.String.Type()})
	if len(gotText) != 1 || gotText[0] != text {
		t.Fatalf("non-star argument parameters = %#v", gotText)
	}
}

func TestSignatureArgumentParametersForExpressionsUsesDefaultValueIndicator(t *testing.T) {
	defaultedLong := NewParameter("duration", ScalarTypes.Long.Type(), WithParameterDefaultValueIndicator("auto"))
	text := NewParameter("text", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), defaultedLong, text).WithLayout(ParameterLayoutRepeating)
	argument := functionCallArgument(t, `f("auto")`, 0)

	got := signature.ArgumentParametersForExpressions([]Expression{argument}, []Symbol{ScalarTypes.String.Type()})
	if len(got) != 1 || got[0] != defaultedLong {
		t.Fatalf("default-indicator argument parameters = %#v", got)
	}
}

func functionCallArgument(t *testing.T, text string, index int) Expression {
	t.Helper()
	call, ok := ParseExpression(text).(*FunctionCallExpression)
	if !ok {
		t.Fatalf("expression %q did not parse as function call", text)
	}
	arguments := call.ArgumentList().Expressions()
	if index < 0 || index >= arguments.Count() {
		t.Fatalf("argument index %d out of %d", index, arguments.Count())
	}
	return arguments.At(index).Element()
}

func parameterTestName(parameter *Parameter) string {
	if parameter == nil {
		return "<nil>"
	}
	if parameter == UnknownParameter {
		return "<unknown>"
	}
	return parameter.Name
}
