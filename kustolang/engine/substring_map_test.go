package kustolang

import "testing"

func TestSubstringMapFindsLongestMatch(t *testing.T) {
	m := NewSubstringMap([]SubstringMapEntry[int]{
		{Key: "a", Value: 1},
		{Key: "ab", Value: 2},
		{Key: "abc", Value: 3},
	})
	match := m.GetLongestMatch("abcdef", 0)
	if match.Key != "abc" || match.Value != 3 {
		t.Fatalf("match = %#v, want abc/3", match)
	}
}

func TestSubstringMapMatchesFromStartOffset(t *testing.T) {
	m := NewSubstringMap([]SubstringMapEntry[string]{
		{Key: "where", Value: "keyword"},
		{Key: "when", Value: "other"},
	})
	match := m.GetLongestMatch("xxwhere y", 2)
	if match.Key != "where" || match.Value != "keyword" {
		t.Fatalf("offset match = %#v, want where/keyword", match)
	}
}

func TestSubstringMapReturnsEmptyEntryForNoMatch(t *testing.T) {
	m := NewSubstringMap([]SubstringMapEntry[int]{{Key: "abc", Value: 1}})
	for _, start := range []int{-1, 0, 3, 10} {
		match := m.GetLongestMatch("zzz", start)
		if match.Key != "" || match.Value != 0 {
			t.Fatalf("no-match at %d = %#v, want empty", start, match)
		}
	}
}

func TestSubstringMapDuplicateKeyLastValueWins(t *testing.T) {
	m := NewSubstringMap([]SubstringMapEntry[string]{
		{Key: "key", Value: "first"},
		{Key: "key", Value: "second"},
	})
	match := m.GetLongestMatch("key", 0)
	if match.Key != "key" || match.Value != "second" {
		t.Fatalf("duplicate match = %#v, want second", match)
	}
}

func TestSubstringMapEmptyKeyIsNotReturnedByLongestMatch(t *testing.T) {
	m := NewSubstringMap([]SubstringMapEntry[string]{
		{Key: "", Value: "empty"},
		{Key: "x", Value: "x"},
	})
	noMatch := m.GetLongestMatch("abc", 0)
	if noMatch.Key != "" || noMatch.Value != "" {
		t.Fatalf("empty-key no-match = %#v, want empty default", noMatch)
	}
	match := m.GetLongestMatch("x", 0)
	if match.Key != "x" || match.Value != "x" {
		t.Fatalf("regular match = %#v, want x", match)
	}
}
