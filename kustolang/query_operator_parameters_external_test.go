package kustolang_test

import (
	"sort"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryOperatorParameterMetadataIsImmutable(t *testing.T) {
	values := []string{"inner", "outer"}
	aliases := []string{"join_kind"}
	parameter := kustolang.NewQueryOperatorParameter(
		"kind",
		kustolang.QueryOperatorParameterValueWord,
		kustolang.WithQueryOperatorParameterValues(values),
		kustolang.WithQueryOperatorParameterAliases(aliases),
		kustolang.WithQueryOperatorParameterCaseSensitive(false),
		kustolang.WithQueryOperatorParameterRepeatable(true),
	)
	values[0] = "changed"
	aliases[0] = "changed"

	if parameter.SymbolName() != "kind" || parameter.ValueKind() != kustolang.QueryOperatorParameterValueWord {
		t.Fatalf("parameter identity = %q/%s", parameter.SymbolName(), parameter.ValueKind())
	}
	if parameter.IsCaseSensitive() || !parameter.IsRepeatable() || parameter.HasNoEquals() {
		t.Fatalf("parameter flags = case:%t repeatable:%t no-equals:%t", parameter.IsCaseSensitive(), parameter.IsRepeatable(), parameter.HasNoEquals())
	}
	if got := parameter.Values(); len(got) != 2 || got[0] != "inner" {
		t.Fatalf("values = %#v", got)
	}
	if got := parameter.Aliases(); len(got) != 1 || got[0] != "join_kind" {
		t.Fatalf("aliases = %#v", got)
	}

	returnedValues := parameter.Values()
	returnedAliases := parameter.Aliases()
	returnedValues[0] = "mutated"
	returnedAliases[0] = "mutated"
	if parameter.Values()[0] != "inner" || parameter.Aliases()[0] != "join_kind" {
		t.Fatal("metadata accessors exposed mutable backing storage")
	}

	revalued := parameter.WithValues([]string{"leftouter"})
	if revalued == parameter || revalued.Values()[0] != "leftouter" || parameter.Values()[0] != "inner" {
		t.Fatal("WithValues did not return an independent parameter")
	}
	if same := parameter.WithIsCaseSensitive(false); same != parameter {
		t.Fatal("WithIsCaseSensitive replaced an unchanged parameter")
	}
	if hidden := parameter.Hide(); hidden == parameter || !hidden.IsHidden() || parameter.IsHidden() {
		t.Fatal("Hide did not preserve immutable clone behavior")
	}
}

func TestPublicQueryOperatorParameterCatalogMatchesPinnedUpstream(t *testing.T) {
	join := kustolang.QueryOperatorParameters.JoinParameters()
	if len(join) != 5 || join[0].SymbolName() != "kind" || join[0].Values()[0] != "inner" {
		t.Fatalf("join parameters = %#v", join)
	}
	if got := kustolang.QueryOperatorParameters.IsFuzzy.Aliases(); len(got) != 1 || got[0] != "__isFuzzy" {
		t.Fatalf("isfuzzy aliases = %#v", got)
	}
	if !kustolang.QueryOperatorParameters.HintDotShuffleKey.IsRepeatable() {
		t.Fatal("hint.shufflekey should be repeatable")
	}
	if !kustolang.QueryOperatorParameters.RenderWithDeprecated.HasNoEquals() ||
		!kustolang.QueryOperatorParameters.RenderByDeprecated.HasNoEquals() {
		t.Fatal("deprecated render parameters should omit equals")
	}
	if len(kustolang.QueryOperatorParameters.RenderWithProperties()) != 17 ||
		len(kustolang.QueryOperatorParameters.UnionParameters()) != 6 ||
		len(kustolang.QueryOperatorParameters.MacroExpandParameters()) != 3 {
		t.Fatal("catalog parameter-list sizes do not match the pinned source")
	}

	join[0] = nil
	if kustolang.QueryOperatorParameters.JoinParameters()[0] == nil {
		t.Fatal("catalog list method exposed mutable backing storage")
	}

	names := kustolang.KnownQueryOperatorParameterNames()
	if len(names) == 0 || !sort.StringsAreSorted(names) {
		t.Fatalf("known parameter names are not sorted: %#v", names)
	}
	foundAlias := false
	for _, name := range names {
		if name == "__isFuzzy" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Fatal("known parameter names omitted aliases")
	}
}

func TestMalformedQueryOperatorParameterStillReportsSyntax(t *testing.T) {
	code := kustolang.Parse(`T | join kind= (U) on Key`)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed query-operator parameter produced no diagnostic")
	}
}
