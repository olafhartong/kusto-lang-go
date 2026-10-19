package kustolang

import "testing"

func TestQueryGrammarFromCachesByRelevantParseOptions(t *testing.T) {
	defaultGrammar := QueryGrammarFrom(nil)
	if defaultGrammar == nil || defaultGrammar != QueryGrammarFrom(DefaultGlobalState) {
		t.Fatalf("default query grammar cache = %#v / %#v", defaultGrammar, QueryGrammarFrom(DefaultGlobalState))
	}

	globals := NewGlobalState(nil).WithCache()
	first := QueryGrammarFrom(globals)
	if first == nil || first != QueryGrammarFrom(globals) {
		t.Fatalf("global query grammar cache = %#v / %#v", first, QueryGrammarFrom(globals))
	}

	parserKindOnly := globals.WithParseOptions(globals.ParseOptions.WithParserKind(ParserKindGrammar))
	if got := QueryGrammarFrom(parserKindOnly); got != first {
		t.Fatalf("parser-kind-only options produced a different grammar: %#v / %#v", first, got)
	}

	nonAdjacent := globals.WithParseOptions(globals.ParseOptions.WithAllowNonAdjacentWildcardParts(true))
	changed := QueryGrammarFrom(nonAdjacent)
	if changed == nil || changed == first {
		t.Fatalf("wildcard parse option did not produce a distinct grammar: %#v / %#v", first, changed)
	}
	if !changed.Options.AllowNonAdjacentWildcardParts {
		t.Fatalf("changed grammar options = %#v", changed.Options)
	}
}

func TestQueryGrammarPublicNameParsers(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	identifier := parseQueryGrammarName(t, grammar.IdentifierName, "Account")
	if identifier.Kind() != SyntaxKindTokenName || NameSimpleName(identifier) != "Account" {
		t.Fatalf("identifier name = kind:%v simple:%q", identifier.Kind(), NameSimpleName(identifier))
	}

	keyword := parseQueryGrammarName(t, grammar.IdentifierName, "access")
	if keyword.Kind() != SyntaxKindTokenName || NameSimpleName(keyword) != "access" {
		t.Fatalf("keyword identifier name = kind:%v simple:%q", keyword.Kind(), NameSimpleName(keyword))
	}
	reserved := grammar.IdentifierName.ParseValue(NewArraySource(Lex("project")), 0)
	if reserved.Length >= 0 || reserved.Value != nil {
		t.Fatalf("reserved project parsed as identifier name: %#v", reserved)
	}

	bracketed := parseQueryGrammarName(t, grammar.BracketedName, "['Column' ' Name']")
	if bracketed.Kind() != SyntaxKindBracketedName || NameSimpleName(bracketed) != "Column Name" {
		t.Fatalf("bracketed name = kind:%v simple:%q", bracketed.Kind(), NameSimpleName(bracketed))
	}

	for _, test := range []struct {
		text string
		want string
	}{
		{text: "{tenant}", want: "tenant"},
		{text: "{workspace[2]}", want: "workspace[2]"},
		{text: "{where[-1]}", want: "where[-1]"},
	} {
		t.Run(test.text, func(t *testing.T) {
			name := parseQueryGrammarName(t, grammar.BracedName, test.text)
			if name.Kind() != SyntaxKindBracedName || NameSimpleName(name) != test.want {
				t.Fatalf("braced name = kind:%v simple:%q, want %q", name.Kind(), NameSimpleName(name), test.want)
			}
			if got := name.Text(IncludeTriviaAll); got != test.text {
				t.Fatalf("braced name text = %q, want %q", got, test.text)
			}
		})
	}
}

func TestQueryGrammarPublicDeclarationsAndRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	plain := parseQueryGrammarDeclaration(t, grammar.SimpleNameDeclaration, "ColumnName")
	if NameDeclarationSimpleName(plain) != "ColumnName" || plain.Name().Kind() != SyntaxKindTokenName {
		t.Fatalf("plain declaration = kind:%v simple:%q", plain.Name().Kind(), NameDeclarationSimpleName(plain))
	}

	bracketed := parseQueryGrammarDeclaration(t, grammar.SimpleNameDeclaration, "['Column Name']")
	if NameDeclarationSimpleName(bracketed) != "Column Name" || bracketed.Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("bracketed declaration = kind:%v simple:%q", bracketed.Name().Kind(), NameDeclarationSimpleName(bracketed))
	}

	recovered := parseQueryGrammarDeclaration(t, grammar.BracketedNameDeclaration, "[")
	diagnostics := recovered.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing string") {
		t.Fatalf("missing declaration value diagnostics = %#v", diagnostics)
	}
	if !diagnosticsContainMessage(diagnostics, "KS005", "Expected: ]") {
		t.Fatalf("missing declaration close diagnostics = %#v", diagnostics)
	}
	name, ok := recovered.Name().(*BracketedName)
	if !ok || name.CloseBracket() == nil || name.CloseBracket().Width() != 0 {
		t.Fatalf("recovered bracketed declaration = %#v", recovered.Name())
	}

	referenceResult := grammar.SimpleNameReference.ParseValue(NewArraySource(Lex("[")), 0)
	reference, ok := referenceResult.Value.(*NameReference)
	if referenceResult.Length < 0 || !ok {
		t.Fatalf("recovered bracketed reference = %#v", referenceResult)
	}
	if diagnostics := reference.ContainedDiagnostics(); !diagnosticsContainMessage(diagnostics, "KS006", "Missing string") ||
		!diagnosticsContainMessage(diagnostics, "KS005", "Expected: ]") {
		t.Fatalf("recovered bracketed reference diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarWildcardAdjacencyOption(t *testing.T) {
	strict := QueryGrammarFrom(NewGlobalState(nil))
	strictResult := strict.WildcardedNameReference.ParseValue(NewArraySource(Lex("Sec * 2024")), 0)
	if strictResult.Length >= 0 || strictResult.Value != nil {
		t.Fatalf("strict wildcard accepted non-adjacent parts: %#v", strictResult)
	}

	options := DefaultParseOptions.WithAllowNonAdjacentWildcardParts(true)
	permissive := QueryGrammarFrom(NewGlobalState(nil).WithParseOptions(options))
	result := permissive.WildcardedNameReference.ParseValue(NewArraySource(Lex("Sec * 2024")), 0)
	reference, ok := result.Value.(*NameReference)
	if result.Length != 3 || !ok {
		t.Fatalf("permissive wildcard parse = %#v", result)
	}
	if reference.Name().Kind() != SyntaxKindWildcardedName || NameReferenceSimpleName(reference) != "Sec*2024" {
		t.Fatalf("permissive wildcard = kind:%v simple:%q", reference.Name().Kind(), NameReferenceSimpleName(reference))
	}
	if got := reference.Text(IncludeTriviaAll); got != "Sec * 2024" {
		t.Fatalf("permissive wildcard text = %q", got)
	}
	pattern := reference.Name().(*WildcardedName).Pattern()
	if pattern.RawText() != "Sec * 2024" || pattern.ValueText() != "Sec*2024" {
		t.Fatalf("permissive wildcard token = raw:%q value:%q", pattern.RawText(), pattern.ValueText())
	}

	star := permissive.WildcardedIdentifier.ParseValue(NewArraySource(Lex("*")), 0)
	if star.Length != 1 || star.Value == nil || star.Value.ValueText() != "*" {
		t.Fatalf("bare wildcard parse = %#v", star)
	}
}

func TestQueryGrammarPublicNameParsersRejectMalformedSyntax(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name   string
		text   string
		parser ResultParser[LexicalToken, Name]
	}{
		{name: "identifier", text: "123", parser: grammar.IdentifierName},
		{name: "bracketed missing close", text: "['Column Name'", parser: grammar.BracketedName},
		{name: "braced missing close", text: "{tenant", parser: grammar.BracedName},
		{name: "braced separated index", text: "{workspace [2]}", parser: grammar.BracedName},
		{name: "braced malformed index", text: "{workspace[x]}", parser: grammar.BracedName},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.parser.ParseValue(NewArraySource(Lex(test.text)), 0)
			if result.Length >= 0 || result.Value != nil {
				t.Fatalf("malformed name %q parsed: %#v", test.text, result)
			}
		})
	}

	for _, text := range []string{"Security", "Security *"} {
		result := grammar.WildcardedIdentifier.ParseValue(NewArraySource(Lex(text)), 0)
		if result.Length >= 0 || result.Value != nil {
			t.Fatalf("malformed wildcard %q parsed: %#v", text, result)
		}
	}
	bracketedWildcard := grammar.WildcardedNameReference.ParseValue(NewArraySource(Lex("[Sec*]")), 0)
	if bracketedWildcard.Length >= 0 || bracketedWildcard.Value != nil {
		t.Fatalf("public unbracketed wildcard rule accepted bracketed syntax: %#v", bracketedWildcard)
	}
}

func TestPredefinedRulesReuseQueryGrammarNameAndWildcardBehavior(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	rules := NewPredefinedRuleParsersWithGrammar(grammar)
	indexed := parsePredefined(t, rules.NameDeclaration, "{workspace[2]}").(*NameDeclaration)
	if indexed.Name().Kind() != SyntaxKindBracedName || NameDeclarationSimpleName(indexed) != "workspace[2]" {
		t.Fatalf("indexed predefined name = kind:%v simple:%q", indexed.Name().Kind(), NameDeclarationSimpleName(indexed))
	}
}

func parseQueryGrammarName(t *testing.T, parser ResultParser[LexicalToken, Name], text string) Name {
	t.Helper()
	result := parser.ParseValue(NewArraySource(Lex(text)), 0)
	if result.Length < 0 || result.Value == nil {
		t.Fatalf("parse name %q failed: %#v", text, result)
	}
	return result.Value
}

func parseQueryGrammarDeclaration(t *testing.T, parser ResultParser[LexicalToken, *NameDeclaration], text string) *NameDeclaration {
	t.Helper()
	result := parser.ParseValue(NewArraySource(Lex(text)), 0)
	if result.Length < 0 || result.Value == nil {
		t.Fatalf("parse declaration %q failed: %#v", text, result)
	}
	return result.Value
}
