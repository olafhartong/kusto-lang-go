package kustolang

import "testing"

func TestTokenParserDirectiveScannerHelpers(t *testing.T) {
	if got := ScanIdentifier("$name_1 rest", 0); got != len("$name_1") {
		t.Fatalf("ScanIdentifier($name_1) = %d", got)
	}
	if got := ScanIdentifier("123abc$", 0); got != len("123abc") {
		t.Fatalf("ScanIdentifier(123abc$) = %d, want %d", got, len("123abc"))
	}
	if got := ScanIdentifier("123 ", 0); got != -1 {
		t.Fatalf("ScanIdentifier(123) = %d, want -1", got)
	}
	if got := ScanLongLiteral("0x10 ", 0); got != len("0x10") {
		t.Fatalf("ScanLongLiteral(0x10) = %d", got)
	}
	if got := ScanLongLiteral("42name", 0); got != -1 {
		t.Fatalf("ScanLongLiteral(42name) = %d, want -1", got)
	}
	if got := ScanRealLiteral("1.5e2;", 0); got != len("1.5e2") {
		t.Fatalf("ScanRealLiteral(1.5e2) = %d", got)
	}
	if got := ScanRealLiteral("1..2", 0); got != -1 {
		t.Fatalf("ScanRealLiteral(1..2) = %d, want -1", got)
	}
}

func TestTryParseClientDirectiveParsesArgumentsAndTextSlices(t *testing.T) {
	text := NewEditString("#set name=\"prod\", count=42; ratio=1.5 bare missing=, 0x10\nT | count")
	directive, ok := TryParseClientDirective(text)
	if !ok {
		t.Fatal("TryParseClientDirective returned false")
	}
	if directive.Text.CurrentText != text.CurrentText || directive.Name != "set" {
		t.Fatalf("directive text/name = text:%q name:%q", directive.Text.CurrentText, directive.Name)
	}
	if got, want := directive.AfterNameText.CurrentText, " name=\"prod\", count=42; ratio=1.5 bare missing=, 0x10\nT | count"; got != want {
		t.Fatalf("AfterNameText = %q, want %q", got, want)
	}
	if got, want := directive.ArgumentsText.CurrentText, "name=\"prod\", count=42; ratio=1.5 bare missing=, 0x10"; got != want {
		t.Fatalf("ArgumentsText = %q, want %q", got, want)
	}
	if got, want := directive.AfterDirectiveText.CurrentText, "T | count"; got != want {
		t.Fatalf("AfterDirectiveText = %q, want %q", got, want)
	}

	if len(directive.Arguments) != 6 {
		t.Fatalf("argument count = %d, want 6: %#v", len(directive.Arguments), directive.Arguments)
	}
	assertDirectiveArg(t, directive.Arguments[0], "name", "\"prod\"", "prod")
	assertDirectiveArg(t, directive.Arguments[1], "count", "42", int64(42))
	assertDirectiveArg(t, directive.Arguments[2], "ratio", "1.5", float64(1.5))
	assertDirectiveArg(t, directive.Arguments[3], "", "bare", "bare")
	assertDirectiveArg(t, directive.Arguments[4], "missing", "", nil)
	assertDirectiveArg(t, directive.Arguments[5], "", "0x10", int64(0))
}

func TestTryParseClientDirectiveSkipsTriviaAndHandlesUnclosedString(t *testing.T) {
	text := NewEditString(" // comment\n  #client arg='unterminated")
	directive, ok := TryParseClientDirective(text)
	if !ok {
		t.Fatal("TryParseClientDirective with trivia returned false")
	}
	if directive.Name != "client" || len(directive.Arguments) != 1 {
		t.Fatalf("directive = name:%q args:%#v", directive.Name, directive.Arguments)
	}
	assertDirectiveArg(t, directive.Arguments[0], "arg", "'unterminated", "unterminated")
	if directive.AfterDirectiveText.CurrentText != "" {
		t.Fatalf("AfterDirectiveText = %q, want empty", directive.AfterDirectiveText.CurrentText)
	}

	if _, ok := TryParseClientDirective(NewEditString("T | count")); ok {
		t.Fatal("non-directive text should not parse as a client directive")
	}
}

func TestTryParseClientDirectiveSeparatesUnicodeWhitespaceArguments(t *testing.T) {
	directive, ok := TryParseClientDirective(NewEditString("#set\u2003one=1\u2003two"))
	if !ok || len(directive.Arguments) != 2 {
		t.Fatalf("Unicode directive = ok:%t args:%#v", ok, directive.Arguments)
	}
	assertDirectiveArg(t, directive.Arguments[0], "one", "1", int64(1))
	assertDirectiveArg(t, directive.Arguments[1], "", "two", "two")
}

func TestParseClientDirectiveInfoCollectsLeadingDirectives(t *testing.T) {
	text := NewEditString("#one value=1\n  #two \"second\"\n\nT | count\n")
	info := ParseClientDirectiveInfo(text)
	if info.Text.CurrentText != text.CurrentText {
		t.Fatalf("info text = %q", info.Text.CurrentText)
	}
	if len(info.Directives) != 2 {
		t.Fatalf("directive count = %d, want 2: %#v", len(info.Directives), info.Directives)
	}
	if info.Directives[0].Name != "one" || info.Directives[1].Name != "two" {
		t.Fatalf("directive names = %q, %q", info.Directives[0].Name, info.Directives[1].Name)
	}
	assertDirectiveArg(t, info.Directives[0].Arguments[0], "value", "1", int64(1))
	assertDirectiveArg(t, info.Directives[1].Arguments[0], "", "\"second\"", "second")
	if got, want := info.RemainingText.CurrentText, "T | count"; got != want {
		t.Fatalf("remaining text = %q, want %q", got, want)
	}

	emptyRemaining := ParseClientDirectiveInfo(NewEditString("#only\n"))
	if len(emptyRemaining.Directives) != 1 || emptyRemaining.RemainingText.CurrentText != "" {
		t.Fatalf("last-line directive info = directives:%#v remaining:%q", emptyRemaining.Directives, emptyRemaining.RemainingText.CurrentText)
	}
}

func assertDirectiveArg(t *testing.T, got ClientDirectiveArgument, wantName string, wantText string, wantValue any) {
	t.Helper()
	if got.Name != wantName || got.Text != wantText || got.Value != wantValue {
		t.Fatalf("directive argument = %#v, want name:%q text:%q value:%#v", got, wantName, wantText, wantValue)
	}
}
