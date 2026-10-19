package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicPlugInResultSchemasMatchPinnedUpstream(t *testing.T) {
	tests := []struct {
		name    string
		columns []*kustolang.ColumnSymbol
		want    []string
	}{
		{"AutoClusterColumns", kustolang.PlugIns.AutoClusterColumns(), []string{"SegmentId:long", "Count:long", "Percent:real"}},
		{"BasketColumns", kustolang.PlugIns.BasketColumns(), []string{"SegmentId:long", "Count:long", "Percent:real"}},
		{"DiffPatternsColumns", kustolang.PlugIns.DiffPatternsColumns(), []string{"SegmentId:long", "CountA:long", "CountB:long", "PercentA:real", "PercentB:real", "PercentDiffAB:real"}},
		{"HttpRequestColumns", kustolang.PlugIns.HttpRequestColumns(), []string{"ResponseHeaders:dynamic", "ResponseBody:dynamic"}},
		{"NarrowColumns", kustolang.PlugIns.NarrowColumns(), []string{"Row:long", "Column:string", "Value:string"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := make([]string, len(test.columns))
			for i, column := range test.columns {
				got[i] = column.SymbolName() + ":" + column.Type.SymbolName()
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("schema = %#v, want %#v", got, test.want)
			}
		})
	}

	columns := kustolang.PlugIns.NarrowColumns()
	columns[0] = nil
	if kustolang.PlugIns.NarrowColumns()[0] == nil {
		t.Fatal("NarrowColumns exposed mutable catalog storage")
	}
	var nilCatalog *kustolang.PlugInCatalog
	if got := nilCatalog.NarrowColumns(); got != nil {
		t.Fatalf("nil catalog columns = %#v, want nil", got)
	}
}

func TestPublicPlugInCatalogMatchesPinnedUpstream(t *testing.T) {
	all := kustolang.PlugIns.All()
	if len(all) != 48 {
		t.Fatalf("plug-in count = %d, want 48", len(all))
	}
	if all[0] != kustolang.PlugIns.ActiveUseCounts || all[len(all)-1] != kustolang.PlugIns.AIEmbeddings {
		t.Fatalf("plug-in order = first:%v last:%v", all[0], all[len(all)-1])
	}
	if kustolang.PlugIns.CSharp.SymbolName() != "csharp" {
		t.Fatalf("CSharp symbol name = %q, want csharp", kustolang.PlugIns.CSharp.SymbolName())
	}
	if kustolang.PlugIns.SchemaMerge.SymbolName() != "schema_merge" {
		t.Fatalf("SchemaMerge symbol name = %q", kustolang.PlugIns.SchemaMerge.SymbolName())
	}
	for _, plugin := range all {
		if plugin == kustolang.PlugIns.SchemaMerge {
			t.Fatal("SchemaMerge should be public but absent from pinned PlugIns.All")
		}
	}
	if got := kustolang.PlugIns.GetPlugIn("bag_unpack"); got != kustolang.PlugIns.BagUnpack {
		t.Fatalf("GetPlugIn(bag_unpack) = %p, want %p", got, kustolang.PlugIns.BagUnpack)
	}
	if got := kustolang.PlugIns.GetPlugIn("schema_merge"); got != nil {
		t.Fatalf("GetPlugIn(schema_merge) = %p, want nil", got)
	}

	all[0] = nil
	if kustolang.PlugIns.All()[0] == nil {
		t.Fatal("PlugIns.All exposed mutable catalog storage")
	}
}

func TestPublicPlugInCatalogDrivesValidAndMalformedBinding(t *testing.T) {
	valid := kustolang.ParseAndAnalyze(`datatable(d:dynamic)[dynamic({"x": 1})] | evaluate bag_unpack(d)`)
	if diagnostics := valid.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid plug-in query diagnostics = %#v", diagnostics)
	}

	malformed := kustolang.ParseAndAnalyze(`datatable(d:dynamic)[dynamic({"x": 1})] | evaluate bag_unpack(d`)
	if diagnostics := malformed.GetDiagnostics(); len(diagnostics) == 0 {
		t.Fatal("malformed plug-in call produced no diagnostic")
	}
}
