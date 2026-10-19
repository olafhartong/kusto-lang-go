package kustolang

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// CustomCodeServiceCallbacks supplies custom editor behavior. Any nil callback
// uses CommonCodeService's default behavior for that operation.
type CustomCodeServiceCallbacks struct {
	Kind                            func() string
	IsFeatureSupported              func(feature string, position int) bool
	TryGetCachedDiagnostics         func() ([]Diagnostic, bool)
	GetDiagnostics                  func(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic
	TryGetCachedAnalyzerDiagnostics func() ([]Diagnostic, bool)
	GetAnalyzerDiagnostics          func(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic
	GetCombinedDiagnostics          func(waitForAnalysis bool, filter *DisabledDiagnostics, cancellationToken CancellationToken) []Diagnostic
	GetAnalyzers                    func() []AnalyzerInfo
	GetCodeActions                  func(position, selectionStart, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo
	ApplyCodeAction                 func(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
	GetClassifications              func(start, length int, clipToRange, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo
	GetOutlines                     func(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo
	ShouldAutoComplete              func(position int, key rune, cancellationToken CancellationToken) bool
	GetCompletionItems              func(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo
	GetQuickInfo                    func(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo
	GetElement                      func(position int, cancellationToken CancellationToken) TextRange
	GetRelatedElements              func(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo
	GetClusterReferences            func(cancellationToken CancellationToken) []ClusterReference
	GetDatabaseReferences           func(cancellationToken CancellationToken) []DatabaseReference
	GetMinimalText                  func(kind MinimalTextKind, cancellationToken CancellationToken) string
	GetFormattedText                func(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText
	GetClientParameters             func() []ClientParameter
}

// NewCustomCodeService adapts public callbacks to CodeService. The supplied
// text is immutable and is the text used by CodeScript for block accounting.
func NewCustomCodeService(text string, callbacks CustomCodeServiceCallbacks) *CodeService {
	var getOutlines func(internal.OutliningOptions, internal.CancellationToken) internal.OutlineInfo
	if callbacks.GetOutlines != nil {
		getOutlines = func(options internal.OutliningOptions, cancellationToken internal.CancellationToken) internal.OutlineInfo {
			result := callbacks.GetOutlines(outliningOptionsFromInternal(options), cancellationToken)
			return result.toInternal()
		}
	}

	service := internal.NewCustomCodeService(text, internal.CustomCodeServiceCallbacks{
		Kind:                            callbacks.Kind,
		IsFeatureSupported:              callbacks.IsFeatureSupported,
		TryGetCachedDiagnostics:         callbacks.TryGetCachedDiagnostics,
		GetDiagnostics:                  callbacks.GetDiagnostics,
		TryGetCachedAnalyzerDiagnostics: callbacks.TryGetCachedAnalyzerDiagnostics,
		GetAnalyzerDiagnostics:          callbacks.GetAnalyzerDiagnostics,
		GetCombinedDiagnostics:          callbacks.GetCombinedDiagnostics,
		GetAnalyzers:                    callbacks.GetAnalyzers,
		GetCodeActions:                  callbacks.GetCodeActions,
		ApplyCodeAction:                 callbacks.ApplyCodeAction,
		GetClassifications:              callbacks.GetClassifications,
		GetOutlines:                     getOutlines,
		ShouldAutoComplete:              callbacks.ShouldAutoComplete,
		GetCompletionItems:              callbacks.GetCompletionItems,
		GetQuickInfo:                    callbacks.GetQuickInfo,
		GetElement:                      callbacks.GetElement,
		GetRelatedElements:              callbacks.GetRelatedElements,
		GetClusterReferences:            callbacks.GetClusterReferences,
		GetDatabaseReferences:           callbacks.GetDatabaseReferences,
		GetMinimalText:                  callbacks.GetMinimalText,
		GetFormattedText:                callbacks.GetFormattedText,
		GetClientParameters:             callbacks.GetClientParameters,
	})
	return wrapCodeService(service)
}
