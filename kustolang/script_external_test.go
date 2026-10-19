package kustolang_test

import (
	"testing"
	"unicode/utf16"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func TestPublicCodeScriptRoutesUTF16BlocksAndMalformedDiagnostics(t *testing.T) {
	text := "// 😀\n\nEvents | where Account ==\n\nEvents | project Account"
	script, err := kustolang.NewCodeScriptForEnvironment(text, nil, environment.LocalData{
		Tables: map[string]environment.TableDetails{"Events": {"Account": "string"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	blocks := script.Blocks()
	if len(blocks) != 3 {
		t.Fatalf("blocks = %#v", blocks)
	}
	secondStart := utf16Length("// 😀\n\n")
	if blocks[1].Start() != secondStart || script.GetBlockAtPosition(secondStart) == nil {
		t.Fatalf("second block start = %d, want %d", blocks[1].Start(), secondStart)
	}
	diagnostics := blocks[1].Service().GetDiagnostics(true, kustolang.CancellationToken{})
	if !hasDiagnostic(diagnostics, "KS006") || diagnostics[0].Start < blocks[1].Start() {
		t.Fatalf("offset malformed diagnostics = %#v", diagnostics)
	}
	if valid := blocks[2].Service().GetDiagnostics(true, kustolang.CancellationToken{}); len(valid) != 0 {
		t.Fatalf("valid third block diagnostics = %#v", valid)
	}
	line, offset, ok := script.TryGetLineAndOffset(blocks[2].Start())
	if !ok || line != 5 || offset != 1 {
		t.Fatalf("third block line/offset = %d/%d/%t", line, offset, ok)
	}
}

func TestPublicCodeScriptDistinguishesSeparatorClosures(t *testing.T) {
	separatorAt := func(position int) kustolang.BlockSeparator {
		return func(string, []int) []int { return []int{0, position} }
	}

	script := kustolang.NewCodeScript("abcde").WithSeparator(separatorAt(2))
	updated := script.WithSeparator(separatorAt(4))
	blocks := updated.Blocks()
	if len(blocks) != 2 || blocks[0].Text() != "abcd" || blocks[1].Text() != "e" {
		t.Fatalf("distinct public separator blocks = %#v", blocks)
	}
}

func TestPublicCodeScriptSameSeparatorHandleReturnsExistingScript(t *testing.T) {
	calls := 0
	handle := kustolang.NewBlockSeparatorHandle(func(string, []int) []int {
		calls++
		return []int{0, 2}
	})

	script := kustolang.NewCodeScript("abcde").WithSeparatorHandle(handle)
	if calls != 1 {
		t.Fatalf("separator calls after initial script = %d, want 1", calls)
	}
	if script.SeparatorHandle() != handle {
		t.Fatal("script did not preserve separator handle identity")
	}
	if updated := script.WithSeparatorHandle(handle); updated != script {
		t.Fatal("same separator handle should return the existing script")
	}
	if updated := script.WithSeparatorHandle(script.SeparatorHandle()); updated != script {
		t.Fatal("separator handle returned by the script should preserve identity")
	}
	if calls != 1 {
		t.Fatalf("separator calls after no-op updates = %d, want 1", calls)
	}
}

func TestPublicCodeScriptSeparatorReplacementReusesBlockServices(t *testing.T) {
	factoryCalls := 0
	factory := kustolang.NewCodeServiceFactory("counting", func(text string) (*kustolang.CodeService, bool) {
		factoryCalls++
		return kustolang.NewCustomCodeService(text, kustolang.CustomCodeServiceCallbacks{}), true
	})
	separatorAtTwo := func() kustolang.BlockSeparator {
		return func(string, []int) []int { return []int{0, 2} }
	}

	script := kustolang.NewCodeScriptWithFactory("abcde", factory).WithSeparator(separatorAtTwo())
	callsBeforeReplacement := factoryCalls
	updated := script.WithSeparator(separatorAtTwo())
	if updated == script {
		t.Fatal("distinct separator callbacks should create a new script")
	}
	if factoryCalls != callsBeforeReplacement {
		t.Fatalf("factory calls after equivalent separator replacement = %d, want %d", factoryCalls, callsBeforeReplacement)
	}
	blocks := updated.Blocks()
	if len(blocks) != 2 || blocks[0].Text() != "ab" || blocks[1].Text() != "cde" {
		t.Fatalf("replacement blocks = %#v", blocks)
	}
}

func TestPublicCombinedDiagnosticsFiltering(t *testing.T) {
	service := newEventsService(t, `Events | where Account contains "failure"`)
	combined := service.GetCombinedDiagnostics(true, nil, kustolang.CancellationToken{})
	if !hasDiagnostic(combined, "KS500") {
		t.Fatalf("combined diagnostics = %#v", combined)
	}
	disabled := kustolang.ParseDisabledDiagnostics("KS500")
	filtered := service.GetCombinedDiagnostics(true, &disabled, kustolang.CancellationToken{})
	if hasDiagnostic(filtered, "KS500") {
		t.Fatalf("filtered diagnostics = %#v", filtered)
	}
}

func utf16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}
