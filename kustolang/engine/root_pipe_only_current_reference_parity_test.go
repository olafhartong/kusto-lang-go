package kustolang

import "testing"

func TestCurrentDeclarationRootCountRecoversAsNameReference(t *testing.T) {
	for _, text := range []string{
		"let X = count; X | project Count",
		"let X = () { count }; X | project Count",
	} {
		query := ParseQuery(text)
		first := query.Statements().At(0)
		if statement, ok := first.Element().(*LetStatement); ok {
			t.Logf("let expression for %q = %T", text, statement.Expression())
		}
		var recovered *NameReference
		WalkSyntax(query, func(element SyntaxElement) {
			if reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok &&
				NameReferenceSimpleName(reference) == "count" && recovered == nil {
				recovered = reference
			}
		})
		if recovered == nil {
			t.Fatalf("%q did not recover count as NameReference", text)
		}
		NewBinder(nil).TryBind(query)
		info, ok := recovered.GetSemanticInfo()
		if !ok || len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS142" ||
			!sameType(info.ResultType, ScalarTypes.Unknown.Type()) {
			t.Fatalf("%q count semantics = %#v, want unresolved unknown scalar", text, info)
		}
	}
}

func TestCurrentPipeOnlyRecoveryControls(t *testing.T) {
	letWhere := ParseQuery("let X = where; X | count")
	var rootWhere *FilterOperator
	WalkSyntax(letWhere, func(element SyntaxElement) {
		if filter, ok := WrapGeneratedSyntaxElement(element).(*FilterOperator); ok {
			rootWhere = filter
		}
	})
	if rootWhere == nil {
		t.Fatal("let root where did not remain a FilterOperator")
	}

	piped := ParseQuery("T | count")
	pipe, ok := WrapGeneratedSyntaxElement(piped.Statements().At(0).Element()).(*ExpressionStatement)
	if !ok {
		t.Fatalf("piped count statement = %T", piped.Statements().At(0).Element())
	}
	pipeExpression, ok := pipe.Expression().(*PipeExpression)
	if !ok {
		t.Fatalf("piped count expression = %T", pipe.Expression())
	}
	if _, ok := pipeExpression.Operator().(*CountOperator); !ok {
		t.Fatalf("piped count operator = %T", pipeExpression.Operator())
	}
}
