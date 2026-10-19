package kustolang

// Ported from src/Kusto.Language/Editor/CompletionHint.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

const (
	CompletionHintNone              = internal.CompletionHintNone
	CompletionHintInherit           = internal.CompletionHintInherit
	CompletionHintExpression        = internal.CompletionHintExpression
	CompletionHintScalar            = internal.CompletionHintScalar
	CompletionHintTabular           = internal.CompletionHintTabular
	CompletionHintNonScalar         = internal.CompletionHintNonScalar
	CompletionHintBoolean           = internal.CompletionHintBoolean
	CompletionHintNumber            = internal.CompletionHintNumber
	CompletionHintLiteral           = internal.CompletionHintLiteral
	CompletionHintAggregate         = internal.CompletionHintAggregate
	CompletionHintTabularFunction   = internal.CompletionHintTabularFunction
	CompletionHintScalarFunction    = internal.CompletionHintScalarFunction
	CompletionHintDatabaseFunction  = internal.CompletionHintDatabaseFunction
	CompletionHintFunction          = internal.CompletionHintFunction
	CompletionHintDeclaration       = internal.CompletionHintDeclaration
	CompletionHintColumn            = internal.CompletionHintColumn
	CompletionHintTable             = internal.CompletionHintTable
	CompletionHintDatabase          = internal.CompletionHintDatabase
	CompletionHintCluster           = internal.CompletionHintCluster
	CompletionHintEntityGroup       = internal.CompletionHintEntityGroup
	CompletionHintGraph             = internal.CompletionHintGraph
	CompletionHintSyntax            = internal.CompletionHintSyntax
	CompletionHintQuery             = internal.CompletionHintQuery
	CompletionHintCommand           = internal.CompletionHintCommand
	CompletionHintKeyword           = internal.CompletionHintKeyword
	CompletionHintClause            = internal.CompletionHintClause
	CompletionHintMaterializedView  = internal.CompletionHintMaterializedView
	CompletionHintOption            = internal.CompletionHintOption
	CompletionHintExternalTable     = internal.CompletionHintExternalTable
	CompletionHintStoredQueryResult = internal.CompletionHintStoredQueryResult
	CompletionHintGraphModel        = internal.CompletionHintGraphModel
	CompletionHintGraphSnapshot     = internal.CompletionHintGraphSnapshot
)
