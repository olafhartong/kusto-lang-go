package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicTypeFactsCommonAndConversionBehavior(t *testing.T) {
	if got := kustolang.GetElementType(kustolang.ScalarTypes.DynamicArrayOfString.Type()); got != kustolang.ScalarTypes.DynamicString.Type() {
		t.Fatalf("dynamic string-array element = %v", got)
	}
	if got := kustolang.GetWidestScalarType(
		kustolang.ScalarTypes.Int.Type(),
		kustolang.ScalarTypes.Long.Type(),
		kustolang.ScalarTypes.Real.Type(),
	); got != kustolang.ScalarTypes.Real {
		t.Fatalf("widest numeric type = %v", got)
	}
	if got := kustolang.GetCommonScalarType([]*kustolang.TypeSymbol{
		kustolang.ScalarTypes.Int.Type(),
		kustolang.ScalarTypes.Long.Type(),
	}); got != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("common int/long type = %v", got)
	}
	if got := kustolang.GetCommonResultType([]*kustolang.TypeSymbol{
		kustolang.ScalarTypes.Int.Type(),
		kustolang.ScalarTypes.String.Type(),
	}); got != nil {
		t.Fatalf("common int/string type = %v, want nil", got)
	}
	if !kustolang.IsConversionAllowed(kustolang.ConversionPromotable, kustolang.ConversionCompatible) ||
		kustolang.IsConversionAllowed(kustolang.ConversionAny, kustolang.ConversionPromotable) {
		t.Fatal("conversion allowance ordering does not match upstream")
	}
	if kustolang.PromoteToLong(kustolang.ScalarTypes.Int.Type()) != kustolang.ScalarTypes.Long.Type() ||
		!kustolang.IsPromotableTo(kustolang.ScalarTypes.Int.Type(), kustolang.ScalarTypes.Long.Type()) ||
		kustolang.IsPromotableTo(kustolang.ScalarTypes.Long.Type(), kustolang.ScalarTypes.Int.Type()) {
		t.Fatal("integer promotion behavior does not match upstream")
	}
}

func TestPublicTryGetCommonTypeSymbolWideAdaptation(t *testing.T) {
	common, ok := kustolang.TryGetCommonType(kustolang.ScalarTypes.Int.Type(), kustolang.ScalarTypes.Real.Type(), kustolang.ConversionPromotable)
	if !ok || common != kustolang.ScalarTypes.Real.Type() {
		t.Fatalf("common scalar type = %#v, %t", common, ok)
	}

	left := kustolang.NewTableSymbol("Left",
		kustolang.NewColumnSymbol("Shared", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("LeftOnly", kustolang.ScalarTypes.Long.Type()),
	)
	right := kustolang.NewExternalTableSymbol("Right",
		kustolang.NewColumnSymbol("Shared", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("RightOnly", kustolang.ScalarTypes.Bool.Type()),
	)
	commonSymbol, ok := kustolang.TryGetCommonSymbolType(left, right, kustolang.ConversionPromotable)
	commonTable, tableOK := commonSymbol.(*kustolang.TableSymbol)
	if !ok || !tableOK || len(commonTable.Columns()) != 1 || commonTable.Columns()[0].SymbolName() != "Shared" {
		t.Fatalf("common table symbol = %#v, %t", commonSymbol, ok)
	}
}

func TestPublicTypeFactsDefaultAndIgnoreDynamicForms(t *testing.T) {
	fallback := kustolang.ScalarTypes.Unknown.Type()
	commonScalar := kustolang.NewParameterKind("value", kustolang.ParameterTypeCommonScalar)
	parameters := []*kustolang.Parameter{commonScalar, commonScalar}

	if got := kustolang.GetCommonArgumentTypeWithDefault(
		parameters,
		[]*kustolang.TypeSymbol{kustolang.ScalarTypes.String.Type(), kustolang.ScalarTypes.Long.Type()},
		fallback,
		false,
	); got != fallback {
		t.Fatalf("incompatible common argument type = %v, want fallback", got)
	}
	if got := kustolang.GetCommonArgumentTypeWithDefault(
		parameters,
		[]*kustolang.TypeSymbol{kustolang.ScalarTypes.Dynamic.Type(), kustolang.ScalarTypes.Long.Type()},
		fallback,
		true,
	); got != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("ignore-dynamic common argument type = %v, want long", got)
	}
	if got := kustolang.GetCommonArgumentTypeIgnoreDynamic(
		parameters,
		[]*kustolang.TypeSymbol{kustolang.ScalarTypes.Dynamic.Type(), kustolang.ScalarTypes.Long.Type()},
	); got != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("ignore-dynamic convenience form = %v, want long", got)
	}

	columns := []*kustolang.ColumnSymbol{
		kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("Value", kustolang.ScalarTypes.Long.Type()),
	}
	if got := kustolang.GetCommonColumnTypeWithDefault(columns, kustolang.ConversionPromotable, fallback); got != fallback {
		t.Fatalf("incompatible common column type = %v, want fallback", got)
	}
	if got := kustolang.GetCommonResultTypeWithDefault(
		[]*kustolang.TypeSymbol{kustolang.ScalarTypes.Dynamic.Type(), kustolang.ScalarTypes.Long.Type()},
		kustolang.ConversionPromotable,
		fallback,
		true,
	); got != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("ignore-dynamic common result type = %v, want long", got)
	}
	if got := kustolang.GetCommonResultTypeWithDefault(
		[]*kustolang.TypeSymbol{kustolang.ScalarTypes.String.Type(), kustolang.ScalarTypes.Long.Type()},
		kustolang.ConversionPromotable,
		fallback,
		false,
	); got != fallback {
		t.Fatalf("incompatible common result type = %v, want fallback", got)
	}
}

func TestPublicTypeFactsScalarCategories(t *testing.T) {
	if !kustolang.IsInteger(kustolang.ScalarTypes.Long) || !kustolang.IsNumeric(kustolang.ScalarTypes.Real) ||
		!kustolang.IsSummable(kustolang.ScalarTypes.TimeSpan) || !kustolang.IsOrderable(kustolang.ScalarTypes.String) {
		t.Fatal("core scalar categories rejected matching types")
	}
	if kustolang.IsAnyScalarExceptBool(kustolang.ScalarTypes.Bool) ||
		!kustolang.IsNumericOrBool(kustolang.ScalarTypes.Bool) ||
		!kustolang.IsStringOrDynamic(kustolang.ScalarTypes.Dynamic) {
		t.Fatal("composite scalar categories do not match upstream")
	}
	if !kustolang.IsDynamicArray(kustolang.ScalarTypes.DynamicArrayOfLong) ||
		!kustolang.IsDynamicBag(kustolang.ScalarTypes.DynamicBag) ||
		!kustolang.IsDynamicArrayOrBag(kustolang.ScalarTypes.Dynamic) {
		t.Fatal("dynamic shape categories do not match upstream")
	}
}

func TestPublicTypeFactsUnionAndIntersectColumns(t *testing.T) {
	account := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
	score := kustolang.NewColumnSymbol("Score", kustolang.ScalarTypes.Int.Type())
	tenant := kustolang.NewColumnSymbol("Tenant", kustolang.ScalarTypes.String.Type())
	union := kustolang.Union(
		[]*kustolang.ColumnSymbol{account, score},
		[]*kustolang.ColumnSymbol{kustolang.NewColumnSymbol("Score", kustolang.ScalarTypes.Long.Type()), tenant},
	)
	if len(union) != 3 || union[0] != account || union[1] != score || union[2] != tenant {
		t.Fatalf("union columns = %#v", union)
	}

	intersection := kustolang.Intersect(
		[]*kustolang.ColumnSymbol{score, account},
		[]*kustolang.ColumnSymbol{kustolang.NewColumnSymbol("Score", kustolang.ScalarTypes.Long.Type())},
	)
	if len(intersection) != 1 || intersection[0].SymbolName() != "Score" || intersection[0].Type != kustolang.ScalarTypes.Long.Type() {
		t.Fatalf("intersection columns = %#v", intersection)
	}
}
