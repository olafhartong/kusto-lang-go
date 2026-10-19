package kustolang

import (
	"strconv"
	"strings"
	"testing"
)

func TestKustoCodeServiceFactoryDetectsKustoAndRejectsSql(t *testing.T) {
	factory := NewKustoCodeServiceFactory(DefaultGlobalState)
	if factory.Globals == nil || factory.Globals.Cache == nil {
		t.Fatalf("factory globals/cache = %#v", factory.Globals)
	}
	if factory.WithGlobals(factory.Globals) != factory {
		t.Fatal("WithGlobals should preserve identity for the same cached globals")
	}

	for _, text := range []string{"select * from T", "  with cte as (select 1)", "-- sql comment"} {
		if service, ok := factory.TryGetCodeService(text); ok || service != nil {
			t.Fatalf("SQL-looking text %q produced service %#v", text, service)
		}
		if KustoCodeServiceFactoryIsKusto(text) {
			t.Fatalf("IsKusto(%q) = true, want false", text)
		}
	}

	for _, text := range []string{"// kusto comment\nT | count", "T | summarize count()", ".show tables"} {
		service, ok := factory.TryGetCodeService(text)
		if !ok || service == nil || service.Text() != text {
			t.Fatalf("Kusto text %q service = %#v, %v", text, service, ok)
		}
	}
}

func TestKustoCodeServiceDiagnosticsAnalyzerCacheAndFeatures(t *testing.T) {
	service := NewKustoCodeService("T | count", NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T")))).WithAnalyzers([]KustoAnalyzer{
		NewKustoAnalyzer(KustoAnalyzerCallbacks{
			Name: "test analyzer",
			Diagnostics: func() []Diagnostic {
				return []Diagnostic{NewDiagnostic("KS998", "metadata")}
			},
			Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
				return append(diagnostics, NewDiagnostic("KS999", "from analyzer"))
			},
		}),
		NewKustoAnalyzer(KustoAnalyzerCallbacks{
			Name: "panic analyzer",
			Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
				panic("boom")
			},
		}),
	})

	if service.Kind() != CodeKindsQuery || !service.IsFeatureSupported(CodeServiceFeatureCompletion, 0) {
		t.Fatalf("service kind/features = %q", service.Kind())
	}
	if diagnostics, ok := service.TryGetCachedDiagnostics(); ok || diagnostics != nil {
		t.Fatalf("cached diagnostics before parse = %#v, %v", diagnostics, ok)
	}
	if diagnostics := service.GetDiagnostics(true, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("syntax diagnostics = %#v, want empty", diagnostics)
	}
	if diagnostics, ok := service.TryGetCachedDiagnostics(); !ok || len(diagnostics) != 0 {
		t.Fatalf("cached diagnostics after parse = %#v, %v", diagnostics, ok)
	}

	analyzerDiagnostics := service.GetAnalyzerDiagnostics(true, CancellationToken{})
	if len(analyzerDiagnostics) != 2 || analyzerDiagnostics[0].Code != "KS999" || analyzerDiagnostics[1].Code != "KS203" {
		t.Fatalf("analyzer diagnostics = %#v", analyzerDiagnostics)
	}
	if cached, ok := service.TryGetCachedAnalyzerDiagnostics(); !ok || len(cached) != len(analyzerDiagnostics) {
		t.Fatalf("cached analyzer diagnostics = %#v, %v", cached, ok)
	}

	infos := service.GetAnalyzers()
	if len(infos) != 2 || infos[0].Name != "test analyzer" || len(infos[0].Diagnostics) != 1 || infos[0].Diagnostics[0].Code != "KS998" {
		t.Fatalf("analyzer infos = %#v", infos)
	}
}

func TestKustoCodeServiceReportsUpstreamFeatureContract(t *testing.T) {
	implemented := []string{
		CodeServiceFeatureDiagnostics,
		CodeServiceFeatureClassification,
		CodeServiceFeatureCompletion,
		CodeServiceFeatureOutlining,
		CodeServiceFeatureQuickInfo,
		CodeServiceFeatureClusterReferences,
		CodeServiceFeatureDatabaseReferences,
		CodeServiceFeatureMinimalText,
		CodeServiceFeatureFormatting,
		CodeServiceFeatureClientParameters,
	}
	for _, text := range []string{"print x=1", ".show version"} {
		service := NewKustoCodeService(text)
		for _, feature := range implemented {
			if !service.IsFeatureSupported(feature, 0) {
				t.Errorf("%s service does not report implemented feature %s", service.Kind(), feature)
			}
		}
		if !service.IsFeatureSupported("FutureFeature", 0) {
			t.Errorf("%s service does not preserve upstream's open feature contract", service.Kind())
		}
	}

	directive := NewKustoCodeService("#set foo")
	directive.kind = CodeKindsDirective
	if !directive.IsFeatureSupported(CodeServiceFeatureClassification, 0) || directive.IsFeatureSupported(CodeServiceFeatureDiagnostics, 0) {
		t.Fatalf("directive feature reporting is inconsistent: kind=%s", directive.Kind())
	}
}

func TestKustoCodeServiceCodeActionsRouteAnalyzerFixers(t *testing.T) {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "ReplaceFirstAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS777", "replace first")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			return append(diagnostics, NewDiagnostic("KS777", "replace first").WithLocation(0, 1))
		},
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			if diagnostic.Code != "KS777" {
				return actions
			}
			return append(actions, NewCodeAction("replace-first", "Replace first", "Replace first character", []string{strconv.Itoa(diagnostic.Start)}))
		},
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			data := action.ApplyData()
			if len(data) != 1 {
				return NewFixEdits(caretPosition, nil)
			}
			start, err := strconv.Atoi(data[0])
			if err != nil {
				return NewFixEdits(caretPosition, nil)
			}
			return NewFixEditsList(start, NewTextReplacement(start, 1, "X"))
		},
	})
	service := NewKustoCodeService("abc").WithAnalyzers([]KustoAnalyzer{analyzer})

	if actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, true, "NoSuchActor", CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("unknown actor actions = %#v, want none", actions)
	}

	filter := ParseDisabledDiagnostics("KS777")
	filteredOptions := CodeActionOptionsDefault.WithDiagnosticFilter(&filter)
	if actions := service.GetCodeActions(0, 0, 0, filteredOptions, true, "", CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("filtered actions = %#v, want none", actions)
	}

	actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("actions = %#v, want one", actions)
	}
	action, ok := actions.Actions[0].(ApplyAction)
	if !ok {
		t.Fatalf("action = %#v, want apply action", actions.Actions[0])
	}
	if action.Base().Kind != "replace-first" || action.Base().Title != "Replace first" {
		t.Fatalf("action metadata = %#v", action.Base())
	}

	result := service.ApplyCodeAction(action, 0, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("apply result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "Xbc" {
		t.Fatalf("change = %#v", result.Actions)
	}
}

func TestKustoCodeServiceCodeActionsWaitFalseUsesCachedAnalyzerDiagnostics(t *testing.T) {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "CachedAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS776", "cached")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			return append(diagnostics, NewDiagnostic("KS776", "cached").WithLocation(0, 1))
		},
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			return append(actions, NewCodeAction("cached", "Cached fix", "", nil))
		},
	})
	service := NewKustoCodeService("abc").WithAnalyzers([]KustoAnalyzer{analyzer})
	if actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, false, "", CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("uncached wait=false actions = %#v, want none", actions)
	}
	_ = service.GetAnalyzerDiagnostics(true, CancellationToken{})
	if actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, false, "", CancellationToken{}); len(actions.Actions) != 1 {
		t.Fatalf("cached wait=false actions = %#v, want one", actions)
	}
}

func TestKustoCodeServiceDiagnosticDedupMatchesReferenceIdentity(t *testing.T) {
	existing := []Diagnostic{
		NewDiagnosticDetailed("KS700", DiagnosticCategoryGeneral, DiagnosticSeverityError, "first", "same").WithLocation(2, 3),
		NewDiagnostic("KS701", "relative"),
	}
	additions := []Diagnostic{
		NewDiagnosticDetailed("KS700", DiagnosticCategoryPerformance, DiagnosticSeveritySuggestion, "different metadata", "same").WithLocation(2, 3),
		NewDiagnostic("KS701", "relative"),
		NewDiagnostic("KS700", "different message").WithLocation(2, 3),
	}
	got := kustoCodeServiceDistinctNewDiagnostics(existing, additions)
	if len(got) != 2 || got[0].Code != "KS701" || got[1].Message != "different message" {
		t.Fatalf("distinct diagnostics = %#v", got)
	}
}

func TestKustoCodeServiceCodeActionsPromoteSelectionFixAll(t *testing.T) {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "ReplaceRepeatedAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS775", "replace repeated")}
		},
		Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
			return append(diagnostics,
				NewDiagnostic("KS775", "replace repeated").WithLocation(0, 1),
				NewDiagnostic("KS775", "replace repeated").WithLocation(2, 1),
			)
		},
		FixAction: func(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
			return append(actions, NewCodeAction("replace-repeated", "Replace repeated", "", []string{strconv.Itoa(diagnostic.Start)}))
		},
		FixEdits: func(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
			data := action.ApplyData()
			if len(data) != 1 {
				return NewFixEdits(caretPosition, nil)
			}
			start, err := strconv.Atoi(data[0])
			if err != nil {
				return NewFixEdits(caretPosition, nil)
			}
			return NewFixEditsList(start, NewTextReplacement(start, 1, "X"))
		},
	})
	service := NewKustoCodeService("aba").WithAnalyzers([]KustoAnalyzer{analyzer})
	actions := service.GetCodeActions(0, 0, len("aba"), CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("actions = %#v, want one menu", actions)
	}
	menu, ok := actions.Actions[0].(*MenuAction)
	if !ok || len(menu.Actions) != 2 {
		t.Fatalf("action = %#v, want apply/fix-all menu", actions.Actions[0])
	}
	if menu.Base().Title != "Replace repeated" || menu.Actions[0].Base().Title != "Apply" || menu.Actions[1].Base().Title != "Fix All" {
		t.Fatalf("menu metadata = %#v actions=%#v", menu.Base(), menu.Actions)
	}
	fixAll, ok := menu.Actions[1].(ApplyAction)
	if !ok {
		t.Fatalf("fix all action = %#v, want apply action", menu.Actions[1])
	}
	result := service.ApplyCodeAction(fixAll, 0, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("fix all result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "XbX" {
		t.Fatalf("fix all change = %#v", result.Actions)
	}
}

func TestKustoCodeServiceHonorsGlobalParseAndAnalysisLimits(t *testing.T) {
	tinyParse := GlobalStateWithProperty(DefaultGlobalState, Properties.MaxParseTextSize, len("T")-1)
	parseLimited := NewKustoCodeService("T", tinyParse)
	diagnostics := parseLimited.GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS244" {
		t.Fatalf("parse-limit diagnostics = %#v", diagnostics)
	}
	if classifications := parseLimited.GetClassifications(0, 1, true, true, CancellationToken{}); len(classifications.Classifications) != 0 {
		t.Fatalf("parse-limited classifications = %#v, want empty", classifications)
	}

	calls := 0
	shallowAnalysis := GlobalStateWithProperty(DefaultGlobalState, Properties.MaxAnalysisDepth, 1)
	analysisLimited := NewKustoCodeService("T | count", shallowAnalysis).WithAnalyzers([]KustoAnalyzer{
		NewKustoAnalyzer(KustoAnalyzerCallbacks{
			Name: "counting analyzer",
			Analyze: func(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
				calls++
				return append(diagnostics, NewDiagnostic("KS997", "ran"))
			},
		}),
	})
	analyzerDiagnostics := analysisLimited.GetAnalyzerDiagnostics(true, CancellationToken{})
	if len(analyzerDiagnostics) != 0 || calls != 0 {
		t.Fatalf("analysis-limited diagnostics = %#v calls=%d", analyzerDiagnostics, calls)
	}
}

func TestKustoCodeServiceClassifierOutlinerParametersAndElements(t *testing.T) {
	text := "// ignored\nT | where Value == long(1)\n| project {tenant}"
	service := NewKustoCodeService(text)

	parameters := service.GetClientParameters()
	if len(parameters) != 1 || parameters[0].Name != "tenant" || parameters[0].Start() != strings.Index(text, "{tenant}") {
		t.Fatalf("client parameters = %#v", parameters)
	}

	classifications := service.GetClassifications(0, len(text), true, true, CancellationToken{}).Classifications
	requireKustoServiceClassification(t, text, classifications, "// ignored", ClassificationKindComment)
	requireKustoServiceClassification(t, text, classifications, "where", ClassificationKindQueryOperator)
	requireKustoServiceClassification(t, text, classifications, "long", ClassificationKindKeyword)
	requireKustoServiceClassification(t, text, classifications, "(1)", ClassificationKindLiteral)
	requireKustoServiceClassification(t, text, classifications, "{tenant}", ClassificationKindClientParameter)

	outlines := service.GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(outlines.Ranges) == 0 {
		t.Fatalf("outlines = %#v, want at least one", outlines)
	}

	comment := service.GetElement(strings.Index(text, "ignored"), CancellationToken{})
	if comment.Start != 0 || comment.Length != len("// ignored\n") {
		t.Fatalf("comment element = %#v", comment)
	}
	tokenStart := strings.Index(text, "project")
	token := service.GetElement(tokenStart+1, CancellationToken{})
	if token.Start != tokenStart || token.Length != len("project") {
		t.Fatalf("token element = %#v", token)
	}

	if got := service.GetMinimalText(MinimalTextKindSingleLine, CancellationToken{}); strings.Contains(got, "\n") || strings.Contains(got, "//") {
		t.Fatalf("single-line minimal text = %q", got)
	}
}

func requireKustoServiceClassification(t *testing.T, text string, ranges []ClassifiedRange, needle string, kind ClassificationKind) {
	t.Helper()
	start := strings.Index(text, needle)
	if start < 0 {
		t.Fatalf("needle %q not found", needle)
	}
	for _, r := range ranges {
		if r.Start == start && r.Length == len(needle) && r.Kind == kind {
			return
		}
	}
	t.Fatalf("classification %q kind=%v not found in %#v", needle, kind, ranges)
}
