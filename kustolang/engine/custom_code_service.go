package kustolang

// CustomCodeServiceCallbacks supplies overrides for CustomCodeService.
// Nil callbacks use the corresponding CommonCodeService behavior.
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

// CustomCodeService adapts callbacks to the editor CodeService contract.
// Text is fixed at construction so CodeScript can safely measure and reuse
// blocks returned by custom factories.
type CustomCodeService struct {
	*CommonCodeService
	callbacks CustomCodeServiceCallbacks
}

func NewCustomCodeService(text string, callbacks CustomCodeServiceCallbacks) *CustomCodeService {
	return &CustomCodeService{
		CommonCodeService: NewCommonCodeService(text),
		callbacks:         callbacks,
	}
}

func (s *CustomCodeService) Kind() string {
	if s != nil && s.callbacks.Kind != nil {
		return s.callbacks.Kind()
	}
	return s.CommonCodeService.Kind()
}

func (s *CustomCodeService) IsFeatureSupported(feature string, position int) bool {
	if s != nil && s.callbacks.IsFeatureSupported != nil {
		return s.callbacks.IsFeatureSupported(feature, position)
	}
	return s.CommonCodeService.IsFeatureSupported(feature, position)
}

func (s *CustomCodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	if s != nil && s.callbacks.TryGetCachedDiagnostics != nil {
		diagnostics, ok := s.callbacks.TryGetCachedDiagnostics()
		if !ok {
			return nil, false
		}
		return snapshotDiagnostics(diagnostics), true
	}
	diagnostics, ok := s.CommonCodeService.TryGetCachedDiagnostics()
	if !ok {
		return nil, false
	}
	return snapshotDiagnostics(diagnostics), true
}

func (s *CustomCodeService) GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	if s != nil && s.callbacks.GetDiagnostics != nil {
		return snapshotDiagnostics(s.callbacks.GetDiagnostics(waitForAnalysis, cancellationToken))
	}
	return snapshotDiagnostics(s.CommonCodeService.GetDiagnostics(waitForAnalysis, cancellationToken))
}

func (s *CustomCodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	if s != nil && s.callbacks.TryGetCachedAnalyzerDiagnostics != nil {
		diagnostics, ok := s.callbacks.TryGetCachedAnalyzerDiagnostics()
		if !ok {
			return nil, false
		}
		return snapshotDiagnostics(diagnostics), true
	}
	diagnostics, ok := s.CommonCodeService.TryGetCachedAnalyzerDiagnostics()
	if !ok {
		return nil, false
	}
	return snapshotDiagnostics(diagnostics), true
}

func (s *CustomCodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	if s != nil && s.callbacks.GetAnalyzerDiagnostics != nil {
		return snapshotDiagnostics(s.callbacks.GetAnalyzerDiagnostics(waitForAnalysis, cancellationToken))
	}
	return snapshotDiagnostics(s.CommonCodeService.GetAnalyzerDiagnostics(waitForAnalysis, cancellationToken))
}

func (s *CustomCodeService) GetCombinedDiagnostics(waitForAnalysis bool, filter *DisabledDiagnostics, cancellationToken CancellationToken) []Diagnostic {
	if s != nil && s.callbacks.GetCombinedDiagnostics != nil {
		return snapshotDiagnostics(s.callbacks.GetCombinedDiagnostics(waitForAnalysis, filter, cancellationToken))
	}
	return codeServiceGetCombinedDiagnosticsDefault(s, waitForAnalysis, filter, cancellationToken)
}

func (s *CustomCodeService) GetAnalyzers() []AnalyzerInfo {
	if s != nil && s.callbacks.GetAnalyzers != nil {
		return cloneAnalyzerInfos(s.callbacks.GetAnalyzers())
	}
	return cloneAnalyzerInfos(s.CommonCodeService.GetAnalyzers())
}

func (s *CustomCodeService) GetCodeActions(position, selectionStart, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo {
	if s != nil && s.callbacks.GetCodeActions != nil {
		return CloneCodeActionInfo(s.callbacks.GetCodeActions(position, selectionStart, selectionLength, options, waitForAnalysis, actorName, cancellationToken))
	}
	return CloneCodeActionInfo(s.CommonCodeService.GetCodeActions(position, selectionStart, selectionLength, options, waitForAnalysis, actorName, cancellationToken))
}

func (s *CustomCodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	if s != nil && s.callbacks.ApplyCodeAction != nil {
		return CloneCodeActionResult(s.callbacks.ApplyCodeAction(action, caretPosition, options, cancellationToken))
	}
	return CloneCodeActionResult(s.CommonCodeService.ApplyCodeAction(action, caretPosition, options, cancellationToken))
}

func (s *CustomCodeService) GetClassifications(start, length int, clipToRange, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo {
	if s != nil && s.callbacks.GetClassifications != nil {
		return snapshotClassificationInfo(s.callbacks.GetClassifications(start, length, clipToRange, waitForAnalysis, cancellationToken))
	}
	classifications := []ClassifiedRange{NewClassifiedRange(ClassificationKindPlainText, start, length)}
	parameters := clientParameterClassifications(s.GetClientParameters(), start, length, clipToRange)
	return NewClassificationInfo(CommonCodeServiceAddClassifications(classifications, parameters))
}

func (s *CustomCodeService) GetOutlines(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo {
	if s != nil && s.callbacks.GetOutlines != nil {
		return snapshotOutlineInfo(s.callbacks.GetOutlines(options, cancellationToken))
	}
	return snapshotOutlineInfo(s.CommonCodeService.GetOutlines(options, cancellationToken))
}

func (s *CustomCodeService) ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) bool {
	if s != nil && s.callbacks.ShouldAutoComplete != nil {
		return s.callbacks.ShouldAutoComplete(position, key, cancellationToken)
	}
	return s.CommonCodeService.ShouldAutoComplete(position, key, cancellationToken)
}

func (s *CustomCodeService) GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo {
	if s != nil && s.callbacks.GetCompletionItems != nil {
		return snapshotCompletionInfo(s.callbacks.GetCompletionItems(position, options, cancellationToken))
	}
	return snapshotCompletionInfo(s.CommonCodeService.GetCompletionItems(position, options, cancellationToken))
}

func (s *CustomCodeService) GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo {
	if s != nil && s.callbacks.GetQuickInfo != nil {
		return snapshotQuickInfo(s.callbacks.GetQuickInfo(position, options, cancellationToken))
	}
	return snapshotQuickInfo(s.CommonCodeService.GetQuickInfo(position, options, cancellationToken))
}

func (s *CustomCodeService) GetElement(position int, cancellationToken CancellationToken) TextRange {
	if s != nil && s.callbacks.GetElement != nil {
		return s.callbacks.GetElement(position, cancellationToken)
	}
	return s.CommonCodeService.GetElement(position, cancellationToken)
}

func (s *CustomCodeService) GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	if s != nil && s.callbacks.GetRelatedElements != nil {
		return snapshotRelatedInfo(s.callbacks.GetRelatedElements(position, options, cancellationToken))
	}
	return snapshotRelatedInfo(s.CommonCodeService.GetRelatedElements(position, options, cancellationToken))
}

func (s *CustomCodeService) GetClusterReferences(cancellationToken CancellationToken) []ClusterReference {
	if s != nil && s.callbacks.GetClusterReferences != nil {
		return snapshotClusterReferences(s.callbacks.GetClusterReferences(cancellationToken))
	}
	return snapshotClusterReferences(s.CommonCodeService.GetClusterReferences(cancellationToken))
}

func (s *CustomCodeService) GetDatabaseReferences(cancellationToken CancellationToken) []DatabaseReference {
	if s != nil && s.callbacks.GetDatabaseReferences != nil {
		return snapshotDatabaseReferences(s.callbacks.GetDatabaseReferences(cancellationToken))
	}
	return snapshotDatabaseReferences(s.CommonCodeService.GetDatabaseReferences(cancellationToken))
}

func (s *CustomCodeService) GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string {
	if s != nil && s.callbacks.GetMinimalText != nil {
		return s.callbacks.GetMinimalText(kind, cancellationToken)
	}
	return s.CommonCodeService.GetMinimalText(kind, cancellationToken)
}

func (s *CustomCodeService) GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText {
	if s != nil && s.callbacks.GetFormattedText != nil {
		return s.callbacks.GetFormattedText(options, caretPosition, cancellationToken)
	}
	return s.CommonCodeService.GetFormattedText(options, caretPosition, cancellationToken)
}

func (s *CustomCodeService) GetClientParameters() []ClientParameter {
	if s != nil && s.callbacks.GetClientParameters != nil {
		return snapshotClientParameters(s.callbacks.GetClientParameters())
	}
	return snapshotClientParameters(s.CommonCodeService.GetClientParameters())
}

var _ CodeService = (*CustomCodeService)(nil)
