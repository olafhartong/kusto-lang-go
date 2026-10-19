package kustolang

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateCatalogAttestation = flag.Bool("update-catalog-attestation", false, "rewrite the pinned catalog attestation golden file")

// TestCatalogInventoriesMatchPinnedUpstream pins the exact symbol name
// inventories of every built-in catalog to the upstream commit recorded in
// third_party/kusto/UPSTREAM.json. A repin that changes any catalog must
// regenerate the golden with -update-catalog-attestation and review the diff
// against the upstream source diff.
func TestCatalogInventoriesMatchPinnedUpstream(t *testing.T) {
	got := renderCatalogAttestation()
	golden := filepath.Join("testdata", "catalog_attestation.golden")
	if *updateCatalogAttestation {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read catalog attestation (run with -update-catalog-attestation to create it): %v", err)
	}
	if string(want) != got {
		t.Fatalf("catalog inventories drifted from %s:\n%s\nIf intentional (upstream repin), regenerate with -update-catalog-attestation and review the diff.",
			golden, catalogAttestationDiff(string(want), got))
	}
}

func renderCatalogAttestation() string {
	var b strings.Builder
	section := func(name string, names []string) {
		fmt.Fprintf(&b, "[%s] count=%d\n", name, len(names))
		for _, name := range names {
			b.WriteString(name)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	functionNames := func(symbols []*FunctionSymbol) []string {
		names := make([]string, len(symbols))
		for i, symbol := range symbols {
			names[i] = symbol.SymbolName()
		}
		return names
	}
	operatorNames := func(symbols []*OperatorSymbol) []string {
		names := make([]string, len(symbols))
		for i, symbol := range symbols {
			names[i] = symbol.SymbolName()
		}
		return names
	}
	commandNames := func(symbols []*CommandSymbol) []string {
		names := make([]string, len(symbols))
		for i, symbol := range symbols {
			names[i] = symbol.SymbolName()
		}
		return names
	}

	section("functions", functionNames(Functions.All))
	section("aggregates", functionNames(Aggregates.All))
	section("plugins", functionNames(PlugIns.All))
	section("operators", operatorNames(Operators.All))
	section("engine-commands", commandNames(EngineCommands.All))
	section("data-manager-commands", commandNames(DataManagerCommands.All))
	return b.String()
}

func catalogAttestationDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var b strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	printed := 0
	for i := 0; i < max && printed < 40; i++ {
		var wantLine, gotLine string
		if i < len(wantLines) {
			wantLine = wantLines[i]
		}
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}
		if wantLine != gotLine {
			fmt.Fprintf(&b, "line %d: golden %q, current %q\n", i+1, wantLine, gotLine)
			printed++
		}
	}
	if printed == 0 {
		return "(no line differences; trailing content differs)"
	}
	return b.String()
}
