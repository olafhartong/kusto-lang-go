package kustolang

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type panickingCaretSyntaxElement struct {
	*SyntaxNode
}

func (*panickingCaretSyntaxElement) Children() []SyntaxElement {
	panic("caret syntax traversal failure")
}

func TestKustoCodeServiceConcurrentLazyInitialization(t *testing.T) {
	const workers = 32
	service := NewKustoCodeService("print value=1")

	start := make(chan struct{})
	codes := make(chan *KustoCode, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			code, ok := service.tryGetCode(true)
			if !ok {
				t.Errorf("tryGetCode returned no code")
				return
			}
			codes <- code
		}()
	}
	close(start)
	wg.Wait()
	close(codes)

	var first *KustoCode
	for code := range codes {
		if first == nil {
			first = code
		} else if code != first {
			t.Fatal("concurrent initialization published multiple code instances")
		}
	}

	var analyzeCalls atomic.Int32
	var metadataCalls atomic.Int32
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "concurrent analyzer",
		Diagnostics: func() []Diagnostic {
			metadataCalls.Add(1)
			return []Diagnostic{NewDiagnostic("KS901", "concurrent")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			analyzeCalls.Add(1)
			return append(diagnostics, NewDiagnostic("KS901", "concurrent").WithLocation(0, 1))
		},
	})
	service = NewKustoCodeServiceForCode(first).WithAnalyzers([]KustoAnalyzer{analyzer})

	start = make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			switch index % 3 {
			case 0:
				service.GetDiagnostics(true, CancellationToken{})
			case 1:
				service.GetAnalyzerDiagnostics(true, CancellationToken{})
			default:
				service.GetAnalyzers()
			}
		}(i)
	}
	close(start)
	wg.Wait()

	if got := analyzeCalls.Load(); got != 1 {
		t.Fatalf("analyzer calls = %d, want 1", got)
	}
	if got := metadataCalls.Load(); got != 1 {
		t.Fatalf("metadata calls = %d, want 1", got)
	}
	if _, ok := service.TryGetCachedDiagnostics(); !ok {
		t.Fatal("binder diagnostics were not published")
	}
	if diagnostics, ok := service.TryGetCachedAnalyzerDiagnostics(); !ok || len(diagnostics) != 1 {
		t.Fatalf("cached analyzer diagnostics = %#v, %v", diagnostics, ok)
	}
}

func TestKustoCodeServiceSeparatesParseOnlyAndBoundLifecycle(t *testing.T) {
	service := NewKustoCodeService("print value=1")
	if code, ok := service.tryGetBoundCode(CancellationToken{}, false); ok || code != nil {
		t.Fatalf("nonblocking bound lookup = %#v, %t; want no code", code, ok)
	}
	if service.unboundCodeState.Load() != nil || service.boundCodeState.Load() != nil {
		t.Fatal("nonblocking lookup initialized code")
	}

	formatted := service.GetFormattedText(FormattingOptionsDefault, 0, CancellationToken{})
	if formatted.Text != "print value = 1" {
		t.Fatalf("formatted text = %q", formatted.Text)
	}
	unboundState := service.unboundCodeState.Load()
	if unboundState == nil || unboundState.code == nil || unboundState.code.HasSemantics() {
		t.Fatalf("syntax-only state = %#v", unboundState)
	}
	if service.boundCodeState.Load() != nil {
		t.Fatal("syntax-only formatting performed semantic analysis")
	}

	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	boundState := service.boundCodeState.Load()
	if boundState == nil || boundState.code == nil || !boundState.code.HasSemantics() {
		t.Fatalf("bound state = %#v", boundState)
	}
	if len(unboundState.code.Tokens) > 0 && &unboundState.code.Tokens[0] != &boundState.code.Tokens[0] {
		t.Fatal("semantic analysis did not reuse parse-only lexical tokens")
	}
}

func TestKustoCodeServiceParseOnlyLifecycleHandlesMalformedSyntax(t *testing.T) {
	service := NewKustoCodeService("T | where (")
	formatted := service.GetFormattedText(FormattingOptionsDefault.WithInsertMissingTokens(false), 0, CancellationToken{})
	if formatted.Text != "T | where (" {
		t.Fatalf("malformed formatted text = %q", formatted.Text)
	}
	unboundState := service.unboundCodeState.Load()
	if unboundState == nil || unboundState.code == nil || unboundState.code.HasSemantics() || len(unboundState.code.GetSyntaxDiagnostics()) == 0 {
		t.Fatalf("malformed parse-only state = %#v", unboundState)
	}
	if service.boundCodeState.Load() != nil {
		t.Fatal("malformed syntax-only formatting performed semantic analysis")
	}

	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) == 0 {
		t.Fatal("malformed bound lifecycle lost syntax diagnostics")
	}
	if bound := service.boundCodeState.Load(); bound == nil || bound.code == nil || !bound.code.HasSemantics() {
		t.Fatalf("malformed bound state = %#v", bound)
	}
}

func TestKustoCodeServiceCanceledAnalyzerInitializationCanRetry(t *testing.T) {
	var calls atomic.Int32
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "cancelable analyzer",
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			calls.Add(1)
			cancellationToken.ThrowIfCancellationRequested()
			return append(diagnostics, NewDiagnostic("KS902", "completed").WithLocation(0, 1))
		},
	})
	service := NewKustoCodeServiceForCode(ParseAndAnalyze("print value=1")).WithAnalyzers([]KustoAnalyzer{analyzer})
	canceled := NewCancellationTokenFunc(func() bool { return true })

	assertOperationCanceled(t, func() { service.GetAnalyzerDiagnostics(true, canceled) })
	if diagnostics, ok := service.TryGetCachedAnalyzerDiagnostics(); ok || diagnostics != nil {
		t.Fatalf("canceled initialization was cached: %#v, %v", diagnostics, ok)
	}
	if diagnostics := service.GetAnalyzerDiagnostics(true, CancellationToken{}); len(diagnostics) != 1 || diagnostics[0].Code != "KS902" {
		t.Fatalf("retried analyzer diagnostics = %#v", diagnostics)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("analyzer calls = %d, want only the uncanceled retry", got)
	}
}

func TestKustoCodeServiceAnalyzerCancellationPanicBecomesFailureDiagnostic(t *testing.T) {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "canceling analyzer",
		Analyze: func(*KustoCode, []Diagnostic, CancellationToken) []Diagnostic {
			panic(OperationCanceledError{})
		},
	})
	service := NewKustoCodeServiceForCode(ParseAndAnalyze("print value=1")).WithAnalyzers([]KustoAnalyzer{analyzer})
	diagnostics := service.GetAnalyzerDiagnostics(true, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS203" ||
		!strings.Contains(diagnostics[0].Message, "canceling analyzer") ||
		!strings.Contains(diagnostics[0].Message, "operation canceled") {
		t.Fatalf("analyzer cancellation diagnostics = %#v", diagnostics)
	}
}

func TestKustoCodeServiceContainsClassifierAndCodeActionCancellation(t *testing.T) {
	canceled := NewCancellationTokenFunc(func() bool { return true })
	classifiedService := NewKustoCodeServiceForCode(ParseAndAnalyze("print value=1"))
	if info := classifiedService.GetClassifications(0, len(classifiedService.Text()), true, true, canceled); len(info.Classifications) != 0 {
		t.Fatalf("canceled classifications = %#v, want empty", info.Classifications)
	}

	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "cancelable fixer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS903", "fix")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			return append(diagnostics, NewDiagnostic("KS903", "fix").WithLocation(0, 1))
		},
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			cancellationToken.ThrowIfCancellationRequested()
			return append(actions, NewCodeAction("cancelable", "Cancelable", "", nil))
		},
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			cancellationToken.ThrowIfCancellationRequested()
			return NewFixEdits(caretPosition, nil)
		},
	})
	service := NewKustoCodeService("abc").WithAnalyzers([]KustoAnalyzer{analyzer})
	service.GetAnalyzerDiagnostics(true, CancellationToken{})

	if info := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, canceled); len(info.Actions) != 0 {
		t.Fatalf("canceled code actions = %#v, want none", info.Actions)
	}
	actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("uncanceled code actions = %#v, want one", actions.Actions)
	}
	action, ok := actions.Actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("code action = %#v, want ApplyAction", actions.Actions[0])
	}
	if result := service.ApplyCodeAction(action, 0, CodeActionOptionsDefault, canceled); result.FailureReason != "Action failed" || len(result.Actions) != 0 {
		t.Fatalf("canceled apply result = %#v", result)
	}
}

func TestKustoCodeServiceCaretInspectionContainsSyntaxTraversalPanics(t *testing.T) {
	code := ParseAndAnalyze("print value=1")
	code.Tree.Root.children = []SyntaxElement{
		&panickingCaretSyntaxElement{SyntaxNode: NewSyntaxNode(SyntaxKindCustomNode, nil)},
	}
	service := NewKustoCodeServiceForCode(code)

	quickInfo := service.GetQuickInfo(0, QuickInfoOptionsDefault, CancellationToken{})
	if quickInfo.Items == nil || len(quickInfo.Items) != 0 || quickInfo.Text() != "" {
		t.Fatalf("quick info after traversal panic = items:%#v text:%q, want canonical empty", quickInfo.Items, quickInfo.Text())
	}

	element := service.GetElement(0, CancellationToken{})
	if element.Start != 0 || element.Length != 0 {
		t.Fatalf("element after traversal panic = %#v, want empty range", element)
	}
}
