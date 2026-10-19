package kustolang

import "testing"

func TestGeneratedSyntaxNodeMetadataCoverage(t *testing.T) {
	nodes := GeneratedSyntaxNodeMetadataAll()
	if len(nodes) != 226 {
		t.Fatalf("generated node metadata count = %d, want 226", len(nodes))
	}

	assertNodeChildShape(t, SyntaxKindQueryBlock, "QueryBlock", "SyntaxNode", []GeneratedSyntaxNodeChild{
		{Name: "Directives", Type: "SyntaxList<Directive>"},
		{Name: "Statements", Type: "SyntaxList<SeparatedElement<Statement>>"},
		{Name: "SkippedTokens", Type: "SkippedTokens", Optional: true},
		{Name: "EndOfQuery", Type: "SyntaxToken", Optional: true},
	})
	assertNodeChildShape(t, SyntaxKindFunctionCallExpression, "FunctionCallExpression", "Expression", []GeneratedSyntaxNodeChild{
		{Name: "Name", Type: "NameReference"},
		{Name: "ArgumentList", Type: "ExpressionList"},
	})
	assertNodeChildShape(t, SyntaxKindPipeExpression, "PipeExpression", "Expression", []GeneratedSyntaxNodeChild{
		{Name: "Expression", Type: "Expression"},
		{Name: "Bar", Type: "SyntaxToken"},
		{Name: "Operator", Type: "QueryOperator"},
	})
	assertNodeChildShape(t, SyntaxKindGraphMergeOperator, "GraphMergeOperator", "QueryOperator", []GeneratedSyntaxNodeChild{
		{Name: "GraphMergeKeyword", Type: "SyntaxToken"},
		{Name: "Graph", Type: "Expression"},
		{Name: "OnClause", Type: "JoinConditionClause", Optional: true},
	})
	assertNodeChildShape(t, SyntaxKindMakeSeriesOperator, "MakeSeriesOperator", "QueryOperator", []GeneratedSyntaxNodeChild{
		{Name: "MakeSeriesKeyword", Type: "SyntaxToken"},
		{Name: "Parameters", Type: "SyntaxList<NamedParameter>"},
		{Name: "Aggregates", Type: "SyntaxList<SeparatedElement<MakeSeriesExpression>>"},
		{Name: "OnClause", Type: "MakeSeriesOnClause"},
		{Name: "RangeClause", Type: "MakeSeriesRangeClause"},
		{Name: "ByClause", Type: "MakeSeriesByClause", Optional: true},
	})
	assertNodeChildShape(t, SyntaxKindProjectOperator, "ProjectOperator", "QueryOperator", []GeneratedSyntaxNodeChild{
		{Name: "ProjectKeyword", Type: "SyntaxToken"},
		{Name: "Expressions", Type: "SyntaxList<SeparatedElement<Expression>>"},
	})
	assertNodeChildShape(t, SyntaxKindSummarizeOperator, "SummarizeOperator", "QueryOperator", []GeneratedSyntaxNodeChild{
		{Name: "SummarizeKeyword", Type: "SyntaxToken"},
		{Name: "Parameters", Type: "SyntaxList<NamedParameter>"},
		{Name: "Aggregates", Type: "SyntaxList<SeparatedElement<Expression>>"},
		{Name: "ByClause", Type: "SummarizeByClause", Optional: true},
	})
	assertNodeChildShape(t, SyntaxKindJoinOperator, "JoinOperator", "QueryOperator", []GeneratedSyntaxNodeChild{
		{Name: "JoinKeyword", Type: "SyntaxToken"},
		{Name: "Parameters", Type: "SyntaxList<NamedParameter>"},
		{Name: "Expression", Type: "Expression"},
		{Name: "ConditionClause", Type: "JoinConditionClause", Optional: true},
	})
	assertNodeChildShape(t, SyntaxKindNamedParameter, "NamedParameter", "SyntaxNode", []GeneratedSyntaxNodeChild{
		{Name: "Name", Type: "NameDeclaration"},
		{Name: "EqualToken", Type: "SyntaxToken"},
		{Name: "Expression", Type: "Expression"},
	})
	assertNodeHintShape(t, "QueryBlock", []CompletionHint{
		CompletionHintNone,
		CompletionHintNonScalar,
		CompletionHintNone,
		CompletionHintNone,
	})
	assertNodeHintShape(t, "FunctionCallExpression", []CompletionHint{
		CompletionHintFunction,
		CompletionHintSyntax,
	})
	assertNodeHintShape(t, "ProjectOperator", []CompletionHint{
		CompletionHintKeyword,
		CompletionHintScalar,
	})
	assertNodeHintShape(t, "NamedParameter", []CompletionHint{
		CompletionHintDeclaration,
		CompletionHintSyntax,
		CompletionHintNone,
	})
	assertNodePayloadShape(t, "NameReference", []GeneratedSyntaxNodePayload{
		{Name: "Match", Type: "Kusto.Language.Symbols.SymbolMatch"},
	})
	assertNodePayloadShape(t, "NamedParameter", []GeneratedSyntaxNodePayload{
		{Name: "ExpressionHint", Type: "CompletionHint"},
	})
	assertNodePayloadShape(t, "CustomCommand", []GeneratedSyntaxNodePayload{
		{Name: "CommandKind", Type: "string"},
	})
	assertNodePayloadShape(t, "PartialCommand", []GeneratedSyntaxNodePayload{
		{Name: "CommandKinds", Type: "IReadOnlyList<string>"},
	})
}

func TestGeneratedSyntaxNodeMetadataReturnsCopies(t *testing.T) {
	node, ok := GeneratedSyntaxNodeMetadataForKind(SyntaxKindProjectOperator)
	if !ok {
		t.Fatal("ProjectOperator metadata not found")
	}
	node.Children[0].Name = "mutated"

	node, ok = GeneratedSyntaxNodeMetadataForKind(SyntaxKindProjectOperator)
	if !ok {
		t.Fatal("ProjectOperator metadata not found after mutation")
	}
	if node.Children[0].Name != "ProjectKeyword" {
		t.Fatalf("metadata was mutated through returned copy: %#v", node.Children)
	}

	node, ok = GeneratedSyntaxNodeMetadataForName("ProjectOperator")
	if !ok {
		t.Fatal("ProjectOperator metadata not found by name")
	}
	if node.Kind != SyntaxKindProjectOperator {
		t.Fatalf("ProjectOperator kind = %d", node.Kind)
	}
	if !node.KindFixed {
		t.Fatal("ProjectOperator should have a fixed kind")
	}

	node, ok = GeneratedSyntaxNodeMetadataForName("LiteralExpression")
	if !ok {
		t.Fatal("LiteralExpression metadata not found by name")
	}
	if node.KindFixed || node.Kind != SyntaxKindNone {
		t.Fatalf("LiteralExpression kind metadata = kind:%d fixed:%t", node.Kind, node.KindFixed)
	}

	node, ok = GeneratedSyntaxNodeMetadataForName("NameReference")
	if !ok {
		t.Fatal("NameReference metadata not found by name")
	}
	node.Payloads[0].Name = "mutated"
	node, ok = GeneratedSyntaxNodeMetadataForName("NameReference")
	if !ok {
		t.Fatal("NameReference metadata not found after mutation")
	}
	if node.Payloads[0].Name != "Match" {
		t.Fatalf("payload metadata was mutated through returned copy: %#v", node.Payloads)
	}
}

func TestNewGeneratedSyntaxNodeUsesUpstreamChildMetadata(t *testing.T) {
	projectKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeyword, "", SyntaxText(SyntaxKindProjectKeyword)))
	expressions := NewTypedSyntaxList("SeparatedElement<Expression>")
	project, err := NewGeneratedSyntaxNode(SyntaxKindProjectOperator, []SyntaxElement{projectKeyword, expressions})
	if err != nil {
		t.Fatalf("NewGeneratedSyntaxNode(ProjectOperator): %v", err)
	}
	if project.ChildCount() != 2 || project.GetChild(0) != projectKeyword || project.GetChild(1) != expressions {
		t.Fatalf("project children = %#v", project.Children())
	}
	if project.GetName(0) != "ProjectKeyword" || project.GetName(1) != "Expressions" {
		t.Fatalf("project child names = %q, %q", project.GetName(0), project.GetName(1))
	}
	if project.IsOptional(0) || project.IsOptional(1) || project.CompletionHint(0) != CompletionHintKeyword || project.CompletionHint(1) != CompletionHintScalar {
		t.Fatalf("project child metadata is inconsistent")
	}
}

func TestNewGeneratedSyntaxNodeValidatesOptionalChildren(t *testing.T) {
	joinKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindJoinKeyword, "", SyntaxText(SyntaxKindJoinKeyword)))
	parameters := NewTypedSyntaxList("NamedParameter")
	expression := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "T"))), SymbolMatchNone)
	join, err := NewGeneratedSyntaxNode(SyntaxKindJoinOperator, []SyntaxElement{joinKeyword, parameters, expression, nil})
	if err != nil {
		t.Fatalf("NewGeneratedSyntaxNode(JoinOperator): %v", err)
	}
	if !join.IsOptional(3) || join.GetName(3) != "ConditionClause" || join.GetChild(3) != nil {
		t.Fatalf("optional join child metadata is inconsistent")
	}
	if _, err := NewGeneratedSyntaxNode(SyntaxKindJoinOperator, []SyntaxElement{joinKeyword, nil, expression, nil}); err == nil {
		t.Fatal("expected required child validation error")
	}
	if _, err := NewGeneratedSyntaxNode(SyntaxKindJoinOperator, []SyntaxElement{joinKeyword, parameters, expression}); err == nil {
		t.Fatal("expected child count validation error")
	}
}

func TestGeneratedSyntaxNodeWrappers(t *testing.T) {
	if generatedSyntaxNodeWrapperCount != 209 {
		t.Fatalf("generated wrapper count = %d, want 209", generatedSyntaxNodeWrapperCount)
	}
	projectKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeyword, "", SyntaxText(SyntaxKindProjectKeyword)))
	expressions := NewTypedSyntaxList("SeparatedElement<Expression>")
	project, err := NewProjectOperator([]SyntaxElement{projectKeyword, expressions})
	if err != nil {
		t.Fatalf("NewProjectOperator: %v", err)
	}
	if project.GeneratedNodeName() != "ProjectOperator" || project.ProjectKeyword() != projectKeyword || project.Expressions().SyntaxList != expressions {
		t.Fatalf("project wrapper = %#v", project)
	}
	if _, ok := WrapGeneratedSyntaxNode(project.SyntaxNode).(*ProjectOperator); !ok {
		t.Fatalf("wrapped type = %T", WrapGeneratedSyntaxNode(project.SyntaxNode))
	}
	if _, err := NewProjectOperator([]SyntaxElement{nil, expressions}); err == nil {
		t.Fatal("expected required child validation error from wrapper constructor")
	}
}

func TestGeneratedTypedConstructors(t *testing.T) {
	if generatedSyntaxNodeTypedConstructorCount != generatedSyntaxNodeWrapperCount {
		t.Fatalf("typed constructor count = %d, want %d", generatedSyntaxNodeTypedConstructorCount, generatedSyntaxNodeWrapperCount)
	}
	literalToken := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'x'"))
	literal := MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, literalToken)
	if literal.Kind() != SyntaxKindStringLiteralExpression || literal.Token() != literalToken {
		t.Fatalf("typed literal = %#v", literal)
	}
	expressions := AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(literal, nil)))
	projectKeyword := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeyword, "", SyntaxText(SyntaxKindProjectKeyword)))
	project, err := NewProjectOperatorTyped(projectKeyword, expressions, DiagnosticMissingText("project"))
	if err != nil {
		t.Fatalf("NewProjectOperatorTyped: %v", err)
	}
	if project.Expressions().At(0).Element() != literal || len(project.Diagnostics()) != 1 {
		t.Fatalf("typed project = %#v", project)
	}
	if _, err := NewProjectOperatorTyped(nil, expressions); err == nil {
		t.Fatal("expected required token validation")
	}
}

func TestGeneratedNodePayloads(t *testing.T) {
	name := MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "columnName")))
	reference := MustNewNameReferenceTyped(name, SymbolMatchColumn)
	if reference.Match() != SymbolMatchColumn || WrapGeneratedSyntaxNode(reference.SyntaxNode).(*NameReference).Match() != SymbolMatchColumn {
		t.Fatalf("name reference payload = %v", reference.Match())
	}
	expression := MustNewLiteralExpressionTyped(SyntaxKindStringLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'x'")))
	parameter := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(cloneSyntaxElementPreservingGeneratedWrapper(name, true).(Name)),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		expression,
		CompletionHintScalar,
	)
	if parameter.ExpressionHint() != CompletionHintScalar || parameter.CompletionHint(2) != CompletionHintScalar {
		t.Fatalf("named parameter payload = %v", parameter.ExpressionHint())
	}
	command := MustNewCustomCommandTyped("show", SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindDotToken, "", ".")), cloneSyntaxElementPreservingGeneratedWrapper(reference, true))
	if command.CommandKind() != "show" {
		t.Fatalf("command kind = %q", command.CommandKind())
	}
}

func TestGeneratedTypedChildAccessors(t *testing.T) {
	name := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "f"))), SymbolMatchNone)
	argumentList := MustNewExpressionListTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "(")),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")")),
	)
	call := MustNewFunctionCallExpressionTyped(name, argumentList)
	if call.Name() != name || call.ArgumentList() != argumentList {
		t.Fatalf("function call accessors returned unexpected children")
	}
	if _, ok := any(call).(Expression); !ok {
		t.Fatal("FunctionCallExpression should implement Expression")
	}
}

func TestGeneratedGenericSyntaxListAccessors(t *testing.T) {
	directiveToken := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "#crp"))
	directives := NewTypedSyntaxList("Directive", MustNewGeneratedSyntaxNodeByName("Directive", []SyntaxElement{directiveToken}))
	skipped := NewTypedSyntaxList("SyntaxToken", SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "x")))
	block := MustNewDirectiveBlock([]SyntaxElement{directives, skipped, nil})
	if got := block.Directives().At(0); got == nil || got.Token() != directiveToken {
		t.Fatalf("typed directive element = %#v", got)
	}
	if got := block.SkippedTokens().At(0); got == nil || got.RawText() != "x" {
		t.Fatalf("typed skipped token = %#v", got)
	}
}

func TestGeneratedDynamicKindNodeWrappers(t *testing.T) {
	token := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'x'"))
	literal, err := NewLiteralExpression(SyntaxKindStringLiteralExpression, []SyntaxElement{token})
	if err != nil {
		t.Fatalf("NewLiteralExpression: %v", err)
	}
	if literal.Kind() != SyntaxKindStringLiteralExpression || literal.GeneratedNodeName() != "LiteralExpression" || literal.Token() != token {
		t.Fatalf("dynamic-kind literal = %#v", literal)
	}
	if _, err := NewGeneratedSyntaxNodeByName("LiteralExpression", []SyntaxElement{token}); err == nil {
		t.Fatal("expected explicit-kind error")
	}
}

func TestGeneratedSyntaxNodeDuplicateKindUsesNodeName(t *testing.T) {
	all := GeneratedSyntaxNodeMetadataAllForKind(SyntaxKindMaterializedViewCombineExpression)
	if len(all) != 2 || all[0].Name != "MaterializedViewCombineExpression" || all[1].Name != "MaterializedViewCombineNameClause" {
		t.Fatalf("duplicate-kind metadata = %#v", all)
	}
	open := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "("))
	close := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")"))
	value := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "ViewName"))
	nameClause := MustNewMaterializedViewCombineNameClause([]SyntaxElement{open, value, close})
	if nameClause.GeneratedNodeName() != "MaterializedViewCombineNameClause" {
		t.Fatalf("generated node name = %q", nameClause.GeneratedNodeName())
	}
	if _, ok := WrapGeneratedSyntaxNode(nameClause.SyntaxNode).(*MaterializedViewCombineNameClause); !ok {
		t.Fatalf("wrapped type = %T", WrapGeneratedSyntaxNode(nameClause.SyntaxNode))
	}
}

func assertNodeChildShape(t *testing.T, kind SyntaxKind, name, baseType string, children []GeneratedSyntaxNodeChild) {
	t.Helper()
	node, ok := GeneratedSyntaxNodeMetadataForKind(kind)
	if !ok {
		t.Fatalf("%s metadata not found", name)
	}
	if node.Name != name || node.BaseType != baseType {
		t.Fatalf("%s metadata = %#v", name, node)
	}
	if len(node.Children) != len(children) {
		t.Fatalf("%s child count = %d, want %d: %#v", name, len(node.Children), len(children), node.Children)
	}
	for i, want := range children {
		got := node.Children[i]
		if got.Name != want.Name || got.Type != want.Type || got.Optional != want.Optional {
			t.Fatalf("%s child %d = %#v, want %#v", name, i, got, want)
		}
	}
}

func assertNodeHintShape(t *testing.T, name string, hints []CompletionHint) {
	t.Helper()
	node, ok := GeneratedSyntaxNodeMetadataForName(name)
	if !ok {
		t.Fatalf("%s metadata not found", name)
	}
	if len(node.Children) != len(hints) {
		t.Fatalf("%s hint count = %d, want %d: %#v", name, len(node.Children), len(hints), node.Children)
	}
	for i, want := range hints {
		if got := node.Children[i].Hint; got != want {
			t.Fatalf("%s hint %d = %d, want %d", name, i, got, want)
		}
	}
}

func assertNodePayloadShape(t *testing.T, name string, payloads []GeneratedSyntaxNodePayload) {
	t.Helper()
	node, ok := GeneratedSyntaxNodeMetadataForName(name)
	if !ok {
		t.Fatalf("%s metadata not found", name)
	}
	if len(node.Payloads) != len(payloads) {
		t.Fatalf("%s payload count = %d, want %d: %#v", name, len(node.Payloads), len(payloads), node.Payloads)
	}
	for i, want := range payloads {
		got := node.Payloads[i]
		if got != want {
			t.Fatalf("%s payload %d = %#v, want %#v", name, i, got, want)
		}
	}
}
