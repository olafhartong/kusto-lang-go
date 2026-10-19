package kustolang

import "testing"

func TestBuiltInOperatorsCatalog(t *testing.T) {
	if len(Operators.All) != 55 {
		t.Fatalf("built-in operator count = %d, want 55", len(Operators.All))
	}
	if Operators.All[0] != Operators.UnaryMinus || Operators.All[len(Operators.All)-1] != Operators.HasAll {
		t.Fatalf("built-in operator order does not match upstream")
	}

	seen := map[OperatorKind]bool{}
	for _, op := range Operators.All {
		if op == nil {
			t.Fatal("built-in operators must not contain nil entries")
		}
		if op.SymbolKind() != SymbolKindOperator || op.SymbolTabularity() != TabularityScalar {
			t.Fatalf("%s metadata = kind:%s tabularity:%s", op.SymbolName(), op.SymbolKind(), op.SymbolTabularity())
		}
		if op.SymbolName() != string(op.OperatorKind) {
			t.Fatalf("operator symbol name = %q, kind = %q", op.SymbolName(), op.OperatorKind)
		}
		if seen[op.OperatorKind] {
			t.Fatalf("duplicate built-in operator %s", op.OperatorKind)
		}
		seen[op.OperatorKind] = true
		for _, signature := range op.Signatures() {
			if signature.Symbol != op {
				t.Fatalf("%s signature should point back to operator", op.OperatorKind)
			}
		}
	}
}

func TestOperatorSignaturesMatchRepresentativeUpstreamShapes(t *testing.T) {
	if got := Operators.Add.Signatures(); len(got) != 8 {
		t.Fatalf("Add signature count = %d, want 8", len(got))
	}
	if got := Operators.Subtract.Signatures(); len(got) != 12 {
		t.Fatalf("Subtract signature count = %d, want 12", len(got))
	}
	if got := Operators.LessThan.Signatures(); len(got) != 7 || !got[0].IsHidden {
		t.Fatalf("LessThan hidden/signature shape changed: len=%d hidden0=%t", len(got), got[0].IsHidden)
	}

	has := Operators.Has.Signatures()
	if len(has) != 2 {
		t.Fatalf("Has signature count = %d, want 2", len(has))
	}
	if right := has[0].Parameters[1]; right.TypeKind != ParameterTypeDeclared || right.DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("Has right parameter should be declared string, got kind=%s types=%#v", right.TypeKind, right.DeclaredTypes)
	}
	if left := has[1].Parameters[0]; left.ArgumentKind != ArgumentKindStarOnly {
		t.Fatalf("Has star overload left argument kind = %s", left.ArgumentKind)
	}

	matchRegex := Operators.MatchRegex.Signatures()[0]
	if matchRegex.Parameters[1].TypeKind != ParameterTypeStringOrDynamic {
		t.Fatalf("MatchRegex dynamic RHS shape was not preserved")
	}
}

func TestNaryOperatorRepeatableParameters(t *testing.T) {
	in := Operators.In.Signatures()
	if len(in) != 4 {
		t.Fatalf("In signature count = %d, want 4", len(in))
	}
	if !in[0].IsHidden || !in[2].IsHidden {
		t.Fatalf("bool In overloads should be hidden")
	}
	repeat := in[3].Parameters[1]
	if repeat.MinOccurring != 1 || repeat.MaxOccurring != MaxRepeat || !in[3].HasRepeatableParameters {
		t.Fatalf("In repeatable parameter = %d..%d repeat=%t", repeat.MinOccurring, repeat.MaxOccurring, in[3].HasRepeatableParameters)
	}
	if in[1].Parameters[1].Tabularity() != TabularityTabular {
		t.Fatalf("In tabular parameter should report tabular")
	}
}

func TestUnaryOperatorCustomReturnType(t *testing.T) {
	signature := Operators.UnaryMinus.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil {
		t.Fatalf("UnaryMinus should use custom return type")
	}

	nonConstant := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Int.Type()}})
	if nonConstant != ScalarTypes.Long.Type() {
		t.Fatalf("non-constant int unary return = %v, want long", nonConstant)
	}
	constant := signature.CustomReturnType(CustomReturnTypeContext{
		ArgumentTypes:      []*TypeSymbol{ScalarTypes.Int.Type()},
		ArgumentIsConstant: []bool{true},
	})
	if constant != ScalarTypes.Int.Type() {
		t.Fatalf("constant int unary return = %v, want int", constant)
	}
	real := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Real.Type()}})
	if real != ScalarTypes.Real.Type() {
		t.Fatalf("real unary return = %v, want real", real)
	}
}
