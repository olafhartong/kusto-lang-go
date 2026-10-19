package kustolang

// Ported from src/Kusto.Language/Symbols/Scope.cs, EmptyScope.cs,
// Syntax/FakeExpression.cs, and Utils/SubstringMap.cs. See
// third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type Scope = internal.Scope
type EmptyScope = internal.EmptyScope

var EmptyScopeInstance Scope = internal.EmptyScopeInstance

// GetScopeSymbols returns symbols matching a name and kind from a scope.
func GetScopeSymbols(scope Scope, name string, match SymbolMatch) []Symbol {
	return internal.GetScopeSymbols(scope, name, match)
}

// GetAllScopeSymbols is the Go spelling of Scope.GetSymbols(SymbolMatch, List).
func GetAllScopeSymbols(scope Scope, match SymbolMatch) []Symbol {
	return internal.GetScopeSymbols(scope, "", match)
}

// CreateFakeExpression creates a missing-token expression with an attached
// semantic result type.
func CreateFakeExpression(resultType *TypeSymbol) Expression {
	return internal.FakeExpressionCreate(resultType)
}

// CreateNamedFakeExpression creates a named fake expression with an attached
// semantic result type.
func CreateNamedFakeExpression(name string, resultType *TypeSymbol) Expression {
	return internal.FakeExpressionCreateNamed(name, resultType)
}

type SubstringMapEntry[T any] struct {
	Key   string
	Value T
}

// SubstringMap finds the longest configured key at a position in a string.
// Public positions are UTF-16 code units, matching Kusto.Language.
type SubstringMap[T any] struct {
	inner *internal.SubstringMap[T]
}

func NewSubstringMap[T any](entries []SubstringMapEntry[T]) *SubstringMap[T] {
	innerEntries := make([]internal.SubstringMapEntry[T], len(entries))
	for i, entry := range entries {
		innerEntries[i] = internal.SubstringMapEntry[T]{Key: entry.Key, Value: entry.Value}
	}
	return &SubstringMap[T]{inner: internal.NewSubstringMap(innerEntries)}
}

func (m *SubstringMap[T]) GetLongestMatch(text string, start int) SubstringMapEntry[T] {
	if start < 0 {
		panic(internal.ArgumentOutOfRangeError{ParameterName: "start"})
	}
	byteStart, ok := editStringTryUTF16Position(text, start)
	if m == nil || m.inner == nil || !ok {
		return SubstringMapEntry[T]{}
	}
	match := m.inner.GetLongestMatch(text, byteStart)
	return SubstringMapEntry[T]{Key: match.Key, Value: match.Value}
}
