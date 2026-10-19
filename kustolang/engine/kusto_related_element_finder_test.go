package kustolang

import (
	"strings"
	"testing"
)

func TestKustoRelatedElementFinderMatchesStructuralReferenceBaselines(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		position     int
		wantCurrent  int
		wantElements []RelatedElement
	}{
		{
			name:         "outer parens from open",
			text:         `print x=(1 + (2))`,
			position:     8,
			wantCurrent:  0,
			wantElements: []RelatedElement{NewRelatedElement(8, 1, RelatedElementKindSyntax), NewRelatedElement(16, 1, RelatedElementKindSyntax)},
		},
		{
			name:         "outer parens from close end",
			text:         `print x=(1 + (2))`,
			position:     17,
			wantCurrent:  1,
			wantElements: []RelatedElement{NewRelatedElement(8, 1, RelatedElementKindSyntax), NewRelatedElement(16, 1, RelatedElementKindSyntax)},
		},
		{
			name:         "inner parens",
			text:         `print x=(1 + (2))`,
			position:     13,
			wantCurrent:  0,
			wantElements: []RelatedElement{NewRelatedElement(13, 1, RelatedElementKindSyntax), NewRelatedElement(15, 1, RelatedElementKindSyntax)},
		},
		{
			name:         "dynamic array brackets",
			text:         `print x=dynamic([1, 2])`,
			position:     16,
			wantCurrent:  0,
			wantElements: []RelatedElement{NewRelatedElement(16, 1, RelatedElementKindSyntax), NewRelatedElement(21, 1, RelatedElementKindSyntax)},
		},
		{name: "bracketed name excluded", text: `print ['x'] = 1`, position: 6},
		{name: "whitespace has no affinity", text: `print x = (1)`, position: 7},
		{
			name:         "UTF-16 positions",
			text:         "// 😀\nprint x=(1)",
			position:     14,
			wantCurrent:  0,
			wantElements: []RelatedElement{NewRelatedElement(14, 1, RelatedElementKindSyntax), NewRelatedElement(16, 1, RelatedElementKindSyntax)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewKustoCodeService(tt.text, nil)
			got := service.GetRelatedElements(tt.position, FindRelatedOptionsNone, CancellationToken{})
			if got.CurrentIndex != tt.wantCurrent || len(got.Elements) != len(tt.wantElements) {
				t.Fatalf("related info = %#v, want current=%d elements=%#v", got, tt.wantCurrent, tt.wantElements)
			}
			for i, want := range tt.wantElements {
				actual := got.Elements[i]
				if actual.Start() != want.Start() || actual.Length() != want.Length() || actual.Kind != want.Kind ||
					actual.CursorLeft != want.CursorLeft || actual.CursorRight != want.CursorRight {
					t.Fatalf("element %d = %#v, want %#v", i, actual, want)
				}
			}
		})
	}
}

func TestKustoRelatedElementFinderMatchesFunctionBodyBraces(t *testing.T) {
	text := `let f=(){print x=1}; f()`
	service := NewKustoCodeService(text, nil)
	got := service.GetRelatedElements(8, FindRelatedOptionsNone, CancellationToken{})
	if len(got.Elements) != 2 || got.Elements[0].Start() != 8 || got.Elements[1].Start() != 18 {
		t.Fatalf("brace related info = %#v", got)
	}
}

func TestKustoRelatedElementFinderMatchesBoundColumnDeclarationAndReferences(t *testing.T) {
	text := `datatable(Account:string)["a"] | where Account == "a" | project Account`
	service := NewKustoCodeService(text)
	selected := strings.Index(text, "Account ==") + 1
	got := service.GetRelatedElements(selected, FindRelatedOptionsNone, CancellationToken{})
	if len(got.Elements) != 3 || got.CurrentIndex != 1 {
		t.Fatalf("column related info = %#v", got)
	}
	wantStarts := []int{strings.Index(text, "Account"), strings.Index(text, "Account =="), strings.LastIndex(text, "Account")}
	for i, element := range got.Elements {
		wantKind := RelatedElementKindReference
		if i == 0 {
			wantKind = RelatedElementKindDeclaration
		}
		if element.Start() != wantStarts[i] || element.Length() != len("Account") || element.Kind != wantKind ||
			element.CursorLeft != wantStarts[i] || element.CursorRight != wantStarts[i] {
			t.Fatalf("column element %d = %#v, want start=%d kind=%v", i, element, wantStarts[i], wantKind)
		}
	}
}

func TestKustoRelatedElementFinderSeeThroughVariablesIsAsymmetric(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	text := `T | as X | union X, T`
	service := NewKustoCodeService(text, NewGlobalState(NewDatabaseSymbol("db", table)))

	xPosition := strings.Index(text, "X |")
	direct := service.GetRelatedElements(xPosition, FindRelatedOptionsNone, CancellationToken{})
	if len(direct.Elements) != 2 || direct.Elements[0].Kind != RelatedElementKindDeclaration || direct.Elements[1].Kind != RelatedElementKindReference {
		t.Fatalf("direct alias related info = %#v", direct)
	}
	throughAlias := service.GetRelatedElements(xPosition, FindRelatedOptionsSeeThroughVariables, CancellationToken{})
	if len(throughAlias.Elements) != 4 {
		t.Fatalf("see-through alias related info = %#v", throughAlias)
	}

	tPosition := strings.LastIndex(text, "T")
	throughTable := service.GetRelatedElements(tPosition, FindRelatedOptionsSeeThroughVariables, CancellationToken{})
	if len(throughTable.Elements) != 3 {
		t.Fatalf("see-through table related info = %#v", throughTable)
	}
	for _, element := range throughTable.Elements {
		if element.Start() == strings.Index(text, "X, T") {
			t.Fatalf("table-side see-through incorrectly included alias reference: %#v", throughTable)
		}
	}
}

func TestKustoRelatedElementFinderTreatsBracketedNamesSemantically(t *testing.T) {
	text := `datatable(['A B']:string)["x"] | project ['A B']`
	got := NewKustoCodeService(text).GetRelatedElements(strings.Index(text, "["), FindRelatedOptionsNone, CancellationToken{})
	if len(got.Elements) != 2 || got.Elements[0].Kind != RelatedElementKindDeclaration || got.Elements[1].Kind != RelatedElementKindReference {
		t.Fatalf("bracketed semantic related info = %#v", got)
	}
}

func TestKustoRelatedElementFinderSemanticRangesUseUTF16(t *testing.T) {
	text := "// 😀\n" + `datatable(Account:string)["a"] | project Account`
	position := byteOffsetToUTF16Position(text, strings.LastIndex(text, "Account"))
	got := NewKustoCodeService(text).GetRelatedElements(position, FindRelatedOptionsNone, CancellationToken{})
	wantDeclaration := byteOffsetToUTF16Position(text, strings.Index(text, "Account"))
	if len(got.Elements) != 2 || got.Elements[0].Start() != wantDeclaration || got.Elements[1].Start() != position {
		t.Fatalf("UTF-16 semantic related info = %#v", got)
	}
}

func TestKustoRelatedElementFinderMatchesLetAndFunctionParameterSymbols(t *testing.T) {
	letText := `let Threshold=10; print Value=Threshold`
	letRelated := NewKustoCodeService(letText).GetRelatedElements(strings.LastIndex(letText, "Threshold"), FindRelatedOptionsNone, CancellationToken{})
	if len(letRelated.Elements) != 2 || letRelated.Elements[0].Start() != strings.Index(letText, "Threshold") ||
		letRelated.Elements[0].Kind != RelatedElementKindDeclaration || letRelated.Elements[1].Kind != RelatedElementKindReference {
		t.Fatalf("let related info = %#v", letRelated)
	}

	functionText := `let F=(x:long) { x }; print Value=F(1)`
	parameterRelated := NewKustoCodeService(functionText).GetRelatedElements(strings.Index(functionText, "x }"), FindRelatedOptionsNone, CancellationToken{})
	if len(parameterRelated.Elements) != 2 || parameterRelated.Elements[0].Start() != strings.Index(functionText, "x:long") ||
		parameterRelated.Elements[0].Kind != RelatedElementKindDeclaration || parameterRelated.Elements[1].Kind != RelatedElementKindReference {
		t.Fatalf("function parameter related info = %#v", parameterRelated)
	}
}
