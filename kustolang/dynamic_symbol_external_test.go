package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicDynamicBagConstructionAndUpdates(t *testing.T) {
	account := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
	score := kustolang.NewColumnSymbol("Score", kustolang.ScalarTypes.Long.Type())
	input := []*kustolang.ColumnSymbol{account, score}
	bag := kustolang.NewDynamicBagSymbol(input...)
	input[0] = nil

	if bag.SymbolKind() != kustolang.SymbolKindBag || bag.SymbolTabularity() != kustolang.TabularityScalar {
		t.Fatalf("dynamic bag metadata = kind:%v tabularity:%v", bag.SymbolKind(), bag.SymbolTabularity())
	}
	if got, ok := bag.TryGetProperty("Account"); !ok || got != account {
		t.Fatalf("Account property = %#v, %t", got, ok)
	}
	properties := bag.Properties()
	properties[0] = nil
	if got, ok := bag.TryGetProperty("Account"); !ok || got != account {
		t.Fatal("Properties returned mutable internal storage")
	}

	updatedAccount := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.Long.Type())
	updated := bag.AddOrUpdateProperty(updatedAccount)
	if got, ok := updated.TryGetProperty("Account"); !ok || got != updatedAccount {
		t.Fatalf("updated Account property = %#v, %t", got, ok)
	}
	if original, _ := bag.TryGetProperty("Account"); original != account {
		t.Fatal("AddOrUpdateProperty mutated the original bag")
	}

	source := kustolang.CreateFakeExpression(kustolang.ScalarTypes.Dynamic.Type())
	sourced := updated.WithSource(source)
	for _, property := range sourced.Properties() {
		if property.Source != source {
			t.Fatalf("property %s source = %#v, want fake expression", property.SymbolName(), property.Source)
		}
	}
}

func TestPublicDynamicBagEmptyIsCanonical(t *testing.T) {
	if kustolang.DynamicBagEmpty == nil || len(kustolang.DynamicBagEmpty.Properties()) != 0 {
		t.Fatalf("DynamicBagEmpty = %#v", kustolang.DynamicBagEmpty)
	}
	if kustolang.DynamicBagEmpty != kustolang.ScalarTypes.DynamicBag {
		t.Fatal("DynamicBagEmpty does not share the scalar catalog singleton")
	}
}

func TestPublicDynamicSymbolHierarchyAndAnySingleton(t *testing.T) {
	if kustolang.DynamicAnySymbolInstance == nil ||
		kustolang.DynamicAnySymbolInstance.SymbolName() != "dynamic" ||
		kustolang.DynamicAnySymbolInstance.Type() != kustolang.ScalarTypes.Dynamic.Type() {
		t.Fatalf("DynamicAnySymbolInstance = %#v", kustolang.DynamicAnySymbolInstance)
	}

	baseSymbols := []*kustolang.DynamicSymbol{
		&kustolang.DynamicAnySymbolInstance.DynamicSymbol,
		&kustolang.ScalarTypes.DynamicLong.DynamicSymbol,
		&kustolang.ScalarTypes.DynamicArray.DynamicSymbol,
		&kustolang.ScalarTypes.DynamicBag.DynamicSymbol,
	}
	for index, symbol := range baseSymbols {
		if symbol == nil || symbol.SymbolName() != "dynamic" || symbol.SymbolTabularity() != kustolang.TabularityScalar {
			t.Fatalf("dynamic base[%d] = %#v", index, symbol)
		}
	}
}
