package kustolang

import "testing"

func TestConnectionInfoParsesNamedPartsCaseInsensitively(t *testing.T) {
	info := ParseConnectionInfo("Data Source=https://cluster;Initial Catalog = database ;User ID='user'")
	if info.Text() != "Data Source=https://cluster;Initial Catalog = database ;User ID='user'" {
		t.Fatalf("text = %q", info.Text())
	}
	if got := info.DataSource(); got != "https://cluster" {
		t.Fatalf("data source = %q, want https://cluster", got)
	}
	if got := info.GetPart("data source"); got != "https://cluster" {
		t.Fatalf("case-insensitive data source = %q, want https://cluster", got)
	}
	if got := info.GetPart("INITIAL CATALOG "); got != "database" {
		t.Fatalf("initial catalog = %q, want database", got)
	}
	if got := info.GetPart("user id"); got != "user" {
		t.Fatalf("user id = %q, want user", got)
	}
}

func TestConnectionInfoTreatsBareValueAsDataSource(t *testing.T) {
	info := ParseConnectionInfo("  https://cluster.kusto.windows.net ;Initial Catalog=db")
	if got := info.DataSource(); got != "https://cluster.kusto.windows.net" {
		t.Fatalf("bare data source = %q, want cluster URI", got)
	}
	if got := info.GetPart("Initial Catalog"); got != "db" {
		t.Fatalf("initial catalog after bare value = %q, want db", got)
	}
}

func TestConnectionInfoQuotedValuesPreserveSemicolonsAndSpaces(t *testing.T) {
	info := ParseConnectionInfo("Data Source= \"https://cluster;with;semi\" ;User=' name '")
	if got := info.DataSource(); got != "https://cluster;with;semi" {
		t.Fatalf("quoted data source = %q, want value with semicolons", got)
	}
	if got := info.GetPart("User"); got != " name " {
		t.Fatalf("quoted user = %q, want spaces preserved", got)
	}
}

func TestConnectionInfoUnclosedQuoteConsumesRemainder(t *testing.T) {
	info := ParseConnectionInfo("Data Source='abc;Initial Catalog=db")
	if got := info.DataSource(); got != "abc;Initial Catalog=db" {
		t.Fatalf("unclosed quoted value = %q, want remainder after quote", got)
	}
	if got := info.GetPart("Initial Catalog"); got != "" {
		t.Fatalf("unclosed quote should not parse later part, got %q", got)
	}
}

func TestConnectionInfoMissingPartAndNil(t *testing.T) {
	info := ParseConnectionInfo("Data Source=cluster")
	if got := info.GetPart("Missing"); got != "" {
		t.Fatalf("missing part = %q, want empty", got)
	}
	var nilInfo *ConnectionInfo
	if nilInfo.Text() != "" || nilInfo.DataSource() != "" || nilInfo.GetPart("x") != "" {
		t.Fatal("nil ConnectionInfo accessors should return empty strings")
	}
}

func TestConnectionInfoUsesPinnedKustoWhitespace(t *testing.T) {
	if got := ParseConnectionInfo("A=\u200b'x'").GetPart("A"); got != "x" {
		t.Fatalf("zero-width-space value = %q, want x", got)
	}
	if got := ParseConnectionInfo("A=\u2028'x'").GetPart("A"); got != "'x'" {
		t.Fatalf("line-separator value = %q, want quoted literal text", got)
	}
}
