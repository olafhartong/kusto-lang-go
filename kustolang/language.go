package kustolang

// Ported from the public language, syntax, parser, binder, and symbol surface
// under src/Kusto.Language. See third_party/kusto/LICENSE for the preserved
// Apache-2.0 license and third_party/kusto/UPSTREAM.json for the pinned source.

//go:generate go run ../tools/gen-public-language

import (
	"context"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
	api "github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

// Core parsed-code and syntax model.
type CodeKind = internal.CodeKind
type KustoDialect = internal.KustoDialect
type ParseOptions = internal.ParseOptions
type ParserKind = internal.ParserKind
type LexicalToken = internal.LexicalToken
type SyntaxKind = internal.SyntaxKind
type SyntaxCategory = internal.SyntaxCategory
type IncludeTrivia = internal.IncludeTrivia
type DiagnosticsInclude = internal.DiagnosticsInclude
type CompletionHint = internal.CompletionHint
type SyntaxElement = internal.SyntaxElement
type SyntaxNode = internal.SyntaxNode
type SyntaxToken = internal.SyntaxToken
type TimeSpanValue = internal.TimeSpanValue
type SyntaxTree = internal.SyntaxTree
type SyntaxList = internal.SyntaxList
type SeparatedElement = internal.SeparatedElement
type QueryBlock = internal.QueryBlock
type SyntaxListOf[T SyntaxElement] = internal.SyntaxListOf[T]
type SeparatedElementOf[T SyntaxElement] = internal.SeparatedElementOf[T]
type SeparatedSyntaxListOf[T SyntaxElement] = internal.SeparatedSyntaxListOf[T]

// Semantic and symbol model.
type SemanticInfo = internal.SemanticInfo
type ValueInfo = internal.ValueInfo
type Symbol = internal.Symbol
type BaseSymbol = internal.BaseSymbol
type SymbolKind = internal.SymbolKind
type SymbolMatch = internal.SymbolMatch
type Tabularity = internal.Tabularity
type TypeSymbol = internal.TypeSymbol
type ScalarSymbol = internal.ScalarSymbol
type ColumnSymbol = internal.ColumnSymbol
type TableSymbol = internal.TableSymbol
type DatabaseSymbol = internal.DatabaseSymbol
type ClusterSymbol = internal.ClusterSymbol
type VariableSymbol = internal.VariableSymbol
type ParameterSymbol = internal.ParameterSymbol
type OptionSymbol = internal.OptionSymbol
type FunctionSymbol = internal.FunctionSymbol
type Signature = internal.Signature
type Parameter = internal.Parameter
type ParameterOption = internal.ParameterOption
type CustomReturnType = internal.CustomReturnType
type CustomReturnTypeContext = internal.CustomReturnTypeContext
type CustomAvailability = internal.CustomAvailability
type CustomAvailabilityContext = internal.CustomAvailabilityContext
type ParameterTypeKind = internal.ParameterTypeKind
type ReturnTypeKind = internal.ReturnTypeKind
type ArgumentKind = internal.ArgumentKind
type Conversion = internal.Conversion
type CommandSymbol = internal.CommandSymbol
type OperatorSymbol = internal.OperatorSymbol
type OperatorKind = internal.OperatorKind
type PatternSymbol = internal.PatternSymbol
type PatternSignature = internal.PatternSignature
type DynamicSymbol = internal.DynamicSymbol
type DynamicAnySymbol = internal.DynamicAnySymbol
type DynamicPrimitiveSymbol = internal.DynamicPrimitiveSymbol
type DynamicArraySymbol = internal.DynamicArraySymbol
type DynamicBagSymbol = internal.DynamicBagSymbol
type TupleSymbol = internal.TupleSymbol
type ErrorSymbol = internal.ErrorSymbol
type VoidSymbol = internal.VoidSymbol
type ExternalTableSymbol = internal.ExternalTableSymbol
type MaterializedViewSymbol = internal.MaterializedViewSymbol
type StoredQueryResultSymbol = internal.StoredQueryResultSymbol
type EntityGroupSymbol = internal.EntityGroupSymbol
type EntityGroupElementSymbol = internal.EntityGroupElementSymbol
type GroupSymbol = internal.GroupSymbol
type GraphSymbol = internal.GraphSymbol
type GraphModelSymbol = internal.GraphModelSymbol
type GraphSnapshotSymbol = internal.GraphSnapshotSymbol
type GlobalState = internal.GlobalState
type Binder = internal.Binder
type LocalScope = internal.LocalScope

const (
	CodeKindUnknown   = internal.CodeKindUnknown
	CodeKindQuery     = internal.CodeKindQuery
	CodeKindCommand   = internal.CodeKindCommand
	CodeKindDirective = internal.CodeKindDirective

	ParserKindGrammar = internal.ParserKindGrammar
	ParserKindDefault = internal.ParserKindDefault

	KustoDialectUnknown               = internal.KustoDialectUnknown
	KustoDialectClusterManagerCommand = internal.KustoDialectClusterManagerCommand
	KustoDialectDataManagerCommand    = internal.KustoDialectDataManagerCommand
	KustoDialectEngineCommand         = internal.KustoDialectEngineCommand
	KustoDialectQuery                 = internal.KustoDialectQuery

	IncludeTriviaAll        = internal.IncludeTriviaAll
	IncludeTriviaInterior   = internal.IncludeTriviaInterior
	IncludeTriviaMinimal    = internal.IncludeTriviaMinimal
	IncludeTriviaSingleLine = internal.IncludeTriviaSingleLine

	DiagnosticsIncludeSyntactic = internal.DiagnosticsIncludeSyntactic
	DiagnosticsIncludeSemantic  = internal.DiagnosticsIncludeSemantic
	DiagnosticsIncludeExpansion = internal.DiagnosticsIncludeExpansion
	DiagnosticsIncludeDefault   = internal.DiagnosticsIncludeDefault

	SymbolMatchNone               = internal.SymbolMatchNone
	SymbolMatchColumn             = internal.SymbolMatchColumn
	SymbolMatchTable              = internal.SymbolMatchTable
	SymbolMatchExternalTable      = internal.SymbolMatchExternalTable
	SymbolMatchFunction           = internal.SymbolMatchFunction
	SymbolMatchView               = internal.SymbolMatchView
	SymbolMatchLocal              = internal.SymbolMatchLocal
	SymbolMatchDatabase           = internal.SymbolMatchDatabase
	SymbolMatchCluster            = internal.SymbolMatchCluster
	SymbolMatchEntityGroup        = internal.SymbolMatchEntityGroup
	SymbolMatchEntityGroupElement = internal.SymbolMatchEntityGroupElement
	SymbolMatchScalar             = internal.SymbolMatchScalar
	SymbolMatchTabular            = internal.SymbolMatchTabular
	SymbolMatchNonScalar          = internal.SymbolMatchNonScalar
	SymbolMatchMaterializedView   = internal.SymbolMatchMaterializedView
	SymbolMatchOption             = internal.SymbolMatchOption
	SymbolMatchGraph              = internal.SymbolMatchGraph
	SymbolMatchStoredQueryResult  = internal.SymbolMatchStoredQueryResult
	SymbolMatchGraphModel         = internal.SymbolMatchGraphModel
	SymbolMatchGraphSnapshot      = internal.SymbolMatchGraphSnapshot
	SymbolMatchAny                = internal.SymbolMatchAny
	SymbolMatchDefault            = internal.SymbolMatchDefault

	SymbolKindNone                   = internal.SymbolKindNone
	SymbolKindDatabase               = internal.SymbolKindDatabase
	SymbolKindTable                  = internal.SymbolKindTable
	SymbolKindColumn                 = internal.SymbolKindColumn
	SymbolKindFunction               = internal.SymbolKindFunction
	SymbolKindVariable               = internal.SymbolKindVariable
	SymbolKindParameter              = internal.SymbolKindParameter
	SymbolKindOption                 = internal.SymbolKindOption
	SymbolKindOperator               = internal.SymbolKindOperator
	SymbolKindPattern                = internal.SymbolKindPattern
	SymbolKindTuple                  = internal.SymbolKindTuple
	SymbolKindBag                    = internal.SymbolKindBag
	SymbolKindArray                  = internal.SymbolKindArray
	SymbolKindCluster                = internal.SymbolKindCluster
	SymbolKindGroup                  = internal.SymbolKindGroup
	SymbolKindVoid                   = internal.SymbolKindVoid
	SymbolKindCommand                = internal.SymbolKindCommand
	SymbolKindGraph                  = internal.SymbolKindGraph
	SymbolKindPrimitive              = internal.SymbolKindPrimitive
	SymbolKindMaterializedView       = internal.SymbolKindMaterializedView
	SymbolKindQueryOperatorParameter = internal.SymbolKindQueryOperatorParameter
	SymbolKindEntityGroup            = internal.SymbolKindEntityGroup
	SymbolKindEntityGroupElement     = internal.SymbolKindEntityGroupElement
	SymbolKindStoredQueryResult      = internal.SymbolKindStoredQueryResult
	SymbolKindGraphModel             = internal.SymbolKindGraphModel
	SymbolKindGraphSnapshot          = internal.SymbolKindGraphSnapshot
	SymbolKindType                   = internal.SymbolKindType
	SymbolKindScalar                 = internal.SymbolKindScalar
	SymbolKindError                  = internal.SymbolKindError

	TabularityNone        = internal.TabularityNone
	TabularityScalar      = internal.TabularityScalar
	TabularityTabular     = internal.TabularityTabular
	TabularityOther       = internal.TabularityOther
	TabularityUnknown     = internal.TabularityUnknown
	TabularityUnspecified = internal.TabularityUnspecified
)

var (
	DefaultParseOptions    = internal.DefaultParseOptions
	DefaultGlobalState     = internal.DefaultGlobalState
	ScalarTypes            = newScalarTypesAPI()
	EmptyEntityGroupSymbol = internal.EmptyEntityGroupSymbol
	EmptyGraphSymbol       = internal.EmptyGraphSymbol
	EmptyTableSymbol       = internal.EmptyTableSymbol
	EmptyTupleSymbol       = internal.EmptyTupleSymbol
	Error                  = internal.Error
	ErrorSymbolInstance    = internal.Error
	UnknownClusterSymbol   = internal.UnknownClusterSymbol
	UnknownDatabaseSymbol  = internal.UnknownDatabaseSymbol
	UnknownParameter       = internal.UnknownParameter
	Void                   = internal.Void
	VoidSymbolInstance     = internal.Void
)

// NewGlobalStateForEnvironment converts the analyzer's JSON-compatible
// environment model into the same GlobalState used by the parser and binder.
func NewGlobalStateForEnvironment(env *api.EnvironmentDefinition, local api.LocalData) (*GlobalState, error) {
	catalog, err := api.NewCatalog(env, local)
	if err != nil {
		return nil, err
	}
	return catalog.Globals, nil
}

func Parse(text string) *KustoCode { return wrapKustoCode(internal.Parse(text)) }

func ParseWithGlobals(text string, globals *GlobalState) *KustoCode {
	return wrapKustoCode(internal.ParseWithGlobals(text, globals))
}

func ParseAndAnalyze(text string, globals ...*GlobalState) *KustoCode {
	return wrapKustoCode(internal.ParseAndAnalyze(text, globals...))
}

func ParseAndAnalyzeWithCancellation(text string, globals *GlobalState, token CancellationToken) *KustoCode {
	return wrapKustoCode(internal.ParseAndAnalyzeWithCancellation(text, globals, token))
}

func ParseForEnvironment(text string, env *api.EnvironmentDefinition, local api.LocalData) (*KustoCode, error) {
	globals, err := NewGlobalStateForEnvironment(env, local)
	if err != nil {
		return nil, err
	}
	return ParseWithGlobals(text, globals), nil
}

func ParseAndAnalyzeForEnvironment(text string, env *api.EnvironmentDefinition, local api.LocalData) (*KustoCode, error) {
	return ParseAndAnalyzeForEnvironmentContext(context.Background(), text, env, local)
}

func ParseAndAnalyzeForEnvironmentContext(ctx context.Context, text string, env *api.EnvironmentDefinition, local api.LocalData) (*KustoCode, error) {
	return parseAndAnalyzeForEnvironmentContext(ctx, text, env, local)
}

func parseAndAnalyzeForEnvironmentContext(ctx context.Context, text string, env *api.EnvironmentDefinition, local api.LocalData) (code *KustoCode, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if canceled, ok := recovered.(OperationCanceledError); ok {
				code = nil
				err = canceled
				return
			}
			panic(recovered)
		}
	}()
	globals, err := NewGlobalStateForEnvironment(env, local)
	if err != nil {
		return nil, err
	}
	token := NewCancellationToken(ctx)
	if err := token.Err(); err != nil {
		return nil, err
	}
	code = ParseAndAnalyzeWithCancellation(text, globals, token)
	if err := token.Err(); err != nil {
		return nil, err
	}
	return code, nil
}

func CodeKindOf(text string) CodeKind { return internal.CodeKindOf(text) }

// GetKind gets the kind of code contained in the text.
func GetKind(text string) CodeKind { return CodeKindOf(text) }

func Lex(text string) []LexicalToken { return internal.Lex(text) }

func LexWithOptions(text string, options ParseOptions) []LexicalToken {
	return internal.LexWithOptions(text, options)
}

func ParseToken(text string, start ...int) *LexicalToken {
	offset := 0
	if len(start) > 0 {
		offset = start[0]
	}
	return internal.ParseToken(text, offset)
}

func ParseTokenWithOptions(text string, start int, options ParseOptions) *LexicalToken {
	return internal.ParseTokenWithOptions(text, start, options)
}

func ParseTokens(text string) []LexicalToken { return internal.ParseTokens(text) }

func ParseTokensWithOptions(text string, options ParseOptions) []LexicalToken {
	return internal.ParseTokensWithOptions(text, options)
}

// AppendTokens parses text and appends its tokens to dst without modifying the
// existing prefix. When options is omitted, DefaultParseOptions is used.
func AppendTokens(dst []LexicalToken, text string, options ...ParseOptions) []LexicalToken {
	parsed := internal.ParseTokens(text)
	if len(options) > 0 {
		parsed = internal.ParseTokensWithOptions(text, options[0])
	}
	return append(dst, parsed...)
}

func ParseQuery(text string) *QueryBlock { return internal.ParseQuery(text) }

func ParseQueryWithOptions(text string, options ParseOptions) *QueryBlock {
	return internal.ParseQueryWithOptions(text, options)
}

func ParseQueryTokens(tokens []LexicalToken) *QueryBlock {
	return internal.ParseQueryTokens(tokens)
}

func ParseQueryTokensWithOptions(tokens []LexicalToken, options ParseOptions) *QueryBlock {
	return internal.ParseQueryTokensWithOptions(tokens, options)
}

func ParseQueryTokensFrom(tokens []LexicalToken, start int) *QueryBlock {
	return internal.ParseQueryTokensFrom(tokens, start)
}

func ParseQueryTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *QueryBlock {
	return internal.ParseQueryTokensFromWithOptions(tokens, start, options)
}

func ParseExpression(text string) Expression { return internal.ParseExpression(text) }

func ParseExpressionWithOptions(text string, options ParseOptions) Expression {
	return internal.ParseExpressionWithOptions(text, options)
}

func ParseExpressionTokens(tokens []LexicalToken) Expression {
	return internal.ParseExpressionTokens(tokens)
}

func ParseExpressionTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return internal.ParseExpressionTokensWithOptions(tokens, options)
}

func ParseExpressionTokensFrom(tokens []LexicalToken, start int) Expression {
	return internal.ParseExpressionTokensFrom(tokens, start)
}

func ParseExpressionTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return internal.ParseExpressionTokensFromWithOptions(tokens, start, options)
}

func ParseLiteral(text string) Expression { return internal.ParseLiteral(text) }

func ParseLiteralWithOptions(text string, options ParseOptions) Expression {
	return internal.ParseLiteralWithOptions(text, options)
}

func ParseLiteralTokens(tokens []LexicalToken) Expression {
	return internal.ParseLiteralTokens(tokens)
}

func ParseLiteralTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return internal.ParseLiteralTokensWithOptions(tokens, options)
}

func ParseLiteralTokensFrom(tokens []LexicalToken, start int) Expression {
	return internal.ParseLiteralTokensFrom(tokens, start)
}

func ParseLiteralTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return internal.ParseLiteralTokensFromWithOptions(tokens, start, options)
}

func ParseEntityPath(text string) Expression { return internal.ParseEntityPath(text) }

func ParseEntityPathWithOptions(text string, options ParseOptions) Expression {
	return internal.ParseEntityPathWithOptions(text, options)
}

func ParseEntityPathTokens(tokens []LexicalToken) Expression {
	return internal.ParseEntityPathTokens(tokens)
}

func ParseEntityPathTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return internal.ParseEntityPathTokensWithOptions(tokens, options)
}

func ParseEntityPathTokensFrom(tokens []LexicalToken, start int) Expression {
	return internal.ParseEntityPathTokensFrom(tokens, start)
}

func ParseEntityPathTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return internal.ParseEntityPathTokensFromWithOptions(tokens, start, options)
}

func ParseEntityGroup(text string) Expression { return internal.ParseEntityGroup(text) }

func ParseEntityGroupWithOptions(text string, options ParseOptions) Expression {
	return internal.ParseEntityGroupWithOptions(text, options)
}

func ParseEntityGroupTokens(tokens []LexicalToken) Expression {
	return internal.ParseEntityGroupTokens(tokens)
}

func ParseEntityGroupTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return internal.ParseEntityGroupTokensWithOptions(tokens, options)
}

func ParseEntityGroupTokensFrom(tokens []LexicalToken, start int) Expression {
	return internal.ParseEntityGroupTokensFrom(tokens, start)
}

func ParseEntityGroupTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return internal.ParseEntityGroupTokensFromWithOptions(tokens, start, options)
}

func ParseRowSchema(text string) *RowSchema { return internal.ParseRowSchema(text) }

func ParseRowSchemaWithOptions(text string, options ParseOptions) *RowSchema {
	return internal.ParseRowSchemaWithOptions(text, options)
}

func ParseRowSchemaTokens(tokens []LexicalToken) *RowSchema {
	return internal.ParseRowSchemaTokens(tokens)
}

func ParseRowSchemaTokensWithOptions(tokens []LexicalToken, options ParseOptions) *RowSchema {
	return internal.ParseRowSchemaTokensWithOptions(tokens, options)
}

func ParseRowSchemaTokensFrom(tokens []LexicalToken, start int) *RowSchema {
	return internal.ParseRowSchemaTokensFrom(tokens, start)
}

func ParseRowSchemaTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *RowSchema {
	return internal.ParseRowSchemaTokensFromWithOptions(tokens, start, options)
}

func ParseFunctionParameters(text string) *FunctionParameters {
	return internal.ParseFunctionParameters(text)
}

func ParseFunctionParametersWithOptions(text string, options ParseOptions) *FunctionParameters {
	return internal.ParseFunctionParametersWithOptions(text, options)
}

func ParseFunctionParametersTokens(tokens []LexicalToken) *FunctionParameters {
	return internal.ParseFunctionParametersTokens(tokens)
}

func ParseFunctionParametersTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameters {
	return internal.ParseFunctionParametersTokensWithOptions(tokens, options)
}

func ParseFunctionParametersTokensFrom(tokens []LexicalToken, start int) *FunctionParameters {
	return internal.ParseFunctionParametersTokensFrom(tokens, start)
}

func ParseFunctionParametersTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameters {
	return internal.ParseFunctionParametersTokensFromWithOptions(tokens, start, options)
}

func ParseFunctionParameter(text string) *FunctionParameter {
	return internal.ParseFunctionParameter(text)
}

func ParseFunctionParameterWithOptions(text string, options ParseOptions) *FunctionParameter {
	return internal.ParseFunctionParameterWithOptions(text, options)
}

func ParseFunctionParameterTokens(tokens []LexicalToken) *FunctionParameter {
	return internal.ParseFunctionParameterTokens(tokens)
}

func ParseFunctionParameterTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameter {
	return internal.ParseFunctionParameterTokensWithOptions(tokens, options)
}

func ParseFunctionParameterTokensFrom(tokens []LexicalToken, start int) *FunctionParameter {
	return internal.ParseFunctionParameterTokensFrom(tokens, start)
}

func ParseFunctionParameterTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameter {
	return internal.ParseFunctionParameterTokensFromWithOptions(tokens, start, options)
}

func ParseFunctionBody(text string) *FunctionBody { return internal.ParseFunctionBody(text) }

func ParseFunctionBodyWithOptions(text string, options ParseOptions) *FunctionBody {
	return internal.ParseFunctionBodyWithOptions(text, options)
}

func ParseFunctionBodyTokens(tokens []LexicalToken) *FunctionBody {
	return internal.ParseFunctionBodyTokens(tokens)
}

func ParseFunctionBodyTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionBody {
	return internal.ParseFunctionBodyTokensWithOptions(tokens, options)
}

func ParseFunctionBodyTokensFrom(tokens []LexicalToken, start int) *FunctionBody {
	return internal.ParseFunctionBodyTokensFrom(tokens, start)
}

func ParseFunctionBodyTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionBody {
	return internal.ParseFunctionBodyTokensFromWithOptions(tokens, start, options)
}

func ParseCommandBlock(text string) *CommandBlock { return internal.ParseCommandBlock(text) }

func NewSyntaxTree(root *SyntaxNode, original *SyntaxTree, offsetInOriginal int) *SyntaxTree {
	if root != nil {
		if root.Parent() != nil {
			panic(internal.InvalidOperationError{Message: "Cannot assign tree. This node is not the root element."})
		}
		if root.Tree() != nil {
			panic(internal.InvalidOperationError{Message: "Cannot assign tree. This node already has a syntax tree."})
		}
	}
	return internal.NewSyntaxTree(root, original, offsetInOriginal)
}

func CopyAsFragment[T SyntaxElement](node T) T {
	clone := internal.CopySyntaxElementAsFragment(node)
	if typed, ok := clone.(T); ok {
		return typed
	}
	panic("CopyAsFragment could not preserve the syntax node type")
}

func WrapGeneratedSyntaxElement(element SyntaxElement) SyntaxElement {
	return internal.WrapGeneratedSyntaxElement(element)
}

func WrapGeneratedSyntaxNode(node *SyntaxNode) SyntaxElement {
	return internal.WrapGeneratedSyntaxNode(node)
}

func AcceptSyntaxVisitor(element SyntaxElement, visitor SyntaxVisitor) {
	internal.AcceptSyntaxVisitor(element, visitor)
}

func AcceptSyntaxVisitorResult[T any](element SyntaxElement, visitor SyntaxVisitorResult[T]) T {
	return internal.AcceptSyntaxVisitorResult(element, visitor)
}

func SyntaxIsMissing(element SyntaxElement) bool { return internal.SyntaxIsMissing(element) }

func SyntaxHasMissingChildren(element SyntaxElement) bool {
	return internal.SyntaxHasMissingChildren(element)
}

func SyntaxGetChildIndex(element, child SyntaxElement) int {
	return internal.SyntaxGetChildIndex(element, child)
}

func SyntaxDepth(element SyntaxElement) int { return internal.SyntaxDepth(element) }

func SyntaxGetCommonAncestor(a, b SyntaxElement) *SyntaxNode {
	return internal.SyntaxGetCommonAncestor(a, b)
}

func SyntaxGetDescendantIndex(element, descendant SyntaxElement) int {
	return internal.SyntaxGetDescendantIndex(element, descendant)
}

func SyntaxIsAncestorOf(element, descendant SyntaxElement) bool {
	return internal.SyntaxIsAncestorOf(element, descendant)
}

func SyntaxNameInParent(element SyntaxElement) string {
	return internal.SyntaxNameInParent(element)
}

func GetFirstAncestor[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return internal.GetFirstAncestor(element, predicate...)
}

func GetFirstDescendant[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return internal.GetFirstDescendant(element, predicate...)
}

func GetDescendants[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return internal.GetDescendants(element, predicate...)
}

func GetSyntaxTokens(element SyntaxElement, includeZeroWidthTokens ...bool) []*SyntaxToken {
	return internal.GetSyntaxTokens(element, includeZeroWidthTokens...)
}

func WalkSyntaxTokens(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	internal.WalkSyntaxTokens(element, start, end, action)
}

func WalkSyntaxTokensByte(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	internal.WalkSyntaxTokensByte(element, start, end, action)
}

func WalkSyntaxTokensUTF16(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	internal.WalkSyntaxTokensUTF16(element, start, end, action)
}

func WalkSyntaxElements(root SyntaxElement, before, after func(SyntaxElement), descend func(SyntaxElement) bool) {
	internal.WalkSyntaxElements(root, before, after, descend)
}

func WalkSyntaxNodes(root *SyntaxNode, before, after func(*SyntaxNode), descend func(*SyntaxNode) bool) {
	internal.WalkSyntaxNodes(root, before, after, descend)
}

func GetContainedDiagnostics(root SyntaxElement, include DiagnosticsInclude) []Diagnostic {
	if root == nil {
		return []Diagnostic{}
	}
	return internal.GetContainedDiagnosticsUTF16(root, include, true, CancellationToken{})
}

func GetContainedDiagnosticsWithCancellation(root SyntaxElement, include DiagnosticsInclude, token CancellationToken) []Diagnostic {
	if root == nil {
		return []Diagnostic{}
	}
	return internal.GetContainedDiagnosticsUTF16(root, include, true, token)
}

func GetContainedSyntaxDiagnostics(root SyntaxElement) []Diagnostic {
	if root == nil {
		return []Diagnostic{}
	}
	return internal.GetContainedDiagnosticsUTF16(root, internal.DiagnosticsIncludeSyntactic, false, CancellationToken{})
}

func WithDiagnostics(element SyntaxElement, diagnostics ...Diagnostic) SyntaxElement {
	return internal.WithDiagnostics(element, diagnostics...)
}

func WithAdditionalDiagnostics(element SyntaxElement, diagnostics ...Diagnostic) SyntaxElement {
	return internal.WithAdditionalDiagnostics(element, diagnostics...)
}

func ExpressionResultType(expression Expression) *TypeSymbol {
	return internal.ExpressionResultType(expression)
}

func ExpressionRawResultType(expression Expression) *TypeSymbol {
	return internal.ExpressionRawResultType(expression)
}

func ExpressionIsConstant(expression Expression) bool {
	return internal.ExpressionIsConstant(expression)
}

func ExpressionIsLiteral(expression Expression) bool {
	return internal.ExpressionIsLiteral(expression)
}

func ExpressionLiteralValueInfo(expression Expression) *ValueInfo {
	return internal.ExpressionLiteralValueInfo(expression)
}

func ExpressionLiteralValue(expression Expression) any {
	return internal.ExpressionLiteralValue(expression)
}

func ExpressionConstantValueInfo(expression Expression) *ValueInfo {
	return internal.ExpressionConstantValueInfo(expression)
}

func ExpressionConstantValue(expression Expression) any {
	return internal.ExpressionConstantValue(expression)
}

func TryBindSyntaxTree(tree *SyntaxTree, globals *GlobalState, tokens ...CancellationToken) bool {
	return internal.TryBindSyntaxTree(tree, globals, tokens...)
}

func GetReferencedSymbol(tree *SyntaxTree, position int, name string, globals *GlobalState, match SymbolMatch, tokens ...CancellationToken) Symbol {
	return internal.GetReferencedSymbol(tree, syntaxTreeBytePosition(tree, position), name, globals, match, tokens...)
}

func GetRowScope(tree *SyntaxTree, position int, globals *GlobalState, tokens ...CancellationToken) *TableSymbol {
	return internal.GetRowScope(tree, syntaxTreeBytePosition(tree, position), globals, tokens...)
}

func GetSymbolsInScope(tree *SyntaxTree, position int, globals *GlobalState, match SymbolMatch, include IncludeFunctionKind, tokens ...CancellationToken) []Symbol {
	return internal.GetSymbolsInScope(tree, syntaxTreeBytePosition(tree, position), globals, match, include, tokens...)
}

func syntaxTreeBytePosition(tree *SyntaxTree, position int) int {
	if tree == nil || tree.Root == nil {
		return 0
	}
	return utf16PositionToByteOffset(tree.Root.Text(IncludeTriviaAll), position)
}

func GetComputedReturnType(signature *Signature, globals *GlobalState, argumentTypes []*TypeSymbol) *TypeSymbol {
	return internal.GetComputedReturnType(signature, globals, argumentTypes)
}

func NewGlobalState(database *DatabaseSymbol) *GlobalState { return internal.NewGlobalState(database) }

// WithClientParameters enables Kusto client-parameter syntax and installs the
// supplied typed symbols in an immutable GlobalState clone.
func WithClientParameters(globals *GlobalState, symbols ...Symbol) *GlobalState {
	if globals == nil {
		globals = DefaultGlobalState
	}
	return internal.GlobalStateWithProperty(globals, internal.Properties.AllowClientParameters, true).
		WithClientSymbols(symbols)
}

func NewColumnSymbol(name string, typ *TypeSymbol) *ColumnSymbol {
	return internal.NewColumnSymbol(name, typ)
}

func NewTableSymbol(name string, columns ...*ColumnSymbol) *TableSymbol {
	return internal.NewTableSymbol(name, columns...)
}

func NewTableSymbolWithDescription(name string, columns []*ColumnSymbol, description string) *TableSymbol {
	return internal.NewTableSymbolWithDescription(name, columns, description)
}

func TableSymbolFromSchema(schema string) (*TableSymbol, error) {
	return internal.TableSymbolFromSchema(schema)
}

func NewTableSymbolFromSchema(name, schema string, description ...string) (*TableSymbol, error) {
	return internal.NewTableSymbolFromSchema(name, schema, description...)
}

func NewDatabaseSymbol(name string, members ...Symbol) *DatabaseSymbol {
	return internal.NewDatabaseSymbol(name, members...)
}

func NewDatabaseSymbolWithMembers(name string, members []Symbol, isOpen bool) *DatabaseSymbol {
	return internal.NewDatabaseSymbolWithAlternateName(name, "", members, isOpen)
}

func NewDatabaseSymbolWithAlternateName(name, alternateName string, members []Symbol, isOpen bool) *DatabaseSymbol {
	return internal.NewDatabaseSymbolWithAlternateName(name, alternateName, members, isOpen)
}

func NewClusterSymbol(name string, databases ...*DatabaseSymbol) *ClusterSymbol {
	return internal.NewClusterSymbol(name, databases...)
}

func NewClusterSymbolWithDatabases(name string, databases []*DatabaseSymbol, isOpen bool) *ClusterSymbol {
	return internal.NewClusterSymbolWithDatabases(name, databases, isOpen)
}

func NewClusterSymbolWithMembers(name string, members []Symbol, isOpen bool) *ClusterSymbol {
	return internal.NewClusterSymbolWithMembers(name, members, isOpen)
}

func NewOptionSymbol(name, description string, types []*ScalarSymbol, examples []string) *OptionSymbol {
	return internal.NewOptionSymbol(name, description, types, examples)
}

func NewOptionSymbolWithType(name, description string, typ *ScalarSymbol, examples []string) *OptionSymbol {
	return internal.NewOptionSymbolWithType(name, description, typ, examples)
}

func NewParameter(name string, typ *TypeSymbol, options ...ParameterOption) *Parameter {
	return internal.NewParameter(name, typ, options...)
}

func NewParameterKind(name string, kind ParameterTypeKind, options ...ParameterOption) *Parameter {
	return internal.NewParameterKind(name, kind, options...)
}

func NewParameterTypes(name string, types []*TypeSymbol, options ...ParameterOption) *Parameter {
	return internal.NewParameterTypes(name, types, options...)
}

func NewParameterSymbolType(name string, symbol Symbol, options ...ParameterOption) *Parameter {
	return internal.NewParameterSymbolType(name, symbol, options...)
}

func NewParameterFromSymbol(parameter *ParameterSymbol, isOptional bool, defaultValue Expression) *Parameter {
	return internal.NewParameterFromSymbol(parameter, isOptional, defaultValue)
}

func NewParameterFromFunctionParameter(declaration *FunctionParameter) *Parameter {
	return internal.NewParameterFromFunctionParameter(declaration)
}

func NewParametersFromFunctionParameters(declaration *FunctionParameters) []*Parameter {
	return internal.NewParametersFromFunctionParameters(declaration)
}

func ParseParameterList(parameters string) []*Parameter {
	return internal.ParseParameterList(parameters)
}

func ParameterDeclaration(parameter *Parameter) string {
	return internal.ParameterDeclaration(parameter)
}

func ParameterListDeclaration(parameters []*Parameter) string {
	return internal.ParameterListDeclaration(parameters)
}

func WithArgumentKind(kind ArgumentKind) ParameterOption {
	return internal.WithArgumentKind(kind)
}

func WithParameterOccurrence(minimum, maximum int) ParameterOption {
	return internal.WithParameterOccurrence(minimum, maximum)
}

func WithParameterDescription(description string) ParameterOption {
	return internal.WithParameterDescription(description)
}

func WithParameterValues(values ...any) ParameterOption {
	return internal.WithParameterValues(values...)
}

func WithParameterExamples(examples ...string) ParameterOption {
	return internal.WithParameterExamples(examples...)
}

func WithParameterCaseSensitive(isCaseSensitive bool) ParameterOption {
	return internal.WithParameterCaseSensitive(isCaseSensitive)
}

func WithParameterDefaultValueIndicator(indicator string) ParameterOption {
	return internal.WithParameterDefaultValueIndicator(indicator)
}

func WithParameterDefaultValue(defaultValue Expression) ParameterOption {
	return internal.WithParameterDefaultValue(defaultValue)
}

func NewSignature(returnType *TypeSymbol, parameters ...*Parameter) *Signature {
	return internal.NewSignature(returnType, parameters...)
}

func NewSymbolSignature(returnSymbol Symbol, parameters ...*Parameter) *Signature {
	return internal.NewSymbolSignature(returnSymbol, parameters...)
}

func NewComputedSignature(body string, tabularity Tabularity, parameters ...*Parameter) *Signature {
	return internal.NewComputedSignature(body, tabularity, parameters...)
}

func NewDeclarationSignature(declaration *FunctionBody, parameters ...*Parameter) *Signature {
	return internal.NewDeclarationSignature(declaration, parameters...)
}

func NewReturnKindSignature(returnKind ReturnTypeKind, parameters ...*Parameter) *Signature {
	return internal.NewReturnKindSignature(returnKind, parameters...)
}

func NewCustomSignature(returnType CustomReturnType, tabularity Tabularity, parameters ...*Parameter) *Signature {
	return internal.NewCustomSignature(returnType, tabularity, parameters...)
}

func NewFunctionSymbol(name string, signatures ...*Signature) *FunctionSymbol {
	return internal.NewFunctionSymbol(name, signatures...)
}

func NewScalarFunctionSymbol(name string, returnType *TypeSymbol, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewScalarFunctionSymbol(name, returnType, parameters...)
}

func NewComputedFunctionSymbol(name, body string, tabularity Tabularity, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewComputedFunctionSymbol(name, body, tabularity, parameters...)
}

func NewReturnKindFunctionSymbol(name string, returnKind ReturnTypeKind, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewReturnKindFunctionSymbol(name, returnKind, parameters...)
}

func NewCustomFunctionSymbol(name string, customReturnType CustomReturnType, tabularity Tabularity, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewCustomFunctionSymbol(name, customReturnType, tabularity, parameters...)
}

func NewFunctionSymbolFromParameterList(name, parameterList, body string, description ...string) *FunctionSymbol {
	return internal.NewFunctionSymbolFromParameterList(name, parameterList, body, description...)
}

func NewFunctionSymbolFromParameterListWithTabularity(name, parameterList, body string, tabularity Tabularity, description ...string) *FunctionSymbol {
	return internal.NewFunctionSymbolFromParameterListWithTabularity(name, parameterList, body, tabularity, description...)
}

func NewDeclarationFunctionSymbol(name string, declaration *FunctionBody, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewDeclarationFunctionSymbol(name, declaration, parameters...)
}

func NewExternalTableSymbol(name string, columns ...*ColumnSymbol) *ExternalTableSymbol {
	return internal.NewExternalTableSymbol(name, columns...)
}

func NewExternalTableSymbolWithDescription(name string, columns []*ColumnSymbol, description string) *ExternalTableSymbol {
	return internal.NewExternalTableSymbolWithDescription(name, columns, description)
}

func NewExternalTableSymbolFromSchema(name, schema string, description ...string) (*ExternalTableSymbol, error) {
	return internal.NewExternalTableSymbolFromSchema(name, schema, description...)
}

func NewMaterializedViewSymbol(name string, columns []*ColumnSymbol, query string, description ...string) *MaterializedViewSymbol {
	return internal.NewMaterializedViewSymbol(name, columns, query, description...)
}

func NewMaterializedViewSymbolFromSchema(name, schema, query string, description ...string) (*MaterializedViewSymbol, error) {
	return internal.NewMaterializedViewSymbolFromSchema(name, schema, query, description...)
}

func NewStoredQueryResultSymbol(name string, columns ...*ColumnSymbol) *StoredQueryResultSymbol {
	return internal.NewStoredQueryResultSymbol(name, columns...)
}

func NewStoredQueryResultSymbolFromColumns(name string, columns []*ColumnSymbol) *StoredQueryResultSymbol {
	return internal.NewStoredQueryResultSymbolFromColumns(name, columns)
}

func NewEntityGroupSymbol(name string, members ...Symbol) *EntityGroupSymbol {
	return internal.NewEntityGroupSymbol(name, members...)
}

func NewEntityGroupSymbolWithMembers(name string, members []Symbol, description string) *EntityGroupSymbol {
	return internal.NewEntityGroupSymbolWithMembers(name, members, description)
}

func NewEntityGroupSymbolFromDefinition(name, definition, description string) *EntityGroupSymbol {
	return internal.NewEntityGroupSymbolFromDefinition(name, definition, description)
}

func EntityGroupBodyFromDefinition(definition string) string {
	return internal.EntityGroupBodyFromDefinition(definition)
}

func NewEntityGroupElementSymbol(name string, entityGroup *EntityGroupSymbol, underlying Symbol) *EntityGroupElementSymbol {
	return internal.NewEntityGroupElementSymbol(name, entityGroup, underlying)
}

func NewGraphSymbol(name string, edgeShape, nodeShape *TableSymbol) *GraphSymbol {
	return internal.NewGraphSymbol(name, edgeShape, nodeShape)
}

func NewGraphSymbolFromNodeShapes(name string, edgeShape *TableSymbol, nodeShapes []*TableSymbol) *GraphSymbol {
	return internal.NewGraphSymbolFromNodeShapes(name, edgeShape, nodeShapes)
}

func NewGraphSymbolFromShapes(name string, edgeShapes, nodeShapes []*TableSymbol) *GraphSymbol {
	return internal.NewGraphSymbolFromShapes(name, edgeShapes, nodeShapes)
}

func NewGraphSymbolFromSchema(name, edgeSchema string, nodeSchema ...string) (*GraphSymbol, error) {
	return internal.NewGraphSymbolFromSchema(name, edgeSchema, nodeSchema...)
}

func GraphSymbolFromSchema(edgeSchema string, nodeSchema ...string) (*GraphSymbol, error) {
	return internal.GraphSymbolFromSchema(edgeSchema, nodeSchema...)
}

func MergeGraphSymbols(left, right *GraphSymbol) *GraphSymbol {
	return internal.MergeGraphSymbols(left, right)
}

func NewGraphModelSymbol(name string, edges, nodes []*Signature, snapshots []*GraphSnapshotSymbol) *GraphModelSymbol {
	return internal.NewGraphModelSymbol(name, edges, nodes, snapshots)
}

func NewGraphModelSymbolFromNames(name string, edges, nodes, snapshots []string) *GraphModelSymbol {
	return internal.NewGraphModelSymbolFromNames(name, edges, nodes, snapshots)
}

func NewGraphSnapshotSymbol(name string) *GraphSnapshotSymbol {
	return internal.NewGraphSnapshotSymbol(name)
}

func NewGroupSymbol(symbols ...Symbol) *GroupSymbol {
	return internal.NewGroupSymbol(symbols...)
}

func NewCommandSymbol(name, resultSchema string, construction ...string) *CommandSymbol {
	return internal.NewCommandSymbol(name, resultSchema, construction...)
}

func NewCommandSymbolWithResultType(name string, resultType *TableSymbol, construction ...string) *CommandSymbol {
	return internal.NewCommandSymbolWithResultType(name, resultType, construction...)
}

func NewOperatorSymbol(kind OperatorKind, signatures ...*Signature) *OperatorSymbol {
	return internal.NewOperatorSymbol(kind, signatures...)
}

func NewOperatorSymbolWithResult(kind OperatorKind, resultType *TypeSymbol) *OperatorSymbol {
	return internal.NewOperatorSymbolWithResult(kind, resultType)
}

func NewPatternSignature(argumentValues []string, pathValue, body string, hasPathValue ...bool) *PatternSignature {
	return internal.NewPatternSignature(argumentValues, pathValue, body, hasPathValue...)
}

func NewPatternSignatureWithFunctionBody(argumentValues []string, pathValue string, body *FunctionBody, hasPathValue ...bool) *PatternSignature {
	return internal.NewPatternSignatureWithFunctionBody(argumentValues, pathValue, body, hasPathValue...)
}

func NewPatternSignatureFromFunctionBody(argumentValues []string, pathValue string, body *FunctionBody, hasPathValue ...bool) *PatternSignature {
	return internal.NewPatternSignatureFromFunctionBody(argumentValues, pathValue, body, hasPathValue...)
}

func NewPatternSymbol(name string, parameters []*Parameter, pathParameter *Parameter, signatures []*PatternSignature) *PatternSymbol {
	return internal.NewPatternSymbol(name, parameters, pathParameter, signatures)
}

func NewNamedPatternSymbol(name string) *PatternSymbol {
	return internal.NewNamedPatternSymbol(name)
}

func NewVariableSymbol(name string, typ Symbol, isConstant bool, constantValueInfo any, source Expression) *VariableSymbol {
	return internal.NewVariableSymbol(name, typ, isConstant, constantValueInfo, source)
}

func NewVariableSymbolWithValueInfo(name string, typ Symbol, isConstant bool, constantValueInfo *ValueInfo, source Expression) *VariableSymbol {
	return internal.NewVariableSymbolWithValueInfo(name, typ, isConstant, constantValueInfo, source)
}

func NewParameterSymbol(name string, typ Symbol, description string) *ParameterSymbol {
	return internal.NewParameterSymbol(name, typ, description)
}

func FindFirstMember(symbol Symbol, name string, ignoreCase ...bool) Symbol {
	ignore := false
	if len(ignoreCase) > 0 {
		ignore = ignoreCase[0]
	}
	return internal.FindFirstMember(symbol, name, ignore)
}

func FindFirstMemberMatch(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) Symbol {
	return internal.FindFirstMemberMatch(symbol, name, match, ignoreCase)
}

func SymbolAssignableTo(source, target Symbol, allowedConversion ...Conversion) bool {
	return internal.SymbolAssignableTo(source, target, allowedConversion...)
}

func SymbolAssignableToAny(source Symbol, targets []Symbol, allowedConversion ...Conversion) bool {
	return internal.SymbolAssignableToAny(source, targets, allowedConversion...)
}

func AreTablesEquivalent(left, right Symbol) bool {
	return internal.AreTablesEquivalent(left, right)
}

func AreTableResultsEquivalent(left, right Symbol) bool {
	return internal.AreTableResultsEquivalent(left, right)
}

func AreTableColumnsEquivalent(left, right Symbol) bool {
	return internal.AreTableColumnsEquivalent(left, right)
}

func IsScalar(symbol Symbol) bool { return internal.IsScalar(symbol) }

func IsTabular(symbol Symbol) bool { return internal.IsTabular(symbol) }

func GetResultType(symbol Symbol) Symbol { return internal.GetResultType(symbol) }

func TableSymbolForType(typ *TypeSymbol) (*TableSymbol, bool) {
	return internal.TableSymbolForType(typ)
}

func FindMembers(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) []Symbol {
	return internal.FindMembers(symbol, name, match, ignoreCase)
}

func FindMembersByMatch(symbol Symbol, match SymbolMatch) []Symbol {
	return internal.FindMembersByMatch(symbol, match)
}

func SymbolMatches(symbol Symbol, match SymbolMatch) bool {
	return internal.SymbolMatches(symbol, match)
}

func SymbolMatchesName(symbol Symbol, name string, match SymbolMatch, ignoreCase bool) bool {
	return internal.SymbolMatchesName(symbol, name, match, ignoreCase)
}

func SyntaxText(kind SyntaxKind) string { return internal.SyntaxText(kind) }

func SyntaxCategoryOf(kind SyntaxKind) SyntaxCategory { return internal.SyntaxCategoryOf(kind) }

func SyntaxKindFromText(text string) (SyntaxKind, bool) { return internal.SyntaxKindFromText(text) }

func IsKeywordKind(kind SyntaxKind) bool { return internal.IsKeywordKind(kind) }

func IsOperatorKind(kind SyntaxKind) bool { return internal.IsOperatorKind(kind) }

func IsLiteralKind(kind SyntaxKind) bool { return internal.IsLiteralKind(kind) }
