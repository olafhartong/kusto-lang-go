package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCompletionHintFlags(t *testing.T) {
	hints := []kustolang.CompletionHint{
		kustolang.CompletionHintInherit,
		kustolang.CompletionHintExpression,
		kustolang.CompletionHintScalar,
		kustolang.CompletionHintTabular,
		kustolang.CompletionHintGraphSnapshot,
	}
	seen := map[kustolang.CompletionHint]bool{}
	for _, hint := range hints {
		if hint == kustolang.CompletionHintNone || seen[hint] || hint&(hint-1) != 0 {
			t.Fatalf("completion hint is not a unique flag: %d", hint)
		}
		seen[hint] = true
	}
}
