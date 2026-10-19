package kustolang

import (
	"strconv"
	"testing"
)

func TestQueryParserDatabaseElementCommaTailOwnership(t *testing.T) {
	query := ParseQuery(`database("db")["T"], Tail | project A`)
	skipped, ok := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
	if !ok || skipped == nil || skipped.Tokens() == nil || skipped.Tokens().Count() == 0 || skipped.Tokens().At(0).Kind() != SyntaxKindCommaToken {
		t.Fatalf("skipped tail = %T %#v", query.GetChild(2), query.GetChild(2))
	}
}

func TestQueryParserJoinCommaTailBelongsEntirelyToSkippedTokens(t *testing.T) {
	const text = "datatable(A:long)[1]\n| join (datatable(A:long)[1]), datatable(B:long)[1] on A"
	for _, parse := range []func(string) *QueryBlock{ParseQuery, ParseQuery} {
		query := parse(text)
		skipped, ok := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
		if !ok || skipped == nil || skipped.Tokens() == nil || skipped.Tokens().Count() == 0 {
			t.Fatalf("skipped join tail = %T %#v", query.GetChild(2), query.GetChild(2))
		}
		if first := skipped.Tokens().At(0); first.Kind() != SyntaxKindCommaToken || !diagnosticsContain(first.Diagnostics(), "KS198") {
			t.Fatalf("first skipped token = %#v, want comma with KS198", first)
		}
		if got := skipped.Text(IncludeTriviaInterior); got != `, datatable(B:long)[1] on A` {
			t.Fatalf("skipped join tail text = %q", got)
		}
	}
}

func TestQueryParserContextualDataTableCommaTailBelongsToSkippedTokens(t *testing.T) {
	const text = `__contextual_datatable guid(00000000-0000-0000-0000-000000000001) (A:string), Extra`
	for _, parse := range []func(string) *QueryBlock{ParseQuery, ParseQuery} {
		query := parse(text)
		skipped, ok := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
		if !ok || skipped == nil || skipped.Tokens() == nil || skipped.Tokens().Count() == 0 {
			t.Fatalf("skipped contextual-datatable tail = %T %#v", query.GetChild(2), query.GetChild(2))
		}
		first := skipped.Tokens().At(0)
		if first.Kind() != SyntaxKindCommaToken || !diagnosticsContain(first.Diagnostics(), "KS198") {
			t.Fatalf("first skipped token = %#v, want comma with KS198", first)
		}
		if got := skipped.Text(IncludeTriviaInterior); got != `, Extra` {
			t.Fatalf("skipped contextual-datatable tail text = %q", got)
		}
	}
}

func TestQueryParserOperatorTailKS198BelongsToFirstSkippedToken(t *testing.T) {
	tests := []struct {
		row   int
		query string
		token string
	}{
		{119, "T | take 1, 2", ","},
		{136, "T | render timechart with (xcolumn=Timestamp), Extra", ","},
		{173, "T | where A > 1, Extra", ","},
		{192, "T | reduce by X, Y | project Pattern, Count", ","},
		{212, `T | __partitionby X id "slice", Z (project P=X) | project P`, ","},
		{254, "SigninLogs | project UserPrincipalName | __executeAndCache, Extra", ","},
		{259, "T | partition by X, Z in dynamic([1]) (project P=X) | project P", ","},
		{263, "T | partition by User (summarize Count=count() by User), Extra | project User, Count", ","},
		{316, "T | as Alias, Extra | project X", ","},
		{387, "materialized-view-combine(\"v\") base(T) delta(T) aggregations(summarize Count=count() by Account), Extra | project Account, Count", ","},
		{392, "T | count, Extra", ","},
		{393, "datatable(x:long)[1] | count as Total, Extra", ","},
		{425, "range x from 1 2 to 10 step 1 | project x", "to"},
		{445, "range i from 1 to 3 step 1 | getschema kind=csl, Extra", ","},
		{447, "T | getschema, Extra", ","},
		{523, "LeftTable | join (RightTable) where A > 1, Extra", ","},
		{768, "T | invoke Keep(), Extra", ","},
		{826, "T | mv-apply Item=Items to typeof(string) on (summarize ItemCount=count() by Item), Extra", ","},
		{829, "T | consume, Extra", ","},
		{834, "find in (T, U) where Score > 0, Extra project x, y", ","},
		{860, "T | make-series Hits=count() on TimeGenerated from ago(1d) step 1h to now() by Account", "to"},
		{1099, "T | top 5 by A with others=x | project A", "others"},
	}

	for _, tt := range tests {
		t.Run(parserCorpusRowName(tt.row), func(t *testing.T) {
			parser := newQueryParserState(Lex(tt.query))
			parser.parseQueryBlockStatementList()
			skipped := parser.parseQuerySkippedTokens()
			if skipped == nil || skipped.Tokens().Count() == 0 {
				t.Fatal("operator parser consumed the unexpected tail")
			}
			token := skipped.Tokens().At(0)
			if token.RawText() != tt.token {
				t.Fatalf("first skipped token = %q, want %q", token.RawText(), tt.token)
			}
			if !diagnosticsContain(token.Diagnostics(), "KS198") {
				t.Fatalf("first skipped token does not own KS198: %#v", token.Diagnostics())
			}
		})
	}
}

func TestQueryParserMalformedGraphPatternRetainsTypedStatementsAndSkippedTail(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []diagnosticLocation
	}{
		{
			name:  "graph-match",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-match (a)-e]->(b) where e.Weight > 0 project Src=a.Source, Weight=e.Weight`,
			diagnostics: []diagnosticLocation{
				{"KS005", 111, 1},
				{"KS142", 112, 1},
				{"KS198", 113, 1},
			},
		},
		{
			name:  "graph-shortest-paths",
			query: `let G=datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target; G | graph-shortest-paths (a)-e]->(b) where e.Weight > 0 project Src=a.Source, Dst=b.Target`,
			diagnostics: []diagnosticLocation{
				{"KS005", 120, 1},
				{"KS142", 121, 1},
				{"KS198", 122, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.query)
			statementList, _ := query.GetChild(1).(*SyntaxList)
			if statements := AsSeparatedSyntaxListOf[Statement](statementList); statements == nil || statements.Count() < 2 {
				t.Fatalf("statement list = %#v, want retained typed statements", query.GetChild(1))
			}
			skipped, ok := WrapGeneratedSyntaxElement(query.GetChild(2)).(*SkippedTokens)
			if !ok || skipped == nil || skipped.Tokens() == nil || skipped.Tokens().Count() == 0 {
				t.Fatalf("skipped graph tail = %T %#v", query.GetChild(2), query.GetChild(2))
			}
			if first := skipped.Tokens().At(0); first.Kind() != SyntaxKindCloseBracketToken || !diagnosticsContain(first.Diagnostics(), "KS198") {
				t.Fatalf("first skipped token = %#v, want ] with KS198", first)
			}
			assertDiagnosticLocations(t, ParseAndAnalyze(tt.query).GetDiagnostics(), tt.diagnostics...)
		})
	}
}

func parserCorpusRowName(row int) string {
	return "row-" + strconv.Itoa(row)
}
