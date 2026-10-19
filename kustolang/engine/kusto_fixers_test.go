package kustolang

import "testing"

func TestKustoFixersCatalogAllFollowsAnalyzerFixers(t *testing.T) {
	analyzers := KustoAnalyzers.All()
	fixers := KustoFixers.All()
	if len(fixers) != len(analyzers) {
		t.Fatalf("fixer count = %d, want %d", len(fixers), len(analyzers))
	}
	for i, analyzer := range analyzers {
		if fixers[i] != analyzer.Fixer() {
			t.Fatalf("fixer[%d] = %#v, want analyzer fixer %#v", i, fixers[i], analyzer.Fixer())
		}
		if fixers[i].Name() != analyzer.Name() {
			t.Fatalf("fixer[%d] name = %q, want %q", i, fixers[i].Name(), analyzer.Name())
		}
	}
}

func TestKustoFixersAllReturnsCopy(t *testing.T) {
	all := KustoFixers.All()
	if len(all) == 0 {
		t.Fatal("expected fixers")
	}
	original := all[0]
	all[0] = nil
	if KustoFixers.All()[0] != original {
		t.Fatal("All did not return a defensive copy")
	}
	if got := (*KustoFixersCatalog)(nil).All(); got == nil || len(got) != 0 {
		t.Fatalf("nil catalog All = %#v, want non-nil empty", got)
	}
}
