package kustolang

import "testing"

func TestOffsetCodeServiceAdjustsRangesAndReferences(t *testing.T) {
	service := newMockCodeService()
	offset := NewOffsetCodeService(service, 10)

	classifications := offset.GetClassifications(15, 4, true, true, CancellationToken{})
	if service.lastClassificationStart != 5 || len(classifications.Classifications) != 1 || classifications.Classifications[0].Start != 12 {
		t.Fatalf("classifications start/result = start:%d result:%#v", service.lastClassificationStart, classifications)
	}

	parameters := offset.GetClientParameters()
	if len(parameters) != 1 || parameters[0].Start() != 13 || parameters[0].Index != nil {
		t.Fatalf("client parameters = %#v", parameters)
	}
	clusters := offset.GetClusterReferences(CancellationToken{})
	if len(clusters) != 1 || clusters[0].Start() != 14 {
		t.Fatalf("cluster references = %#v", clusters)
	}
	databases := offset.GetDatabaseReferences(CancellationToken{})
	if len(databases) != 1 || databases[0].Start() != 15 {
		t.Fatalf("database references = %#v", databases)
	}
	outlines := offset.GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(outlines.Ranges) != 1 || outlines.Ranges[0].Start != 16 {
		t.Fatalf("outlines = %#v", outlines)
	}
	completions := offset.GetCompletionItems(21, CompletionOptionsDefault, CancellationToken{})
	if service.lastCompletionPosition != 11 || completions.EditStart != 17 {
		t.Fatalf("completion position/result = pos:%d result:%#v", service.lastCompletionPosition, completions)
	}
	formatted := offset.GetFormattedText(FormattingOptionsDefault, 20, CancellationToken{})
	if service.lastFormatCaret != 10 || formatted.Position != 18 {
		t.Fatalf("formatted text = caret:%d result:%#v", service.lastFormatCaret, formatted)
	}
	element := offset.GetElement(30, CancellationToken{})
	if service.lastElementPosition != 20 || element.Start != 19 {
		t.Fatalf("element = pos:%d result:%#v", service.lastElementPosition, element)
	}
	related := offset.GetRelatedElements(31, FindRelatedOptionsSeeThroughVariables, CancellationToken{})
	if service.lastRelatedPosition != 21 || len(related.Elements) != 1 || related.Elements[0].Start() != 20 || related.Elements[0].CursorLeft != 21 || related.Elements[0].CursorRight != 22 {
		t.Fatalf("related elements = pos:%d result:%#v", service.lastRelatedPosition, related)
	}
}

func TestOffsetCodeServiceAdjustsDiagnosticsAndCombinedFiltering(t *testing.T) {
	service := newMockCodeService()
	offset := NewOffsetCodeService(service, 10)

	cached, ok := offset.TryGetCachedDiagnostics()
	if !ok || len(cached) != 1 || cached[0].Start != 11 {
		t.Fatalf("cached diagnostics = ok:%t result:%#v", ok, cached)
	}
	diagnostics := offset.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Start != 11 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	cachedAnalyzer, ok := offset.TryGetCachedAnalyzerDiagnostics()
	if !ok || len(cachedAnalyzer) != 1 || cachedAnalyzer[0].Start != 12 {
		t.Fatalf("cached analyzer diagnostics = ok:%t result:%#v", ok, cachedAnalyzer)
	}
	analyzerDiagnostics := offset.GetAnalyzerDiagnostics(true, CancellationToken{})
	if len(analyzerDiagnostics) != 1 || analyzerDiagnostics[0].Start != 12 {
		t.Fatalf("analyzer diagnostics = %#v", analyzerDiagnostics)
	}

	filter := ParseDisabledDiagnostics("KS900")
	combined := CodeServiceGetCombinedDiagnostics(offset, true, &filter, CancellationToken{})
	if len(combined) != 1 || combined[0].Code != "KS901" || combined[0].Start != 12 {
		t.Fatalf("combined diagnostics = %#v", combined)
	}
}

func TestOffsetCodeServiceDelegatesPositionsActionsAndPassThroughs(t *testing.T) {
	service := newMockCodeService()
	offset := NewOffsetCodeService(service, 10)

	if offset.Kind() != "Query" || offset.Text() != service.text {
		t.Fatalf("kind/text = %q/%q", offset.Kind(), offset.Text())
	}
	if !offset.IsFeatureSupported(CodeServiceFeatureCompletion, 25) || service.lastFeaturePosition != 15 {
		t.Fatalf("feature position = %d", service.lastFeaturePosition)
	}
	if !offset.ShouldAutoComplete(22, '.', CancellationToken{}) || service.lastAutoCompletePosition != 12 {
		t.Fatalf("autocomplete position = %d", service.lastAutoCompletePosition)
	}
	quickInfo := offset.GetQuickInfo(24, QuickInfoOptionsDefault, CancellationToken{})
	if service.lastQuickInfoPosition != 14 || quickInfo.Text() != "quick" {
		t.Fatalf("quick info = pos:%d result:%q", service.lastQuickInfoPosition, quickInfo.Text())
	}
	if got := offset.GetMinimalText(MinimalTextKindSingleLine, CancellationToken{}); got != "minimal" {
		t.Fatalf("minimal text = %q", got)
	}
	analyzers := offset.GetAnalyzers()
	if len(analyzers) != 1 || analyzers[0].Name != "mock" {
		t.Fatalf("analyzers = %#v", analyzers)
	}

	actions := offset.GetCodeActions(25, 22, 5, CodeActionOptionsDefault, true, "actor", CancellationToken{})
	if service.lastCodeActionPosition != 15 || service.lastSelectionStart != 12 || service.lastSelectionLength != 5 || service.lastActorName != "actor" || len(actions.Actions) != 1 {
		t.Fatalf("code actions state = pos:%d start:%d length:%d actor:%q result:%#v", service.lastCodeActionPosition, service.lastSelectionStart, service.lastSelectionLength, service.lastActorName, actions)
	}

	result := offset.ApplyCodeAction(NewCodeAction("kind", "title", "", nil), 30, CodeActionOptionsDefault, CancellationToken{})
	if service.lastApplyCaret != 20 || len(result.Actions) != 1 {
		t.Fatalf("apply action = caret:%d result:%#v", service.lastApplyCaret, result)
	}
	move, ok := result.Actions[0].(MoveCaretAction)
	if !ok || move.NewCaretPosition != 14 {
		t.Fatalf("adjusted move action = %#v", result.Actions[0])
	}

	reoffset := offset.WithOffset(3)
	if reoffset.service != service || reoffset.offset != 3 {
		t.Fatalf("WithOffset result = %#v", reoffset)
	}
}

type mockCodeService struct {
	text string

	lastClassificationStart  int
	lastCompletionPosition   int
	lastFormatCaret          int
	lastElementPosition      int
	lastRelatedPosition      int
	lastFeaturePosition      int
	lastAutoCompletePosition int
	lastQuickInfoPosition    int
	lastCodeActionPosition   int
	lastSelectionStart       int
	lastSelectionLength      int
	lastActorName            string
	lastApplyCaret           int
}

func newMockCodeService() *mockCodeService {
	return &mockCodeService{text: "01234567890123456789"}
}

func (s *mockCodeService) Kind() string {
	return CodeKindsQuery
}

func (s *mockCodeService) Text() string {
	return s.text
}

func (s *mockCodeService) IsFeatureSupported(feature string, position int) bool {
	s.lastFeaturePosition = position
	return feature == CodeServiceFeatureCompletion
}

func (s *mockCodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	return []Diagnostic{NewDiagnostic("KS900", "syntax").WithLocation(1, 2)}, true
}

func (s *mockCodeService) GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return []Diagnostic{NewDiagnostic("KS900", "syntax").WithLocation(1, 2)}
}

func (s *mockCodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	return []Diagnostic{NewDiagnostic("KS901", "analyzer").WithLocation(2, 3)}, true
}

func (s *mockCodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return []Diagnostic{NewDiagnostic("KS901", "analyzer").WithLocation(2, 3)}
}

func (s *mockCodeService) GetAnalyzers() []AnalyzerInfo {
	return []AnalyzerInfo{NewAnalyzerInfo("mock", []Diagnostic{NewDiagnostic("KS901", "analyzer")})}
}

func (s *mockCodeService) GetCodeActions(position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo {
	s.lastCodeActionPosition = position
	s.lastSelectionStart = selectionStart
	s.lastSelectionLength = selectionLength
	s.lastActorName = actorName
	return NewCodeActionInfo([]CodeAction{NewCodeAction("kind", "title", "", nil)})
}

func (s *mockCodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	s.lastApplyCaret = caretPosition
	return NewCodeActionResult(NewMoveCaretAction(4))
}

func (s *mockCodeService) GetClassifications(start int, length int, clipToRange bool, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo {
	s.lastClassificationStart = start
	return NewClassificationInfo([]ClassifiedRange{NewClassifiedRange(ClassificationKindColumn, 2, 3)})
}

func (s *mockCodeService) GetOutlines(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo {
	return NewOutlineInfo([]OutlineRange{NewOutlineRange(6, 4, "collapse")})
}

func (s *mockCodeService) ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) bool {
	s.lastAutoCompletePosition = position
	return key == '.'
}

func (s *mockCodeService) GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo {
	s.lastCompletionPosition = position
	return NewCompletionInfo(nil, 7, 2)
}

func (s *mockCodeService) GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo {
	s.lastQuickInfoPosition = position
	return NewQuickInfoText("quick")
}

func (s *mockCodeService) GetElement(position int, cancellationToken CancellationToken) TextRange {
	s.lastElementPosition = position
	return NewTextRange(9, 2)
}

func (s *mockCodeService) GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	s.lastRelatedPosition = position
	return NewRelatedInfo([]RelatedElement{NewRelatedElementWithCursors(10, 2, RelatedElementKindReference, 11, 12)}, 0)
}

func (s *mockCodeService) GetClusterReferences(cancellationToken CancellationToken) []ClusterReference {
	return []ClusterReference{NewClusterReference("cluster", 4, 2)}
}

func (s *mockCodeService) GetDatabaseReferences(cancellationToken CancellationToken) []DatabaseReference {
	return []DatabaseReference{NewDatabaseReference("db", "cluster", 5, 2)}
}

func (s *mockCodeService) GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string {
	return "minimal"
}

func (s *mockCodeService) GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText {
	s.lastFormatCaret = caretPosition
	return NewFormattedText("formatted", 8)
}

func (s *mockCodeService) GetClientParameters() []ClientParameter {
	index := 1
	return []ClientParameter{NewClientParameterAt("client", &index, 3, 4)}
}
