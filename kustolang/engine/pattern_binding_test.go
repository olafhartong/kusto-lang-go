package kustolang

import "testing"

func TestBindPatternStatementBuildsPatternSymbol(t *testing.T) {
	statement := parseFirstPatternStatement(t, `declare pattern LoginPattern = (User:string, App:string) [Path:string] { ("admin", "portal").["signin"] = { let Label = User; print Value=Label }; }; print x=1`)
	info := BindPatternStatement(statement)
	if len(info.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", info.Diagnostics)
	}
	pattern, ok := info.ReferencedSymbol().(*PatternSymbol)
	if !ok {
		t.Fatalf("referenced symbol = %#v", info.ReferencedSymbol())
	}
	if pattern.SymbolName() != "LoginPattern" || pattern.SymbolKind() != SymbolKindPattern {
		t.Fatalf("pattern metadata = %q/%s", pattern.SymbolName(), pattern.SymbolKind())
	}
	if info.ResultType != &pattern.TypeSymbol {
		t.Fatalf("result type = %#v, want pattern type", info.ResultType)
	}
	parameters := pattern.Parameters()
	if len(parameters) != 2 || parameters[0].Name != "User" || parameters[1].Name != "App" {
		t.Fatalf("parameters = %#v", parameters)
	}
	if pattern.PathParameter() == nil || pattern.PathParameter().Name != "Path" {
		t.Fatalf("path parameter = %#v", pattern.PathParameter())
	}
	signatures := pattern.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("signatures = %#v", signatures)
	}
	if len(signatures[0].ArgumentValues) != 2 || signatures[0].ArgumentValues[0] != "admin" || signatures[0].ArgumentValues[1] != "portal" {
		t.Fatalf("signature values = %#v", signatures[0].ArgumentValues)
	}
	if signatures[0].PathValue != "signin" {
		t.Fatalf("path value = %q", signatures[0].PathValue)
	}
	if signatures[0].Signature() == nil || signatures[0].Signature().Body == "" {
		t.Fatalf("computed signature = %#v", signatures[0].Signature())
	}
}

func TestBindPatternStatementDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		query string
		code  string
		msg   string
	}{
		{
			name:  "value count",
			query: `declare pattern Bad = (User:string, App:string) { ("admin") = { print x=1 }; }; print x=1`,
			code:  "KS185",
			msg:   "The number of values must equal the number of parameters.",
		},
		{
			name:  "non string value",
			query: `declare pattern Bad = (User:string) { (42) = { print x=1 }; }; print x=1`,
			code:  "KS141",
			msg:   "The expression must have the type string.",
		},
		{
			name:  "non literal value",
			query: `declare pattern Bad = (User:string) { (User) = { print x=1 }; }; print x=1`,
			code:  "KS131",
			msg:   "The expression must be a literal.",
		},
		{
			name:  "path value without parameter",
			query: `declare pattern Bad = (User:string) { ("admin").["signin"] = { print x=1 }; }; print x=1`,
			code:  "KS186",
			msg:   "A path value can only be specified when a path name is part of the declaration.",
		},
		{
			name:  "missing path value",
			query: `declare pattern Bad = (User:string) [Path:string] { ("admin") = { print x=1 }; }; print x=1`,
			code:  "KS187",
			msg:   "A path value is expected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statement := parseFirstPatternStatement(t, tt.query)
			info := BindPatternStatement(statement)
			if !diagnosticsContainMessage(info.Diagnostics, tt.code, tt.msg) {
				t.Fatalf("expected %s %q, got %#v", tt.code, tt.msg, info.Diagnostics)
			}
		})
	}
}

func TestBindPatternCallMatchesExactAndPartialSignatures(t *testing.T) {
	pattern := NewPatternSymbol("LoginPattern",
		[]*Parameter{NewParameter("User", ScalarTypes.String.Type()), NewParameter("App", ScalarTypes.String.Type())},
		nil,
		[]*PatternSignature{
			NewPatternSignature([]string{"admin", "portal"}, "", "{ print Exact=1 }"),
			NewPatternSignature([]string{"admin"}, "", "{ print Partial=1 }"),
		})

	exact := BindPatternCall(pattern, []Expression{ParseExpression(`"admin"`), ParseExpression(`"portal"`)}, DefaultGlobalState)
	if len(exact.Diagnostics) != 0 {
		t.Fatalf("exact diagnostics = %#v", exact.Diagnostics)
	}
	if len(exact.MatchingSignatures) != 1 || exact.MatchingSignatures[0].ArgumentValues[1] != "portal" {
		t.Fatalf("exact signatures = %#v", exact.MatchingSignatures)
	}

	partial := BindPatternCall(pattern, []Expression{ParseExpression(`"admin"`), ParseExpression(`"missing"`)}, DefaultGlobalState)
	if len(partial.Diagnostics) != 0 {
		t.Fatalf("partial diagnostics = %#v", partial.Diagnostics)
	}
	if len(partial.MatchingSignatures) != 1 || len(partial.MatchingSignatures[0].ArgumentValues) != 1 {
		t.Fatalf("partial signatures = %#v", partial.MatchingSignatures)
	}
}

func TestBindPatternCallDiagnostics(t *testing.T) {
	pattern := NewPatternSymbol("LoginPattern",
		[]*Parameter{NewParameter("User", ScalarTypes.String.Type()), NewParameter("Count", ScalarTypes.Long.Type())},
		nil,
		[]*PatternSignature{NewPatternSignature([]string{"admin", "7"}, "", "{ print Match=1 }")})

	result := BindPatternCall(pattern, []Expression{ParseExpression(`User`), ParseExpression(`"seven"`), ParseExpression(`"extra"`)}, DefaultGlobalState)
	if !diagnosticsContainCode(result.Diagnostics, "KS118") {
		t.Fatalf("missing count diagnostic: %#v", result.Diagnostics)
	}
	if !diagnosticsContainCode(result.Diagnostics, "KS141") {
		t.Fatalf("missing type diagnostic: %#v", result.Diagnostics)
	}
	if !diagnosticsContainCode(result.Diagnostics, "KS188") {
		t.Fatalf("missing no-match diagnostic: %#v", result.Diagnostics)
	}
	if len(result.DiagnosticArgumentIndices) != len(result.Diagnostics) ||
		result.DiagnosticArgumentIndices[0] != -1 ||
		result.DiagnosticArgumentIndices[1] != 0 ||
		result.DiagnosticArgumentIndices[2] != 1 ||
		result.DiagnosticArgumentIndices[3] != -1 {
		t.Fatalf("diagnostic targets = %#v for diagnostics %#v", result.DiagnosticArgumentIndices, result.Diagnostics)
	}

	nonLiteral := BindPatternCallWithArgumentTypes(pattern,
		[]Expression{ParseExpression(`User`), ParseExpression(`7`)},
		[]*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()},
		DefaultGlobalState)
	if !diagnosticsContainCode(nonLiteral.Diagnostics, "KS131") {
		t.Fatalf("missing literal diagnostic: %#v", nonLiteral.Diagnostics)
	}
	if len(nonLiteral.DiagnosticArgumentIndices) != len(nonLiteral.Diagnostics) || nonLiteral.DiagnosticArgumentIndices[0] != 0 {
		t.Fatalf("non-literal diagnostic targets = %#v", nonLiteral.DiagnosticArgumentIndices)
	}
}

func TestBindPatternCallPathReturnSymbol(t *testing.T) {
	pattern := NewPatternSymbol("LoginPattern",
		[]*Parameter{NewParameter("User", ScalarTypes.String.Type())},
		NewParameter("Path", ScalarTypes.String.Type()),
		[]*PatternSignature{
			NewPatternSignature([]string{"admin"}, "signin", "{ print Signin=1 }", true),
			NewPatternSignature([]string{"admin"}, "portal", "{ print Portal=1 }", true),
		})

	result := BindPatternCall(pattern, []Expression{ParseExpression(`"admin"`)}, DefaultGlobalState)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
	group, ok := result.Result.(*GroupSymbol)
	if !ok {
		t.Fatalf("result = %#v, want group", result.Result)
	}
	members := group.Members()
	if len(members) != 2 || members[0].SymbolName() != "signin" || members[1].SymbolName() != "portal" {
		t.Fatalf("members = %#v", members)
	}
}

func parseFirstPatternStatement(t *testing.T, queryText string) *PatternStatement {
	t.Helper()
	query := ParseQuery(queryText)
	statements, ok := query.GetChild(1).(*SyntaxList)
	if !ok || statements.Count() == 0 {
		t.Fatalf("statements = %#v", query.GetChild(1))
	}
	first, ok := statements.At(0).(*SeparatedElement)
	if !ok {
		t.Fatalf("first statement element = %#v", statements.At(0))
	}
	statement, ok := WrapGeneratedSyntaxElement(first.Element()).(*PatternStatement)
	if !ok {
		t.Fatalf("first statement = %#v", first.Element())
	}
	return statement
}

func diagnosticsContainCode(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
