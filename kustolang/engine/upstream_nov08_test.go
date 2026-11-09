package kustolang

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

const nov08Commit = "dbd7ca7684adfd7a38536f74728670dc54a67acc"

type nov08Fixture struct {
	Commit    string
	DateTimes []struct {
		Text, Interpretation string
		Ticks                *int64
	}
	Diagnostics []struct {
		Query       string
		Diagnostics []attestedDiagnostic
	}
	Parsers []struct {
		Query       string
		Diagnostics []attestedDiagnostic
		Kinds       []string
	}
	Completions []struct {
		Text  string
		Items []string
	}
	Prefixes []struct {
		Text   string
		Length int
	}
	Commands []attestedLanguageResult
}

func readNov08Fixture(t *testing.T) nov08Fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/upstream_dbd7ca7_behavior.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture nov08Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != nov08Commit || len(fixture.DateTimes) < 100 || len(fixture.Commands) != 18 || len(fixture.Parsers) != 6 || len(fixture.Completions) != 5 || len(fixture.Prefixes) != 2 {
		t.Fatal("incomplete or incorrectly pinned November oracle")
	}
	return fixture
}
func TestUpstreamNov08DateTimeInterpretation(t *testing.T) {
	for _, want := range readNov08Fixture(t).DateTimes {
		t.Run(want.Text, func(t *testing.T) {
			interpretation, value := parseKustoDateTime(want.Text)
			if string(interpretation) != want.Interpretation {
				t.Fatalf("interpretation = %s, want %s", interpretation, want.Interpretation)
			}
			if want.Ticks != nil && datetimeToTicks(value) != *want.Ticks {
				t.Errorf("ticks = %d, want %d", datetimeToTicks(value), *want.Ticks)
			}
		})
	}
}
func TestUpstreamNov08DateTimeAnalyzer(t *testing.T) {
	for _, want := range readNov08Fixture(t).Diagnostics {
		t.Run(want.Query, func(t *testing.T) {
			service := NewKustoCodeService(want.Query).WithAnalyzers([]KustoAnalyzer{NewDateTimeLiteralAnalyzer()})
			got := []attestedDiagnostic{}
			for _, d := range service.GetAnalyzerDiagnostics(true, CancellationToken{}) {
				got = append(got, attestedDiagnostic{d.Code, d.Message, d.Start, d.Length})
			}
			if !reflect.DeepEqual(got, want.Diagnostics) {
				t.Errorf("diagnostics = %#v, want %#v", got, want.Diagnostics)
			}
		})
	}
}
func TestUpstreamNov08PolicyCommands(t *testing.T) {
	for _, want := range readNov08Fixture(t).Commands {
		t.Run(want.Query, func(t *testing.T) {
			got := currentAttestedLanguageResult(want.Query)
			if len(got.Columns) == 0 {
				got.Columns = nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("command = %#v, want %#v", got, want)
			}
		})
	}
	for _, policy := range []string{"snapshot_indexing", "graph_query"} {
		text := ".show graph_model M policy "
		items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
		assertCompletionIncludes(t, items, policy)
	}
}
func TestUpstreamNov08Grammar(t *testing.T) {
	for _, want := range readNov08Fixture(t).Parsers {
		t.Run(want.Query, func(t *testing.T) {
			globals := DefaultGlobalState.WithParseOptions(DefaultParseOptions.WithParserKind(ParserKindGrammar))
			code := ParseAndAnalyze(want.Query, globals)
			got := []attestedDiagnostic{}
			for _, d := range code.GetDiagnostics() {
				got = append(got, attestedDiagnostic{d.Code, d.Message, d.Start, d.Length})
			}
			if !reflect.DeepEqual(got, want.Diagnostics) {
				t.Errorf("diagnostics = %#v, want %#v", got, want.Diagnostics)
			}
			// Ensure has_any/all are expression nodes in both parser paths.
			walkSyntaxWithTriviaStart(code.Syntax(), 0, func(e SyntaxElement, _ int) {
				switch WrapGeneratedSyntaxElement(e).(type) {
				case *HasAnyExpression:
					if e.Kind() != SyntaxKindHasAnyExpression {
						t.Error("has_any keyword used as node kind")
					}
				case *HasAllExpression:
					if e.Kind() != SyntaxKindHasAllExpression {
						t.Error("has_all keyword used as node kind")
					}
				}
			})
		})
	}
}

func TestUpstreamNov08CompletionAndRecovery(t *testing.T) {
	fixture := readNov08Fixture(t)
	options := CompletionOptionsDefault.WithIncludeSymbols(false).WithIncludeExtendedSyntax(false)
	for _, want := range fixture.Completions {
		items := NewKustoCodeService(want.Text).GetCompletionItems(utf16TextLength(want.Text), options, CancellationToken{}).Items
		got := completionDisplayTexts(items)
		sort.Strings(got)
		if !reflect.DeepEqual(got, want.Items) {
			t.Errorf("%q completions = %v, want %v", want.Text, got, want.Items)
		}
	}
	grammar := QueryGrammarFrom(nil)
	for i, want := range fixture.Prefixes {
		var got int
		if i == 0 {
			got = grammar.Expression.Scan(NewArraySource(Lex(want.Text)), 0)
		} else {
			got = grammar.QueryOperator.Scan(NewArraySource(Lex(want.Text)), 0)
		}
		if got != want.Length {
			t.Errorf("%q consumed %d tokens, want %d", want.Text, got, want.Length)
		}
	}
}
