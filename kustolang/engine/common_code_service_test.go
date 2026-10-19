package kustolang

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"unicode/utf16"
)

func TestScanClientParameterParts(t *testing.T) {
	length, nameStart, nameLength, indexStart, indexLength := ScanClientParameterParts("{tenant}", 0)
	if length != len("{tenant}") || nameStart != 1 || nameLength != len("tenant") || indexStart != 0 || indexLength != 0 {
		t.Fatalf("ScanClientParameterParts({tenant}) = len:%d name:%d/%d index:%d/%d", length, nameStart, nameLength, indexStart, indexLength)
	}

	text := "xx{name[-2]}yy"
	length, nameStart, nameLength, indexStart, indexLength = ScanClientParameterParts(text, 2)
	if length != len("{name[-2]}") || nameStart != 3 || nameLength != len("name") || indexStart != 8 || indexLength != len("-2") {
		t.Fatalf("ScanClientParameterParts(indexed) = len:%d name:%d/%d index:%d/%d", length, nameStart, nameLength, indexStart, indexLength)
	}

	if got := ScanClientParameter("{name[]}", 0); got != -1 {
		t.Fatalf("ScanClientParameter({name[]}) = %d, want -1", got)
	}
	if got := ScanClientParameter("{123}", 0); got != -1 {
		t.Fatalf("ScanClientParameter({123}) = %d, want -1", got)
	}
}

func TestUnknownCodeServiceDefaultsAndClientParameters(t *testing.T) {
	text := "  // lead\nT | where Tenant == {tenant[-1]}\n"
	service := NewUnknownCodeService(text)

	if service.Kind() != CodeKindsUnknown || service.Text() != text {
		t.Fatalf("unknown service = kind:%q text:%q", service.Kind(), service.Text())
	}
	if !service.IsFeatureSupported(CodeServiceFeatureCompletion, 5) {
		t.Fatal("default service should report all features as supported")
	}
	if diagnostics, ok := service.TryGetCachedDiagnostics(); ok || diagnostics != nil {
		t.Fatalf("cached diagnostics = %#v, %v", diagnostics, ok)
	}
	if diagnostics := service.GetDiagnostics(false, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want empty", diagnostics)
	}
	if diagnostics, ok := service.TryGetCachedAnalyzerDiagnostics(); ok || diagnostics != nil {
		t.Fatalf("cached analyzer diagnostics = %#v, %v", diagnostics, ok)
	}
	if diagnostics := service.GetAnalyzerDiagnostics(false, CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("analyzer diagnostics = %#v, want empty", diagnostics)
	}
	if analyzers := service.GetAnalyzers(); len(analyzers) != 0 {
		t.Fatalf("analyzers = %#v, want empty", analyzers)
	}
	if actions := service.GetCodeActions(0, 0, 0, CodeActionOptionsDefault, false, "", CancellationToken{}); len(actions.Actions) != 0 {
		t.Fatalf("code actions = %#v, want empty", actions)
	}
	if result := service.ApplyCodeAction(nil, 0, CodeActionOptionsDefault, CancellationToken{}); len(result.Actions) != 0 || result.FailureReason != "" {
		t.Fatalf("apply code action = %#v, want nothing", result)
	}
	if service.ShouldAutoComplete(0, '.', CancellationToken{}) {
		t.Fatal("ShouldAutoComplete = true, want false")
	}
	if completions := service.GetCompletionItems(0, CompletionOptionsDefault, CancellationToken{}); len(completions.Items) != 0 {
		t.Fatalf("completion items = %#v, want empty", completions)
	}
	if quickInfo := service.GetQuickInfo(0, QuickInfoOptionsDefault, CancellationToken{}); len(quickInfo.Items) != 0 {
		t.Fatalf("quick info = %#v, want empty", quickInfo)
	}
	if element := service.GetElement(5, CancellationToken{}); element.Start != 0 || element.Length != 0 {
		t.Fatalf("element = %#v, want 0/0", element)
	}
	if related := service.GetRelatedElements(0, FindRelatedOptionsNone, CancellationToken{}); len(related.Elements) != 0 {
		t.Fatalf("related = %#v, want empty", related)
	}
	if refs := service.GetClusterReferences(CancellationToken{}); len(refs) != 0 {
		t.Fatalf("cluster refs = %#v, want empty", refs)
	}
	if refs := service.GetDatabaseReferences(CancellationToken{}); len(refs) != 0 {
		t.Fatalf("database refs = %#v, want empty", refs)
	}
	if formatted := service.GetFormattedText(FormattingOptionsDefault, 12, CancellationToken{}); formatted.Text != text || formatted.Position != 12 {
		t.Fatalf("formatted = %#v", formatted)
	}

	parameters := service.GetClientParameters()
	if len(parameters) != 1 {
		t.Fatalf("parameters = %#v, want one", parameters)
	}
	parameter := parameters[0]
	if parameter.Name != "tenant" || parameter.Index == nil || *parameter.Index != -1 || parameter.Start() != strings.Index(text, "{tenant[-1]}") || parameter.Length() != len("{tenant[-1]}") {
		t.Fatalf("parameter = %#v", parameter)
	}

	if got := service.GetMinimalText(MinimalTextKindRemoveLeadingWhitespaceAndComments, CancellationToken{}); got != "T | where Tenant == {tenant[-1]}\n" {
		t.Fatalf("minimal interior = %q", got)
	}
	if got := service.GetMinimalText(MinimalTextKindMinimizeWhitespaceAndRemoveComments, CancellationToken{}); got != "T | where Tenant == {tenant[-1]}" {
		t.Fatalf("minimal compact = %q", got)
	}
	if got := service.GetMinimalText(MinimalTextKindSingleLine, CancellationToken{}); got != "T | where Tenant == {tenant[-1]}" {
		t.Fatalf("minimal single-line = %q", got)
	}
}

func TestCommonCodeServiceClassificationsPreferClientParameters(t *testing.T) {
	text := "print x={tenant}, y={workspace[2]}"
	service := NewCommonCodeService(text)
	firstStart := strings.Index(text, "{tenant}")
	secondStart := strings.Index(text, "{workspace[2]}")

	info := service.GetClassifications(0, len(text), true, false, CancellationToken{})
	want := []ClassifiedRange{
		NewClassifiedRange(ClassificationKindPlainText, 0, firstStart),
		NewClassifiedRange(ClassificationKindClientParameter, firstStart, len("{tenant}")),
		NewClassifiedRange(ClassificationKindPlainText, firstStart+len("{tenant}"), secondStart-(firstStart+len("{tenant}"))),
		NewClassifiedRange(ClassificationKindClientParameter, secondStart, len("{workspace[2]}")),
	}
	trailingStart := secondStart + len("{workspace[2]}")
	if trailingStart < len(text) {
		want = append(want, NewClassifiedRange(ClassificationKindPlainText, trailingStart, len(text)-trailingStart))
	}
	if !reflect.DeepEqual(info.Classifications, want) {
		t.Fatalf("classifications = %#v, want %#v", info.Classifications, want)
	}

	clipped := service.GetClassifications(firstStart+1, len("{tenant}")-1, true, false, CancellationToken{})
	want = []ClassifiedRange{NewClassifiedRange(ClassificationKindClientParameter, firstStart+1, len("{tenant}")-1)}
	if !reflect.DeepEqual(clipped.Classifications, want) {
		t.Fatalf("clipped classifications = %#v, want %#v", clipped.Classifications, want)
	}

	upstreamQuirk := service.GetClassifications(firstStart, len("{tenant}")-1, true, false, CancellationToken{})
	want = []ClassifiedRange{NewClassifiedRange(ClassificationKindPlainText, firstStart, len("{tenant}")-1)}
	if !reflect.DeepEqual(upstreamQuirk.Classifications, want) {
		t.Fatalf("partial end classifications = %#v, want %#v", upstreamQuirk.Classifications, want)
	}
}

func TestCommonCodeServiceAddClassificationsSecondListTakesPrecedence(t *testing.T) {
	base := []ClassifiedRange{
		NewClassifiedRange(ClassificationKindPlainText, 0, 10),
		NewClassifiedRange(ClassificationKindKeyword, 12, 4),
	}
	priority := []ClassifiedRange{
		NewClassifiedRange(ClassificationKindClientParameter, 3, 4),
		NewClassifiedRange(ClassificationKindLiteral, 12, 2),
	}

	got := CommonCodeServiceAddClassifications(base, priority)
	want := []ClassifiedRange{
		NewClassifiedRange(ClassificationKindPlainText, 0, 3),
		NewClassifiedRange(ClassificationKindClientParameter, 3, 4),
		NewClassifiedRange(ClassificationKindPlainText, 7, 3),
		NewClassifiedRange(ClassificationKindLiteral, 12, 2),
		NewClassifiedRange(ClassificationKindKeyword, 14, 2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged classifications = %#v, want %#v", got, want)
	}
}

func TestCommonCodeServiceOutlinesUseFirstTokenLine(t *testing.T) {
	text := "  T | take 1\n| count  \n"
	service := NewCommonCodeService(text)
	outlines := service.GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(outlines.Ranges) != 1 {
		t.Fatalf("outlines = %#v, want one range", outlines)
	}
	outline := outlines.Ranges[0]
	if outline.Start != 0 || outline.Length != TrimEnd(text) || outline.CollapsedText != "T | take 1 ..." {
		t.Fatalf("outline = %#v", outline)
	}

	empty := NewCommonCodeService("  // comment only\n").GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(empty.Ranges) != 0 {
		t.Fatalf("comment-only outlines = %#v, want empty", empty)
	}
}

func TestCommonCodeServiceUsesUTF16PublicRangesConcurrently(t *testing.T) {
	text := "// 😀\nprint x={tenant}"
	service := NewCommonCodeService(text)
	parameterByteStart := strings.Index(text, "{tenant}")
	parameterStart := len(utf16.Encode([]rune(text[:parameterByteStart])))
	textLength := len(utf16.Encode([]rune(text)))

	parameters := service.GetClientParameters()
	if len(parameters) != 1 || parameters[0].Start() != parameterStart || parameters[0].Length() != len("{tenant}") {
		t.Fatalf("UTF-16 common parameters = %#v, want %d+8", parameters, parameterStart)
	}
	classifications := service.GetClassifications(0, textLength, true, false, CancellationToken{}).Classifications
	wantParameter := NewClassifiedRange(ClassificationKindClientParameter, parameterStart, len("{tenant}"))
	found := false
	for _, classification := range classifications {
		if reflect.DeepEqual(classification, wantParameter) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("UTF-16 common classifications = %#v, want %#v", classifications, wantParameter)
	}
	outlines := service.GetOutlines(OutliningOptionsDefault, CancellationToken{})
	if len(outlines.Ranges) != 1 || outlines.Ranges[0].Length != textLength {
		t.Fatalf("UTF-16 common outlines = %#v, want length %d", outlines, textLength)
	}

	var wait sync.WaitGroup
	for i := 0; i < 32; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got := service.GetClientParameters()
			if len(got) != 1 || got[0].Start() != parameterStart {
				t.Errorf("concurrent UTF-16 parameters = %#v", got)
			}
		}()
	}
	wait.Wait()
}

func TestCommonCodeServiceClassifiesRangeInsideSurrogatePair(t *testing.T) {
	service := NewCommonCodeService("😀{tenant}")
	classifications := service.GetClientParametersClassifications(1, 9, true)
	want := []ClassifiedRange{NewClassifiedRange(ClassificationKindClientParameter, 2, len("{tenant}"))}
	if !reflect.DeepEqual(classifications, want) {
		t.Fatalf("surrogate-interior classifications = %#v, want %#v", classifications, want)
	}
}
