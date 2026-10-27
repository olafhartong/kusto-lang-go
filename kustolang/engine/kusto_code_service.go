package kustolang

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Ported from selected supported behavior in:
//   src/Kusto.Language/Editor/Kusto/KustoCodeService.cs
//   src/Kusto.Language/Editor/Kusto/KustoCodeServiceFactory.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.
//
// Completion, quick-info, extraction, structural related-element behavior and
// formatting are routed to native helpers. Remaining completion depth and
// editor behavior stay explicitly tracked as incomplete upstream units.

type KustoCodeService struct {
	*CommonCodeService

	kind    string
	globals *GlobalState

	codeInitMu       sync.Mutex
	unboundCodeState atomic.Pointer[kustoCodeServiceCodeState]
	boundCodeState   atomic.Pointer[kustoCodeServiceCodeState]
	codeFailureState atomic.Pointer[kustoCodeServiceCodeState]

	diagnosticsInitMu sync.Mutex
	diagnosticsState  atomic.Pointer[kustoCodeServiceDiagnosticsState]

	analyzerDiagnosticsInitMu sync.Mutex
	analyzerDiagnosticsState  atomic.Pointer[kustoCodeServiceDiagnosticsState]

	analyzerInfosInitMu sync.Mutex
	analyzerInfosState  atomic.Pointer[kustoCodeServiceAnalyzerInfosState]

	fixersInitMu sync.Mutex
	fixersState  atomic.Pointer[kustoCodeServiceFixersState]

	analyzers []KustoAnalyzer
	actors    []KustoActor
}

type kustoCodeServiceCodeState struct {
	code      *KustoCode
	recovered any
}

type kustoCodeServiceDiagnosticsState struct {
	diagnostics []Diagnostic
}

type kustoCodeServiceAnalyzerInfosState struct {
	infos []AnalyzerInfo
}

type kustoCodeServiceFixersState struct {
	fixers           []KustoFixer
	byDiagnosticCode map[string]KustoFixer
	byName           map[string]KustoFixer
}

func NewKustoCodeService(text string, globals ...*GlobalState) *KustoCodeService {
	globalState := DefaultGlobalState
	if len(globals) > 0 && globals[0] != nil {
		globalState = globals[0]
	}
	return newKustoCodeService(text, globalState, nil, nil, nil)
}

func NewKustoCodeServiceForCode(code *KustoCode) *KustoCodeService {
	if code == nil {
		return NewKustoCodeService("")
	}
	return newKustoCodeService(code.Text, code.Globals, code, nil, nil)
}

func newKustoCodeService(text string, globals *GlobalState, code *KustoCode, analyzers []KustoAnalyzer, actors []KustoActor) *KustoCodeService {
	if globals == nil {
		globals = DefaultGlobalState
	}
	if analyzers == nil {
		analyzers = KustoAnalyzers.All()
	}
	if actors == nil {
		actors = KustoActorsAll()
	}
	service := &KustoCodeService{
		CommonCodeService: NewCommonCodeService(text),
		kind:              string(CodeKindOf(text)),
		globals:           globals,
		analyzers:         ListToReadOnly(analyzers),
		actors:            ListToReadOnly(actors),
	}
	if code != nil {
		service.boundCodeState.Store(&kustoCodeServiceCodeState{code: code})
	}
	return service
}

func (s *KustoCodeService) WithAnalyzers(analyzers []KustoAnalyzer) *KustoCodeService {
	if s == nil {
		return newKustoCodeService("", DefaultGlobalState, nil, analyzers, nil)
	}
	var code *KustoCode
	if state := s.boundCodeState.Load(); state != nil {
		code = state.code
	}
	return newKustoCodeService(s.Text(), s.globals, code, analyzers, s.actors)
}

func (s *KustoCodeService) WithActors(actors []KustoActor) *KustoCodeService {
	if s == nil {
		return newKustoCodeService("", DefaultGlobalState, nil, nil, actors)
	}
	var code *KustoCode
	if state := s.boundCodeState.Load(); state != nil {
		code = state.code
	}
	return newKustoCodeService(s.Text(), s.globals, code, s.analyzers, actors)
}

func (s *KustoCodeService) Kind() string {
	if s == nil {
		return CodeKindsQuery
	}
	return s.kind
}

func (s *KustoCodeService) IsFeatureSupported(feature string, position int) bool {
	switch s.Kind() {
	case CodeKindsQuery, CodeKindsCommand:
		// Upstream treats feature names as an open contract for Kusto query and
		// command services; callers may probe newer features by name.
		return true
	case CodeKindsDirective:
		return feature == CodeServiceFeatureClassification
	default:
		return false
	}
}

func (s *KustoCodeService) tryGetBoundOrUnboundCode(cancellationToken CancellationToken, waitForAnalysis bool) (*KustoCode, bool) {
	if s == nil {
		return nil, false
	}
	if state := s.boundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if state := s.unboundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if s.codeFailureState.Load() != nil || !waitForAnalysis {
		return nil, false
	}

	s.codeInitMu.Lock()
	defer s.codeInitMu.Unlock()
	if state := s.boundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if state := s.unboundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if s.codeFailureState.Load() != nil {
		return nil, false
	}

	state := &kustoCodeServiceCodeState{}
	if s.canBeParsed() {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					state.recovered = recovered
				}
			}()
			state.code = ParseWithGlobals(s.Text(), s.globals)
		}()
	}
	if state.recovered != nil {
		s.codeFailureState.Store(state)
		return nil, false
	}
	if state.code != nil {
		s.unboundCodeState.Store(state)
	}

	return state.code, state.code != nil
}

func (s *KustoCodeService) tryGetBoundCode(cancellationToken CancellationToken, waitForAnalysis bool) (*KustoCode, bool) {
	if s == nil {
		return nil, false
	}
	if state := s.boundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if s.codeFailureState.Load() != nil || !waitForAnalysis {
		return nil, false
	}

	s.codeInitMu.Lock()
	defer s.codeInitMu.Unlock()
	if state := s.boundCodeState.Load(); state != nil {
		return state.code, state.code != nil
	}
	if s.codeFailureState.Load() != nil {
		return nil, false
	}

	state := &kustoCodeServiceCodeState{}
	if s.canBeParsed() {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					state.recovered = recovered
				}
			}()
			if unbound := s.unboundCodeState.Load(); unbound != nil && unbound.code != nil {
				state.code = unbound.code.AnalyzeWithCancellation(s.globals, cancellationToken)
			} else {
				state.code = ParseAndAnalyzeWithCancellation(s.Text(), s.globals, cancellationToken)
			}
			cancellationToken.ThrowIfCancellationRequested()
		}()
	}
	if state.recovered != nil {
		if kustoCodeServiceIsCancellationPanic(state.recovered) {
			return nil, false
		}
		s.codeFailureState.Store(state)
		return nil, false
	}
	if state.code != nil {
		s.boundCodeState.Store(state)
	}

	return state.code, state.code != nil
}

// tryGetCode retains the original native helper shape for internal callers and
// tests. Code-service operations should choose the bound or parse-only path.
func (s *KustoCodeService) tryGetCode(waitForAnalysis bool) (*KustoCode, bool) {
	return s.tryGetBoundCode(CancellationToken{}, waitForAnalysis)
}

func (s *KustoCodeService) recoveredCodePanic() any {
	if s == nil {
		return nil
	}
	if state := s.codeFailureState.Load(); state != nil {
		return state.recovered
	}
	return nil
}

func (s *KustoCodeService) canBeParsed() bool {
	if s == nil {
		return false
	}
	globals := s.globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	return utf16TextWidth(s.Text()) <= GlobalStateGetProperty(globals, Properties.MaxParseTextSize)
}

func (s *KustoCodeService) canBeAnalyzed(code *KustoCode) bool {
	if s == nil || code == nil || code.Tree == nil {
		return false
	}
	globals := s.globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	return code.Tree.IsSafeToRecurse(globals)
}

func (s *KustoCodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	diagnostics, ok := s.tryGetCachedDiagnosticsBytes()
	if !ok {
		return nil, false
	}
	return kustoCodeServiceDiagnosticsToUTF16(s.Text(), diagnostics), true
}

func (s *KustoCodeService) tryGetCachedDiagnosticsBytes() ([]Diagnostic, bool) {
	if s == nil {
		return nil, false
	}
	state := s.diagnosticsState.Load()
	if state == nil {
		return nil, false
	}
	return ListToReadOnly(state.diagnostics), true
}

func (s *KustoCodeService) GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	if s == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	return kustoCodeServiceDiagnosticsToUTF16(s.Text(), s.getDiagnosticsBytes(waitForAnalysis, cancellationToken))
}

func (s *KustoCodeService) getDiagnosticsBytes(waitForAnalysis bool, cancellationToken CancellationToken) (result []Diagnostic) {
	if s == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	if state := s.diagnosticsState.Load(); state != nil {
		return ListToReadOnly(state.diagnostics)
	}

	code, hasCode := s.tryGetBoundCode(cancellationToken, waitForAnalysis)
	if cancellationToken.Err() != nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	s.diagnosticsInitMu.Lock()
	defer s.diagnosticsInitMu.Unlock()
	if state := s.diagnosticsState.Load(); state != nil {
		return ListToReadOnly(state.diagnostics)
	}

	var diagnostics []Diagnostic
	ready := false
	switch {
	case hasCode:
		ready = true
		var recovered any
		func() {
			defer func() {
				recovered = recover()
			}()
			diagnostics = code.GetDiagnosticsWithCancellation(cancellationToken)
			cancellationToken.ThrowIfCancellationRequested()
		}()
		if recovered != nil {
			if kustoCodeServiceIsCancellationPanic(recovered) {
				return EmptyReadOnlyList[Diagnostic]()
			}
			diagnostics = []Diagnostic{DiagnosticInternalFailure()}
		}
	case s.recoveredCodePanic() != nil:
		diagnostics = []Diagnostic{DiagnosticInternalFailure()}
		ready = true
	case !s.canBeParsed():
		diagnostics = []Diagnostic{DiagnosticQueryTextSizeExceeded()}
		ready = true
	}
	if !ready {
		return EmptyReadOnlyList[Diagnostic]()
	}
	state := &kustoCodeServiceDiagnosticsState{diagnostics: ListToReadOnly(diagnostics)}
	s.diagnosticsState.Store(state)
	return ListToReadOnly(state.diagnostics)
}

func (s *KustoCodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	diagnostics, ok := s.tryGetCachedAnalyzerDiagnosticsBytes()
	if !ok {
		return nil, false
	}
	return kustoCodeServiceDiagnosticsToUTF16(s.Text(), diagnostics), true
}

func (s *KustoCodeService) tryGetCachedAnalyzerDiagnosticsBytes() ([]Diagnostic, bool) {
	if s == nil {
		return nil, false
	}
	state := s.analyzerDiagnosticsState.Load()
	if state == nil {
		return nil, false
	}
	return ListToReadOnly(state.diagnostics), true
}

func (s *KustoCodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	if s == nil {
		return EmptyReadOnlyList[Diagnostic]()
	}
	return kustoCodeServiceDiagnosticsToUTF16(s.Text(), s.getAnalyzerDiagnosticsBytes(waitForAnalysis, cancellationToken))
}

func (s *KustoCodeService) getAnalyzerDiagnosticsBytes(waitForAnalysis bool, cancellationToken CancellationToken) (result []Diagnostic) {
	result = EmptyReadOnlyList[Diagnostic]()
	defer func() {
		if recovered := recover(); recovered != nil {
			if kustoCodeServiceIsCancellationPanic(recovered) {
				panic(recovered)
			}
			result = EmptyReadOnlyList[Diagnostic]()
		}
	}()
	if s == nil {
		return result
	}
	if state := s.analyzerDiagnosticsState.Load(); state != nil {
		return ListToReadOnly(state.diagnostics)
	}
	if !waitForAnalysis {
		return result
	}

	s.analyzerDiagnosticsInitMu.Lock()
	defer s.analyzerDiagnosticsInitMu.Unlock()
	if state := s.analyzerDiagnosticsState.Load(); state != nil {
		return ListToReadOnly(state.diagnostics)
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || !s.canBeAnalyzed(code) {
		return result
	}

	var analyzerDiagnostics []Diagnostic
	for _, analyzer := range s.analyzers {
		cancellationToken.ThrowIfCancellationRequested()
		analyzerDiagnostics = kustoCodeServiceAnalyze(analyzer, code, analyzerDiagnostics, cancellationToken)
	}
	binderDiagnostics := s.getDiagnosticsBytes(true, cancellationToken)
	state := &kustoCodeServiceDiagnosticsState{
		diagnostics: kustoCodeServiceDistinctNewDiagnostics(binderDiagnostics, analyzerDiagnostics),
	}
	s.analyzerDiagnosticsState.Store(state)
	return ListToReadOnly(state.diagnostics)
}

func kustoCodeServiceAnalyze(analyzer KustoAnalyzer, code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) (result []Diagnostic) {
	if analyzer == nil {
		return diagnostics
	}
	result = diagnostics
	defer func() {
		if recovered := recover(); recovered != nil {
			result = append(result, DiagnosticAnalyzerFailure(analyzer.Name(), fmt.Sprint(recovered)))
		}
	}()
	return analyzer.Analyze(code, diagnostics, cancellationToken)
}

func kustoCodeServiceIsCancellationPanic(recovered any) bool {
	err, ok := recovered.(error)
	if !ok {
		return false
	}
	var canceled OperationCanceledError
	return errors.As(err, &canceled)
}

func kustoCodeServiceDistinctNewDiagnostics(existing []Diagnostic, additions []Diagnostic) []Diagnostic {
	seen := map[diagnosticIdentity]struct{}{}
	for _, diagnostic := range existing {
		if identity, ok := diagnosticIdentityForEquality(diagnostic); ok {
			seen[identity] = struct{}{}
		}
	}
	result := make([]Diagnostic, 0, len(additions))
	for _, diagnostic := range additions {
		if identity, hasIdentity := diagnosticIdentityForEquality(diagnostic); hasIdentity {
			if _, exists := seen[identity]; exists {
				continue
			}
			seen[identity] = struct{}{}
		}
		result = append(result, diagnostic)
	}
	return ListToReadOnly(result)
}

func (s *KustoCodeService) GetAnalyzers() (result []AnalyzerInfo) {
	result = EmptyReadOnlyList[AnalyzerInfo]()
	defer func() {
		if recover() != nil {
			result = EmptyReadOnlyList[AnalyzerInfo]()
		}
	}()
	if s == nil {
		return result
	}
	return cloneAnalyzerInfos(s.kustoCodeServiceGetAnalyzerInfosState().infos)
}

func (s *KustoCodeService) kustoCodeServiceGetAnalyzerInfosState() *kustoCodeServiceAnalyzerInfosState {
	if s == nil {
		return &kustoCodeServiceAnalyzerInfosState{infos: EmptyReadOnlyList[AnalyzerInfo]()}
	}
	if state := s.analyzerInfosState.Load(); state != nil {
		return state
	}
	s.analyzerInfosInitMu.Lock()
	defer s.analyzerInfosInitMu.Unlock()
	if state := s.analyzerInfosState.Load(); state != nil {
		return state
	}
	infos := make([]AnalyzerInfo, 0, len(s.analyzers))
	for _, analyzer := range s.analyzers {
		if analyzer != nil {
			infos = append(infos, NewAnalyzerInfo(analyzer.Name(), analyzer.Diagnostics()))
		}
	}
	state := &kustoCodeServiceAnalyzerInfosState{infos: ListToReadOnly(infos)}
	s.analyzerInfosState.Store(state)
	return state
}

func (s *KustoCodeService) GetCodeActions(position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) (result CodeActionInfo) {
	result = CodeActionInfoNoActions
	defer func() {
		if recover() != nil {
			result = CodeActionInfoNoActions
		}
	}()
	if s == nil {
		return result
	}
	if actorName != "" {
		if _, known := kustoActorByName(s.actors, actorName); !known {
			return CodeActionInfoNoActions
		}
	}
	code, ok := s.tryGetBoundCode(cancellationToken, waitForAnalysis)
	if !ok || code == nil {
		return CodeActionInfoNoActions
	}
	actions := []CodeAction{}
	for _, actor := range s.actors {
		if actor == nil || actorName != "" && actor.Name() != actorName {
			continue
		}
		for _, action := range kustoActorGetActions(actor, s, code, position, selectionStart, selectionLength, options, waitForAnalysis, cancellationToken) {
			actions = append(actions, kustoCodeServiceAddActionData(action, actor.Name()))
		}
	}
	if len(actions) > 1 {
		kustoSortCodeActions(actions)
	}
	if len(actions) == 0 {
		return CodeActionInfoNoActions
	}
	return NewCodeActionInfo(actions)
}

func (s *KustoCodeService) kustoCodeServiceGetDiagnosticFixActions(code *KustoCode, position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, cancellationToken CancellationToken) []CodeAction {
	diagnostics, ok := s.kustoCodeServiceAnalyzerDiagnosticsForActions(waitForAnalysis, cancellationToken)
	if !ok || len(diagnostics) == 0 {
		return nil
	}

	diagnosticsAtPosition := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if options.DiagnosticFilter.IsDiagnosticEnabled(diagnostic) &&
			TextRangesOverlap(position, 0, diagnostic.Start, diagnostic.Length) {
			diagnosticsAtPosition = append(diagnosticsAtPosition, diagnostic)
		}
	}
	if len(diagnosticsAtPosition) == 0 {
		return nil
	}

	if selectionLength == 0 {
		selectionStart = 0
		selectionLength = len(code.Text)
	}
	diagnosticsInSelectionByCode := map[string][]Diagnostic(nil)
	if position >= selectionStart && position < selectionStart+selectionLength {
		diagnosticsInSelectionByCode = kustoCodeServiceDiagnosticsByCode(diagnostics, selectionStart, selectionLength, options.DiagnosticFilter)
	}

	fixersByCode := s.kustoCodeServiceFixersByDiagnosticCode()
	actions := []CodeAction{}
	fixActions := []CodeAction{}
	for _, diagnostic := range diagnosticsAtPosition {
		cancellationToken.ThrowIfCancellationRequested()
		fixer := fixersByCode[diagnostic.Code]
		if fixer == nil {
			continue
		}
		selectionDiagnostics := EmptyReadOnlyList[Diagnostic]()
		if diagnosticsInSelectionByCode != nil {
			if matches := diagnosticsInSelectionByCode[diagnostic.Code]; len(matches) > 0 {
				selectionDiagnostics = matches
			}
		}
		fixActions = fixActions[:0]
		fixActions = fixer.GetFixActions(code, diagnostic, selectionDiagnostics, options, fixActions, cancellationToken)
		for _, action := range fixActions {
			action = kustoCodeServiceAddActionData(action, fixer.Name())
			actions = append(actions, action)
		}
	}
	return actions
}

func (s *KustoCodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) (result CodeActionResult) {
	result = NewCodeActionFailure()
	defer func() {
		if recover() != nil {
			result = NewCodeActionFailure()
		}
	}()
	if s == nil || action == nil {
		return NewCodeActionFailure("Unknown actor")
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil {
		return NewCodeActionFailure("No semantic information available for this query")
	}
	data := action.ApplyData()
	if len(data) == 0 {
		return NewCodeActionFailure("Unknown actor")
	}
	actorName := data[len(data)-1]
	action = action.RemoveData(1)
	actor, known := kustoActorByName(s.actors, actorName)
	if !known {
		return NewCodeActionFailure("Unknown actor: " + actorName)
	}
	return kustoActorApply(actor, s, code, action, caretPosition, options, cancellationToken)
}

const (
	kustoFunctionInlinerActorName      = KustoActorNameFunctionInliner
	kustoExtractExpressionActorName    = KustoActorNameExtractExpression
	kustoConvertToMacroExpandActorName = KustoActorNameConvertToMacroExpand
	kustoRenameActorName               = KustoActorNameRename
	kustoDiagnosticFixActorName        = KustoActorNameDiagnosticFixer
)

func (s *KustoCodeService) kustoCodeServiceAnalyzerDiagnosticsForActions(waitForAnalysis bool, cancellationToken CancellationToken) ([]Diagnostic, bool) {
	if s == nil {
		return nil, false
	}
	if waitForAnalysis {
		return s.getAnalyzerDiagnosticsBytes(waitForAnalysis, cancellationToken), true
	}
	return s.tryGetCachedAnalyzerDiagnosticsBytes()
}

func (s *KustoCodeService) kustoCodeServiceFixersByDiagnosticCode() map[string]KustoFixer {
	return s.kustoCodeServiceGetFixersState().byDiagnosticCode
}

func (s *KustoCodeService) kustoCodeServiceFixersByName() map[string]KustoFixer {
	return s.kustoCodeServiceGetFixersState().byName
}

func (s *KustoCodeService) kustoCodeServiceGetFixersState() *kustoCodeServiceFixersState {
	if s == nil {
		return &kustoCodeServiceFixersState{
			fixers:           EmptyReadOnlyList[KustoFixer](),
			byDiagnosticCode: map[string]KustoFixer{},
			byName:           map[string]KustoFixer{},
		}
	}
	if state := s.fixersState.Load(); state != nil {
		return state
	}
	// Initialize analyzer metadata first because the native fixer delegates its
	// diagnostic metadata to the analyzer.
	s.kustoCodeServiceGetAnalyzerInfosState()
	s.fixersInitMu.Lock()
	defer s.fixersInitMu.Unlock()
	if state := s.fixersState.Load(); state != nil {
		return state
	}

	fixers := make([]KustoFixer, 0, len(s.analyzers))
	byDiagnosticCode := make(map[string]KustoFixer, len(s.analyzers))
	byName := make(map[string]KustoFixer, len(s.analyzers))
	for _, analyzer := range s.analyzers {
		if analyzer != nil {
			fixer := analyzer.Fixer()
			fixers = append(fixers, fixer)
			if fixer == nil {
				continue
			}
			byName[fixer.Name()] = fixer
			for _, diagnostic := range fixer.Diagnostics() {
				byDiagnosticCode[diagnostic.Code] = fixer
			}
		}
	}
	state := &kustoCodeServiceFixersState{
		fixers:           ListToReadOnly(fixers),
		byDiagnosticCode: byDiagnosticCode,
		byName:           byName,
	}
	s.fixersState.Store(state)
	return state
}

func kustoCodeServiceDiagnosticsByCode(diagnostics []Diagnostic, selectionStart int, selectionLength int, filter DisabledDiagnostics) map[string][]Diagnostic {
	result := map[string][]Diagnostic{}
	for _, diagnostic := range diagnostics {
		if filter.IsDiagnosticEnabled(diagnostic) &&
			TextRangesOverlap(selectionStart, selectionLength, diagnostic.Start, diagnostic.Length) {
			result[diagnostic.Code] = append(result[diagnostic.Code], diagnostic)
		}
	}
	return result
}

func kustoCodeServiceAddActionData(action CodeAction, data ...string) CodeAction {
	switch typed := action.(type) {
	case nil:
		return action
	case *MenuAction:
		if typed == nil {
			return action
		}
		actions := make([]CodeAction, 0, len(typed.Actions))
		for _, nested := range typed.Actions {
			actions = append(actions, kustoCodeServiceAddActionData(nested, data...))
		}
		return typed.WithActions(actions)
	case ApplyAction:
		return typed.AddDataStrings(data...)
	default:
		return action
	}
}

func (s *KustoCodeService) GetClassifications(start int, length int, clipToRange bool, waitForAnalysis bool, cancellationToken CancellationToken) (result ClassificationInfo) {
	result = ClassificationInfoEmpty
	defer func() {
		if recover() != nil {
			result = ClassificationInfoEmpty
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, waitForAnalysis)
	if !ok || code == nil || code.Syntax() == nil {
		return ClassificationInfoEmpty
	}
	text := s.Text()
	byteStart, byteLength, rangeOK := utf16RangeToBytes(text, start, length)
	if !rangeOK {
		return ClassificationInfoEmpty
	}

	classifications := KustoClassifierGetClassifications(code.Syntax(), byteStart, byteLength, clipToRange, nil, cancellationToken)
	clientParameterClassifications := s.CommonCodeService.getClientParametersClassificationsBytes(byteStart, byteLength, clipToRange)
	if len(clientParameterClassifications) > 0 {
		classifications = CommonCodeServiceAddClassifications(classifications, clientParameterClassifications)
	}
	converted := make([]ClassifiedRange, 0, len(classifications))
	for _, classification := range classifications {
		convertedStart, convertedLength := byteRangeToUTF16(text, classification.Start, classification.Length)
		converted = append(converted, NewClassifiedRange(classification.Kind, convertedStart, convertedLength))
	}
	return NewClassificationInfo(converted)
}

func (s *KustoCodeService) GetClientParameters() []ClientParameter {
	if s == nil {
		return EmptyReadOnlyList[ClientParameter]()
	}
	parameters := s.CommonCodeService.getClientParametersBytes()
	if len(parameters) == 0 {
		return parameters
	}
	code, ok := s.tryGetBoundOrUnboundCode(CancellationToken{}, true)
	if !ok || code == nil {
		return clientParametersToUTF16(s.Text(), parameters)
	}
	filtered := make([]ClientParameter, 0, len(parameters))
	for _, parameter := range parameters {
		token := kustoCodeServiceTokenAt(code, parameter.Start())
		if token != nil && parameter.Start() >= token.TextStart() {
			start, length := byteRangeToUTF16(s.Text(), parameter.Start(), parameter.Length())
			filtered = append(filtered, NewClientParameterAt(parameter.Name, parameter.Index, start, length))
		}
	}
	return ListToReadOnly(filtered)
}

func (s *KustoCodeService) GetOutlines(options OutliningOptions, cancellationToken CancellationToken) (result OutlineInfo) {
	if s == nil {
		return OutlineInfoEmpty
	}
	defer func() {
		if recover() != nil {
			result = s.CommonCodeService.GetOutlines(options, cancellationToken)
		}
	}()
	if code, ok := s.tryGetBoundOrUnboundCode(cancellationToken, true); ok && code != nil {
		info := KustoOutlinerGetOutlines(code, options)
		converted := make([]OutlineRange, 0, len(info.Ranges))
		for _, outline := range info.Ranges {
			start, length := byteRangeToUTF16(s.Text(), outline.Start, outline.Length)
			converted = append(converted, NewOutlineRange(start, length, outline.CollapsedText))
		}
		return NewOutlineInfo(converted)
	}
	return s.CommonCodeService.GetOutlines(options, cancellationToken)
}

func (s *KustoCodeService) ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) (result bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = false
		}
	}()
	if s == nil {
		return false
	}
	code, ok := s.tryGetBoundOrUnboundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) {
		return false
	}
	bytePosition := utf16PositionToByteOffset(s.Text(), position)
	return newKustoCompleter(code, CompletionOptionsDefault, cancellationToken).shouldAutoComplete(bytePosition, key)
}

func (s *KustoCodeService) GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) (result CompletionInfo) {
	result = CompletionInfoEmpty
	defer func() {
		if recovered := recover(); recovered != nil {
			result = CompletionInfoEmpty
		}
	}()
	if s == nil {
		return CompletionInfoEmpty
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) {
		return CompletionInfoEmpty
	}
	text := s.Text()
	bytePosition := utf16PositionToByteOffset(text, position)
	completion := newKustoCompleter(code, options, cancellationToken).getCompletionItems(bytePosition)
	editStart := byteOffsetToUTF16Position(text, completion.EditStart)
	editEnd := byteOffsetToUTF16Position(text, completion.EditStart+completion.EditLength)
	return NewCompletionInfo(completion.Items, editStart, editEnd-editStart)
}

func (s *KustoCodeService) GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) (result QuickInfo) {
	result = QuickInfoEmpty
	defer func() {
		if recover() != nil {
			result = QuickInfoEmpty
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) {
		if recovered := s.recoveredCodePanic(); recovered != nil {
			return NewQuickInfoText(fmt.Sprint(recovered))
		}
		return result
	}
	bytePosition := utf16PositionToByteOffset(s.Text(), position)
	result = newKustoQuickInfoBuilder(s, code, options).getQuickInfo(bytePosition, cancellationToken)
	return result
}

func (s *KustoCodeService) GetElement(position int, cancellationToken CancellationToken) (result TextRange) {
	result = NewTextRange(0, 0)
	defer func() {
		if recover() != nil {
			result = NewTextRange(0, 0)
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || code.Syntax() == nil {
		return result
	}
	text := s.Text()
	bytePosition := utf16PositionToByteOffset(text, position)

	refs := kustoServiceTokenRefs(code.Syntax())
	if len(refs) == 0 {
		return result
	}
	index := kustoCodeServiceTokenRefIndexAt(refs, bytePosition)
	if index < 0 {
		return result
	}

	token := refs[index].positionedToken()
	var previous *SyntaxToken
	if index > 0 {
		previous = refs[index-1].positionedToken()
	}

	if bytePosition == token.TriviaStart() && (token.TriviaWidth() > 0 || token.Kind() == SyntaxKindEndOfTextToken) && previous != nil {
		token = previous
	}

	if (bytePosition > token.TriviaStart() || (bytePosition == token.TriviaStart() && previous == nil)) &&
		(bytePosition < token.TextStart() || token.Kind() == SyntaxKindEndOfTextToken) {
		if commentStart, commentLength, ok := TryGetCommentSpan(token.Trivia(), bytePosition-token.TriviaStart()); ok && bytePosition < token.TriviaStart()+commentStart+commentLength {
			start, length := byteRangeToUTF16(text, token.TriviaStart()+commentStart, commentLength)
			return NewTextRange(start, length)
		}
		return NewTextRange(position, 0)
	}

	start, length := byteRangeToUTF16(text, token.TextStart(), len(token.RawText()))
	return NewTextRange(start, length)
}

func kustoCodeServiceTokenAt(code *KustoCode, position int) *SyntaxToken {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	refs := kustoServiceTokenRefs(code.Syntax())
	index := kustoCodeServiceTokenRefIndexAt(refs, position)
	if index < 0 {
		return nil
	}
	return refs[index].positionedToken()
}

func kustoCodeServiceTokenRefIndexAt(refs []kustoServiceTokenRef, position int) int {
	if len(refs) == 0 {
		return -1
	}
	if position >= refs[len(refs)-1].end() {
		return len(refs) - 1
	}
	for i, ref := range refs {
		if ref.triviaStart <= position && position < ref.end() {
			return i
		}
	}
	return -1
}

func (s *KustoCodeService) GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) (result RelatedInfo) {
	result = RelatedInfoEmpty
	defer func() {
		if recover() != nil {
			result = RelatedInfoEmpty
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) {
		return result
	}
	return findKustoRelatedElements(code, position, options, cancellationToken)
}

func (s *KustoCodeService) GetClusterReferences(cancellationToken CancellationToken) (result []ClusterReference) {
	result = EmptyReadOnlyList[ClusterReference]()
	defer func() {
		if recover() != nil {
			result = EmptyReadOnlyList[ClusterReference]()
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) || !code.HasSemantics() {
		return result
	}
	result = append(result, kustoDirectiveClusterReferences(code)...)
	newKustoReferenceFinder(code).getClusterReferences(code.Syntax(), nil, &result, cancellationToken)
	return result
}

func (s *KustoCodeService) GetDatabaseReferences(cancellationToken CancellationToken) (result []DatabaseReference) {
	result = EmptyReadOnlyList[DatabaseReference]()
	defer func() {
		if recover() != nil {
			result = EmptyReadOnlyList[DatabaseReference]()
		}
	}()
	if s == nil {
		return result
	}
	code, ok := s.tryGetBoundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) || !code.HasSemantics() {
		return result
	}
	result = append(result, kustoDirectiveDatabaseReferences(code, code.Globals.Cluster)...)
	finder := newKustoReferenceFinder(code)
	finder.getDatabaseReferences(code.Syntax(), nil, code.Globals.Cluster, code.Globals.Database, &result, cancellationToken)
	return result
}

func KustoCodeServiceIncludeTrivia(kind MinimalTextKind) IncludeTrivia {
	switch kind {
	case MinimalTextKindRemoveLeadingWhitespaceAndComments:
		return IncludeTriviaInterior
	case MinimalTextKindMinimizeWhitespaceAndRemoveComments:
		return IncludeTriviaMinimal
	case MinimalTextKindSingleLine:
		return IncludeTriviaSingleLine
	default:
		panic(InvalidOperationError{Message: fmt.Sprintf("Unhandled MinimalTextKind '%d'", kind)})
	}
}

func (s *KustoCodeService) GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string {
	if s == nil {
		return ""
	}
	if code, ok := s.tryGetBoundOrUnboundCode(cancellationToken, true); ok && code != nil && code.Syntax() != nil {
		return code.Syntax().Text(KustoCodeServiceIncludeTrivia(kind))
	}
	return s.Text()
}

func (s *KustoCodeService) GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) (result FormattedText) {
	if s == nil {
		return NewFormattedText("", caretPosition)
	}
	result = NewFormattedText(s.Text(), caretPosition)
	defer func() {
		if recover() != nil {
			result = NewFormattedText(s.Text(), caretPosition)
		}
	}()
	cancellationToken.ThrowIfCancellationRequested()
	code, ok := s.tryGetBoundOrUnboundCode(cancellationToken, true)
	if !ok || code == nil || !s.canBeAnalyzed(code) {
		return result
	}
	byteCaret, caretOK := utf16RangeStartToByte(s.Text(), caretPosition)
	if !caretOK {
		return result
	}
	formatted := FormatKustoCode(code, options, byteCaret)
	return NewFormattedText(formatted.Text, byteOffsetToUTF16Position(formatted.Text, formatted.Position))
}

func byteRangeToUTF16(text string, start int, length int) (int, int) {
	convertedStart := byteOffsetToUTF16Position(text, start)
	convertedEnd := byteOffsetToUTF16Position(text, start+length)
	return convertedStart, convertedEnd - convertedStart
}

func kustoCodeServiceDiagnosticsToUTF16(text string, diagnostics []Diagnostic) []Diagnostic {
	return ListToReadOnly(DiagnosticsToUTF16(text, diagnostics))
}

type KustoCodeServiceFactory struct {
	Globals *GlobalState
}

func NewKustoCodeServiceFactory(globals *GlobalState) *KustoCodeServiceFactory {
	if globals == nil {
		panic(ArgumentError{ParameterName: "globals", Message: "globals cannot be nil"})
	}
	return &KustoCodeServiceFactory{Globals: globals.WithCache()}
}

func (f *KustoCodeServiceFactory) WithGlobals(globals *GlobalState) *KustoCodeServiceFactory {
	if f != nil && f.Globals == globals {
		return f
	}
	return NewKustoCodeServiceFactory(globals)
}

func (f *KustoCodeServiceFactory) TryGetCodeService(text string) (CodeService, bool) {
	if KustoCodeServiceFactoryIsKusto(text) {
		globals := DefaultGlobalState
		if f != nil && f.Globals != nil {
			globals = f.Globals
		}
		return NewKustoCodeService(text, globals), true
	}
	return nil, false
}

func (f *KustoCodeServiceFactory) GetFactory(match func(CodeServiceFactory) bool) CodeServiceFactory {
	if f != nil && (match == nil || match(f)) {
		return f
	}
	return nil
}

func (f *KustoCodeServiceFactory) WithFactory(factory CodeServiceFactory) CodeServiceFactory {
	return codeServiceFactoryWithFactory(f, factory)
}

func (f *KustoCodeServiceFactory) factoryKey() string {
	return "KustoCodeServiceFactory"
}

func KustoCodeServiceFactoryIsKusto(text string) bool {
	return !KustoCodeServiceFactoryIsSql(text)
}

func KustoCodeServiceFactoryIsSql(text string) bool {
	tokens := Lex(text)
	if len(tokens) == 0 {
		return false
	}
	firstToken := tokens[0]
	if !IsWhitespaceOnly(firstToken.Trivia) {
		return false
	}
	if firstToken.Text == "--" {
		return true
	}
	if firstToken.Text == "-" && len(tokens) > 1 && tokens[1].Text == "-" && tokens[1].Trivia == "" {
		return true
	}
	return ordinalEqualFold(firstToken.Text, "select") || ordinalEqualFold(firstToken.Text, "with")
}
