package kustolang

import "testing"

func TestNativeBinderPatternDeclarationPublishesResultAndLocal(t *testing.T) {
	statement := parseFirstPatternStatement(t, `declare pattern LoginPattern = (User:string, Count:long) [Path:string] { ("admin", "7").["signin"] = { print Value=User }; }; print x=1`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(statement.Root())

	declaration := statement.Pattern()
	if !declaration.IsBound() || declaration.ReferencedSymbol() != nil || len(declaration.SemanticDiagnostics()) != 0 {
		t.Fatalf("pattern declaration = bound:%t ref:%#v diagnostics:%#v", declaration.IsBound(), declaration.ReferencedSymbol(), declaration.SemanticDiagnostics())
	}
	value, _ := declaration.SyntaxNode.Payload(patternDeclarationSymbolPayload)
	pattern, ok := value.(*PatternSymbol)
	if !ok || declaration.RawResultType() != &pattern.TypeSymbol || pattern.SymbolName() != "LoginPattern" {
		t.Fatalf("pattern result = payload:%#v type:%#v", value, declaration.RawResultType())
	}
	parameters := pattern.Parameters()
	if len(parameters) != 2 || parameters[0].Name != "User" || parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || parameters[1].Name != "Count" || parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() {
		t.Fatalf("pattern parameters = %#v", parameters)
	}
	local, ok := statement.Name().ReferencedSymbol().(*VariableSymbol)
	if !ok || local.SymbolName() != "LoginPattern" || local.Type != pattern {
		t.Fatalf("pattern local = %#v", statement.Name().ReferencedSymbol())
	}
	if bodyUser := findBoundNameReference(statement, "User"); bodyUser == nil {
		t.Fatal("pattern body parameter reference was not bound")
	} else if parameter, ok := bodyUser.ReferencedSymbol().(*ParameterSymbol); !ok || parameter.SymbolName() != "User" {
		t.Fatalf("pattern body User reference = %#v", bodyUser.ReferencedSymbol())
	}
}

func TestNativeBinderPatternDeclarationDiagnosticsHaveExactLocations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		code     string
		location func(*PatternMatch) SyntaxElement
	}{
		{
			name:  "value count",
			query: `declare pattern Bad = (User:string, App:string) { ("admin") = { print x=1 }; }; print x=1`,
			code:  "KS185",
			location: func(match *PatternMatch) SyntaxElement {
				return match.ParameterValues()
			},
		},
		{
			name:  "non string value",
			query: `declare pattern Bad = (User:string) { (42) = { print x=1 }; }; print x=1`,
			code:  "KS141",
			location: func(match *PatternMatch) SyntaxElement {
				return match.ParameterValues().Expressions().At(0).Element()
			},
		},
		{
			name:  "non literal value",
			query: `declare pattern Bad = (User:string) { (User) = { print x=1 }; }; print x=1`,
			code:  "KS131",
			location: func(match *PatternMatch) SyntaxElement {
				return match.ParameterValues().Expressions().At(0).Element()
			},
		},
		{
			name:  "path without parameter",
			query: `declare pattern Bad = (User:string) { ("admin").["signin"] = { print x=1 }; }; print x=1`,
			code:  "KS186",
			location: func(match *PatternMatch) SyntaxElement {
				return match.PathValue()
			},
		},
		{
			name:  "missing path",
			query: `declare pattern Bad = (User:string) [Path:string] { ("admin") = { print x=1 }; }; print x=1`,
			code:  "KS187",
			location: func(match *PatternMatch) SyntaxElement {
				return match.EqualToken()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statement := parseFirstPatternStatement(t, test.query)
			NewBinder(NewGlobalState(nil)).TryBind(statement.Root())
			declaration := statement.Pattern()
			diagnostics := declaration.SemanticDiagnostics()
			if len(diagnostics) != 1 || diagnostics[0].Code != test.code {
				t.Fatalf("diagnostics = %#v, want %s", diagnostics, test.code)
			}
			match := declaration.Patterns().At(0)
			location := test.location(match)
			if diagnostics[0].Start != location.TextStart() || diagnostics[0].Length != location.Width() {
				t.Fatalf("diagnostic location = %d+%d, want %d+%d", diagnostics[0].Start, diagnostics[0].Length, location.TextStart(), location.Width())
			}
		})
	}
}

func TestKustoCodePatternDeclarationUsesNativeBinderDiagnostics(t *testing.T) {
	const query = `declare pattern Bad = (User:string) { (42) = { print x=1 }; }; print x=1`
	for _, analyze := range []func(string) *KustoCode{
		func(text string) *KustoCode { return ParseAndAnalyze(text, NewGlobalState(nil)) },
		func(text string) *KustoCode { return Parse(text).Analyze() },
	} {
		code := analyze(query)
		diagnostics := code.GetDiagnostics()
		var matches []Diagnostic
		for _, diagnostic := range diagnostics {
			if diagnostic.Code == "KS141" {
				matches = append(matches, diagnostic)
			}
		}
		if len(matches) != 1 || matches[0].Start != 39 || matches[0].Length != 2 {
			t.Fatalf("pattern KustoCode diagnostics = %#v", diagnostics)
		}
	}
}

func findBoundNameReference(root SyntaxElement, name string) *NameReference {
	var found *NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		if found != nil {
			return
		}
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if ok && NameReferenceSimpleName(reference) == name && reference.IsBound() {
			found = reference
		}
	})
	return found
}
