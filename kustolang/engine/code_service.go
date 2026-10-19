package kustolang

// Ported from:
//   src/Kusto.Language/Editor/CodeService.cs
//   src/Kusto.Language/Editor/OffsetCodeService.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CodeService interface {
	Kind() string
	Text() string
	IsFeatureSupported(feature string, position int) bool
	TryGetCachedDiagnostics() ([]Diagnostic, bool)
	GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic
	TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool)
	GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic
	GetAnalyzers() []AnalyzerInfo
	GetCodeActions(position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo
	ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult
	GetClassifications(start int, length int, clipToRange bool, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo
	GetOutlines(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo
	ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) bool
	GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo
	GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo
	GetElement(position int, cancellationToken CancellationToken) TextRange
	GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo
	GetClusterReferences(cancellationToken CancellationToken) []ClusterReference
	GetDatabaseReferences(cancellationToken CancellationToken) []DatabaseReference
	GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string
	GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText
	GetClientParameters() []ClientParameter
}

func CodeServiceGetCombinedDiagnostics(service CodeService, waitForAnalysis bool, filter *DisabledDiagnostics, cancellationToken CancellationToken) []Diagnostic {
	if service == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	if custom, ok := service.(interface {
		GetCombinedDiagnostics(bool, *DisabledDiagnostics, CancellationToken) []Diagnostic
	}); ok {
		return snapshotDiagnostics(custom.GetCombinedDiagnostics(waitForAnalysis, filter, cancellationToken))
	}
	return codeServiceGetCombinedDiagnosticsDefault(service, waitForAnalysis, filter, cancellationToken)
}

func codeServiceGetCombinedDiagnosticsDefault(service CodeService, waitForAnalysis bool, filter *DisabledDiagnostics, cancellationToken CancellationToken) []Diagnostic {
	base := service.GetDiagnostics(waitForAnalysis, cancellationToken)
	analyzer := service.GetAnalyzerDiagnostics(waitForAnalysis, cancellationToken)
	diagnostics := make([]Diagnostic, 0, len(base)+len(analyzer))
	diagnostics = append(diagnostics, base...)
	diagnostics = append(diagnostics, analyzer...)
	if filter == nil {
		return ListToReadOnly(diagnostics)
	}
	filtered := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if filter.IsDiagnosticEnabled(diagnostic) {
			filtered = append(filtered, diagnostic)
		}
	}
	return ListToReadOnly(filtered)
}

type OffsetCodeService struct {
	service CodeService
	offset  int
}

func NewOffsetCodeService(service CodeService, offset int) *OffsetCodeService {
	return &OffsetCodeService{service: service, offset: offset}
}

func (s *OffsetCodeService) WithOffset(offset int) *OffsetCodeService {
	if s == nil {
		return NewOffsetCodeService(nil, offset)
	}
	return NewOffsetCodeService(s.service, offset)
}

func (s *OffsetCodeService) Kind() string {
	return s.service.Kind()
}

func (s *OffsetCodeService) Text() string {
	return s.service.Text()
}

func (s *OffsetCodeService) IsFeatureSupported(feature string, position int) bool {
	if position != -1 {
		position -= s.offset
	}
	return s.service.IsFeatureSupported(feature, position)
}

func (s *OffsetCodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	diagnostics, ok := s.service.TryGetCachedDiagnostics()
	if !ok {
		return nil, false
	}
	return s.adjustDiagnostics(diagnostics), true
}

func (s *OffsetCodeService) GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return s.adjustDiagnostics(s.service.GetDiagnostics(waitForAnalysis, cancellationToken))
}

func (s *OffsetCodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	diagnostics, ok := s.service.TryGetCachedAnalyzerDiagnostics()
	if !ok {
		return nil, false
	}
	return s.adjustDiagnostics(diagnostics), true
}

func (s *OffsetCodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return s.adjustDiagnostics(s.service.GetAnalyzerDiagnostics(waitForAnalysis, cancellationToken))
}

func (s *OffsetCodeService) GetAnalyzers() []AnalyzerInfo {
	return cloneAnalyzerInfos(s.service.GetAnalyzers())
}

func (s *OffsetCodeService) GetCodeActions(position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo {
	position -= s.offset
	if selectionLength > 0 {
		selectionStart -= s.offset
		selectionEnd := minInt(selectionStart+selectionLength, utf16TextWidth(s.service.Text()))
		selectionLength = selectionEnd - selectionStart
	}
	return CloneCodeActionInfo(s.service.GetCodeActions(position, selectionStart, selectionLength, options, waitForAnalysis, actorName, cancellationToken))
}

func (s *OffsetCodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	result := s.service.ApplyCodeAction(action, caretPosition-s.offset, options, cancellationToken)
	return CloneCodeActionResult(result.WithAdjustedPosition(s.offset))
}

func (s *OffsetCodeService) GetClassifications(start int, length int, clipToRange bool, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo {
	result := s.service.GetClassifications(start-s.offset, length, clipToRange, waitForAnalysis, cancellationToken)
	if len(result.Classifications) == 0 || s.offset <= 0 {
		return snapshotClassificationInfo(result)
	}
	adjusted := make([]ClassifiedRange, 0, len(result.Classifications))
	for _, classification := range result.Classifications {
		adjusted = append(adjusted, NewClassifiedRange(classification.Kind, classification.Start+s.offset, classification.Length))
	}
	return NewClassificationInfo(adjusted)
}

func (s *OffsetCodeService) GetOutlines(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo {
	result := s.service.GetOutlines(options, cancellationToken)
	if len(result.Ranges) == 0 || s.offset <= 0 {
		return snapshotOutlineInfo(result)
	}
	adjusted := make([]OutlineRange, 0, len(result.Ranges))
	for _, outline := range result.Ranges {
		adjusted = append(adjusted, NewOutlineRange(outline.Start+s.offset, outline.Length, outline.CollapsedText))
	}
	return NewOutlineInfo(adjusted)
}

func (s *OffsetCodeService) ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) bool {
	return s.service.ShouldAutoComplete(position-s.offset, key, cancellationToken)
}

func (s *OffsetCodeService) GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo {
	result := s.service.GetCompletionItems(position-s.offset, options, cancellationToken)
	if s.offset > 0 {
		return NewCompletionInfo(result.Items, result.EditStart+s.offset, result.EditLength)
	}
	return snapshotCompletionInfo(result)
}

func (s *OffsetCodeService) GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo {
	return snapshotQuickInfo(s.service.GetQuickInfo(position-s.offset, options, cancellationToken))
}

func (s *OffsetCodeService) GetElement(position int, cancellationToken CancellationToken) TextRange {
	rng := s.service.GetElement(position-s.offset, cancellationToken)
	if s.offset > 0 {
		return NewTextRange(rng.Start+s.offset, rng.Length)
	}
	return rng
}

func (s *OffsetCodeService) GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	result := s.service.GetRelatedElements(position-s.offset, options, cancellationToken)
	if len(result.Elements) == 0 || s.offset <= 0 {
		return snapshotRelatedInfo(result)
	}
	adjusted := make([]RelatedElement, 0, len(result.Elements))
	for _, element := range result.Elements {
		adjusted = append(adjusted, NewRelatedElementWithCursors(
			element.Start()+s.offset,
			element.Length(),
			element.Kind,
			element.CursorLeft+s.offset,
			element.CursorRight+s.offset,
		))
	}
	return NewRelatedInfo(adjusted, result.CurrentIndex)
}

func (s *OffsetCodeService) GetClusterReferences(cancellationToken CancellationToken) []ClusterReference {
	result := s.service.GetClusterReferences(cancellationToken)
	if len(result) == 0 || s.offset <= 0 {
		return snapshotClusterReferences(result)
	}
	adjusted := make([]ClusterReference, 0, len(result))
	for _, reference := range result {
		adjusted = append(adjusted, NewClusterReference(reference.Cluster, reference.Start()+s.offset, reference.Length()))
	}
	return ListToReadOnly(adjusted)
}

func (s *OffsetCodeService) GetDatabaseReferences(cancellationToken CancellationToken) []DatabaseReference {
	result := s.service.GetDatabaseReferences(cancellationToken)
	if len(result) == 0 || s.offset <= 0 {
		return snapshotDatabaseReferences(result)
	}
	adjusted := make([]DatabaseReference, 0, len(result))
	for _, reference := range result {
		adjusted = append(adjusted, NewDatabaseReference(reference.Database, reference.Cluster, reference.Start()+s.offset, reference.Length()))
	}
	return ListToReadOnly(adjusted)
}

func (s *OffsetCodeService) GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string {
	return s.service.GetMinimalText(kind, cancellationToken)
}

func (s *OffsetCodeService) GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText {
	result := s.service.GetFormattedText(options, caretPosition-s.offset, cancellationToken)
	if s.offset > 0 {
		return NewFormattedText(result.Text, result.Position+s.offset)
	}
	return result
}

func (s *OffsetCodeService) GetClientParameters() []ClientParameter {
	result := s.service.GetClientParameters()
	if len(result) == 0 || s.offset <= 0 {
		return snapshotClientParameters(result)
	}
	adjusted := make([]ClientParameter, 0, len(result))
	for _, parameter := range result {
		adjusted = append(adjusted, NewClientParameterReference(parameter.Name, parameter.Start()+s.offset, parameter.Length()))
	}
	return ListToReadOnly(adjusted)
}

func (s *OffsetCodeService) adjustDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	if len(diagnostics) == 0 || s.offset <= 0 {
		return snapshotDiagnostics(diagnostics)
	}
	adjusted := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		adjusted = append(adjusted, diagnostic.WithLocation(diagnostic.Start+s.offset, diagnostic.Length))
	}
	return ListToReadOnly(adjusted)
}

func snapshotDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	return ListToReadOnly(diagnostics)
}

func snapshotClassificationInfo(info ClassificationInfo) ClassificationInfo {
	return NewClassificationInfo(info.Classifications)
}

func snapshotOutlineInfo(info OutlineInfo) OutlineInfo {
	return NewOutlineInfo(info.Ranges)
}

func snapshotCompletionInfo(info CompletionInfo) CompletionInfo {
	return NewCompletionInfo(info.Items, info.EditStart, info.EditLength)
}

func snapshotQuickInfo(info QuickInfo) QuickInfo {
	return NewQuickInfoItems(info.Items)
}

func snapshotRelatedInfo(info RelatedInfo) RelatedInfo {
	return NewRelatedInfo(info.Elements, info.CurrentIndex)
}

func snapshotClusterReferences(references []ClusterReference) []ClusterReference {
	if len(references) == 0 {
		return EmptyReadOnlyList[ClusterReference]()
	}
	result := make([]ClusterReference, 0, len(references))
	for _, reference := range references {
		result = append(result, NewClusterReference(reference.Cluster, reference.Start(), reference.Length()))
	}
	return ListToReadOnly(result)
}

func snapshotDatabaseReferences(references []DatabaseReference) []DatabaseReference {
	if len(references) == 0 {
		return EmptyReadOnlyList[DatabaseReference]()
	}
	result := make([]DatabaseReference, 0, len(references))
	for _, reference := range references {
		result = append(result, NewDatabaseReference(reference.Database, reference.Cluster, reference.Start(), reference.Length()))
	}
	return ListToReadOnly(result)
}

func snapshotClientParameters(parameters []ClientParameter) []ClientParameter {
	if len(parameters) == 0 {
		return EmptyReadOnlyList[ClientParameter]()
	}
	result := make([]ClientParameter, 0, len(parameters))
	for _, parameter := range parameters {
		result = append(result, NewClientParameter(parameter.Name, parameter.Index, parameter.Range))
	}
	return ListToReadOnly(result)
}
