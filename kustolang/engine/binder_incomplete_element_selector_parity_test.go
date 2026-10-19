package kustolang

import "testing"

func TestNativeIncompleteElementSelectorStillBindsSemanticDiagnostic(t *testing.T) {
	query := "let F=(T:(k:long)){T | summarize c=count()}; F(datatablek:long)[1,2])"
	code := ParseAndAnalyze(query)
	wantCodes := []string{"KS111", "KS142", "KS228", "KS142", "KS147", "KS005", "KS198"}
	wantStarts := []int{47, 47, 58, 58, 63, 65, 65}
	diagnostics := code.GetDiagnostics()
	if len(diagnostics) != len(wantCodes) {
		t.Fatalf("diagnostics = %#v, want codes %v", diagnostics, wantCodes)
	}
	for i := range wantCodes {
		if diagnostics[i].Code != wantCodes[i] || diagnostics[i].Start != wantStarts[i] {
			t.Fatalf("diagnostic %d = %#v, want %s at %d", i, diagnostics[i], wantCodes[i], wantStarts[i])
		}
	}
}
