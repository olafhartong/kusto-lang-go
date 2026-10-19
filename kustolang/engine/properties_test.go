package kustolang

import "testing"

func TestGlobalStatePropertiesMirrorUpstreamDefaultsAndImmutability(t *testing.T) {
	globals := NewGlobalState(nil)
	if GlobalStateGetProperty(globals, Properties.AllowClientParameters) {
		t.Fatal("AllowClientParameters default = true, want false")
	}
	if got := GlobalStateGetProperty(globals, Properties.MaxParseTextSize); got != 4*1024*1024 {
		t.Fatalf("MaxParseTextSize = %d", got)
	}
	if got := GlobalStateGetProperty(globals, Properties.MaxAnalysisDepth); got != 500 {
		t.Fatalf("MaxAnalysisDepth = %d", got)
	}
	if got := GlobalStateGetProperty(globals, Properties.MaxCachedExpansions); got != 10 {
		t.Fatalf("MaxCachedExpansions = %d", got)
	}
	if got := GlobalStateGetProperty(globals, Properties.MaxCachedResultTypes); got != 50 {
		t.Fatalf("MaxCachedResultTypes = %d", got)
	}

	cached := globals.WithCache()
	changed := GlobalStateWithProperty(cached, Properties.MaxParseTextSize, 128)
	if changed == cached {
		t.Fatal("WithProperty should clone when value changes")
	}
	if changed.Cache == nil || changed.Cache == cached.Cache || changed.Cache.Globals != changed {
		t.Fatalf("property change should rebind cache, got old:%#v new:%#v", cached.Cache, changed.Cache)
	}
	if got := GlobalStateGetProperty(changed, Properties.MaxParseTextSize); got != 128 {
		t.Fatalf("changed MaxParseTextSize = %d", got)
	}
	if got := GlobalStateGetProperty(cached, Properties.MaxParseTextSize); got != 4*1024*1024 {
		t.Fatalf("original MaxParseTextSize mutated = %d", got)
	}
	if same := GlobalStateWithProperty(changed, Properties.MaxParseTextSize, 128); same != changed {
		t.Fatal("WithProperty with same value should return the same GlobalState")
	}
	if got := changed.GetProperty(Properties.MaxParseTextSize); got != 128 {
		t.Fatalf("non-generic GetProperty = %#v", got)
	}
}

func TestSyntaxTreeIsSafeToRecurseUsesGlobalProperty(t *testing.T) {
	code := Parse("T | where Value == 1 | project Value")
	if !code.Tree.IsSafeToRecurse(DefaultGlobalState) {
		t.Fatal("default MaxAnalysisDepth should allow normal query tree")
	}
	shallow := GlobalStateWithProperty(DefaultGlobalState, Properties.MaxAnalysisDepth, 1)
	if code.Tree.IsSafeToRecurse(shallow) {
		t.Fatal("MaxAnalysisDepth=1 should reject normal query tree")
	}
}
