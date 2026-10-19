package kustolang

import "testing"

func TestQueryGrammarPublicNativeStatementFamilyAST(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name string
		text string
		kind SyntaxKind
	}{
		{name: "alias", text: "alias database Current = cluster('help').database('Samples')", kind: SyntaxKindAliasStatement},
		{name: "let scalar", text: "let threshold = 10", kind: SyntaxKindLetStatement},
		{name: "set", text: "set query_take_max_records = 100", kind: SyntaxKindSetOptionStatement},
		{name: "query parameters", text: "declare query_parameters (tenant:string, count:long=1)", kind: SyntaxKindQueryParametersStatement},
		{name: "restrict", text: "restrict access to (Events, cluster('help').database('*').*)", kind: SyntaxKindRestrictStatement},
		{name: "pattern", text: "declare pattern route = (tenant:string)[path:string] { ('contoso', 1).['west'] = { let n = 1; Events | take n; }; }", kind: SyntaxKindPatternStatement},
		{name: "query", text: "Events | take 1", kind: SyntaxKindExpressionStatement},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statement := parseQueryGrammarProduction(t, grammar.Statement, test.text)
			if statement.Kind() != test.kind || len(statement.ContainedDiagnostics()) != 0 {
				t.Fatalf("statement = kind:%v type:%T diagnostics:%#v", statement.Kind(), statement, statement.ContainedDiagnostics())
			}
		})
	}
}

func TestQueryGrammarPublicLetStatementGuardedBranches(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	function := parseQueryGrammarProduction(t, grammar.Statement, "let f = (x:string) { let y = x; y }").(*LetStatement)
	declaration, ok := function.Expression().(*FunctionDeclaration)
	if !ok || declaration.Parameters().Parameters().Count() != 1 || declaration.Body().Statements().Count() != 1 {
		t.Fatalf("function let expression = %#v", function.Expression())
	}

	view := parseQueryGrammarProduction(t, grammar.Statement, "let f = view () { Events }").(*LetStatement)
	viewDeclaration, ok := view.Expression().(*FunctionDeclaration)
	if !ok || viewDeclaration.ViewKeyword() == nil {
		t.Fatalf("view let expression = %#v", view.Expression())
	}

	materialized := parseQueryGrammarProduction(t, grammar.Statement, "let cached = materialize(Events | take 1)").(*LetStatement)
	materialize, ok := materialized.Expression().(*MaterializeExpression)
	if !ok || materialize.Expression().Kind() != SyntaxKindPipeExpression {
		t.Fatalf("materialize let expression = %#v", materialized.Expression())
	}

	grouped := parseQueryGrammarProduction(t, grammar.Statement, "let scope = entity_group [cluster('help'), database('Samples')]").(*LetStatement)
	if _, ok := grouped.Expression().(*EntityGroup); !ok {
		t.Fatalf("entity-group let expression = %T, want *EntityGroup", grouped.Expression())
	}
}

func TestQueryGrammarPublicSetParametersAndRestrictionDetails(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	set := parseQueryGrammarProduction(t, grammar.Statement, "set notruncation").(*SetOptionStatement)
	if set.ValueClause() != nil {
		t.Fatalf("set option value = %#v, want nil", set.ValueClause())
	}
	setWithValue := parseQueryGrammarProduction(t, grammar.Statement, "set truncationmaxsize = 1024").(*SetOptionStatement)
	if setWithValue.ValueClause() == nil || setWithValue.ValueClause().Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("set option value = %#v", setWithValue.ValueClause())
	}

	parameters := parseQueryGrammarProduction(t, grammar.Statement, "declare query_parameters (tenant:string, mode:string=access)").(*QueryParametersStatement)
	if parameters.Parameters().Count() != 2 || parameters.Parameters().At(1).Element().DefaultValue().Value().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("query parameters = %#v", parameters.Parameters())
	}

	restrict := parseQueryGrammarProduction(t, grammar.Statement, "restrict access to (Events, cluster('help').database('*').*) with (AllowManagedIdentityAuthentication=true)").(*RestrictStatement)
	if restrict.Expressions().Count() != 2 || restrict.WithClause() == nil || restrict.WithClause().Properties().Count() != 1 {
		t.Fatalf("restrict statement = expressions:%#v with:%#v", restrict.Expressions(), restrict.WithClause())
	}
	property := restrict.WithClause().Properties().At(0).Element()
	if property.Name().Text(IncludeTriviaMinimal) != "AllowManagedIdentityAuthentication" || property.Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("restrict property = %#v", property)
	}
}

func TestQueryGrammarPublicPatternDeclarationShape(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	statement := parseQueryGrammarProduction(t, grammar.Statement,
		"declare pattern route = (tenant:string, code:long)[path:string] { ('contoso', 1).['west'] = { let n = code; Events | take n; }; }").(*PatternStatement)
	declaration := statement.Pattern()
	if declaration == nil || declaration.Parameters().Count() != 2 || declaration.PathParameter() == nil || declaration.Patterns().Count() != 1 {
		t.Fatalf("pattern declaration = %#v", declaration)
	}
	match := declaration.Patterns().At(0)
	if match.ParameterValues().Expressions().Count() != 2 || match.PathValue() == nil ||
		match.PathValue().Value().Kind() != SyntaxKindStringLiteralExpression || match.Body().Statements().Count() != 1 ||
		match.Body().Expression().Kind() != SyntaxKindPipeExpression || match.SemicolonToken().Width() == 0 {
		t.Fatalf("pattern match = %#v", match)
	}

	optional := parseQueryGrammarProduction(t, grammar.Statement, "declare pattern route").(*PatternStatement)
	if optional.Pattern() != nil {
		t.Fatalf("optional pattern declaration = %#v, want nil", optional.Pattern())
	}
}

func TestQueryGrammarPublicStatementRecoveryFactories(t *testing.T) {
	grammar := QueryGrammarFrom(nil)

	alias := parseQueryGrammarProduction(t, grammar.Statement, "alias").(*AliasStatement)
	if alias.DatabaseKeyword().Width() != 0 || alias.Name().Width() != 0 || alias.Equal().Width() != 0 || alias.Expression().Width() != 0 ||
		!diagnosticsContainMessage(alias.Name().ContainedDiagnostics(), "KS006", "Missing name") ||
		!diagnosticsContainMessage(alias.Expression().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("alias recovery = %#v diagnostics:%#v", alias, alias.ContainedDiagnostics())
	}

	set := parseQueryGrammarProduction(t, grammar.Statement, "set option =").(*SetOptionStatement)
	if set.ValueClause() == nil || set.ValueClause().Expression().Width() != 0 ||
		!diagnosticsContainMessage(set.ValueClause().Expression().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("set recovery = %#v", set)
	}

	parameters := parseQueryGrammarProduction(t, grammar.Statement, "declare query_parameters ()").(*QueryParametersStatement)
	if parameters.Parameters().Count() != 1 ||
		!diagnosticsContainMessage(parameters.Parameters().At(0).Element().ContainedDiagnostics(), "KS006", "Missing parameter") {
		t.Fatalf("query-parameter recovery = %#v diagnostics:%#v", parameters.Parameters(), parameters.ContainedDiagnostics())
	}

	restrict := parseQueryGrammarProduction(t, grammar.Statement, "restrict access to ()").(*RestrictStatement)
	if restrict.Expressions().Count() != 1 ||
		!diagnosticsContainMessage(restrict.Expressions().At(0).Element().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("restriction recovery = %#v diagnostics:%#v", restrict.Expressions(), restrict.ContainedDiagnostics())
	}

	pattern := parseQueryGrammarProduction(t, grammar.Statement, "declare pattern p = () { }").(*PatternStatement)
	if pattern.Pattern().Patterns().Count() != 1 ||
		!diagnosticsContainMessage(pattern.Pattern().Patterns().At(0).Diagnostics(), "KS006", "Missing pattern match clause") {
		t.Fatalf("missing pattern match = %#v diagnostics:%#v", pattern.Pattern().Patterns(), pattern.ContainedDiagnostics())
	}
	missing := pattern.Pattern().Patterns().At(0)
	if missing.ParameterValues().Expressions().Count() != 0 || missing.PathValue() != nil || missing.Body().Expression() != nil || missing.Body().Semicolon() != nil {
		t.Fatalf("missing pattern factory shape = %#v", missing)
	}
	if diagnostics := missing.ContainedDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Message != "Missing pattern match clause" {
		t.Fatalf("missing pattern factory diagnostics = %#v", diagnostics)
	}

	path := parseQueryGrammarProduction(t, grammar.Statement, "declare pattern p = () { ('x').[] = { 1 }; }").(*PatternStatement).
		Pattern().Patterns().At(0).PathValue()
	if path == nil || path.Value().Width() != 0 || path.Value().Kind() != SyntaxKindStringLiteralExpression ||
		!diagnosticsContainMessage(path.Value().ContainedDiagnostics(), "KS006", "Missing string") {
		t.Fatalf("missing pattern path value = %#v diagnostics:%#v", path, path.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicStatementTailsAndIntegration(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tokens := Lex("let n = 1; print n")
	result := grammar.Statement.ParseValue(NewArraySource(tokens), 0)
	if result.Length <= 0 || result.Value == nil || tokens[result.Length].Kind != SyntaxKindSemicolonToken {
		t.Fatalf("statement tail = length:%d value:%#v next:%v", result.Length, result.Value, tokens[result.Length].Kind)
	}

	block := parseQueryGrammarQueryBlock(t, grammar.QueryBlock, "declare query_parameters (tenant:string); let rows = materialize(Events | take 1); rows")
	if block.Statements().Count() != 3 {
		t.Fatalf("query block statements = %#v", block.Statements())
	}
	if _, ok := block.Statements().At(0).Element().(*QueryParametersStatement); !ok {
		t.Fatalf("query block first statement = %T", block.Statements().At(0).Element())
	}

	macro := parseQueryGrammarProduction(t, grammar.MacroExpandSubQuery, "let n = 1; Events | take n")
	if macro.Count() != 2 {
		t.Fatalf("macro-expand subquery statements = %#v", macro)
	}
}

func TestQueryGrammarPublicStatementAnnotations(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	for _, tagged := range []struct {
		parser Parser[LexicalToken]
		tag    string
	}{
		{parser: grammar.Statement, tag: "<statement>"},
		{parser: grammar.aliasStatement, tag: "<alias>"},
		{parser: grammar.letStatement, tag: "<let>"},
		{parser: grammar.setOptionStatement, tag: "<set-option>"},
		{parser: grammar.declarePatternStatement, tag: "<pattern-statement>"},
		{parser: grammar.declareQueryParametersStatement, tag: "<query-parameters>"},
		{parser: grammar.restrictStatement, tag: "<restrict>"},
	} {
		if tagged.parser.Tag() != tagged.tag || !queryGrammarParserTreeHasTag(grammar.Statement, tagged.tag, 0) {
			t.Fatalf("statement parser tag = got:%q want:%q", tagged.parser.Tag(), tagged.tag)
		}
	}

	items := queryGrammarDispatchCompletionItems(grammar.Statement, "")
	for _, visible := range []string{"let", "set"} {
		if item, ok := completionItemByDisplayText(items, visible); !ok || item.Kind != CompletionKindQueryPrefix {
			t.Fatalf("statement completion missing %q: %#v", visible, completionDisplayTexts(items))
		}
	}
	for _, hidden := range []string{"alias", "declare", "restrict"} {
		if _, ok := completionItemByDisplayText(items, hidden); ok {
			t.Fatalf("statement completion exposed %q: %#v", hidden, completionDisplayTexts(items))
		}
	}

	tokens := Lex("restrict access to (")
	wantHint := CompletionHintTable | CompletionHintMaterializedView | CompletionHintExternalTable | CompletionHintGraphModel
	found := false
	for _, annotated := range FindAnnotatedParsers(NewArraySource(tokens), len(tokens)-1, grammar.Statement, 0) {
		for _, annotation := range annotated.Annotations() {
			if hint, ok := annotation.(CompletionHint); ok && hint == wantHint {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("restriction completion hint %v not found", wantHint)
	}
}
