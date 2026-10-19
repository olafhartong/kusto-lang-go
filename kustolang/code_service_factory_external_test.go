package kustolang_test

// Tests for public CodeService factories and concrete service facades ported
// from Kusto.Language.Editor. See third_party/kusto/LICENSE.

import (
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCompletionDisplayText(t *testing.T) {
	table := kustolang.NewTableSymbol("Events")
	if got := kustolang.CompletionDisplayText(table); got != "Events" {
		t.Fatalf("table display = %q", got)
	}
	external := kustolang.NewExternalTableSymbol("Archive")
	if got := kustolang.CompletionDisplayText(external); got != "external_table('Archive')" {
		t.Fatalf("external display = %q", got)
	}
	left := kustolang.NewParameter("left", kustolang.ScalarTypes.Long.Type())
	right := kustolang.NewParameter("right", kustolang.ScalarTypes.Long.Type(), kustolang.WithParameterOccurrence(0, 1))
	signature := kustolang.NewSignature(kustolang.ScalarTypes.Long.Type(), left, right)
	function := kustolang.NewFunctionSymbol("sum2", signature)
	if got := kustolang.CompletionDisplayText(function); got != "sum2(left, [right])" {
		t.Fatalf("function display = %q", got)
	}
	if got := kustolang.CompletionDisplayText(function, true); got != "sum2" {
		t.Fatalf("name-only display = %q", got)
	}
}

func TestPublicConcreteAndOffsetCodeServices(t *testing.T) {
	common := kustolang.NewCommonCodeService("common text")
	if common.Text() != "common text" {
		t.Fatalf("common service text = %q", common.Text())
	}

	text := "// 😀\nT | where ("
	kusto := kustolang.NewKustoCodeService(text)
	if kusto.Kind() != kustolang.CodeKindsQuery {
		t.Fatalf("Kusto kind = %q", kusto.Kind())
	}
	baseDiagnostics := kusto.GetDiagnostics(true, kustolang.CancellationToken{})
	if len(baseDiagnostics) == 0 {
		t.Fatal("malformed KQL produced no diagnostics")
	}

	offset := kustolang.NewOffsetCodeService(kusto.CodeService, 17)
	adjusted := offset.GetDiagnostics(true, kustolang.CancellationToken{})
	if len(adjusted) != len(baseDiagnostics) {
		t.Fatalf("offset diagnostics = %d, want %d", len(adjusted), len(baseDiagnostics))
	}
	for index := range adjusted {
		if adjusted[index].Start != baseDiagnostics[index].Start+17 || adjusted[index].Length != baseDiagnostics[index].Length {
			t.Fatalf("offset diagnostic %d = %#v, base %#v", index, adjusted[index], baseDiagnostics[index])
		}
	}
	if moved := offset.WithOffset(31).GetDiagnostics(true, kustolang.CancellationToken{}); moved[0].Start != baseDiagnostics[0].Start+31 {
		t.Fatalf("WithOffset diagnostic = %#v", moved[0])
	}

	unknown := kustolang.NewUnknownCodeService("not a language")
	if unknown.Kind() != kustolang.CodeKindsUnknown || unknown.Text() != "not a language" {
		t.Fatalf("unknown service = %q/%q", unknown.Kind(), unknown.Text())
	}
	block := kustolang.NewCodeBlock(9, unknown.CodeService)
	if block.Start() != 9 || block.Kind() != kustolang.CodeKindsUnknown {
		t.Fatalf("unknown block = %d/%q", block.Start(), block.Kind())
	}
}

func TestPublicCodeServiceFactoryCompositionAndScript(t *testing.T) {
	kustoFactory := kustolang.NewKustoCodeServiceFactory(kustolang.DefaultGlobalState)
	if service, ok := kustoFactory.TryGetCodeService("T | count"); !ok || service.Kind() != kustolang.CodeKindsQuery {
		t.Fatalf("Kusto factory result = %T/%t", service, ok)
	}
	if _, ok := kustoFactory.TryGetCodeService("select * from T"); ok {
		t.Fatal("Kusto factory accepted SQL-looking text")
	}

	sqlFactory := kustolang.NewCodeServiceFactory("sql", func(text string) (*kustolang.CodeService, bool) {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "select") {
			return nil, false
		}
		return kustolang.NewUnknownCodeService(text).CodeService, true
	})
	combined := kustoFactory.CodeServiceFactory.WithFactory(sqlFactory)
	if service, ok := combined.TryGetCodeService("select * from T"); !ok || service.Kind() != kustolang.CodeKindsUnknown {
		t.Fatalf("combined SQL service = %T/%t", service, ok)
	}
	if found := combined.GetFactory(func(factory *kustolang.CodeServiceFactory) bool { return factory.Key() == "sql" }); found == nil {
		t.Fatal("combined factory did not expose custom SQL factory")
	}

	script := kustolang.NewCodeScriptWithFactory("select * from T", combined)
	blocks := script.Blocks()
	if len(blocks) != 1 || blocks[0].Kind() != kustolang.CodeKindsUnknown {
		t.Fatalf("custom factory script blocks = %#v", blocks)
	}

	replacement := kustolang.NewCodeServiceFactory("sql", func(text string) (*kustolang.CodeService, bool) {
		return kustolang.NewUnknownCodeService("replacement:" + text).CodeService, true
	})
	replaced := combined.WithFactory(replacement)
	if service, ok := replaced.TryGetCodeService("select 1"); !ok || service.Text() != "replacement:select 1" {
		t.Fatalf("replacement factory service = %T/%t", service, ok)
	}
}

func TestPublicCodeServiceFactoryRetrievesConcreteKustoFactory(t *testing.T) {
	initial := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("initial"))
	kustoFactory := kustolang.NewKustoCodeServiceFactory(initial)

	direct := kustoFactory.CodeServiceFactory.GetKustoCodeServiceFactory()
	if direct == nil || direct.Globals().Database.SymbolName() != "initial" {
		t.Fatalf("direct Kusto factory = %#v", direct)
	}
	if direct.WithGlobals(direct.Globals()) != direct {
		t.Fatal("WithGlobals should preserve the retrieved factory for the same globals")
	}

	customFactory := kustolang.NewCodeServiceFactory("custom", func(text string) (*kustolang.CodeService, bool) {
		return kustolang.NewUnknownCodeService(text).CodeService, true
	})
	combined := customFactory.WithFactory(kustoFactory.CodeServiceFactory)
	retrieved := combined.GetKustoCodeServiceFactory()
	if retrieved == nil || retrieved.Globals().Database.SymbolName() != "initial" {
		t.Fatalf("aggregate Kusto factory = %#v", retrieved)
	}
	legacy := combined.GetFactory(func(factory *kustolang.CodeServiceFactory) bool {
		return factory.Key() == "KustoCodeServiceFactory"
	})
	if legacy == nil || legacy.GetKustoCodeServiceFactory() == nil {
		t.Fatal("GetFactory result did not recover the concrete Kusto factory")
	}

	updatedGlobals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("updated"))
	updated := retrieved.WithGlobals(updatedGlobals)
	if updated == retrieved || updated.Globals().Database.SymbolName() != "updated" {
		t.Fatalf("WithGlobals result = same:%t globals:%#v", updated == retrieved, updated.Globals())
	}
	if retrieved.Globals().Database.SymbolName() != "initial" {
		t.Fatal("WithGlobals mutated the retrieved factory")
	}
	if found := combined.GetFactory(func(factory *kustolang.CodeServiceFactory) bool {
		return factory.Key() == "custom"
	}); found == nil {
		t.Fatal("concrete retrieval changed the aggregate")
	}

	if got := customFactory.GetKustoCodeServiceFactory(); got != nil {
		t.Fatalf("custom-only Kusto factory = %#v, want nil", got)
	}
	var nilFactory *kustolang.CodeServiceFactory
	if got := nilFactory.GetKustoCodeServiceFactory(); got != nil {
		t.Fatalf("nil Kusto factory = %#v, want nil", got)
	}
}

func TestPublicCodeScriptGlobalsPreserveComposedFactories(t *testing.T) {
	initialGlobals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("initial"))
	kustoFactory := kustolang.NewKustoCodeServiceFactory(initialGlobals)
	customFactory := kustolang.NewCodeServiceFactory("custom", func(text string) (*kustolang.CodeService, bool) {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "select") {
			return nil, false
		}
		return kustolang.NewUnknownCodeService("custom:" + text).CodeService, true
	})
	combined := kustoFactory.CodeServiceFactory.WithFactory(customFactory)
	script := kustolang.NewCodeScriptWithFactory("select 1", combined)

	if globals := script.Globals(); globals == nil || globals.Database.SymbolName() != "initial" {
		t.Fatalf("composed script globals = %#v", globals)
	}

	updatedGlobals := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("updated"))
	updated := script.WithGlobals(updatedGlobals)
	if globals := updated.Globals(); globals == nil || globals.Database.SymbolName() != "updated" {
		t.Fatalf("updated composed script globals = %#v", globals)
	}
	if found := updated.Factory().GetFactory(func(factory *kustolang.CodeServiceFactory) bool {
		return factory.Key() == "custom"
	}); found == nil {
		t.Fatal("WithGlobals discarded sibling custom factory")
	}
	if blocks := updated.Blocks(); len(blocks) != 1 || blocks[0].Text() != "custom:select 1" {
		t.Fatalf("updated composed script blocks = %#v", blocks)
	}

	customOnly := kustolang.NewCodeScriptWithFactory("select 1", customFactory)
	if got := customOnly.WithGlobals(updatedGlobals); got != customOnly {
		t.Fatal("WithGlobals should not add a Kusto factory when none is composed")
	}
}
