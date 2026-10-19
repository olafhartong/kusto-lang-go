package kustolang

import "testing"

func TestCurrentMalformedLetOperatorRecoveryNativeSemantics(t *testing.T) {
	for _, test := range []struct {
		keyword string
		codes   []string
	}{
		{keyword: "print", codes: []string{"KS142", "KS005", "KS006", "KS198"}},
		{keyword: "project", codes: []string{"KS142", "KS005", "KS174", "KS198"}},
		{keyword: "where", codes: []string{"KS142", "KS005", "KS174", "KS006", "KS198"}},
	} {
		t.Run(test.keyword, func(t *testing.T) {
			code := ParseAndAnalyze(`let ` + test.keyword + `=(){"s"}; print x=1`)
			diagnostics := code.GetDiagnostics()
			if len(diagnostics) != len(test.codes) {
				t.Fatalf("diagnostics = %#v, want codes %#v", diagnostics, test.codes)
			}
			for i, want := range test.codes {
				if diagnostics[i].Code != want {
					t.Fatalf("diagnostic %d = %#v, want %s", i, diagnostics[i], want)
				}
			}
			result, _ := TableSymbolForType(code.ResultType)
			if test.keyword == "print" {
				if result == nil || len(result.Columns()) != 1 || result.Columns()[0].SymbolName() != "print_0" ||
					result.Columns()[0].Type != ScalarTypes.Unknown.Type() {
					t.Fatalf("print result = %#v", result)
				}
			} else if result == nil || len(result.Columns()) != 0 {
				t.Fatalf("%s result = %#v, want empty table", test.keyword, result)
			}
		})
	}
}
