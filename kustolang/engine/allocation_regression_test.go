package kustolang

import (
	"strings"
	"testing"
)

func TestLexAllocationsDoNotScalePerToken(t *testing.T) {
	text := strings.Repeat("x + ", 1000) + "x"
	if tokens := Lex(text); len(tokens) != 2002 {
		t.Fatalf("got %d tokens, want 2002 including the end token", len(tokens))
	}
	if allocations := testing.AllocsPerRun(10, func() { Lex(text) }); allocations > 20 {
		t.Fatalf("plain tokenization allocated %v times; expected buffer growth, not a heap object per token", allocations)
	}
}

func TestLexLongInitialTriviaAndLiteralKeepSmallBuffers(t *testing.T) {
	for _, text := range []string{"//" + strings.Repeat("x", 1<<20) + "\nx", "'" + strings.Repeat("x", 1<<20) + "'"} {
		tokens := Lex(text)
		if len(tokens) != 2 || cap(tokens) > 4 {
			t.Fatalf("single-token input retained buffer len=%d cap=%d", len(tokens), cap(tokens))
		}
	}
	options := DefaultParseOptions
	options.AlwaysProduceEndToken = false
	if tokens := LexWithOptions("", options); tokens != nil {
		t.Fatal("empty input without an end token must preserve the nil result")
	}
}

func TestSyntaxMetadataReadsPreservePublicOwnership(t *testing.T) {
	node := ParseExpression("x + 1").(*BinaryExpression).SyntaxNode
	name, optional, hint := node.GetName(0), node.IsOptional(0), node.CompletionHint(0)
	metadata, ok := GeneratedSyntaxNodeMetadataForName(node.GeneratedNodeName())
	if !ok || len(metadata.Children) == 0 {
		t.Fatal("missing binary metadata")
	}
	metadata.Children[0].Name = "changed"
	metadata.Children[0].Optional = !optional
	metadata.Children[0].Hint = CompletionHintNone
	if node.GetName(0) != name || node.IsOptional(0) != optional || node.CompletionHint(0) != hint {
		t.Fatal("public metadata mutation changed an existing node")
	}
	if allocations := testing.AllocsPerRun(20, func() {
		node.GetName(0)
		node.IsOptional(0)
		node.CompletionHint(0)
	}); allocations != 0 {
		t.Fatalf("internal metadata reads allocated %v times", allocations)
	}
}

func BenchmarkLexerAllocationWorkloads(b *testing.B) {
	for _, test := range []struct{ name, text string }{
		{"short", "T | where User has 'test' | project User"},
		{"dense", strings.Repeat("x + ", 1000) + "1"},
		{"long_literal", "'" + strings.Repeat("x", 1<<16) + "'"},
		{"long_comment", "//" + strings.Repeat("x", 1<<16) + "\nx"},
	} {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				Lex(test.text)
			}
		})
	}
}
