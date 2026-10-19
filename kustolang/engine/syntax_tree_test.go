package kustolang

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSyntaxTreeOwnsNodesAndProjectsNestedOriginalPositions(t *testing.T) {
	leaf := NewSyntaxNode(SyntaxKindNameReference, nil)
	fragmentRoot := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{leaf})
	original := NewSyntaxTree(NewSyntaxNode(SyntaxKindQueryBlock, nil), nil, 0)
	fragment := NewSyntaxTree(fragmentRoot, original, 12)
	nestedLeaf := NewSyntaxNode(SyntaxKindStringLiteralExpression, nil)
	nested := NewSyntaxTree(NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{nestedLeaf}), fragment, 7)

	if fragmentRoot.Tree() != fragment || leaf.Tree() != fragment || nestedLeaf.Tree() != nested {
		t.Fatalf("tree ownership = root:%p leaf:%p nested:%p", fragmentRoot.Tree(), leaf.Tree(), nestedLeaf.Tree())
	}
	if got := nestedLeaf.GetPositionInOriginalTree(3); got != 22 {
		t.Fatalf("nested original position = %d, want 22", got)
	}
	if got := leaf.GetPositionInOriginalTree(3); got != 15 {
		t.Fatalf("fragment original position = %d, want 15", got)
	}
}

func TestParseBuildsSyntaxTreeAndTokenStarts(t *testing.T) {
	code := Parse("print a=true | project a")
	if code.Kind != CodeKindQuery {
		t.Fatalf("Kind = %s", code.Kind)
	}
	if code.Syntax() == nil {
		t.Fatal("missing syntax root")
	}
	if code.MaxDepth() < 5 {
		t.Fatalf("MaxDepth = %d", code.MaxDepth())
	}
	if len(code.TokenStarts) != len(code.Tokens)+1 {
		t.Fatalf("starts=%d tokens=%d", len(code.TokenStarts), len(code.Tokens))
	}
	if got := code.Syntax().Text(IncludeTriviaAll); got != code.Text {
		t.Fatalf("syntax text = %q, want %q", got, code.Text)
	}
}

func TestParseSurfacesTokenDiagnostics(t *testing.T) {
	for _, query := range []string{
		"print a='unterminated",
		"print a=datetime(2024-01-01",
	} {
		t.Run(query, func(t *testing.T) {
			code := Parse(query)
			diags := code.GetDiagnostics()
			if len(diags) == 0 {
				t.Fatalf("diagnostics = %#v", diags)
			}
			if diags[0].Code != "KS001" {
				t.Fatalf("diagnostic code = %s", diags[0].Code)
			}
			syntaxDiags := code.GetSyntaxDiagnostics()
			if len(syntaxDiags) == 0 || syntaxDiags[0].Code != "KS001" {
				t.Fatalf("syntax diagnostics = %#v", syntaxDiags)
			}
		})
	}
}

func TestKustoCodeLineOffsetAndTokenHelpers(t *testing.T) {
	const text = "print a=true\n| project a\n"
	code := Parse(text)

	line, offset, ok := code.TryGetLineAndOffset(0)
	if !ok || line != 1 || offset != 1 {
		t.Fatalf("line/offset at start = %d, %d, %t", line, offset, ok)
	}
	barPosition := 13
	line, offset, ok = code.TryGetLineAndOffset(barPosition)
	if !ok || line != 2 || offset != 1 {
		t.Fatalf("line/offset at pipe = %d, %d, %t", line, offset, ok)
	}
	line, offset, ok = code.TryGetLineAndOffset(len(text))
	if !ok || line != 3 || offset != 1 {
		t.Fatalf("line/offset at trailing empty line = %d, %d, %t", line, offset, ok)
	}
	if _, _, ok := code.TryGetLineAndOffset(len(text) + 1); ok {
		t.Fatal("position beyond text should not resolve")
	}

	if got := code.GetTokenIndex(-1); got != -1 {
		t.Fatalf("token index before start = %d", got)
	}
	if got := code.GetTokenIndex(0); got != 0 {
		t.Fatalf("token index at first token = %d", got)
	}
	if got := code.GetTokenIndex(6); got != 1 {
		t.Fatalf("token index in second token = %d", got)
	}
	if got := code.GetTokenIndex(len(text) + 10); got != len(code.Tokens)-1 {
		t.Fatalf("token index beyond text = %d, want end token %d", got, len(code.Tokens)-1)
	}

	tokens := code.GetLexicalTokens()
	if len(tokens) != len(code.Tokens) {
		t.Fatalf("lexical token count = %d, want %d", len(tokens), len(code.Tokens))
	}
	tokens[0].Text = "mutated"
	if code.Tokens[0].Text == "mutated" {
		t.Fatal("GetLexicalTokens should return a defensive copy")
	}
}

func TestKustoCodeNonAnalyzedSemanticScopeHelpers(t *testing.T) {
	code := Parse("T | project Account")
	if code.HasSemantics() {
		t.Fatal("native KustoCode should not claim full semantics yet")
	}
	if got := code.GetSymbolsInScope(0, SymbolMatchAny, IncludeFunctionKindAll); len(got) != 0 {
		t.Fatalf("symbols in scope without semantics = %#v", got)
	}
	if got := code.GetSpeculativeReferencedSymbol(0, "Account", SymbolMatchAny); got != nil {
		t.Fatalf("speculative symbol without semantics = %#v", got)
	}
	if got := code.GetColumnsInScope(0); got != nil {
		t.Fatalf("columns in scope without semantics = %#v", got)
	}
	if IncludeFunctionKindAll != IncludeFunctionKindBuiltInFunctions|IncludeFunctionKindDatabaseFunctions|IncludeFunctionKindLocalFunctions|IncludeFunctionKindLocalViews {
		t.Fatalf("include function flags = %d", IncludeFunctionKindAll)
	}
}

func TestTextFactsLineHelpersMirrorUpstreamEdges(t *testing.T) {
	const text = "alpha\r\n  beta\n\nomega"
	if got := NextLineBreakStart(text, 0); got != 5 {
		t.Fatalf("next line break start = %d", got)
	}
	if got := NextLineBreakEnd(text, 0); got != 7 {
		t.Fatalf("next line break end = %d", got)
	}
	if got := LineStart(text, 9); got != 7 {
		t.Fatalf("line start = %d", got)
	}
	if got := LineEnd(text, 7); got != 13 {
		t.Fatalf("line end = %d", got)
	}
	if previous, ok := TryGetPreviousLineEnd(text, 7); !ok || previous != 5 {
		t.Fatalf("previous line end = %d, %t", previous, ok)
	}
	if next, ok := TryGetNextLineStart(text, 7); !ok || next != 14 {
		t.Fatalf("next line start = %d, %t", next, ok)
	}
	if got := FirstLineBreakStart(text); got != 5 {
		t.Fatalf("first line break start = %d", got)
	}
	if got := FirstLineBreakEnd(text); got != 7 {
		t.Fatalf("first line break end = %d", got)
	}
	if got := LastLineBreakStart(text); got != 14 {
		t.Fatalf("last line break start = %d", got)
	}
	if got := LastLineBreakEnd(text); got != 15 {
		t.Fatalf("last line break end = %d", got)
	}
	if got := LineLength(text, 7, true); got != 7 {
		t.Fatalf("line length including break = %d", got)
	}
	if got := PositionAfterLeadingWhitespace(text, 7); got != 9 {
		t.Fatalf("position after leading whitespace = %d", got)
	}
	if !IsBlankLine(text, 14) {
		t.Fatal("empty line should be blank")
	}
	if IsBlankLine(text, 7) {
		t.Fatal("line with content should not be blank")
	}
	if start, ok := TryGetTextLineStart(text, 4); !ok || start != 15 {
		t.Fatalf("text line start = %d, %t", start, ok)
	}
	if _, ok := TryGetTextLineStart(text, 5); ok {
		t.Fatal("line beyond text should not resolve")
	}
	if line, offset, ok := TryGetTextLineAndOffset(text, 16); !ok || line != 4 || offset != 2 {
		t.Fatalf("text line/offset = %d, %d, %t", line, offset, ok)
	}
	if position, ok := TryGetTextPosition(text, 4, 2); !ok || position != 16 {
		t.Fatalf("text position = %d, %t", position, ok)
	}
	if _, ok := TryGetTextPosition(text, 4, 99); ok {
		t.Fatal("position beyond text should not resolve")
	}
}

func TestKustoCodeLineStartsCacheIsConcurrentAndPerInstance(t *testing.T) {
	code := Parse("one\ntwo\r\nthree")
	if code.lineStartsCache == nil {
		t.Fatal("parsed code is missing line-start cache")
	}

	type expectation struct {
		position int
		line     int
		offset   int
	}
	expectations := []expectation{
		{position: 0, line: 1, offset: 1},
		{position: 5, line: 2, offset: 2},
		{position: 9, line: 3, offset: 1},
	}
	problems := make(chan string, 1)
	var group sync.WaitGroup
	for range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			for _, want := range expectations {
				line, offset, ok := code.TryGetLineAndOffset(want.position)
				if ok && line == want.line && offset == want.offset {
					continue
				}
				select {
				case problems <- fmt.Sprintf("position %d = %d/%d/%t, want %d/%d/true", want.position, line, offset, ok, want.line, want.offset):
				default:
				}
				return
			}
		}()
	}
	group.Wait()
	select {
	case problem := <-problems:
		t.Fatal(problem)
	default:
	}
	if got := code.lineStartsCache.starts; len(got) != 3 || got[0] != 0 || got[1] != 4 || got[2] != 9 {
		t.Fatalf("cached line starts = %#v", got)
	}

	clone := code.WithGlobals(NewGlobalState(NewDatabaseSymbol("other")))
	if clone == code || clone.lineStartsCache == nil || clone.lineStartsCache == code.lineStartsCache {
		t.Fatalf("line-start cache must be fresh for rebuilt code: original=%p clone=%p", code.lineStartsCache, clone.lineStartsCache)
	}
	if _, _, ok := clone.TryGetLineAndOffset(9); !ok || len(clone.lineStartsCache.starts) != 3 {
		t.Fatalf("rebuilt code line starts = %#v", clone.lineStartsCache.starts)
	}
}

func TestKustoCodeParseAndAnalyzeAttachesSemantics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	code := ParseAndAnalyze("T | project Account", NewGlobalState(NewDatabaseSymbol("db", table)))
	if !code.HasSemantics() {
		t.Fatal("ParseAndAnalyze did not record performed semantic analysis")
	}
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("analyzed diagnostics = %#v", diagnostics)
	}
	project := firstBoundNode[*ProjectOperator](t, code.Query)
	if result := expressionTableResult(project); result == nil || result.GetColumn("Account") == nil {
		t.Fatalf("analyzed project result = %#v", result)
	}
}

func TestSyntaxTokenLiteralValues(t *testing.T) {
	code := Parse("print a='x', b=12, c=0xC, d=1.5, e=true")
	values := map[string]any{}
	WalkSyntax(code.Syntax(), func(child SyntaxElement) {
		token, ok := child.(*SyntaxToken)
		if !ok {
			return
		}
		switch token.Kind() {
		case SyntaxKindStringLiteralToken, SyntaxKindIntLiteralToken, SyntaxKindLongLiteralToken, SyntaxKindRealLiteralToken, SyntaxKindBooleanLiteralToken:
			values[token.RawText()] = token.Value()
		}
	})
	if values["'x'"] != "x" || values["12"] != int64(12) || values["0xC"] != int64(12) || values["1.5"] != 1.5 || values["true"] != true {
		t.Fatalf("values = %#v", values)
	}
}

func TestSyntaxTokenTypedLiteralValuesAndNulls(t *testing.T) {
	guid := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	tests := []struct {
		name string
		kind SyntaxKind
		text string
		want any
	}{
		{name: "decimal", kind: SyntaxKindDecimalLiteralToken, text: "decimal(1.25)", want: big.NewRat(5, 4)},
		{name: "timespan suffix", kind: SyntaxKindTimespanLiteralToken, text: "timespan(1.5h)", want: TimeSpanValue(54_000_000_000)},
		{name: "timespan clock", kind: SyntaxKindTimespanLiteralToken, text: "time(1.02:03:04.005)", want: TimeSpanValue(937_840_050_000)},
		{name: "datetime", kind: SyntaxKindDateTimeLiteralToken, text: "datetime(2024-01-02T03:04:05Z)", want: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)},
		{name: "guid", kind: SyntaxKindGuidLiteralToken, text: "guid(00010203-0405-0607-0809-0a0b0c0d0e0f)", want: guid},
		{name: "raw guid", kind: SyntaxKindRawGuidLiteralToken, text: "00010203-0405-0607-0809-0a0b0c0d0e0f", want: guid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := NewLiteralToken("", test.text, test.kind)
			if err != nil {
				t.Fatal(err)
			}
			if got := token.Value(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Value() = %#v (%T), want %#v (%T)", got, got, test.want, test.want)
			}
		})
	}

	for _, test := range []struct {
		kind SyntaxKind
		text string
	}{
		{kind: SyntaxKindBooleanLiteralToken, text: "bool(null)"},
		{kind: SyntaxKindIntLiteralToken, text: "int(null)"},
		{kind: SyntaxKindLongLiteralToken, text: "long(null)"},
		{kind: SyntaxKindRealLiteralToken, text: "real(null)"},
		{kind: SyntaxKindDecimalLiteralToken, text: "decimal(null)"},
		{kind: SyntaxKindTimespanLiteralToken, text: "timespan(null)"},
		{kind: SyntaxKindDateTimeLiteralToken, text: "datetime(null)"},
		{kind: SyntaxKindGuidLiteralToken, text: "guid(null)"},
	} {
		token, err := NewLiteralToken("", test.text, test.kind)
		if err != nil {
			t.Fatal(err)
		}
		if got := token.Value(); got != nil {
			t.Fatalf("%v null Value() = %#v, want nil", test.kind, got)
		}
	}

	zeroGuid := [16]byte{}
	zeroValues := []struct {
		kind SyntaxKind
		text string
		want any
	}{
		{kind: SyntaxKindDecimalLiteralToken, text: "decimal(bad)", want: new(big.Rat)},
		{kind: SyntaxKindTimespanLiteralToken, text: "timespan(bad)", want: TimeSpanValue(0)},
		{kind: SyntaxKindDateTimeLiteralToken, text: "datetime(bad)", want: time.Time{}},
		{kind: SyntaxKindGuidLiteralToken, text: "guid(bad)", want: zeroGuid},
		{kind: SyntaxKindGuidLiteralToken, text: `guid("null")`, want: zeroGuid},
		{kind: SyntaxKindStringLiteralToken, text: `"null"`, want: "null"},
	}
	for _, test := range zeroValues {
		token, err := NewLiteralToken("", test.text, test.kind)
		if err != nil {
			t.Fatal(err)
		}
		if got := token.Value(); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%s Value() = %#v (%T), want %#v (%T)", test.text, got, got, test.want, test.want)
		}
	}
}

func TestSyntaxTokenDecimalAndTimeSpanPreserveUpstreamRange(t *testing.T) {
	decimal, _ := NewLiteralToken("", "decimal(79228162514264337593543950335)", SyntaxKindDecimalLiteralToken)
	wantDecimal, _ := new(big.Rat).SetString("79228162514264337593543950335")
	if got := decimal.Value(); !reflect.DeepEqual(got, wantDecimal) {
		t.Fatalf("decimal max = %v, want %v", got, wantDecimal)
	}

	span, _ := NewLiteralToken("", "timespan(max)", SyntaxKindTimespanLiteralToken)
	value := span.Value().(TimeSpanValue)
	if value.Ticks() != math.MaxInt64 {
		t.Fatalf("timespan max ticks = %d", value.Ticks())
	}
	if _, ok := value.Duration(); ok {
		t.Fatal("full TimeSpan range must not silently overflow time.Duration")
	}
	short := TimeSpanValue(timeSpanTicksPerSecond)
	if duration, ok := short.Duration(); !ok || duration != time.Second {
		t.Fatalf("one-second duration = %v, %t", duration, ok)
	}
}

func TestSyntaxTokenPrimitiveMinMaxValues(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		text string
		want any
	}{
		{SyntaxKindIntLiteralToken, "int(min)", int(math.MinInt32)},
		{SyntaxKindIntLiteralToken, "int(max)", int(math.MaxInt32)},
		{SyntaxKindLongLiteralToken, "long(min)", int64(math.MinInt64)},
		{SyntaxKindLongLiteralToken, "long(max)", int64(math.MaxInt64)},
		{SyntaxKindRealLiteralToken, "real(min)", -math.MaxFloat64},
		{SyntaxKindRealLiteralToken, "real(max)", math.MaxFloat64},
	}
	for _, test := range tests {
		token, err := NewLiteralToken("", test.text, test.kind)
		if err != nil {
			t.Fatal(err)
		}
		if got := token.Value(); got != test.want {
			t.Fatalf("%s Value() = %#v, want %#v", test.text, got, test.want)
		}
	}
}

func TestSyntaxTreeDepthCountsOnlyNodes(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x"))
	leaf := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{token})
	list := NewSyntaxList(leaf)
	tree := NewSyntaxTree(NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{list}), nil, 0)

	if got := tree.Depth(); got != 2 {
		t.Fatalf("Depth() = %d, want 2 node levels", got)
	}
}

func TestCodeKindOf(t *testing.T) {
	if CodeKindOf(".show tables") != CodeKindCommand {
		t.Fatal("expected command")
	}
	if CodeKindOf("#crp") != CodeKindQuery {
		t.Fatal("expected standalone directive text to fall back to query")
	}
	if CodeKindOf("#crp\n.show tables") != CodeKindCommand {
		t.Fatal("expected directive line before command to be skipped")
	}
	if CodeKindOf("Table | count") != CodeKindQuery {
		t.Fatal("expected query")
	}
}
