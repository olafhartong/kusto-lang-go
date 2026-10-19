package kustolang

import "testing"

func TestStringFilterReturnsInputWhenAllCharactersMatch(t *testing.T) {
	got := StringFilter("abc", func(ch rune) bool {
		return ch >= 'a' && ch <= 'z'
	})
	if got != "abc" {
		t.Fatalf("StringFilter all-match = %q, want abc", got)
	}
}

func TestStringFilterRemovesCharactersThatDoNotMatch(t *testing.T) {
	got := StringFilter("a1b2c3", func(ch rune) bool {
		return ch >= 'a' && ch <= 'z'
	})
	if got != "abc" {
		t.Fatalf("StringFilter filtered = %q, want abc", got)
	}
}

func TestStringFilterPreservesMatchingRunes(t *testing.T) {
	got := StringFilter("aβ1ç2", func(ch rune) bool {
		return ch == 'β' || ch == 'ç'
	})
	if got != "βç" {
		t.Fatalf("StringFilter unicode = %q, want βç", got)
	}
}
