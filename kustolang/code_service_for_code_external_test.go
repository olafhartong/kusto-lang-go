package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicKustoCodeServiceUsesExistingAnalyzedCode(t *testing.T) {
	code := kustolang.ParseAndAnalyze(`datatable(A:long)[1] | project Missing, B=`)
	service := kustolang.NewKustoCodeServiceForCode(code)
	if service.Text() != code.Text || service.Kind() != kustolang.CodeKindsQuery {
		t.Fatalf("service = text:%q kind:%q", service.Text(), service.Kind())
	}
	got := service.GetDiagnostics(true, kustolang.CancellationToken{})
	want := code.GetDiagnostics()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
}

func TestPublicCodeScriptFactoryGlobalsAndImmutableReplacement(t *testing.T) {
	first := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("first", kustolang.NewTableSymbol("T", kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.Long.Type()))))
	second := kustolang.NewGlobalState(kustolang.NewDatabaseSymbol("second", kustolang.NewTableSymbol("T", kustolang.NewColumnSymbol("A", kustolang.ScalarTypes.String.Type()))))
	script := kustolang.NewCodeScript("T | project A", first)
	if script.Factory() == nil || script.Factory().Key() != "KustoCodeServiceFactory" || script.Globals() == nil || script.Separator() != nil {
		t.Fatalf("script metadata = factory:%#v globals:%#v separator:%#v", script.Factory(), script.Globals(), script.Separator())
	}

	withGlobals := script.WithGlobals(second)
	if withGlobals == script || withGlobals.Text() != script.Text() || withGlobals.Globals() == nil || withGlobals.Globals() == script.Globals() {
		t.Fatalf("WithGlobals = same:%t text:%q globals:%#v", withGlobals == script, withGlobals.Text(), withGlobals.Globals())
	}
	if script.Globals() == nil {
		t.Fatal("WithGlobals mutated the source script")
	}

	unknown := kustolang.NewUnknownCodeServiceFactory()
	replaced := script.WithTextAndFactory("plain text", unknown)
	if replaced.Text() != "plain text" || replaced.Factory() == nil || replaced.Factory().Key() != unknown.Key() || replaced.Globals() != nil {
		t.Fatalf("WithTextAndFactory = text:%q factory:%#v globals:%#v", replaced.Text(), replaced.Factory(), replaced.Globals())
	}
	if unchanged := replaced.WithFactory(replaced.Factory()); unchanged != replaced {
		t.Fatal("equivalent factory replacement should preserve script identity")
	}
}
