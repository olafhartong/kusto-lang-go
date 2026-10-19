package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicScalarFlagsAndPrimitiveSymbolMatchUpstream(t *testing.T) {
	if kustolang.ScalarFlagsNone != 0 ||
		kustolang.ScalarFlagsInteger != 1 ||
		kustolang.ScalarFlagsNumeric != 2 ||
		kustolang.ScalarFlagsInterval != 4 ||
		kustolang.ScalarFlagsSummable != 8 ||
		kustolang.ScalarFlagsOrderable != 16 ||
		kustolang.ScalarFlagsAll != 31 {
		t.Fatalf("scalar flags = none:%d integer:%d numeric:%d interval:%d summable:%d orderable:%d all:%d",
			kustolang.ScalarFlagsNone,
			kustolang.ScalarFlagsInteger,
			kustolang.ScalarFlagsNumeric,
			kustolang.ScalarFlagsInterval,
			kustolang.ScalarFlagsSummable,
			kustolang.ScalarFlagsOrderable,
			kustolang.ScalarFlagsAll,
		)
	}

	aliases := []string{"distance"}
	primitive := kustolang.NewPrimitiveSymbol(
		"distance_type",
		aliases,
		kustolang.ScalarFlagsNumeric|kustolang.ScalarFlagsOrderable,
		kustolang.ScalarTypes.Long,
	)
	aliases[0] = "mutated"
	if primitive.SymbolName() != "distance_type" || !reflect.DeepEqual(primitive.Aliases(), []string{"distance"}) {
		t.Fatalf("primitive identity = %q/%#v", primitive.SymbolName(), primitive.Aliases())
	}
	if primitive.IsInteger() || !primitive.IsNumeric() || primitive.IsInterval() || primitive.IsSummable() || !primitive.IsOrderable() || primitive.IsMultiValue() {
		t.Fatalf("primitive flags = integer:%t numeric:%t interval:%t summable:%t orderable:%t multi:%t",
			primitive.IsInteger(), primitive.IsNumeric(), primitive.IsInterval(), primitive.IsSummable(), primitive.IsOrderable(), primitive.IsMultiValue())
	}
	if !primitive.IsWiderThan(kustolang.ScalarTypes.Long) || primitive.IsWiderThan(primitive) {
		t.Fatalf("primitive widening = long:%t self:%t", primitive.IsWiderThan(kustolang.ScalarTypes.Long), primitive.IsWiderThan(primitive))
	}
	if got := kustolang.PrimitiveSymbolFrom("int64"); got != kustolang.ScalarTypes.Long {
		t.Fatalf("PrimitiveSymbolFrom(int64) = %p, want %p", got, kustolang.ScalarTypes.Long)
	}
}

func TestPublicSchemaDisplayMatchesUpstreamText(t *testing.T) {
	table := kustolang.NewTableSymbol("Events",
		kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("bad name", kustolang.ScalarTypes.Long.Type()),
		kustolang.NewColumnSymbol("where", kustolang.ScalarTypes.DynamicArray.Type()),
	)
	if got, want := kustolang.SchemaDisplay.GetText(table), "(Account: string, ['bad name']: long, ['where']: dynamic)"; got != want {
		t.Fatalf("SchemaDisplay.GetText(table) = %q, want %q", got, want)
	}
	if got := kustolang.SchemaDisplay.GetScalarTypeText(kustolang.ScalarTypes.DynamicArray.Type()); got != "dynamic" {
		t.Fatalf("dynamic array scalar display = %q", got)
	}

	function := kustolang.NewScalarFunctionSymbol("f", kustolang.ScalarTypes.Bool.Type(),
		kustolang.NewParameter("value", kustolang.ScalarTypes.String.Type()),
		kustolang.NewParameterKind("rows", kustolang.ParameterTypeTabular),
		kustolang.NewParameterTypes("choice", []*kustolang.TypeSymbol{kustolang.ScalarTypes.String.Type(), kustolang.ScalarTypes.Long.Type()}),
	)
	if got, want := kustolang.SchemaDisplay.GetText(function), "(value: string, rows: (*), choice: dynamic)"; got != want {
		t.Fatalf("SchemaDisplay.GetText(function) = %q, want %q", got, want)
	}
}

func TestPublicCompletionDisplayAndSymbolEnums(t *testing.T) {
	required := kustolang.NewParameter("left", kustolang.ScalarTypes.Long.Type())
	optional := kustolang.NewParameter("right", kustolang.ScalarTypes.Long.Type(), kustolang.WithParameterOccurrence(0, 1))
	repeat := kustolang.NewParameter("more", kustolang.ScalarTypes.Long.Type(), kustolang.WithParameterOccurrence(0, 100))
	function := kustolang.NewFunctionSymbol("sum2", kustolang.NewSignature(kustolang.ScalarTypes.Long.Type(), required, optional, repeat))
	if got, want := kustolang.CompletionDisplay.GetText(function), "sum2(left, [right], [more], ...)"; got != want {
		t.Fatalf("CompletionDisplay.GetText(function) = %q, want %q", got, want)
	}
	if got := kustolang.CompletionDisplay.GetText(function, true); got != "sum2" {
		t.Fatalf("name-only function display = %q", got)
	}
	external := kustolang.NewExternalTableSymbol("Archive")
	if got := kustolang.CompletionDisplay.GetText(external); got != "external_table('Archive')" {
		t.Fatalf("external table display = %q", got)
	}

	if kustolang.ServerKindEngine != "Engine" ||
		kustolang.ServerKindDataManager != "DataManager" ||
		kustolang.ServerKindClusterManager != "ClusterManager" ||
		kustolang.ServerKindAriaBridge != "AriaBridge" {
		t.Fatalf("server kinds = %q/%q/%q/%q", kustolang.ServerKindEngine, kustolang.ServerKindDataManager, kustolang.ServerKindClusterManager, kustolang.ServerKindAriaBridge)
	}
	view := kustolang.NewMaterializedViewSymbol("View", nil, "T | summarize count()")
	if view.MaterializedViewKind() != kustolang.MaterializedViewKindUnknown || kustolang.MaterializedViewKindDownsampling == kustolang.MaterializedViewKindOther {
		t.Fatalf("materialized view kind = %q", view.MaterializedViewKind())
	}
}
