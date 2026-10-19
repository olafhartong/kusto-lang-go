package kustolang

// Ported from the public Statement/StatementList/Directive/QueryBlock parser
// fields in src/Kusto.Language/Parser/QueryGrammar.cs.

func (g *QueryGrammar) initializeQueryBlockParsers() {
	g.Statement = asResultParser[LexicalToken, Statement](FirstValue(
		g.aliasStatement,
		g.letStatement,
		g.setOptionStatement,
		g.declarePatternStatement,
		g.declareQueryParametersStatement,
		g.restrictStatement,
		g.queryStatement,
	).WithTag("<statement>"))
	g.StatementList = SyntaxSeparatedListParser(
		g.Statement,
		SyntaxKindSemicolonToken,
		func(Source[LexicalToken], int) Statement { return queryGrammarMissingStatement() },
		EndOfTextSyntaxParser(),
		false,
		true,
	)
	g.MacroExpandSubQuery = asResultParser[LexicalToken, *SeparatedSyntaxListOf[Statement]](
		Forward(func() ResultParser[LexicalToken, *SeparatedSyntaxListOf[Statement]] {
			return g.StatementList
		}).WithTag("<macro-expand-subquery>"),
	)
	g.Directive = Rule1(SyntaxTokenKindParser(SyntaxKindDirectiveToken), func(token *SyntaxToken) *Directive {
		return MustNewDirectiveTyped(token)
	})
	directives := SyntaxListParser(g.Directive, nil, false)
	g.QueryBlock = Rule4(
		directives,
		g.StatementList,
		OptionalValue(g.SkippedTokens),
		OptionalValue(SyntaxTokenKindParser(SyntaxKindEndOfTextToken)),
		func(directives *SyntaxListOf[*Directive], statements *SeparatedSyntaxListOf[Statement], skipped *SkippedTokens, end *SyntaxToken) *QueryBlock {
			return MustNewQueryBlockTyped(directives, statements, skipped, end)
		},
	)
}

func queryGrammarMissingStatement() Statement {
	expression := MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone)
	return MustNewExpressionStatementTyped(expression, DiagnosticMissingStatement())
}

func NewQueryBlockTyped(
	directives *SyntaxListOf[*Directive],
	statements *SeparatedSyntaxListOf[Statement],
	skipped *SkippedTokens,
	endOfQuery *SyntaxToken,
	diagnostics ...Diagnostic,
) (*QueryBlock, error) {
	children := []SyntaxElement{
		directives,
		syntaxElementFromSeparatedSyntaxListOf(statements),
		syntaxElementFromTyped(skipped),
		syntaxElementFromTyped(endOfQuery),
	}
	node, err := NewGeneratedSyntaxNodeByName("QueryBlock", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &QueryBlock{SyntaxNode: node}, nil
}

func MustNewQueryBlockTyped(
	directives *SyntaxListOf[*Directive],
	statements *SeparatedSyntaxListOf[Statement],
	skipped *SkippedTokens,
	endOfQuery *SyntaxToken,
	diagnostics ...Diagnostic,
) *QueryBlock {
	block, err := NewQueryBlockTyped(directives, statements, skipped, endOfQuery, diagnostics...)
	if err != nil {
		panic(err)
	}
	return block
}

func (q *QueryBlock) Directives() *SyntaxListOf[*Directive] {
	if q == nil || q.SyntaxNode == nil {
		return nil
	}
	switch list := q.GetChild(0).(type) {
	case *SyntaxListOf[*Directive]:
		return list
	case *SyntaxList:
		return AsSyntaxListOf[*Directive](list)
	default:
		return nil
	}
}

func (q *QueryBlock) Statements() *SeparatedSyntaxListOf[Statement] {
	if q == nil || q.SyntaxNode == nil {
		return nil
	}
	switch list := q.GetChild(1).(type) {
	case *SeparatedSyntaxListOf[Statement]:
		return list
	case *SyntaxList:
		return AsSeparatedSyntaxListOf[Statement](list)
	default:
		return nil
	}
}

func (q *QueryBlock) SkippedTokens() *SkippedTokens {
	if q == nil || q.SyntaxNode == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(q.GetChild(2)).(*SkippedTokens)
	return child
}

func (q *QueryBlock) EndOfQuery() *SyntaxToken {
	if q == nil || q.SyntaxNode == nil {
		return nil
	}
	child, _ := q.GetChild(3).(*SyntaxToken)
	return child
}
