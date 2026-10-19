package kustolang

import (
	"strings"
	"testing"
)

func TestCodeBlockWrapsServiceWithOffset(t *testing.T) {
	text := "print x={tenant}"
	block := NewCodeBlock(20, NewUnknownCodeService(text))

	if block.Text() != text || block.Start != 20 || block.Length() != len(text) || block.End() != 20+len(text) || block.Kind() != CodeKindsUnknown {
		t.Fatalf("block = text:%q start:%d length:%d end:%d kind:%q", block.Text(), block.Start, block.Length(), block.End(), block.Kind())
	}
	parameters := block.Service().GetClientParameters()
	if len(parameters) != 1 || parameters[0].Start() != 20+strings.Index(text, "{tenant}") {
		t.Fatalf("offset parameters = %#v", parameters)
	}

	moved := block.WithStart(5)
	parameters = moved.Service().GetClientParameters()
	if moved.Start != 5 || len(parameters) != 1 || parameters[0].Start() != 5+strings.Index(text, "{tenant}") {
		t.Fatalf("moved block = start:%d parameters:%#v", moved.Start, parameters)
	}
}

func TestCodeServiceFactoryAggregateAndReplacement(t *testing.T) {
	first := NewCodeServiceFactoryFunc("first", func(text string) (CodeService, bool) {
		return nil, false
	})
	second := NewCodeServiceFactoryFunc("second", func(text string) (CodeService, bool) {
		return NewUnknownCodeService("second:" + text), true
	})

	combined := first.WithFactory(second)
	service, ok := combined.TryGetCodeService("query")
	if !ok || service.Text() != "second:query" {
		t.Fatalf("combined service = %v %#v", ok, service)
	}
	if got := combined.GetFactory(func(factory CodeServiceFactory) bool { return factory.factoryKey() == "second" }); got != second {
		t.Fatalf("GetFactory(second) = %#v, want second", got)
	}

	replacement := NewCodeServiceFactoryFunc("second", func(text string) (CodeService, bool) {
		return NewUnknownCodeService("replacement:" + text), true
	})
	combined = combined.WithFactory(replacement)
	service, ok = combined.TryGetCodeService("query")
	if !ok || service.Text() != "replacement:query" {
		t.Fatalf("replacement service = %v %#v", ok, service)
	}
}

func TestCodeScriptSplitsBlocksPositionsAndReusesUnchangedText(t *testing.T) {
	calls := 0
	factory := NewCodeServiceFactoryFunc("counting", func(text string) (CodeService, bool) {
		calls++
		return NewUnknownCodeService(text), true
	})
	text := "T | count\n\nU | count"
	secondStart := strings.Index(text, "U | count")
	script := NewCodeScript(text, factory)

	blocks := script.Blocks()
	if len(blocks) != 2 || blocks[0].Start != 0 || blocks[0].Text() != text[:secondStart] || blocks[1].Start != secondStart || blocks[1].Text() != text[secondStart:] {
		t.Fatalf("blocks = %#v", blocks)
	}
	if calls != 2 {
		t.Fatalf("factory calls = %d, want 2", calls)
	}
	if got := script.GetBlockAtPosition(0); got != blocks[0] {
		t.Fatalf("block at 0 = %#v, want first", got)
	}
	if got := script.GetBlockAtPosition(secondStart); got != blocks[1] {
		t.Fatalf("block at second start = %#v, want second", got)
	}
	if got := script.GetBlockAtPosition(len(text)); got != blocks[1] {
		t.Fatalf("block at end = %#v, want second", got)
	}

	if position, ok := script.TryGetTextPosition(1, 10); !ok || position != len("T | count") {
		t.Fatalf("line 1 offset 10 = %d, %v", position, ok)
	}
	if _, ok := script.TryGetTextPosition(1, 11); ok {
		t.Fatal("line 1 offset 11 should be outside virtual end-of-line")
	}
	if line, offset, ok := script.TryGetLineAndOffset(secondStart); !ok || line != 3 || offset != 1 {
		t.Fatalf("line/offset at second block = %d/%d %v", line, offset, ok)
	}

	updated := script.WithText("T | count\n\nV | count")
	if calls != 3 {
		t.Fatalf("factory calls after unchanged first block reuse = %d, want 3", calls)
	}
	if len(updated.Blocks()) != 2 || updated.Blocks()[0].Text() != blocks[0].Text() || updated.Blocks()[1].Text() != "V | count" {
		t.Fatalf("updated blocks = %#v", updated.Blocks())
	}
}

func TestCodeScriptCustomSeparatorAndUnknownFallback(t *testing.T) {
	failingFactory := NewCodeServiceFactoryFunc("failing", func(text string) (CodeService, bool) {
		return nil, false
	})
	separator := func(text string, lineStarts []int) []int {
		return []int{0, 3}
	}

	script := NewCodeScript("abcde", failingFactory).WithSeparator(separator)
	blocks := script.Blocks()
	if len(blocks) != 2 || blocks[0].Text() != "abc" || blocks[1].Text() != "de" {
		t.Fatalf("custom separated blocks = %#v", blocks)
	}
	if blocks[0].Kind() != CodeKindsUnknown || blocks[1].Kind() != CodeKindsUnknown {
		t.Fatalf("fallback block kinds = %q/%q, want unknown", blocks[0].Kind(), blocks[1].Kind())
	}
	if script.WithSeparator(separator) != script {
		t.Fatal("same separator should return existing script")
	}
}

func TestCodeScriptUsesUTF16PositionsAcrossNonBMPBlocks(t *testing.T) {
	text := "// 😀\n\nprint x={tenant}"
	secondByteStart := strings.Index(text, "print")
	secondStart := byteOffsetToUTF16Position(text, secondByteStart)
	textLength := byteOffsetToUTF16Position(text, len(text))
	script := NewCodeScript(text, NewUnknownCodeServiceFactory())

	blocks := script.Blocks()
	if len(blocks) != 2 || blocks[1].Start != secondStart || blocks[1].Text() != text[secondByteStart:] {
		t.Fatalf("UTF-16 blocks = %#v, want second start %d", blocks, secondStart)
	}
	if blocks[1].Length() != byteOffsetToUTF16Position(blocks[1].Text(), len(blocks[1].Text())) || blocks[1].End() != textLength {
		t.Fatalf("UTF-16 second block = start:%d length:%d end:%d, want end %d", blocks[1].Start, blocks[1].Length(), blocks[1].End(), textLength)
	}
	if script.GetBlockAtPosition(secondStart) != blocks[1] || script.GetBlockAtPosition(textLength) != blocks[1] {
		t.Fatalf("UTF-16 block lookup failed at %d or %d", secondStart, textLength)
	}
	parameters := blocks[1].Service().GetClientParameters()
	wantParameterStart := secondStart + strings.Index(blocks[1].Text(), "{tenant}")
	if len(parameters) != 1 || parameters[0].Start() != wantParameterStart {
		t.Fatalf("UTF-16 second-block parameters = %#v, want start %d", parameters, wantParameterStart)
	}
	if position, ok := script.TryGetTextPosition(1, 6); !ok || position != 5 {
		t.Fatalf("UTF-16 first-line virtual end = %d/%t, want 5/true", position, ok)
	}
	if _, ok := script.TryGetTextPosition(1, 7); ok {
		t.Fatal("UTF-16 first-line offset beyond virtual end should fail")
	}
	if line, offset, ok := script.TryGetLineAndOffset(secondStart); !ok || line != 3 || offset != 1 {
		t.Fatalf("UTF-16 second-block line/offset = %d/%d/%t", line, offset, ok)
	}

	kustoScript := NewCodeScript("// 😀\n\nprint", NewKustoCodeServiceFactory(DefaultGlobalState))
	diagnostics := kustoScript.Blocks()[1].Service().GetDiagnostics(true, CancellationToken{})
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS006" || diagnostics[0].Start != secondStart+len("print") {
		t.Fatalf("UTF-16 second-block diagnostics = %#v, want KS006 at %d", diagnostics, secondStart+len("print"))
	}
}

func TestCodeScriptTryGetTextPositionExcludesUnicodeLineSeparators(t *testing.T) {
	script := NewCodeScript("a\u2028b\u2029c\r\nd", NewUnknownCodeServiceFactory())
	tests := []struct {
		line       int
		lineOffset int
		position   int
		ok         bool
	}{
		{line: 1, lineOffset: 2, position: 1, ok: true},
		{line: 1, lineOffset: 3, ok: false},
		{line: 2, lineOffset: 2, position: 3, ok: true},
		{line: 2, lineOffset: 3, ok: false},
		{line: 3, lineOffset: 2, position: 5, ok: true},
		{line: 3, lineOffset: 3, ok: false},
		{line: 4, lineOffset: 2, position: 8, ok: true},
	}
	for _, test := range tests {
		position, ok := script.TryGetTextPosition(test.line, test.lineOffset)
		if ok != test.ok || ok && position != test.position {
			t.Fatalf("line %d offset %d = %d/%t, want %d/%t",
				test.line, test.lineOffset, position, ok, test.position, test.ok)
		}
	}
}

func TestCodeScriptCustomSeparatorUsesUTF16Positions(t *testing.T) {
	text := "😀abc"
	separator := func(gotText string, lineStarts []int) []int {
		if gotText != text || len(lineStarts) != 1 || lineStarts[0] != 0 {
			t.Fatalf("separator input = %q %#v", gotText, lineStarts)
		}
		return []int{0, 2}
	}
	script := NewCodeScript(text, NewUnknownCodeServiceFactory()).WithSeparator(separator)
	blocks := script.Blocks()
	if len(blocks) != 2 || blocks[0].Text() != "😀" || blocks[1].Text() != "abc" || blocks[1].Start != 2 {
		t.Fatalf("UTF-16 custom-separated blocks = %#v", blocks)
	}
}
