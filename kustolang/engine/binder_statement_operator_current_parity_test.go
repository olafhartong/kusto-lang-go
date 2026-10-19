package kustolang

import "testing"

func TestCurrentAssertSchemaMissingTypeDoesNotCascade(t *testing.T) {
	code := ParseAndAnalyze(`datatable(A:long)[1] | assert-schema (,, A:long) | getschema kind=csl`)
	assertKustoCodeDiagnosticCodes(t, code, "KS006")
}

func TestCurrentStandaloneAssertSchemaPublishesEmptyTableResult(t *testing.T) {
	code := ParseAndAnalyze(`assert-schema (A:string)`)
	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("assert-schema diagnostics = %#v", diagnostics)
	}
	result, ok := TableSymbolForType(code.ResultType)
	if !ok || result == nil || len(result.Columns()) != 0 || result.IsOpen() {
		t.Fatalf("assert-schema result = %#v, want closed empty table", code.ResultType)
	}
}

func TestCurrentExternalDataMalformedURIRecoveryDoesNotCascade(t *testing.T) {
	query := `externaldata(User:string, Score:long)["https://example.test/input.csv" with (format="csv", ignoreFirstRecord=true) | project User, Score`
	code := ParseAndAnalyze(query)
	assertKustoCodeDiagnosticCodes(t, code, "KS005", "KS211", "KS005")
}

func TestCurrentGetSchemaInvalidKindReportsInsteadOfPanicking(t *testing.T) {
	code := ParseAndAnalyze(`datatable(A:long)[1] | getschema kind=1`)
	assertKustoCodeDiagnosticCodes(t, code, "KS141")
	result, ok := TableSymbolForType(code.ResultType)
	if !ok || result == nil || len(result.Columns()) != 4 {
		t.Fatalf("getschema result = %#v, want default four-column schema", code.ResultType)
	}
}

func TestCurrentMissingQueryParameterTypeReportsInvalidType(t *testing.T) {
	for _, query := range []string{
		`declare query_parameters`,
		`declare query_parameters(User:string,, Limit:long=10); print User=User`,
		`declare query_parameters(User:string="guest",`,
		`let Transform=(T:(Key:long,Value:string),`,
	} {
		code := ParseAndAnalyze(query)
		if !diagnosticsContain(code.GetDiagnostics(), "KS160") {
			t.Fatalf("%q diagnostics = %#v, want KS160", query, code.GetDiagnostics())
		}
	}
}

func TestCurrentDetachedGraphToTablePublishesNoResultType(t *testing.T) {
	query := `datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target  graph-to-table edges with_source_id=SrcId with_target_id=DstId`
	code := ParseAndAnalyze(query)
	if code.ResultType != nil {
		t.Fatalf("detached graph-to-table result = %#v, want nil", code.ResultType)
	}
	assertKustoCodeDiagnosticCodes(t, code, "KS005", "KS174", "KS219")
}

func TestCurrentTrailingBracketStatementPublishesErrorResult(t *testing.T) {
	query := `datatable(s:string)['a=1 b=x'] | parse-kv s as (a:long,b:string) with (pair_delimiter=' ', kv_delimiter='=') [`
	code := ParseAndAnalyze(query)
	if code.ResultType != &Error.TypeSymbol {
		last := queryBlockExpressionAtForTest(code.Query, code.Query.Statements().Count()-1)
		t.Fatalf("result = %#v, want error; last expression = %T raw=%#v", code.ResultType, last, ExpressionRawResultType(last))
	}
}

func assertKustoCodeDiagnosticCodes(t *testing.T, code *KustoCode, want ...string) {
	t.Helper()
	diagnostics := code.GetDiagnostics()
	got := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		got[i] = diagnostic.Code
	}
	if len(got) != len(want) {
		t.Fatalf("diagnostic codes = %v, want %v; diagnostics = %#v", got, want, diagnostics)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic codes = %v, want %v; diagnostics = %#v", got, want, diagnostics)
		}
	}
}
