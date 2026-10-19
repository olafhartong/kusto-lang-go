package kustolang

// Ported from the generic traversal surface on
// src/Kusto.Language/Syntax/SyntaxElement.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func GetFirstAncestorOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return internal.GetFirstAncestorOrSelf(element, predicate...)
}

func GetAncestors[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return internal.GetAncestors(element, predicate...)
}

func GetAncestorsOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return internal.GetAncestorsOrSelf(element, predicate...)
}

func GetFirstDescendantOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return internal.GetFirstDescendantOrSelf(element, predicate...)
}

func GetDescendantsOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return internal.GetDescendantsOrSelf(element, predicate...)
}
