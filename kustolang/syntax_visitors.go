package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxVisitor.cs. The concrete method
// set is generated from the pinned syntax model in language_generated.go.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// DefaultSyntaxVisitor forwards every non-overridden visit to DefaultVisit.
type DefaultSyntaxVisitor = internal.SyntaxVisitorFuncs

// DefaultSyntaxVisitorResult is the Go spelling of DefaultSyntaxVisitor<TResult>.
type DefaultSyntaxVisitorResult[T any] = internal.SyntaxVisitorResultFuncs[T]

func NewDefaultSyntaxVisitor(defaultVisit func(SyntaxElement)) DefaultSyntaxVisitor {
	return DefaultSyntaxVisitor{DefaultVisit: defaultVisit}
}

func NewDefaultSyntaxVisitorResult[T any](defaultVisit func(SyntaxElement) T) DefaultSyntaxVisitorResult[T] {
	return DefaultSyntaxVisitorResult[T]{DefaultVisit: defaultVisit}
}
