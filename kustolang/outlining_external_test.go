package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicOutliningTypesArePackageOwnedFacades(t *testing.T) {
	const publicPackage = "github.com/olafhartong/kusto-lang-go/kustolang"
	types := []reflect.Type{
		reflect.TypeOf(kustolang.OutliningOptions{}),
		reflect.TypeOf(kustolang.OutlineRange{}),
		reflect.TypeOf(kustolang.OutlineInfo{}),
	}
	for _, typ := range types {
		if got := typ.PkgPath(); got != publicPackage {
			t.Fatalf("%v package = %q, want %q", typ, got, publicPackage)
		}
	}
}

func TestPublicOutliningOptionsMatchPinnedDefaultsAndWithMethods(t *testing.T) {
	defaults := kustolang.OutliningOptionsDefault
	if defaults.MaxCollapsedTextLength != 120 || !defaults.Queries || defaults.Statements {
		t.Fatalf("outlining defaults = %#v", defaults)
	}

	changed := defaults.
		WithMaxCollapesedTextlength(60).
		WithQueries(false).
		WithStatements(true)
	if changed.MaxCollapsedTextLength != 60 || changed.Queries || !changed.Statements {
		t.Fatalf("changed outlining options = %#v", changed)
	}
	if corrected := defaults.WithMaxCollapsedTextLength(30); corrected.MaxCollapsedTextLength != 30 {
		t.Fatalf("correctly spelled helper = %#v", corrected)
	}
	if defaults != kustolang.OutliningOptionsDefault {
		t.Fatalf("with methods mutated defaults: %#v", defaults)
	}
}

func TestPublicOutlineValuesMatchPinnedConstructionAndOwnership(t *testing.T) {
	ranges := []kustolang.OutlineRange{
		kustolang.NewOutlineRange(2, 8, "query ..."),
		kustolang.NewOutlineRangeOptional(12, 4, nil),
	}
	info := kustolang.NewOutlineInfo(ranges)
	ranges[0].CollapsedText = "changed"

	if got := info.Ranges[0]; got.Start != 2 || got.Length != 8 || got.CollapsedText != "query ..." {
		t.Fatalf("first outline = %#v", got)
	}
	if got := info.Ranges[1].CollapsedText; got != "" {
		t.Fatalf("nil collapsed text = %q, want empty", got)
	}
	copyRanges := info.RangesList()
	copyRanges[0].CollapsedText = "changed copy"
	if info.Ranges[0].CollapsedText != "query ..." {
		t.Fatalf("RangesList exposed outline storage: %#v", info)
	}
	if kustolang.OutlineInfoEmpty.Ranges == nil || len(kustolang.OutlineInfoEmpty.Ranges) != 0 {
		t.Fatalf("empty outline info = %#v", kustolang.OutlineInfoEmpty)
	}
}

func TestPublicCustomCodeServiceConvertsOutliningFacades(t *testing.T) {
	wantOptions := kustolang.OutliningOptionsDefault.
		WithMaxCollapsedTextLength(42).
		WithQueries(false).
		WithStatements(true)
	callbackRanges := []kustolang.OutlineRange{kustolang.NewOutlineRange(1, 7, "custom")}
	var gotOptions kustolang.OutliningOptions
	service := kustolang.NewCustomCodeService("custom query", kustolang.CustomCodeServiceCallbacks{
		GetOutlines: func(options kustolang.OutliningOptions, _ kustolang.CancellationToken) kustolang.OutlineInfo {
			gotOptions = options
			return kustolang.OutlineInfo{Ranges: callbackRanges}
		},
	})

	got := service.GetOutlines(wantOptions, kustolang.CancellationToken{})
	if gotOptions != wantOptions {
		t.Fatalf("callback options = %#v, want %#v", gotOptions, wantOptions)
	}
	if !reflect.DeepEqual(got.Ranges, callbackRanges) {
		t.Fatalf("outlines = %#v, want %#v", got.Ranges, callbackRanges)
	}
	got.Ranges[0].CollapsedText = "changed"
	if callbackRanges[0].CollapsedText != "custom" {
		t.Fatalf("callback range storage was exposed: %#v", callbackRanges)
	}
}

func TestPublicNilCodeServiceReturnsPublicEmptyOutlineInfo(t *testing.T) {
	var service *kustolang.CodeService
	got := service.GetOutlines(kustolang.OutliningOptionsDefault, kustolang.CancellationToken{})
	if got.Ranges == nil || len(got.Ranges) != 0 {
		t.Fatalf("nil service outlines = %#v", got)
	}
	if reflect.TypeOf(got).PkgPath() != "github.com/olafhartong/kusto-lang-go/kustolang" {
		t.Fatalf("nil service returned non-public type %T", got)
	}
}
