package kustolang

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestUpstream13SourceOracleCompatibility(t *testing.T) {
	var fixture struct {
		CompletionOrder []string
		ValueExamples   []struct {
			Index int
			Text  string
		}
		Commit      string
		Comparisons []struct {
			Left, Right, Lower  string
			Equal               bool
			Order, NaturalOrder int
		}
		Wildcards []struct {
			Pattern, Text          string
			Sensitive, Insensitive bool
		}
		Literals []struct{ Kind, Text, Value string }
		Scanners []struct {
			Pattern, Text string
			Result        int
		}
		TextSources []struct {
			Offset, Length, Start  int
			Pattern                string
			Sensitive, Insensitive bool
		}
	}
	data, err := os.ReadFile("testdata/upstream_328e05d_compatibility.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != "328e05d10d6c73654e416675fc75af7ed0e05ffd" || len(fixture.Comparisons) != 19 || len(fixture.Wildcards) != 13 || len(fixture.Literals) != 31 || len(fixture.Scanners) != 19 || len(fixture.TextSources) != 6 {
		t.Fatal("incomplete or incorrectly pinned source oracle fixture")
	}
	examples := []any{"Iİ", time.Date(2026, 10, 22, 13, 14, 15, 0, time.UTC), TimeSpanValue(937840050001), TimeSpanValue(-1), big.NewRat(5, 4), [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 1234.5, true}
	if len(fixture.ValueExamples) != len(examples) || len(fixture.CompletionOrder) != 7 {
		t.Fatal("missing completion fixture")
	}
	for _, example := range fixture.ValueExamples {
		if got := functionParameterValueExample(nil, examples[example.Index]); got != example.Text {
			t.Errorf("example %d = %q, want %q", example.Index, got, example.Text)
		}
	}
	c := newKustoCompleter(nil, CompletionOptionsDefault, CancellationToken{})
	items := []CompletionItem{}
	for _, name := range []string{"ab", "a-b", "_a", "é", "z", "\ue000", "😀"} {
		items = append(items, NewCompletionItem(CompletionKindColumn, name))
	}
	gotOrder := []string{}
	for _, item := range c.orderCompletionItems(items) {
		gotOrder = append(gotOrder, item.DisplayText)
	}
	if !reflect.DeepEqual(gotOrder, fixture.CompletionOrder) {
		t.Errorf("completion order = %q, want %q", gotOrder, fixture.CompletionOrder)
	}
	for _, test := range fixture.Comparisons {
		t.Run("compare/"+test.Left+"/"+test.Right, func(t *testing.T) {
			if got := ordinalEqualFold(test.Left, test.Right); got != test.Equal {
				t.Errorf("equal = %v, want %v", got, test.Equal)
			}
			if got := sign13(ordinalCompare(test.Left, test.Right, true)); got != test.Order {
				t.Errorf("order = %d, want %d", got, test.Order)
			}
			if got := sign13(StringAndNumberCompareOrdinalIgnoreCase(test.Left+"1", test.Right+"1")); got != test.NaturalOrder {
				t.Errorf("natural order = %d, want %d", got, test.NaturalOrder)
			}
			if got := invariantLower(test.Left); got != test.Lower {
				t.Errorf("lower = %q, want %q", got, test.Lower)
			}
			if got := ValueAreEquivalent(test.Left, test.Right); got != test.Equal {
				t.Errorf("ValueAreEquivalent = %v, want %v", got, test.Equal)
			}
			if got := IsHostName(test.Left, test.Right); got != test.Equal {
				t.Errorf("IsHostName = %v, want %v", got, test.Equal)
			}
			if got := ParseConnectionInfo(test.Left+"=value").GetPart(test.Right) == "value"; got != test.Equal {
				t.Errorf("connection key equality = %v, want %v", got, test.Equal)
			}
			if utf16TextWidth(test.Left) == utf16TextWidth(test.Right) {
				if got := ordinalHasPrefix(test.Left+"!", test.Right); got != test.Equal {
					t.Errorf("prefix equality = %v, want %v", got, test.Equal)
				}
				if got := ordinalHasSuffix("!"+test.Left, test.Right); got != test.Equal {
					t.Errorf("suffix equality = %v, want %v", got, test.Equal)
				}
			}

			symbol := NewColumnSymbol(test.Left, ScalarTypes.String.Type())
			if got := SymbolMatchesName(symbol, test.Right, SymbolMatchColumn, true); got != test.Equal {
				t.Errorf("SymbolMatchesName = %v, want %v", got, test.Equal)
			}
		})
	}
	for _, test := range fixture.Wildcards {
		t.Run("wildcard/"+test.Pattern+"/"+test.Text, func(t *testing.T) {
			if got := KustoFactsMatches(test.Pattern, test.Text); got != test.Sensitive {
				t.Errorf("sensitive = %v, want %v", got, test.Sensitive)
			}
			if got := KustoFactsMatches(test.Pattern, test.Text, true); got != test.Insensitive {
				t.Errorf("insensitive = %v, want %v", got, test.Insensitive)
			}
		})
	}
	kinds := map[string]SyntaxKind{"RealLiteralToken": SyntaxKindRealLiteralToken, "DecimalLiteralToken": SyntaxKindDecimalLiteralToken, "IntLiteralToken": SyntaxKindIntLiteralToken, "LongLiteralToken": SyntaxKindLongLiteralToken, "TimespanLiteralToken": SyntaxKindTimespanLiteralToken, "DateTimeLiteralToken": SyntaxKindDateTimeLiteralToken}
	for _, test := range fixture.Literals {
		t.Run(test.Text, func(t *testing.T) {
			token, err := NewLiteralToken("", test.Text, kinds[test.Kind])
			if err != nil {
				t.Fatal(err)
			}
			value := token.Value()
			switch got := value.(type) {
			case *big.Rat:
				want, ok := new(big.Rat).SetString(test.Value)
				if !ok || got.Cmp(want) != 0 {
					t.Errorf("value = %v, want %s", got, test.Value)
				}
			case time.Time:
				if got.Format("2006-01-02T15:04:05.0000000") != test.Value {
					t.Errorf("value = %v, want %s", got, test.Value)
				}
			case float64:
				want, err := strconv.ParseFloat(test.Value, 64)
				if err != nil {
					t.Fatal(err)
				}
				if got != want && !(math.IsNaN(got) && math.IsNaN(want)) {
					t.Errorf("value = %v, want %s", got, test.Value)
				}
			default:
				if fmt.Sprint(value) != test.Value {
					t.Errorf("value = %v, want %s", value, test.Value)
				}
			}
		})
	}
	for _, test := range fixture.Scanners {
		t.Run("scanner/"+test.Pattern+"/"+test.Text, func(t *testing.T) {
			if got := UTF16Chars(test.Pattern, true).Scan(NewArraySource(UTF16CodeUnits(test.Text)), 0); got != test.Result {
				t.Errorf("scan = %d, want %d", got, test.Result)
			}
			if test.Result >= 0 && test.Pattern != "" {
				for _, source := range []Source[byte]{NewTextSource(test.Text), NewArraySource([]byte(test.Text))} {
					if got := Chars(test.Pattern, true).Scan(source, 0); got != len(test.Text) {
						t.Errorf("byte scan = %d, want %d", got, len(test.Text))
					}
				}
			}
		})
	}
	for i, test := range fixture.TextSources {
		t.Run(fmt.Sprintf("source/%d", i), func(t *testing.T) {
			source := NewTextSource("abc", test.Offset, test.Length)
			if got := source.Matches(test.Start, test.Pattern); got != test.Sensitive {
				t.Errorf("sensitive = %v, want %v", got, test.Sensitive)
			}
			if got := source.Matches(test.Start, test.Pattern, true); got != test.Insensitive {
				t.Errorf("insensitive = %v, want %v", got, test.Insensitive)
			}
			if source.Matches(test.Start, test.Pattern, false) != (test.Sensitive || test.Pattern == "") {
				t.Error("explicit sensitive differs")
			}
			if test.Pattern != "" && !test.Insensitive && MatchText(test.Pattern, true).Scan(source, test.Start) >= 0 {
				t.Error("MatchText consumed past the source boundary")
			}
		})
	}
}

func sign13(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func TestUpstream13ByteCharacterBoundaries(t *testing.T) {
	if got := Char('𐐀', true).Scan(NewTextSource("𐐨"), 0); got != 4 {
		t.Fatalf("supplementary Char consumed %d bytes, want 4", got)
	}
	if got := MatchByte(0x80, true).Scan(NewArraySource([]byte{0x81}), 0); got >= 0 {
		t.Fatal("distinct high bytes matched through invalid UTF-8")
	}
	if got := MatchByte(0xc0, true).Scan(NewArraySource([]byte{0xe0}), 0); got != 1 {
		t.Fatal("byte character ordinal casing lost")
	}
}
