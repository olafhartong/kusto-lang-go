package kustolang

import "testing"

func TestPredefinedRuleParsersLiteralAndTypeRules(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	guid := parsePredefined(t, rules.GuidLiteral, "guid(00000000-0000-0000-0000-000000000001)").(*LiteralExpression)
	if guid.Kind() != SyntaxKindGuidLiteralExpression || guid.Token().Kind() != SyntaxKindGuidLiteralToken {
		t.Fatalf("guid literal = kind:%v token:%v", guid.Kind(), guid.Token().Kind())
	}

	raw := parsePredefined(t, rules.RawGuidLiteral, "00000000-0000-0000-0000-000000000002").(*LiteralExpression)
	if raw.Kind() != SyntaxKindGuidLiteralExpression || raw.Token().Kind() != SyntaxKindRawGuidLiteralToken {
		t.Fatalf("raw guid literal = kind:%v token:%v", raw.Kind(), raw.Token().Kind())
	}

	bracketed := parsePredefined(t, rules.BracketedStringLiteral, "[storage/account/path]").(*LiteralExpression)
	if bracketed.Kind() != SyntaxKindStringLiteralExpression || bracketed.Token().RawText() != "[storage/account/path]" {
		t.Fatalf("bracketed string literal = kind:%v token:%q", bracketed.Kind(), bracketed.Token().RawText())
	}

	value := parsePredefined(t, rules.Value, "'abc'").(*LiteralExpression)
	if value.Kind() != SyntaxKindStringLiteralExpression || value.Token().ValueText() != "abc" {
		t.Fatalf("value literal = kind:%v token:%q", value.Kind(), value.Token().ValueText())
	}

	typ := parsePredefined(t, rules.Type, "dynamic")
	if typ.Kind() != SyntaxKindPrimitiveTypeExpression {
		t.Fatalf("type parser kind = %v", typ.Kind())
	}
}

func TestPredefinedRuleParsersNameDeclarationForms(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	plain := parsePredefined(t, rules.NameDeclaration, "ColumnName").(*NameDeclaration)
	if NameDeclarationSimpleName(plain) != "ColumnName" {
		t.Fatalf("plain name = %q", NameDeclarationSimpleName(plain))
	}

	quoted := parsePredefined(t, rules.NameDeclaration, "'Column Name'").(*NameDeclaration)
	if NameDeclarationSimpleName(quoted) != "Column Name" {
		t.Fatalf("quoted name = %q", NameDeclarationSimpleName(quoted))
	}

	bracketed := parsePredefined(t, rules.NameDeclaration, "['Column Name']").(*NameDeclaration)
	if NameDeclarationSimpleName(bracketed) != "Column Name" || bracketed.Name().Kind() != SyntaxKindBracketedName {
		t.Fatalf("bracketed name = simple:%q kind:%v", NameDeclarationSimpleName(bracketed), bracketed.Name().Kind())
	}

	commandBracketed := parsePredefined(t, rules.NameDeclaration, "[raw-command-name]").(*NameDeclaration)
	if NameDeclarationSimpleName(commandBracketed) != "[raw-command-name]" || commandBracketed.Name().Kind() != SyntaxKindTokenName {
		t.Fatalf("command bracketed name = simple:%q kind:%v", NameDeclarationSimpleName(commandBracketed), commandBracketed.Name().Kind())
	}

	braced := parsePredefined(t, rules.NameDeclaration, "{access}").(*NameDeclaration)
	if NameDeclarationSimpleName(braced) != "access" || braced.Name().Kind() != SyntaxKindBracedName {
		t.Fatalf("braced name = simple:%q kind:%v", NameDeclarationSimpleName(braced), braced.Name().Kind())
	}
}

func TestPredefinedRuleParsersQualifiedReferencesAndRecovery(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	table := parsePredefined(t, rules.TableNameReference, "SigninLogs").(*NameReference)
	if table.Match() != SymbolMatchTable || NameReferenceSimpleName(table) != "SigninLogs" {
		t.Fatalf("table reference = match:%v name:%q", table.Match(), NameReferenceSimpleName(table))
	}
	if completionHintFromAnnotations(rules.TableNameReference.Annotations()) != CompletionHintTable {
		t.Fatalf("table completion hint = %#v", rules.TableNameReference.Annotations())
	}

	qualified := parsePredefined(t, rules.DatabaseTableNameReference, "Samples.SigninLogs").(*PathExpression)
	left := qualified.Expression().(*NameReference)
	right := qualified.Selector().(*NameReference)
	if left.Match() != SymbolMatchDatabase|SymbolMatchTable || right.Match() != SymbolMatchTable {
		t.Fatalf("database table matches = left:%v right:%v", left.Match(), right.Match())
	}

	recovered := parsePredefined(t, rules.DatabaseTableNameReference, "Samples.").(*PathExpression)
	missing := recovered.Selector().(*NameReference)
	if !diagnosticsContain(missing.Diagnostics(), "KS006") {
		t.Fatalf("missing selector diagnostics = %#v", missing.Diagnostics())
	}

	modelOnly := parsePredefined(t, rules.GraphModelSnapshotNameReference, "GraphModel").(*NameReference)
	if modelOnly.Match() != SymbolMatchGraphModel {
		t.Fatalf("model-only match = %v", modelOnly.Match())
	}

	modelSnapshot := parsePredefined(t, rules.GraphModelSnapshotNameReference, "GraphModel.Snapshot").(*PathExpression)
	if modelSnapshot.Selector().(*NameReference).Match() != SymbolMatchGraphSnapshot {
		t.Fatalf("model snapshot selector = %#v", modelSnapshot.Selector())
	}
}

func TestPredefinedRuleParsersWildcardAndFunctionRules(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	wildcard := parsePredefined(t, rules.WildcardedNameDeclaration, "Sec*2024").(*NameDeclaration)
	if wildcard.Name().Kind() != SyntaxKindWildcardedName || NameDeclarationSimpleName(wildcard) != "Sec*2024" {
		t.Fatalf("wildcard declaration = simple:%q kind:%v", NameDeclarationSimpleName(wildcard), wildcard.Name().Kind())
	}

	qualified := parsePredefined(t, rules.QualifiedWildcardedNameDeclaration, "Db.Sec*").(*PathExpression)
	selector := qualified.Selector().(*NameDeclaration)
	if selector.Name().Kind() != SyntaxKindWildcardedName || NameDeclarationSimpleName(selector) != "Sec*" {
		t.Fatalf("qualified wildcard selector = simple:%q kind:%v", NameDeclarationSimpleName(selector), selector.Name().Kind())
	}

	body := parsePredefined(t, rules.FunctionBody, "{ let x = 1; x }").(*FunctionBody)
	if body.Statements().Count() != 1 || body.Expression() == nil {
		t.Fatalf("function body = statements:%d expression:%#v", body.Statements().Count(), body.Expression())
	}

	declaration := parsePredefined(t, rules.FunctionDeclaration, "(x:string) { x }").(*FunctionDeclaration)
	if declaration.ViewKeyword() != nil || declaration.Parameters().Parameters().Count() != 1 || declaration.Body().Expression() == nil {
		t.Fatalf("function declaration = view:%#v parameters:%d body:%#v", declaration.ViewKeyword(), declaration.Parameters().Parameters().Count(), declaration.Body())
	}
}

func TestPredefinedRuleParsersFunctionBodyForwardBarrier(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	rules := NewPredefinedRuleParsersWithGrammar(grammar)

	if rules.FunctionBody.ChildParserCount() != 1 || !rules.FunctionBody.ChildParser(0).IsForward() {
		t.Fatalf("predefined FunctionBody parser = %#v, want conversion over Forward", rules.FunctionBody)
	}
	forwardBody := rules.FunctionBody.ChildParser(0)
	if forwardBody.ChildParserCount() != 1 || forwardBody.ChildParser(0) != grammar.FunctionBody {
		t.Fatalf("predefined FunctionBody forward child = %#v, want public FunctionBody", forwardBody.ChildParser(0))
	}
	declarationRule := rules.FunctionDeclaration.ChildParser(0)
	if declarationRule.ChildParserCount() != 2 || declarationRule.ChildParser(1) != forwardBody {
		t.Fatalf("predefined FunctionDeclaration body parser = %#v, want exposed Forward", declarationRule.ChildParser(1))
	}

	const text = "{ T |"
	source := NewArraySource(Lex(text))
	direct := grammar.FunctionBody.Scan(source, 0)
	if partial := ScanPartial(rules.FunctionBody, source, 0); partial != direct {
		t.Fatalf("forwarded FunctionBody partial scan = %d, want barrier at direct scan %d", partial, direct)
	}
}

func TestPredefinedRuleParsersFunctionDeclarationRejectsViewAndRecovers(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	const viewText = "view (x:long) { x }"
	viewTokens := Lex(viewText)
	viewResult := rules.FunctionDeclaration.ParseValue(NewArraySource(viewTokens), 0)
	viewDeclaration, ok := viewResult.Value.(*FunctionDeclaration)
	if viewResult.Length < 0 || !ok {
		t.Fatalf("view rejection result = %#v", viewResult)
	}
	if viewDeclaration.ViewKeyword() != nil {
		t.Fatalf("predefined declaration view keyword = %#v, want nil", viewDeclaration.ViewKeyword())
	}
	if viewResult.Length >= len(viewTokens)-1 {
		t.Fatalf("predefined declaration consumed view form = length:%d declaration:%#v", viewResult.Length, viewDeclaration)
	}
	if len(viewDeclaration.ContainedDiagnostics()) == 0 {
		t.Fatalf("predefined declaration rejected view without recovery diagnostics = %#v", viewDeclaration)
	}

	recovered := parsePredefined(t, rules.FunctionDeclaration, "(x:long { x }").(*FunctionDeclaration)
	if recovered.ViewKeyword() != nil || recovered.Parameters().CloseParen().Width() != 0 || recovered.Body().Expression() == nil {
		t.Fatalf("recovered declaration = view:%#v parameters:%#v body:%#v", recovered.ViewKeyword(), recovered.Parameters(), recovered.Body())
	}
	if !diagnosticsContainMessage(recovered.Parameters().CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("recovered close paren diagnostics = %#v", recovered.Parameters().CloseParen().Diagnostics())
	}
}

func TestPredefinedRuleParsersInputText(t *testing.T) {
	rules := NewPredefinedRuleParsers()

	input := rules.InputText.ParseValue(NewArraySource(Lex("show version | take 1")), 0)
	token := input.Value.(*SyntaxToken)
	if input.Length != len(Lex("show version | take 1"))-1 || token.Kind() != SyntaxKindInputTextToken || token.RawText() != "show version | take 1" {
		t.Fatalf("input text = length:%d token:%#v", input.Length, token)
	}

	bracketed := rules.BracketedInputText.ParseValue(NewArraySource(Lex("storage/account ] rest")), 0)
	bracketedToken := bracketed.Value.(*SyntaxToken)
	if bracketed.Length != 3 || bracketedToken.RawText() != "storage/account" {
		t.Fatalf("bracketed input text before close = length:%d token:%q", bracketed.Length, bracketedToken.RawText())
	}

	unclosed := rules.BracketedInputText.ParseValue(NewArraySource(Lex("storage/account")), 0)
	if unclosed.Length != 3 || unclosed.Value.(*SyntaxToken).RawText() != "storage/account" {
		t.Fatalf("unclosed bracketed input text = length:%d token:%q", unclosed.Length, unclosed.Value.(*SyntaxToken).RawText())
	}
}

func parsePredefined(t *testing.T, parser ResultParser[LexicalToken, SyntaxElement], text string) SyntaxElement {
	t.Helper()
	result := parser.ParseValue(NewArraySource(Lex(text)), 0)
	if result.Length < 0 || result.Value == nil {
		t.Fatalf("parse %q failed: %#v", text, result)
	}
	return result.Value
}
