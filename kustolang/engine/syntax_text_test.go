package kustolang

import (
	"reflect"
	"strings"
	"testing"
)

// Preserve the pre-optimization renderer as an independent compatibility oracle.
func legacySyntaxTokenText(t *SyntaxToken, trivia IncludeTrivia, before, after bool) string {
	var b strings.Builder
	if t.trivia != "" {
		switch trivia {
		case IncludeTriviaAll:
			b.WriteString(t.trivia)
		case IncludeTriviaInterior:
			if !before {
				b.WriteString(t.trivia)
			}
		case IncludeTriviaMinimal:
			if !before && !after {
				if strings.Contains(t.trivia, "\n") {
					b.WriteString("\n")
				} else {
					b.WriteString(" ")
				}
			}
		case IncludeTriviaSingleLine:
			if !before && !after {
				b.WriteString(" ")
			}
		}
	}
	b.WriteString(t.text)
	return b.String()
}

func legacySyntaxText(element SyntaxElement, trivia IncludeTrivia) string {
	if token, ok := element.(*SyntaxToken); ok {
		return legacySyntaxTokenText(token, trivia, false, token.kind == SyntaxKindEndOfTextToken)
	}
	var b strings.Builder
	var children []SyntaxElement
	switch node := element.(type) {
	case interface{ syntaxNodeRoot() *SyntaxNode }:
		children = node.syntaxNodeRoot().children
	case *SyntaxList:
		children = node.children()
	case *SeparatedElement:
		children = node.children()
	default:
		children = element.Children()
	}
	for i, child := range children {
		if child == nil {
			continue
		}
		if token, ok := child.(*SyntaxToken); ok {
			b.WriteString(legacySyntaxTokenText(token, trivia, i == 0, i == len(children)-1))
		} else {
			b.WriteString(legacySyntaxText(child, trivia))
		}
	}
	return b.String()
}

func TestSyntaxTextPreservesAllTriviaAndSubtrees(t *testing.T) {
	for _, text := range []string{
		"", " // trailing only\r\n", "// 😀\r\n let x = 'é'; T | where x == 'é' // tail\n",
		"print x = f(1, // comment\n 2), y = dynamic([1,2])", "T | project x,", "print f(a=, b=2,,)",
		"print x='unterminated", "print x=\xff", "T | where x between (1 .. )", "restrict access to (a,,b); T",
	} {
		root := ParseQuery(text)
		for _, node := range GetDescendantsOrSelf[SyntaxElement](root) {
			for _, trivia := range []IncludeTrivia{IncludeTriviaAll, IncludeTriviaInterior, IncludeTriviaMinimal, IncludeTriviaSingleLine, IncludeTrivia(-1)} {
				want, got := legacySyntaxText(node, trivia), node.Text(trivia)
				if got != want {
					t.Fatalf("%q node %T mode %d: got %q, want %q", text, node, trivia, got, want)
				}
			}
		}
	}
	// Optional nil slots must still participate in first/last child boundaries.
	for _, trivia := range []IncludeTrivia{IncludeTriviaAll, IncludeTriviaInterior, IncludeTriviaMinimal, IncludeTriviaSingleLine} {
		root := NewSyntaxNode(SyntaxKindNone, []SyntaxElement{nil, NewIdentifierToken(" // before\n", "x"), nil})
		if got, want := root.Text(trivia), legacySyntaxText(root, trivia); got != want {
			t.Fatalf("nil-slot boundary: got %q, want %q", got, want)
		}
	}
}

type customTextSyntaxNode struct{ *SyntaxNode }

func (*customTextSyntaxNode) Text(IncludeTrivia) string { return "custom node" }

type customTextGeneratedNode struct{ *NameReference }

func (*customTextGeneratedNode) Text(IncludeTrivia) string { return "custom generated node" }

type customTextSyntaxToken struct{ *SyntaxToken }

func (*customTextSyntaxToken) Text(IncludeTrivia) string { return "custom token" }

func TestSyntaxTextPreservesCustomOverrides(t *testing.T) {
	custom := &customTextSyntaxNode{NewSyntaxNode(SyntaxKindNone, nil)}
	generated := &customTextGeneratedNode{ParseExpression("x").(*NameReference)}
	token := &customTextSyntaxToken{NewIdentifierToken("", "hidden")}
	root := NewSyntaxList(custom, generated, token)
	if got := root.Text(IncludeTriviaAll); got != "custom nodecustom generated nodecustom token" {
		t.Fatalf("custom Text bypassed: %q", got)
	}
	// Typed list wrappers retain their existing Text dispatch, too.
	typed := AsSyntaxListOf[*SyntaxToken](NewSyntaxList(NewIdentifierToken(" ", "x")))
	outer := NewSyntaxList(typed)
	if got, want := outer.Text(IncludeTriviaMinimal), typed.Text(IncludeTriviaMinimal); got != want {
		t.Fatalf("typed list = %q, want %q", got, want)
	}
}

func TestSyntaxTextGeneratedRegistryCoversWrappers(t *testing.T) {
	if len(generatedSyntaxTextTypes) < 100 {
		t.Fatal("generated renderer registry was not initialized")
	}
	root := ParseExpression("f(x + 1)")
	if _, exists := generatedSyntaxTextTypes[reflect.TypeOf(root)]; !exists {
		t.Fatal("generated call missing from renderer registry")
	}
}

func BenchmarkSyntaxTextNested(b *testing.B) {
	root := ParseExpression(strings.Repeat("f(", 40) + "x" + strings.Repeat(")", 40))
	for _, legacy := range []bool{true, false} {
		name := "single-buffer"
		if legacy {
			name = "legacy"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if legacy {
					_ = legacySyntaxText(root, IncludeTriviaAll)
				} else {
					_ = root.Text(IncludeTriviaAll)
				}
			}
		})
	}
}
