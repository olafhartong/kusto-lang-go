package kustolang

import "testing"

func TestCurrentMaterializedViewMalformedSourceKeepsConfiguredCatalogBinding(t *testing.T) {
	view := NewMaterializedViewSymbol("MV", []*ColumnSymbol{
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("B", ScalarTypes.Long.Type()),
	}, "")
	database := NewDatabaseSymbol("db", view)
	globals := NewGlobalState(database)

	for _, test := range []struct {
		query string
		codes []string
	}{
		{`MV | project `, nil},
		{`MV | project A,`, []string{"KS006"}},
		{`MV | project A MissingTail`, []string{"KS005", "KS142"}},
		{`database("db").MV  project A`, []string{"KS005", "KS174", "KS142"}},
		{`find in (MV) where A = "needle" project A`, []string{"KS141", "KS198"}},
		{`materialized_view("MV") | project `, nil},
	} {
		code := ParseAndAnalyze(test.query, globals)
		diagnostics := code.GetDiagnostics()
		codes := make([]string, len(diagnostics))
		for i, diagnostic := range diagnostics {
			codes[i] = diagnostic.Code
			if diagnostic.Code == "KS204" || diagnostic.Code == "KS210" {
				t.Fatalf("%q lost configured materialized-view binding: %#v", test.query, diagnostics)
			}
		}
		if len(codes) != len(test.codes) {
			t.Fatalf("%q diagnostics = %v, want %v", test.query, codes, test.codes)
		}
		for i := range codes {
			if codes[i] != test.codes[i] {
				t.Fatalf("%q diagnostics = %v, want %v", test.query, codes, test.codes)
			}
		}
	}
}
