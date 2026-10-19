package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

var _ *kustolang.ScalarTypesAPI = kustolang.ScalarTypes

func TestPublicScalarTypesSourceCatalogAndLookup(t *testing.T) {
	if kustolang.ScalarTypes.Type == nil || kustolang.ScalarTypes.Type != kustolang.ScalarTypes.TypeLit || kustolang.ScalarTypes.Type.SymbolName() != "type" {
		t.Fatalf("type symbol = %#v / %#v", kustolang.ScalarTypes.Type, kustolang.ScalarTypes.TypeLit)
	}
	want := []*kustolang.ScalarSymbol{
		kustolang.ScalarTypes.Bool,
		kustolang.ScalarTypes.Int,
		kustolang.ScalarTypes.Long,
		kustolang.ScalarTypes.Real,
		kustolang.ScalarTypes.Decimal,
		kustolang.ScalarTypes.String,
		kustolang.ScalarTypes.DateTime,
		kustolang.ScalarTypes.TimeSpan,
		kustolang.ScalarTypes.Guid,
		kustolang.ScalarTypes.Type,
		kustolang.ScalarTypes.Dynamic,
		kustolang.ScalarTypes.Null,
	}
	all := kustolang.ScalarTypes.All()
	if !reflect.DeepEqual(all, want) {
		t.Fatalf("scalar catalog = %#v", all)
	}
	all[0] = kustolang.ScalarTypes.Unknown
	if kustolang.ScalarTypes.All()[0] != kustolang.ScalarTypes.Bool {
		t.Fatal("All returned mutable catalog storage")
	}

	if kustolang.ScalarTypes.GetSymbol("long") != kustolang.ScalarTypes.Long || kustolang.ScalarTypes.GetSymbol("int64") != kustolang.ScalarTypes.Long {
		t.Fatal("strict scalar lookup did not resolve source names and aliases")
	}
	if kustolang.ScalarTypes.GetSymbol("Long") != nil || kustolang.ScalarTypes.GetSymbol(" long ") != nil || kustolang.ScalarTypes.GetSymbol("unknown") != nil {
		t.Fatal("GetSymbol should use ordinal, untrimmed lookup")
	}
	if kustolang.ScalarTypes.Get(" Long ") != kustolang.ScalarTypes.Long || kustolang.ScalarTypes.Get("unknown") != kustolang.ScalarTypes.Unknown {
		t.Fatal("compatibility Get should remain case-insensitive and trimmed")
	}
}

func TestPublicScalarTypesSchemaHelpers(t *testing.T) {
	bag, err := kustolang.ScalarTypes.GetDynamicBagFromSchema("Name:string, Count:long")
	if err != nil {
		t.Fatalf("GetDynamicBagFromSchema: %v", err)
	}
	if bag == nil || len(bag.Properties()) != 2 || bag.Properties()[0].SymbolName() != "Name" || bag.Properties()[1].Type != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("dynamic bag = %#v", bag)
	}

	tuple, err := kustolang.ScalarTypes.GetTupleFromSchema("Name:string, Count:long")
	if err != nil {
		t.Fatalf("GetTupleFromSchema: %v", err)
	}
	if tuple == nil || len(tuple.Columns()) != 2 || tuple.Columns()[1].SymbolName() != "Count" {
		t.Fatalf("tuple = %#v", tuple)
	}
}
