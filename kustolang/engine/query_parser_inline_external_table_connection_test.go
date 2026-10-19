package kustolang

import "testing"

func TestQueryParserInlineExternalTableExplicitEmptyConnectionsArePresent(t *testing.T) {
	query := ParseQuery(`inline_external_table(Account:string) kind=storage dataformat=csv ()`)
	if query == nil {
		t.Fatal("ParseQuery returned nil")
	}
	if diagnosticsContain(query.ContainedDiagnostics(), "KS256") {
		t.Fatalf("explicit empty connection clause reported KS256: %#v", query.ContainedDiagnostics())
	}
}
