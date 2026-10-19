package kustolang

import "testing"

// Ported from src/Kusto.Language/Symbols/DatabaseSymbol.cs at the pinned
// source revision. See third_party/kusto/LICENSE.
func TestDatabaseSymbolContainsPinnedSourceBehavior(t *testing.T) {
	member := NewTableSymbol("Member")
	firstProbe := NewTableSymbol("FirstProbe")
	database := NewDatabaseSymbol("db", member)

	if !database.Contains(firstProbe) {
		t.Fatal("first query against a non-empty database should populate the pinned cache with that query")
	}
	if database.Contains(member) {
		t.Fatal("a different symbol should not be present after the pinned cache has been initialized")
	}
	if !database.Contains(firstProbe) {
		t.Fatal("the first queried identity should remain cached")
	}
}

func TestDatabaseSymbolContainsEmptyAndFirstMemberCases(t *testing.T) {
	member := NewTableSymbol("Member")
	if NewDatabaseSymbol("empty").Contains(member) {
		t.Fatal("an empty database should not cache the queried symbol")
	}

	database := NewDatabaseSymbol("db", member, NewFunctionSymbol("F"))
	if !database.Contains(member) {
		t.Fatal("a member queried first should be cached")
	}
	if database.Contains(NewTableSymbol("Other")) {
		t.Fatal("a later non-member should not be cached")
	}
}

func TestDatabaseSymbolContainsCachesNullWhenQueriedFirst(t *testing.T) {
	database := NewDatabaseSymbol("db", NewTableSymbol("Member"))
	if !database.Contains(nil) {
		t.Fatal("the pinned HashSet behavior should cache a null first query")
	}
	if database.Contains(NewTableSymbol("Later")) {
		t.Fatal("a later symbol should not replace the cached null identity")
	}
	if !database.Contains(nil) {
		t.Fatal("the cached null identity should remain present")
	}
}
