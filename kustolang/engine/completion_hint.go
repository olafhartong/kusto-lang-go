package kustolang

// Ported from src/Kusto.Language/Editor/CompletionHint.cs.

type CompletionHint uint64

const (
	CompletionHintNone CompletionHint = 0

	CompletionHintInherit           CompletionHint = 1
	CompletionHintExpression        CompletionHint = CompletionHintInherit << 1
	CompletionHintScalar            CompletionHint = CompletionHintExpression << 1
	CompletionHintTabular           CompletionHint = CompletionHintScalar << 1
	CompletionHintNonScalar         CompletionHint = CompletionHintTabular << 1
	CompletionHintBoolean           CompletionHint = CompletionHintNonScalar << 1
	CompletionHintNumber            CompletionHint = CompletionHintBoolean << 1
	CompletionHintLiteral           CompletionHint = CompletionHintNumber << 1
	CompletionHintAggregate         CompletionHint = CompletionHintLiteral << 1
	CompletionHintTabularFunction   CompletionHint = CompletionHintAggregate << 1
	CompletionHintScalarFunction    CompletionHint = CompletionHintTabularFunction << 1
	CompletionHintDatabaseFunction  CompletionHint = CompletionHintScalarFunction << 1
	CompletionHintFunction          CompletionHint = CompletionHintDatabaseFunction << 1
	CompletionHintDeclaration       CompletionHint = CompletionHintFunction << 1
	CompletionHintColumn            CompletionHint = CompletionHintDeclaration << 1
	CompletionHintTable             CompletionHint = CompletionHintColumn << 1
	CompletionHintDatabase          CompletionHint = CompletionHintTable << 1
	CompletionHintCluster           CompletionHint = CompletionHintDatabase << 1
	CompletionHintEntityGroup       CompletionHint = CompletionHintCluster << 1
	CompletionHintGraph             CompletionHint = CompletionHintEntityGroup << 1
	CompletionHintSyntax            CompletionHint = CompletionHintGraph << 1
	CompletionHintQuery             CompletionHint = CompletionHintSyntax << 1
	CompletionHintCommand           CompletionHint = CompletionHintQuery << 1
	CompletionHintKeyword           CompletionHint = CompletionHintCommand << 1
	CompletionHintClause            CompletionHint = CompletionHintKeyword << 1
	CompletionHintMaterializedView  CompletionHint = CompletionHintClause << 1
	CompletionHintOption            CompletionHint = CompletionHintMaterializedView << 1
	CompletionHintExternalTable     CompletionHint = CompletionHintOption << 1
	CompletionHintStoredQueryResult CompletionHint = CompletionHintExternalTable << 1
	CompletionHintGraphModel        CompletionHint = CompletionHintStoredQueryResult << 1
	CompletionHintGraphSnapshot     CompletionHint = CompletionHintGraphModel << 1
)
