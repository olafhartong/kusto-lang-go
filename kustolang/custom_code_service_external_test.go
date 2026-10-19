package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestCustomCodeServiceDelegatesEveryCallback(t *testing.T) {
	calls := map[string]int{}
	called := func(name string) { calls[name]++ }
	diagnostic := kustolang.NewDiagnostic("CUSTOM001", "custom diagnostic").WithLocation(1, 2)
	action := kustolang.NewCodeAction("custom", "Custom action", "", nil)

	service := kustolang.NewCustomCodeService("custom text", kustolang.CustomCodeServiceCallbacks{
		Kind: func() string {
			called("Kind")
			return "Custom"
		},
		IsFeatureSupported: func(feature string, position int) bool {
			called("IsFeatureSupported")
			return feature == "custom-feature" && position == 3
		},
		TryGetCachedDiagnostics: func() ([]kustolang.Diagnostic, bool) {
			called("TryGetCachedDiagnostics")
			return []kustolang.Diagnostic{diagnostic}, true
		},
		GetDiagnostics: func(waitForAnalysis bool, token kustolang.CancellationToken) []kustolang.Diagnostic {
			called("GetDiagnostics")
			return []kustolang.Diagnostic{diagnostic}
		},
		TryGetCachedAnalyzerDiagnostics: func() ([]kustolang.Diagnostic, bool) {
			called("TryGetCachedAnalyzerDiagnostics")
			return []kustolang.Diagnostic{diagnostic}, true
		},
		GetAnalyzerDiagnostics: func(waitForAnalysis bool, token kustolang.CancellationToken) []kustolang.Diagnostic {
			called("GetAnalyzerDiagnostics")
			return []kustolang.Diagnostic{diagnostic}
		},
		GetCombinedDiagnostics: func(waitForAnalysis bool, filter *kustolang.DisabledDiagnostics, token kustolang.CancellationToken) []kustolang.Diagnostic {
			called("GetCombinedDiagnostics")
			return []kustolang.Diagnostic{diagnostic}
		},
		GetAnalyzers: func() []kustolang.AnalyzerInfo {
			called("GetAnalyzers")
			return []kustolang.AnalyzerInfo{kustolang.NewAnalyzerInfo("custom", nil)}
		},
		GetCodeActions: func(position, selectionStart, selectionLength int, options kustolang.CodeActionOptions, waitForAnalysis bool, actorName string, token kustolang.CancellationToken) kustolang.CodeActionInfo {
			called("GetCodeActions")
			return kustolang.NewCodeActionInfo([]kustolang.CodeAction{action})
		},
		ApplyCodeAction: func(applied kustolang.ApplyAction, caretPosition int, options kustolang.CodeActionOptions, token kustolang.CancellationToken) kustolang.CodeActionResult {
			called("ApplyCodeAction")
			return kustolang.NewCodeActionResult(kustolang.NewMoveCaretAction(caretPosition + 1))
		},
		GetClassifications: func(start, length int, clipToRange, waitForAnalysis bool, token kustolang.CancellationToken) kustolang.ClassificationInfo {
			called("GetClassifications")
			return kustolang.NewClassificationInfo([]kustolang.ClassifiedRange{kustolang.NewClassifiedRange(kustolang.ClassificationKindKeyword, start, length)})
		},
		GetOutlines: func(options kustolang.OutliningOptions, token kustolang.CancellationToken) kustolang.OutlineInfo {
			called("GetOutlines")
			return kustolang.NewOutlineInfo([]kustolang.OutlineRange{kustolang.NewOutlineRange(0, 4, "custom")})
		},
		ShouldAutoComplete: func(position int, key rune, token kustolang.CancellationToken) bool {
			called("ShouldAutoComplete")
			return position == 2 && key == '.'
		},
		GetCompletionItems: func(position int, options kustolang.CompletionOptions, token kustolang.CancellationToken) kustolang.CompletionInfo {
			called("GetCompletionItems")
			return kustolang.NewCompletionInfo(nil, position, 0)
		},
		GetQuickInfo: func(position int, options kustolang.QuickInfoOptions, token kustolang.CancellationToken) kustolang.QuickInfo {
			called("GetQuickInfo")
			return kustolang.NewQuickInfoText("custom quick info")
		},
		GetElement: func(position int, token kustolang.CancellationToken) kustolang.TextRange {
			called("GetElement")
			return kustolang.NewTextRange(position, 1)
		},
		GetRelatedElements: func(position int, options kustolang.FindRelatedOptions, token kustolang.CancellationToken) kustolang.RelatedInfo {
			called("GetRelatedElements")
			return kustolang.NewRelatedInfo([]kustolang.RelatedElement{kustolang.NewRelatedElement(position, 1, kustolang.RelatedElementKindReference)}, 0)
		},
		GetClusterReferences: func(token kustolang.CancellationToken) []kustolang.ClusterReference {
			called("GetClusterReferences")
			return []kustolang.ClusterReference{kustolang.NewClusterReference("custom-cluster", 1, 2)}
		},
		GetDatabaseReferences: func(token kustolang.CancellationToken) []kustolang.DatabaseReference {
			called("GetDatabaseReferences")
			return []kustolang.DatabaseReference{kustolang.NewDatabaseReference("custom-db", "custom-cluster", 1, 2)}
		},
		GetMinimalText: func(kind kustolang.MinimalTextKind, token kustolang.CancellationToken) string {
			called("GetMinimalText")
			return "minimal"
		},
		GetFormattedText: func(options kustolang.FormattingOptions, caretPosition int, token kustolang.CancellationToken) kustolang.FormattedText {
			called("GetFormattedText")
			return kustolang.NewFormattedText("formatted", caretPosition)
		},
		GetClientParameters: func() []kustolang.ClientParameter {
			called("GetClientParameters")
			return []kustolang.ClientParameter{kustolang.NewClientParameterReference("custom", 1, 2)}
		},
	})

	token := kustolang.CancellationToken{}
	if service.Text() != "custom text" || service.Kind() != "Custom" {
		t.Fatalf("text/kind = %q/%q", service.Text(), service.Kind())
	}
	if !service.IsFeatureSupported("custom-feature", 3) {
		t.Fatal("custom feature was not supported")
	}
	service.TryGetCachedDiagnostics()
	service.GetDiagnostics(true, token)
	service.TryGetCachedAnalyzerDiagnostics()
	service.GetAnalyzerDiagnostics(true, token)
	service.GetCombinedDiagnostics(true, nil, token)
	service.GetAnalyzers()
	service.GetCodeActions(3, 1, 2, kustolang.CodeActionOptionsDefault, true, "custom", token)
	service.ApplyCodeAction(action, 3, kustolang.CodeActionOptionsDefault, token)
	service.GetClassifications(1, 2, true, true, token)
	service.GetOutlines(kustolang.OutliningOptionsDefault, token)
	service.ShouldAutoComplete(2, '.', token)
	service.GetCompletionItems(2, kustolang.CompletionOptionsDefault, token)
	service.GetQuickInfo(2, kustolang.QuickInfoOptionsDefault, token)
	service.GetElement(2, token)
	service.GetRelatedElements(2, kustolang.FindRelatedOptionsNone, token)
	service.GetClusterReferences(token)
	service.GetDatabaseReferences(token)
	service.GetMinimalText(kustolang.MinimalTextKindSingleLine, token)
	service.GetFormattedText(kustolang.FormattingOptionsDefault, 2, token)
	service.GetClientParameters()

	for _, name := range []string{
		"Kind", "IsFeatureSupported", "TryGetCachedDiagnostics", "GetDiagnostics",
		"TryGetCachedAnalyzerDiagnostics", "GetAnalyzerDiagnostics", "GetCombinedDiagnostics", "GetAnalyzers",
		"GetCodeActions", "ApplyCodeAction", "GetClassifications", "GetOutlines",
		"ShouldAutoComplete", "GetCompletionItems", "GetQuickInfo", "GetElement",
		"GetRelatedElements", "GetClusterReferences", "GetDatabaseReferences",
		"GetMinimalText", "GetFormattedText", "GetClientParameters",
	} {
		if calls[name] != 1 {
			t.Errorf("%s callback calls = %d, want 1", name, calls[name])
		}
	}
}

func TestCustomCodeServiceCombinedDiagnosticsDoesNotMutateCallbackSlices(t *testing.T) {
	parserDiagnostic := kustolang.NewDiagnostic("CUSTOM001", "parser")
	analyzerDiagnostic := kustolang.NewDiagnostic("CUSTOM002", "analyzer")
	sentinel := kustolang.NewDiagnostic("SENTINEL", "must remain untouched")
	storage := make([]kustolang.Diagnostic, 2)
	storage[0] = parserDiagnostic
	storage[1] = sentinel
	parserDiagnostics := storage[:1]

	service := kustolang.NewCustomCodeService("print 1", kustolang.CustomCodeServiceCallbacks{
		GetDiagnostics: func(bool, kustolang.CancellationToken) []kustolang.Diagnostic {
			return parserDiagnostics
		},
		GetAnalyzerDiagnostics: func(bool, kustolang.CancellationToken) []kustolang.Diagnostic {
			return []kustolang.Diagnostic{analyzerDiagnostic}
		},
	})
	got := service.GetCombinedDiagnostics(true, nil, kustolang.CancellationToken{})
	if len(got) != 2 || got[0].Code != "CUSTOM001" || got[1].Code != "CUSTOM002" {
		t.Fatalf("combined diagnostics = %#v", got)
	}
	if storage[1] != sentinel {
		t.Fatalf("combined diagnostics overwrote callback-owned capacity: %#v", storage)
	}

	got[0] = sentinel
	if parserDiagnostics[0] != parserDiagnostic {
		t.Fatal("mutating the combined result changed callback-owned diagnostics")
	}
}

func TestCustomCodeServiceCommonClassificationUsesCustomClientParameters(t *testing.T) {
	service := kustolang.NewCustomCodeService("a{p}b", kustolang.CustomCodeServiceCallbacks{
		GetClientParameters: func() []kustolang.ClientParameter {
			return []kustolang.ClientParameter{kustolang.NewClientParameterReference("p", 1, 3)}
		},
	})
	info := service.GetClassifications(0, 5, true, false, kustolang.CancellationToken{})
	found := false
	for _, classification := range info.Classifications {
		if classification.Kind == kustolang.ClassificationKindClientParameter && classification.Start == 1 && classification.Length == 3 {
			found = true
		}
	}
	if !found {
		t.Fatalf("custom client parameter was absent from common classifications: %#v", info.Classifications)
	}
}

func TestCustomCodeServiceFactoryAndCodeScriptOffsets(t *testing.T) {
	factory := kustolang.NewCodeServiceFactory("custom", func(text string) (*kustolang.CodeService, bool) {
		return kustolang.NewCustomCodeService(text, kustolang.CustomCodeServiceCallbacks{
			Kind: func() string { return "Custom" },
			GetDiagnostics: func(bool, kustolang.CancellationToken) []kustolang.Diagnostic {
				return []kustolang.Diagnostic{kustolang.NewDiagnostic("CUSTOM001", "custom").WithLocation(1, 1)}
			},
		}), true
	})

	script := kustolang.NewCodeScriptWithFactory("first\n\nsecond", factory)
	blocks := script.Blocks()
	if len(blocks) != 2 || blocks[1].Start() != 7 || blocks[1].Kind() != "Custom" {
		t.Fatalf("custom blocks = %#v", blocks)
	}
	if got := blocks[1].Service().GetDiagnostics(true, kustolang.CancellationToken{}); len(got) != 1 || got[0].Start != 8 {
		t.Fatalf("offset custom diagnostics = %#v", got)
	}
	if found := script.Factory().GetFactory(func(candidate *kustolang.CodeServiceFactory) bool {
		return candidate.Key() == "custom"
	}); found == nil {
		t.Fatal("CodeScript did not preserve the custom factory")
	}
}

func TestCustomCodeServiceUsesCommonDefaults(t *testing.T) {
	service := kustolang.NewCustomCodeService("print   1", kustolang.CustomCodeServiceCallbacks{})
	if service.Text() != "print   1" || service.Kind() != "" {
		t.Fatalf("default text/kind = %q/%q", service.Text(), service.Kind())
	}
	if !service.IsFeatureSupported("any-feature", -1) {
		t.Fatal("CommonCodeService feature default was not used")
	}
	if got := service.GetMinimalText(kustolang.MinimalTextKindSingleLine, kustolang.CancellationToken{}); got != "print 1" {
		t.Fatalf("default minimal text = %q", got)
	}
}
