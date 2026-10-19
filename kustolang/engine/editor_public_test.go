package kustolang

import "testing"

func TestCodeKindsAndServiceFeatureConstantsMatchUpstream(t *testing.T) {
	if CodeKindsQuery != "Query" || CodeKindsCommand != "Command" || CodeKindsDirective != "Directive" || CodeKindsUnknown != "Unknown" {
		t.Fatalf("code kinds = query:%q command:%q directive:%q unknown:%q", CodeKindsQuery, CodeKindsCommand, CodeKindsDirective, CodeKindsUnknown)
	}
	if CodeKindsQuery != string(CodeKindQuery) || CodeKindsCommand != string(CodeKindCommand) || CodeKindsDirective != string(CodeKindDirective) || CodeKindsUnknown != string(CodeKindUnknown) {
		t.Fatalf("code kind constants diverged from KustoCode constants")
	}

	features := []string{
		CodeServiceFeatureDiagnostics,
		CodeServiceFeatureClassification,
		CodeServiceFeatureCompletion,
		CodeServiceFeatureOutlining,
		CodeServiceFeatureQuickInfo,
		CodeServiceFeatureClusterReferences,
		CodeServiceFeatureDatabaseReferences,
		CodeServiceFeatureMinimalText,
		CodeServiceFeatureFormatting,
		CodeServiceFeatureClientParameters,
	}
	want := []string{
		"Diagnostics",
		"Classification",
		"Completion",
		"Outlining",
		"QuickInfo",
		"ClusterReferences",
		"DatabaseReferences",
		"MinimalText",
		"Formatting",
		"ClientParameters",
	}
	for i := range want {
		if features[i] != want[i] {
			t.Fatalf("feature[%d] = %q, want %q", i, features[i], want[i])
		}
	}
}

func TestAnalyzerInfoNormalizesAndCopiesDiagnostics(t *testing.T) {
	diagnostics := []Diagnostic{NewDiagnostic("KS000", "message")}
	info := NewAnalyzerInfo("", diagnostics)
	diagnostics[0] = diagnostics[0].WithCode("changed")

	if info.Name != "" {
		t.Fatalf("analyzer name = %q, want empty", info.Name)
	}
	if len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS000" {
		t.Fatalf("analyzer diagnostics should be copied, got %#v", info.Diagnostics)
	}
	empty := NewAnalyzerInfo("none", nil)
	if empty.Diagnostics == nil || len(empty.Diagnostics) != 0 {
		t.Fatalf("empty analyzer diagnostics = %#v, want non-nil empty", empty.Diagnostics)
	}
}

func TestClientParameterReferencesAndString(t *testing.T) {
	index := 3
	parameter := NewClientParameterAt("tenant", &index, 5, 6)
	index = 9
	if parameter.Name != "tenant" || parameter.Index == nil || *parameter.Index != 3 {
		t.Fatalf("client parameter index/name = %#v", parameter)
	}
	if parameter.Range != NewTextRange(5, 6) || parameter.Start() != 5 || parameter.Length() != 6 || parameter.End() != 11 {
		t.Fatalf("client parameter range = %#v start=%d length=%d end=%d", parameter.Range, parameter.Start(), parameter.Length(), parameter.End())
	}
	if got := parameter.String(); got != "tenant[3]" {
		t.Fatalf("indexed client parameter string = %q", got)
	}

	rangeParameter := NewClientParameterRange("workspace", NewTextRange(2, 4))
	if rangeParameter.Index != nil || rangeParameter.String() != "workspace" || rangeParameter.Start() != 2 || rangeParameter.End() != 6 {
		t.Fatalf("range client parameter = %#v string=%q", rangeParameter, rangeParameter.String())
	}

	reference := NewClientParameterReference("account", 10, 7)
	if reference.Index != nil || reference.Start() != 10 || reference.Length() != 7 || reference.String() != "account" {
		t.Fatalf("reference client parameter = %#v string=%q", reference, reference.String())
	}
}

func TestClientDirectiveArgumentPreservesNameTextAndValue(t *testing.T) {
	numeric := NewClientDirectiveArgument("limit", "42", int64(42))
	if numeric.Name != "limit" || numeric.Text != "42" || numeric.Value != int64(42) {
		t.Fatalf("numeric client directive argument = %#v", numeric)
	}

	unnamed := NewClientDirectiveArgument("", "\"value\"", "value")
	if unnamed.Name != "" || unnamed.Text != "\"value\"" || unnamed.Value != "value" {
		t.Fatalf("unnamed client directive argument = %#v", unnamed)
	}

	missing := NewClientDirectiveArgument("optional", "", nil)
	if missing.Name != "optional" || missing.Text != "" || missing.Value != nil {
		t.Fatalf("missing-value client directive argument = %#v", missing)
	}
}
