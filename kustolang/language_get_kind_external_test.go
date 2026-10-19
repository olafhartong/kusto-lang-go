package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicGetKindSpellingMatchesCodeKindOf(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "valid query", text: `Events | count`},
		{name: "valid command", text: `.show tables`},
		{name: "malformed query", text: `Events | where`},
		{name: "malformed command", text: `.show`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got, want := kustolang.GetKind(test.text), kustolang.CodeKindOf(test.text); got != want {
				t.Fatalf("GetKind(%q) = %v, want %v", test.text, got, want)
			}
		})
	}
}
