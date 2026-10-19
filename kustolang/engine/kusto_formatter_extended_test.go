package kustolang

import "testing"

func TestKustoFormatterDataTableSchemaAndRows(t *testing.T) {
	input := `datatable(alpha:string,beta:long,gamma:datetime,delta:bool,epsilon:real)["one",1,datetime(2024-01-01),true,1.5,"two",2,datetime(2024-01-02),false,2.5]`
	want := `datatable
(
    alpha: string,
    beta: long,
    gamma: datetime,
    delta: bool,
    epsilon: real
)
[
    "one", 1, datetime(2024-01-01), true, 1.5,
    "two", 2, datetime(2024-01-02), false, 2.5
]`
	got := formatKustoForTest(input, FormattingOptionsDefault, len(input))
	if got.Text != want {
		t.Fatalf("formatted datatable mismatch\n got: %q\nwant: %q", got.Text, want)
	}
	if got.Position != len(want) {
		t.Fatalf("caret = %d, want %d", got.Position, len(want))
	}
}

func TestKustoFormatterDataTableBalancesColumnsAndPreservesComments(t *testing.T) {
	balancedInput := `datatable(value:string)[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30]`
	balancedWant := `datatable (value: string)
[
    1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
    11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
    21, 22, 23, 24, 25, 26, 27, 28, 29, 30
]`
	if got := formatKustoForTest(balancedInput, FormattingOptionsDefault, 0); got.Text != balancedWant {
		t.Fatalf("balanced values mismatch\n got: %q\nwant: %q", got.Text, balancedWant)
	}

	commentInput := "datatable(value:string)[\"this value is deliberately wider than forty characters\", // keep\n\"short\"]"
	commentWant := `datatable (value: string)
[
    "this value is deliberately wider than forty characters", // keep
    "short"
]`
	if got := formatKustoForTest(commentInput, FormattingOptionsDefault, 0); got.Text != commentWant {
		t.Fatalf("wide commented values mismatch\n got: %q\nwant: %q", got.Text, commentWant)
	}
}

func TestKustoFormatterSplitsWideDirectBinaryChains(t *testing.T) {
	input := `T | where alpha_identifier == 1111111111 and beta_identifier == 2222222222 and gamma_identifier == 3333333333`
	want := `T
| where alpha_identifier == 1111111111
    and beta_identifier == 2222222222
    and gamma_identifier == 3333333333`
	if got := formatKustoForTest(input, FormattingOptionsDefault, 0); got.Text != want {
		t.Fatalf("binary chain mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterWrapsDirectQueryOperatorCommaLists(t *testing.T) {
	input := `T | project first_really_long_column_name, second_really_long_column_name, third_really_long_column_name`
	want := `T
| project
    first_really_long_column_name,
    second_really_long_column_name,
    third_really_long_column_name`
	if got := formatKustoForTest(input, FormattingOptionsDefault, 0); got.Text != want {
		t.Fatalf("query operator list mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterAlignsQueryOperatorClausesAndSubElements(t *testing.T) {
	input := "T | summarize\ncount()\nby very_long_grouping_column, another_grouping_column"
	want := `T
| summarize
    count()
    by very_long_grouping_column, another_grouping_column`
	options := FormattingOptionsDefault.WithIndentationSize(4)
	if got := formatKustoForTest(input, options, 0); got.Text != want {
		t.Fatalf("clause alignment mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}
