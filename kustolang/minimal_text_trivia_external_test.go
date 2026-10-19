package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestKustoCodeServiceIncludeTrivia(t *testing.T) {
	tests := []struct {
		kind kustolang.MinimalTextKind
		want kustolang.IncludeTrivia
	}{
		{kustolang.MinimalTextKindRemoveLeadingWhitespaceAndComments, kustolang.IncludeTriviaInterior},
		{kustolang.MinimalTextKindMinimizeWhitespaceAndRemoveComments, kustolang.IncludeTriviaMinimal},
		{kustolang.MinimalTextKindSingleLine, kustolang.IncludeTriviaSingleLine},
	}
	for _, tt := range tests {
		if got := kustolang.KustoCodeServiceIncludeTrivia(tt.kind); got != tt.want {
			t.Fatalf("include trivia for %v = %v, want %v", tt.kind, got, tt.want)
		}
	}
}

func TestKustoCodeServiceIncludeTriviaRejectsUnknownKind(t *testing.T) {
	defer func() {
		value := recover()
		if _, ok := value.(kustolang.InvalidOperationError); !ok {
			t.Fatalf("panic = %#v, want InvalidOperationError", value)
		}
	}()
	kustolang.KustoCodeServiceIncludeTrivia(kustolang.MinimalTextKind(99))
}
