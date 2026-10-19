package kustolang

import (
	"strings"
	"testing"
)

func TestKustoClassifierClassifiesSyntaxAndTrivia(t *testing.T) {
	text := "// c\nSigninLogs | where ResultType != long(0) and Message == 'x'"
	ranges := kustoClassifierRanges(Parse(text), text, true)

	kustoClassifierRequireRange(t, text, ranges, "// c", ClassificationKindComment)
	kustoClassifierRequireRange(t, text, ranges, "SigninLogs", ClassificationKindIdentifier)
	kustoClassifierRequireRange(t, text, ranges, "|", ClassificationKindPunctuation)
	kustoClassifierRequireRange(t, text, ranges, "where", ClassificationKindQueryOperator)
	kustoClassifierRequireRange(t, text, ranges, "!=", ClassificationKindMathOperator)
	kustoClassifierRequireRange(t, text, ranges, "long", ClassificationKindKeyword)
	kustoClassifierRequireRange(t, text, ranges, "(0)", ClassificationKindLiteral)
	kustoClassifierRequireRange(t, text, ranges, "and", ClassificationKindScalarOperator)
	kustoClassifierRequireRange(t, text, ranges, "'x'", ClassificationKindStringLiteral)
}

func TestKustoClassifierHandlesUnicodeTriviaByRune(t *testing.T) {
	text := "\u2003// c\u2028T"
	ranges := kustoClassifierRanges(Parse(text), text, true)
	kustoClassifierRequireRange(t, text, ranges, "// c", ClassificationKindComment)
	kustoClassifierRequireRange(t, text, ranges, "T", ClassificationKindIdentifier)

	for _, r := range ranges {
		if r.Start < len("\u2003") {
			t.Fatalf("Unicode whitespace should not be classified: %#v", ranges)
		}
		separatorStart := len("\u2003// c")
		if r.Start < separatorStart+len("\u2028") && r.End() > separatorStart {
			t.Fatalf("Unicode line separator should not be part of a classification: %#v", ranges)
		}
	}
}

func TestKustoClassifierClipsAndMergesRanges(t *testing.T) {
	commentText := "// comment\nT"
	commentCode := Parse(commentText)
	commentRanges := KustoClassifierGetClassifications(commentCode.Syntax(), 3, 4, true, nil, CancellationToken{})
	if len(commentRanges) != 1 {
		t.Fatalf("comment ranges = %#v, want one clipped range", commentRanges)
	}
	if got := commentRanges[0]; got.Start != 3 || got.Length != 4 || got.Kind != ClassificationKindComment {
		t.Fatalf("clipped comment range = %#v, want start=3 length=4 kind=Comment", got)
	}

	punctuationText := "T | summarize count()"
	punctuationRanges := kustoClassifierRanges(Parse(punctuationText), punctuationText, true)
	kustoClassifierRequireRange(t, punctuationText, punctuationRanges, "()", ClassificationKindPunctuation)
}

func TestKustoClassifierSemanticIdentifierKinds(t *testing.T) {
	text := "T | project Account"
	code := Parse(text)
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	lastParsedNameReferenceBySimpleName(t, code, "Account").SetSemanticInfo(NewSemanticInfoForSymbol(column, ScalarTypes.String.Type()))

	ranges := kustoClassifierRanges(code, text, true)
	kustoClassifierRequireRange(t, text, ranges, "Account", ClassificationKindColumn)
}

func TestKustoClassifierQueryAndClientParameters(t *testing.T) {
	queryParameterText := `externaldata(x:string)["https://example.test/data.csv"] with (foo=1)`
	queryParameterRanges := kustoClassifierRanges(Parse(queryParameterText), queryParameterText, true)
	kustoClassifierRequireRange(t, queryParameterText, queryParameterRanges, "foo", ClassificationKindQueryParameter)

	clientParameterText := "print {tenant}"
	clientParameterRanges := kustoClassifierRanges(Parse(clientParameterText), clientParameterText, true)
	kustoClassifierRequireRange(t, clientParameterText, clientParameterRanges, "{tenant}", ClassificationKindClientParameter)
}

func TestKustoClassifierClassifiesCustomAndUnknownCommandVerbs(t *testing.T) {
	tests := []struct {
		name string
		text string
		verb string
	}{
		{name: "custom command", text: ".show version", verb: "show"},
		{name: "unknown command", text: ".not_a_real_command version", verb: "not_a_real_command"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info := NewKustoCodeService(test.text).GetClassifications(0, len(test.text), true, true, CancellationToken{})
			kustoClassifierRequireRange(t, test.text, info.Classifications, "."+test.verb, ClassificationKindCommand)
		})
	}
}

func kustoClassifierRanges(code *KustoCode, text string, clipToRange bool) []ClassifiedRange {
	return KustoClassifierGetClassifications(code.Syntax(), 0, len(text), clipToRange, nil, CancellationToken{})
}

func kustoClassifierRequireRange(t *testing.T, text string, ranges []ClassifiedRange, needle string, kind ClassificationKind) {
	t.Helper()
	start := strings.Index(text, needle)
	if start < 0 {
		t.Fatalf("needle %q not found in %q", needle, text)
	}
	length := len(needle)
	for _, r := range ranges {
		if r.Start == start && r.Length == length && r.Kind == kind {
			return
		}
	}
	t.Fatalf("range %q at [%d,%d) kind=%v not found in %#v", needle, start, start+length, kind, ranges)
}
