package kustolang

import "testing"

func TestCurrentPatternRecoveryUsesReferenceDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		diagnostics []patternRecoveryDiagnostic
	}{
		{
			name:  "missing match has one diagnostic",
			query: `declare pattern P = () { }; print x=1`,
			diagnostics: []patternRecoveryDiagnostic{
				{"KS006", 24, 0, "Missing pattern match clause"},
			},
		},
		{
			name:  "missing values expect strings",
			query: `declare pattern P = (x:string) { (,) = { print x=1 }; }; print x=1`,
			diagnostics: []patternRecoveryDiagnostic{
				{"KS185", 33, 3, "The number of values must equal the number of parameters."},
				{"KS006", 34, 0, "Missing string"},
				{"KS006", 35, 0, "Missing string"},
			},
		},
		{
			name:  "invalid path value does not add literal diagnostic",
			query: `declare pattern Route = (tenant:string)[path:string] { ("contoso".["west"] = { print Tenant=tenant, Path=path }; }; Route("contoso").west`,
			diagnostics: []patternRecoveryDiagnostic{
				{"KS187", 75, 1, "A path value is expected."},
				{"KS142", 66, 8, "The name 'west' does not refer to any known column, table, variable or function."},
				{"KS005", 75, 1, "Expected: )"},
				{"KS188", 116, 5, "No pattern matches the specified arguments."},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := ParseAndAnalyze(test.query)
			assertPatternRecoveryDiagnostics(t, code.GetDiagnostics(), test.diagnostics...)
		})
	}

	statement := parseFirstPatternStatement(t, `declare pattern P = () { }; print x=1`)
	match := statement.Pattern().Patterns().At(0)
	tokens := []*SyntaxToken{
		match.ParameterValues().OpenParen(),
		match.ParameterValues().CloseParen(),
		match.EqualToken(),
		match.Body().OpenBrace(),
		match.Body().CloseBrace(),
		match.SemicolonToken(),
	}
	for _, token := range tokens {
		if token == nil || token.Width() != 0 || len(token.Diagnostics()) != 0 {
			t.Fatalf("synthetic token = %#v, want diagnostic-free omitted token", token)
		}
	}
}

func TestCurrentDeclarationlessPatternPublishesFallbackSymbol(t *testing.T) {
	const query = `declare pattern P; P("a")`
	code := ParseAndAnalyze(query)
	assertPatternRecoveryDiagnostics(t, code.GetDiagnostics(),
		patternRecoveryDiagnostic{"KS118", 19, 1, "No arguments expected."},
		patternRecoveryDiagnostic{"KS188", 19, 1, "No pattern matches the specified arguments."},
	)

	statement := firstPatternStatementInCode(t, code)
	local, ok := statement.Name().ReferencedSymbol().(*VariableSymbol)
	if !ok || local == nil {
		t.Fatalf("pattern local = %T %#v", statement.Name().ReferencedSymbol(), statement.Name().ReferencedSymbol())
	}
	pattern, ok := local.Type.(*PatternSymbol)
	if !ok || pattern == nil || pattern.SymbolName() != "P" || len(pattern.Parameters()) != 0 || len(pattern.Signatures()) != 0 {
		t.Fatalf("fallback pattern = %T %#v", local.Type, local.Type)
	}
}

type patternRecoveryDiagnostic struct {
	code    string
	start   int
	length  int
	message string
}

func assertPatternRecoveryDiagnostics(t *testing.T, got []Diagnostic, want ...patternRecoveryDiagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i, expected := range want {
		diagnostic := got[i]
		if diagnostic.Code != expected.code || diagnostic.Message != expected.message ||
			diagnostic.Start != expected.start || diagnostic.Length != expected.length {
			t.Fatalf("diagnostic %d = %s %q %d:%d, want %s %q %d:%d", i,
				diagnostic.Code, diagnostic.Message, diagnostic.Start, diagnostic.Length,
				expected.code, expected.message, expected.start, expected.length)
		}
	}
}
