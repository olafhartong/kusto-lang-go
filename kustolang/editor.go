// Package kustolang exposes the native Go port of Kusto.Language editor
// services to applications outside this module's internal packages.
//
// Portions are ported from Microsoft's Kusto Query Language project:
// https://github.com/microsoft/Kusto-Query-Language
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.
package kustolang

import (
	"context"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
	api "github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

type DiagnosticLocationKind = internal.DiagnosticLocationKind
type Diagnostic = internal.Diagnostic
type InvalidOperationError = internal.InvalidOperationError
type OperationCanceledError = internal.OperationCanceledError
type CancellationToken = internal.CancellationToken
type AnalyzerInfo = internal.AnalyzerInfo
type ClientParameter = internal.ClientParameter
type FormattedText = internal.FormattedText
type SyntaxReference = internal.SyntaxReference
type ClusterReference = internal.ClusterReference
type DatabaseReference = internal.DatabaseReference
type ClassificationKind = internal.ClassificationKind
type ClassifiedRange = internal.ClassifiedRange
type ClassificationInfo = internal.ClassificationInfo
type TextRange = internal.TextRange
type FormattingOptions = internal.FormattingOptions
type QuickInfo = internal.QuickInfo
type QuickInfoItem = internal.QuickInfoItem
type QuickInfoKind = internal.QuickInfoKind
type ClassifiedText = internal.ClassifiedText
type RelatedInfo = internal.RelatedInfo
type RelatedElementKind = internal.RelatedElementKind
type RelatedElement = internal.RelatedElement
type FindRelatedOptions = internal.FindRelatedOptions
type BrackettingStyle = internal.BrackettingStyle
type BracketingStyle = internal.BrackettingStyle
type PlacementStyle = internal.PlacementStyle
type SpacingStyle = internal.SpacingStyle
type DualSpacingStyle = internal.DualSpacingStyle
type PositionBias = internal.PositionBias
type MinimalTextKind = internal.MinimalTextKind
type CompletionKind = internal.CompletionKind
type CompletionRank = internal.CompletionRank
type CompletionPriority = internal.CompletionPriority
type CompletionText = internal.CompletionText
type CompletionItem = internal.CompletionItem
type CompletionItemOption = internal.CompletionItemOption
type CompletionInfo = internal.CompletionInfo
type CompletionOptions = internal.CompletionOptions
type IncludeFunctionKind = internal.IncludeFunctionKind
type TextEdit = internal.TextEdit
type DisabledDiagnostics = internal.DisabledDiagnostics
type DisabledDiagnosticsBuilder = internal.DisabledDiagnosticsBuilder
type QuickInfoOptions = internal.QuickInfoOptions
type CodeAction = internal.CodeAction
type CodeActionBase = internal.CodeActionBase
type ApplyAction = internal.ApplyAction
type ApplyActionBase = internal.ApplyActionBase
type SingleAction = internal.SingleAction
type MultiAction = internal.MultiAction
type MenuAction = internal.MenuAction
type CodeActionInfo = internal.CodeActionInfo
type CodeActionOptions = internal.CodeActionOptions
type ResultAction = internal.ResultAction
type ResultActionBase = internal.ResultActionBase
type ChangeTextAction = internal.ChangeTextAction
type MoveCaretAction = internal.MoveCaretAction
type RenameAction = internal.RenameAction
type CodeActionResult = internal.CodeActionResult

const (
	CodeKindsQuery     = internal.CodeKindsQuery
	CodeKindsCommand   = internal.CodeKindsCommand
	CodeKindsDirective = internal.CodeKindsDirective
	CodeKindsUnknown   = internal.CodeKindsUnknown

	CodeServiceFeatureDiagnostics        = internal.CodeServiceFeatureDiagnostics
	CodeServiceFeatureClassification     = internal.CodeServiceFeatureClassification
	CodeServiceFeatureCompletion         = internal.CodeServiceFeatureCompletion
	CodeServiceFeatureOutlining          = internal.CodeServiceFeatureOutlining
	CodeServiceFeatureQuickInfo          = internal.CodeServiceFeatureQuickInfo
	CodeServiceFeatureClusterReferences  = internal.CodeServiceFeatureClusterReferences
	CodeServiceFeatureDatabaseReferences = internal.CodeServiceFeatureDatabaseReferences
	CodeServiceFeatureMinimalText        = internal.CodeServiceFeatureMinimalText
	CodeServiceFeatureFormatting         = internal.CodeServiceFeatureFormatting
	CodeServiceFeatureClientParameters   = internal.CodeServiceFeatureClientParameters

	DiagnosticCategoryGeneral     = internal.DiagnosticCategoryGeneral
	DiagnosticCategoryCorrectness = internal.DiagnosticCategoryCorrectness
	DiagnosticCategoryPerformance = internal.DiagnosticCategoryPerformance
	DiagnosticSeverityError       = internal.DiagnosticSeverityError
	DiagnosticSeverityWarning     = internal.DiagnosticSeverityWarning
	DiagnosticSeveritySuggestion  = internal.DiagnosticSeveritySuggestion
	DiagnosticSeverityInformation = internal.DiagnosticSeverityInformation
	DiagnosticSeverityHidden      = internal.DiagnosticSeverityHidden
	DiagnosticLocationRelative    = internal.DiagnosticLocationRelative
	DiagnosticLocationAbsolute    = internal.DiagnosticLocationAbsolute
	DiagnosticLocationRelativeEnd = internal.DiagnosticLocationRelativeEnd

	KustoAnalyzerCodeAvoidUsingContains                   = internal.KustoAnalyzerCodeAvoidUsingContains
	KustoAnalyzerCodeAvoidUsingIsNullStringComparision    = internal.KustoAnalyzerCodeAvoidUsingIsNullStringComparision
	KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison  = internal.KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison
	KustoAnalyzerCodeAvoidUsingShortStringComparision     = internal.KustoAnalyzerCodeAvoidUsingShortStringComparision
	KustoAnalyzerCodeAvoidUsingToBoolOnNumerics           = internal.KustoAnalyzerCodeAvoidUsingToBoolOnNumerics
	KustoAnalyzerCodeNullAggregation                      = internal.KustoAnalyzerCodeNullAggregation
	KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate  = internal.KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate
	KustoAnalyzerCodeAvoidUsingObsoleteFunctions          = internal.KustoAnalyzerCodeAvoidUsingObsoleteFunctions
	KustoAnalyzerCodeAvoidJoinWithoutKind                 = internal.KustoAnalyzerCodeAvoidJoinWithoutKind
	KustoAnalyzerCodeStdevTimespanConversion              = internal.KustoAnalyzerCodeStdevTimespanConversion
	KustoAnalyzerCodeColumnHasSameNameAsVariable          = internal.KustoAnalyzerCodeColumnHasSameNameAsVariable
	KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic = internal.KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic
	KustoAnalyzerCodeCalledFunctionHasErrors              = internal.KustoAnalyzerCodeCalledFunctionHasErrors
	KustoAnalyzerCodeAvoidUsingLegacyPartition            = internal.KustoAnalyzerCodeAvoidUsingLegacyPartition
	KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings         = internal.KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings
	KustoAnalyzerCodePreferUsingOptimizedAlternative      = internal.KustoAnalyzerCodePreferUsingOptimizedAlternative
	KustoAnalyzerCodeAvoidStrlenWithDynamic               = internal.KustoAnalyzerCodeAvoidStrlenWithDynamic

	BrackettingStyleDefault  = internal.BrackettingStyleDefault
	BrackettingStyleNone     = internal.BrackettingStyleNone
	BrackettingStyleVertical = internal.BrackettingStyleVertical
	BrackettingStyleDiagonal = internal.BrackettingStyleDiagonal
	BracketingStyleDefault   = internal.BrackettingStyleDefault
	BracketingStyleNone      = internal.BrackettingStyleNone
	BracketingStyleVertical  = internal.BrackettingStyleVertical
	BracketingStyleDiagonal  = internal.BrackettingStyleDiagonal
	PlacementStyleNone       = internal.PlacementStyleNone
	PlacementStyleNewLine    = internal.PlacementStyleNewLine
	PlacementStyleSmart      = internal.PlacementStyleSmart
	SpacingStyleAsIs         = internal.SpacingStyleAsIs
	SpacingStyleMinimal      = internal.SpacingStyleMinimal
	SpacingStyleOne          = internal.SpacingStyleOne
	DualSpacingStyleAsIs     = internal.DualSpacingStyleAsIs
	DualSpacingStyleNeither  = internal.DualSpacingStyleNeither
	DualSpacingStyleBefore   = internal.DualSpacingStyleBefore
	DualSpacingStyleAfter    = internal.DualSpacingStyleAfter
	DualSpacingStyleBoth     = internal.DualSpacingStyleBoth
	PositionBiasLeft         = internal.PositionBiasLeft
	PositionBiasRight        = internal.PositionBiasRight

	MinimalTextKindRemoveLeadingWhitespaceAndComments  = internal.MinimalTextKindRemoveLeadingWhitespaceAndComments
	MinimalTextKindMinimizeWhitespaceAndRemoveComments = internal.MinimalTextKindMinimizeWhitespaceAndRemoveComments
	MinimalTextKindSingleLine                          = internal.MinimalTextKindSingleLine

	CompletionKindUnknown           = internal.CompletionKindUnknown
	CompletionKindKeyword           = internal.CompletionKindKeyword
	CompletionKindPunctuation       = internal.CompletionKindPunctuation
	CompletionKindSyntax            = internal.CompletionKindSyntax
	CompletionKindIdentifier        = internal.CompletionKindIdentifier
	CompletionKindExample           = internal.CompletionKindExample
	CompletionKindScalarPrefix      = internal.CompletionKindScalarPrefix
	CompletionKindTabularPrefix     = internal.CompletionKindTabularPrefix
	CompletionKindTabularSuffix     = internal.CompletionKindTabularSuffix
	CompletionKindQueryPrefix       = internal.CompletionKindQueryPrefix
	CompletionKindCommandPrefix     = internal.CompletionKindCommandPrefix
	CompletionKindScalarInfix       = internal.CompletionKindScalarInfix
	CompletionKindRenderChart       = internal.CompletionKindRenderChart
	CompletionKindColumn            = internal.CompletionKindColumn
	CompletionKindTable             = internal.CompletionKindTable
	CompletionKindBuiltInFunction   = internal.CompletionKindBuiltInFunction
	CompletionKindLocalFunction     = internal.CompletionKindLocalFunction
	CompletionKindDatabaseFunction  = internal.CompletionKindDatabaseFunction
	CompletionKindAggregateFunction = internal.CompletionKindAggregateFunction
	CompletionKindParameter         = internal.CompletionKindParameter
	CompletionKindVariable          = internal.CompletionKindVariable
	CompletionKindDatabase          = internal.CompletionKindDatabase
	CompletionKindCluster           = internal.CompletionKindCluster
	CompletionKindMaterialiedView   = internal.CompletionKindMaterialiedView
	CompletionKindMaterializedView  = internal.CompletionKindMaterialiedView
	CompletionKindEntityGroup       = internal.CompletionKindEntityGroup
	CompletionKindGraph             = internal.CompletionKindGraph
	CompletionKindScalarType        = internal.CompletionKindScalarType
	CompletionKindOption            = internal.CompletionKindOption
	CompletionKindStoredQueryResult = internal.CompletionKindStoredQueryResult

	CompletionRankLiteral        = internal.CompletionRankLiteral
	CompletionRankAggregate      = internal.CompletionRankAggregate
	CompletionRankColumn         = internal.CompletionRankColumn
	CompletionRankTable          = internal.CompletionRankTable
	CompletionRankEntity         = internal.CompletionRankEntity
	CompletionRankVariable       = internal.CompletionRankVariable
	CompletionRankFunction       = internal.CompletionRankFunction
	CompletionRankKeyword        = internal.CompletionRankKeyword
	CompletionRankStringOperator = internal.CompletionRankStringOperator
	CompletionRankMathOperator   = internal.CompletionRankMathOperator
	CompletionRankOther          = internal.CompletionRankOther
	CompletionRankDefault        = internal.CompletionRankDefault
	CompletionPriorityTop        = internal.CompletionPriorityTop
	CompletionPriorityHigh       = internal.CompletionPriorityHigh
	CompletionPriorityNormal     = internal.CompletionPriorityNormal
	CompletionPriorityLow        = internal.CompletionPriorityLow
	CompletionPriorityDefault    = internal.CompletionPriorityDefault

	IncludeFunctionKindBuiltInFunctions  = internal.IncludeFunctionKindBuiltInFunctions
	IncludeFunctionKindDatabaseFunctions = internal.IncludeFunctionKindDatabaseFunctions
	IncludeFunctionKindLocalFunctions    = internal.IncludeFunctionKindLocalFunctions
	IncludeFunctionKindLocalViews        = internal.IncludeFunctionKindLocalViews
	IncludeFunctionKindNone              = internal.IncludeFunctionKindNone
	IncludeFunctionKindAll               = internal.IncludeFunctionKindAll

	ClassificationKindPlainText          = internal.ClassificationKindPlainText
	ClassificationKindComment            = internal.ClassificationKindComment
	ClassificationKindPunctuation        = internal.ClassificationKindPunctuation
	ClassificationKindDirective          = internal.ClassificationKindDirective
	ClassificationKindLiteral            = internal.ClassificationKindLiteral
	ClassificationKindStringLiteral      = internal.ClassificationKindStringLiteral
	ClassificationKindType               = internal.ClassificationKindType
	ClassificationKindColumn             = internal.ClassificationKindColumn
	ClassificationKindTable              = internal.ClassificationKindTable
	ClassificationKindDatabase           = internal.ClassificationKindDatabase
	ClassificationKindFunction           = internal.ClassificationKindFunction
	ClassificationKindParameter          = internal.ClassificationKindParameter
	ClassificationKindVariable           = internal.ClassificationKindVariable
	ClassificationKindIdentifier         = internal.ClassificationKindIdentifier
	ClassificationKindClientParameter    = internal.ClassificationKindClientParameter
	ClassificationKindQueryParameter     = internal.ClassificationKindQueryParameter
	ClassificationKindScalarOperator     = internal.ClassificationKindScalarOperator
	ClassificationKindMathOperator       = internal.ClassificationKindMathOperator
	ClassificationKindQueryOperator      = internal.ClassificationKindQueryOperator
	ClassificationKindCommand            = internal.ClassificationKindCommand
	ClassificationKindKeyword            = internal.ClassificationKindKeyword
	ClassificationKindMaterializedView   = internal.ClassificationKindMaterializedView
	ClassificationKindSchemaMember       = internal.ClassificationKindSchemaMember
	ClassificationKindSignatureParameter = internal.ClassificationKindSignatureParameter
	ClassificationKindOption             = internal.ClassificationKindOption

	QuickInfoKindText             = internal.QuickInfoKindText
	QuickInfoKindError            = internal.QuickInfoKindError
	QuickInfoKindWarning          = internal.QuickInfoKindWarning
	QuickInfoKindSuggestion       = internal.QuickInfoKindSuggestion
	QuickInfoKindColumn           = internal.QuickInfoKindColumn
	QuickInfoKindTable            = internal.QuickInfoKindTable
	QuickInfoKindDatabase         = internal.QuickInfoKindDatabase
	QuickInfoKindCluster          = internal.QuickInfoKindCluster
	QuickInfoKindLiteral          = internal.QuickInfoKindLiteral
	QuickInfoKindType             = internal.QuickInfoKindType
	QuickInfoKindPattern          = internal.QuickInfoKindPattern
	QuickInfoKindParameter        = internal.QuickInfoKindParameter
	QuickInfoKindScalar           = internal.QuickInfoKindScalar
	QuickInfoKindVariable         = internal.QuickInfoKindVariable
	QuickInfoKindLocalFunction    = internal.QuickInfoKindLocalFunction
	QuickInfoKindBuiltInFunction  = internal.QuickInfoKindBuiltInFunction
	QuickInfoKindDatabaseFunction = internal.QuickInfoKindDatabaseFunction
	QuickInfoKindOperator         = internal.QuickInfoKindOperator
	QuickInfoKindCommand          = internal.QuickInfoKindCommand
	QuickInfoKindOption           = internal.QuickInfoKindOption
	QuickInfoKindGraph            = internal.QuickInfoKindGraph

	RelatedElementKindSyntax              = internal.RelatedElementKindSyntax
	RelatedElementKindReference           = internal.RelatedElementKindReference
	RelatedElementKindDeclaration         = internal.RelatedElementKindDeclaration
	RelatedElementKindOther               = internal.RelatedElementKindOther
	FindRelatedOptionsNone                = internal.FindRelatedOptionsNone
	FindRelatedOptionsSeeThroughVariables = internal.FindRelatedOptionsSeeThroughVariables

	CodeActionFixAllKind = internal.CodeActionFixAllKind
	CodeActionMenuKind   = internal.CodeActionMenuKind

	KustoActorNameFunctionInliner      = internal.KustoActorNameFunctionInliner
	KustoActorNameExtractExpression    = internal.KustoActorNameExtractExpression
	KustoActorNameConvertToMacroExpand = internal.KustoActorNameConvertToMacroExpand
	KustoActorNameDiagnosticFixer      = internal.KustoActorNameDiagnosticFixer
	KustoActorNameRename               = internal.KustoActorNameRename

	// DiagnosticFixActorName requests the analyzer diagnostic fixer used by
	// KustoCodeService in the upstream editor pipeline.
	DiagnosticFixActorName = internal.KustoActorNameDiagnosticFixer
)

var (
	FormattingOptionsDefault   = internal.FormattingOptionsDefault
	QuickInfoOptionsDefault    = internal.QuickInfoOptionsDefault
	CompletionOptionsDefault   = internal.CompletionOptionsDefault
	CompletionInfoEmpty        = internal.CompletionInfoEmpty
	CodeActionOptionsDefault   = internal.CodeActionOptionsDefault
	CodeActionInfoNoActions    = internal.CodeActionInfoNoActions
	CodeActionResultNothing    = internal.CodeActionResultNothing
	ClassificationInfoEmpty    = internal.ClassificationInfoEmpty
	DisabledDiagnosticsDefault = internal.DisabledDiagnosticsDefault
	QuickInfoEmpty             = internal.QuickInfoEmpty
	RelatedInfoEmpty           = internal.RelatedInfoEmpty
	RenameActionInstance       = internal.RenameActionInstance
	TextRangeEmpty             = internal.TextRangeEmpty
)

// CodeService is the public editor-service facade over the native
// KustoCodeService port.
type CodeService struct {
	service internal.CodeService
}

func wrapCodeService(service internal.CodeService) *CodeService {
	if service == nil {
		return nil
	}
	return &CodeService{service: service}
}

// NewCodeService creates an editor service using Kusto.Language's default
// global state.
func NewCodeService(text string) *CodeService {
	return wrapCodeService(internal.NewKustoCodeService(text))
}

// NewCodeServiceWithGlobals creates an editor service with an explicitly
// configured Kusto.Language GlobalState.
func NewCodeServiceWithGlobals(text string, globals *GlobalState) *CodeService {
	return wrapCodeService(internal.NewKustoCodeService(text, globals))
}

// NewCodeServiceForEnvironment validates an analyzer environment and creates
// an editor service with its tables, functions, clusters, and databases.
func NewCodeServiceForEnvironment(text string, env *api.EnvironmentDefinition, local api.LocalData) (*CodeService, error) {
	catalog, err := api.NewCatalog(env, local)
	if err != nil {
		return nil, err
	}
	return wrapCodeService(internal.NewKustoCodeService(text, catalog.Globals)), nil
}

func NewCancellationToken(ctx context.Context) CancellationToken {
	return internal.NewCancellationToken(ctx)
}

func NewAnalyzerInfo(name string, diagnostics []Diagnostic) AnalyzerInfo {
	return internal.NewAnalyzerInfo(name, diagnostics)
}

func NewClientParameter(name string, index *int, rng TextRange) ClientParameter {
	return internal.NewClientParameter(name, index, rng)
}

func NewClientParameterAt(name string, index *int, start, length int) ClientParameter {
	return internal.NewClientParameterAt(name, index, start, length)
}

func NewClientParameterRange(name string, rng TextRange) ClientParameter {
	return internal.NewClientParameterRange(name, rng)
}

func NewClientParameterReference(name string, start, length int) ClientParameter {
	return internal.NewClientParameterReference(name, start, length)
}

func NewFormattedText(text string, position int) FormattedText {
	return internal.NewFormattedText(text, position)
}

func NewSyntaxReference(start, length int) SyntaxReference {
	return internal.NewSyntaxReference(start, length)
}

func NewSyntaxReferenceRange(rng TextRange) SyntaxReference {
	return internal.NewSyntaxReferenceRange(rng)
}

func NewClusterReference(cluster string, start, length int) ClusterReference {
	return internal.NewClusterReference(cluster, start, length)
}

func NewDatabaseReference(database, cluster string, start, length int) DatabaseReference {
	return internal.NewDatabaseReference(database, cluster, start, length)
}

func NewClassifiedRange(kind ClassificationKind, start, length int) ClassifiedRange {
	return internal.NewClassifiedRange(kind, start, length)
}

func NewClassificationInfo(classifications []ClassifiedRange) ClassificationInfo {
	return internal.NewClassificationInfo(classifications)
}

func NewQuickInfoText(value string) QuickInfo {
	return internal.NewQuickInfoText(value)
}

func NewQuickInfo(items ...QuickInfoItem) QuickInfo {
	return internal.NewQuickInfo(items...)
}

func NewQuickInfoItems(items []QuickInfoItem) QuickInfo {
	return internal.NewQuickInfoItems(items)
}

func NewQuickInfoItem(kind QuickInfoKind, parts ...ClassifiedText) QuickInfoItem {
	return internal.NewQuickInfoItem(kind, parts...)
}

func NewQuickInfoItemText(kind QuickInfoKind, text string) QuickInfoItem {
	return internal.NewQuickInfoItemText(kind, text)
}

func NewClassifiedText(text string) ClassifiedText {
	return internal.NewClassifiedText(text)
}

func NewClassifiedTextKind(kind ClassificationKind, text string) ClassifiedText {
	return internal.NewClassifiedTextKind(kind, text)
}

func NewRelatedInfo(elements []RelatedElement, currentIndex int) RelatedInfo {
	return internal.NewRelatedInfo(elements, currentIndex)
}

func NewRelatedElementRangeWithCursors(rng TextRange, kind RelatedElementKind, cursorLeft, cursorRight int) RelatedElement {
	return internal.NewRelatedElementRangeWithCursors(rng, kind, cursorLeft, cursorRight)
}

func NewRelatedElementRange(rng TextRange, kind RelatedElementKind) RelatedElement {
	return internal.NewRelatedElementRange(rng, kind)
}

func NewRelatedElementWithCursors(start, length int, kind RelatedElementKind, cursorLeft, cursorRight int) RelatedElement {
	return internal.NewRelatedElementWithCursors(start, length, kind, cursorLeft, cursorRight)
}

func NewRelatedElement(start, length int, kind RelatedElementKind) RelatedElement {
	return internal.NewRelatedElement(start, length, kind)
}

func KustoCodeServiceIncludeTrivia(kind MinimalTextKind) IncludeTrivia {
	return internal.KustoCodeServiceIncludeTrivia(kind)
}

func ParseDisabledDiagnostics(text string) DisabledDiagnostics {
	return internal.ParseDisabledDiagnostics(text)
}

func NewDisabledDiagnostics(text ...string) DisabledDiagnostics {
	return internal.NewDisabledDiagnostics(text...)
}

func NewDisabledDiagnosticsBuilder(source ...DisabledDiagnostics) *DisabledDiagnosticsBuilder {
	return internal.NewDisabledDiagnosticsBuilder(source...)
}

func CanDisableSeverity(severity string) bool {
	return internal.CanDisableSeverity(severity)
}

func CanDisableCode(code string) bool {
	return internal.CanDisableCode(code)
}

func NewCompletionText(text string, caret, selectText bool) CompletionText {
	return internal.NewCompletionText(text, caret, selectText)
}

func NewCompletionTextEdit(text string) CompletionText {
	return internal.NewCompletionTextEdit(text)
}

func ParseCompletionText(textWithMarkers string, markers ...string) []CompletionText {
	return internal.ParseCompletionText(textWithMarkers, markers...)
}

func WithCompletionBeforeText(text string) CompletionItemOption {
	return internal.WithCompletionBeforeText(text)
}

func WithCompletionAfterText(text string) CompletionItemOption {
	return internal.WithCompletionAfterText(text)
}

func WithCompletionMatchText(text string) CompletionItemOption {
	return internal.WithCompletionMatchText(text)
}

func WithCompletionOrderText(text string) CompletionItemOption {
	return internal.WithCompletionOrderText(text)
}

func WithCompletionApplyTexts(texts ...CompletionText) CompletionItemOption {
	return internal.WithCompletionApplyTexts(texts...)
}

func WithCompletionApplyTextMarkers(textWithMarkers string) CompletionItemOption {
	return internal.WithCompletionApplyTextMarkers(textWithMarkers)
}

func WithCompletionRank(rank CompletionRank) CompletionItemOption {
	return internal.WithCompletionRank(rank)
}

func WithCompletionPriority(priority CompletionPriority) CompletionItemOption {
	return internal.WithCompletionPriority(priority)
}

func WithCompletionRetrigger(retrigger bool) CompletionItemOption {
	return internal.WithCompletionRetrigger(retrigger)
}

func NewCompletionItem(kind CompletionKind, displayText string, options ...CompletionItemOption) CompletionItem {
	return internal.NewCompletionItem(kind, displayText, options...)
}

func NewSyntaxCompletionItem(displayText string, options ...CompletionItemOption) CompletionItem {
	return internal.NewSyntaxCompletionItem(displayText, options...)
}

func NewCompletionInfo(items []CompletionItem, editStart, editLength int) CompletionInfo {
	return internal.NewCompletionInfo(items, editStart, editLength)
}

func NewCodeActionOptions(filter *DisabledDiagnostics, formatting *FormattingOptions) CodeActionOptions {
	return internal.NewCodeActionOptions(filter, formatting)
}

func NewSingleAction(kind, title, description string, data []string) *SingleAction {
	return internal.NewSingleAction(kind, title, description, data)
}

func NewCodeAction(kind, title, description string, data []string) *SingleAction {
	return internal.NewCodeAction(kind, title, description, data)
}

func NewCodeActionTitle(title, description string, data []string) *SingleAction {
	return internal.NewCodeActionTitle(title, description, data)
}

func NewMultiAction(kind, title, description string, actions []ApplyAction) *MultiAction {
	return internal.NewMultiAction(kind, title, description, actions)
}

func NewCodeActionFixAll(title, description string, actions []ApplyAction) *MultiAction {
	return internal.NewCodeActionFixAll(title, description, actions)
}

func NewMenuAction(title, description string, actions []CodeAction) *MenuAction {
	return internal.NewMenuAction(title, description, actions)
}

func NewCodeActionMenu(title, description string, actions []CodeAction) *MenuAction {
	return internal.NewCodeActionMenu(title, description, actions)
}

func NewCodeActionInfo(actions []CodeAction) CodeActionInfo {
	return internal.NewCodeActionInfo(actions)
}

func NewChangeTextAction(changes []TextEdit, changedText string) ChangeTextAction {
	return internal.NewChangeTextAction(changes, changedText)
}

func NewChangeTextActionEditString(editString EditString) ChangeTextAction {
	return internal.NewChangeTextAction(editString.GetChanges(), editString.CurrentText())
}

func NewMoveCaretAction(newCaretPosition int) MoveCaretAction {
	return internal.NewMoveCaretAction(newCaretPosition)
}

func NewRenameAction() RenameAction {
	return internal.NewRenameAction()
}

func NewCodeActionResult(actions ...ResultAction) CodeActionResult {
	return internal.NewCodeActionResult(actions...)
}

func NewCodeActionResultActions(actions []ResultAction) CodeActionResult {
	return internal.NewCodeActionResultActions(actions)
}

func NewCodeActionChange(changes []TextEdit, changedText string) CodeActionResult {
	return internal.NewCodeActionChange(changes, changedText)
}

func NewCodeActionChangeEditString(editString EditString) CodeActionResult {
	return internal.NewCodeActionChange(editString.GetChanges(), editString.CurrentText())
}

func NewCodeActionChangeAndMove(changes []TextEdit, changedText string, newCaretPosition int) CodeActionResult {
	return internal.NewCodeActionChangeAndMove(changes, changedText, newCaretPosition)
}

func NewCodeActionChangeAndMoveEditString(editString EditString, newCaretPosition int) CodeActionResult {
	return internal.NewCodeActionChangeAndMove(editString.GetChanges(), editString.CurrentText(), newCaretPosition)
}

func NewCodeActionFailure(failureReason ...string) CodeActionResult {
	return internal.NewCodeActionFailure(failureReason...)
}

func NewTextEdit(start, deleteLength int, insertText string) TextEdit {
	return internal.NewTextEdit(start, deleteLength, insertText)
}

func NewTextRange(start, length int) TextRange {
	return internal.NewTextRange(start, length)
}

func TextRangeFromBounds(start, end int) TextRange {
	return internal.TextRangeFromBounds(start, end)
}

func TextRangesOverlap(startA, lengthA, startB, lengthB int) bool {
	return internal.TextRangesOverlap(startA, lengthA, startB, lengthB)
}

func NewTextReplacement(start, deleteLength int, insertText string) TextEdit {
	return internal.NewTextReplacement(start, deleteLength, insertText)
}

func NewTextDeletion(start, deleteLength int) TextEdit {
	return internal.NewTextDeletion(start, deleteLength)
}

func NewTextInsertion(start int, text string) TextEdit {
	return internal.NewTextInsertion(start, text)
}

func CompletionDisplayText(symbol Symbol, nameOnly ...bool) string {
	return internal.CompletionDisplayText(symbol, nameOnly...)
}

func KustoActorNamesAll() []string {
	return internal.KustoActorNamesAll()
}

func (s *CodeService) Text() string {
	if s == nil || s.service == nil {
		return ""
	}
	return s.service.Text()
}

func (s *CodeService) Kind() string {
	if s == nil || s.service == nil {
		return CodeKindsQuery
	}
	return s.service.Kind()
}

func (s *CodeService) IsFeatureSupported(feature string, position int) bool {
	return s != nil && s.service != nil && s.service.IsFeatureSupported(feature, position)
}

// IsFeatureSupportedDefault considers support over the entire text.
func (s *CodeService) IsFeatureSupportedDefault(feature string) bool {
	return s.IsFeatureSupported(feature, -1)
}

func (s *CodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	if s == nil || s.service == nil {
		return nil, false
	}
	diagnostics, ok := s.service.TryGetCachedDiagnostics()
	if !ok {
		return nil, false
	}
	return internal.ListToReadOnly(diagnostics), true
}

func (s *CodeService) GetDiagnostics(waitForAnalysis bool, token CancellationToken) []Diagnostic {
	if s == nil || s.service == nil {
		return []Diagnostic{}
	}
	return internal.ListToReadOnly(s.service.GetDiagnostics(waitForAnalysis, token))
}

// GetDiagnosticsDefault waits for analysis without cancellation.
func (s *CodeService) GetDiagnosticsDefault() []Diagnostic {
	return s.GetDiagnostics(true, CancellationToken{})
}

func (s *CodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	if s == nil || s.service == nil {
		return nil, false
	}
	diagnostics, ok := s.service.TryGetCachedAnalyzerDiagnostics()
	if !ok {
		return nil, false
	}
	return internal.ListToReadOnly(diagnostics), true
}

func (s *CodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, token CancellationToken) []Diagnostic {
	if s == nil || s.service == nil {
		return []Diagnostic{}
	}
	return internal.ListToReadOnly(s.service.GetAnalyzerDiagnostics(waitForAnalysis, token))
}

// GetAnalyzerDiagnosticsDefault waits for analysis without cancellation.
func (s *CodeService) GetAnalyzerDiagnosticsDefault() []Diagnostic {
	return s.GetAnalyzerDiagnostics(true, CancellationToken{})
}

// GetCombinedDiagnostics returns parser, semantic, and analyzer diagnostics in
// the same order as Kusto.Language's CodeService, optionally removing disabled
// diagnostic codes.
func (s *CodeService) GetCombinedDiagnostics(waitForAnalysis bool, filter *DisabledDiagnostics, token CancellationToken) []Diagnostic {
	if s == nil || s.service == nil {
		return []Diagnostic{}
	}
	return internal.ListToReadOnly(internal.CodeServiceGetCombinedDiagnostics(s.service, waitForAnalysis, filter, token))
}

// GetCombinedDiagnosticsDefault waits for analysis with no diagnostic filter.
func (s *CodeService) GetCombinedDiagnosticsDefault() []Diagnostic {
	return s.GetCombinedDiagnostics(true, nil, CancellationToken{})
}

func (s *CodeService) GetAnalyzers() []AnalyzerInfo {
	if s == nil || s.service == nil {
		return []AnalyzerInfo{}
	}
	infos := s.service.GetAnalyzers()
	result := make([]AnalyzerInfo, len(infos))
	for i, info := range infos {
		result[i] = internal.NewAnalyzerInfo(info.Name, info.Diagnostics)
	}
	return internal.ListToReadOnly(result)
}

func (s *CodeService) GetCodeActions(position, selectionStart, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, token CancellationToken) CodeActionInfo {
	if s == nil || s.service == nil {
		return internal.CodeActionInfoNoActions
	}
	result := s.service.GetCodeActions(position, selectionStart, selectionLength, options, waitForAnalysis, actorName, token)
	return internal.CloneCodeActionInfo(result)
}

// GetCodeActionsDefault uses the default options, waits for analysis, and queries all actors.
func (s *CodeService) GetCodeActionsDefault(position, selectionStart, selectionLength int) CodeActionInfo {
	return s.GetCodeActions(position, selectionStart, selectionLength, CodeActionOptionsDefault, true, "", CancellationToken{})
}

func (s *CodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, token CancellationToken) CodeActionResult {
	if s == nil || s.service == nil {
		return internal.NewCodeActionFailure("code service is nil")
	}
	return internal.CloneCodeActionResult(s.service.ApplyCodeAction(action, caretPosition, options, token))
}

// ApplyCodeActionDefault uses the default options without cancellation.
func (s *CodeService) ApplyCodeActionDefault(action ApplyAction, caretPosition int) CodeActionResult {
	return s.ApplyCodeAction(action, caretPosition, CodeActionOptionsDefault, CancellationToken{})
}

func (s *CodeService) GetClassifications(start, length int, clipToRange, waitForAnalysis bool, token CancellationToken) ClassificationInfo {
	if s == nil || s.service == nil {
		return internal.ClassificationInfoEmpty
	}
	result := s.service.GetClassifications(start, length, clipToRange, waitForAnalysis, token)
	return internal.NewClassificationInfo(result.Classifications)
}

// GetClassificationsDefault clips to the range and waits for analysis.
func (s *CodeService) GetClassificationsDefault(start, length int) ClassificationInfo {
	return s.GetClassifications(start, length, true, true, CancellationToken{})
}

func (s *CodeService) GetOutlines(options OutliningOptions, token CancellationToken) OutlineInfo {
	if s == nil || s.service == nil {
		return OutlineInfoEmpty
	}
	return outlineInfoFromInternal(s.service.GetOutlines(options.toInternal(), token))
}

// GetOutlinesDefault uses the pinned Kusto.Language default outlining options.
func (s *CodeService) GetOutlinesDefault(token CancellationToken) OutlineInfo {
	return s.GetOutlines(OutliningOptionsDefault, token)
}

func (s *CodeService) ShouldAutoComplete(position int, key rune, token CancellationToken) bool {
	return s != nil && s.service != nil && s.service.ShouldAutoComplete(position, key, token)
}

func (s *CodeService) GetCompletionItems(position int, options CompletionOptions, token CancellationToken) CompletionInfo {
	if s == nil || s.service == nil {
		return internal.CompletionInfoEmpty
	}
	result := s.service.GetCompletionItems(position, options, token)
	return internal.NewCompletionInfo(result.Items, result.EditStart, result.EditLength)
}

// GetCompletionItemsDefault uses the pinned Kusto.Language default completion options.
func (s *CodeService) GetCompletionItemsDefault(position int, token CancellationToken) CompletionInfo {
	return s.GetCompletionItems(position, CompletionOptionsDefault, token)
}

func (s *CodeService) GetQuickInfo(position int, options QuickInfoOptions, token CancellationToken) QuickInfo {
	if s == nil || s.service == nil {
		return internal.QuickInfoEmpty
	}
	result := s.service.GetQuickInfo(position, options, token)
	return internal.NewQuickInfoItems(result.Items)
}

// GetQuickInfoDefault uses the pinned Kusto.Language default quick-info options.
func (s *CodeService) GetQuickInfoDefault(position int, token CancellationToken) QuickInfo {
	return s.GetQuickInfo(position, QuickInfoOptionsDefault, token)
}

func (s *CodeService) GetElement(position int, token CancellationToken) TextRange {
	if s == nil || s.service == nil {
		return TextRange{}
	}
	return s.service.GetElement(position, token)
}

func (s *CodeService) GetRelatedElements(position int, options FindRelatedOptions, token CancellationToken) RelatedInfo {
	if s == nil || s.service == nil {
		return internal.RelatedInfoEmpty
	}
	result := s.service.GetRelatedElements(position, options, token)
	return internal.NewRelatedInfo(result.Elements, result.CurrentIndex)
}

func (s *CodeService) GetClusterReferences(token CancellationToken) []ClusterReference {
	if s == nil || s.service == nil {
		return []ClusterReference{}
	}
	return internal.ListToReadOnly(s.service.GetClusterReferences(token))
}

func (s *CodeService) GetDatabaseReferences(token CancellationToken) []DatabaseReference {
	if s == nil || s.service == nil {
		return []DatabaseReference{}
	}
	return internal.ListToReadOnly(s.service.GetDatabaseReferences(token))
}

func (s *CodeService) GetMinimalText(kind MinimalTextKind, token CancellationToken) string {
	if s == nil || s.service == nil {
		return ""
	}
	return s.service.GetMinimalText(kind, token)
}

func (s *CodeService) GetFormattedText(options FormattingOptions, caretPosition int, token CancellationToken) FormattedText {
	if s == nil || s.service == nil {
		return FormattedText{}
	}
	return s.service.GetFormattedText(options, caretPosition, token)
}

// GetFormattedTextDefault uses the pinned Kusto.Language default formatting options.
func (s *CodeService) GetFormattedTextDefault(caretPosition int, token CancellationToken) FormattedText {
	return s.GetFormattedText(FormattingOptionsDefault, caretPosition, token)
}

func (s *CodeService) GetClientParameters() []ClientParameter {
	if s == nil || s.service == nil {
		return []ClientParameter{}
	}
	parameters := s.service.GetClientParameters()
	result := make([]ClientParameter, len(parameters))
	for i, parameter := range parameters {
		result[i] = internal.NewClientParameter(parameter.Name, parameter.Index, parameter.Range)
	}
	return internal.ListToReadOnly(result)
}
