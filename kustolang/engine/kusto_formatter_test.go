package kustolang

import "testing"

func formatKustoForTest(text string, options FormattingOptions, caret int) FormattedText {
	return FormatKustoCode(Parse(text), options, caret)
}

func TestKustoFormatterDefaultStructureAndLiterals(t *testing.T) {
	input := `let threshold=10;T|where Message=="a  b"|project A,B`
	options := FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine)
	got := formatKustoForTest(input, options, len(input))
	want := "let threshold = 10;\nT\n| where Message == \"a  b\"\n| project A, B"
	if got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
	if got.Position != len(want) {
		t.Fatalf("caret = %d, want %d", got.Position, len(want))
	}
}

func TestKustoFormatterPreservesAndIndentsComments(t *testing.T) {
	input := "T // source\n | where x==1 // predicate\n | project x"
	got := formatKustoForTest(input, FormattingOptionsDefault, 0)
	want := "T // source\n| where x == 1 // predicate\n| project x"
	if got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterSpacingOptions(t *testing.T) {
	options := FormattingOptionsDefault.
		WithCommaSpacing(DualSpacingStyleBoth).
		WithAssignmentSpacing(DualSpacingStyleNeither).
		WithInfixOperatorSpacing(DualSpacingStyleNeither).
		WithArgumentListSpacing(SpacingStyleOne).
		WithBeforeArgumentListSpacing(SpacingStyleOne)
	got := formatKustoForTest(`print f(1,2),x=1+2`, options, 0)
	want := `print f ( 1 , 2 ) , x=1+2`
	if got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterPipeAndStatementPlacementOptions(t *testing.T) {
	options := FormattingOptionsDefault.
		WithPipeOperatorStyle(PlacementStyleNone).
		WithStatementStyle(PlacementStyleNone)
	got := formatKustoForTest(`let x=1;T|take x`, options, 0)
	want := `let x = 1; T | take x`
	if got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterSmartStatementPlacementUsesEnclosingList(t *testing.T) {
	input := "let a=1; let b=2; T\n| take 1"
	tests := []struct {
		name  string
		style PlacementStyle
		want  string
	}{
		{
			name:  "smart",
			style: PlacementStyleSmart,
			want:  "let a = 1;\nlet b = 2;\nT\n| take 1",
		},
		{
			name:  "none",
			style: PlacementStyleNone,
			want:  "let a = 1; let b = 2; T\n| take 1",
		},
		{
			name:  "new line",
			style: PlacementStyleNewLine,
			want:  "let a = 1;\nlet b = 2;\nT\n| take 1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := FormattingOptionsDefault.WithStatementStyle(test.style)
			if got := formatKustoForTest(input, options, 0); got.Text != test.want {
				t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, test.want)
			}
		})
	}
}

func TestKustoFormatterSemicolonPlacementOptions(t *testing.T) {
	input := "set notruncation; let y=T\n|take 1; T"
	tests := []struct {
		name  string
		style PlacementStyle
		want  string
	}{
		{
			name:  "none",
			style: PlacementStyleNone,
			want:  "set notruncation;\nlet y = T\n    | take 1;\nT",
		},
		{
			name:  "new line",
			style: PlacementStyleNewLine,
			want:  "set notruncation;\nlet y = T\n    | take 1\n    ;\nT",
		},
		{
			name:  "smart",
			style: PlacementStyleSmart,
			want:  "set notruncation;\nlet y = T\n    | take 1\n    ;\nT",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := FormattingOptionsDefault.WithSemicolonStyle(test.style)
			if got := formatKustoForTest(input, options, 0); got.Text != test.want {
				t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, test.want)
			}
		})
	}
}

func TestKustoFormatterSemicolonPlacementHandlesMissingSeparator(t *testing.T) {
	input := "set notruncation\ndeclare query_parameters(x:long)\nT"
	options := FormattingOptionsDefault.
		WithSemicolonStyle(PlacementStyleNewLine).
		WithInsertMissingTokens(true)
	want := "set notruncation;\ndeclare query_parameters (x: long);\nT"
	if got := formatKustoForTest(input, options, 0); got.Text != want {
		t.Fatalf("formatted malformed query mismatch\n got: %q\nwant: %q", got.Text, want)
	}
	if diagnostics := ParseAndAnalyze(input).GetDiagnostics(); !diagnosticsContain(diagnostics, "KS005") {
		t.Fatalf("malformed query diagnostics = %#v, want KS005", diagnostics)
	}
}

func TestKustoFormatterSmartPipeNormalizesBarAndParenthesizedQuery(t *testing.T) {
	input := "let x=(T |\n where value==1); x"
	want := "let x = (\n    T\n    | where value == 1);\nx"
	if got := formatKustoForTest(input, FormattingOptionsDefault, 0); got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterSmartPipeHandlesUnclosedParenthesizedQuery(t *testing.T) {
	input := "(T |\n where value=="
	options := FormattingOptionsDefault.WithInsertMissingTokens(false)
	want := "(\nT\n| where value =="
	if got := formatKustoForTest(input, options, 0); got.Text != want {
		t.Fatalf("formatted malformed query mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoFormatterCaretMapsInsideTokenAndAfterRoot(t *testing.T) {
	input := `T|where abc==1`
	options := FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine)
	inside := formatKustoForTest(input, options, 10)
	if inside.Text != "T\n| where abc == 1" || inside.Position != 12 {
		t.Fatalf("inside caret result = %#v", inside)
	}
	after := formatKustoForTest(input, options, len(input)+5)
	if after.Position != len(after.Text)+5 {
		t.Fatalf("after caret = %d, want %d", after.Position, len(after.Text)+5)
	}
}

func TestKustoFormatterMalformedRecoveryAndMissingTokenOption(t *testing.T) {
	input := `T|where (x==1`
	withoutOptions := FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine).WithInsertMissingTokens(false)
	withOptions := FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine).WithInsertMissingTokens(true)
	without := formatKustoForTest(input, withoutOptions, len(input))
	with := formatKustoForTest(input, withOptions, len(input))
	if without.Text != "T\n| where (x == 1" {
		t.Fatalf("without insertion = %q", without.Text)
	}
	if with.Text != "T\n| where (x == 1)" {
		t.Fatalf("with insertion = %q", with.Text)
	}
}

func TestKustoFormatterMultilineFunctionCallUsesDiagonalLayout(t *testing.T) {
	input := "print strcat(\n'a',\n'b'\n)"
	got := formatKustoForTest(input, FormattingOptionsDefault, 0)
	want := "print strcat(\n          'a',\n          'b'\n      )"
	if got.Text != want {
		t.Fatalf("formatted text mismatch\n got: %q\nwant: %q", got.Text, want)
	}
}

func TestKustoCodeServiceReportsAndUsesFormattingCapability(t *testing.T) {
	service := NewKustoCodeService(`T|where x==1`)
	options := FormattingOptionsDefault.WithPipeOperatorStyle(PlacementStyleNewLine)
	formatted := service.GetFormattedText(options, len(service.Text()), CancellationToken{})
	if formatted.Text != "T\n| where x == 1" || formatted.Position != len(formatted.Text) {
		t.Fatalf("service formatted text = %#v", formatted)
	}
	if !service.IsFeatureSupported(CodeServiceFeatureFormatting, 0) {
		t.Fatal("implemented formatting capability is not advertised")
	}

	canceled := NewCancellationTokenFunc(func() bool { return true })
	unchanged := service.GetFormattedText(FormattingOptionsDefault, 3, canceled)
	if unchanged.Text != service.Text() || unchanged.Position != 3 {
		t.Fatalf("canceled formatting = %#v", unchanged)
	}
}
