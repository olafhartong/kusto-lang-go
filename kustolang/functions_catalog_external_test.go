package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicFunctionCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.Functions.All()
	if len(all) != 429 {
		t.Fatalf("function count = %d, want 429", len(all))
	}
	if all[0] != kustolang.Functions.Cluster || all[len(all)-1] != kustolang.Functions.ColumnNamesOf {
		t.Fatalf("function order = first:%v last:%v", all[0], all[len(all)-1])
	}

	tests := []struct {
		field *kustolang.FunctionSymbol
		name  string
	}{
		{kustolang.Functions.Strcat, "strcat"},
		{kustolang.Functions.GetType, "gettype"},
		{kustolang.Functions.ToDynamic, "todynamic"},
		{kustolang.Functions.GraphAll, "all"},
		{kustolang.Functions.GeoDistance2Points, "geo_distance_2points"},
		{kustolang.Functions.ConvertVolume, "convert_volume"},
		{kustolang.Functions.ColumnNamesOf, "column_names_of"},
	}
	for _, test := range tests {
		if test.field == nil || test.field.SymbolName() != test.name {
			t.Fatalf("function field for %q = %#v", test.name, test.field)
		}
	}
	if got := kustolang.DefaultGlobalState.GetFunction("strcat"); got != kustolang.Functions.Strcat {
		t.Fatalf("default strcat = %p, catalog = %p", got, kustolang.Functions.Strcat)
	}

	all[0] = nil
	if kustolang.Functions.All()[0] == nil {
		t.Fatal("Functions.All exposed mutable catalog storage")
	}
	signatures := kustolang.Functions.Strcat.Signatures()
	signatures[0] = nil
	if kustolang.Functions.Strcat.Signatures()[0] == nil {
		t.Fatal("FunctionSymbol.Signatures exposed mutable storage")
	}
}

func TestPublicFunctionCatalogDrivesValidAndMalformedBinding(t *testing.T) {
	valid := kustolang.ParseAndAnalyze(`print text=tostring(42), radians=convert_angle(180.0, "Degree", "Radian")`)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid function query diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze(`print value=strcat("a",`)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed function call produced no diagnostic")
	}
}
