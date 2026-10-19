package kustolang

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestKustoCodeServiceEditorRangesUseUTF16(t *testing.T) {
	text := "// 😀\nEvents | project Account, P={tenant}"
	account := NewColumnSymbolFull("Account", ScalarTypes.String.Type(), "User account", nil, nil, nil)
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("Events", account)))
	service := NewKustoCodeService(text, globals)

	accountByteStart := strings.Index(text, "Account")
	accountStart := utf16TextLength(text[:accountByteStart])
	quickInfo := service.GetQuickInfo(accountStart+1, QuickInfoOptionsDefault, CancellationToken{})
	if len(quickInfo.Items) == 0 || !strings.Contains(quickInfo.Items[0].Text(), "Account: string") {
		t.Fatalf("UTF-16 quick info = %#v", quickInfo)
	}

	element := service.GetElement(accountStart+1, CancellationToken{})
	if element.Start != accountStart || element.Length != utf16TextLength("Account") {
		t.Fatalf("UTF-16 token element = %#v, want %d+7", element, accountStart)
	}
	comment := service.GetElement(1, CancellationToken{})
	if comment.Start != 0 || comment.Length != utf16TextLength("// 😀\n") {
		t.Fatalf("UTF-16 comment element = %#v", comment)
	}

	parameterByteStart := strings.Index(text, "{tenant}")
	parameterStart := utf16TextLength(text[:parameterByteStart])
	parameters := service.GetClientParameters()
	if len(parameters) != 1 || parameters[0].Start() != parameterStart || parameters[0].Length() != len("{tenant}") {
		t.Fatalf("UTF-16 client parameters = %#v, want %d+8", parameters, parameterStart)
	}

	classifications := service.GetClassifications(0, utf16TextLength(text), true, true, CancellationToken{}).Classifications
	requireUTF16Classification(t, classifications, ClassificationKindColumn, accountStart, utf16TextLength("Account"))
	requireUTF16Classification(t, classifications, ClassificationKindClientParameter, parameterStart, utf16TextLength("{tenant}"))
}

func TestKustoCodeServiceOutlinesAndFormattingUseUTF16(t *testing.T) {
	text := "// 😀\nEvents\n| project Account\n"
	service := NewKustoCodeService(text)
	outlines := service.GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(outlines.Ranges) == 0 {
		t.Fatalf("UTF-16 outlines = %#v", outlines)
	}
	wantLength := utf16TextLength(text[:TrimEnd(text)])
	if outlines.Ranges[0].Start != 0 || outlines.Ranges[0].Length != wantLength {
		t.Fatalf("UTF-16 root outline = %#v, want 0+%d", outlines.Ranges[0], wantLength)
	}

	formatText := "// 😀\nEvents|project Account"
	formatted := NewKustoCodeService(formatText).GetFormattedText(
		FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine),
		utf16TextLength(formatText),
		CancellationToken{},
	)
	if formatted.Position != utf16TextLength(formatted.Text) {
		t.Fatalf("UTF-16 formatted caret = %d, want %d for %q", formatted.Position, utf16TextLength(formatted.Text), formatted.Text)
	}
}

func TestKustoCodeServiceMalformedEditorInputUsesUTF16(t *testing.T) {
	text := "// 😀\nEvents | where ("
	service := NewKustoCodeService(text)
	position := utf16TextLength(text)

	classifications := service.GetClassifications(0, position, true, true, CancellationToken{}).Classifications
	if len(classifications) == 0 {
		t.Fatal("malformed UTF-16 classifications are empty")
	}
	wantStart := utf16TextLength(text[:strings.LastIndex(text, "(")])
	element := service.GetElement(wantStart, CancellationToken{})
	if element.Start != wantStart || element.Length != 1 {
		t.Fatalf("malformed UTF-16 element = %#v, want %d+1", element, wantStart)
	}
	formatted := service.GetFormattedText(FormattingOptionsDefault, position, CancellationToken{})
	if formatted.Position < 0 || formatted.Position > utf16TextLength(formatted.Text) {
		t.Fatalf("malformed UTF-16 formatted result = %#v", formatted)
	}
	_ = service.GetQuickInfo(position, QuickInfoOptionsDefault, CancellationToken{})
}

func TestKustoCodeServiceDiagnosticsUseUTF16(t *testing.T) {
	text := "// 😀\nprint x=Missing"
	missingByteStart := strings.Index(text, "Missing")
	missingStart := utf16TextLength(text[:missingByteStart])
	service := NewKustoCodeService(text)

	diagnostics := service.GetDiagnostics(true, CancellationToken{})
	requireUTF16Diagnostic(t, diagnostics, "KS142", missingStart, utf16TextLength("Missing"))
	cached, ok := service.TryGetCachedDiagnostics()
	if !ok {
		t.Fatal("UTF-16 diagnostics were not cached")
	}
	requireUTF16Diagnostic(t, cached, "KS142", missingStart, utf16TextLength("Missing"))
	byteDiagnostics := service.getDiagnosticsBytes(true, CancellationToken{})
	if len(byteDiagnostics) != 1 || byteDiagnostics[0].Start != missingByteStart {
		t.Fatalf("internal diagnostics lost byte indexing: %#v", byteDiagnostics)
	}

	analyzerText := "// 😀\nprint Value=1"
	valueByteStart := strings.Index(analyzerText, "Value")
	valueStart := utf16TextLength(analyzerText[:valueByteStart])
	analyzerService := NewKustoCodeService(analyzerText).WithAnalyzers([]KustoAnalyzer{
		NewKustoAnalyzer(KustoAnalyzerCallbacks{
			Name: "UTF-16 boundary analyzer",
			Analyze: func(_ *KustoCode, diagnostics []Diagnostic, _ CancellationToken) []Diagnostic {
				return append(diagnostics, NewDiagnostic("KS998", "test").WithByteLocation(valueByteStart, len("Value")))
			},
		}),
	})
	analyzerDiagnostics := analyzerService.GetAnalyzerDiagnostics(true, CancellationToken{})
	requireUTF16Diagnostic(t, analyzerDiagnostics, "KS998", valueStart, utf16TextLength("Value"))
	cachedAnalyzer, ok := analyzerService.TryGetCachedAnalyzerDiagnostics()
	if !ok {
		t.Fatal("UTF-16 analyzer diagnostics were not cached")
	}
	requireUTF16Diagnostic(t, cachedAnalyzer, "KS998", valueStart, utf16TextLength("Value"))
}

func TestKustoCodeServiceDiagnosticFixActionsUseUTF16(t *testing.T) {
	text := "// 😀\nabc"
	byteStart := strings.Index(text, "abc")
	utf16Start := utf16TextLength(text[:byteStart])
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "UTF-16 fixer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS997", "replace")}
		},
		Analyze: func(_ *KustoCode, diagnostics []Diagnostic, _ CancellationToken) []Diagnostic {
			return append(diagnostics, NewDiagnostic("KS997", "replace").WithByteLocation(byteStart, 3))
		},
		FixAction: func(_ *KustoCode, diagnostic Diagnostic, _ CodeActionOptions, actions []CodeAction, _ CancellationToken) []CodeAction {
			return append(actions, NewCodeAction("replace", "Replace", "", []string{strconv.Itoa(diagnostic.Start)}))
		},
		FixEdits: func(_ *KustoCode, action ApplyAction, _ int, _ CodeActionOptions, _ CancellationToken) FixEdits {
			start, err := strconv.Atoi(action.ApplyData()[0])
			if err != nil {
				return NewFixEdits(0, nil)
			}
			return NewFixEditsList(start, NewTextReplacement(start, 3, "xyz"))
		},
	})
	service := NewKustoCodeService(text).WithAnalyzers([]KustoAnalyzer{analyzer})
	actions := service.GetCodeActions(utf16Start, 0, 0, CodeActionOptionsDefault, true, kustoDiagnosticFixActorName, CancellationToken{})
	if len(actions.Actions) != 1 {
		t.Fatalf("UTF-16 fixer actions = %#v", actions.Actions)
	}
	result := service.ApplyCodeAction(actions.Actions[0].(ApplyAction), utf16Start, CodeActionOptionsDefault, CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("UTF-16 fixer result = %#v", result)
	}
	change, ok := result.Actions[0].(ChangeTextAction)
	if !ok || change.ChangedText != "// 😀\nxyz" || len(change.Changes) != 1 || change.Changes[0].Start != utf16Start || change.Changes[0].DeleteLength != 3 {
		t.Fatalf("UTF-16 fixer change = %#v, want start %d", result.Actions[0], utf16Start)
	}
}

func requireUTF16Classification(t *testing.T, ranges []ClassifiedRange, kind ClassificationKind, start, length int) {
	t.Helper()
	for _, classified := range ranges {
		if classified.Kind == kind && classified.Start == start && classified.Length == length {
			return
		}
	}
	t.Fatalf("classification %v at %d+%d not found in %#v", kind, start, length, ranges)
}

func requireUTF16Diagnostic(t *testing.T, diagnostics []Diagnostic, code string, start, length int) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Start == start && diagnostic.Length == length {
			return
		}
	}
	t.Fatalf("diagnostic %s at %d+%d not found in %#v", code, start, length, diagnostics)
}

func utf16TextLength(text string) int {
	return len(utf16.Encode([]rune(text)))
}
