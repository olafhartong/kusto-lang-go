package kustolang

// Code generated from the pinned upstream Kusto.Language sources; DO NOT EDIT.

const generatedSyntaxNodeWrapperCount = 209

const generatedSyntaxNodeTypedConstructorCount = 209

type SyntaxNodeElement interface {
	SyntaxElement
	syntaxNodeElement()
}

type Expression interface {
	SyntaxNodeElement
	ResultType() *TypeSymbol
	RawResultType() *TypeSymbol
	IsConstant() bool
	ConstantValueInfo() *ValueInfo
	ConstantValue() any
	IsLiteral() bool
	LiteralValueInfo() *ValueInfo
	LiteralValue() any
	isExpression()
}

type Clause interface {
	SyntaxNodeElement
	isClause()
}

type QueryOperator interface {
	SyntaxNodeElement
	isQueryOperator()
}

type Statement interface {
	SyntaxNodeElement
	isStatement()
}

type Name interface {
	SyntaxNodeElement
	isName()
}

type JsonExpression interface {
	SyntaxNodeElement
	isJsonExpression()
}

type NamedExpression interface {
	SyntaxNodeElement
	isNamedExpression()
}

type FacetWithClause interface {
	SyntaxNodeElement
	isFacetWithClause()
}

type TypeExpression interface {
	SyntaxNodeElement
	isTypeExpression()
}

type MakeSeriesRangeClause interface {
	SyntaxNodeElement
	isMakeSeriesRangeClause()
}

type PartitionOperand interface {
	SyntaxNodeElement
	isPartitionOperand()
}

type MakeGraphWithClause interface {
	SyntaxNodeElement
	isMakeGraphWithClause()
}

type GraphMatchPatternNotation interface {
	SyntaxNodeElement
	isGraphMatchPatternNotation()
}

type JoinConditionClause interface {
	SyntaxNodeElement
	isJoinConditionClause()
}

type CommandWithClause interface {
	SyntaxNodeElement
	isCommandWithClause()
}

type Command interface {
	SyntaxNodeElement
	isCommand()
}

type DirectiveBlock struct {
	*SyntaxNode
}

func NewDirectiveBlock(children []SyntaxElement, diagnostics ...Diagnostic) (*DirectiveBlock, error) {
	node, err := NewGeneratedSyntaxNodeByName("DirectiveBlock", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DirectiveBlock{SyntaxNode: node}, nil
}

func MustNewDirectiveBlock(children []SyntaxElement, diagnostics ...Diagnostic) *DirectiveBlock {
	node, err := NewDirectiveBlock(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDirectiveBlockTyped(directives *SyntaxListOf[*Directive], skippedTokens *SyntaxListOf[*SyntaxToken], endOfText *SyntaxToken, diagnostics ...Diagnostic) (*DirectiveBlock, error) {
	children := []SyntaxElement{
		syntaxElementFromSyntaxListOf(directives),
		syntaxElementFromSyntaxListOf(skippedTokens),
		syntaxElementFromTyped(endOfText),
	}
	created, err := NewDirectiveBlock(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDirectiveBlockTyped(directives *SyntaxListOf[*Directive], skippedTokens *SyntaxListOf[*SyntaxToken], endOfText *SyntaxToken, diagnostics ...Diagnostic) *DirectiveBlock {
	node, err := NewDirectiveBlockTyped(directives, skippedTokens, endOfText, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DirectiveBlock) CloneTyped(includeDiagnostics bool) *DirectiveBlock {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DirectiveBlock{SyntaxNode: cloned}
}

func (n *DirectiveBlock) Directives() *SyntaxListOf[*Directive] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*Directive](list)
}

func (n *DirectiveBlock) SkippedTokens() *SyntaxListOf[*SyntaxToken] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*SyntaxToken](list)
}

func (n *DirectiveBlock) EndOfText() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type Directive struct {
	*SyntaxNode
}

func NewDirective(children []SyntaxElement, diagnostics ...Diagnostic) (*Directive, error) {
	node, err := NewGeneratedSyntaxNodeByName("Directive", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &Directive{SyntaxNode: node}, nil
}

func MustNewDirective(children []SyntaxElement, diagnostics ...Diagnostic) *Directive {
	node, err := NewDirective(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDirectiveTyped(token *SyntaxToken, diagnostics ...Diagnostic) (*Directive, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(token),
	}
	created, err := NewDirective(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDirectiveTyped(token *SyntaxToken, diagnostics ...Diagnostic) *Directive {
	node, err := NewDirectiveTyped(token, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *Directive) CloneTyped(includeDiagnostics bool) *Directive {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &Directive{SyntaxNode: cloned}
}

func (n *Directive) Token() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SkippedTokens struct {
	*SyntaxNode
}

func NewSkippedTokens(children []SyntaxElement, diagnostics ...Diagnostic) (*SkippedTokens, error) {
	node, err := NewGeneratedSyntaxNodeByName("SkippedTokens", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SkippedTokens{SyntaxNode: node}, nil
}

func MustNewSkippedTokens(children []SyntaxElement, diagnostics ...Diagnostic) *SkippedTokens {
	node, err := NewSkippedTokens(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSkippedTokensTyped(tokens *SyntaxListOf[*SyntaxToken], diagnostics ...Diagnostic) (*SkippedTokens, error) {
	children := []SyntaxElement{
		syntaxElementFromSyntaxListOf(tokens),
	}
	created, err := NewSkippedTokens(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSkippedTokensTyped(tokens *SyntaxListOf[*SyntaxToken], diagnostics ...Diagnostic) *SkippedTokens {
	node, err := NewSkippedTokensTyped(tokens, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SkippedTokens) CloneTyped(includeDiagnostics bool) *SkippedTokens {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SkippedTokens{SyntaxNode: cloned}
}

func (n *SkippedTokens) Tokens() *SyntaxListOf[*SyntaxToken] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*SyntaxToken](list)
}

type TypeOfLiteralExpression struct {
	*SyntaxNode
}

func (*TypeOfLiteralExpression) isExpression() {}

func NewTypeOfLiteralExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*TypeOfLiteralExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("TypeOfLiteralExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TypeOfLiteralExpression{SyntaxNode: node}, nil
}

func MustNewTypeOfLiteralExpression(children []SyntaxElement, diagnostics ...Diagnostic) *TypeOfLiteralExpression {
	node, err := NewTypeOfLiteralExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTypeOfLiteralExpressionTyped(typeOfKeyword *SyntaxToken, openParen *SyntaxToken, types *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*TypeOfLiteralExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(typeOfKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(types),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewTypeOfLiteralExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTypeOfLiteralExpressionTyped(typeOfKeyword *SyntaxToken, openParen *SyntaxToken, types *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) *TypeOfLiteralExpression {
	node, err := NewTypeOfLiteralExpressionTyped(typeOfKeyword, openParen, types, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TypeOfLiteralExpression) CloneTyped(includeDiagnostics bool) *TypeOfLiteralExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TypeOfLiteralExpression{SyntaxNode: cloned}
}

func (n *TypeOfLiteralExpression) TypeOfKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TypeOfLiteralExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TypeOfLiteralExpression) Types() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *TypeOfLiteralExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type BadQueryOperator struct {
	*SyntaxNode
}

func (*BadQueryOperator) isQueryOperator() {}

func (*BadQueryOperator) isExpression() {}

func NewBadQueryOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*BadQueryOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("BadQueryOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BadQueryOperator{SyntaxNode: node}, nil
}

func MustNewBadQueryOperator(children []SyntaxElement, diagnostics ...Diagnostic) *BadQueryOperator {
	node, err := NewBadQueryOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBadQueryOperatorTyped(keyword *SyntaxToken, diagnostics ...Diagnostic) (*BadQueryOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(keyword),
	}
	created, err := NewBadQueryOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBadQueryOperatorTyped(keyword *SyntaxToken, diagnostics ...Diagnostic) *BadQueryOperator {
	node, err := NewBadQueryOperatorTyped(keyword, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BadQueryOperator) CloneTyped(includeDiagnostics bool) *BadQueryOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BadQueryOperator{SyntaxNode: cloned}
}

func (n *BadQueryOperator) Keyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type CompoundStringLiteralExpression struct {
	*SyntaxNode
}

func (*CompoundStringLiteralExpression) isExpression() {}

func NewCompoundStringLiteralExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*CompoundStringLiteralExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("CompoundStringLiteralExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CompoundStringLiteralExpression{SyntaxNode: node}, nil
}

func MustNewCompoundStringLiteralExpression(children []SyntaxElement, diagnostics ...Diagnostic) *CompoundStringLiteralExpression {
	node, err := NewCompoundStringLiteralExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCompoundStringLiteralExpressionTyped(tokens *SyntaxListOf[*SyntaxToken], diagnostics ...Diagnostic) (*CompoundStringLiteralExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromSyntaxListOf(tokens),
	}
	created, err := NewCompoundStringLiteralExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCompoundStringLiteralExpressionTyped(tokens *SyntaxListOf[*SyntaxToken], diagnostics ...Diagnostic) *CompoundStringLiteralExpression {
	node, err := NewCompoundStringLiteralExpressionTyped(tokens, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CompoundStringLiteralExpression) CloneTyped(includeDiagnostics bool) *CompoundStringLiteralExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CompoundStringLiteralExpression{SyntaxNode: cloned}
}

func (n *CompoundStringLiteralExpression) Tokens() *SyntaxListOf[*SyntaxToken] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*SyntaxToken](list)
}

type TokenName struct {
	*SyntaxNode
}

func (*TokenName) isName() {}

func NewTokenName(children []SyntaxElement, diagnostics ...Diagnostic) (*TokenName, error) {
	node, err := NewGeneratedSyntaxNodeByName("TokenName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TokenName{SyntaxNode: node}, nil
}

func MustNewTokenName(children []SyntaxElement, diagnostics ...Diagnostic) *TokenName {
	node, err := NewTokenName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTokenNameTyped(name *SyntaxToken, diagnostics ...Diagnostic) (*TokenName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
	}
	created, err := NewTokenName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTokenNameTyped(name *SyntaxToken, diagnostics ...Diagnostic) *TokenName {
	node, err := NewTokenNameTyped(name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TokenName) CloneTyped(includeDiagnostics bool) *TokenName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TokenName{SyntaxNode: cloned}
}

func (n *TokenName) Name() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type BracketedName struct {
	*SyntaxNode
}

func (*BracketedName) isName() {}

func NewBracketedName(children []SyntaxElement, diagnostics ...Diagnostic) (*BracketedName, error) {
	node, err := NewGeneratedSyntaxNodeByName("BracketedName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BracketedName{SyntaxNode: node}, nil
}

func MustNewBracketedName(children []SyntaxElement, diagnostics ...Diagnostic) *BracketedName {
	node, err := NewBracketedName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBracketedNameTyped(openBracket *SyntaxToken, name Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*BracketedName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewBracketedName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBracketedNameTyped(openBracket *SyntaxToken, name Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) *BracketedName {
	node, err := NewBracketedNameTyped(openBracket, name, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BracketedName) CloneTyped(includeDiagnostics bool) *BracketedName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BracketedName{SyntaxNode: cloned}
}

func (n *BracketedName) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracketedName) Name() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *BracketedName) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type BracedName struct {
	*SyntaxNode
}

func (*BracedName) isName() {}

func NewBracedName(children []SyntaxElement, diagnostics ...Diagnostic) (*BracedName, error) {
	node, err := NewGeneratedSyntaxNodeByName("BracedName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BracedName{SyntaxNode: node}, nil
}

func MustNewBracedName(children []SyntaxElement, diagnostics ...Diagnostic) *BracedName {
	node, err := NewBracedName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBracedNameTyped(openBrace *SyntaxToken, name *SyntaxToken, closeBrace *SyntaxToken, diagnostics ...Diagnostic) (*BracedName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBrace),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(closeBrace),
	}
	created, err := NewBracedName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBracedNameTyped(openBrace *SyntaxToken, name *SyntaxToken, closeBrace *SyntaxToken, diagnostics ...Diagnostic) *BracedName {
	node, err := NewBracedNameTyped(openBrace, name, closeBrace, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BracedName) CloneTyped(includeDiagnostics bool) *BracedName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BracedName{SyntaxNode: cloned}
}

func (n *BracedName) OpenBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracedName) Name() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracedName) CloseBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type WildcardedName struct {
	*SyntaxNode
}

func (*WildcardedName) isName() {}

func NewWildcardedName(children []SyntaxElement, diagnostics ...Diagnostic) (*WildcardedName, error) {
	node, err := NewGeneratedSyntaxNodeByName("WildcardedName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &WildcardedName{SyntaxNode: node}, nil
}

func MustNewWildcardedName(children []SyntaxElement, diagnostics ...Diagnostic) *WildcardedName {
	node, err := NewWildcardedName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewWildcardedNameTyped(pattern *SyntaxToken, diagnostics ...Diagnostic) (*WildcardedName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(pattern),
	}
	created, err := NewWildcardedName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewWildcardedNameTyped(pattern *SyntaxToken, diagnostics ...Diagnostic) *WildcardedName {
	node, err := NewWildcardedNameTyped(pattern, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *WildcardedName) CloneTyped(includeDiagnostics bool) *WildcardedName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &WildcardedName{SyntaxNode: cloned}
}

func (n *WildcardedName) Pattern() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type BracketedWildcardedName struct {
	*SyntaxNode
}

func (*BracketedWildcardedName) isName() {}

func NewBracketedWildcardedName(children []SyntaxElement, diagnostics ...Diagnostic) (*BracketedWildcardedName, error) {
	node, err := NewGeneratedSyntaxNodeByName("BracketedWildcardedName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BracketedWildcardedName{SyntaxNode: node}, nil
}

func MustNewBracketedWildcardedName(children []SyntaxElement, diagnostics ...Diagnostic) *BracketedWildcardedName {
	node, err := NewBracketedWildcardedName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBracketedWildcardedNameTyped(openBracket *SyntaxToken, pattern *SyntaxToken, closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*BracketedWildcardedName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(pattern),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewBracketedWildcardedName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBracketedWildcardedNameTyped(openBracket *SyntaxToken, pattern *SyntaxToken, closeBracket *SyntaxToken, diagnostics ...Diagnostic) *BracketedWildcardedName {
	node, err := NewBracketedWildcardedNameTyped(openBracket, pattern, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BracketedWildcardedName) CloneTyped(includeDiagnostics bool) *BracketedWildcardedName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BracketedWildcardedName{SyntaxNode: cloned}
}

func (n *BracketedWildcardedName) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracketedWildcardedName) Pattern() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracketedWildcardedName) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type NameDeclaration struct {
	*SyntaxNode
}

func (*NameDeclaration) isExpression() {}

func NewNameDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*NameDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("NameDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NameDeclaration{SyntaxNode: node}, nil
}

func MustNewNameDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *NameDeclaration {
	node, err := NewNameDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNameDeclarationTyped(name Name, diagnostics ...Diagnostic) (*NameDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
	}
	created, err := NewNameDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewNameDeclarationTyped(name Name, diagnostics ...Diagnostic) *NameDeclaration {
	node, err := NewNameDeclarationTyped(name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NameDeclaration) CloneTyped(includeDiagnostics bool) *NameDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NameDeclaration{SyntaxNode: cloned}
}

func (n *NameDeclaration) Name() Name {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Name)
	return child
}

type NameReference struct {
	*SyntaxNode
}

func (*NameReference) isExpression() {}

func NewNameReference(children []SyntaxElement, diagnostics ...Diagnostic) (*NameReference, error) {
	node, err := NewGeneratedSyntaxNodeByName("NameReference", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NameReference{SyntaxNode: node}, nil
}

func MustNewNameReference(children []SyntaxElement, diagnostics ...Diagnostic) *NameReference {
	node, err := NewNameReference(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNameReferenceTyped(name Name, match SymbolMatch, diagnostics ...Diagnostic) (*NameReference, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
	}
	created, err := NewNameReference(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	created.SyntaxNode.setPayload("Match", match)
	return created, nil
}

func MustNewNameReferenceTyped(name Name, match SymbolMatch, diagnostics ...Diagnostic) *NameReference {
	node, err := NewNameReferenceTyped(name, match, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NameReference) CloneTyped(includeDiagnostics bool) *NameReference {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NameReference{SyntaxNode: cloned}
}

func (n *NameReference) Name() Name {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Name)
	return child
}

func (n *NameReference) Match() SymbolMatch {
	if n == nil || n.SyntaxNode == nil {
		return SymbolMatchNone
	}
	payload, ok := n.SyntaxNode.Payload("Match")
	if !ok {
		return SymbolMatchNone
	}
	value, _ := payload.(SymbolMatch)
	return value
}

type LiteralExpression struct {
	*SyntaxNode
}

func (*LiteralExpression) isExpression() {}

func NewLiteralExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*LiteralExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("LiteralExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &LiteralExpression{SyntaxNode: node}, nil
}

func MustNewLiteralExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *LiteralExpression {
	node, err := NewLiteralExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewLiteralExpressionTyped(kind SyntaxKind, token *SyntaxToken, diagnostics ...Diagnostic) (*LiteralExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(token),
	}
	created, err := NewLiteralExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewLiteralExpressionTyped(kind SyntaxKind, token *SyntaxToken, diagnostics ...Diagnostic) *LiteralExpression {
	node, err := NewLiteralExpressionTyped(kind, token, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *LiteralExpression) CloneTyped(includeDiagnostics bool) *LiteralExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &LiteralExpression{SyntaxNode: cloned}
}

func (n *LiteralExpression) Token() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type StarExpression struct {
	*SyntaxNode
}

func (*StarExpression) isExpression() {}

func NewStarExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*StarExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("StarExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &StarExpression{SyntaxNode: node}, nil
}

func MustNewStarExpression(children []SyntaxElement, diagnostics ...Diagnostic) *StarExpression {
	node, err := NewStarExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewStarExpressionTyped(asteriskToken *SyntaxToken, diagnostics ...Diagnostic) (*StarExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asteriskToken),
	}
	created, err := NewStarExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewStarExpressionTyped(asteriskToken *SyntaxToken, diagnostics ...Diagnostic) *StarExpression {
	node, err := NewStarExpressionTyped(asteriskToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *StarExpression) CloneTyped(includeDiagnostics bool) *StarExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &StarExpression{SyntaxNode: cloned}
}

func (n *StarExpression) AsteriskToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type AtExpression struct {
	*SyntaxNode
}

func (*AtExpression) isExpression() {}

func NewAtExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*AtExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("AtExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &AtExpression{SyntaxNode: node}, nil
}

func MustNewAtExpression(children []SyntaxElement, diagnostics ...Diagnostic) *AtExpression {
	node, err := NewAtExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewAtExpressionTyped(atToken *SyntaxToken, diagnostics ...Diagnostic) (*AtExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(atToken),
	}
	created, err := NewAtExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewAtExpressionTyped(atToken *SyntaxToken, diagnostics ...Diagnostic) *AtExpression {
	node, err := NewAtExpressionTyped(atToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *AtExpression) CloneTyped(includeDiagnostics bool) *AtExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &AtExpression{SyntaxNode: cloned}
}

func (n *AtExpression) AtToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type JsonPair struct {
	*SyntaxNode
}

func NewJsonPair(children []SyntaxElement, diagnostics ...Diagnostic) (*JsonPair, error) {
	node, err := NewGeneratedSyntaxNodeByName("JsonPair", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JsonPair{SyntaxNode: node}, nil
}

func MustNewJsonPair(children []SyntaxElement, diagnostics ...Diagnostic) *JsonPair {
	node, err := NewJsonPair(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJsonPairTyped(name *SyntaxToken, colon *SyntaxToken, value Expression, diagnostics ...Diagnostic) (*JsonPair, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(colon),
		syntaxElementFromTyped(value),
	}
	created, err := NewJsonPair(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJsonPairTyped(name *SyntaxToken, colon *SyntaxToken, value Expression, diagnostics ...Diagnostic) *JsonPair {
	node, err := NewJsonPairTyped(name, colon, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JsonPair) CloneTyped(includeDiagnostics bool) *JsonPair {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JsonPair{SyntaxNode: cloned}
}

func (n *JsonPair) Name() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JsonPair) Colon() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JsonPair) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type JsonObjectExpression struct {
	*SyntaxNode
}

func (*JsonObjectExpression) isJsonExpression() {}

func (*JsonObjectExpression) isExpression() {}

func NewJsonObjectExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*JsonObjectExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("JsonObjectExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JsonObjectExpression{SyntaxNode: node}, nil
}

func MustNewJsonObjectExpression(children []SyntaxElement, diagnostics ...Diagnostic) *JsonObjectExpression {
	node, err := NewJsonObjectExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJsonObjectExpressionTyped(openBrace *SyntaxToken, pairs *SeparatedSyntaxListOf[*JsonPair], closeBrace *SyntaxToken, diagnostics ...Diagnostic) (*JsonObjectExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBrace),
		syntaxElementFromSeparatedSyntaxListOf(pairs),
		syntaxElementFromTyped(closeBrace),
	}
	created, err := NewJsonObjectExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJsonObjectExpressionTyped(openBrace *SyntaxToken, pairs *SeparatedSyntaxListOf[*JsonPair], closeBrace *SyntaxToken, diagnostics ...Diagnostic) *JsonObjectExpression {
	node, err := NewJsonObjectExpressionTyped(openBrace, pairs, closeBrace, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JsonObjectExpression) CloneTyped(includeDiagnostics bool) *JsonObjectExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JsonObjectExpression{SyntaxNode: cloned}
}

func (n *JsonObjectExpression) OpenBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JsonObjectExpression) Pairs() *SeparatedSyntaxListOf[*JsonPair] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*JsonPair](list)
}

func (n *JsonObjectExpression) CloseBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type JsonArrayExpression struct {
	*SyntaxNode
}

func (*JsonArrayExpression) isJsonExpression() {}

func (*JsonArrayExpression) isExpression() {}

func NewJsonArrayExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*JsonArrayExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("JsonArrayExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JsonArrayExpression{SyntaxNode: node}, nil
}

func MustNewJsonArrayExpression(children []SyntaxElement, diagnostics ...Diagnostic) *JsonArrayExpression {
	node, err := NewJsonArrayExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJsonArrayExpressionTyped(openBracket *SyntaxToken, values *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*JsonArrayExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBracket),
		syntaxElementFromSeparatedSyntaxListOf(values),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewJsonArrayExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJsonArrayExpressionTyped(openBracket *SyntaxToken, values *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) *JsonArrayExpression {
	node, err := NewJsonArrayExpressionTyped(openBracket, values, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JsonArrayExpression) CloneTyped(includeDiagnostics bool) *JsonArrayExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JsonArrayExpression{SyntaxNode: cloned}
}

func (n *JsonArrayExpression) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JsonArrayExpression) Values() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *JsonArrayExpression) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type DynamicExpression struct {
	*SyntaxNode
}

func (*DynamicExpression) isExpression() {}

func NewDynamicExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*DynamicExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("DynamicExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DynamicExpression{SyntaxNode: node}, nil
}

func MustNewDynamicExpression(children []SyntaxElement, diagnostics ...Diagnostic) *DynamicExpression {
	node, err := NewDynamicExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDynamicExpressionTyped(dynamic *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*DynamicExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dynamic),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewDynamicExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDynamicExpressionTyped(dynamic *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *DynamicExpression {
	node, err := NewDynamicExpressionTyped(dynamic, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DynamicExpression) CloneTyped(includeDiagnostics bool) *DynamicExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DynamicExpression{SyntaxNode: cloned}
}

func (n *DynamicExpression) Dynamic() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DynamicExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DynamicExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *DynamicExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ParenthesizedExpression struct {
	*SyntaxNode
}

func (*ParenthesizedExpression) isExpression() {}

func NewParenthesizedExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ParenthesizedExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ParenthesizedExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ParenthesizedExpression{SyntaxNode: node}, nil
}

func MustNewParenthesizedExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ParenthesizedExpression {
	node, err := NewParenthesizedExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewParenthesizedExpressionTyped(openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ParenthesizedExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewParenthesizedExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewParenthesizedExpressionTyped(openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *ParenthesizedExpression {
	node, err := NewParenthesizedExpressionTyped(openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ParenthesizedExpression) CloneTyped(includeDiagnostics bool) *ParenthesizedExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ParenthesizedExpression{SyntaxNode: cloned}
}

func (n *ParenthesizedExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParenthesizedExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ParenthesizedExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExpressionList struct {
	*SyntaxNode
}

func NewExpressionList(children []SyntaxElement, diagnostics ...Diagnostic) (*ExpressionList, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExpressionList", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExpressionList{SyntaxNode: node}, nil
}

func MustNewExpressionList(children []SyntaxElement, diagnostics ...Diagnostic) *ExpressionList {
	node, err := NewExpressionList(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExpressionListTyped(openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ExpressionList, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewExpressionList(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExpressionListTyped(openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) *ExpressionList {
	node, err := NewExpressionListTyped(openParen, expressions, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExpressionList) CloneTyped(includeDiagnostics bool) *ExpressionList {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExpressionList{SyntaxNode: cloned}
}

func (n *ExpressionList) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExpressionList) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *ExpressionList) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExpressionCouple struct {
	*SyntaxNode
}

func NewExpressionCouple(children []SyntaxElement, diagnostics ...Diagnostic) (*ExpressionCouple, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExpressionCouple", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExpressionCouple{SyntaxNode: node}, nil
}

func MustNewExpressionCouple(children []SyntaxElement, diagnostics ...Diagnostic) *ExpressionCouple {
	node, err := NewExpressionCouple(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExpressionCoupleTyped(openParen *SyntaxToken, first Expression, dotDot *SyntaxToken, second Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ExpressionCouple, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(first),
		syntaxElementFromTyped(dotDot),
		syntaxElementFromTyped(second),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewExpressionCouple(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExpressionCoupleTyped(openParen *SyntaxToken, first Expression, dotDot *SyntaxToken, second Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *ExpressionCouple {
	node, err := NewExpressionCoupleTyped(openParen, first, dotDot, second, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExpressionCouple) CloneTyped(includeDiagnostics bool) *ExpressionCouple {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExpressionCouple{SyntaxNode: cloned}
}

func (n *ExpressionCouple) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExpressionCouple) First() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ExpressionCouple) DotDot() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExpressionCouple) Second() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ExpressionCouple) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PrefixUnaryExpression struct {
	*SyntaxNode
}

func (*PrefixUnaryExpression) isExpression() {}

func NewPrefixUnaryExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*PrefixUnaryExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("PrefixUnaryExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PrefixUnaryExpression{SyntaxNode: node}, nil
}

func MustNewPrefixUnaryExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *PrefixUnaryExpression {
	node, err := NewPrefixUnaryExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPrefixUnaryExpressionTyped(kind SyntaxKind, operator *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*PrefixUnaryExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(expression),
	}
	created, err := NewPrefixUnaryExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPrefixUnaryExpressionTyped(kind SyntaxKind, operator *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *PrefixUnaryExpression {
	node, err := NewPrefixUnaryExpressionTyped(kind, operator, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PrefixUnaryExpression) CloneTyped(includeDiagnostics bool) *PrefixUnaryExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PrefixUnaryExpression{SyntaxNode: cloned}
}

func (n *PrefixUnaryExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PrefixUnaryExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type BinaryExpression struct {
	*SyntaxNode
}

func (*BinaryExpression) isExpression() {}

func NewBinaryExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*BinaryExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("BinaryExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BinaryExpression{SyntaxNode: node}, nil
}

func MustNewBinaryExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *BinaryExpression {
	node, err := NewBinaryExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBinaryExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right Expression, diagnostics ...Diagnostic) (*BinaryExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(left),
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(right),
	}
	created, err := NewBinaryExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBinaryExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right Expression, diagnostics ...Diagnostic) *BinaryExpression {
	node, err := NewBinaryExpressionTyped(kind, left, operator, right, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BinaryExpression) CloneTyped(includeDiagnostics bool) *BinaryExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BinaryExpression{SyntaxNode: cloned}
}

func (n *BinaryExpression) Left() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *BinaryExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BinaryExpression) Right() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type InExpression struct {
	*SyntaxNode
}

func (*InExpression) isExpression() {}

func NewInExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*InExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("InExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InExpression{SyntaxNode: node}, nil
}

func MustNewInExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *InExpression {
	node, err := NewInExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) (*InExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(left),
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(right),
	}
	created, err := NewInExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) *InExpression {
	node, err := NewInExpressionTyped(kind, left, operator, right, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InExpression) CloneTyped(includeDiagnostics bool) *InExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InExpression{SyntaxNode: cloned}
}

func (n *InExpression) Left() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *InExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InExpression) Right() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

type HasAnyExpression struct {
	*SyntaxNode
}

func (*HasAnyExpression) isExpression() {}

func NewHasAnyExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*HasAnyExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("HasAnyExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &HasAnyExpression{SyntaxNode: node}, nil
}

func MustNewHasAnyExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *HasAnyExpression {
	node, err := NewHasAnyExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewHasAnyExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) (*HasAnyExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(left),
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(right),
	}
	created, err := NewHasAnyExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewHasAnyExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) *HasAnyExpression {
	node, err := NewHasAnyExpressionTyped(kind, left, operator, right, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *HasAnyExpression) CloneTyped(includeDiagnostics bool) *HasAnyExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &HasAnyExpression{SyntaxNode: cloned}
}

func (n *HasAnyExpression) Left() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *HasAnyExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *HasAnyExpression) Right() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

type HasAllExpression struct {
	*SyntaxNode
}

func (*HasAllExpression) isExpression() {}

func NewHasAllExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*HasAllExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("HasAllExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &HasAllExpression{SyntaxNode: node}, nil
}

func MustNewHasAllExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *HasAllExpression {
	node, err := NewHasAllExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewHasAllExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) (*HasAllExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(left),
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(right),
	}
	created, err := NewHasAllExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewHasAllExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionList, diagnostics ...Diagnostic) *HasAllExpression {
	node, err := NewHasAllExpressionTyped(kind, left, operator, right, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *HasAllExpression) CloneTyped(includeDiagnostics bool) *HasAllExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &HasAllExpression{SyntaxNode: cloned}
}

func (n *HasAllExpression) Left() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *HasAllExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *HasAllExpression) Right() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

type BetweenExpression struct {
	*SyntaxNode
}

func (*BetweenExpression) isExpression() {}

func NewBetweenExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*BetweenExpression, error) {
	node, err := NewGeneratedSyntaxNodeByNameWithKind("BetweenExpression", kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BetweenExpression{SyntaxNode: node}, nil
}

func MustNewBetweenExpression(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *BetweenExpression {
	node, err := NewBetweenExpression(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBetweenExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionCouple, diagnostics ...Diagnostic) (*BetweenExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(left),
		syntaxElementFromTyped(operator),
		syntaxElementFromTyped(right),
	}
	created, err := NewBetweenExpression(kind, children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBetweenExpressionTyped(kind SyntaxKind, left Expression, operator *SyntaxToken, right *ExpressionCouple, diagnostics ...Diagnostic) *BetweenExpression {
	node, err := NewBetweenExpressionTyped(kind, left, operator, right, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BetweenExpression) CloneTyped(includeDiagnostics bool) *BetweenExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BetweenExpression{SyntaxNode: cloned}
}

func (n *BetweenExpression) Left() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *BetweenExpression) Operator() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BetweenExpression) Right() *ExpressionCouple {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionCouple)
	return child
}

type FunctionCallExpression struct {
	*SyntaxNode
}

func (*FunctionCallExpression) isExpression() {}

func NewFunctionCallExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*FunctionCallExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("FunctionCallExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FunctionCallExpression{SyntaxNode: node}, nil
}

func MustNewFunctionCallExpression(children []SyntaxElement, diagnostics ...Diagnostic) *FunctionCallExpression {
	node, err := NewFunctionCallExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFunctionCallExpressionTyped(name *NameReference, argumentList *ExpressionList, diagnostics ...Diagnostic) (*FunctionCallExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(argumentList),
	}
	created, err := NewFunctionCallExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFunctionCallExpressionTyped(name *NameReference, argumentList *ExpressionList, diagnostics ...Diagnostic) *FunctionCallExpression {
	node, err := NewFunctionCallExpressionTyped(name, argumentList, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FunctionCallExpression) CloneTyped(includeDiagnostics bool) *FunctionCallExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FunctionCallExpression{SyntaxNode: cloned}
}

func (n *FunctionCallExpression) Name() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *FunctionCallExpression) ArgumentList() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

type ToScalarExpression struct {
	*SyntaxNode
}

func (*ToScalarExpression) isExpression() {}

func NewToScalarExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ToScalarExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ToScalarExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ToScalarExpression{SyntaxNode: node}, nil
}

func MustNewToScalarExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ToScalarExpression {
	node, err := NewToScalarExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewToScalarExpressionTyped(toScalar *SyntaxToken, kindParameter *NamedParameter, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ToScalarExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(toScalar),
		syntaxElementFromTyped(kindParameter),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewToScalarExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewToScalarExpressionTyped(toScalar *SyntaxToken, kindParameter *NamedParameter, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *ToScalarExpression {
	node, err := NewToScalarExpressionTyped(toScalar, kindParameter, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ToScalarExpression) CloneTyped(includeDiagnostics bool) *ToScalarExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ToScalarExpression{SyntaxNode: cloned}
}

func (n *ToScalarExpression) ToScalar() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ToScalarExpression) KindParameter() *NamedParameter {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NamedParameter)
	return child
}

func (n *ToScalarExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ToScalarExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ToScalarExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ToTableExpression struct {
	*SyntaxNode
}

func (*ToTableExpression) isExpression() {}

func NewToTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ToTableExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ToTableExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ToTableExpression{SyntaxNode: node}, nil
}

func MustNewToTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ToTableExpression {
	node, err := NewToTableExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewToTableExpressionTyped(toTable *SyntaxToken, kindParameter *NamedParameter, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ToTableExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(toTable),
		syntaxElementFromTyped(kindParameter),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewToTableExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewToTableExpressionTyped(toTable *SyntaxToken, kindParameter *NamedParameter, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *ToTableExpression {
	node, err := NewToTableExpressionTyped(toTable, kindParameter, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ToTableExpression) CloneTyped(includeDiagnostics bool) *ToTableExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ToTableExpression{SyntaxNode: cloned}
}

func (n *ToTableExpression) ToTable() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ToTableExpression) KindParameter() *NamedParameter {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NamedParameter)
	return child
}

func (n *ToTableExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ToTableExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ToTableExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type MaterializedViewCombineExpression struct {
	*SyntaxNode
}

func (*MaterializedViewCombineExpression) isExpression() {}

func NewMaterializedViewCombineExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MaterializedViewCombineExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MaterializedViewCombineExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MaterializedViewCombineExpression{SyntaxNode: node}, nil
}

func MustNewMaterializedViewCombineExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MaterializedViewCombineExpression {
	node, err := NewMaterializedViewCombineExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMaterializedViewCombineExpressionTyped(materializedViewCombineKeyword *SyntaxToken, viewName *MaterializedViewCombineNameClause, baseClause *MaterializedViewCombineClause, deltaClause *MaterializedViewCombineClause, aggregationsClause *MaterializedViewCombineClause, diagnostics ...Diagnostic) (*MaterializedViewCombineExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(materializedViewCombineKeyword),
		syntaxElementFromTyped(viewName),
		syntaxElementFromTyped(baseClause),
		syntaxElementFromTyped(deltaClause),
		syntaxElementFromTyped(aggregationsClause),
	}
	created, err := NewMaterializedViewCombineExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMaterializedViewCombineExpressionTyped(materializedViewCombineKeyword *SyntaxToken, viewName *MaterializedViewCombineNameClause, baseClause *MaterializedViewCombineClause, deltaClause *MaterializedViewCombineClause, aggregationsClause *MaterializedViewCombineClause, diagnostics ...Diagnostic) *MaterializedViewCombineExpression {
	node, err := NewMaterializedViewCombineExpressionTyped(materializedViewCombineKeyword, viewName, baseClause, deltaClause, aggregationsClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MaterializedViewCombineExpression) CloneTyped(includeDiagnostics bool) *MaterializedViewCombineExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MaterializedViewCombineExpression{SyntaxNode: cloned}
}

func (n *MaterializedViewCombineExpression) MaterializedViewCombineKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializedViewCombineExpression) ViewName() *MaterializedViewCombineNameClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MaterializedViewCombineNameClause)
	return child
}

func (n *MaterializedViewCombineExpression) BaseClause() *MaterializedViewCombineClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MaterializedViewCombineClause)
	return child
}

func (n *MaterializedViewCombineExpression) DeltaClause() *MaterializedViewCombineClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MaterializedViewCombineClause)
	return child
}

func (n *MaterializedViewCombineExpression) AggregationsClause() *MaterializedViewCombineClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MaterializedViewCombineClause)
	return child
}

type MaterializedViewCombineNameClause struct {
	*SyntaxNode
}

func NewMaterializedViewCombineNameClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MaterializedViewCombineNameClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MaterializedViewCombineNameClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MaterializedViewCombineNameClause{SyntaxNode: node}, nil
}

func MustNewMaterializedViewCombineNameClause(children []SyntaxElement, diagnostics ...Diagnostic) *MaterializedViewCombineNameClause {
	node, err := NewMaterializedViewCombineNameClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMaterializedViewCombineNameClauseTyped(openParen *SyntaxToken, value Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MaterializedViewCombineNameClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(value),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMaterializedViewCombineNameClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMaterializedViewCombineNameClauseTyped(openParen *SyntaxToken, value Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *MaterializedViewCombineNameClause {
	node, err := NewMaterializedViewCombineNameClauseTyped(openParen, value, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MaterializedViewCombineNameClause) CloneTyped(includeDiagnostics bool) *MaterializedViewCombineNameClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MaterializedViewCombineNameClause{SyntaxNode: cloned}
}

func (n *MaterializedViewCombineNameClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializedViewCombineNameClause) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MaterializedViewCombineNameClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type MaterializedViewCombineClause struct {
	*SyntaxNode
}

func NewMaterializedViewCombineClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MaterializedViewCombineClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MaterializedViewCombineClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MaterializedViewCombineClause{SyntaxNode: node}, nil
}

func MustNewMaterializedViewCombineClause(children []SyntaxElement, diagnostics ...Diagnostic) *MaterializedViewCombineClause {
	node, err := NewMaterializedViewCombineClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMaterializedViewCombineClauseTyped(keyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MaterializedViewCombineClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(keyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMaterializedViewCombineClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMaterializedViewCombineClauseTyped(keyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *MaterializedViewCombineClause {
	node, err := NewMaterializedViewCombineClauseTyped(keyword, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MaterializedViewCombineClause) CloneTyped(includeDiagnostics bool) *MaterializedViewCombineClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MaterializedViewCombineClause{SyntaxNode: cloned}
}

func (n *MaterializedViewCombineClause) Keyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializedViewCombineClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializedViewCombineClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MaterializedViewCombineClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SimpleNamedExpression struct {
	*SyntaxNode
}

func (*SimpleNamedExpression) isNamedExpression() {}

func (*SimpleNamedExpression) isExpression() {}

func NewSimpleNamedExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*SimpleNamedExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("SimpleNamedExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SimpleNamedExpression{SyntaxNode: node}, nil
}

func MustNewSimpleNamedExpression(children []SyntaxElement, diagnostics ...Diagnostic) *SimpleNamedExpression {
	node, err := NewSimpleNamedExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSimpleNamedExpressionTyped(name *NameDeclaration, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*SimpleNamedExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(expression),
	}
	created, err := NewSimpleNamedExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSimpleNamedExpressionTyped(name *NameDeclaration, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *SimpleNamedExpression {
	node, err := NewSimpleNamedExpressionTyped(name, equalToken, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SimpleNamedExpression) CloneTyped(includeDiagnostics bool) *SimpleNamedExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SimpleNamedExpression{SyntaxNode: cloned}
}

func (n *SimpleNamedExpression) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *SimpleNamedExpression) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SimpleNamedExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type RenameList struct {
	*SyntaxNode
}

func NewRenameList(children []SyntaxElement, diagnostics ...Diagnostic) (*RenameList, error) {
	node, err := NewGeneratedSyntaxNodeByName("RenameList", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RenameList{SyntaxNode: node}, nil
}

func MustNewRenameList(children []SyntaxElement, diagnostics ...Diagnostic) *RenameList {
	node, err := NewRenameList(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRenameListTyped(openParen *SyntaxToken, names *SeparatedSyntaxListOf[*NameDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*RenameList, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(names),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewRenameList(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRenameListTyped(openParen *SyntaxToken, names *SeparatedSyntaxListOf[*NameDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) *RenameList {
	node, err := NewRenameListTyped(openParen, names, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RenameList) CloneTyped(includeDiagnostics bool) *RenameList {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RenameList{SyntaxNode: cloned}
}

func (n *RenameList) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenameList) Names() *SeparatedSyntaxListOf[*NameDeclaration] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NameDeclaration](list)
}

func (n *RenameList) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type CompoundNamedExpression struct {
	*SyntaxNode
}

func (*CompoundNamedExpression) isNamedExpression() {}

func (*CompoundNamedExpression) isExpression() {}

func NewCompoundNamedExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*CompoundNamedExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("CompoundNamedExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CompoundNamedExpression{SyntaxNode: node}, nil
}

func MustNewCompoundNamedExpression(children []SyntaxElement, diagnostics ...Diagnostic) *CompoundNamedExpression {
	node, err := NewCompoundNamedExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCompoundNamedExpressionTyped(names *RenameList, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*CompoundNamedExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(names),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(expression),
	}
	created, err := NewCompoundNamedExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCompoundNamedExpressionTyped(names *RenameList, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *CompoundNamedExpression {
	node, err := NewCompoundNamedExpressionTyped(names, equalToken, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CompoundNamedExpression) CloneTyped(includeDiagnostics bool) *CompoundNamedExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CompoundNamedExpression{SyntaxNode: cloned}
}

func (n *CompoundNamedExpression) Names() *RenameList {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RenameList)
	return child
}

func (n *CompoundNamedExpression) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CompoundNamedExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type BracketedExpression struct {
	*SyntaxNode
}

func (*BracketedExpression) isExpression() {}

func NewBracketedExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*BracketedExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("BracketedExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BracketedExpression{SyntaxNode: node}, nil
}

func MustNewBracketedExpression(children []SyntaxElement, diagnostics ...Diagnostic) *BracketedExpression {
	node, err := NewBracketedExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBracketedExpressionTyped(openBracket *SyntaxToken, expression Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*BracketedExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewBracketedExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBracketedExpressionTyped(openBracket *SyntaxToken, expression Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) *BracketedExpression {
	node, err := NewBracketedExpressionTyped(openBracket, expression, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BracketedExpression) CloneTyped(includeDiagnostics bool) *BracketedExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BracketedExpression{SyntaxNode: cloned}
}

func (n *BracketedExpression) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *BracketedExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *BracketedExpression) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PathExpression struct {
	*SyntaxNode
}

func (*PathExpression) isExpression() {}

func NewPathExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*PathExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("PathExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PathExpression{SyntaxNode: node}, nil
}

func MustNewPathExpression(children []SyntaxElement, diagnostics ...Diagnostic) *PathExpression {
	node, err := NewPathExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPathExpressionTyped(expression Expression, dot *SyntaxToken, selector Expression, diagnostics ...Diagnostic) (*PathExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(dot),
		syntaxElementFromTyped(selector),
	}
	created, err := NewPathExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPathExpressionTyped(expression Expression, dot *SyntaxToken, selector Expression, diagnostics ...Diagnostic) *PathExpression {
	node, err := NewPathExpressionTyped(expression, dot, selector, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PathExpression) CloneTyped(includeDiagnostics bool) *PathExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PathExpression{SyntaxNode: cloned}
}

func (n *PathExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PathExpression) Dot() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PathExpression) Selector() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type ElementExpression struct {
	*SyntaxNode
}

func (*ElementExpression) isExpression() {}

func NewElementExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ElementExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ElementExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ElementExpression{SyntaxNode: node}, nil
}

func MustNewElementExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ElementExpression {
	node, err := NewElementExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewElementExpressionTyped(expression Expression, selector Expression, diagnostics ...Diagnostic) (*ElementExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(selector),
	}
	created, err := NewElementExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewElementExpressionTyped(expression Expression, selector Expression, diagnostics ...Diagnostic) *ElementExpression {
	node, err := NewElementExpressionTyped(expression, selector, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ElementExpression) CloneTyped(includeDiagnostics bool) *ElementExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ElementExpression{SyntaxNode: cloned}
}

func (n *ElementExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ElementExpression) Selector() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type PipeExpression struct {
	*SyntaxNode
}

func (*PipeExpression) isExpression() {}

func NewPipeExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*PipeExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("PipeExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PipeExpression{SyntaxNode: node}, nil
}

func MustNewPipeExpression(children []SyntaxElement, diagnostics ...Diagnostic) *PipeExpression {
	node, err := NewPipeExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPipeExpressionTyped(expression Expression, bar *SyntaxToken, operator QueryOperator, diagnostics ...Diagnostic) (*PipeExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(bar),
		syntaxElementFromTyped(operator),
	}
	created, err := NewPipeExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPipeExpressionTyped(expression Expression, bar *SyntaxToken, operator QueryOperator, diagnostics ...Diagnostic) *PipeExpression {
	node, err := NewPipeExpressionTyped(expression, bar, operator, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PipeExpression) CloneTyped(includeDiagnostics bool) *PipeExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PipeExpression{SyntaxNode: cloned}
}

func (n *PipeExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PipeExpression) Bar() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PipeExpression) Operator() QueryOperator {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(QueryOperator)
	return child
}

type RangeOperator struct {
	*SyntaxNode
}

func (*RangeOperator) isQueryOperator() {}

func (*RangeOperator) isExpression() {}

func NewRangeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*RangeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("RangeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RangeOperator{SyntaxNode: node}, nil
}

func MustNewRangeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *RangeOperator {
	node, err := NewRangeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRangeOperatorTyped(rangeToken *SyntaxToken, name *NameDeclaration, fromToken *SyntaxToken, from Expression, toToken *SyntaxToken, to Expression, stepToken *SyntaxToken, step Expression, diagnostics ...Diagnostic) (*RangeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(rangeToken),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(fromToken),
		syntaxElementFromTyped(from),
		syntaxElementFromTyped(toToken),
		syntaxElementFromTyped(to),
		syntaxElementFromTyped(stepToken),
		syntaxElementFromTyped(step),
	}
	created, err := NewRangeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRangeOperatorTyped(rangeToken *SyntaxToken, name *NameDeclaration, fromToken *SyntaxToken, from Expression, toToken *SyntaxToken, to Expression, stepToken *SyntaxToken, step Expression, diagnostics ...Diagnostic) *RangeOperator {
	node, err := NewRangeOperatorTyped(rangeToken, name, fromToken, from, toToken, to, stepToken, step, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RangeOperator) CloneTyped(includeDiagnostics bool) *RangeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RangeOperator{SyntaxNode: cloned}
}

func (n *RangeOperator) RangeToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RangeOperator) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *RangeOperator) FromToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RangeOperator) From() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *RangeOperator) ToToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RangeOperator) To() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *RangeOperator) StepToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RangeOperator) Step() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type NamedParameter struct {
	*SyntaxNode
}

func NewNamedParameter(children []SyntaxElement, diagnostics ...Diagnostic) (*NamedParameter, error) {
	node, err := NewGeneratedSyntaxNodeByName("NamedParameter", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NamedParameter{SyntaxNode: node}, nil
}

func MustNewNamedParameter(children []SyntaxElement, diagnostics ...Diagnostic) *NamedParameter {
	node, err := NewNamedParameter(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNamedParameterTyped(name *NameDeclaration, equalToken *SyntaxToken, expression Expression, expressionHint CompletionHint, diagnostics ...Diagnostic) (*NamedParameter, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(expression),
	}
	created, err := NewNamedParameter(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	created.SyntaxNode.setPayload("ExpressionHint", expressionHint)
	return created, nil
}

func MustNewNamedParameterTyped(name *NameDeclaration, equalToken *SyntaxToken, expression Expression, expressionHint CompletionHint, diagnostics ...Diagnostic) *NamedParameter {
	node, err := NewNamedParameterTyped(name, equalToken, expression, expressionHint, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NamedParameter) CloneTyped(includeDiagnostics bool) *NamedParameter {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NamedParameter{SyntaxNode: cloned}
}

func (n *NamedParameter) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *NamedParameter) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *NamedParameter) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *NamedParameter) ExpressionHint() CompletionHint {
	if n == nil || n.SyntaxNode == nil {
		return CompletionHintNone
	}
	payload, ok := n.SyntaxNode.Payload("ExpressionHint")
	if !ok {
		return CompletionHintNone
	}
	value, _ := payload.(CompletionHint)
	return value
}

type ConsumeOperator struct {
	*SyntaxNode
}

func (*ConsumeOperator) isQueryOperator() {}

func (*ConsumeOperator) isExpression() {}

func NewConsumeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ConsumeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ConsumeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ConsumeOperator{SyntaxNode: node}, nil
}

func MustNewConsumeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ConsumeOperator {
	node, err := NewConsumeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewConsumeOperatorTyped(consumeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) (*ConsumeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(consumeKeyword),
		syntaxElementFromSyntaxListOf(parameters),
	}
	created, err := NewConsumeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewConsumeOperatorTyped(consumeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) *ConsumeOperator {
	node, err := NewConsumeOperatorTyped(consumeKeyword, parameters, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ConsumeOperator) CloneTyped(includeDiagnostics bool) *ConsumeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ConsumeOperator{SyntaxNode: cloned}
}

func (n *ConsumeOperator) ConsumeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ConsumeOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

type CountOperator struct {
	*SyntaxNode
}

func (*CountOperator) isQueryOperator() {}

func (*CountOperator) isExpression() {}

func NewCountOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*CountOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("CountOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CountOperator{SyntaxNode: node}, nil
}

func MustNewCountOperator(children []SyntaxElement, diagnostics ...Diagnostic) *CountOperator {
	node, err := NewCountOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCountOperatorTyped(countKeyword *SyntaxToken, asIdentifier *CountAsIdentifierClause, diagnostics ...Diagnostic) (*CountOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(countKeyword),
		syntaxElementFromTyped(asIdentifier),
	}
	created, err := NewCountOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCountOperatorTyped(countKeyword *SyntaxToken, asIdentifier *CountAsIdentifierClause, diagnostics ...Diagnostic) *CountOperator {
	node, err := NewCountOperatorTyped(countKeyword, asIdentifier, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CountOperator) CloneTyped(includeDiagnostics bool) *CountOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CountOperator{SyntaxNode: cloned}
}

func (n *CountOperator) CountKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CountOperator) AsIdentifier() *CountAsIdentifierClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*CountAsIdentifierClause)
	return child
}

type CountAsIdentifierClause struct {
	*SyntaxNode
}

func (*CountAsIdentifierClause) isClause() {}

func NewCountAsIdentifierClause(children []SyntaxElement, diagnostics ...Diagnostic) (*CountAsIdentifierClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("CountAsIdentifierClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CountAsIdentifierClause{SyntaxNode: node}, nil
}

func MustNewCountAsIdentifierClause(children []SyntaxElement, diagnostics ...Diagnostic) *CountAsIdentifierClause {
	node, err := NewCountAsIdentifierClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCountAsIdentifierClauseTyped(asKeyword *SyntaxToken, identifier *SyntaxToken, diagnostics ...Diagnostic) (*CountAsIdentifierClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asKeyword),
		syntaxElementFromTyped(identifier),
	}
	created, err := NewCountAsIdentifierClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCountAsIdentifierClauseTyped(asKeyword *SyntaxToken, identifier *SyntaxToken, diagnostics ...Diagnostic) *CountAsIdentifierClause {
	node, err := NewCountAsIdentifierClauseTyped(asKeyword, identifier, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CountAsIdentifierClause) CloneTyped(includeDiagnostics bool) *CountAsIdentifierClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CountAsIdentifierClause{SyntaxNode: cloned}
}

func (n *CountAsIdentifierClause) AsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CountAsIdentifierClause) Identifier() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExecuteAndCacheOperator struct {
	*SyntaxNode
}

func (*ExecuteAndCacheOperator) isQueryOperator() {}

func (*ExecuteAndCacheOperator) isExpression() {}

func NewExecuteAndCacheOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ExecuteAndCacheOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExecuteAndCacheOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExecuteAndCacheOperator{SyntaxNode: node}, nil
}

func MustNewExecuteAndCacheOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ExecuteAndCacheOperator {
	node, err := NewExecuteAndCacheOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExecuteAndCacheOperatorTyped(executeAndCacheKeyword *SyntaxToken, diagnostics ...Diagnostic) (*ExecuteAndCacheOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(executeAndCacheKeyword),
	}
	created, err := NewExecuteAndCacheOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExecuteAndCacheOperatorTyped(executeAndCacheKeyword *SyntaxToken, diagnostics ...Diagnostic) *ExecuteAndCacheOperator {
	node, err := NewExecuteAndCacheOperatorTyped(executeAndCacheKeyword, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExecuteAndCacheOperator) CloneTyped(includeDiagnostics bool) *ExecuteAndCacheOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExecuteAndCacheOperator{SyntaxNode: cloned}
}

func (n *ExecuteAndCacheOperator) ExecuteAndCacheKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExtendOperator struct {
	*SyntaxNode
}

func (*ExtendOperator) isQueryOperator() {}

func (*ExtendOperator) isExpression() {}

func NewExtendOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ExtendOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExtendOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExtendOperator{SyntaxNode: node}, nil
}

func MustNewExtendOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ExtendOperator {
	node, err := NewExtendOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExtendOperatorTyped(extendKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ExtendOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(extendKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewExtendOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExtendOperatorTyped(extendKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ExtendOperator {
	node, err := NewExtendOperatorTyped(extendKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExtendOperator) CloneTyped(includeDiagnostics bool) *ExtendOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExtendOperator{SyntaxNode: cloned}
}

func (n *ExtendOperator) ExtendKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExtendOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type FacetOperator struct {
	*SyntaxNode
}

func (*FacetOperator) isQueryOperator() {}

func (*FacetOperator) isExpression() {}

func NewFacetOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*FacetOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("FacetOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FacetOperator{SyntaxNode: node}, nil
}

func MustNewFacetOperator(children []SyntaxElement, diagnostics ...Diagnostic) *FacetOperator {
	node, err := NewFacetOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFacetOperatorTyped(facetKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], withClause FacetWithClause, diagnostics ...Diagnostic) (*FacetOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(facetKeyword),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewFacetOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFacetOperatorTyped(facetKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], withClause FacetWithClause, diagnostics ...Diagnostic) *FacetOperator {
	node, err := NewFacetOperatorTyped(facetKeyword, byKeyword, expressions, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FacetOperator) CloneTyped(includeDiagnostics bool) *FacetOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FacetOperator{SyntaxNode: cloned}
}

func (n *FacetOperator) FacetKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FacetOperator) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FacetOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *FacetOperator) WithClause() FacetWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(FacetWithClause)
	return child
}

type FacetWithOperatorClause struct {
	*SyntaxNode
}

func (*FacetWithOperatorClause) isFacetWithClause() {}

func (*FacetWithOperatorClause) isClause() {}

func NewFacetWithOperatorClause(children []SyntaxElement, diagnostics ...Diagnostic) (*FacetWithOperatorClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("FacetWithOperatorClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FacetWithOperatorClause{SyntaxNode: node}, nil
}

func MustNewFacetWithOperatorClause(children []SyntaxElement, diagnostics ...Diagnostic) *FacetWithOperatorClause {
	node, err := NewFacetWithOperatorClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFacetWithOperatorClauseTyped(withKeyword *SyntaxToken, operator QueryOperator, diagnostics ...Diagnostic) (*FacetWithOperatorClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(operator),
	}
	created, err := NewFacetWithOperatorClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFacetWithOperatorClauseTyped(withKeyword *SyntaxToken, operator QueryOperator, diagnostics ...Diagnostic) *FacetWithOperatorClause {
	node, err := NewFacetWithOperatorClauseTyped(withKeyword, operator, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FacetWithOperatorClause) CloneTyped(includeDiagnostics bool) *FacetWithOperatorClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FacetWithOperatorClause{SyntaxNode: cloned}
}

func (n *FacetWithOperatorClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FacetWithOperatorClause) Operator() QueryOperator {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(QueryOperator)
	return child
}

type FacetWithExpressionClause struct {
	*SyntaxNode
}

func (*FacetWithExpressionClause) isFacetWithClause() {}

func (*FacetWithExpressionClause) isClause() {}

func NewFacetWithExpressionClause(children []SyntaxElement, diagnostics ...Diagnostic) (*FacetWithExpressionClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("FacetWithExpressionClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FacetWithExpressionClause{SyntaxNode: node}, nil
}

func MustNewFacetWithExpressionClause(children []SyntaxElement, diagnostics ...Diagnostic) *FacetWithExpressionClause {
	node, err := NewFacetWithExpressionClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFacetWithExpressionClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*FacetWithExpressionClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewFacetWithExpressionClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFacetWithExpressionClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *FacetWithExpressionClause {
	node, err := NewFacetWithExpressionClauseTyped(withKeyword, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FacetWithExpressionClause) CloneTyped(includeDiagnostics bool) *FacetWithExpressionClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FacetWithExpressionClause{SyntaxNode: cloned}
}

func (n *FacetWithExpressionClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FacetWithExpressionClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FacetWithExpressionClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *FacetWithExpressionClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type FilterOperator struct {
	*SyntaxNode
}

func (*FilterOperator) isQueryOperator() {}

func (*FilterOperator) isExpression() {}

func NewFilterOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*FilterOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("FilterOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FilterOperator{SyntaxNode: node}, nil
}

func MustNewFilterOperator(children []SyntaxElement, diagnostics ...Diagnostic) *FilterOperator {
	node, err := NewFilterOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFilterOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], condition Expression, diagnostics ...Diagnostic) (*FilterOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(keyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(condition),
	}
	created, err := NewFilterOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFilterOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], condition Expression, diagnostics ...Diagnostic) *FilterOperator {
	node, err := NewFilterOperatorTyped(keyword, parameters, condition, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FilterOperator) CloneTyped(includeDiagnostics bool) *FilterOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FilterOperator{SyntaxNode: cloned}
}

func (n *FilterOperator) Keyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FilterOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *FilterOperator) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type GetSchemaOperator struct {
	*SyntaxNode
}

func (*GetSchemaOperator) isQueryOperator() {}

func (*GetSchemaOperator) isExpression() {}

func NewGetSchemaOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GetSchemaOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GetSchemaOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GetSchemaOperator{SyntaxNode: node}, nil
}

func MustNewGetSchemaOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GetSchemaOperator {
	node, err := NewGetSchemaOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGetSchemaOperatorTyped(getSchemaKeyword *SyntaxToken, kindParameter *NamedParameter, diagnostics ...Diagnostic) (*GetSchemaOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(getSchemaKeyword),
		syntaxElementFromTyped(kindParameter),
	}
	created, err := NewGetSchemaOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGetSchemaOperatorTyped(getSchemaKeyword *SyntaxToken, kindParameter *NamedParameter, diagnostics ...Diagnostic) *GetSchemaOperator {
	node, err := NewGetSchemaOperatorTyped(getSchemaKeyword, kindParameter, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GetSchemaOperator) CloneTyped(includeDiagnostics bool) *GetSchemaOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GetSchemaOperator{SyntaxNode: cloned}
}

func (n *GetSchemaOperator) GetSchemaKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GetSchemaOperator) KindParameter() *NamedParameter {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NamedParameter)
	return child
}

type FindOperator struct {
	*SyntaxNode
}

func (*FindOperator) isQueryOperator() {}

func (*FindOperator) isExpression() {}

func NewFindOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*FindOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("FindOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FindOperator{SyntaxNode: node}, nil
}

func MustNewFindOperator(children []SyntaxElement, diagnostics ...Diagnostic) *FindOperator {
	node, err := NewFindOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFindOperatorTyped(findKeyword *SyntaxToken, dataScope *DataScopeClause, parameters *SyntaxListOf[*NamedParameter], inClause *FindInClause, whereKeyword *SyntaxToken, condition Expression, project *FindProjectClause, projectAway *FindProjectClause, diagnostics ...Diagnostic) (*FindOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(findKeyword),
		syntaxElementFromTyped(dataScope),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(inClause),
		syntaxElementFromTyped(whereKeyword),
		syntaxElementFromTyped(condition),
		syntaxElementFromTyped(project),
		syntaxElementFromTyped(projectAway),
	}
	created, err := NewFindOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFindOperatorTyped(findKeyword *SyntaxToken, dataScope *DataScopeClause, parameters *SyntaxListOf[*NamedParameter], inClause *FindInClause, whereKeyword *SyntaxToken, condition Expression, project *FindProjectClause, projectAway *FindProjectClause, diagnostics ...Diagnostic) *FindOperator {
	node, err := NewFindOperatorTyped(findKeyword, dataScope, parameters, inClause, whereKeyword, condition, project, projectAway, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FindOperator) CloneTyped(includeDiagnostics bool) *FindOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FindOperator{SyntaxNode: cloned}
}

func (n *FindOperator) FindKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FindOperator) DataScope() *DataScopeClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*DataScopeClause)
	return child
}

func (n *FindOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *FindOperator) InClause() *FindInClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FindInClause)
	return child
}

func (n *FindOperator) WhereKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FindOperator) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *FindOperator) Project() *FindProjectClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FindProjectClause)
	return child
}

func (n *FindOperator) ProjectAway() *FindProjectClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FindProjectClause)
	return child
}

type DataScopeClause struct {
	*SyntaxNode
}

func (*DataScopeClause) isClause() {}

func NewDataScopeClause(children []SyntaxElement, diagnostics ...Diagnostic) (*DataScopeClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("DataScopeClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DataScopeClause{SyntaxNode: node}, nil
}

func MustNewDataScopeClause(children []SyntaxElement, diagnostics ...Diagnostic) *DataScopeClause {
	node, err := NewDataScopeClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDataScopeClauseTyped(dataScopeKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) (*DataScopeClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dataScopeKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(value),
	}
	created, err := NewDataScopeClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDataScopeClauseTyped(dataScopeKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) *DataScopeClause {
	node, err := NewDataScopeClauseTyped(dataScopeKeyword, equalToken, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DataScopeClause) CloneTyped(includeDiagnostics bool) *DataScopeClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DataScopeClause{SyntaxNode: cloned}
}

func (n *DataScopeClause) DataScopeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DataScopeClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DataScopeClause) Value() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type TypedColumnReference struct {
	*SyntaxNode
}

func (*TypedColumnReference) isExpression() {}

func NewTypedColumnReference(children []SyntaxElement, diagnostics ...Diagnostic) (*TypedColumnReference, error) {
	node, err := NewGeneratedSyntaxNodeByName("TypedColumnReference", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TypedColumnReference{SyntaxNode: node}, nil
}

func MustNewTypedColumnReference(children []SyntaxElement, diagnostics ...Diagnostic) *TypedColumnReference {
	node, err := NewTypedColumnReference(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTypedColumnReferenceTyped(column *NameReference, colonToken *SyntaxToken, typeChild TypeExpression, diagnostics ...Diagnostic) (*TypedColumnReference, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(column),
		syntaxElementFromTyped(colonToken),
		syntaxElementFromTyped(typeChild),
	}
	created, err := NewTypedColumnReference(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTypedColumnReferenceTyped(column *NameReference, colonToken *SyntaxToken, typeChild TypeExpression, diagnostics ...Diagnostic) *TypedColumnReference {
	node, err := NewTypedColumnReferenceTyped(column, colonToken, typeChild, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TypedColumnReference) CloneTyped(includeDiagnostics bool) *TypedColumnReference {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TypedColumnReference{SyntaxNode: cloned}
}

func (n *TypedColumnReference) Column() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *TypedColumnReference) ColonToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TypedColumnReference) Type() TypeExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(TypeExpression)
	return child
}

type FindInClause struct {
	*SyntaxNode
}

func (*FindInClause) isClause() {}

func NewFindInClause(children []SyntaxElement, diagnostics ...Diagnostic) (*FindInClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("FindInClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FindInClause{SyntaxNode: node}, nil
}

func MustNewFindInClause(children []SyntaxElement, diagnostics ...Diagnostic) *FindInClause {
	node, err := NewFindInClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFindInClauseTyped(inKeyword *SyntaxToken, openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*FindInClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(inKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewFindInClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFindInClauseTyped(inKeyword *SyntaxToken, openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) *FindInClause {
	node, err := NewFindInClauseTyped(inKeyword, openParen, expressions, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FindInClause) CloneTyped(includeDiagnostics bool) *FindInClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FindInClause{SyntaxNode: cloned}
}

func (n *FindInClause) InKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FindInClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FindInClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *FindInClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type FindProjectClause struct {
	*SyntaxNode
}

func (*FindProjectClause) isClause() {}

func NewFindProjectClause(children []SyntaxElement, diagnostics ...Diagnostic) (*FindProjectClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("FindProjectClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FindProjectClause{SyntaxNode: node}, nil
}

func MustNewFindProjectClause(children []SyntaxElement, diagnostics ...Diagnostic) *FindProjectClause {
	node, err := NewFindProjectClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFindProjectClauseTyped(projectKeyword *SyntaxToken, columns *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*FindProjectClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectKeyword),
		syntaxElementFromSeparatedSyntaxListOf(columns),
	}
	created, err := NewFindProjectClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFindProjectClauseTyped(projectKeyword *SyntaxToken, columns *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *FindProjectClause {
	node, err := NewFindProjectClauseTyped(projectKeyword, columns, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FindProjectClause) CloneTyped(includeDiagnostics bool) *FindProjectClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FindProjectClause{SyntaxNode: cloned}
}

func (n *FindProjectClause) ProjectKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FindProjectClause) Columns() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type PackExpression struct {
	*SyntaxNode
}

func (*PackExpression) isExpression() {}

func NewPackExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*PackExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("PackExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PackExpression{SyntaxNode: node}, nil
}

func MustNewPackExpression(children []SyntaxElement, diagnostics ...Diagnostic) *PackExpression {
	node, err := NewPackExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPackExpressionTyped(packKeyword *SyntaxToken, openParen *SyntaxToken, asteriskToken *SyntaxToken, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*PackExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(packKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(asteriskToken),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewPackExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPackExpressionTyped(packKeyword *SyntaxToken, openParen *SyntaxToken, asteriskToken *SyntaxToken, closeParen *SyntaxToken, diagnostics ...Diagnostic) *PackExpression {
	node, err := NewPackExpressionTyped(packKeyword, openParen, asteriskToken, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PackExpression) CloneTyped(includeDiagnostics bool) *PackExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PackExpression{SyntaxNode: cloned}
}

func (n *PackExpression) PackKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PackExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PackExpression) AsteriskToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PackExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type NameAndTypeDeclaration struct {
	*SyntaxNode
}

func (*NameAndTypeDeclaration) isExpression() {}

func NewNameAndTypeDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*NameAndTypeDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("NameAndTypeDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NameAndTypeDeclaration{SyntaxNode: node}, nil
}

func MustNewNameAndTypeDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *NameAndTypeDeclaration {
	node, err := NewNameAndTypeDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNameAndTypeDeclarationTyped(name *NameDeclaration, colon *SyntaxToken, typeChild TypeExpression, diagnostics ...Diagnostic) (*NameAndTypeDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(colon),
		syntaxElementFromTyped(typeChild),
	}
	created, err := NewNameAndTypeDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewNameAndTypeDeclarationTyped(name *NameDeclaration, colon *SyntaxToken, typeChild TypeExpression, diagnostics ...Diagnostic) *NameAndTypeDeclaration {
	node, err := NewNameAndTypeDeclarationTyped(name, colon, typeChild, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NameAndTypeDeclaration) CloneTyped(includeDiagnostics bool) *NameAndTypeDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NameAndTypeDeclaration{SyntaxNode: cloned}
}

func (n *NameAndTypeDeclaration) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *NameAndTypeDeclaration) Colon() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *NameAndTypeDeclaration) Type() TypeExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(TypeExpression)
	return child
}

type PrimitiveTypeExpression struct {
	*SyntaxNode
}

func (*PrimitiveTypeExpression) isTypeExpression() {}

func (*PrimitiveTypeExpression) isExpression() {}

func NewPrimitiveTypeExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*PrimitiveTypeExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("PrimitiveTypeExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PrimitiveTypeExpression{SyntaxNode: node}, nil
}

func MustNewPrimitiveTypeExpression(children []SyntaxElement, diagnostics ...Diagnostic) *PrimitiveTypeExpression {
	node, err := NewPrimitiveTypeExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPrimitiveTypeExpressionTyped(typeChild *SyntaxToken, diagnostics ...Diagnostic) (*PrimitiveTypeExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(typeChild),
	}
	created, err := NewPrimitiveTypeExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPrimitiveTypeExpressionTyped(typeChild *SyntaxToken, diagnostics ...Diagnostic) *PrimitiveTypeExpression {
	node, err := NewPrimitiveTypeExpressionTyped(typeChild, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PrimitiveTypeExpression) CloneTyped(includeDiagnostics bool) *PrimitiveTypeExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PrimitiveTypeExpression{SyntaxNode: cloned}
}

func (n *PrimitiveTypeExpression) Type() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SearchOperator struct {
	*SyntaxNode
}

func (*SearchOperator) isQueryOperator() {}

func (*SearchOperator) isExpression() {}

func NewSearchOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SearchOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SearchOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SearchOperator{SyntaxNode: node}, nil
}

func MustNewSearchOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SearchOperator {
	node, err := NewSearchOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSearchOperatorTyped(searchKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], dataScope *DataScopeClause, inClause *FindInClause, condition Expression, diagnostics ...Diagnostic) (*SearchOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(searchKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(dataScope),
		syntaxElementFromTyped(inClause),
		syntaxElementFromTyped(condition),
	}
	created, err := NewSearchOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSearchOperatorTyped(searchKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], dataScope *DataScopeClause, inClause *FindInClause, condition Expression, diagnostics ...Diagnostic) *SearchOperator {
	node, err := NewSearchOperatorTyped(searchKeyword, parameters, dataScope, inClause, condition, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SearchOperator) CloneTyped(includeDiagnostics bool) *SearchOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SearchOperator{SyntaxNode: cloned}
}

func (n *SearchOperator) SearchKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SearchOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SearchOperator) DataScope() *DataScopeClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*DataScopeClause)
	return child
}

func (n *SearchOperator) InClause() *FindInClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FindInClause)
	return child
}

func (n *SearchOperator) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type ForkOperator struct {
	*SyntaxNode
}

func (*ForkOperator) isQueryOperator() {}

func (*ForkOperator) isExpression() {}

func NewForkOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ForkOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ForkOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ForkOperator{SyntaxNode: node}, nil
}

func MustNewForkOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ForkOperator {
	node, err := NewForkOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewForkOperatorTyped(forkKeyword *SyntaxToken, expressions *SyntaxListOf[*ForkExpression], diagnostics ...Diagnostic) (*ForkOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(forkKeyword),
		syntaxElementFromSyntaxListOf(expressions),
	}
	created, err := NewForkOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewForkOperatorTyped(forkKeyword *SyntaxToken, expressions *SyntaxListOf[*ForkExpression], diagnostics ...Diagnostic) *ForkOperator {
	node, err := NewForkOperatorTyped(forkKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ForkOperator) CloneTyped(includeDiagnostics bool) *ForkOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ForkOperator{SyntaxNode: cloned}
}

func (n *ForkOperator) ForkKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ForkOperator) Expressions() *SyntaxListOf[*ForkExpression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*ForkExpression](list)
}

type NameEqualsClause struct {
	*SyntaxNode
}

func (*NameEqualsClause) isExpression() {}

func NewNameEqualsClause(children []SyntaxElement, diagnostics ...Diagnostic) (*NameEqualsClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("NameEqualsClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NameEqualsClause{SyntaxNode: node}, nil
}

func MustNewNameEqualsClause(children []SyntaxElement, diagnostics ...Diagnostic) *NameEqualsClause {
	node, err := NewNameEqualsClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNameEqualsClauseTyped(name *NameDeclaration, equalToken *SyntaxToken, diagnostics ...Diagnostic) (*NameEqualsClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equalToken),
	}
	created, err := NewNameEqualsClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewNameEqualsClauseTyped(name *NameDeclaration, equalToken *SyntaxToken, diagnostics ...Diagnostic) *NameEqualsClause {
	node, err := NewNameEqualsClauseTyped(name, equalToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NameEqualsClause) CloneTyped(includeDiagnostics bool) *NameEqualsClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NameEqualsClause{SyntaxNode: cloned}
}

func (n *NameEqualsClause) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *NameEqualsClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ForkExpression struct {
	*SyntaxNode
}

func (*ForkExpression) isExpression() {}

func NewForkExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ForkExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ForkExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ForkExpression{SyntaxNode: node}, nil
}

func MustNewForkExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ForkExpression {
	node, err := NewForkExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewForkExpressionTyped(nameEquals *NameEqualsClause, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ForkExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(nameEquals),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewForkExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewForkExpressionTyped(nameEquals *NameEqualsClause, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *ForkExpression {
	node, err := NewForkExpressionTyped(nameEquals, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ForkExpression) CloneTyped(includeDiagnostics bool) *ForkExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ForkExpression{SyntaxNode: cloned}
}

func (n *ForkExpression) NameEquals() *NameEqualsClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameEqualsClause)
	return child
}

func (n *ForkExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ForkExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ForkExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type MakeSeriesOperator struct {
	*SyntaxNode
}

func (*MakeSeriesOperator) isQueryOperator() {}

func (*MakeSeriesOperator) isExpression() {}

func NewMakeSeriesOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesOperator{SyntaxNode: node}, nil
}

func MustNewMakeSeriesOperator(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesOperator {
	node, err := NewMakeSeriesOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesOperatorTyped(makeSeriesKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[*MakeSeriesExpression], onClause *MakeSeriesOnClause, rangeClause MakeSeriesRangeClause, byClause *MakeSeriesByClause, diagnostics ...Diagnostic) (*MakeSeriesOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(makeSeriesKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(aggregates),
		syntaxElementFromTyped(onClause),
		syntaxElementFromTyped(rangeClause),
		syntaxElementFromTyped(byClause),
	}
	created, err := NewMakeSeriesOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesOperatorTyped(makeSeriesKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[*MakeSeriesExpression], onClause *MakeSeriesOnClause, rangeClause MakeSeriesRangeClause, byClause *MakeSeriesByClause, diagnostics ...Diagnostic) *MakeSeriesOperator {
	node, err := NewMakeSeriesOperatorTyped(makeSeriesKeyword, parameters, aggregates, onClause, rangeClause, byClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesOperator) CloneTyped(includeDiagnostics bool) *MakeSeriesOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesOperator{SyntaxNode: cloned}
}

func (n *MakeSeriesOperator) MakeSeriesKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *MakeSeriesOperator) Aggregates() *SeparatedSyntaxListOf[*MakeSeriesExpression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*MakeSeriesExpression](list)
}

func (n *MakeSeriesOperator) OnClause() *MakeSeriesOnClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeSeriesOnClause)
	return child
}

func (n *MakeSeriesOperator) RangeClause() MakeSeriesRangeClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(MakeSeriesRangeClause)
	return child
}

func (n *MakeSeriesOperator) ByClause() *MakeSeriesByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeSeriesByClause)
	return child
}

type MakeSeriesExpression struct {
	*SyntaxNode
}

func (*MakeSeriesExpression) isExpression() {}

func NewMakeSeriesExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesExpression{SyntaxNode: node}, nil
}

func MustNewMakeSeriesExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesExpression {
	node, err := NewMakeSeriesExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesExpressionTyped(expression Expression, defaultExpression *DefaultExpressionClause, diagnostics ...Diagnostic) (*MakeSeriesExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(defaultExpression),
	}
	created, err := NewMakeSeriesExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesExpressionTyped(expression Expression, defaultExpression *DefaultExpressionClause, diagnostics ...Diagnostic) *MakeSeriesExpression {
	node, err := NewMakeSeriesExpressionTyped(expression, defaultExpression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesExpression) CloneTyped(includeDiagnostics bool) *MakeSeriesExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesExpression{SyntaxNode: cloned}
}

func (n *MakeSeriesExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MakeSeriesExpression) DefaultExpression() *DefaultExpressionClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*DefaultExpressionClause)
	return child
}

type DefaultExpressionClause struct {
	*SyntaxNode
}

func (*DefaultExpressionClause) isExpression() {}

func NewDefaultExpressionClause(children []SyntaxElement, diagnostics ...Diagnostic) (*DefaultExpressionClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("DefaultExpressionClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DefaultExpressionClause{SyntaxNode: node}, nil
}

func MustNewDefaultExpressionClause(children []SyntaxElement, diagnostics ...Diagnostic) *DefaultExpressionClause {
	node, err := NewDefaultExpressionClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDefaultExpressionClauseTyped(defaultKeyword *SyntaxToken, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*DefaultExpressionClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(defaultKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(expression),
	}
	created, err := NewDefaultExpressionClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDefaultExpressionClauseTyped(defaultKeyword *SyntaxToken, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *DefaultExpressionClause {
	node, err := NewDefaultExpressionClauseTyped(defaultKeyword, equalToken, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DefaultExpressionClause) CloneTyped(includeDiagnostics bool) *DefaultExpressionClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DefaultExpressionClause{SyntaxNode: cloned}
}

func (n *DefaultExpressionClause) DefaultKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DefaultExpressionClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DefaultExpressionClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeSeriesOnClause struct {
	*SyntaxNode
}

func (*MakeSeriesOnClause) isClause() {}

func NewMakeSeriesOnClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesOnClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesOnClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesOnClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesOnClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesOnClause {
	node, err := NewMakeSeriesOnClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesOnClauseTyped(onKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*MakeSeriesOnClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(onKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewMakeSeriesOnClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesOnClauseTyped(onKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *MakeSeriesOnClause {
	node, err := NewMakeSeriesOnClauseTyped(onKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesOnClause) CloneTyped(includeDiagnostics bool) *MakeSeriesOnClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesOnClause{SyntaxNode: cloned}
}

func (n *MakeSeriesOnClause) OnKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesOnClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeSeriesFromClause struct {
	*SyntaxNode
}

func (*MakeSeriesFromClause) isClause() {}

func NewMakeSeriesFromClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesFromClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesFromClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesFromClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesFromClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesFromClause {
	node, err := NewMakeSeriesFromClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesFromClauseTyped(fromKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*MakeSeriesFromClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(fromKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewMakeSeriesFromClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesFromClauseTyped(fromKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *MakeSeriesFromClause {
	node, err := NewMakeSeriesFromClauseTyped(fromKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesFromClause) CloneTyped(includeDiagnostics bool) *MakeSeriesFromClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesFromClause{SyntaxNode: cloned}
}

func (n *MakeSeriesFromClause) FromKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesFromClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeSeriesToClause struct {
	*SyntaxNode
}

func (*MakeSeriesToClause) isClause() {}

func NewMakeSeriesToClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesToClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesToClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesToClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesToClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesToClause {
	node, err := NewMakeSeriesToClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesToClauseTyped(toKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*MakeSeriesToClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(toKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewMakeSeriesToClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesToClauseTyped(toKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *MakeSeriesToClause {
	node, err := NewMakeSeriesToClauseTyped(toKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesToClause) CloneTyped(includeDiagnostics bool) *MakeSeriesToClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesToClause{SyntaxNode: cloned}
}

func (n *MakeSeriesToClause) ToKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesToClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeSeriesStepClause struct {
	*SyntaxNode
}

func (*MakeSeriesStepClause) isClause() {}

func NewMakeSeriesStepClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesStepClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesStepClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesStepClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesStepClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesStepClause {
	node, err := NewMakeSeriesStepClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesStepClauseTyped(stepKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*MakeSeriesStepClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(stepKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewMakeSeriesStepClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesStepClauseTyped(stepKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *MakeSeriesStepClause {
	node, err := NewMakeSeriesStepClauseTyped(stepKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesStepClause) CloneTyped(includeDiagnostics bool) *MakeSeriesStepClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesStepClause{SyntaxNode: cloned}
}

func (n *MakeSeriesStepClause) StepKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesStepClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeSeriesInRangeClause struct {
	*SyntaxNode
}

func (*MakeSeriesInRangeClause) isMakeSeriesRangeClause() {}

func (*MakeSeriesInRangeClause) isClause() {}

func NewMakeSeriesInRangeClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesInRangeClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesInRangeClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesInRangeClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesInRangeClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesInRangeClause {
	node, err := NewMakeSeriesInRangeClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesInRangeClauseTyped(inKeyword *SyntaxToken, rangeKeyword *SyntaxToken, arguments *ExpressionList, diagnostics ...Diagnostic) (*MakeSeriesInRangeClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(inKeyword),
		syntaxElementFromTyped(rangeKeyword),
		syntaxElementFromTyped(arguments),
	}
	created, err := NewMakeSeriesInRangeClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesInRangeClauseTyped(inKeyword *SyntaxToken, rangeKeyword *SyntaxToken, arguments *ExpressionList, diagnostics ...Diagnostic) *MakeSeriesInRangeClause {
	node, err := NewMakeSeriesInRangeClauseTyped(inKeyword, rangeKeyword, arguments, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesInRangeClause) CloneTyped(includeDiagnostics bool) *MakeSeriesInRangeClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesInRangeClause{SyntaxNode: cloned}
}

func (n *MakeSeriesInRangeClause) InKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesInRangeClause) RangeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesInRangeClause) Arguments() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

type MakeSeriesFromToStepClause struct {
	*SyntaxNode
}

func (*MakeSeriesFromToStepClause) isMakeSeriesRangeClause() {}

func (*MakeSeriesFromToStepClause) isClause() {}

func NewMakeSeriesFromToStepClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesFromToStepClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesFromToStepClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesFromToStepClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesFromToStepClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesFromToStepClause {
	node, err := NewMakeSeriesFromToStepClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesFromToStepClauseTyped(makeSeriesFromClause *MakeSeriesFromClause, makeSeriesToClause *MakeSeriesToClause, makeSeriesStepClause *MakeSeriesStepClause, diagnostics ...Diagnostic) (*MakeSeriesFromToStepClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(makeSeriesFromClause),
		syntaxElementFromTyped(makeSeriesToClause),
		syntaxElementFromTyped(makeSeriesStepClause),
	}
	created, err := NewMakeSeriesFromToStepClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesFromToStepClauseTyped(makeSeriesFromClause *MakeSeriesFromClause, makeSeriesToClause *MakeSeriesToClause, makeSeriesStepClause *MakeSeriesStepClause, diagnostics ...Diagnostic) *MakeSeriesFromToStepClause {
	node, err := NewMakeSeriesFromToStepClauseTyped(makeSeriesFromClause, makeSeriesToClause, makeSeriesStepClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesFromToStepClause) CloneTyped(includeDiagnostics bool) *MakeSeriesFromToStepClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesFromToStepClause{SyntaxNode: cloned}
}

func (n *MakeSeriesFromToStepClause) MakeSeriesFromClause() *MakeSeriesFromClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeSeriesFromClause)
	return child
}

func (n *MakeSeriesFromToStepClause) MakeSeriesToClause() *MakeSeriesToClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeSeriesToClause)
	return child
}

func (n *MakeSeriesFromToStepClause) MakeSeriesStepClause() *MakeSeriesStepClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeSeriesStepClause)
	return child
}

type MakeSeriesByClause struct {
	*SyntaxNode
}

func (*MakeSeriesByClause) isClause() {}

func NewMakeSeriesByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeSeriesByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeSeriesByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeSeriesByClause{SyntaxNode: node}, nil
}

func MustNewMakeSeriesByClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeSeriesByClause {
	node, err := NewMakeSeriesByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeSeriesByClauseTyped(byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*MakeSeriesByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewMakeSeriesByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeSeriesByClauseTyped(byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *MakeSeriesByClause {
	node, err := NewMakeSeriesByClauseTyped(byKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeSeriesByClause) CloneTyped(includeDiagnostics bool) *MakeSeriesByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeSeriesByClause{SyntaxNode: cloned}
}

func (n *MakeSeriesByClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeSeriesByClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type MvExpandOperator struct {
	*SyntaxNode
}

func (*MvExpandOperator) isQueryOperator() {}

func (*MvExpandOperator) isExpression() {}

func NewMvExpandOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*MvExpandOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvExpandOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvExpandOperator{SyntaxNode: node}, nil
}

func MustNewMvExpandOperator(children []SyntaxElement, diagnostics ...Diagnostic) *MvExpandOperator {
	node, err := NewMvExpandOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvExpandOperatorTyped(mvExpandKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvExpandExpression], rowLimitClause *MvExpandRowLimitClause, diagnostics ...Diagnostic) (*MvExpandOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(mvExpandKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(rowLimitClause),
	}
	created, err := NewMvExpandOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvExpandOperatorTyped(mvExpandKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvExpandExpression], rowLimitClause *MvExpandRowLimitClause, diagnostics ...Diagnostic) *MvExpandOperator {
	node, err := NewMvExpandOperatorTyped(mvExpandKeyword, parameters, expressions, rowLimitClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvExpandOperator) CloneTyped(includeDiagnostics bool) *MvExpandOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvExpandOperator{SyntaxNode: cloned}
}

func (n *MvExpandOperator) MvExpandKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvExpandOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *MvExpandOperator) Expressions() *SeparatedSyntaxListOf[*MvExpandExpression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*MvExpandExpression](list)
}

func (n *MvExpandOperator) RowLimitClause() *MvExpandRowLimitClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MvExpandRowLimitClause)
	return child
}

type MvExpandExpression struct {
	*SyntaxNode
}

func (*MvExpandExpression) isExpression() {}

func NewMvExpandExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MvExpandExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvExpandExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvExpandExpression{SyntaxNode: node}, nil
}

func MustNewMvExpandExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MvExpandExpression {
	node, err := NewMvExpandExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvExpandExpressionTyped(expression Expression, toTypeOf *ToTypeOfClause, diagnostics ...Diagnostic) (*MvExpandExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(toTypeOf),
	}
	created, err := NewMvExpandExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvExpandExpressionTyped(expression Expression, toTypeOf *ToTypeOfClause, diagnostics ...Diagnostic) *MvExpandExpression {
	node, err := NewMvExpandExpressionTyped(expression, toTypeOf, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvExpandExpression) CloneTyped(includeDiagnostics bool) *MvExpandExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvExpandExpression{SyntaxNode: cloned}
}

func (n *MvExpandExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MvExpandExpression) ToTypeOf() *ToTypeOfClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ToTypeOfClause)
	return child
}

type MvExpandRowLimitClause struct {
	*SyntaxNode
}

func (*MvExpandRowLimitClause) isExpression() {}

func NewMvExpandRowLimitClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MvExpandRowLimitClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvExpandRowLimitClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvExpandRowLimitClause{SyntaxNode: node}, nil
}

func MustNewMvExpandRowLimitClause(children []SyntaxElement, diagnostics ...Diagnostic) *MvExpandRowLimitClause {
	node, err := NewMvExpandRowLimitClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvExpandRowLimitClauseTyped(limitKeyword *SyntaxToken, rowLimit Expression, diagnostics ...Diagnostic) (*MvExpandRowLimitClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(limitKeyword),
		syntaxElementFromTyped(rowLimit),
	}
	created, err := NewMvExpandRowLimitClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvExpandRowLimitClauseTyped(limitKeyword *SyntaxToken, rowLimit Expression, diagnostics ...Diagnostic) *MvExpandRowLimitClause {
	node, err := NewMvExpandRowLimitClauseTyped(limitKeyword, rowLimit, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvExpandRowLimitClause) CloneTyped(includeDiagnostics bool) *MvExpandRowLimitClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvExpandRowLimitClause{SyntaxNode: cloned}
}

func (n *MvExpandRowLimitClause) LimitKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvExpandRowLimitClause) RowLimit() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MvApplyOperator struct {
	*SyntaxNode
}

func (*MvApplyOperator) isQueryOperator() {}

func (*MvApplyOperator) isExpression() {}

func NewMvApplyOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*MvApplyOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvApplyOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvApplyOperator{SyntaxNode: node}, nil
}

func MustNewMvApplyOperator(children []SyntaxElement, diagnostics ...Diagnostic) *MvApplyOperator {
	node, err := NewMvApplyOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvApplyOperatorTyped(mvApplyKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvApplyExpression], rowLimitClause *MvApplyRowLimitClause, contextIdClause *MvApplyContextIdClause, onKeyword *SyntaxToken, subquery *MvApplySubqueryExpression, diagnostics ...Diagnostic) (*MvApplyOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(mvApplyKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(rowLimitClause),
		syntaxElementFromTyped(contextIdClause),
		syntaxElementFromTyped(onKeyword),
		syntaxElementFromTyped(subquery),
	}
	created, err := NewMvApplyOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvApplyOperatorTyped(mvApplyKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[*MvApplyExpression], rowLimitClause *MvApplyRowLimitClause, contextIdClause *MvApplyContextIdClause, onKeyword *SyntaxToken, subquery *MvApplySubqueryExpression, diagnostics ...Diagnostic) *MvApplyOperator {
	node, err := NewMvApplyOperatorTyped(mvApplyKeyword, parameters, expressions, rowLimitClause, contextIdClause, onKeyword, subquery, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvApplyOperator) CloneTyped(includeDiagnostics bool) *MvApplyOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvApplyOperator{SyntaxNode: cloned}
}

func (n *MvApplyOperator) MvApplyKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvApplyOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *MvApplyOperator) Expressions() *SeparatedSyntaxListOf[*MvApplyExpression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*MvApplyExpression](list)
}

func (n *MvApplyOperator) RowLimitClause() *MvApplyRowLimitClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MvApplyRowLimitClause)
	return child
}

func (n *MvApplyOperator) ContextIdClause() *MvApplyContextIdClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MvApplyContextIdClause)
	return child
}

func (n *MvApplyOperator) OnKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvApplyOperator) Subquery() *MvApplySubqueryExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MvApplySubqueryExpression)
	return child
}

type MvApplyExpression struct {
	*SyntaxNode
}

func (*MvApplyExpression) isExpression() {}

func NewMvApplyExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MvApplyExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvApplyExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvApplyExpression{SyntaxNode: node}, nil
}

func MustNewMvApplyExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MvApplyExpression {
	node, err := NewMvApplyExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvApplyExpressionTyped(expression Expression, toTypeOf *ToTypeOfClause, diagnostics ...Diagnostic) (*MvApplyExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(toTypeOf),
	}
	created, err := NewMvApplyExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvApplyExpressionTyped(expression Expression, toTypeOf *ToTypeOfClause, diagnostics ...Diagnostic) *MvApplyExpression {
	node, err := NewMvApplyExpressionTyped(expression, toTypeOf, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvApplyExpression) CloneTyped(includeDiagnostics bool) *MvApplyExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvApplyExpression{SyntaxNode: cloned}
}

func (n *MvApplyExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MvApplyExpression) ToTypeOf() *ToTypeOfClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ToTypeOfClause)
	return child
}

type MvApplyRowLimitClause struct {
	*SyntaxNode
}

func (*MvApplyRowLimitClause) isExpression() {}

func NewMvApplyRowLimitClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MvApplyRowLimitClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvApplyRowLimitClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvApplyRowLimitClause{SyntaxNode: node}, nil
}

func MustNewMvApplyRowLimitClause(children []SyntaxElement, diagnostics ...Diagnostic) *MvApplyRowLimitClause {
	node, err := NewMvApplyRowLimitClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvApplyRowLimitClauseTyped(limitKeyword *SyntaxToken, rowLimit Expression, diagnostics ...Diagnostic) (*MvApplyRowLimitClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(limitKeyword),
		syntaxElementFromTyped(rowLimit),
	}
	created, err := NewMvApplyRowLimitClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvApplyRowLimitClauseTyped(limitKeyword *SyntaxToken, rowLimit Expression, diagnostics ...Diagnostic) *MvApplyRowLimitClause {
	node, err := NewMvApplyRowLimitClauseTyped(limitKeyword, rowLimit, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvApplyRowLimitClause) CloneTyped(includeDiagnostics bool) *MvApplyRowLimitClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvApplyRowLimitClause{SyntaxNode: cloned}
}

func (n *MvApplyRowLimitClause) LimitKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvApplyRowLimitClause) RowLimit() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MvApplyContextIdClause struct {
	*SyntaxNode
}

func (*MvApplyContextIdClause) isExpression() {}

func NewMvApplyContextIdClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MvApplyContextIdClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvApplyContextIdClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvApplyContextIdClause{SyntaxNode: node}, nil
}

func MustNewMvApplyContextIdClause(children []SyntaxElement, diagnostics ...Diagnostic) *MvApplyContextIdClause {
	node, err := NewMvApplyContextIdClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvApplyContextIdClauseTyped(idKeyword *SyntaxToken, id Expression, diagnostics ...Diagnostic) (*MvApplyContextIdClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(idKeyword),
		syntaxElementFromTyped(id),
	}
	created, err := NewMvApplyContextIdClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvApplyContextIdClauseTyped(idKeyword *SyntaxToken, id Expression, diagnostics ...Diagnostic) *MvApplyContextIdClause {
	node, err := NewMvApplyContextIdClauseTyped(idKeyword, id, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvApplyContextIdClause) CloneTyped(includeDiagnostics bool) *MvApplyContextIdClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvApplyContextIdClause{SyntaxNode: cloned}
}

func (n *MvApplyContextIdClause) IdKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvApplyContextIdClause) Id() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MvApplySubqueryExpression struct {
	*SyntaxNode
}

func (*MvApplySubqueryExpression) isExpression() {}

func NewMvApplySubqueryExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MvApplySubqueryExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MvApplySubqueryExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MvApplySubqueryExpression{SyntaxNode: node}, nil
}

func MustNewMvApplySubqueryExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MvApplySubqueryExpression {
	node, err := NewMvApplySubqueryExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMvApplySubqueryExpressionTyped(openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MvApplySubqueryExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMvApplySubqueryExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMvApplySubqueryExpressionTyped(openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *MvApplySubqueryExpression {
	node, err := NewMvApplySubqueryExpressionTyped(openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MvApplySubqueryExpression) CloneTyped(includeDiagnostics bool) *MvApplySubqueryExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MvApplySubqueryExpression{SyntaxNode: cloned}
}

func (n *MvApplySubqueryExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MvApplySubqueryExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MvApplySubqueryExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ToTypeOfClause struct {
	*SyntaxNode
}

func NewToTypeOfClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ToTypeOfClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ToTypeOfClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ToTypeOfClause{SyntaxNode: node}, nil
}

func MustNewToTypeOfClause(children []SyntaxElement, diagnostics ...Diagnostic) *ToTypeOfClause {
	node, err := NewToTypeOfClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewToTypeOfClauseTyped(toKeyword *SyntaxToken, typeOf *TypeOfLiteralExpression, diagnostics ...Diagnostic) (*ToTypeOfClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(toKeyword),
		syntaxElementFromTyped(typeOf),
	}
	created, err := NewToTypeOfClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewToTypeOfClauseTyped(toKeyword *SyntaxToken, typeOf *TypeOfLiteralExpression, diagnostics ...Diagnostic) *ToTypeOfClause {
	node, err := NewToTypeOfClauseTyped(toKeyword, typeOf, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ToTypeOfClause) CloneTyped(includeDiagnostics bool) *ToTypeOfClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ToTypeOfClause{SyntaxNode: cloned}
}

func (n *ToTypeOfClause) ToKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ToTypeOfClause) TypeOf() *TypeOfLiteralExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*TypeOfLiteralExpression)
	return child
}

type EvaluateSchemaClause struct {
	*SyntaxNode
}

func NewEvaluateSchemaClause(children []SyntaxElement, diagnostics ...Diagnostic) (*EvaluateSchemaClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("EvaluateSchemaClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &EvaluateSchemaClause{SyntaxNode: node}, nil
}

func MustNewEvaluateSchemaClause(children []SyntaxElement, diagnostics ...Diagnostic) *EvaluateSchemaClause {
	node, err := NewEvaluateSchemaClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewEvaluateSchemaClauseTyped(colonToken *SyntaxToken, schema *EvaluateRowSchema, diagnostics ...Diagnostic) (*EvaluateSchemaClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(colonToken),
		syntaxElementFromTyped(schema),
	}
	created, err := NewEvaluateSchemaClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewEvaluateSchemaClauseTyped(colonToken *SyntaxToken, schema *EvaluateRowSchema, diagnostics ...Diagnostic) *EvaluateSchemaClause {
	node, err := NewEvaluateSchemaClauseTyped(colonToken, schema, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *EvaluateSchemaClause) CloneTyped(includeDiagnostics bool) *EvaluateSchemaClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &EvaluateSchemaClause{SyntaxNode: cloned}
}

func (n *EvaluateSchemaClause) ColonToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateSchemaClause) Schema() *EvaluateRowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*EvaluateRowSchema)
	return child
}

type EvaluateOperator struct {
	*SyntaxNode
}

func (*EvaluateOperator) isQueryOperator() {}

func (*EvaluateOperator) isExpression() {}

func NewEvaluateOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*EvaluateOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("EvaluateOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &EvaluateOperator{SyntaxNode: node}, nil
}

func MustNewEvaluateOperator(children []SyntaxElement, diagnostics ...Diagnostic) *EvaluateOperator {
	node, err := NewEvaluateOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewEvaluateOperatorTyped(evaluateKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], functionCall *FunctionCallExpression, schema *EvaluateSchemaClause, diagnostics ...Diagnostic) (*EvaluateOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(evaluateKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(functionCall),
		syntaxElementFromTyped(schema),
	}
	created, err := NewEvaluateOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewEvaluateOperatorTyped(evaluateKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], functionCall *FunctionCallExpression, schema *EvaluateSchemaClause, diagnostics ...Diagnostic) *EvaluateOperator {
	node, err := NewEvaluateOperatorTyped(evaluateKeyword, parameters, functionCall, schema, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *EvaluateOperator) CloneTyped(includeDiagnostics bool) *EvaluateOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &EvaluateOperator{SyntaxNode: cloned}
}

func (n *EvaluateOperator) EvaluateKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *EvaluateOperator) FunctionCall() *FunctionCallExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
	return child
}

func (n *EvaluateOperator) Schema() *EvaluateSchemaClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*EvaluateSchemaClause)
	return child
}

type ParseOperator struct {
	*SyntaxNode
}

func (*ParseOperator) isQueryOperator() {}

func (*ParseOperator) isExpression() {}

func NewParseOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ParseOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ParseOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ParseOperator{SyntaxNode: node}, nil
}

func MustNewParseOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ParseOperator {
	node, err := NewParseOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewParseOperatorTyped(parseKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, withKeyword *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement], diagnostics ...Diagnostic) (*ParseOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(parseKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromSyntaxListOf(patterns),
	}
	created, err := NewParseOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewParseOperatorTyped(parseKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, withKeyword *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement], diagnostics ...Diagnostic) *ParseOperator {
	node, err := NewParseOperatorTyped(parseKeyword, parameters, expression, withKeyword, patterns, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ParseOperator) CloneTyped(includeDiagnostics bool) *ParseOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ParseOperator{SyntaxNode: cloned}
}

func (n *ParseOperator) ParseKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *ParseOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ParseOperator) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseOperator) Patterns() *SyntaxListOf[SyntaxNodeElement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[SyntaxNodeElement](list)
}

type ParseWhereOperator struct {
	*SyntaxNode
}

func (*ParseWhereOperator) isQueryOperator() {}

func (*ParseWhereOperator) isExpression() {}

func NewParseWhereOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ParseWhereOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ParseWhereOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ParseWhereOperator{SyntaxNode: node}, nil
}

func MustNewParseWhereOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ParseWhereOperator {
	node, err := NewParseWhereOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewParseWhereOperatorTyped(parseKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, withKeyword *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement], diagnostics ...Diagnostic) (*ParseWhereOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(parseKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromSyntaxListOf(patterns),
	}
	created, err := NewParseWhereOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewParseWhereOperatorTyped(parseKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, withKeyword *SyntaxToken, patterns *SyntaxListOf[SyntaxNodeElement], diagnostics ...Diagnostic) *ParseWhereOperator {
	node, err := NewParseWhereOperatorTyped(parseKeyword, parameters, expression, withKeyword, patterns, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ParseWhereOperator) CloneTyped(includeDiagnostics bool) *ParseWhereOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ParseWhereOperator{SyntaxNode: cloned}
}

func (n *ParseWhereOperator) ParseKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseWhereOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *ParseWhereOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ParseWhereOperator) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseWhereOperator) Patterns() *SyntaxListOf[SyntaxNodeElement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[SyntaxNodeElement](list)
}

type ParseKvWithClause struct {
	*SyntaxNode
}

func (*ParseKvWithClause) isClause() {}

func NewParseKvWithClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ParseKvWithClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ParseKvWithClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ParseKvWithClause{SyntaxNode: node}, nil
}

func MustNewParseKvWithClause(children []SyntaxElement, diagnostics ...Diagnostic) *ParseKvWithClause {
	node, err := NewParseKvWithClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewParseKvWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ParseKvWithClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(properties),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewParseKvWithClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewParseKvWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *ParseKvWithClause {
	node, err := NewParseKvWithClauseTyped(withKeyword, openParen, properties, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ParseKvWithClause) CloneTyped(includeDiagnostics bool) *ParseKvWithClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ParseKvWithClause{SyntaxNode: cloned}
}

func (n *ParseKvWithClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseKvWithClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseKvWithClause) Properties() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

func (n *ParseKvWithClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ParseKvOperator struct {
	*SyntaxNode
}

func (*ParseKvOperator) isQueryOperator() {}

func (*ParseKvOperator) isExpression() {}

func NewParseKvOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ParseKvOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ParseKvOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ParseKvOperator{SyntaxNode: node}, nil
}

func MustNewParseKvOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ParseKvOperator {
	node, err := NewParseKvOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewParseKvOperatorTyped(parseKvKeyword *SyntaxToken, expression Expression, asKeyword *SyntaxToken, keys *RowSchema, withClause *ParseKvWithClause, diagnostics ...Diagnostic) (*ParseKvOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(parseKvKeyword),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(asKeyword),
		syntaxElementFromTyped(keys),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewParseKvOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewParseKvOperatorTyped(parseKvKeyword *SyntaxToken, expression Expression, asKeyword *SyntaxToken, keys *RowSchema, withClause *ParseKvWithClause, diagnostics ...Diagnostic) *ParseKvOperator {
	node, err := NewParseKvOperatorTyped(parseKvKeyword, expression, asKeyword, keys, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ParseKvOperator) CloneTyped(includeDiagnostics bool) *ParseKvOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ParseKvOperator{SyntaxNode: cloned}
}

func (n *ParseKvOperator) ParseKvKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseKvOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ParseKvOperator) AsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ParseKvOperator) Keys() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

func (n *ParseKvOperator) WithClause() *ParseKvWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ParseKvWithClause)
	return child
}

type PartitionOperator struct {
	*SyntaxNode
}

func (*PartitionOperator) isQueryOperator() {}

func (*PartitionOperator) isExpression() {}

func NewPartitionOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionOperator{SyntaxNode: node}, nil
}

func MustNewPartitionOperator(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionOperator {
	node, err := NewPartitionOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionOperatorTyped(partitionKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, byExpression Expression, scope *PartitionScope, operand PartitionOperand, diagnostics ...Diagnostic) (*PartitionOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(byExpression),
		syntaxElementFromTyped(scope),
		syntaxElementFromTyped(operand),
	}
	created, err := NewPartitionOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionOperatorTyped(partitionKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, byExpression Expression, scope *PartitionScope, operand PartitionOperand, diagnostics ...Diagnostic) *PartitionOperator {
	node, err := NewPartitionOperatorTyped(partitionKeyword, parameters, byKeyword, byExpression, scope, operand, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionOperator) CloneTyped(includeDiagnostics bool) *PartitionOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionOperator{SyntaxNode: cloned}
}

func (n *PartitionOperator) PartitionKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *PartitionOperator) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionOperator) ByExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PartitionOperator) Scope() *PartitionScope {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*PartitionScope)
	return child
}

func (n *PartitionOperator) Operand() PartitionOperand {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(PartitionOperand)
	return child
}

type PartitionQuery struct {
	*SyntaxNode
}

func (*PartitionQuery) isPartitionOperand() {}

func (*PartitionQuery) isExpression() {}

func NewPartitionQuery(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionQuery, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionQuery", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionQuery{SyntaxNode: node}, nil
}

func MustNewPartitionQuery(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionQuery {
	node, err := NewPartitionQuery(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionQueryTyped(openBrace *SyntaxToken, query Expression, closeBrace *SyntaxToken, diagnostics ...Diagnostic) (*PartitionQuery, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBrace),
		syntaxElementFromTyped(query),
		syntaxElementFromTyped(closeBrace),
	}
	created, err := NewPartitionQuery(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionQueryTyped(openBrace *SyntaxToken, query Expression, closeBrace *SyntaxToken, diagnostics ...Diagnostic) *PartitionQuery {
	node, err := NewPartitionQueryTyped(openBrace, query, closeBrace, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionQuery) CloneTyped(includeDiagnostics bool) *PartitionQuery {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionQuery{SyntaxNode: cloned}
}

func (n *PartitionQuery) OpenBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionQuery) Query() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PartitionQuery) CloseBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PartitionScope struct {
	*SyntaxNode
}

func NewPartitionScope(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionScope, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionScope", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionScope{SyntaxNode: node}, nil
}

func MustNewPartitionScope(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionScope {
	node, err := NewPartitionScope(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionScopeTyped(inKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*PartitionScope, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(inKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewPartitionScope(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionScopeTyped(inKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *PartitionScope {
	node, err := NewPartitionScopeTyped(inKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionScope) CloneTyped(includeDiagnostics bool) *PartitionScope {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionScope{SyntaxNode: cloned}
}

func (n *PartitionScope) InKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionScope) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type PartitionSubquery struct {
	*SyntaxNode
}

func (*PartitionSubquery) isPartitionOperand() {}

func (*PartitionSubquery) isExpression() {}

func NewPartitionSubquery(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionSubquery, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionSubquery", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionSubquery{SyntaxNode: node}, nil
}

func MustNewPartitionSubquery(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionSubquery {
	node, err := NewPartitionSubquery(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionSubqueryTyped(openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*PartitionSubquery, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(subquery),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewPartitionSubquery(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionSubqueryTyped(openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *PartitionSubquery {
	node, err := NewPartitionSubqueryTyped(openParen, subquery, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionSubquery) CloneTyped(includeDiagnostics bool) *PartitionSubquery {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionSubquery{SyntaxNode: cloned}
}

func (n *PartitionSubquery) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionSubquery) Subquery() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PartitionSubquery) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ProjectOperator struct {
	*SyntaxNode
}

func (*ProjectOperator) isQueryOperator() {}

func (*ProjectOperator) isExpression() {}

func NewProjectOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectOperator{SyntaxNode: node}, nil
}

func MustNewProjectOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectOperator {
	node, err := NewProjectOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectOperatorTyped(projectKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectOperatorTyped(projectKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectOperator {
	node, err := NewProjectOperatorTyped(projectKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectOperator) CloneTyped(includeDiagnostics bool) *ProjectOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectOperator{SyntaxNode: cloned}
}

func (n *ProjectOperator) ProjectKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ProjectAwayOperator struct {
	*SyntaxNode
}

func (*ProjectAwayOperator) isQueryOperator() {}

func (*ProjectAwayOperator) isExpression() {}

func NewProjectAwayOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectAwayOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectAwayOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectAwayOperator{SyntaxNode: node}, nil
}

func MustNewProjectAwayOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectAwayOperator {
	node, err := NewProjectAwayOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectAwayOperatorTyped(projectAwayKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectAwayOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectAwayKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectAwayOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectAwayOperatorTyped(projectAwayKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectAwayOperator {
	node, err := NewProjectAwayOperatorTyped(projectAwayKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectAwayOperator) CloneTyped(includeDiagnostics bool) *ProjectAwayOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectAwayOperator{SyntaxNode: cloned}
}

func (n *ProjectAwayOperator) ProjectAwayKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectAwayOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ProjectByNamesOperator struct {
	*SyntaxNode
}

func (*ProjectByNamesOperator) isQueryOperator() {}

func (*ProjectByNamesOperator) isExpression() {}

func NewProjectByNamesOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectByNamesOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectByNamesOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectByNamesOperator{SyntaxNode: node}, nil
}

func MustNewProjectByNamesOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectByNamesOperator {
	node, err := NewProjectByNamesOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectByNamesOperatorTyped(projectByNamesKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectByNamesOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectByNamesKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectByNamesOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectByNamesOperatorTyped(projectByNamesKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectByNamesOperator {
	node, err := NewProjectByNamesOperatorTyped(projectByNamesKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectByNamesOperator) CloneTyped(includeDiagnostics bool) *ProjectByNamesOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectByNamesOperator{SyntaxNode: cloned}
}

func (n *ProjectByNamesOperator) ProjectByNamesKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectByNamesOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ProjectKeepOperator struct {
	*SyntaxNode
}

func (*ProjectKeepOperator) isQueryOperator() {}

func (*ProjectKeepOperator) isExpression() {}

func NewProjectKeepOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectKeepOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectKeepOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectKeepOperator{SyntaxNode: node}, nil
}

func MustNewProjectKeepOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectKeepOperator {
	node, err := NewProjectKeepOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectKeepOperatorTyped(projectKeepKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectKeepOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectKeepKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectKeepOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectKeepOperatorTyped(projectKeepKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectKeepOperator {
	node, err := NewProjectKeepOperatorTyped(projectKeepKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectKeepOperator) CloneTyped(includeDiagnostics bool) *ProjectKeepOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectKeepOperator{SyntaxNode: cloned}
}

func (n *ProjectKeepOperator) ProjectKeepKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectKeepOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ProjectRenameOperator struct {
	*SyntaxNode
}

func (*ProjectRenameOperator) isQueryOperator() {}

func (*ProjectRenameOperator) isExpression() {}

func NewProjectRenameOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectRenameOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectRenameOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectRenameOperator{SyntaxNode: node}, nil
}

func MustNewProjectRenameOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectRenameOperator {
	node, err := NewProjectRenameOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectRenameOperatorTyped(projectRenameKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectRenameOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectRenameKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectRenameOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectRenameOperatorTyped(projectRenameKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectRenameOperator {
	node, err := NewProjectRenameOperatorTyped(projectRenameKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectRenameOperator) CloneTyped(includeDiagnostics bool) *ProjectRenameOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectRenameOperator{SyntaxNode: cloned}
}

func (n *ProjectRenameOperator) ProjectRenameKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectRenameOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ProjectReorderOperator struct {
	*SyntaxNode
}

func (*ProjectReorderOperator) isQueryOperator() {}

func (*ProjectReorderOperator) isExpression() {}

func NewProjectReorderOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectReorderOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectReorderOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectReorderOperator{SyntaxNode: node}, nil
}

func MustNewProjectReorderOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectReorderOperator {
	node, err := NewProjectReorderOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectReorderOperatorTyped(projectReorderKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectReorderOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectReorderKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectReorderOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectReorderOperatorTyped(projectReorderKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectReorderOperator {
	node, err := NewProjectReorderOperatorTyped(projectReorderKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectReorderOperator) CloneTyped(includeDiagnostics bool) *ProjectReorderOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectReorderOperator{SyntaxNode: cloned}
}

func (n *ProjectReorderOperator) ProjectReorderKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectReorderOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type SampleOperator struct {
	*SyntaxNode
}

func (*SampleOperator) isQueryOperator() {}

func (*SampleOperator) isExpression() {}

func NewSampleOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SampleOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SampleOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SampleOperator{SyntaxNode: node}, nil
}

func MustNewSampleOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SampleOperator {
	node, err := NewSampleOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSampleOperatorTyped(sampleKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, diagnostics ...Diagnostic) (*SampleOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(sampleKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
	}
	created, err := NewSampleOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSampleOperatorTyped(sampleKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, diagnostics ...Diagnostic) *SampleOperator {
	node, err := NewSampleOperatorTyped(sampleKeyword, parameters, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SampleOperator) CloneTyped(includeDiagnostics bool) *SampleOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SampleOperator{SyntaxNode: cloned}
}

func (n *SampleOperator) SampleKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SampleOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SampleOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type SampleDistinctOperator struct {
	*SyntaxNode
}

func (*SampleDistinctOperator) isQueryOperator() {}

func (*SampleDistinctOperator) isExpression() {}

func NewSampleDistinctOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SampleDistinctOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SampleDistinctOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SampleDistinctOperator{SyntaxNode: node}, nil
}

func MustNewSampleDistinctOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SampleDistinctOperator {
	node, err := NewSampleDistinctOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSampleDistinctOperatorTyped(sampleDistinctKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, diagnostics ...Diagnostic) (*SampleDistinctOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(sampleDistinctKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(ofKeyword),
		syntaxElementFromTyped(ofExpression),
	}
	created, err := NewSampleDistinctOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSampleDistinctOperatorTyped(sampleDistinctKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, diagnostics ...Diagnostic) *SampleDistinctOperator {
	node, err := NewSampleDistinctOperatorTyped(sampleDistinctKeyword, parameters, expression, ofKeyword, ofExpression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SampleDistinctOperator) CloneTyped(includeDiagnostics bool) *SampleDistinctOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SampleDistinctOperator{SyntaxNode: cloned}
}

func (n *SampleDistinctOperator) SampleDistinctKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SampleDistinctOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SampleDistinctOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *SampleDistinctOperator) OfKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SampleDistinctOperator) OfExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type EntityGroup struct {
	*SyntaxNode
}

func (*EntityGroup) isExpression() {}

func NewEntityGroup(children []SyntaxElement, diagnostics ...Diagnostic) (*EntityGroup, error) {
	node, err := NewGeneratedSyntaxNodeByName("EntityGroup", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &EntityGroup{SyntaxNode: node}, nil
}

func MustNewEntityGroup(children []SyntaxElement, diagnostics ...Diagnostic) *EntityGroup {
	node, err := NewEntityGroup(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewEntityGroupTyped(entityGroupKeyword *SyntaxToken, openBracket *SyntaxToken, entities *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*EntityGroup, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(entityGroupKeyword),
		syntaxElementFromTyped(openBracket),
		syntaxElementFromSeparatedSyntaxListOf(entities),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewEntityGroup(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewEntityGroupTyped(entityGroupKeyword *SyntaxToken, openBracket *SyntaxToken, entities *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) *EntityGroup {
	node, err := NewEntityGroupTyped(entityGroupKeyword, openBracket, entities, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *EntityGroup) CloneTyped(includeDiagnostics bool) *EntityGroup {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &EntityGroup{SyntaxNode: cloned}
}

func (n *EntityGroup) EntityGroupKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EntityGroup) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EntityGroup) Entities() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *EntityGroup) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ReduceByOperator struct {
	*SyntaxNode
}

func (*ReduceByOperator) isQueryOperator() {}

func (*ReduceByOperator) isExpression() {}

func NewReduceByOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ReduceByOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ReduceByOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ReduceByOperator{SyntaxNode: node}, nil
}

func MustNewReduceByOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ReduceByOperator {
	node, err := NewReduceByOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewReduceByOperatorTyped(reduceKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, expression Expression, with *ReduceByWithClause, diagnostics ...Diagnostic) (*ReduceByOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(reduceKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(with),
	}
	created, err := NewReduceByOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewReduceByOperatorTyped(reduceKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, expression Expression, with *ReduceByWithClause, diagnostics ...Diagnostic) *ReduceByOperator {
	node, err := NewReduceByOperatorTyped(reduceKeyword, parameters, byKeyword, expression, with, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ReduceByOperator) CloneTyped(includeDiagnostics bool) *ReduceByOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ReduceByOperator{SyntaxNode: cloned}
}

func (n *ReduceByOperator) ReduceKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ReduceByOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *ReduceByOperator) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ReduceByOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ReduceByOperator) With() *ReduceByWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ReduceByWithClause)
	return child
}

type ReduceByWithClause struct {
	*SyntaxNode
}

func (*ReduceByWithClause) isClause() {}

func NewReduceByWithClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ReduceByWithClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ReduceByWithClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ReduceByWithClause{SyntaxNode: node}, nil
}

func MustNewReduceByWithClause(children []SyntaxElement, diagnostics ...Diagnostic) *ReduceByWithClause {
	node, err := NewReduceByWithClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewReduceByWithClauseTyped(withKeyword *SyntaxToken, parameters *SeparatedSyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) (*ReduceByWithClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromSeparatedSyntaxListOf(parameters),
	}
	created, err := NewReduceByWithClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewReduceByWithClauseTyped(withKeyword *SyntaxToken, parameters *SeparatedSyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) *ReduceByWithClause {
	node, err := NewReduceByWithClauseTyped(withKeyword, parameters, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ReduceByWithClause) CloneTyped(includeDiagnostics bool) *ReduceByWithClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ReduceByWithClause{SyntaxNode: cloned}
}

func (n *ReduceByWithClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ReduceByWithClause) Parameters() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

type SummarizeOperator struct {
	*SyntaxNode
}

func (*SummarizeOperator) isQueryOperator() {}

func (*SummarizeOperator) isExpression() {}

func NewSummarizeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SummarizeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SummarizeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SummarizeOperator{SyntaxNode: node}, nil
}

func MustNewSummarizeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SummarizeOperator {
	node, err := NewSummarizeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSummarizeOperatorTyped(summarizeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[Expression], byClause *SummarizeByClause, diagnostics ...Diagnostic) (*SummarizeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(summarizeKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(aggregates),
		syntaxElementFromTyped(byClause),
	}
	created, err := NewSummarizeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSummarizeOperatorTyped(summarizeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], aggregates *SeparatedSyntaxListOf[Expression], byClause *SummarizeByClause, diagnostics ...Diagnostic) *SummarizeOperator {
	node, err := NewSummarizeOperatorTyped(summarizeKeyword, parameters, aggregates, byClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SummarizeOperator) CloneTyped(includeDiagnostics bool) *SummarizeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SummarizeOperator{SyntaxNode: cloned}
}

func (n *SummarizeOperator) SummarizeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SummarizeOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SummarizeOperator) Aggregates() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *SummarizeOperator) ByClause() *SummarizeByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*SummarizeByClause)
	return child
}

type MacroExpandScopeReferenceName struct {
	*SyntaxNode
}

func (*MacroExpandScopeReferenceName) isClause() {}

func NewMacroExpandScopeReferenceName(children []SyntaxElement, diagnostics ...Diagnostic) (*MacroExpandScopeReferenceName, error) {
	node, err := NewGeneratedSyntaxNodeByName("MacroExpandScopeReferenceName", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MacroExpandScopeReferenceName{SyntaxNode: node}, nil
}

func MustNewMacroExpandScopeReferenceName(children []SyntaxElement, diagnostics ...Diagnostic) *MacroExpandScopeReferenceName {
	node, err := NewMacroExpandScopeReferenceName(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMacroExpandScopeReferenceNameTyped(asKeyword *SyntaxToken, entityGroupReferenceName *NameDeclaration, diagnostics ...Diagnostic) (*MacroExpandScopeReferenceName, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asKeyword),
		syntaxElementFromTyped(entityGroupReferenceName),
	}
	created, err := NewMacroExpandScopeReferenceName(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMacroExpandScopeReferenceNameTyped(asKeyword *SyntaxToken, entityGroupReferenceName *NameDeclaration, diagnostics ...Diagnostic) *MacroExpandScopeReferenceName {
	node, err := NewMacroExpandScopeReferenceNameTyped(asKeyword, entityGroupReferenceName, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MacroExpandScopeReferenceName) CloneTyped(includeDiagnostics bool) *MacroExpandScopeReferenceName {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MacroExpandScopeReferenceName{SyntaxNode: cloned}
}

func (n *MacroExpandScopeReferenceName) AsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MacroExpandScopeReferenceName) EntityGroupReferenceName() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

type MacroExpandOperator struct {
	*SyntaxNode
}

func (*MacroExpandOperator) isQueryOperator() {}

func (*MacroExpandOperator) isExpression() {}

func NewMacroExpandOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*MacroExpandOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("MacroExpandOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MacroExpandOperator{SyntaxNode: node}, nil
}

func MustNewMacroExpandOperator(children []SyntaxElement, diagnostics ...Diagnostic) *MacroExpandOperator {
	node, err := NewMacroExpandOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMacroExpandOperatorTyped(macroExpandKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], entityGroup Expression, scopeReferenceName *MacroExpandScopeReferenceName, openParen *SyntaxToken, statementList *SeparatedSyntaxListOf[Statement], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MacroExpandOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(macroExpandKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(entityGroup),
		syntaxElementFromTyped(scopeReferenceName),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(statementList),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMacroExpandOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMacroExpandOperatorTyped(macroExpandKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], entityGroup Expression, scopeReferenceName *MacroExpandScopeReferenceName, openParen *SyntaxToken, statementList *SeparatedSyntaxListOf[Statement], closeParen *SyntaxToken, diagnostics ...Diagnostic) *MacroExpandOperator {
	node, err := NewMacroExpandOperatorTyped(macroExpandKeyword, parameters, entityGroup, scopeReferenceName, openParen, statementList, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MacroExpandOperator) CloneTyped(includeDiagnostics bool) *MacroExpandOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MacroExpandOperator{SyntaxNode: cloned}
}

func (n *MacroExpandOperator) MacroExpandKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MacroExpandOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *MacroExpandOperator) EntityGroup() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MacroExpandOperator) ScopeReferenceName() *MacroExpandScopeReferenceName {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MacroExpandScopeReferenceName)
	return child
}

func (n *MacroExpandOperator) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MacroExpandOperator) StatementList() *SeparatedSyntaxListOf[Statement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Statement](list)
}

func (n *MacroExpandOperator) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SummarizeByClause struct {
	*SyntaxNode
}

func (*SummarizeByClause) isClause() {}

func NewSummarizeByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*SummarizeByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("SummarizeByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SummarizeByClause{SyntaxNode: node}, nil
}

func MustNewSummarizeByClause(children []SyntaxElement, diagnostics ...Diagnostic) *SummarizeByClause {
	node, err := NewSummarizeByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSummarizeByClauseTyped(byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], binClause NamedExpression, diagnostics ...Diagnostic) (*SummarizeByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(binClause),
	}
	created, err := NewSummarizeByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSummarizeByClauseTyped(byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], binClause NamedExpression, diagnostics ...Diagnostic) *SummarizeByClause {
	node, err := NewSummarizeByClauseTyped(byKeyword, expressions, binClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SummarizeByClause) CloneTyped(includeDiagnostics bool) *SummarizeByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SummarizeByClause{SyntaxNode: cloned}
}

func (n *SummarizeByClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SummarizeByClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *SummarizeByClause) BinClause() NamedExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(NamedExpression)
	return child
}

type DistinctOperator struct {
	*SyntaxNode
}

func (*DistinctOperator) isQueryOperator() {}

func (*DistinctOperator) isExpression() {}

func NewDistinctOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*DistinctOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("DistinctOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DistinctOperator{SyntaxNode: node}, nil
}

func MustNewDistinctOperator(children []SyntaxElement, diagnostics ...Diagnostic) *DistinctOperator {
	node, err := NewDistinctOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDistinctOperatorTyped(distinctKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*DistinctOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(distinctKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewDistinctOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDistinctOperatorTyped(distinctKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *DistinctOperator {
	node, err := NewDistinctOperatorTyped(distinctKeyword, parameters, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DistinctOperator) CloneTyped(includeDiagnostics bool) *DistinctOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DistinctOperator{SyntaxNode: cloned}
}

func (n *DistinctOperator) DistinctKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DistinctOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *DistinctOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type TakeOperator struct {
	*SyntaxNode
}

func (*TakeOperator) isQueryOperator() {}

func (*TakeOperator) isExpression() {}

func NewTakeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*TakeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("TakeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TakeOperator{SyntaxNode: node}, nil
}

func MustNewTakeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *TakeOperator {
	node, err := NewTakeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTakeOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, diagnostics ...Diagnostic) (*TakeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(keyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
	}
	created, err := NewTakeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTakeOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, diagnostics ...Diagnostic) *TakeOperator {
	node, err := NewTakeOperatorTyped(keyword, parameters, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TakeOperator) CloneTyped(includeDiagnostics bool) *TakeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TakeOperator{SyntaxNode: cloned}
}

func (n *TakeOperator) Keyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TakeOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *TakeOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type SortOperator struct {
	*SyntaxNode
}

func (*SortOperator) isQueryOperator() {}

func (*SortOperator) isExpression() {}

func NewSortOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SortOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SortOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SortOperator{SyntaxNode: node}, nil
}

func MustNewSortOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SortOperator {
	node, err := NewSortOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSortOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*SortOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(keyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewSortOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSortOperatorTyped(keyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *SortOperator {
	node, err := NewSortOperatorTyped(keyword, parameters, byKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SortOperator) CloneTyped(includeDiagnostics bool) *SortOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SortOperator{SyntaxNode: cloned}
}

func (n *SortOperator) Keyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SortOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SortOperator) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SortOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type OrderedExpression struct {
	*SyntaxNode
}

func (*OrderedExpression) isExpression() {}

func NewOrderedExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*OrderedExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("OrderedExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &OrderedExpression{SyntaxNode: node}, nil
}

func MustNewOrderedExpression(children []SyntaxElement, diagnostics ...Diagnostic) *OrderedExpression {
	node, err := NewOrderedExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewOrderedExpressionTyped(expression Expression, ordering *OrderingClause, diagnostics ...Diagnostic) (*OrderedExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(ordering),
	}
	created, err := NewOrderedExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewOrderedExpressionTyped(expression Expression, ordering *OrderingClause, diagnostics ...Diagnostic) *OrderedExpression {
	node, err := NewOrderedExpressionTyped(expression, ordering, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *OrderedExpression) CloneTyped(includeDiagnostics bool) *OrderedExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &OrderedExpression{SyntaxNode: cloned}
}

func (n *OrderedExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *OrderedExpression) Ordering() *OrderingClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*OrderingClause)
	return child
}

type OrderingClause struct {
	*SyntaxNode
}

func NewOrderingClause(children []SyntaxElement, diagnostics ...Diagnostic) (*OrderingClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("OrderingClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &OrderingClause{SyntaxNode: node}, nil
}

func MustNewOrderingClause(children []SyntaxElement, diagnostics ...Diagnostic) *OrderingClause {
	node, err := NewOrderingClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewOrderingClauseTyped(ascOrDescKeyword *SyntaxToken, nullsClause *OrderingNullsClause, diagnostics ...Diagnostic) (*OrderingClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(ascOrDescKeyword),
		syntaxElementFromTyped(nullsClause),
	}
	created, err := NewOrderingClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewOrderingClauseTyped(ascOrDescKeyword *SyntaxToken, nullsClause *OrderingNullsClause, diagnostics ...Diagnostic) *OrderingClause {
	node, err := NewOrderingClauseTyped(ascOrDescKeyword, nullsClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *OrderingClause) CloneTyped(includeDiagnostics bool) *OrderingClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &OrderingClause{SyntaxNode: cloned}
}

func (n *OrderingClause) AscOrDescKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *OrderingClause) NullsClause() *OrderingNullsClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*OrderingNullsClause)
	return child
}

type OrderingNullsClause struct {
	*SyntaxNode
}

func NewOrderingNullsClause(children []SyntaxElement, diagnostics ...Diagnostic) (*OrderingNullsClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("OrderingNullsClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &OrderingNullsClause{SyntaxNode: node}, nil
}

func MustNewOrderingNullsClause(children []SyntaxElement, diagnostics ...Diagnostic) *OrderingNullsClause {
	node, err := NewOrderingNullsClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewOrderingNullsClauseTyped(nullsKeyword *SyntaxToken, firstOrLastKeyword *SyntaxToken, diagnostics ...Diagnostic) (*OrderingNullsClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(nullsKeyword),
		syntaxElementFromTyped(firstOrLastKeyword),
	}
	created, err := NewOrderingNullsClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewOrderingNullsClauseTyped(nullsKeyword *SyntaxToken, firstOrLastKeyword *SyntaxToken, diagnostics ...Diagnostic) *OrderingNullsClause {
	node, err := NewOrderingNullsClauseTyped(nullsKeyword, firstOrLastKeyword, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *OrderingNullsClause) CloneTyped(includeDiagnostics bool) *OrderingNullsClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &OrderingNullsClause{SyntaxNode: cloned}
}

func (n *OrderingNullsClause) NullsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *OrderingNullsClause) FirstOrLastKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type TopHittersOperator struct {
	*SyntaxNode
}

func (*TopHittersOperator) isQueryOperator() {}

func (*TopHittersOperator) isExpression() {}

func NewTopHittersOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*TopHittersOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopHittersOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopHittersOperator{SyntaxNode: node}, nil
}

func MustNewTopHittersOperator(children []SyntaxElement, diagnostics ...Diagnostic) *TopHittersOperator {
	node, err := NewTopHittersOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopHittersOperatorTyped(topHittersKeyword *SyntaxToken, expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, byClause *TopHittersByClause, diagnostics ...Diagnostic) (*TopHittersOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(topHittersKeyword),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(ofKeyword),
		syntaxElementFromTyped(ofExpression),
		syntaxElementFromTyped(byClause),
	}
	created, err := NewTopHittersOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopHittersOperatorTyped(topHittersKeyword *SyntaxToken, expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, byClause *TopHittersByClause, diagnostics ...Diagnostic) *TopHittersOperator {
	node, err := NewTopHittersOperatorTyped(topHittersKeyword, expression, ofKeyword, ofExpression, byClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopHittersOperator) CloneTyped(includeDiagnostics bool) *TopHittersOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopHittersOperator{SyntaxNode: cloned}
}

func (n *TopHittersOperator) TopHittersKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopHittersOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *TopHittersOperator) OfKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopHittersOperator) OfExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *TopHittersOperator) ByClause() *TopHittersByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*TopHittersByClause)
	return child
}

type TopHittersByClause struct {
	*SyntaxNode
}

func (*TopHittersByClause) isClause() {}

func NewTopHittersByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*TopHittersByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopHittersByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopHittersByClause{SyntaxNode: node}, nil
}

func MustNewTopHittersByClause(children []SyntaxElement, diagnostics ...Diagnostic) *TopHittersByClause {
	node, err := NewTopHittersByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopHittersByClauseTyped(byKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*TopHittersByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewTopHittersByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopHittersByClauseTyped(byKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *TopHittersByClause {
	node, err := NewTopHittersByClauseTyped(byKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopHittersByClause) CloneTyped(includeDiagnostics bool) *TopHittersByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopHittersByClause{SyntaxNode: cloned}
}

func (n *TopHittersByClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopHittersByClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type TopOperator struct {
	*SyntaxNode
}

func (*TopOperator) isQueryOperator() {}

func (*TopOperator) isExpression() {}

func NewTopOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*TopOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopOperator{SyntaxNode: node}, nil
}

func MustNewTopOperator(children []SyntaxElement, diagnostics ...Diagnostic) *TopOperator {
	node, err := NewTopOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopOperatorTyped(topKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, byKeyword *SyntaxToken, byExpression Expression, diagnostics ...Diagnostic) (*TopOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(topKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(byExpression),
	}
	created, err := NewTopOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopOperatorTyped(topKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, byKeyword *SyntaxToken, byExpression Expression, diagnostics ...Diagnostic) *TopOperator {
	node, err := NewTopOperatorTyped(topKeyword, parameters, expression, byKeyword, byExpression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopOperator) CloneTyped(includeDiagnostics bool) *TopOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopOperator{SyntaxNode: cloned}
}

func (n *TopOperator) TopKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *TopOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *TopOperator) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopOperator) ByExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type TopNestedOperator struct {
	*SyntaxNode
}

func (*TopNestedOperator) isQueryOperator() {}

func (*TopNestedOperator) isExpression() {}

func NewTopNestedOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*TopNestedOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopNestedOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopNestedOperator{SyntaxNode: node}, nil
}

func MustNewTopNestedOperator(children []SyntaxElement, diagnostics ...Diagnostic) *TopNestedOperator {
	node, err := NewTopNestedOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopNestedOperatorTyped(clauses *SeparatedSyntaxListOf[*TopNestedClause], diagnostics ...Diagnostic) (*TopNestedOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromSeparatedSyntaxListOf(clauses),
	}
	created, err := NewTopNestedOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopNestedOperatorTyped(clauses *SeparatedSyntaxListOf[*TopNestedClause], diagnostics ...Diagnostic) *TopNestedOperator {
	node, err := NewTopNestedOperatorTyped(clauses, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopNestedOperator) CloneTyped(includeDiagnostics bool) *TopNestedOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopNestedOperator{SyntaxNode: cloned}
}

func (n *TopNestedOperator) Clauses() *SeparatedSyntaxListOf[*TopNestedClause] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*TopNestedClause](list)
}

type TopNestedClause struct {
	*SyntaxNode
}

func (*TopNestedClause) isClause() {}

func NewTopNestedClause(children []SyntaxElement, diagnostics ...Diagnostic) (*TopNestedClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopNestedClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopNestedClause{SyntaxNode: node}, nil
}

func MustNewTopNestedClause(children []SyntaxElement, diagnostics ...Diagnostic) *TopNestedClause {
	node, err := NewTopNestedClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopNestedClauseTyped(topNestedKeyword *SyntaxToken, expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, withOthersClause *TopNestedWithOthersClause, byKeyword *SyntaxToken, byExpression Expression, diagnostics ...Diagnostic) (*TopNestedClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(topNestedKeyword),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(ofKeyword),
		syntaxElementFromTyped(ofExpression),
		syntaxElementFromTyped(withOthersClause),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(byExpression),
	}
	created, err := NewTopNestedClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopNestedClauseTyped(topNestedKeyword *SyntaxToken, expression Expression, ofKeyword *SyntaxToken, ofExpression Expression, withOthersClause *TopNestedWithOthersClause, byKeyword *SyntaxToken, byExpression Expression, diagnostics ...Diagnostic) *TopNestedClause {
	node, err := NewTopNestedClauseTyped(topNestedKeyword, expression, ofKeyword, ofExpression, withOthersClause, byKeyword, byExpression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopNestedClause) CloneTyped(includeDiagnostics bool) *TopNestedClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopNestedClause{SyntaxNode: cloned}
}

func (n *TopNestedClause) TopNestedKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *TopNestedClause) OfKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedClause) OfExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *TopNestedClause) WithOthersClause() *TopNestedWithOthersClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*TopNestedWithOthersClause)
	return child
}

func (n *TopNestedClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedClause) ByExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type TopNestedWithOthersClause struct {
	*SyntaxNode
}

func (*TopNestedWithOthersClause) isClause() {}

func NewTopNestedWithOthersClause(children []SyntaxElement, diagnostics ...Diagnostic) (*TopNestedWithOthersClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("TopNestedWithOthersClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &TopNestedWithOthersClause{SyntaxNode: node}, nil
}

func MustNewTopNestedWithOthersClause(children []SyntaxElement, diagnostics ...Diagnostic) *TopNestedWithOthersClause {
	node, err := NewTopNestedWithOthersClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewTopNestedWithOthersClauseTyped(withKeyword *SyntaxToken, othersKeyword *SyntaxToken, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*TopNestedWithOthersClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(othersKeyword),
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(expression),
	}
	created, err := NewTopNestedWithOthersClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewTopNestedWithOthersClauseTyped(withKeyword *SyntaxToken, othersKeyword *SyntaxToken, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *TopNestedWithOthersClause {
	node, err := NewTopNestedWithOthersClauseTyped(withKeyword, othersKeyword, equal, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *TopNestedWithOthersClause) CloneTyped(includeDiagnostics bool) *TopNestedWithOthersClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &TopNestedWithOthersClause{SyntaxNode: cloned}
}

func (n *TopNestedWithOthersClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedWithOthersClause) OthersKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedWithOthersClause) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *TopNestedWithOthersClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type UnionOperator struct {
	*SyntaxNode
}

func (*UnionOperator) isQueryOperator() {}

func (*UnionOperator) isExpression() {}

func NewUnionOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*UnionOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("UnionOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &UnionOperator{SyntaxNode: node}, nil
}

func MustNewUnionOperator(children []SyntaxElement, diagnostics ...Diagnostic) *UnionOperator {
	node, err := NewUnionOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewUnionOperatorTyped(unionKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*UnionOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(unionKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewUnionOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewUnionOperatorTyped(unionKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *UnionOperator {
	node, err := NewUnionOperatorTyped(unionKeyword, parameters, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *UnionOperator) CloneTyped(includeDiagnostics bool) *UnionOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &UnionOperator{SyntaxNode: cloned}
}

func (n *UnionOperator) UnionKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *UnionOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *UnionOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type AsOperator struct {
	*SyntaxNode
}

func (*AsOperator) isQueryOperator() {}

func (*AsOperator) isExpression() {}

func NewAsOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*AsOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("AsOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &AsOperator{SyntaxNode: node}, nil
}

func MustNewAsOperator(children []SyntaxElement, diagnostics ...Diagnostic) *AsOperator {
	node, err := NewAsOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewAsOperatorTyped(asKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], name *NameDeclaration, diagnostics ...Diagnostic) (*AsOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(name),
	}
	created, err := NewAsOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewAsOperatorTyped(asKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], name *NameDeclaration, diagnostics ...Diagnostic) *AsOperator {
	node, err := NewAsOperatorTyped(asKeyword, parameters, name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *AsOperator) CloneTyped(includeDiagnostics bool) *AsOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &AsOperator{SyntaxNode: cloned}
}

func (n *AsOperator) AsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *AsOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *AsOperator) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

type SerializeOperator struct {
	*SyntaxNode
}

func (*SerializeOperator) isQueryOperator() {}

func (*SerializeOperator) isExpression() {}

func NewSerializeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*SerializeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("SerializeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SerializeOperator{SyntaxNode: node}, nil
}

func MustNewSerializeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *SerializeOperator {
	node, err := NewSerializeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSerializeOperatorTyped(serializeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*SerializeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(serializeKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewSerializeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSerializeOperatorTyped(serializeKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *SerializeOperator {
	node, err := NewSerializeOperatorTyped(serializeKeyword, parameters, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SerializeOperator) CloneTyped(includeDiagnostics bool) *SerializeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SerializeOperator{SyntaxNode: cloned}
}

func (n *SerializeOperator) SerializeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SerializeOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *SerializeOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type InvokeOperator struct {
	*SyntaxNode
}

func (*InvokeOperator) isQueryOperator() {}

func (*InvokeOperator) isExpression() {}

func NewInvokeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*InvokeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("InvokeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InvokeOperator{SyntaxNode: node}, nil
}

func MustNewInvokeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *InvokeOperator {
	node, err := NewInvokeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInvokeOperatorTyped(invokeKeyword *SyntaxToken, function Expression, diagnostics ...Diagnostic) (*InvokeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(invokeKeyword),
		syntaxElementFromTyped(function),
	}
	created, err := NewInvokeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInvokeOperatorTyped(invokeKeyword *SyntaxToken, function Expression, diagnostics ...Diagnostic) *InvokeOperator {
	node, err := NewInvokeOperatorTyped(invokeKeyword, function, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InvokeOperator) CloneTyped(includeDiagnostics bool) *InvokeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InvokeOperator{SyntaxNode: cloned}
}

func (n *InvokeOperator) InvokeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InvokeOperator) Function() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type RenderOperator struct {
	*SyntaxNode
}

func (*RenderOperator) isQueryOperator() {}

func (*RenderOperator) isExpression() {}

func NewRenderOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*RenderOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("RenderOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RenderOperator{SyntaxNode: node}, nil
}

func MustNewRenderOperator(children []SyntaxElement, diagnostics ...Diagnostic) *RenderOperator {
	node, err := NewRenderOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRenderOperatorTyped(renderKeyword *SyntaxToken, chartType *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], withClause *RenderWithClause, diagnostics ...Diagnostic) (*RenderOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(renderKeyword),
		syntaxElementFromTyped(chartType),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewRenderOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRenderOperatorTyped(renderKeyword *SyntaxToken, chartType *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], withClause *RenderWithClause, diagnostics ...Diagnostic) *RenderOperator {
	node, err := NewRenderOperatorTyped(renderKeyword, chartType, parameters, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RenderOperator) CloneTyped(includeDiagnostics bool) *RenderOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RenderOperator{SyntaxNode: cloned}
}

func (n *RenderOperator) RenderKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenderOperator) ChartType() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenderOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *RenderOperator) WithClause() *RenderWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RenderWithClause)
	return child
}

type MakeGraphOperator struct {
	*SyntaxNode
}

func (*MakeGraphOperator) isQueryOperator() {}

func (*MakeGraphOperator) isExpression() {}

func NewMakeGraphOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeGraphOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeGraphOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeGraphOperator{SyntaxNode: node}, nil
}

func MustNewMakeGraphOperator(children []SyntaxElement, diagnostics ...Diagnostic) *MakeGraphOperator {
	node, err := NewMakeGraphOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeGraphOperatorTyped(makeGraphKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], sourceColumn *NameReference, directionToken *SyntaxToken, targetColumn *NameReference, withClause MakeGraphWithClause, partitionedByClause *MakeGraphPartitionedByClause, diagnostics ...Diagnostic) (*MakeGraphOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(makeGraphKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(sourceColumn),
		syntaxElementFromTyped(directionToken),
		syntaxElementFromTyped(targetColumn),
		syntaxElementFromTyped(withClause),
		syntaxElementFromTyped(partitionedByClause),
	}
	created, err := NewMakeGraphOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeGraphOperatorTyped(makeGraphKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], sourceColumn *NameReference, directionToken *SyntaxToken, targetColumn *NameReference, withClause MakeGraphWithClause, partitionedByClause *MakeGraphPartitionedByClause, diagnostics ...Diagnostic) *MakeGraphOperator {
	node, err := NewMakeGraphOperatorTyped(makeGraphKeyword, parameters, sourceColumn, directionToken, targetColumn, withClause, partitionedByClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeGraphOperator) CloneTyped(includeDiagnostics bool) *MakeGraphOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeGraphOperator{SyntaxNode: cloned}
}

func (n *MakeGraphOperator) MakeGraphKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *MakeGraphOperator) SourceColumn() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *MakeGraphOperator) DirectionToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphOperator) TargetColumn() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *MakeGraphOperator) WithClause() MakeGraphWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(MakeGraphWithClause)
	return child
}

func (n *MakeGraphOperator) PartitionedByClause() *MakeGraphPartitionedByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*MakeGraphPartitionedByClause)
	return child
}

type MakeGraphWithTablesAndKeysClause struct {
	*SyntaxNode
}

func (*MakeGraphWithTablesAndKeysClause) isMakeGraphWithClause() {}

func NewMakeGraphWithTablesAndKeysClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeGraphWithTablesAndKeysClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeGraphWithTablesAndKeysClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeGraphWithTablesAndKeysClause{SyntaxNode: node}, nil
}

func MustNewMakeGraphWithTablesAndKeysClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeGraphWithTablesAndKeysClause {
	node, err := NewMakeGraphWithTablesAndKeysClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeGraphWithTablesAndKeysClauseTyped(withKeyword *SyntaxToken, tablesAndKeys *SeparatedSyntaxListOf[*MakeGraphTableAndKeyClause], diagnostics ...Diagnostic) (*MakeGraphWithTablesAndKeysClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromSeparatedSyntaxListOf(tablesAndKeys),
	}
	created, err := NewMakeGraphWithTablesAndKeysClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeGraphWithTablesAndKeysClauseTyped(withKeyword *SyntaxToken, tablesAndKeys *SeparatedSyntaxListOf[*MakeGraphTableAndKeyClause], diagnostics ...Diagnostic) *MakeGraphWithTablesAndKeysClause {
	node, err := NewMakeGraphWithTablesAndKeysClauseTyped(withKeyword, tablesAndKeys, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeGraphWithTablesAndKeysClause) CloneTyped(includeDiagnostics bool) *MakeGraphWithTablesAndKeysClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeGraphWithTablesAndKeysClause{SyntaxNode: cloned}
}

func (n *MakeGraphWithTablesAndKeysClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphWithTablesAndKeysClause) TablesAndKeys() *SeparatedSyntaxListOf[*MakeGraphTableAndKeyClause] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*MakeGraphTableAndKeyClause](list)
}

type MakeGraphWithImplicitIdClause struct {
	*SyntaxNode
}

func (*MakeGraphWithImplicitIdClause) isMakeGraphWithClause() {}

func NewMakeGraphWithImplicitIdClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeGraphWithImplicitIdClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeGraphWithImplicitIdClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeGraphWithImplicitIdClause{SyntaxNode: node}, nil
}

func MustNewMakeGraphWithImplicitIdClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeGraphWithImplicitIdClause {
	node, err := NewMakeGraphWithImplicitIdClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeGraphWithImplicitIdClauseTyped(withNodeIdKeyword *SyntaxToken, equalToken *SyntaxToken, name *NameDeclaration, diagnostics ...Diagnostic) (*MakeGraphWithImplicitIdClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withNodeIdKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(name),
	}
	created, err := NewMakeGraphWithImplicitIdClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeGraphWithImplicitIdClauseTyped(withNodeIdKeyword *SyntaxToken, equalToken *SyntaxToken, name *NameDeclaration, diagnostics ...Diagnostic) *MakeGraphWithImplicitIdClause {
	node, err := NewMakeGraphWithImplicitIdClauseTyped(withNodeIdKeyword, equalToken, name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeGraphWithImplicitIdClause) CloneTyped(includeDiagnostics bool) *MakeGraphWithImplicitIdClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeGraphWithImplicitIdClause{SyntaxNode: cloned}
}

func (n *MakeGraphWithImplicitIdClause) WithNodeIdKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphWithImplicitIdClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphWithImplicitIdClause) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

type GraphMarkComponentsOperator struct {
	*SyntaxNode
}

func (*GraphMarkComponentsOperator) isQueryOperator() {}

func (*GraphMarkComponentsOperator) isExpression() {}

func NewGraphMarkComponentsOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMarkComponentsOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMarkComponentsOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMarkComponentsOperator{SyntaxNode: node}, nil
}

func MustNewGraphMarkComponentsOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMarkComponentsOperator {
	node, err := NewGraphMarkComponentsOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMarkComponentsOperatorTyped(graphMarkComponentsKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) (*GraphMarkComponentsOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphMarkComponentsKeyword),
		syntaxElementFromSyntaxListOf(parameters),
	}
	created, err := NewGraphMarkComponentsOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMarkComponentsOperatorTyped(graphMarkComponentsKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) *GraphMarkComponentsOperator {
	node, err := NewGraphMarkComponentsOperatorTyped(graphMarkComponentsKeyword, parameters, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMarkComponentsOperator) CloneTyped(includeDiagnostics bool) *GraphMarkComponentsOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMarkComponentsOperator{SyntaxNode: cloned}
}

func (n *GraphMarkComponentsOperator) GraphMarkComponentsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMarkComponentsOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

type GraphWhereNodesOperator struct {
	*SyntaxNode
}

func (*GraphWhereNodesOperator) isQueryOperator() {}

func (*GraphWhereNodesOperator) isExpression() {}

func NewGraphWhereNodesOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphWhereNodesOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphWhereNodesOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphWhereNodesOperator{SyntaxNode: node}, nil
}

func MustNewGraphWhereNodesOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphWhereNodesOperator {
	node, err := NewGraphWhereNodesOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphWhereNodesOperatorTyped(graphWhereNodesKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) (*GraphWhereNodesOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphWhereNodesKeyword),
		syntaxElementFromTyped(condition),
	}
	created, err := NewGraphWhereNodesOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphWhereNodesOperatorTyped(graphWhereNodesKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) *GraphWhereNodesOperator {
	node, err := NewGraphWhereNodesOperatorTyped(graphWhereNodesKeyword, condition, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphWhereNodesOperator) CloneTyped(includeDiagnostics bool) *GraphWhereNodesOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphWhereNodesOperator{SyntaxNode: cloned}
}

func (n *GraphWhereNodesOperator) GraphWhereNodesKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphWhereNodesOperator) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type GraphWhereEdgesOperator struct {
	*SyntaxNode
}

func (*GraphWhereEdgesOperator) isQueryOperator() {}

func (*GraphWhereEdgesOperator) isExpression() {}

func NewGraphWhereEdgesOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphWhereEdgesOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphWhereEdgesOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphWhereEdgesOperator{SyntaxNode: node}, nil
}

func MustNewGraphWhereEdgesOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphWhereEdgesOperator {
	node, err := NewGraphWhereEdgesOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphWhereEdgesOperatorTyped(graphWhereEdgesKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) (*GraphWhereEdgesOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphWhereEdgesKeyword),
		syntaxElementFromTyped(condition),
	}
	created, err := NewGraphWhereEdgesOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphWhereEdgesOperatorTyped(graphWhereEdgesKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) *GraphWhereEdgesOperator {
	node, err := NewGraphWhereEdgesOperatorTyped(graphWhereEdgesKeyword, condition, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphWhereEdgesOperator) CloneTyped(includeDiagnostics bool) *GraphWhereEdgesOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphWhereEdgesOperator{SyntaxNode: cloned}
}

func (n *GraphWhereEdgesOperator) GraphWhereEdgesKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphWhereEdgesOperator) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MakeGraphTableAndKeyClause struct {
	*SyntaxNode
}

func NewMakeGraphTableAndKeyClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeGraphTableAndKeyClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeGraphTableAndKeyClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeGraphTableAndKeyClause{SyntaxNode: node}, nil
}

func MustNewMakeGraphTableAndKeyClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeGraphTableAndKeyClause {
	node, err := NewMakeGraphTableAndKeyClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeGraphTableAndKeyClauseTyped(table Expression, onKeyword *SyntaxToken, column *NameReference, diagnostics ...Diagnostic) (*MakeGraphTableAndKeyClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(table),
		syntaxElementFromTyped(onKeyword),
		syntaxElementFromTyped(column),
	}
	created, err := NewMakeGraphTableAndKeyClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeGraphTableAndKeyClauseTyped(table Expression, onKeyword *SyntaxToken, column *NameReference, diagnostics ...Diagnostic) *MakeGraphTableAndKeyClause {
	node, err := NewMakeGraphTableAndKeyClauseTyped(table, onKeyword, column, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeGraphTableAndKeyClause) CloneTyped(includeDiagnostics bool) *MakeGraphTableAndKeyClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeGraphTableAndKeyClause{SyntaxNode: cloned}
}

func (n *MakeGraphTableAndKeyClause) Table() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MakeGraphTableAndKeyClause) OnKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphTableAndKeyClause) Column() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

type MakeGraphPartitionedByClause struct {
	*SyntaxNode
}

func NewMakeGraphPartitionedByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*MakeGraphPartitionedByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("MakeGraphPartitionedByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MakeGraphPartitionedByClause{SyntaxNode: node}, nil
}

func MustNewMakeGraphPartitionedByClause(children []SyntaxElement, diagnostics ...Diagnostic) *MakeGraphPartitionedByClause {
	node, err := NewMakeGraphPartitionedByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMakeGraphPartitionedByClauseTyped(partitionedByKeyword *SyntaxToken, entity *NameReference, openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MakeGraphPartitionedByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionedByKeyword),
		syntaxElementFromTyped(entity),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(subquery),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMakeGraphPartitionedByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMakeGraphPartitionedByClauseTyped(partitionedByKeyword *SyntaxToken, entity *NameReference, openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *MakeGraphPartitionedByClause {
	node, err := NewMakeGraphPartitionedByClauseTyped(partitionedByKeyword, entity, openParen, subquery, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MakeGraphPartitionedByClause) CloneTyped(includeDiagnostics bool) *MakeGraphPartitionedByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MakeGraphPartitionedByClause{SyntaxNode: cloned}
}

func (n *MakeGraphPartitionedByClause) PartitionedByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphPartitionedByClause) Entity() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *MakeGraphPartitionedByClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MakeGraphPartitionedByClause) Subquery() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MakeGraphPartitionedByClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type GraphToTableOperator struct {
	*SyntaxNode
}

func (*GraphToTableOperator) isQueryOperator() {}

func (*GraphToTableOperator) isExpression() {}

func NewGraphToTableOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphToTableOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphToTableOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphToTableOperator{SyntaxNode: node}, nil
}

func MustNewGraphToTableOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphToTableOperator {
	node, err := NewGraphToTableOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphToTableOperatorTyped(graphToTableKeyword *SyntaxToken, outputClause *SeparatedSyntaxListOf[*GraphToTableOutputClause], diagnostics ...Diagnostic) (*GraphToTableOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphToTableKeyword),
		syntaxElementFromSeparatedSyntaxListOf(outputClause),
	}
	created, err := NewGraphToTableOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphToTableOperatorTyped(graphToTableKeyword *SyntaxToken, outputClause *SeparatedSyntaxListOf[*GraphToTableOutputClause], diagnostics ...Diagnostic) *GraphToTableOperator {
	node, err := NewGraphToTableOperatorTyped(graphToTableKeyword, outputClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphToTableOperator) CloneTyped(includeDiagnostics bool) *GraphToTableOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphToTableOperator{SyntaxNode: cloned}
}

func (n *GraphToTableOperator) GraphToTableKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphToTableOperator) OutputClause() *SeparatedSyntaxListOf[*GraphToTableOutputClause] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*GraphToTableOutputClause](list)
}

type GraphToTableOutputClause struct {
	*SyntaxNode
}

func (*GraphToTableOutputClause) isClause() {}

func NewGraphToTableOutputClause(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphToTableOutputClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphToTableOutputClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphToTableOutputClause{SyntaxNode: node}, nil
}

func MustNewGraphToTableOutputClause(children []SyntaxElement, diagnostics ...Diagnostic) *GraphToTableOutputClause {
	node, err := NewGraphToTableOutputClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphToTableOutputClauseTyped(entityKeyword *SyntaxToken, asClause *GraphToTableAsClause, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) (*GraphToTableOutputClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(entityKeyword),
		syntaxElementFromTyped(asClause),
		syntaxElementFromSyntaxListOf(parameters),
	}
	created, err := NewGraphToTableOutputClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphToTableOutputClauseTyped(entityKeyword *SyntaxToken, asClause *GraphToTableAsClause, parameters *SyntaxListOf[*NamedParameter], diagnostics ...Diagnostic) *GraphToTableOutputClause {
	node, err := NewGraphToTableOutputClauseTyped(entityKeyword, asClause, parameters, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphToTableOutputClause) CloneTyped(includeDiagnostics bool) *GraphToTableOutputClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphToTableOutputClause{SyntaxNode: cloned}
}

func (n *GraphToTableOutputClause) EntityKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphToTableOutputClause) AsClause() *GraphToTableAsClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*GraphToTableAsClause)
	return child
}

func (n *GraphToTableOutputClause) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

type GraphToTableAsClause struct {
	*SyntaxNode
}

func (*GraphToTableAsClause) isClause() {}

func NewGraphToTableAsClause(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphToTableAsClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphToTableAsClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphToTableAsClause{SyntaxNode: node}, nil
}

func MustNewGraphToTableAsClause(children []SyntaxElement, diagnostics ...Diagnostic) *GraphToTableAsClause {
	node, err := NewGraphToTableAsClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphToTableAsClauseTyped(asKeyword *SyntaxToken, name *NameDeclaration, diagnostics ...Diagnostic) (*GraphToTableAsClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asKeyword),
		syntaxElementFromTyped(name),
	}
	created, err := NewGraphToTableAsClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphToTableAsClauseTyped(asKeyword *SyntaxToken, name *NameDeclaration, diagnostics ...Diagnostic) *GraphToTableAsClause {
	node, err := NewGraphToTableAsClauseTyped(asKeyword, name, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphToTableAsClause) CloneTyped(includeDiagnostics bool) *GraphToTableAsClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphToTableAsClause{SyntaxNode: cloned}
}

func (n *GraphToTableAsClause) AsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphToTableAsClause) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

type GraphMatchOperator struct {
	*SyntaxNode
}

func (*GraphMatchOperator) isQueryOperator() {}

func (*GraphMatchOperator) isExpression() {}

func NewGraphMatchOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMatchOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMatchOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMatchOperator{SyntaxNode: node}, nil
}

func MustNewGraphMatchOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMatchOperator {
	node, err := NewGraphMatchOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMatchOperatorTyped(graphMatchKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], whereClause *WhereClause, projectClause *ProjectClause, diagnostics ...Diagnostic) (*GraphMatchOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphMatchKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(patterns),
		syntaxElementFromTyped(whereClause),
		syntaxElementFromTyped(projectClause),
	}
	created, err := NewGraphMatchOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMatchOperatorTyped(graphMatchKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], whereClause *WhereClause, projectClause *ProjectClause, diagnostics ...Diagnostic) *GraphMatchOperator {
	node, err := NewGraphMatchOperatorTyped(graphMatchKeyword, parameters, patterns, whereClause, projectClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMatchOperator) CloneTyped(includeDiagnostics bool) *GraphMatchOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMatchOperator{SyntaxNode: cloned}
}

func (n *GraphMatchOperator) GraphMatchKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMatchOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *GraphMatchOperator) Patterns() *SeparatedSyntaxListOf[*GraphMatchPattern] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*GraphMatchPattern](list)
}

func (n *GraphMatchOperator) WhereClause() *WhereClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*WhereClause)
	return child
}

func (n *GraphMatchOperator) ProjectClause() *ProjectClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ProjectClause)
	return child
}

type GraphShortestPathsOperator struct {
	*SyntaxNode
}

func (*GraphShortestPathsOperator) isQueryOperator() {}

func (*GraphShortestPathsOperator) isExpression() {}

func NewGraphShortestPathsOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphShortestPathsOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphShortestPathsOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphShortestPathsOperator{SyntaxNode: node}, nil
}

func MustNewGraphShortestPathsOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphShortestPathsOperator {
	node, err := NewGraphShortestPathsOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphShortestPathsOperatorTyped(graphShortestPathsKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], whereClause *WhereClause, projectClause *ProjectClause, diagnostics ...Diagnostic) (*GraphShortestPathsOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphShortestPathsKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromSeparatedSyntaxListOf(patterns),
		syntaxElementFromTyped(whereClause),
		syntaxElementFromTyped(projectClause),
	}
	created, err := NewGraphShortestPathsOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphShortestPathsOperatorTyped(graphShortestPathsKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], patterns *SeparatedSyntaxListOf[*GraphMatchPattern], whereClause *WhereClause, projectClause *ProjectClause, diagnostics ...Diagnostic) *GraphShortestPathsOperator {
	node, err := NewGraphShortestPathsOperatorTyped(graphShortestPathsKeyword, parameters, patterns, whereClause, projectClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphShortestPathsOperator) CloneTyped(includeDiagnostics bool) *GraphShortestPathsOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphShortestPathsOperator{SyntaxNode: cloned}
}

func (n *GraphShortestPathsOperator) GraphShortestPathsKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphShortestPathsOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *GraphShortestPathsOperator) Patterns() *SeparatedSyntaxListOf[*GraphMatchPattern] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*GraphMatchPattern](list)
}

func (n *GraphShortestPathsOperator) WhereClause() *WhereClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*WhereClause)
	return child
}

func (n *GraphShortestPathsOperator) ProjectClause() *ProjectClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ProjectClause)
	return child
}

type GraphMatchPattern struct {
	*SyntaxNode
}

func (*GraphMatchPattern) isQueryOperator() {}

func (*GraphMatchPattern) isExpression() {}

func NewGraphMatchPattern(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMatchPattern, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMatchPattern", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMatchPattern{SyntaxNode: node}, nil
}

func MustNewGraphMatchPattern(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMatchPattern {
	node, err := NewGraphMatchPattern(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMatchPatternTyped(patternElements *SyntaxListOf[GraphMatchPatternNotation], diagnostics ...Diagnostic) (*GraphMatchPattern, error) {
	children := []SyntaxElement{
		syntaxElementFromSyntaxListOf(patternElements),
	}
	created, err := NewGraphMatchPattern(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMatchPatternTyped(patternElements *SyntaxListOf[GraphMatchPatternNotation], diagnostics ...Diagnostic) *GraphMatchPattern {
	node, err := NewGraphMatchPatternTyped(patternElements, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMatchPattern) CloneTyped(includeDiagnostics bool) *GraphMatchPattern {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMatchPattern{SyntaxNode: cloned}
}

func (n *GraphMatchPattern) PatternElements() *SyntaxListOf[GraphMatchPatternNotation] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[GraphMatchPatternNotation](list)
}

type GraphMatchPatternNode struct {
	*SyntaxNode
}

func (*GraphMatchPatternNode) isGraphMatchPatternNotation() {}

func NewGraphMatchPatternNode(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMatchPatternNode, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMatchPatternNode", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMatchPatternNode{SyntaxNode: node}, nil
}

func MustNewGraphMatchPatternNode(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMatchPatternNode {
	node, err := NewGraphMatchPatternNode(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMatchPatternNodeTyped(open *SyntaxToken, name *NameDeclaration, close *SyntaxToken, diagnostics ...Diagnostic) (*GraphMatchPatternNode, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(open),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(close),
	}
	created, err := NewGraphMatchPatternNode(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMatchPatternNodeTyped(open *SyntaxToken, name *NameDeclaration, close *SyntaxToken, diagnostics ...Diagnostic) *GraphMatchPatternNode {
	node, err := NewGraphMatchPatternNodeTyped(open, name, close, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMatchPatternNode) CloneTyped(includeDiagnostics bool) *GraphMatchPatternNode {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMatchPatternNode{SyntaxNode: cloned}
}

func (n *GraphMatchPatternNode) Open() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMatchPatternNode) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *GraphMatchPatternNode) Close() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type GraphMatchPatternEdge struct {
	*SyntaxNode
}

func (*GraphMatchPatternEdge) isGraphMatchPatternNotation() {}

func NewGraphMatchPatternEdge(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMatchPatternEdge, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMatchPatternEdge", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMatchPatternEdge{SyntaxNode: node}, nil
}

func MustNewGraphMatchPatternEdge(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMatchPatternEdge {
	node, err := NewGraphMatchPatternEdge(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMatchPatternEdgeTyped(firstToken *SyntaxToken, name *NameDeclaration, rangeChild *GraphMatchPatternEdgeRange, lastToken *SyntaxToken, diagnostics ...Diagnostic) (*GraphMatchPatternEdge, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(firstToken),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(rangeChild),
		syntaxElementFromTyped(lastToken),
	}
	created, err := NewGraphMatchPatternEdge(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMatchPatternEdgeTyped(firstToken *SyntaxToken, name *NameDeclaration, rangeChild *GraphMatchPatternEdgeRange, lastToken *SyntaxToken, diagnostics ...Diagnostic) *GraphMatchPatternEdge {
	node, err := NewGraphMatchPatternEdgeTyped(firstToken, name, rangeChild, lastToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMatchPatternEdge) CloneTyped(includeDiagnostics bool) *GraphMatchPatternEdge {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMatchPatternEdge{SyntaxNode: cloned}
}

func (n *GraphMatchPatternEdge) FirstToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMatchPatternEdge) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *GraphMatchPatternEdge) Range() *GraphMatchPatternEdgeRange {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*GraphMatchPatternEdgeRange)
	return child
}

func (n *GraphMatchPatternEdge) LastToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type GraphMatchPatternEdgeRange struct {
	*SyntaxNode
}

func NewGraphMatchPatternEdgeRange(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMatchPatternEdgeRange, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMatchPatternEdgeRange", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMatchPatternEdgeRange{SyntaxNode: node}, nil
}

func MustNewGraphMatchPatternEdgeRange(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMatchPatternEdgeRange {
	node, err := NewGraphMatchPatternEdgeRange(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMatchPatternEdgeRangeTyped(asterisk *SyntaxToken, rangeStart Expression, dotDotToken *SyntaxToken, rangeEnd Expression, diagnostics ...Diagnostic) (*GraphMatchPatternEdgeRange, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(asterisk),
		syntaxElementFromTyped(rangeStart),
		syntaxElementFromTyped(dotDotToken),
		syntaxElementFromTyped(rangeEnd),
	}
	created, err := NewGraphMatchPatternEdgeRange(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMatchPatternEdgeRangeTyped(asterisk *SyntaxToken, rangeStart Expression, dotDotToken *SyntaxToken, rangeEnd Expression, diagnostics ...Diagnostic) *GraphMatchPatternEdgeRange {
	node, err := NewGraphMatchPatternEdgeRangeTyped(asterisk, rangeStart, dotDotToken, rangeEnd, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMatchPatternEdgeRange) CloneTyped(includeDiagnostics bool) *GraphMatchPatternEdgeRange {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMatchPatternEdgeRange{SyntaxNode: cloned}
}

func (n *GraphMatchPatternEdgeRange) Asterisk() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMatchPatternEdgeRange) RangeStart() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *GraphMatchPatternEdgeRange) DotDotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMatchPatternEdgeRange) RangeEnd() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type WhereClause struct {
	*SyntaxNode
}

func NewWhereClause(children []SyntaxElement, diagnostics ...Diagnostic) (*WhereClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("WhereClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &WhereClause{SyntaxNode: node}, nil
}

func MustNewWhereClause(children []SyntaxElement, diagnostics ...Diagnostic) *WhereClause {
	node, err := NewWhereClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewWhereClauseTyped(whereKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) (*WhereClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(whereKeyword),
		syntaxElementFromTyped(condition),
	}
	created, err := NewWhereClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewWhereClauseTyped(whereKeyword *SyntaxToken, condition Expression, diagnostics ...Diagnostic) *WhereClause {
	node, err := NewWhereClauseTyped(whereKeyword, condition, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *WhereClause) CloneTyped(includeDiagnostics bool) *WhereClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &WhereClause{SyntaxNode: cloned}
}

func (n *WhereClause) WhereKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *WhereClause) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type ProjectClause struct {
	*SyntaxNode
}

func NewProjectClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ProjectClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ProjectClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ProjectClause{SyntaxNode: node}, nil
}

func MustNewProjectClause(children []SyntaxElement, diagnostics ...Diagnostic) *ProjectClause {
	node, err := NewProjectClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewProjectClauseTyped(projectKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ProjectClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(projectKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewProjectClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewProjectClauseTyped(projectKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ProjectClause {
	node, err := NewProjectClauseTyped(projectKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ProjectClause) CloneTyped(includeDiagnostics bool) *ProjectClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ProjectClause{SyntaxNode: cloned}
}

func (n *ProjectClause) ProjectKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ProjectClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type NameReferenceList struct {
	*SyntaxNode
}

func (*NameReferenceList) isExpression() {}

func NewNameReferenceList(children []SyntaxElement, diagnostics ...Diagnostic) (*NameReferenceList, error) {
	node, err := NewGeneratedSyntaxNodeByName("NameReferenceList", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &NameReferenceList{SyntaxNode: node}, nil
}

func MustNewNameReferenceList(children []SyntaxElement, diagnostics ...Diagnostic) *NameReferenceList {
	node, err := NewNameReferenceList(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewNameReferenceListTyped(names *SeparatedSyntaxListOf[*NameReference], diagnostics ...Diagnostic) (*NameReferenceList, error) {
	children := []SyntaxElement{
		syntaxElementFromSeparatedSyntaxListOf(names),
	}
	created, err := NewNameReferenceList(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewNameReferenceListTyped(names *SeparatedSyntaxListOf[*NameReference], diagnostics ...Diagnostic) *NameReferenceList {
	node, err := NewNameReferenceListTyped(names, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *NameReferenceList) CloneTyped(includeDiagnostics bool) *NameReferenceList {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &NameReferenceList{SyntaxNode: cloned}
}

func (n *NameReferenceList) Names() *SeparatedSyntaxListOf[*NameReference] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NameReference](list)
}

type RenderWithClause struct {
	*SyntaxNode
}

func (*RenderWithClause) isClause() {}

func NewRenderWithClause(children []SyntaxElement, diagnostics ...Diagnostic) (*RenderWithClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("RenderWithClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RenderWithClause{SyntaxNode: node}, nil
}

func MustNewRenderWithClause(children []SyntaxElement, diagnostics ...Diagnostic) *RenderWithClause {
	node, err := NewRenderWithClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRenderWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, leadingComma *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*RenderWithClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(leadingComma),
		syntaxElementFromSeparatedSyntaxListOf(properties),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewRenderWithClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRenderWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, leadingComma *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *RenderWithClause {
	node, err := NewRenderWithClauseTyped(withKeyword, openParen, leadingComma, properties, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RenderWithClause) CloneTyped(includeDiagnostics bool) *RenderWithClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RenderWithClause{SyntaxNode: cloned}
}

func (n *RenderWithClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenderWithClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenderWithClause) LeadingComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RenderWithClause) Properties() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

func (n *RenderWithClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PrintOperator struct {
	*SyntaxNode
}

func (*PrintOperator) isQueryOperator() {}

func (*PrintOperator) isExpression() {}

func NewPrintOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*PrintOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("PrintOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PrintOperator{SyntaxNode: node}, nil
}

func MustNewPrintOperator(children []SyntaxElement, diagnostics ...Diagnostic) *PrintOperator {
	node, err := NewPrintOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPrintOperatorTyped(printKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*PrintOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(printKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewPrintOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPrintOperatorTyped(printKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *PrintOperator {
	node, err := NewPrintOperatorTyped(printKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PrintOperator) CloneTyped(includeDiagnostics bool) *PrintOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PrintOperator{SyntaxNode: cloned}
}

func (n *PrintOperator) PrintKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PrintOperator) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type AliasStatement struct {
	*SyntaxNode
}

func (*AliasStatement) isStatement() {}

func NewAliasStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*AliasStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("AliasStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &AliasStatement{SyntaxNode: node}, nil
}

func MustNewAliasStatement(children []SyntaxElement, diagnostics ...Diagnostic) *AliasStatement {
	node, err := NewAliasStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewAliasStatementTyped(aliasKeyword *SyntaxToken, databaseKeyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*AliasStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(aliasKeyword),
		syntaxElementFromTyped(databaseKeyword),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(expression),
	}
	created, err := NewAliasStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewAliasStatementTyped(aliasKeyword *SyntaxToken, databaseKeyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *AliasStatement {
	node, err := NewAliasStatementTyped(aliasKeyword, databaseKeyword, name, equal, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *AliasStatement) CloneTyped(includeDiagnostics bool) *AliasStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &AliasStatement{SyntaxNode: cloned}
}

func (n *AliasStatement) AliasKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *AliasStatement) DatabaseKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *AliasStatement) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *AliasStatement) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *AliasStatement) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type LetStatement struct {
	*SyntaxNode
}

func (*LetStatement) isStatement() {}

func NewLetStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*LetStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("LetStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &LetStatement{SyntaxNode: node}, nil
}

func MustNewLetStatement(children []SyntaxElement, diagnostics ...Diagnostic) *LetStatement {
	node, err := NewLetStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewLetStatementTyped(letKeyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*LetStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(letKeyword),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(expression),
	}
	created, err := NewLetStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewLetStatementTyped(letKeyword *SyntaxToken, name *NameDeclaration, equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *LetStatement {
	node, err := NewLetStatementTyped(letKeyword, name, equal, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *LetStatement) CloneTyped(includeDiagnostics bool) *LetStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &LetStatement{SyntaxNode: cloned}
}

func (n *LetStatement) LetKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *LetStatement) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *LetStatement) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *LetStatement) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type FunctionDeclaration struct {
	*SyntaxNode
}

func (*FunctionDeclaration) isExpression() {}

func NewFunctionDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*FunctionDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("FunctionDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FunctionDeclaration{SyntaxNode: node}, nil
}

func MustNewFunctionDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *FunctionDeclaration {
	node, err := NewFunctionDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFunctionDeclarationTyped(viewKeyword *SyntaxToken, parameters *FunctionParameters, body *FunctionBody, diagnostics ...Diagnostic) (*FunctionDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(viewKeyword),
		syntaxElementFromTyped(parameters),
		syntaxElementFromTyped(body),
	}
	created, err := NewFunctionDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFunctionDeclarationTyped(viewKeyword *SyntaxToken, parameters *FunctionParameters, body *FunctionBody, diagnostics ...Diagnostic) *FunctionDeclaration {
	node, err := NewFunctionDeclarationTyped(viewKeyword, parameters, body, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FunctionDeclaration) CloneTyped(includeDiagnostics bool) *FunctionDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FunctionDeclaration{SyntaxNode: cloned}
}

func (n *FunctionDeclaration) ViewKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FunctionDeclaration) Parameters() *FunctionParameters {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FunctionParameters)
	return child
}

func (n *FunctionDeclaration) Body() *FunctionBody {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FunctionBody)
	return child
}

type FunctionParameters struct {
	*SyntaxNode
}

func NewFunctionParameters(children []SyntaxElement, diagnostics ...Diagnostic) (*FunctionParameters, error) {
	node, err := NewGeneratedSyntaxNodeByName("FunctionParameters", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FunctionParameters{SyntaxNode: node}, nil
}

func MustNewFunctionParameters(children []SyntaxElement, diagnostics ...Diagnostic) *FunctionParameters {
	node, err := NewFunctionParameters(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFunctionParametersTyped(openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*FunctionParameters, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(parameters),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewFunctionParameters(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFunctionParametersTyped(openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *FunctionParameters {
	node, err := NewFunctionParametersTyped(openParen, parameters, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FunctionParameters) CloneTyped(includeDiagnostics bool) *FunctionParameters {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FunctionParameters{SyntaxNode: cloned}
}

func (n *FunctionParameters) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FunctionParameters) Parameters() *SeparatedSyntaxListOf[*FunctionParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*FunctionParameter](list)
}

func (n *FunctionParameters) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type FunctionParameter struct {
	*SyntaxNode
}

func NewFunctionParameter(children []SyntaxElement, diagnostics ...Diagnostic) (*FunctionParameter, error) {
	node, err := NewGeneratedSyntaxNodeByName("FunctionParameter", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FunctionParameter{SyntaxNode: node}, nil
}

func MustNewFunctionParameter(children []SyntaxElement, diagnostics ...Diagnostic) *FunctionParameter {
	node, err := NewFunctionParameter(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFunctionParameterTyped(nameAndType *NameAndTypeDeclaration, defaultValue *DefaultValueDeclaration, diagnostics ...Diagnostic) (*FunctionParameter, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(nameAndType),
		syntaxElementFromTyped(defaultValue),
	}
	created, err := NewFunctionParameter(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFunctionParameterTyped(nameAndType *NameAndTypeDeclaration, defaultValue *DefaultValueDeclaration, diagnostics ...Diagnostic) *FunctionParameter {
	node, err := NewFunctionParameterTyped(nameAndType, defaultValue, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FunctionParameter) CloneTyped(includeDiagnostics bool) *FunctionParameter {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FunctionParameter{SyntaxNode: cloned}
}

func (n *FunctionParameter) NameAndType() *NameAndTypeDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameAndTypeDeclaration)
	return child
}

func (n *FunctionParameter) DefaultValue() *DefaultValueDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*DefaultValueDeclaration)
	return child
}

type DefaultValueDeclaration struct {
	*SyntaxNode
}

func NewDefaultValueDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*DefaultValueDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("DefaultValueDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DefaultValueDeclaration{SyntaxNode: node}, nil
}

func MustNewDefaultValueDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *DefaultValueDeclaration {
	node, err := NewDefaultValueDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDefaultValueDeclarationTyped(equal *SyntaxToken, value Expression, diagnostics ...Diagnostic) (*DefaultValueDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(value),
	}
	created, err := NewDefaultValueDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDefaultValueDeclarationTyped(equal *SyntaxToken, value Expression, diagnostics ...Diagnostic) *DefaultValueDeclaration {
	node, err := NewDefaultValueDeclarationTyped(equal, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DefaultValueDeclaration) CloneTyped(includeDiagnostics bool) *DefaultValueDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DefaultValueDeclaration{SyntaxNode: cloned}
}

func (n *DefaultValueDeclaration) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DefaultValueDeclaration) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type FunctionBody struct {
	*SyntaxNode
}

func NewFunctionBody(children []SyntaxElement, diagnostics ...Diagnostic) (*FunctionBody, error) {
	node, err := NewGeneratedSyntaxNodeByName("FunctionBody", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &FunctionBody{SyntaxNode: node}, nil
}

func MustNewFunctionBody(children []SyntaxElement, diagnostics ...Diagnostic) *FunctionBody {
	node, err := NewFunctionBody(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewFunctionBodyTyped(openBrace *SyntaxToken, statements *SeparatedSyntaxListOf[Statement], expression Expression, semicolon *SyntaxToken, closeBrace *SyntaxToken, diagnostics ...Diagnostic) (*FunctionBody, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBrace),
		syntaxElementFromSeparatedSyntaxListOf(statements),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(semicolon),
		syntaxElementFromTyped(closeBrace),
	}
	created, err := NewFunctionBody(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewFunctionBodyTyped(openBrace *SyntaxToken, statements *SeparatedSyntaxListOf[Statement], expression Expression, semicolon *SyntaxToken, closeBrace *SyntaxToken, diagnostics ...Diagnostic) *FunctionBody {
	node, err := NewFunctionBodyTyped(openBrace, statements, expression, semicolon, closeBrace, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *FunctionBody) CloneTyped(includeDiagnostics bool) *FunctionBody {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &FunctionBody{SyntaxNode: cloned}
}

func (n *FunctionBody) OpenBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FunctionBody) Statements() *SeparatedSyntaxListOf[Statement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Statement](list)
}

func (n *FunctionBody) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *FunctionBody) Semicolon() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *FunctionBody) CloseBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SchemaTypeExpression struct {
	*SyntaxNode
}

func (*SchemaTypeExpression) isTypeExpression() {}

func (*SchemaTypeExpression) isExpression() {}

func NewSchemaTypeExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*SchemaTypeExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("SchemaTypeExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SchemaTypeExpression{SyntaxNode: node}, nil
}

func MustNewSchemaTypeExpression(children []SyntaxElement, diagnostics ...Diagnostic) *SchemaTypeExpression {
	node, err := NewSchemaTypeExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSchemaTypeExpressionTyped(openParen *SyntaxToken, columns *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*SchemaTypeExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(columns),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewSchemaTypeExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSchemaTypeExpressionTyped(openParen *SyntaxToken, columns *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) *SchemaTypeExpression {
	node, err := NewSchemaTypeExpressionTyped(openParen, columns, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SchemaTypeExpression) CloneTyped(includeDiagnostics bool) *SchemaTypeExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SchemaTypeExpression{SyntaxNode: cloned}
}

func (n *SchemaTypeExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SchemaTypeExpression) Columns() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *SchemaTypeExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExpressionStatement struct {
	*SyntaxNode
}

func (*ExpressionStatement) isStatement() {}

func NewExpressionStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*ExpressionStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExpressionStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExpressionStatement{SyntaxNode: node}, nil
}

func MustNewExpressionStatement(children []SyntaxElement, diagnostics ...Diagnostic) *ExpressionStatement {
	node, err := NewExpressionStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExpressionStatementTyped(expression Expression, diagnostics ...Diagnostic) (*ExpressionStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
	}
	created, err := NewExpressionStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExpressionStatementTyped(expression Expression, diagnostics ...Diagnostic) *ExpressionStatement {
	node, err := NewExpressionStatementTyped(expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExpressionStatement) CloneTyped(includeDiagnostics bool) *ExpressionStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExpressionStatement{SyntaxNode: cloned}
}

func (n *ExpressionStatement) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type MaterializeExpression struct {
	*SyntaxNode
}

func (*MaterializeExpression) isExpression() {}

func NewMaterializeExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*MaterializeExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("MaterializeExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &MaterializeExpression{SyntaxNode: node}, nil
}

func MustNewMaterializeExpression(children []SyntaxElement, diagnostics ...Diagnostic) *MaterializeExpression {
	node, err := NewMaterializeExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewMaterializeExpressionTyped(materializeKeyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*MaterializeExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(materializeKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewMaterializeExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewMaterializeExpressionTyped(materializeKeyword *SyntaxToken, openParen *SyntaxToken, expression Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *MaterializeExpression {
	node, err := NewMaterializeExpressionTyped(materializeKeyword, openParen, expression, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *MaterializeExpression) CloneTyped(includeDiagnostics bool) *MaterializeExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &MaterializeExpression{SyntaxNode: cloned}
}

func (n *MaterializeExpression) MaterializeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializeExpression) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *MaterializeExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *MaterializeExpression) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type SetOptionStatement struct {
	*SyntaxNode
}

func (*SetOptionStatement) isStatement() {}

func NewSetOptionStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*SetOptionStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("SetOptionStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &SetOptionStatement{SyntaxNode: node}, nil
}

func MustNewSetOptionStatement(children []SyntaxElement, diagnostics ...Diagnostic) *SetOptionStatement {
	node, err := NewSetOptionStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewSetOptionStatementTyped(setKeyword *SyntaxToken, name *NameDeclaration, valueClause *OptionValueClause, diagnostics ...Diagnostic) (*SetOptionStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(setKeyword),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(valueClause),
	}
	created, err := NewSetOptionStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewSetOptionStatementTyped(setKeyword *SyntaxToken, name *NameDeclaration, valueClause *OptionValueClause, diagnostics ...Diagnostic) *SetOptionStatement {
	node, err := NewSetOptionStatementTyped(setKeyword, name, valueClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *SetOptionStatement) CloneTyped(includeDiagnostics bool) *SetOptionStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &SetOptionStatement{SyntaxNode: cloned}
}

func (n *SetOptionStatement) SetKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *SetOptionStatement) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *SetOptionStatement) ValueClause() *OptionValueClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*OptionValueClause)
	return child
}

type OptionValueClause struct {
	*SyntaxNode
}

func (*OptionValueClause) isClause() {}

func NewOptionValueClause(children []SyntaxElement, diagnostics ...Diagnostic) (*OptionValueClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("OptionValueClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &OptionValueClause{SyntaxNode: node}, nil
}

func MustNewOptionValueClause(children []SyntaxElement, diagnostics ...Diagnostic) *OptionValueClause {
	node, err := NewOptionValueClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewOptionValueClauseTyped(equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*OptionValueClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(expression),
	}
	created, err := NewOptionValueClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewOptionValueClauseTyped(equal *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *OptionValueClause {
	node, err := NewOptionValueClauseTyped(equal, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *OptionValueClause) CloneTyped(includeDiagnostics bool) *OptionValueClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &OptionValueClause{SyntaxNode: cloned}
}

func (n *OptionValueClause) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *OptionValueClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type QueryParametersStatement struct {
	*SyntaxNode
}

func (*QueryParametersStatement) isStatement() {}

func NewQueryParametersStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*QueryParametersStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("QueryParametersStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &QueryParametersStatement{SyntaxNode: node}, nil
}

func MustNewQueryParametersStatement(children []SyntaxElement, diagnostics ...Diagnostic) *QueryParametersStatement {
	node, err := NewQueryParametersStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewQueryParametersStatementTyped(declareKeyword *SyntaxToken, queryParametersKeyword *SyntaxToken, openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*QueryParametersStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(declareKeyword),
		syntaxElementFromTyped(queryParametersKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(parameters),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewQueryParametersStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewQueryParametersStatementTyped(declareKeyword *SyntaxToken, queryParametersKeyword *SyntaxToken, openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *QueryParametersStatement {
	node, err := NewQueryParametersStatementTyped(declareKeyword, queryParametersKeyword, openParen, parameters, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *QueryParametersStatement) CloneTyped(includeDiagnostics bool) *QueryParametersStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &QueryParametersStatement{SyntaxNode: cloned}
}

func (n *QueryParametersStatement) DeclareKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *QueryParametersStatement) QueryParametersKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *QueryParametersStatement) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *QueryParametersStatement) Parameters() *SeparatedSyntaxListOf[*FunctionParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*FunctionParameter](list)
}

func (n *QueryParametersStatement) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type RestrictStatementWithClause struct {
	*SyntaxNode
}

func NewRestrictStatementWithClause(children []SyntaxElement, diagnostics ...Diagnostic) (*RestrictStatementWithClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("RestrictStatementWithClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RestrictStatementWithClause{SyntaxNode: node}, nil
}

func MustNewRestrictStatementWithClause(children []SyntaxElement, diagnostics ...Diagnostic) *RestrictStatementWithClause {
	node, err := NewRestrictStatementWithClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRestrictStatementWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*RestrictStatementWithClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(properties),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewRestrictStatementWithClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRestrictStatementWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *RestrictStatementWithClause {
	node, err := NewRestrictStatementWithClauseTyped(withKeyword, openParen, properties, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RestrictStatementWithClause) CloneTyped(includeDiagnostics bool) *RestrictStatementWithClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RestrictStatementWithClause{SyntaxNode: cloned}
}

func (n *RestrictStatementWithClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatementWithClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatementWithClause) Properties() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

func (n *RestrictStatementWithClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type RestrictStatement struct {
	*SyntaxNode
}

func (*RestrictStatement) isStatement() {}

func NewRestrictStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*RestrictStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("RestrictStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RestrictStatement{SyntaxNode: node}, nil
}

func MustNewRestrictStatement(children []SyntaxElement, diagnostics ...Diagnostic) *RestrictStatement {
	node, err := NewRestrictStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRestrictStatementTyped(restrictKeyword *SyntaxToken, accessKeyword *SyntaxToken, toKeyword *SyntaxToken, openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, withClause *RestrictStatementWithClause, diagnostics ...Diagnostic) (*RestrictStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(restrictKeyword),
		syntaxElementFromTyped(accessKeyword),
		syntaxElementFromTyped(toKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
		syntaxElementFromTyped(closeParen),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewRestrictStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRestrictStatementTyped(restrictKeyword *SyntaxToken, accessKeyword *SyntaxToken, toKeyword *SyntaxToken, openParen *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, withClause *RestrictStatementWithClause, diagnostics ...Diagnostic) *RestrictStatement {
	node, err := NewRestrictStatementTyped(restrictKeyword, accessKeyword, toKeyword, openParen, expressions, closeParen, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RestrictStatement) CloneTyped(includeDiagnostics bool) *RestrictStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RestrictStatement{SyntaxNode: cloned}
}

func (n *RestrictStatement) RestrictKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatement) AccessKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatement) ToKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatement) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatement) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *RestrictStatement) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RestrictStatement) WithClause() *RestrictStatementWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RestrictStatementWithClause)
	return child
}

type PatternStatement struct {
	*SyntaxNode
}

func (*PatternStatement) isStatement() {}

func NewPatternStatement(children []SyntaxElement, diagnostics ...Diagnostic) (*PatternStatement, error) {
	node, err := NewGeneratedSyntaxNodeByName("PatternStatement", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PatternStatement{SyntaxNode: node}, nil
}

func MustNewPatternStatement(children []SyntaxElement, diagnostics ...Diagnostic) *PatternStatement {
	node, err := NewPatternStatement(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPatternStatementTyped(declareKeyword *SyntaxToken, patternKeyword *SyntaxToken, name *NameDeclaration, pattern *PatternDeclaration, diagnostics ...Diagnostic) (*PatternStatement, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(declareKeyword),
		syntaxElementFromTyped(patternKeyword),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(pattern),
	}
	created, err := NewPatternStatement(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPatternStatementTyped(declareKeyword *SyntaxToken, patternKeyword *SyntaxToken, name *NameDeclaration, pattern *PatternDeclaration, diagnostics ...Diagnostic) *PatternStatement {
	node, err := NewPatternStatementTyped(declareKeyword, patternKeyword, name, pattern, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PatternStatement) CloneTyped(includeDiagnostics bool) *PatternStatement {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PatternStatement{SyntaxNode: cloned}
}

func (n *PatternStatement) DeclareKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternStatement) PatternKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternStatement) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *PatternStatement) Pattern() *PatternDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*PatternDeclaration)
	return child
}

type PatternDeclaration struct {
	*SyntaxNode
}

func (*PatternDeclaration) isExpression() {}

func NewPatternDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*PatternDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("PatternDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PatternDeclaration{SyntaxNode: node}, nil
}

func MustNewPatternDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *PatternDeclaration {
	node, err := NewPatternDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPatternDeclarationTyped(equalToken *SyntaxToken, openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, pathParameter *PatternPathParameter, openBrace *SyntaxToken, patterns *SyntaxListOf[*PatternMatch], closeBrace *SyntaxToken, diagnostics ...Diagnostic) (*PatternDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(parameters),
		syntaxElementFromTyped(closeParen),
		syntaxElementFromTyped(pathParameter),
		syntaxElementFromTyped(openBrace),
		syntaxElementFromSyntaxListOf(patterns),
		syntaxElementFromTyped(closeBrace),
	}
	created, err := NewPatternDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPatternDeclarationTyped(equalToken *SyntaxToken, openParen *SyntaxToken, parameters *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, pathParameter *PatternPathParameter, openBrace *SyntaxToken, patterns *SyntaxListOf[*PatternMatch], closeBrace *SyntaxToken, diagnostics ...Diagnostic) *PatternDeclaration {
	node, err := NewPatternDeclarationTyped(equalToken, openParen, parameters, closeParen, pathParameter, openBrace, patterns, closeBrace, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PatternDeclaration) CloneTyped(includeDiagnostics bool) *PatternDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PatternDeclaration{SyntaxNode: cloned}
}

func (n *PatternDeclaration) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternDeclaration) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternDeclaration) Parameters() *SeparatedSyntaxListOf[*NameAndTypeDeclaration] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](list)
}

func (n *PatternDeclaration) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternDeclaration) PathParameter() *PatternPathParameter {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*PatternPathParameter)
	return child
}

func (n *PatternDeclaration) OpenBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternDeclaration) Patterns() *SyntaxListOf[*PatternMatch] {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*PatternMatch](list)
}

func (n *PatternDeclaration) CloseBrace() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PatternPathParameter struct {
	*SyntaxNode
}

func NewPatternPathParameter(children []SyntaxElement, diagnostics ...Diagnostic) (*PatternPathParameter, error) {
	node, err := NewGeneratedSyntaxNodeByName("PatternPathParameter", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PatternPathParameter{SyntaxNode: node}, nil
}

func MustNewPatternPathParameter(children []SyntaxElement, diagnostics ...Diagnostic) *PatternPathParameter {
	node, err := NewPatternPathParameter(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPatternPathParameterTyped(openBracket *SyntaxToken, parameter *NameAndTypeDeclaration, closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*PatternPathParameter, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(parameter),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewPatternPathParameter(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPatternPathParameterTyped(openBracket *SyntaxToken, parameter *NameAndTypeDeclaration, closeBracket *SyntaxToken, diagnostics ...Diagnostic) *PatternPathParameter {
	node, err := NewPatternPathParameterTyped(openBracket, parameter, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PatternPathParameter) CloneTyped(includeDiagnostics bool) *PatternPathParameter {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PatternPathParameter{SyntaxNode: cloned}
}

func (n *PatternPathParameter) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternPathParameter) Parameter() *NameAndTypeDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameAndTypeDeclaration)
	return child
}

func (n *PatternPathParameter) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PatternMatch struct {
	*SyntaxNode
}

func NewPatternMatch(children []SyntaxElement, diagnostics ...Diagnostic) (*PatternMatch, error) {
	node, err := NewGeneratedSyntaxNodeByName("PatternMatch", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PatternMatch{SyntaxNode: node}, nil
}

func MustNewPatternMatch(children []SyntaxElement, diagnostics ...Diagnostic) *PatternMatch {
	node, err := NewPatternMatch(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPatternMatchTyped(parameterValues *ExpressionList, pathValue *PatternPathValue, equalToken *SyntaxToken, body *FunctionBody, semicolonToken *SyntaxToken, diagnostics ...Diagnostic) (*PatternMatch, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(parameterValues),
		syntaxElementFromTyped(pathValue),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(body),
		syntaxElementFromTyped(semicolonToken),
	}
	created, err := NewPatternMatch(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPatternMatchTyped(parameterValues *ExpressionList, pathValue *PatternPathValue, equalToken *SyntaxToken, body *FunctionBody, semicolonToken *SyntaxToken, diagnostics ...Diagnostic) *PatternMatch {
	node, err := NewPatternMatchTyped(parameterValues, pathValue, equalToken, body, semicolonToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PatternMatch) CloneTyped(includeDiagnostics bool) *PatternMatch {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PatternMatch{SyntaxNode: cloned}
}

func (n *PatternMatch) ParameterValues() *ExpressionList {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExpressionList)
	return child
}

func (n *PatternMatch) PathValue() *PatternPathValue {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*PatternPathValue)
	return child
}

func (n *PatternMatch) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternMatch) Body() *FunctionBody {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*FunctionBody)
	return child
}

func (n *PatternMatch) SemicolonToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PatternPathValue struct {
	*SyntaxNode
}

func NewPatternPathValue(children []SyntaxElement, diagnostics ...Diagnostic) (*PatternPathValue, error) {
	node, err := NewGeneratedSyntaxNodeByName("PatternPathValue", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PatternPathValue{SyntaxNode: node}, nil
}

func MustNewPatternPathValue(children []SyntaxElement, diagnostics ...Diagnostic) *PatternPathValue {
	node, err := NewPatternPathValue(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPatternPathValueTyped(dotToken *SyntaxToken, openBracket *SyntaxToken, value Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*PatternPathValue, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dotToken),
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(value),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewPatternPathValue(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPatternPathValueTyped(dotToken *SyntaxToken, openBracket *SyntaxToken, value Expression, closeBracket *SyntaxToken, diagnostics ...Diagnostic) *PatternPathValue {
	node, err := NewPatternPathValueTyped(dotToken, openBracket, value, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PatternPathValue) CloneTyped(includeDiagnostics bool) *PatternPathValue {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PatternPathValue{SyntaxNode: cloned}
}

func (n *PatternPathValue) DotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternPathValue) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PatternPathValue) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PatternPathValue) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type DataScopeExpression struct {
	*SyntaxNode
}

func (*DataScopeExpression) isExpression() {}

func NewDataScopeExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*DataScopeExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("DataScopeExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DataScopeExpression{SyntaxNode: node}, nil
}

func MustNewDataScopeExpression(children []SyntaxElement, diagnostics ...Diagnostic) *DataScopeExpression {
	node, err := NewDataScopeExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDataScopeExpressionTyped(expression Expression, dataScopeClause *DataScopeClause, diagnostics ...Diagnostic) (*DataScopeExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(dataScopeClause),
	}
	created, err := NewDataScopeExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDataScopeExpressionTyped(expression Expression, dataScopeClause *DataScopeClause, diagnostics ...Diagnostic) *DataScopeExpression {
	node, err := NewDataScopeExpressionTyped(expression, dataScopeClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DataScopeExpression) CloneTyped(includeDiagnostics bool) *DataScopeExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DataScopeExpression{SyntaxNode: cloned}
}

func (n *DataScopeExpression) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *DataScopeExpression) DataScopeClause() *DataScopeClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*DataScopeClause)
	return child
}

type DataTableExpression struct {
	*SyntaxNode
}

func (*DataTableExpression) isExpression() {}

func NewDataTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*DataTableExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("DataTableExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DataTableExpression{SyntaxNode: node}, nil
}

func MustNewDataTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) *DataTableExpression {
	node, err := NewDataTableExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDataTableExpressionTyped(dataTableKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, openBracket *SyntaxToken, leadingComma *SyntaxToken, values *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) (*DataTableExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dataTableKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(schema),
		syntaxElementFromTyped(openBracket),
		syntaxElementFromTyped(leadingComma),
		syntaxElementFromSeparatedSyntaxListOf(values),
		syntaxElementFromTyped(closeBracket),
	}
	created, err := NewDataTableExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDataTableExpressionTyped(dataTableKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, openBracket *SyntaxToken, leadingComma *SyntaxToken, values *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, diagnostics ...Diagnostic) *DataTableExpression {
	node, err := NewDataTableExpressionTyped(dataTableKeyword, parameters, schema, openBracket, leadingComma, values, closeBracket, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DataTableExpression) CloneTyped(includeDiagnostics bool) *DataTableExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DataTableExpression{SyntaxNode: cloned}
}

func (n *DataTableExpression) DataTableKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DataTableExpression) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *DataTableExpression) Schema() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

func (n *DataTableExpression) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DataTableExpression) LeadingComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DataTableExpression) Values() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *DataTableExpression) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type RowSchema struct {
	*SyntaxNode
}

func NewRowSchema(children []SyntaxElement, diagnostics ...Diagnostic) (*RowSchema, error) {
	node, err := NewGeneratedSyntaxNodeByName("RowSchema", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &RowSchema{SyntaxNode: node}, nil
}

func MustNewRowSchema(children []SyntaxElement, diagnostics ...Diagnostic) *RowSchema {
	node, err := NewRowSchema(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewRowSchemaTyped(openParen *SyntaxToken, leadingComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*RowSchema, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(leadingComma),
		syntaxElementFromSeparatedSyntaxListOf(columns),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewRowSchema(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewRowSchemaTyped(openParen *SyntaxToken, leadingComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) *RowSchema {
	node, err := NewRowSchemaTyped(openParen, leadingComma, columns, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *RowSchema) CloneTyped(includeDiagnostics bool) *RowSchema {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &RowSchema{SyntaxNode: cloned}
}

func (n *RowSchema) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RowSchema) LeadingComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *RowSchema) Columns() *SeparatedSyntaxListOf[*NameAndTypeDeclaration] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](list)
}

func (n *RowSchema) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type EvaluateRowSchema struct {
	*SyntaxNode
}

func NewEvaluateRowSchema(children []SyntaxElement, diagnostics ...Diagnostic) (*EvaluateRowSchema, error) {
	node, err := NewGeneratedSyntaxNodeByName("EvaluateRowSchema", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &EvaluateRowSchema{SyntaxNode: node}, nil
}

func MustNewEvaluateRowSchema(children []SyntaxElement, diagnostics ...Diagnostic) *EvaluateRowSchema {
	node, err := NewEvaluateRowSchema(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewEvaluateRowSchemaTyped(openParen *SyntaxToken, leadingComma *SyntaxToken, asteriskToken *SyntaxToken, asteriskTokenComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*EvaluateRowSchema, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(leadingComma),
		syntaxElementFromTyped(asteriskToken),
		syntaxElementFromTyped(asteriskTokenComma),
		syntaxElementFromSeparatedSyntaxListOf(columns),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewEvaluateRowSchema(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewEvaluateRowSchemaTyped(openParen *SyntaxToken, leadingComma *SyntaxToken, asteriskToken *SyntaxToken, asteriskTokenComma *SyntaxToken, columns *SeparatedSyntaxListOf[*NameAndTypeDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) *EvaluateRowSchema {
	node, err := NewEvaluateRowSchemaTyped(openParen, leadingComma, asteriskToken, asteriskTokenComma, columns, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *EvaluateRowSchema) CloneTyped(includeDiagnostics bool) *EvaluateRowSchema {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &EvaluateRowSchema{SyntaxNode: cloned}
}

func (n *EvaluateRowSchema) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateRowSchema) LeadingComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateRowSchema) AsteriskToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateRowSchema) AsteriskTokenComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *EvaluateRowSchema) Columns() *SeparatedSyntaxListOf[*NameAndTypeDeclaration] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](list)
}

func (n *EvaluateRowSchema) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ExternalDataExpression struct {
	*SyntaxNode
}

func (*ExternalDataExpression) isExpression() {}

func NewExternalDataExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ExternalDataExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExternalDataExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExternalDataExpression{SyntaxNode: node}, nil
}

func MustNewExternalDataExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ExternalDataExpression {
	node, err := NewExternalDataExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExternalDataExpressionTyped(externalDataKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, openBracket *SyntaxToken, uRIs *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, withClause *ExternalDataWithClause, diagnostics ...Diagnostic) (*ExternalDataExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(externalDataKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(schema),
		syntaxElementFromTyped(openBracket),
		syntaxElementFromSeparatedSyntaxListOf(uRIs),
		syntaxElementFromTyped(closeBracket),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewExternalDataExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExternalDataExpressionTyped(externalDataKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, openBracket *SyntaxToken, uRIs *SeparatedSyntaxListOf[Expression], closeBracket *SyntaxToken, withClause *ExternalDataWithClause, diagnostics ...Diagnostic) *ExternalDataExpression {
	node, err := NewExternalDataExpressionTyped(externalDataKeyword, parameters, schema, openBracket, uRIs, closeBracket, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExternalDataExpression) CloneTyped(includeDiagnostics bool) *ExternalDataExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExternalDataExpression{SyntaxNode: cloned}
}

func (n *ExternalDataExpression) ExternalDataKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExternalDataExpression) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *ExternalDataExpression) Schema() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

func (n *ExternalDataExpression) OpenBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExternalDataExpression) URIs() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *ExternalDataExpression) CloseBracket() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExternalDataExpression) WithClause() *ExternalDataWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExternalDataWithClause)
	return child
}

type ContextualDataTableExpression struct {
	*SyntaxNode
}

func (*ContextualDataTableExpression) isExpression() {}

func NewContextualDataTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*ContextualDataTableExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("ContextualDataTableExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ContextualDataTableExpression{SyntaxNode: node}, nil
}

func MustNewContextualDataTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) *ContextualDataTableExpression {
	node, err := NewContextualDataTableExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewContextualDataTableExpressionTyped(contextualDataTableKeyword *SyntaxToken, id Expression, schema *RowSchema, diagnostics ...Diagnostic) (*ContextualDataTableExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(contextualDataTableKeyword),
		syntaxElementFromTyped(id),
		syntaxElementFromTyped(schema),
	}
	created, err := NewContextualDataTableExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewContextualDataTableExpressionTyped(contextualDataTableKeyword *SyntaxToken, id Expression, schema *RowSchema, diagnostics ...Diagnostic) *ContextualDataTableExpression {
	node, err := NewContextualDataTableExpressionTyped(contextualDataTableKeyword, id, schema, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ContextualDataTableExpression) CloneTyped(includeDiagnostics bool) *ContextualDataTableExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ContextualDataTableExpression{SyntaxNode: cloned}
}

func (n *ContextualDataTableExpression) ContextualDataTableKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ContextualDataTableExpression) Id() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ContextualDataTableExpression) Schema() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

type ExternalDataWithClause struct {
	*SyntaxNode
}

func (*ExternalDataWithClause) isExpression() {}

func NewExternalDataWithClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ExternalDataWithClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ExternalDataWithClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ExternalDataWithClause{SyntaxNode: node}, nil
}

func MustNewExternalDataWithClause(children []SyntaxElement, diagnostics ...Diagnostic) *ExternalDataWithClause {
	node, err := NewExternalDataWithClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewExternalDataWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ExternalDataWithClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(properties),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewExternalDataWithClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewExternalDataWithClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *ExternalDataWithClause {
	node, err := NewExternalDataWithClauseTyped(withKeyword, openParen, properties, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ExternalDataWithClause) CloneTyped(includeDiagnostics bool) *ExternalDataWithClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ExternalDataWithClause{SyntaxNode: cloned}
}

func (n *ExternalDataWithClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExternalDataWithClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ExternalDataWithClause) Properties() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

func (n *ExternalDataWithClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type InlineExternalTableKindClause struct {
	*SyntaxNode
}

func (*InlineExternalTableKindClause) isClause() {}

func NewInlineExternalTableKindClause(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTableKindClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTableKindClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTableKindClause{SyntaxNode: node}, nil
}

func MustNewInlineExternalTableKindClause(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTableKindClause {
	node, err := NewInlineExternalTableKindClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTableKindClauseTyped(kindKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) (*InlineExternalTableKindClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(kindKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(value),
	}
	created, err := NewInlineExternalTableKindClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTableKindClauseTyped(kindKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) *InlineExternalTableKindClause {
	node, err := NewInlineExternalTableKindClauseTyped(kindKeyword, equalToken, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTableKindClause) CloneTyped(includeDiagnostics bool) *InlineExternalTableKindClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTableKindClause{SyntaxNode: cloned}
}

func (n *InlineExternalTableKindClause) KindKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableKindClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableKindClause) Value() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type InlineExternalTableDataFormatClause struct {
	*SyntaxNode
}

func (*InlineExternalTableDataFormatClause) isClause() {}

func NewInlineExternalTableDataFormatClause(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTableDataFormatClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTableDataFormatClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTableDataFormatClause{SyntaxNode: node}, nil
}

func MustNewInlineExternalTableDataFormatClause(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTableDataFormatClause {
	node, err := NewInlineExternalTableDataFormatClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTableDataFormatClauseTyped(dataFormatKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) (*InlineExternalTableDataFormatClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dataFormatKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(value),
	}
	created, err := NewInlineExternalTableDataFormatClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTableDataFormatClauseTyped(dataFormatKeyword *SyntaxToken, equalToken *SyntaxToken, value *SyntaxToken, diagnostics ...Diagnostic) *InlineExternalTableDataFormatClause {
	node, err := NewInlineExternalTableDataFormatClauseTyped(dataFormatKeyword, equalToken, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTableDataFormatClause) CloneTyped(includeDiagnostics bool) *InlineExternalTableDataFormatClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTableDataFormatClause{SyntaxNode: cloned}
}

func (n *InlineExternalTableDataFormatClause) DataFormatKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableDataFormatClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableDataFormatClause) Value() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type DateTimePattern struct {
	*SyntaxNode
}

func (*DateTimePattern) isExpression() {}

func NewDateTimePattern(children []SyntaxElement, diagnostics ...Diagnostic) (*DateTimePattern, error) {
	node, err := NewGeneratedSyntaxNodeByName("DateTimePattern", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &DateTimePattern{SyntaxNode: node}, nil
}

func MustNewDateTimePattern(children []SyntaxElement, diagnostics ...Diagnostic) *DateTimePattern {
	node, err := NewDateTimePattern(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewDateTimePatternTyped(dateTimePatternKeyword *SyntaxToken, openParen *SyntaxToken, stringLiteral *LiteralExpression, comma *SyntaxToken, partitionColumn *NameReference, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*DateTimePattern, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dateTimePatternKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(stringLiteral),
		syntaxElementFromTyped(comma),
		syntaxElementFromTyped(partitionColumn),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewDateTimePattern(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewDateTimePatternTyped(dateTimePatternKeyword *SyntaxToken, openParen *SyntaxToken, stringLiteral *LiteralExpression, comma *SyntaxToken, partitionColumn *NameReference, closeParen *SyntaxToken, diagnostics ...Diagnostic) *DateTimePattern {
	node, err := NewDateTimePatternTyped(dateTimePatternKeyword, openParen, stringLiteral, comma, partitionColumn, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *DateTimePattern) CloneTyped(includeDiagnostics bool) *DateTimePattern {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &DateTimePattern{SyntaxNode: cloned}
}

func (n *DateTimePattern) DateTimePatternKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DateTimePattern) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DateTimePattern) StringLiteral() *LiteralExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*LiteralExpression)
	return child
}

func (n *DateTimePattern) Comma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *DateTimePattern) PartitionColumn() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *DateTimePattern) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type InlineExternalTablePathFormatPartitionColumnReference struct {
	*SyntaxNode
}

func (*InlineExternalTablePathFormatPartitionColumnReference) isExpression() {}

func NewInlineExternalTablePathFormatPartitionColumnReference(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTablePathFormatPartitionColumnReference, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTablePathFormatPartitionColumnReference", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node}, nil
}

func MustNewInlineExternalTablePathFormatPartitionColumnReference(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTablePathFormatPartitionColumnReference {
	node, err := NewInlineExternalTablePathFormatPartitionColumnReference(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTablePathFormatPartitionColumnReferenceTyped(partitionColumnExpression Expression, separatorLiteral *LiteralExpression, diagnostics ...Diagnostic) (*InlineExternalTablePathFormatPartitionColumnReference, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionColumnExpression),
		syntaxElementFromTyped(separatorLiteral),
	}
	created, err := NewInlineExternalTablePathFormatPartitionColumnReference(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTablePathFormatPartitionColumnReferenceTyped(partitionColumnExpression Expression, separatorLiteral *LiteralExpression, diagnostics ...Diagnostic) *InlineExternalTablePathFormatPartitionColumnReference {
	node, err := NewInlineExternalTablePathFormatPartitionColumnReferenceTyped(partitionColumnExpression, separatorLiteral, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTablePathFormatPartitionColumnReference) CloneTyped(includeDiagnostics bool) *InlineExternalTablePathFormatPartitionColumnReference {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: cloned}
}

func (n *InlineExternalTablePathFormatPartitionColumnReference) PartitionColumnExpression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *InlineExternalTablePathFormatPartitionColumnReference) SeparatorLiteral() *LiteralExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*LiteralExpression)
	return child
}

type InlineExternalTablePathFormatClause struct {
	*SyntaxNode
}

func (*InlineExternalTablePathFormatClause) isClause() {}

func NewInlineExternalTablePathFormatClause(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTablePathFormatClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTablePathFormatClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTablePathFormatClause{SyntaxNode: node}, nil
}

func MustNewInlineExternalTablePathFormatClause(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTablePathFormatClause {
	node, err := NewInlineExternalTablePathFormatClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTablePathFormatClauseTyped(pathFormatKeyword *SyntaxToken, equalToken *SyntaxToken, openParen *SyntaxToken, optionalSeparatorLiteral *LiteralExpression, pathExpressions *SyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*InlineExternalTablePathFormatClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(pathFormatKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(optionalSeparatorLiteral),
		syntaxElementFromSyntaxListOf(pathExpressions),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewInlineExternalTablePathFormatClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTablePathFormatClauseTyped(pathFormatKeyword *SyntaxToken, equalToken *SyntaxToken, openParen *SyntaxToken, optionalSeparatorLiteral *LiteralExpression, pathExpressions *SyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference], closeParen *SyntaxToken, diagnostics ...Diagnostic) *InlineExternalTablePathFormatClause {
	node, err := NewInlineExternalTablePathFormatClauseTyped(pathFormatKeyword, equalToken, openParen, optionalSeparatorLiteral, pathExpressions, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTablePathFormatClause) CloneTyped(includeDiagnostics bool) *InlineExternalTablePathFormatClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTablePathFormatClause{SyntaxNode: cloned}
}

func (n *InlineExternalTablePathFormatClause) PathFormatKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePathFormatClause) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePathFormatClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePathFormatClause) OptionalSeparatorLiteral() *LiteralExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*LiteralExpression)
	return child
}

func (n *InlineExternalTablePathFormatClause) PathExpressions() *SyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*InlineExternalTablePathFormatPartitionColumnReference](list)
}

func (n *InlineExternalTablePathFormatClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PartitionColumnDeclaration struct {
	*SyntaxNode
}

func (*PartitionColumnDeclaration) isExpression() {}

func NewPartitionColumnDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionColumnDeclaration, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionColumnDeclaration", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionColumnDeclaration{SyntaxNode: node}, nil
}

func MustNewPartitionColumnDeclaration(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionColumnDeclaration {
	node, err := NewPartitionColumnDeclaration(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionColumnDeclarationTyped(name *NameDeclaration, colon *SyntaxToken, typeChild TypeExpression, equal *SyntaxToken, expr Expression, diagnostics ...Diagnostic) (*PartitionColumnDeclaration, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(colon),
		syntaxElementFromTyped(typeChild),
		syntaxElementFromTyped(equal),
		syntaxElementFromTyped(expr),
	}
	created, err := NewPartitionColumnDeclaration(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionColumnDeclarationTyped(name *NameDeclaration, colon *SyntaxToken, typeChild TypeExpression, equal *SyntaxToken, expr Expression, diagnostics ...Diagnostic) *PartitionColumnDeclaration {
	node, err := NewPartitionColumnDeclarationTyped(name, colon, typeChild, equal, expr, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionColumnDeclaration) CloneTyped(includeDiagnostics bool) *PartitionColumnDeclaration {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionColumnDeclaration{SyntaxNode: cloned}
}

func (n *PartitionColumnDeclaration) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *PartitionColumnDeclaration) Colon() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionColumnDeclaration) Type() TypeExpression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(TypeExpression)
	return child
}

func (n *PartitionColumnDeclaration) Equal() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionColumnDeclaration) Expr() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type InlineExternalTablePartitionClause struct {
	*SyntaxNode
}

func NewInlineExternalTablePartitionClause(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTablePartitionClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTablePartitionClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTablePartitionClause{SyntaxNode: node}, nil
}

func MustNewInlineExternalTablePartitionClause(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTablePartitionClause {
	node, err := NewInlineExternalTablePartitionClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTablePartitionClauseTyped(partitionKeyword *SyntaxToken, byKeyword *SyntaxToken, openParen *SyntaxToken, leadingComma *SyntaxToken, partitionColumns *SeparatedSyntaxListOf[*PartitionColumnDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*InlineExternalTablePartitionClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionKeyword),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(leadingComma),
		syntaxElementFromSeparatedSyntaxListOf(partitionColumns),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewInlineExternalTablePartitionClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTablePartitionClauseTyped(partitionKeyword *SyntaxToken, byKeyword *SyntaxToken, openParen *SyntaxToken, leadingComma *SyntaxToken, partitionColumns *SeparatedSyntaxListOf[*PartitionColumnDeclaration], closeParen *SyntaxToken, diagnostics ...Diagnostic) *InlineExternalTablePartitionClause {
	node, err := NewInlineExternalTablePartitionClauseTyped(partitionKeyword, byKeyword, openParen, leadingComma, partitionColumns, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTablePartitionClause) CloneTyped(includeDiagnostics bool) *InlineExternalTablePartitionClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTablePartitionClause{SyntaxNode: cloned}
}

func (n *InlineExternalTablePartitionClause) PartitionKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePartitionClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePartitionClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePartitionClause) LeadingComma() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTablePartitionClause) PartitionColumns() *SeparatedSyntaxListOf[*PartitionColumnDeclaration] {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*PartitionColumnDeclaration](list)
}

func (n *InlineExternalTablePartitionClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type InlineExternalTableConnectionStringsClause struct {
	*SyntaxNode
}

func NewInlineExternalTableConnectionStringsClause(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTableConnectionStringsClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTableConnectionStringsClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTableConnectionStringsClause{SyntaxNode: node}, nil
}

func MustNewInlineExternalTableConnectionStringsClause(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTableConnectionStringsClause {
	node, err := NewInlineExternalTableConnectionStringsClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTableConnectionStringsClauseTyped(openParen *SyntaxToken, connectionStrings *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*InlineExternalTableConnectionStringsClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(connectionStrings),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewInlineExternalTableConnectionStringsClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTableConnectionStringsClauseTyped(openParen *SyntaxToken, connectionStrings *SeparatedSyntaxListOf[Expression], closeParen *SyntaxToken, diagnostics ...Diagnostic) *InlineExternalTableConnectionStringsClause {
	node, err := NewInlineExternalTableConnectionStringsClauseTyped(openParen, connectionStrings, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTableConnectionStringsClause) CloneTyped(includeDiagnostics bool) *InlineExternalTableConnectionStringsClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTableConnectionStringsClause{SyntaxNode: cloned}
}

func (n *InlineExternalTableConnectionStringsClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableConnectionStringsClause) ConnectionStrings() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

func (n *InlineExternalTableConnectionStringsClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type InlineExternalTableExpression struct {
	*SyntaxNode
}

func (*InlineExternalTableExpression) isQueryOperator() {}

func (*InlineExternalTableExpression) isExpression() {}

func NewInlineExternalTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) (*InlineExternalTableExpression, error) {
	node, err := NewGeneratedSyntaxNodeByName("InlineExternalTableExpression", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &InlineExternalTableExpression{SyntaxNode: node}, nil
}

func MustNewInlineExternalTableExpression(children []SyntaxElement, diagnostics ...Diagnostic) *InlineExternalTableExpression {
	node, err := NewInlineExternalTableExpression(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewInlineExternalTableExpressionTyped(inlineExternalTableKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, kindParameter *InlineExternalTableKindClause, partitionClause *InlineExternalTablePartitionClause, pathFormat *InlineExternalTablePathFormatClause, dataFormatParameter *InlineExternalTableDataFormatClause, connectionStrings *InlineExternalTableConnectionStringsClause, withClause *ExternalDataWithClause, diagnostics ...Diagnostic) (*InlineExternalTableExpression, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(inlineExternalTableKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(schema),
		syntaxElementFromTyped(kindParameter),
		syntaxElementFromTyped(partitionClause),
		syntaxElementFromTyped(pathFormat),
		syntaxElementFromTyped(dataFormatParameter),
		syntaxElementFromTyped(connectionStrings),
		syntaxElementFromTyped(withClause),
	}
	created, err := NewInlineExternalTableExpression(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewInlineExternalTableExpressionTyped(inlineExternalTableKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], schema *RowSchema, kindParameter *InlineExternalTableKindClause, partitionClause *InlineExternalTablePartitionClause, pathFormat *InlineExternalTablePathFormatClause, dataFormatParameter *InlineExternalTableDataFormatClause, connectionStrings *InlineExternalTableConnectionStringsClause, withClause *ExternalDataWithClause, diagnostics ...Diagnostic) *InlineExternalTableExpression {
	node, err := NewInlineExternalTableExpressionTyped(inlineExternalTableKeyword, parameters, schema, kindParameter, partitionClause, pathFormat, dataFormatParameter, connectionStrings, withClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *InlineExternalTableExpression) CloneTyped(includeDiagnostics bool) *InlineExternalTableExpression {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &InlineExternalTableExpression{SyntaxNode: cloned}
}

func (n *InlineExternalTableExpression) InlineExternalTableKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *InlineExternalTableExpression) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *InlineExternalTableExpression) Schema() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

func (n *InlineExternalTableExpression) KindParameter() *InlineExternalTableKindClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*InlineExternalTableKindClause)
	return child
}

func (n *InlineExternalTableExpression) PartitionClause() *InlineExternalTablePartitionClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*InlineExternalTablePartitionClause)
	return child
}

func (n *InlineExternalTableExpression) PathFormat() *InlineExternalTablePathFormatClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*InlineExternalTablePathFormatClause)
	return child
}

func (n *InlineExternalTableExpression) DataFormatParameter() *InlineExternalTableDataFormatClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*InlineExternalTableDataFormatClause)
	return child
}

func (n *InlineExternalTableExpression) ConnectionStrings() *InlineExternalTableConnectionStringsClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*InlineExternalTableConnectionStringsClause)
	return child
}

func (n *InlineExternalTableExpression) WithClause() *ExternalDataWithClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(8)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ExternalDataWithClause)
	return child
}

type JoinOperator struct {
	*SyntaxNode
}

func (*JoinOperator) isQueryOperator() {}

func (*JoinOperator) isExpression() {}

func NewJoinOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*JoinOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("JoinOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JoinOperator{SyntaxNode: node}, nil
}

func MustNewJoinOperator(children []SyntaxElement, diagnostics ...Diagnostic) *JoinOperator {
	node, err := NewJoinOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJoinOperatorTyped(joinKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, conditionClause JoinConditionClause, diagnostics ...Diagnostic) (*JoinOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(joinKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(conditionClause),
	}
	created, err := NewJoinOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJoinOperatorTyped(joinKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, conditionClause JoinConditionClause, diagnostics ...Diagnostic) *JoinOperator {
	node, err := NewJoinOperatorTyped(joinKeyword, parameters, expression, conditionClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JoinOperator) CloneTyped(includeDiagnostics bool) *JoinOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JoinOperator{SyntaxNode: cloned}
}

func (n *JoinOperator) JoinKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JoinOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *JoinOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *JoinOperator) ConditionClause() JoinConditionClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(JoinConditionClause)
	return child
}

type LookupOperator struct {
	*SyntaxNode
}

func (*LookupOperator) isQueryOperator() {}

func (*LookupOperator) isExpression() {}

func NewLookupOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*LookupOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("LookupOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &LookupOperator{SyntaxNode: node}, nil
}

func MustNewLookupOperator(children []SyntaxElement, diagnostics ...Diagnostic) *LookupOperator {
	node, err := NewLookupOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewLookupOperatorTyped(lookupKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, lookupClause JoinConditionClause, diagnostics ...Diagnostic) (*LookupOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(lookupKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(expression),
		syntaxElementFromTyped(lookupClause),
	}
	created, err := NewLookupOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewLookupOperatorTyped(lookupKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], expression Expression, lookupClause JoinConditionClause, diagnostics ...Diagnostic) *LookupOperator {
	node, err := NewLookupOperatorTyped(lookupKeyword, parameters, expression, lookupClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *LookupOperator) CloneTyped(includeDiagnostics bool) *LookupOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &LookupOperator{SyntaxNode: cloned}
}

func (n *LookupOperator) LookupKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *LookupOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *LookupOperator) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *LookupOperator) LookupClause() JoinConditionClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(JoinConditionClause)
	return child
}

type JoinOnClause struct {
	*SyntaxNode
}

func (*JoinOnClause) isJoinConditionClause() {}

func (*JoinOnClause) isClause() {}

func NewJoinOnClause(children []SyntaxElement, diagnostics ...Diagnostic) (*JoinOnClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("JoinOnClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JoinOnClause{SyntaxNode: node}, nil
}

func MustNewJoinOnClause(children []SyntaxElement, diagnostics ...Diagnostic) *JoinOnClause {
	node, err := NewJoinOnClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJoinOnClauseTyped(onKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*JoinOnClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(onKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewJoinOnClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJoinOnClauseTyped(onKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *JoinOnClause {
	node, err := NewJoinOnClauseTyped(onKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JoinOnClause) CloneTyped(includeDiagnostics bool) *JoinOnClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JoinOnClause{SyntaxNode: cloned}
}

func (n *JoinOnClause) OnKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JoinOnClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type JoinWhereClause struct {
	*SyntaxNode
}

func (*JoinWhereClause) isJoinConditionClause() {}

func (*JoinWhereClause) isClause() {}

func NewJoinWhereClause(children []SyntaxElement, diagnostics ...Diagnostic) (*JoinWhereClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("JoinWhereClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &JoinWhereClause{SyntaxNode: node}, nil
}

func MustNewJoinWhereClause(children []SyntaxElement, diagnostics ...Diagnostic) *JoinWhereClause {
	node, err := NewJoinWhereClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewJoinWhereClauseTyped(whereKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*JoinWhereClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(whereKeyword),
		syntaxElementFromTyped(expression),
	}
	created, err := NewJoinWhereClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewJoinWhereClauseTyped(whereKeyword *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *JoinWhereClause {
	node, err := NewJoinWhereClauseTyped(whereKeyword, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *JoinWhereClause) CloneTyped(includeDiagnostics bool) *JoinWhereClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &JoinWhereClause{SyntaxNode: cloned}
}

func (n *JoinWhereClause) WhereKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *JoinWhereClause) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type ScanOperator struct {
	*SyntaxNode
}

func (*ScanOperator) isQueryOperator() {}

func (*ScanOperator) isExpression() {}

func NewScanOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanOperator{SyntaxNode: node}, nil
}

func MustNewScanOperator(children []SyntaxElement, diagnostics ...Diagnostic) *ScanOperator {
	node, err := NewScanOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanOperatorTyped(scanKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], orderByClause *ScanOrderByClause, partitionByClause *ScanPartitionByClause, declareClause *ScanDeclareClause, withKeyword *SyntaxToken, openParenToken *SyntaxToken, steps *SyntaxListOf[*ScanStep], closeParenToken *SyntaxToken, diagnostics ...Diagnostic) (*ScanOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(scanKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(orderByClause),
		syntaxElementFromTyped(partitionByClause),
		syntaxElementFromTyped(declareClause),
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParenToken),
		syntaxElementFromSyntaxListOf(steps),
		syntaxElementFromTyped(closeParenToken),
	}
	created, err := NewScanOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanOperatorTyped(scanKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], orderByClause *ScanOrderByClause, partitionByClause *ScanPartitionByClause, declareClause *ScanDeclareClause, withKeyword *SyntaxToken, openParenToken *SyntaxToken, steps *SyntaxListOf[*ScanStep], closeParenToken *SyntaxToken, diagnostics ...Diagnostic) *ScanOperator {
	node, err := NewScanOperatorTyped(scanKeyword, parameters, orderByClause, partitionByClause, declareClause, withKeyword, openParenToken, steps, closeParenToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanOperator) CloneTyped(includeDiagnostics bool) *ScanOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanOperator{SyntaxNode: cloned}
}

func (n *ScanOperator) ScanKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *ScanOperator) OrderByClause() *ScanOrderByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ScanOrderByClause)
	return child
}

func (n *ScanOperator) PartitionByClause() *ScanPartitionByClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ScanPartitionByClause)
	return child
}

func (n *ScanOperator) DeclareClause() *ScanDeclareClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ScanDeclareClause)
	return child
}

func (n *ScanOperator) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanOperator) OpenParenToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanOperator) Steps() *SyntaxListOf[*ScanStep] {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*ScanStep](list)
}

func (n *ScanOperator) CloseParenToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(8)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type AssertSchemaOperator struct {
	*SyntaxNode
}

func (*AssertSchemaOperator) isQueryOperator() {}

func (*AssertSchemaOperator) isExpression() {}

func NewAssertSchemaOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*AssertSchemaOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("AssertSchemaOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &AssertSchemaOperator{SyntaxNode: node}, nil
}

func MustNewAssertSchemaOperator(children []SyntaxElement, diagnostics ...Diagnostic) *AssertSchemaOperator {
	node, err := NewAssertSchemaOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewAssertSchemaOperatorTyped(assertSchemaKeyword *SyntaxToken, schema *RowSchema, diagnostics ...Diagnostic) (*AssertSchemaOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(assertSchemaKeyword),
		syntaxElementFromTyped(schema),
	}
	created, err := NewAssertSchemaOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewAssertSchemaOperatorTyped(assertSchemaKeyword *SyntaxToken, schema *RowSchema, diagnostics ...Diagnostic) *AssertSchemaOperator {
	node, err := NewAssertSchemaOperatorTyped(assertSchemaKeyword, schema, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *AssertSchemaOperator) CloneTyped(includeDiagnostics bool) *AssertSchemaOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &AssertSchemaOperator{SyntaxNode: cloned}
}

func (n *AssertSchemaOperator) AssertSchemaKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *AssertSchemaOperator) Schema() *RowSchema {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*RowSchema)
	return child
}

type ScanDeclareClause struct {
	*SyntaxNode
}

func NewScanDeclareClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanDeclareClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanDeclareClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanDeclareClause{SyntaxNode: node}, nil
}

func MustNewScanDeclareClause(children []SyntaxElement, diagnostics ...Diagnostic) *ScanDeclareClause {
	node, err := NewScanDeclareClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanDeclareClauseTyped(declareKeyword *SyntaxToken, openParen *SyntaxToken, declarations *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*ScanDeclareClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(declareKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(declarations),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewScanDeclareClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanDeclareClauseTyped(declareKeyword *SyntaxToken, openParen *SyntaxToken, declarations *SeparatedSyntaxListOf[*FunctionParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *ScanDeclareClause {
	node, err := NewScanDeclareClauseTyped(declareKeyword, openParen, declarations, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanDeclareClause) CloneTyped(includeDiagnostics bool) *ScanDeclareClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanDeclareClause{SyntaxNode: cloned}
}

func (n *ScanDeclareClause) DeclareKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanDeclareClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanDeclareClause) Declarations() *SeparatedSyntaxListOf[*FunctionParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*FunctionParameter](list)
}

func (n *ScanDeclareClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ScanOrderByClause struct {
	*SyntaxNode
}

func NewScanOrderByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanOrderByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanOrderByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanOrderByClause{SyntaxNode: node}, nil
}

func MustNewScanOrderByClause(children []SyntaxElement, diagnostics ...Diagnostic) *ScanOrderByClause {
	node, err := NewScanOrderByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanOrderByClauseTyped(orderKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ScanOrderByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(orderKeyword),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewScanOrderByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanOrderByClauseTyped(orderKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ScanOrderByClause {
	node, err := NewScanOrderByClauseTyped(orderKeyword, byKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanOrderByClause) CloneTyped(includeDiagnostics bool) *ScanOrderByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanOrderByClause{SyntaxNode: cloned}
}

func (n *ScanOrderByClause) OrderKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanOrderByClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanOrderByClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ScanPartitionByClause struct {
	*SyntaxNode
}

func NewScanPartitionByClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanPartitionByClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanPartitionByClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanPartitionByClause{SyntaxNode: node}, nil
}

func MustNewScanPartitionByClause(children []SyntaxElement, diagnostics ...Diagnostic) *ScanPartitionByClause {
	node, err := NewScanPartitionByClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanPartitionByClauseTyped(partitionKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) (*ScanPartitionByClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionKeyword),
		syntaxElementFromTyped(byKeyword),
		syntaxElementFromSeparatedSyntaxListOf(expressions),
	}
	created, err := NewScanPartitionByClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanPartitionByClauseTyped(partitionKeyword *SyntaxToken, byKeyword *SyntaxToken, expressions *SeparatedSyntaxListOf[Expression], diagnostics ...Diagnostic) *ScanPartitionByClause {
	node, err := NewScanPartitionByClauseTyped(partitionKeyword, byKeyword, expressions, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanPartitionByClause) CloneTyped(includeDiagnostics bool) *ScanPartitionByClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanPartitionByClause{SyntaxNode: cloned}
}

func (n *ScanPartitionByClause) PartitionKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanPartitionByClause) ByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanPartitionByClause) Expressions() *SeparatedSyntaxListOf[Expression] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Expression](list)
}

type ScanStepOutput struct {
	*SyntaxNode
}

func NewScanStepOutput(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanStepOutput, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanStepOutput", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanStepOutput{SyntaxNode: node}, nil
}

func MustNewScanStepOutput(children []SyntaxElement, diagnostics ...Diagnostic) *ScanStepOutput {
	node, err := NewScanStepOutput(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanStepOutputTyped(outputKeyword *SyntaxToken, equalToken *SyntaxToken, outputKind *SyntaxToken, diagnostics ...Diagnostic) (*ScanStepOutput, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(outputKeyword),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(outputKind),
	}
	created, err := NewScanStepOutput(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanStepOutputTyped(outputKeyword *SyntaxToken, equalToken *SyntaxToken, outputKind *SyntaxToken, diagnostics ...Diagnostic) *ScanStepOutput {
	node, err := NewScanStepOutputTyped(outputKeyword, equalToken, outputKind, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanStepOutput) CloneTyped(includeDiagnostics bool) *ScanStepOutput {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanStepOutput{SyntaxNode: cloned}
}

func (n *ScanStepOutput) OutputKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanStepOutput) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanStepOutput) OutputKind() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ScanStep struct {
	*SyntaxNode
}

func NewScanStep(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanStep, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanStep", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanStep{SyntaxNode: node}, nil
}

func MustNewScanStep(children []SyntaxElement, diagnostics ...Diagnostic) *ScanStep {
	node, err := NewScanStep(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanStepTyped(stepKeyword *SyntaxToken, name *NameDeclaration, optionalKeyword *SyntaxToken, scanStepOutput *ScanStepOutput, colonToken *SyntaxToken, condition Expression, computationClause *ScanComputationClause, semicolonToken *SyntaxToken, diagnostics ...Diagnostic) (*ScanStep, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(stepKeyword),
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(optionalKeyword),
		syntaxElementFromTyped(scanStepOutput),
		syntaxElementFromTyped(colonToken),
		syntaxElementFromTyped(condition),
		syntaxElementFromTyped(computationClause),
		syntaxElementFromTyped(semicolonToken),
	}
	created, err := NewScanStep(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanStepTyped(stepKeyword *SyntaxToken, name *NameDeclaration, optionalKeyword *SyntaxToken, scanStepOutput *ScanStepOutput, colonToken *SyntaxToken, condition Expression, computationClause *ScanComputationClause, semicolonToken *SyntaxToken, diagnostics ...Diagnostic) *ScanStep {
	node, err := NewScanStepTyped(stepKeyword, name, optionalKeyword, scanStepOutput, colonToken, condition, computationClause, semicolonToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanStep) CloneTyped(includeDiagnostics bool) *ScanStep {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanStep{SyntaxNode: cloned}
}

func (n *ScanStep) StepKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanStep) Name() *NameDeclaration {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameDeclaration)
	return child
}

func (n *ScanStep) OptionalKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanStep) ScanStepOutput() *ScanStepOutput {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ScanStepOutput)
	return child
}

func (n *ScanStep) ColonToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanStep) Condition() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *ScanStep) ComputationClause() *ScanComputationClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*ScanComputationClause)
	return child
}

func (n *ScanStep) SemicolonToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(7)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type ScanComputationClause struct {
	*SyntaxNode
}

func NewScanComputationClause(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanComputationClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanComputationClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanComputationClause{SyntaxNode: node}, nil
}

func MustNewScanComputationClause(children []SyntaxElement, diagnostics ...Diagnostic) *ScanComputationClause {
	node, err := NewScanComputationClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanComputationClauseTyped(arrowToken *SyntaxToken, assignments *SeparatedSyntaxListOf[*ScanAssignment], diagnostics ...Diagnostic) (*ScanComputationClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(arrowToken),
		syntaxElementFromSeparatedSyntaxListOf(assignments),
	}
	created, err := NewScanComputationClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanComputationClauseTyped(arrowToken *SyntaxToken, assignments *SeparatedSyntaxListOf[*ScanAssignment], diagnostics ...Diagnostic) *ScanComputationClause {
	node, err := NewScanComputationClauseTyped(arrowToken, assignments, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanComputationClause) CloneTyped(includeDiagnostics bool) *ScanComputationClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanComputationClause{SyntaxNode: cloned}
}

func (n *ScanComputationClause) ArrowToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanComputationClause) Assignments() *SeparatedSyntaxListOf[*ScanAssignment] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*ScanAssignment](list)
}

type ScanAssignment struct {
	*SyntaxNode
}

func NewScanAssignment(children []SyntaxElement, diagnostics ...Diagnostic) (*ScanAssignment, error) {
	node, err := NewGeneratedSyntaxNodeByName("ScanAssignment", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &ScanAssignment{SyntaxNode: node}, nil
}

func MustNewScanAssignment(children []SyntaxElement, diagnostics ...Diagnostic) *ScanAssignment {
	node, err := NewScanAssignment(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewScanAssignmentTyped(name *NameReference, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) (*ScanAssignment, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(name),
		syntaxElementFromTyped(equalToken),
		syntaxElementFromTyped(expression),
	}
	created, err := NewScanAssignment(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewScanAssignmentTyped(name *NameReference, equalToken *SyntaxToken, expression Expression, diagnostics ...Diagnostic) *ScanAssignment {
	node, err := NewScanAssignmentTyped(name, equalToken, expression, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *ScanAssignment) CloneTyped(includeDiagnostics bool) *ScanAssignment {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &ScanAssignment{SyntaxNode: cloned}
}

func (n *ScanAssignment) Name() *NameReference {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*NameReference)
	return child
}

func (n *ScanAssignment) EqualToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *ScanAssignment) Expression() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type PartitionByOperator struct {
	*SyntaxNode
}

func (*PartitionByOperator) isQueryOperator() {}

func (*PartitionByOperator) isExpression() {}

func NewPartitionByOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionByOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionByOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionByOperator{SyntaxNode: node}, nil
}

func MustNewPartitionByOperator(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionByOperator {
	node, err := NewPartitionByOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionByOperatorTyped(partitionByKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], entity Expression, idClause *PartitionByIdClause, openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) (*PartitionByOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(partitionByKeyword),
		syntaxElementFromSyntaxListOf(parameters),
		syntaxElementFromTyped(entity),
		syntaxElementFromTyped(idClause),
		syntaxElementFromTyped(openParen),
		syntaxElementFromTyped(subquery),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewPartitionByOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionByOperatorTyped(partitionByKeyword *SyntaxToken, parameters *SyntaxListOf[*NamedParameter], entity Expression, idClause *PartitionByIdClause, openParen *SyntaxToken, subquery Expression, closeParen *SyntaxToken, diagnostics ...Diagnostic) *PartitionByOperator {
	node, err := NewPartitionByOperatorTyped(partitionByKeyword, parameters, entity, idClause, openParen, subquery, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionByOperator) CloneTyped(includeDiagnostics bool) *PartitionByOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionByOperator{SyntaxNode: cloned}
}

func (n *PartitionByOperator) PartitionByKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionByOperator) Parameters() *SyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*NamedParameter](list)
}

func (n *PartitionByOperator) Entity() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PartitionByOperator) IdClause() *PartitionByIdClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*PartitionByIdClause)
	return child
}

func (n *PartitionByOperator) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(4)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionByOperator) Subquery() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(5)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *PartitionByOperator) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(6)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type PartitionByIdClause struct {
	*SyntaxNode
}

func NewPartitionByIdClause(children []SyntaxElement, diagnostics ...Diagnostic) (*PartitionByIdClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartitionByIdClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartitionByIdClause{SyntaxNode: node}, nil
}

func MustNewPartitionByIdClause(children []SyntaxElement, diagnostics ...Diagnostic) *PartitionByIdClause {
	node, err := NewPartitionByIdClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartitionByIdClauseTyped(idKeyword *SyntaxToken, value Expression, diagnostics ...Diagnostic) (*PartitionByIdClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(idKeyword),
		syntaxElementFromTyped(value),
	}
	created, err := NewPartitionByIdClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewPartitionByIdClauseTyped(idKeyword *SyntaxToken, value Expression, diagnostics ...Diagnostic) *PartitionByIdClause {
	node, err := NewPartitionByIdClauseTyped(idKeyword, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartitionByIdClause) CloneTyped(includeDiagnostics bool) *PartitionByIdClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartitionByIdClause{SyntaxNode: cloned}
}

func (n *PartitionByIdClause) IdKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartitionByIdClause) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type CommandWithValueClause struct {
	*SyntaxNode
}

func (*CommandWithValueClause) isCommandWithClause() {}

func (*CommandWithValueClause) isClause() {}

func NewCommandWithValueClause(children []SyntaxElement, diagnostics ...Diagnostic) (*CommandWithValueClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("CommandWithValueClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CommandWithValueClause{SyntaxNode: node}, nil
}

func MustNewCommandWithValueClause(children []SyntaxElement, diagnostics ...Diagnostic) *CommandWithValueClause {
	node, err := NewCommandWithValueClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCommandWithValueClauseTyped(withKeyword *SyntaxToken, value Expression, diagnostics ...Diagnostic) (*CommandWithValueClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(value),
	}
	created, err := NewCommandWithValueClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCommandWithValueClauseTyped(withKeyword *SyntaxToken, value Expression, diagnostics ...Diagnostic) *CommandWithValueClause {
	node, err := NewCommandWithValueClauseTyped(withKeyword, value, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CommandWithValueClause) CloneTyped(includeDiagnostics bool) *CommandWithValueClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CommandWithValueClause{SyntaxNode: cloned}
}

func (n *CommandWithValueClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CommandWithValueClause) Value() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

type CommandWithPropertyListClause struct {
	*SyntaxNode
}

func (*CommandWithPropertyListClause) isCommandWithClause() {}

func (*CommandWithPropertyListClause) isClause() {}

func NewCommandWithPropertyListClause(children []SyntaxElement, diagnostics ...Diagnostic) (*CommandWithPropertyListClause, error) {
	node, err := NewGeneratedSyntaxNodeByName("CommandWithPropertyListClause", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CommandWithPropertyListClause{SyntaxNode: node}, nil
}

func MustNewCommandWithPropertyListClause(children []SyntaxElement, diagnostics ...Diagnostic) *CommandWithPropertyListClause {
	node, err := NewCommandWithPropertyListClause(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCommandWithPropertyListClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) (*CommandWithPropertyListClause, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(withKeyword),
		syntaxElementFromTyped(openParen),
		syntaxElementFromSeparatedSyntaxListOf(properties),
		syntaxElementFromTyped(closeParen),
	}
	created, err := NewCommandWithPropertyListClause(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCommandWithPropertyListClauseTyped(withKeyword *SyntaxToken, openParen *SyntaxToken, properties *SeparatedSyntaxListOf[*NamedParameter], closeParen *SyntaxToken, diagnostics ...Diagnostic) *CommandWithPropertyListClause {
	node, err := NewCommandWithPropertyListClauseTyped(withKeyword, openParen, properties, closeParen, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CommandWithPropertyListClause) CloneTyped(includeDiagnostics bool) *CommandWithPropertyListClause {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CommandWithPropertyListClause{SyntaxNode: cloned}
}

func (n *CommandWithPropertyListClause) WithKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CommandWithPropertyListClause) OpenParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CommandWithPropertyListClause) Properties() *SeparatedSyntaxListOf[*NamedParameter] {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[*NamedParameter](list)
}

func (n *CommandWithPropertyListClause) CloseParen() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type UnknownCommand struct {
	*SyntaxNode
}

func (*UnknownCommand) isCommand() {}

func (*UnknownCommand) isExpression() {}

func NewUnknownCommand(children []SyntaxElement, diagnostics ...Diagnostic) (*UnknownCommand, error) {
	node, err := NewGeneratedSyntaxNodeByName("UnknownCommand", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &UnknownCommand{SyntaxNode: node}, nil
}

func MustNewUnknownCommand(children []SyntaxElement, diagnostics ...Diagnostic) *UnknownCommand {
	node, err := NewUnknownCommand(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewUnknownCommandTyped(dotToken *SyntaxToken, parts *SyntaxListOf[SyntaxElement], diagnostics ...Diagnostic) (*UnknownCommand, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dotToken),
		syntaxElementFromSyntaxListOf(parts),
	}
	created, err := NewUnknownCommand(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewUnknownCommandTyped(dotToken *SyntaxToken, parts *SyntaxListOf[SyntaxElement], diagnostics ...Diagnostic) *UnknownCommand {
	node, err := NewUnknownCommandTyped(dotToken, parts, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *UnknownCommand) CloneTyped(includeDiagnostics bool) *UnknownCommand {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &UnknownCommand{SyntaxNode: cloned}
}

func (n *UnknownCommand) DotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *UnknownCommand) Parts() *SyntaxListOf[SyntaxElement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[SyntaxElement](list)
}

type CustomCommand struct {
	*SyntaxNode
}

func (*CustomCommand) isCommand() {}

func (*CustomCommand) isExpression() {}

func NewCustomCommand(children []SyntaxElement, diagnostics ...Diagnostic) (*CustomCommand, error) {
	node, err := NewGeneratedSyntaxNodeByName("CustomCommand", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CustomCommand{SyntaxNode: node}, nil
}

func MustNewCustomCommand(children []SyntaxElement, diagnostics ...Diagnostic) *CustomCommand {
	node, err := NewCustomCommand(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCustomCommandTyped(commandKind string, dotToken *SyntaxToken, custom SyntaxElement, diagnostics ...Diagnostic) (*CustomCommand, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dotToken),
		syntaxElementFromTyped(custom),
	}
	created, err := NewCustomCommand(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	created.SyntaxNode.setPayload("CommandKind", commandKind)
	return created, nil
}

func MustNewCustomCommandTyped(commandKind string, dotToken *SyntaxToken, custom SyntaxElement, diagnostics ...Diagnostic) *CustomCommand {
	node, err := NewCustomCommandTyped(commandKind, dotToken, custom, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CustomCommand) CloneTyped(includeDiagnostics bool) *CustomCommand {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CustomCommand{SyntaxNode: cloned}
}

func (n *CustomCommand) DotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *CustomCommand) Custom() SyntaxElement {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(SyntaxElement)
	return child
}

func (n *CustomCommand) CommandKind() string {
	if n == nil || n.SyntaxNode == nil {
		return ""
	}
	payload, ok := n.SyntaxNode.Payload("CommandKind")
	if !ok {
		return ""
	}
	value, _ := payload.(string)
	return value
}

type PartialCommand struct {
	*SyntaxNode
}

func (*PartialCommand) isCommand() {}

func (*PartialCommand) isExpression() {}

func NewPartialCommand(children []SyntaxElement, diagnostics ...Diagnostic) (*PartialCommand, error) {
	node, err := NewGeneratedSyntaxNodeByName("PartialCommand", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &PartialCommand{SyntaxNode: node}, nil
}

func MustNewPartialCommand(children []SyntaxElement, diagnostics ...Diagnostic) *PartialCommand {
	node, err := NewPartialCommand(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewPartialCommandTyped(commandKinds []string, dotToken *SyntaxToken, parts *SyntaxListOf[SyntaxElement], diagnostics ...Diagnostic) (*PartialCommand, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dotToken),
		syntaxElementFromSyntaxListOf(parts),
	}
	created, err := NewPartialCommand(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	created.SyntaxNode.setPayload("CommandKinds", append([]string{}, commandKinds...))
	return created, nil
}

func MustNewPartialCommandTyped(commandKinds []string, dotToken *SyntaxToken, parts *SyntaxListOf[SyntaxElement], diagnostics ...Diagnostic) *PartialCommand {
	node, err := NewPartialCommandTyped(commandKinds, dotToken, parts, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *PartialCommand) CloneTyped(includeDiagnostics bool) *PartialCommand {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &PartialCommand{SyntaxNode: cloned}
}

func (n *PartialCommand) DotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *PartialCommand) Parts() *SyntaxListOf[SyntaxElement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[SyntaxElement](list)
}

func (n *PartialCommand) CommandKinds() []string {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	payload, ok := n.SyntaxNode.Payload("CommandKinds")
	if !ok {
		return nil
	}
	value, _ := payload.([]string)
	return append([]string{}, value...)
}

type CommandAndSkippedTokens struct {
	*SyntaxNode
}

func (*CommandAndSkippedTokens) isCommand() {}

func (*CommandAndSkippedTokens) isExpression() {}

func NewCommandAndSkippedTokens(children []SyntaxElement, diagnostics ...Diagnostic) (*CommandAndSkippedTokens, error) {
	node, err := NewGeneratedSyntaxNodeByName("CommandAndSkippedTokens", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CommandAndSkippedTokens{SyntaxNode: node}, nil
}

func MustNewCommandAndSkippedTokens(children []SyntaxElement, diagnostics ...Diagnostic) *CommandAndSkippedTokens {
	node, err := NewCommandAndSkippedTokens(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCommandAndSkippedTokensTyped(command Command, skippedTokens *SkippedTokens, diagnostics ...Diagnostic) (*CommandAndSkippedTokens, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(command),
		syntaxElementFromTyped(skippedTokens),
	}
	created, err := NewCommandAndSkippedTokens(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCommandAndSkippedTokensTyped(command Command, skippedTokens *SkippedTokens, diagnostics ...Diagnostic) *CommandAndSkippedTokens {
	node, err := NewCommandAndSkippedTokensTyped(command, skippedTokens, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CommandAndSkippedTokens) CloneTyped(includeDiagnostics bool) *CommandAndSkippedTokens {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CommandAndSkippedTokens{SyntaxNode: cloned}
}

func (n *CommandAndSkippedTokens) Command() Command {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Command)
	return child
}

func (n *CommandAndSkippedTokens) SkippedTokens() *SkippedTokens {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*SkippedTokens)
	return child
}

type BadCommand struct {
	*SyntaxNode
}

func (*BadCommand) isCommand() {}

func (*BadCommand) isExpression() {}

func NewBadCommand(children []SyntaxElement, diagnostics ...Diagnostic) (*BadCommand, error) {
	node, err := NewGeneratedSyntaxNodeByName("BadCommand", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &BadCommand{SyntaxNode: node}, nil
}

func MustNewBadCommand(children []SyntaxElement, diagnostics ...Diagnostic) *BadCommand {
	node, err := NewBadCommand(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewBadCommandTyped(dotToken *SyntaxToken, diagnostics ...Diagnostic) (*BadCommand, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(dotToken),
	}
	created, err := NewBadCommand(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewBadCommandTyped(dotToken *SyntaxToken, diagnostics ...Diagnostic) *BadCommand {
	node, err := NewBadCommandTyped(dotToken, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *BadCommand) CloneTyped(includeDiagnostics bool) *BadCommand {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &BadCommand{SyntaxNode: cloned}
}

func (n *BadCommand) DotToken() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type CommandBlock struct {
	*SyntaxNode
}

func NewCommandBlock(children []SyntaxElement, diagnostics ...Diagnostic) (*CommandBlock, error) {
	node, err := NewGeneratedSyntaxNodeByName("CommandBlock", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &CommandBlock{SyntaxNode: node}, nil
}

func MustNewCommandBlock(children []SyntaxElement, diagnostics ...Diagnostic) *CommandBlock {
	node, err := NewCommandBlock(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCommandBlockTyped(directives *SyntaxListOf[*Directive], statements *SeparatedSyntaxListOf[Statement], skippedTokens *SkippedTokens, endOfCommand *SyntaxToken, diagnostics ...Diagnostic) (*CommandBlock, error) {
	children := []SyntaxElement{
		syntaxElementFromSyntaxListOf(directives),
		syntaxElementFromSeparatedSyntaxListOf(statements),
		syntaxElementFromTyped(skippedTokens),
		syntaxElementFromTyped(endOfCommand),
	}
	created, err := NewCommandBlock(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewCommandBlockTyped(directives *SyntaxListOf[*Directive], statements *SeparatedSyntaxListOf[Statement], skippedTokens *SkippedTokens, endOfCommand *SyntaxToken, diagnostics ...Diagnostic) *CommandBlock {
	node, err := NewCommandBlockTyped(directives, statements, skippedTokens, endOfCommand, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *CommandBlock) CloneTyped(includeDiagnostics bool) *CommandBlock {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &CommandBlock{SyntaxNode: cloned}
}

func (n *CommandBlock) Directives() *SyntaxListOf[*Directive] {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSyntaxListOf[*Directive](list)
}

func (n *CommandBlock) Statements() *SeparatedSyntaxListOf[Statement] {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	list, _ := element.(*SyntaxList)
	return AsSeparatedSyntaxListOf[Statement](list)
}

func (n *CommandBlock) SkippedTokens() *SkippedTokens {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(*SkippedTokens)
	return child
}

func (n *CommandBlock) EndOfCommand() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(3)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

type GraphMergeOperator struct {
	*SyntaxNode
}

func (*GraphMergeOperator) isQueryOperator() {}

func (*GraphMergeOperator) isExpression() {}

func NewGraphMergeOperator(children []SyntaxElement, diagnostics ...Diagnostic) (*GraphMergeOperator, error) {
	node, err := NewGeneratedSyntaxNodeByName("GraphMergeOperator", children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return &GraphMergeOperator{SyntaxNode: node}, nil
}

func MustNewGraphMergeOperator(children []SyntaxElement, diagnostics ...Diagnostic) *GraphMergeOperator {
	node, err := NewGraphMergeOperator(children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func NewGraphMergeOperatorTyped(graphMergeKeyword *SyntaxToken, graph Expression, onClause JoinConditionClause, diagnostics ...Diagnostic) (*GraphMergeOperator, error) {
	children := []SyntaxElement{
		syntaxElementFromTyped(graphMergeKeyword),
		syntaxElementFromTyped(graph),
		syntaxElementFromTyped(onClause),
	}
	created, err := NewGraphMergeOperator(children, diagnostics...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func MustNewGraphMergeOperatorTyped(graphMergeKeyword *SyntaxToken, graph Expression, onClause JoinConditionClause, diagnostics ...Diagnostic) *GraphMergeOperator {
	node, err := NewGraphMergeOperatorTyped(graphMergeKeyword, graph, onClause, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GraphMergeOperator) CloneTyped(includeDiagnostics bool) *GraphMergeOperator {
	if n == nil || n.SyntaxNode == nil {
		return nil
	}
	cloned, _ := n.SyntaxNode.Clone(includeDiagnostics).(*SyntaxNode)
	return &GraphMergeOperator{SyntaxNode: cloned}
}

func (n *GraphMergeOperator) GraphMergeKeyword() *SyntaxToken {
	if n == nil {
		return nil
	}
	element := n.GetChild(0)
	if element == nil {
		return nil
	}
	child, _ := element.(*SyntaxToken)
	return child
}

func (n *GraphMergeOperator) Graph() Expression {
	if n == nil {
		return nil
	}
	element := n.GetChild(1)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(Expression)
	return child
}

func (n *GraphMergeOperator) OnClause() JoinConditionClause {
	if n == nil {
		return nil
	}
	element := n.GetChild(2)
	if element == nil {
		return nil
	}
	child, _ := WrapGeneratedSyntaxElement(element).(JoinConditionClause)
	return child
}

func WrapGeneratedSyntaxElement(element SyntaxElement) SyntaxElement {
	if node, ok := element.(*SyntaxNode); ok {
		return WrapGeneratedSyntaxNode(node)
	}
	return element
}

func WrapGeneratedSyntaxNode(node *SyntaxNode) SyntaxElement {
	if node == nil {
		return nil
	}
	switch node.GeneratedNodeName() {
	case "DirectiveBlock":
		return &DirectiveBlock{SyntaxNode: node}
	case "Directive":
		return &Directive{SyntaxNode: node}
	case "SkippedTokens":
		return &SkippedTokens{SyntaxNode: node}
	case "TypeOfLiteralExpression":
		return &TypeOfLiteralExpression{SyntaxNode: node}
	case "BadQueryOperator":
		return &BadQueryOperator{SyntaxNode: node}
	case "CompoundStringLiteralExpression":
		return &CompoundStringLiteralExpression{SyntaxNode: node}
	case "TokenName":
		return &TokenName{SyntaxNode: node}
	case "BracketedName":
		return &BracketedName{SyntaxNode: node}
	case "BracedName":
		return &BracedName{SyntaxNode: node}
	case "WildcardedName":
		return &WildcardedName{SyntaxNode: node}
	case "BracketedWildcardedName":
		return &BracketedWildcardedName{SyntaxNode: node}
	case "NameDeclaration":
		return &NameDeclaration{SyntaxNode: node}
	case "NameReference":
		return &NameReference{SyntaxNode: node}
	case "LiteralExpression":
		return &LiteralExpression{SyntaxNode: node}
	case "StarExpression":
		return &StarExpression{SyntaxNode: node}
	case "AtExpression":
		return &AtExpression{SyntaxNode: node}
	case "JsonPair":
		return &JsonPair{SyntaxNode: node}
	case "JsonObjectExpression":
		return &JsonObjectExpression{SyntaxNode: node}
	case "JsonArrayExpression":
		return &JsonArrayExpression{SyntaxNode: node}
	case "DynamicExpression":
		return &DynamicExpression{SyntaxNode: node}
	case "ParenthesizedExpression":
		return &ParenthesizedExpression{SyntaxNode: node}
	case "ExpressionList":
		return &ExpressionList{SyntaxNode: node}
	case "ExpressionCouple":
		return &ExpressionCouple{SyntaxNode: node}
	case "PrefixUnaryExpression":
		return &PrefixUnaryExpression{SyntaxNode: node}
	case "BinaryExpression":
		return &BinaryExpression{SyntaxNode: node}
	case "InExpression":
		return &InExpression{SyntaxNode: node}
	case "HasAnyExpression":
		return &HasAnyExpression{SyntaxNode: node}
	case "HasAllExpression":
		return &HasAllExpression{SyntaxNode: node}
	case "BetweenExpression":
		return &BetweenExpression{SyntaxNode: node}
	case "FunctionCallExpression":
		return &FunctionCallExpression{SyntaxNode: node}
	case "ToScalarExpression":
		return &ToScalarExpression{SyntaxNode: node}
	case "ToTableExpression":
		return &ToTableExpression{SyntaxNode: node}
	case "MaterializedViewCombineExpression":
		return &MaterializedViewCombineExpression{SyntaxNode: node}
	case "MaterializedViewCombineNameClause":
		return &MaterializedViewCombineNameClause{SyntaxNode: node}
	case "MaterializedViewCombineClause":
		return &MaterializedViewCombineClause{SyntaxNode: node}
	case "SimpleNamedExpression":
		return &SimpleNamedExpression{SyntaxNode: node}
	case "RenameList":
		return &RenameList{SyntaxNode: node}
	case "CompoundNamedExpression":
		return &CompoundNamedExpression{SyntaxNode: node}
	case "BracketedExpression":
		return &BracketedExpression{SyntaxNode: node}
	case "PathExpression":
		return &PathExpression{SyntaxNode: node}
	case "ElementExpression":
		return &ElementExpression{SyntaxNode: node}
	case "PipeExpression":
		return &PipeExpression{SyntaxNode: node}
	case "RangeOperator":
		return &RangeOperator{SyntaxNode: node}
	case "NamedParameter":
		return &NamedParameter{SyntaxNode: node}
	case "ConsumeOperator":
		return &ConsumeOperator{SyntaxNode: node}
	case "CountOperator":
		return &CountOperator{SyntaxNode: node}
	case "CountAsIdentifierClause":
		return &CountAsIdentifierClause{SyntaxNode: node}
	case "ExecuteAndCacheOperator":
		return &ExecuteAndCacheOperator{SyntaxNode: node}
	case "ExtendOperator":
		return &ExtendOperator{SyntaxNode: node}
	case "FacetOperator":
		return &FacetOperator{SyntaxNode: node}
	case "FacetWithOperatorClause":
		return &FacetWithOperatorClause{SyntaxNode: node}
	case "FacetWithExpressionClause":
		return &FacetWithExpressionClause{SyntaxNode: node}
	case "FilterOperator":
		return &FilterOperator{SyntaxNode: node}
	case "GetSchemaOperator":
		return &GetSchemaOperator{SyntaxNode: node}
	case "FindOperator":
		return &FindOperator{SyntaxNode: node}
	case "DataScopeClause":
		return &DataScopeClause{SyntaxNode: node}
	case "TypedColumnReference":
		return &TypedColumnReference{SyntaxNode: node}
	case "FindInClause":
		return &FindInClause{SyntaxNode: node}
	case "FindProjectClause":
		return &FindProjectClause{SyntaxNode: node}
	case "PackExpression":
		return &PackExpression{SyntaxNode: node}
	case "NameAndTypeDeclaration":
		return &NameAndTypeDeclaration{SyntaxNode: node}
	case "PrimitiveTypeExpression":
		return &PrimitiveTypeExpression{SyntaxNode: node}
	case "SearchOperator":
		return &SearchOperator{SyntaxNode: node}
	case "ForkOperator":
		return &ForkOperator{SyntaxNode: node}
	case "NameEqualsClause":
		return &NameEqualsClause{SyntaxNode: node}
	case "ForkExpression":
		return &ForkExpression{SyntaxNode: node}
	case "MakeSeriesOperator":
		return &MakeSeriesOperator{SyntaxNode: node}
	case "MakeSeriesExpression":
		return &MakeSeriesExpression{SyntaxNode: node}
	case "DefaultExpressionClause":
		return &DefaultExpressionClause{SyntaxNode: node}
	case "MakeSeriesOnClause":
		return &MakeSeriesOnClause{SyntaxNode: node}
	case "MakeSeriesFromClause":
		return &MakeSeriesFromClause{SyntaxNode: node}
	case "MakeSeriesToClause":
		return &MakeSeriesToClause{SyntaxNode: node}
	case "MakeSeriesStepClause":
		return &MakeSeriesStepClause{SyntaxNode: node}
	case "MakeSeriesInRangeClause":
		return &MakeSeriesInRangeClause{SyntaxNode: node}
	case "MakeSeriesFromToStepClause":
		return &MakeSeriesFromToStepClause{SyntaxNode: node}
	case "MakeSeriesByClause":
		return &MakeSeriesByClause{SyntaxNode: node}
	case "MvExpandOperator":
		return &MvExpandOperator{SyntaxNode: node}
	case "MvExpandExpression":
		return &MvExpandExpression{SyntaxNode: node}
	case "MvExpandRowLimitClause":
		return &MvExpandRowLimitClause{SyntaxNode: node}
	case "MvApplyOperator":
		return &MvApplyOperator{SyntaxNode: node}
	case "MvApplyExpression":
		return &MvApplyExpression{SyntaxNode: node}
	case "MvApplyRowLimitClause":
		return &MvApplyRowLimitClause{SyntaxNode: node}
	case "MvApplyContextIdClause":
		return &MvApplyContextIdClause{SyntaxNode: node}
	case "MvApplySubqueryExpression":
		return &MvApplySubqueryExpression{SyntaxNode: node}
	case "ToTypeOfClause":
		return &ToTypeOfClause{SyntaxNode: node}
	case "EvaluateSchemaClause":
		return &EvaluateSchemaClause{SyntaxNode: node}
	case "EvaluateOperator":
		return &EvaluateOperator{SyntaxNode: node}
	case "ParseOperator":
		return &ParseOperator{SyntaxNode: node}
	case "ParseWhereOperator":
		return &ParseWhereOperator{SyntaxNode: node}
	case "ParseKvWithClause":
		return &ParseKvWithClause{SyntaxNode: node}
	case "ParseKvOperator":
		return &ParseKvOperator{SyntaxNode: node}
	case "PartitionOperator":
		return &PartitionOperator{SyntaxNode: node}
	case "PartitionQuery":
		return &PartitionQuery{SyntaxNode: node}
	case "PartitionScope":
		return &PartitionScope{SyntaxNode: node}
	case "PartitionSubquery":
		return &PartitionSubquery{SyntaxNode: node}
	case "ProjectOperator":
		return &ProjectOperator{SyntaxNode: node}
	case "ProjectAwayOperator":
		return &ProjectAwayOperator{SyntaxNode: node}
	case "ProjectByNamesOperator":
		return &ProjectByNamesOperator{SyntaxNode: node}
	case "ProjectKeepOperator":
		return &ProjectKeepOperator{SyntaxNode: node}
	case "ProjectRenameOperator":
		return &ProjectRenameOperator{SyntaxNode: node}
	case "ProjectReorderOperator":
		return &ProjectReorderOperator{SyntaxNode: node}
	case "SampleOperator":
		return &SampleOperator{SyntaxNode: node}
	case "SampleDistinctOperator":
		return &SampleDistinctOperator{SyntaxNode: node}
	case "EntityGroup":
		return &EntityGroup{SyntaxNode: node}
	case "ReduceByOperator":
		return &ReduceByOperator{SyntaxNode: node}
	case "ReduceByWithClause":
		return &ReduceByWithClause{SyntaxNode: node}
	case "SummarizeOperator":
		return &SummarizeOperator{SyntaxNode: node}
	case "MacroExpandScopeReferenceName":
		return &MacroExpandScopeReferenceName{SyntaxNode: node}
	case "MacroExpandOperator":
		return &MacroExpandOperator{SyntaxNode: node}
	case "SummarizeByClause":
		return &SummarizeByClause{SyntaxNode: node}
	case "DistinctOperator":
		return &DistinctOperator{SyntaxNode: node}
	case "TakeOperator":
		return &TakeOperator{SyntaxNode: node}
	case "SortOperator":
		return &SortOperator{SyntaxNode: node}
	case "OrderedExpression":
		return &OrderedExpression{SyntaxNode: node}
	case "OrderingClause":
		return &OrderingClause{SyntaxNode: node}
	case "OrderingNullsClause":
		return &OrderingNullsClause{SyntaxNode: node}
	case "TopHittersOperator":
		return &TopHittersOperator{SyntaxNode: node}
	case "TopHittersByClause":
		return &TopHittersByClause{SyntaxNode: node}
	case "TopOperator":
		return &TopOperator{SyntaxNode: node}
	case "TopNestedOperator":
		return &TopNestedOperator{SyntaxNode: node}
	case "TopNestedClause":
		return &TopNestedClause{SyntaxNode: node}
	case "TopNestedWithOthersClause":
		return &TopNestedWithOthersClause{SyntaxNode: node}
	case "UnionOperator":
		return &UnionOperator{SyntaxNode: node}
	case "AsOperator":
		return &AsOperator{SyntaxNode: node}
	case "SerializeOperator":
		return &SerializeOperator{SyntaxNode: node}
	case "InvokeOperator":
		return &InvokeOperator{SyntaxNode: node}
	case "RenderOperator":
		return &RenderOperator{SyntaxNode: node}
	case "MakeGraphOperator":
		return &MakeGraphOperator{SyntaxNode: node}
	case "MakeGraphWithTablesAndKeysClause":
		return &MakeGraphWithTablesAndKeysClause{SyntaxNode: node}
	case "MakeGraphWithImplicitIdClause":
		return &MakeGraphWithImplicitIdClause{SyntaxNode: node}
	case "GraphMarkComponentsOperator":
		return &GraphMarkComponentsOperator{SyntaxNode: node}
	case "GraphWhereNodesOperator":
		return &GraphWhereNodesOperator{SyntaxNode: node}
	case "GraphWhereEdgesOperator":
		return &GraphWhereEdgesOperator{SyntaxNode: node}
	case "MakeGraphTableAndKeyClause":
		return &MakeGraphTableAndKeyClause{SyntaxNode: node}
	case "MakeGraphPartitionedByClause":
		return &MakeGraphPartitionedByClause{SyntaxNode: node}
	case "GraphToTableOperator":
		return &GraphToTableOperator{SyntaxNode: node}
	case "GraphToTableOutputClause":
		return &GraphToTableOutputClause{SyntaxNode: node}
	case "GraphToTableAsClause":
		return &GraphToTableAsClause{SyntaxNode: node}
	case "GraphMatchOperator":
		return &GraphMatchOperator{SyntaxNode: node}
	case "GraphShortestPathsOperator":
		return &GraphShortestPathsOperator{SyntaxNode: node}
	case "GraphMatchPattern":
		return &GraphMatchPattern{SyntaxNode: node}
	case "GraphMatchPatternNode":
		return &GraphMatchPatternNode{SyntaxNode: node}
	case "GraphMatchPatternEdge":
		return &GraphMatchPatternEdge{SyntaxNode: node}
	case "GraphMatchPatternEdgeRange":
		return &GraphMatchPatternEdgeRange{SyntaxNode: node}
	case "WhereClause":
		return &WhereClause{SyntaxNode: node}
	case "ProjectClause":
		return &ProjectClause{SyntaxNode: node}
	case "NameReferenceList":
		return &NameReferenceList{SyntaxNode: node}
	case "RenderWithClause":
		return &RenderWithClause{SyntaxNode: node}
	case "PrintOperator":
		return &PrintOperator{SyntaxNode: node}
	case "AliasStatement":
		return &AliasStatement{SyntaxNode: node}
	case "LetStatement":
		return &LetStatement{SyntaxNode: node}
	case "FunctionDeclaration":
		return &FunctionDeclaration{SyntaxNode: node}
	case "FunctionParameters":
		return &FunctionParameters{SyntaxNode: node}
	case "FunctionParameter":
		return &FunctionParameter{SyntaxNode: node}
	case "DefaultValueDeclaration":
		return &DefaultValueDeclaration{SyntaxNode: node}
	case "FunctionBody":
		return &FunctionBody{SyntaxNode: node}
	case "SchemaTypeExpression":
		return &SchemaTypeExpression{SyntaxNode: node}
	case "ExpressionStatement":
		return &ExpressionStatement{SyntaxNode: node}
	case "MaterializeExpression":
		return &MaterializeExpression{SyntaxNode: node}
	case "SetOptionStatement":
		return &SetOptionStatement{SyntaxNode: node}
	case "OptionValueClause":
		return &OptionValueClause{SyntaxNode: node}
	case "QueryParametersStatement":
		return &QueryParametersStatement{SyntaxNode: node}
	case "RestrictStatementWithClause":
		return &RestrictStatementWithClause{SyntaxNode: node}
	case "RestrictStatement":
		return &RestrictStatement{SyntaxNode: node}
	case "PatternStatement":
		return &PatternStatement{SyntaxNode: node}
	case "PatternDeclaration":
		return &PatternDeclaration{SyntaxNode: node}
	case "PatternPathParameter":
		return &PatternPathParameter{SyntaxNode: node}
	case "PatternMatch":
		return &PatternMatch{SyntaxNode: node}
	case "PatternPathValue":
		return &PatternPathValue{SyntaxNode: node}
	case "DataScopeExpression":
		return &DataScopeExpression{SyntaxNode: node}
	case "DataTableExpression":
		return &DataTableExpression{SyntaxNode: node}
	case "RowSchema":
		return &RowSchema{SyntaxNode: node}
	case "EvaluateRowSchema":
		return &EvaluateRowSchema{SyntaxNode: node}
	case "ExternalDataExpression":
		return &ExternalDataExpression{SyntaxNode: node}
	case "ContextualDataTableExpression":
		return &ContextualDataTableExpression{SyntaxNode: node}
	case "ExternalDataWithClause":
		return &ExternalDataWithClause{SyntaxNode: node}
	case "InlineExternalTableKindClause":
		return &InlineExternalTableKindClause{SyntaxNode: node}
	case "InlineExternalTableDataFormatClause":
		return &InlineExternalTableDataFormatClause{SyntaxNode: node}
	case "DateTimePattern":
		return &DateTimePattern{SyntaxNode: node}
	case "InlineExternalTablePathFormatPartitionColumnReference":
		return &InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node}
	case "InlineExternalTablePathFormatClause":
		return &InlineExternalTablePathFormatClause{SyntaxNode: node}
	case "PartitionColumnDeclaration":
		return &PartitionColumnDeclaration{SyntaxNode: node}
	case "InlineExternalTablePartitionClause":
		return &InlineExternalTablePartitionClause{SyntaxNode: node}
	case "InlineExternalTableConnectionStringsClause":
		return &InlineExternalTableConnectionStringsClause{SyntaxNode: node}
	case "InlineExternalTableExpression":
		return &InlineExternalTableExpression{SyntaxNode: node}
	case "JoinOperator":
		return &JoinOperator{SyntaxNode: node}
	case "LookupOperator":
		return &LookupOperator{SyntaxNode: node}
	case "JoinOnClause":
		return &JoinOnClause{SyntaxNode: node}
	case "JoinWhereClause":
		return &JoinWhereClause{SyntaxNode: node}
	case "ScanOperator":
		return &ScanOperator{SyntaxNode: node}
	case "AssertSchemaOperator":
		return &AssertSchemaOperator{SyntaxNode: node}
	case "ScanDeclareClause":
		return &ScanDeclareClause{SyntaxNode: node}
	case "ScanOrderByClause":
		return &ScanOrderByClause{SyntaxNode: node}
	case "ScanPartitionByClause":
		return &ScanPartitionByClause{SyntaxNode: node}
	case "ScanStepOutput":
		return &ScanStepOutput{SyntaxNode: node}
	case "ScanStep":
		return &ScanStep{SyntaxNode: node}
	case "ScanComputationClause":
		return &ScanComputationClause{SyntaxNode: node}
	case "ScanAssignment":
		return &ScanAssignment{SyntaxNode: node}
	case "PartitionByOperator":
		return &PartitionByOperator{SyntaxNode: node}
	case "PartitionByIdClause":
		return &PartitionByIdClause{SyntaxNode: node}
	case "CommandWithValueClause":
		return &CommandWithValueClause{SyntaxNode: node}
	case "CommandWithPropertyListClause":
		return &CommandWithPropertyListClause{SyntaxNode: node}
	case "UnknownCommand":
		return &UnknownCommand{SyntaxNode: node}
	case "CustomCommand":
		return &CustomCommand{SyntaxNode: node}
	case "PartialCommand":
		return &PartialCommand{SyntaxNode: node}
	case "CommandAndSkippedTokens":
		return &CommandAndSkippedTokens{SyntaxNode: node}
	case "BadCommand":
		return &BadCommand{SyntaxNode: node}
	case "CommandBlock":
		return &CommandBlock{SyntaxNode: node}
	case "GraphMergeOperator":
		return &GraphMergeOperator{SyntaxNode: node}
	}
	switch node.Kind() {
	case SyntaxKindDirectiveBlock:
		return &DirectiveBlock{SyntaxNode: node}
	case SyntaxKindDirective:
		return &Directive{SyntaxNode: node}
	case SyntaxKindSkippedTokens:
		return &SkippedTokens{SyntaxNode: node}
	case SyntaxKindTypeOfLiteralExpression:
		return &TypeOfLiteralExpression{SyntaxNode: node}
	case SyntaxKindBadQueryOperator:
		return &BadQueryOperator{SyntaxNode: node}
	case SyntaxKindCompoundStringLiteralExpression:
		return &CompoundStringLiteralExpression{SyntaxNode: node}
	case SyntaxKindTokenName:
		return &TokenName{SyntaxNode: node}
	case SyntaxKindBracketedName:
		return &BracketedName{SyntaxNode: node}
	case SyntaxKindBracedName:
		return &BracedName{SyntaxNode: node}
	case SyntaxKindWildcardedName:
		return &WildcardedName{SyntaxNode: node}
	case SyntaxKindBracketedWildcardedName:
		return &BracketedWildcardedName{SyntaxNode: node}
	case SyntaxKindNameDeclaration:
		return &NameDeclaration{SyntaxNode: node}
	case SyntaxKindNameReference:
		return &NameReference{SyntaxNode: node}
	case SyntaxKindStarExpression:
		return &StarExpression{SyntaxNode: node}
	case SyntaxKindAtExpression:
		return &AtExpression{SyntaxNode: node}
	case SyntaxKindJsonPair:
		return &JsonPair{SyntaxNode: node}
	case SyntaxKindJsonObjectExpression:
		return &JsonObjectExpression{SyntaxNode: node}
	case SyntaxKindJsonArrayExpression:
		return &JsonArrayExpression{SyntaxNode: node}
	case SyntaxKindDynamicExpression:
		return &DynamicExpression{SyntaxNode: node}
	case SyntaxKindParenthesizedExpression:
		return &ParenthesizedExpression{SyntaxNode: node}
	case SyntaxKindExpressionList:
		return &ExpressionList{SyntaxNode: node}
	case SyntaxKindExpressionCouple:
		return &ExpressionCouple{SyntaxNode: node}
	case SyntaxKindFunctionCallExpression:
		return &FunctionCallExpression{SyntaxNode: node}
	case SyntaxKindToScalarExpression:
		return &ToScalarExpression{SyntaxNode: node}
	case SyntaxKindToTableExpression:
		return &ToTableExpression{SyntaxNode: node}
	case SyntaxKindMaterializedViewCombineExpression:
		return &MaterializedViewCombineExpression{SyntaxNode: node}
	case SyntaxKindMaterializedViewCombineClause:
		return &MaterializedViewCombineClause{SyntaxNode: node}
	case SyntaxKindSimpleNamedExpression:
		return &SimpleNamedExpression{SyntaxNode: node}
	case SyntaxKindRenameList:
		return &RenameList{SyntaxNode: node}
	case SyntaxKindCompoundNamedExpression:
		return &CompoundNamedExpression{SyntaxNode: node}
	case SyntaxKindBracketedExpression:
		return &BracketedExpression{SyntaxNode: node}
	case SyntaxKindPathExpression:
		return &PathExpression{SyntaxNode: node}
	case SyntaxKindElementExpression:
		return &ElementExpression{SyntaxNode: node}
	case SyntaxKindPipeExpression:
		return &PipeExpression{SyntaxNode: node}
	case SyntaxKindRangeOperator:
		return &RangeOperator{SyntaxNode: node}
	case SyntaxKindNamedParameter:
		return &NamedParameter{SyntaxNode: node}
	case SyntaxKindConsumeOperator:
		return &ConsumeOperator{SyntaxNode: node}
	case SyntaxKindCountOperator:
		return &CountOperator{SyntaxNode: node}
	case SyntaxKindCountAsIdentifierClause:
		return &CountAsIdentifierClause{SyntaxNode: node}
	case SyntaxKindExecuteAndCacheOperator:
		return &ExecuteAndCacheOperator{SyntaxNode: node}
	case SyntaxKindExtendOperator:
		return &ExtendOperator{SyntaxNode: node}
	case SyntaxKindFacetOperator:
		return &FacetOperator{SyntaxNode: node}
	case SyntaxKindFacetWithOperatorClause:
		return &FacetWithOperatorClause{SyntaxNode: node}
	case SyntaxKindFacetWithExpressionClause:
		return &FacetWithExpressionClause{SyntaxNode: node}
	case SyntaxKindFilterOperator:
		return &FilterOperator{SyntaxNode: node}
	case SyntaxKindGetSchemaOperator:
		return &GetSchemaOperator{SyntaxNode: node}
	case SyntaxKindFindOperator:
		return &FindOperator{SyntaxNode: node}
	case SyntaxKindDataScopeClause:
		return &DataScopeClause{SyntaxNode: node}
	case SyntaxKindTypedColumnReference:
		return &TypedColumnReference{SyntaxNode: node}
	case SyntaxKindFindInClause:
		return &FindInClause{SyntaxNode: node}
	case SyntaxKindFindProjectClause:
		return &FindProjectClause{SyntaxNode: node}
	case SyntaxKindPackExpression:
		return &PackExpression{SyntaxNode: node}
	case SyntaxKindNameAndTypeDeclaration:
		return &NameAndTypeDeclaration{SyntaxNode: node}
	case SyntaxKindPrimitiveTypeExpression:
		return &PrimitiveTypeExpression{SyntaxNode: node}
	case SyntaxKindSearchOperator:
		return &SearchOperator{SyntaxNode: node}
	case SyntaxKindForkOperator:
		return &ForkOperator{SyntaxNode: node}
	case SyntaxKindNameEqualsClause:
		return &NameEqualsClause{SyntaxNode: node}
	case SyntaxKindForkExpression:
		return &ForkExpression{SyntaxNode: node}
	case SyntaxKindMakeSeriesOperator:
		return &MakeSeriesOperator{SyntaxNode: node}
	case SyntaxKindMakeSeriesExpression:
		return &MakeSeriesExpression{SyntaxNode: node}
	case SyntaxKindDefaultExpressionClause:
		return &DefaultExpressionClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesOnClause:
		return &MakeSeriesOnClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesFromClause:
		return &MakeSeriesFromClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesToClause:
		return &MakeSeriesToClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesStepClause:
		return &MakeSeriesStepClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesInRangeClause:
		return &MakeSeriesInRangeClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesFromToStepClause:
		return &MakeSeriesFromToStepClause{SyntaxNode: node}
	case SyntaxKindMakeSeriesByClause:
		return &MakeSeriesByClause{SyntaxNode: node}
	case SyntaxKindMvExpandOperator:
		return &MvExpandOperator{SyntaxNode: node}
	case SyntaxKindMvExpandExpression:
		return &MvExpandExpression{SyntaxNode: node}
	case SyntaxKindMvExpandRowLimitClause:
		return &MvExpandRowLimitClause{SyntaxNode: node}
	case SyntaxKindMvApplyOperator:
		return &MvApplyOperator{SyntaxNode: node}
	case SyntaxKindMvApplyExpression:
		return &MvApplyExpression{SyntaxNode: node}
	case SyntaxKindMvApplyRowLimitClause:
		return &MvApplyRowLimitClause{SyntaxNode: node}
	case SyntaxKindMvApplyContextIdClause:
		return &MvApplyContextIdClause{SyntaxNode: node}
	case SyntaxKindMvApplySubqueryExpression:
		return &MvApplySubqueryExpression{SyntaxNode: node}
	case SyntaxKindToTypeOfClause:
		return &ToTypeOfClause{SyntaxNode: node}
	case SyntaxKindEvaluateSchemaClause:
		return &EvaluateSchemaClause{SyntaxNode: node}
	case SyntaxKindEvaluateOperator:
		return &EvaluateOperator{SyntaxNode: node}
	case SyntaxKindParseOperator:
		return &ParseOperator{SyntaxNode: node}
	case SyntaxKindParseWhereOperator:
		return &ParseWhereOperator{SyntaxNode: node}
	case SyntaxKindParseKvWithClause:
		return &ParseKvWithClause{SyntaxNode: node}
	case SyntaxKindParseKvOperator:
		return &ParseKvOperator{SyntaxNode: node}
	case SyntaxKindPartitionOperator:
		return &PartitionOperator{SyntaxNode: node}
	case SyntaxKindPartitionQuery:
		return &PartitionQuery{SyntaxNode: node}
	case SyntaxKindPartitionScope:
		return &PartitionScope{SyntaxNode: node}
	case SyntaxKindPartitionSubquery:
		return &PartitionSubquery{SyntaxNode: node}
	case SyntaxKindProjectOperator:
		return &ProjectOperator{SyntaxNode: node}
	case SyntaxKindProjectAwayOperator:
		return &ProjectAwayOperator{SyntaxNode: node}
	case SyntaxKindProjectByNamesOperator:
		return &ProjectByNamesOperator{SyntaxNode: node}
	case SyntaxKindProjectKeepOperator:
		return &ProjectKeepOperator{SyntaxNode: node}
	case SyntaxKindProjectRenameOperator:
		return &ProjectRenameOperator{SyntaxNode: node}
	case SyntaxKindProjectReorderOperator:
		return &ProjectReorderOperator{SyntaxNode: node}
	case SyntaxKindSampleOperator:
		return &SampleOperator{SyntaxNode: node}
	case SyntaxKindSampleDistinctOperator:
		return &SampleDistinctOperator{SyntaxNode: node}
	case SyntaxKindEntityGroupExpression:
		return &EntityGroup{SyntaxNode: node}
	case SyntaxKindReduceByOperator:
		return &ReduceByOperator{SyntaxNode: node}
	case SyntaxKindReduceByWithClause:
		return &ReduceByWithClause{SyntaxNode: node}
	case SyntaxKindSummarizeOperator:
		return &SummarizeOperator{SyntaxNode: node}
	case SyntaxKindMacroExpandScopeReferenceName:
		return &MacroExpandScopeReferenceName{SyntaxNode: node}
	case SyntaxKindMacroExpandOperator:
		return &MacroExpandOperator{SyntaxNode: node}
	case SyntaxKindSummarizeByClause:
		return &SummarizeByClause{SyntaxNode: node}
	case SyntaxKindDistinctOperator:
		return &DistinctOperator{SyntaxNode: node}
	case SyntaxKindTakeOperator:
		return &TakeOperator{SyntaxNode: node}
	case SyntaxKindSortOperator:
		return &SortOperator{SyntaxNode: node}
	case SyntaxKindOrderedExpression:
		return &OrderedExpression{SyntaxNode: node}
	case SyntaxKindOrderingClause:
		return &OrderingClause{SyntaxNode: node}
	case SyntaxKindOrderingNullsClause:
		return &OrderingNullsClause{SyntaxNode: node}
	case SyntaxKindTopHittersOperator:
		return &TopHittersOperator{SyntaxNode: node}
	case SyntaxKindTopHittersByClause:
		return &TopHittersByClause{SyntaxNode: node}
	case SyntaxKindTopOperator:
		return &TopOperator{SyntaxNode: node}
	case SyntaxKindTopNestedOperator:
		return &TopNestedOperator{SyntaxNode: node}
	case SyntaxKindTopNestedClause:
		return &TopNestedClause{SyntaxNode: node}
	case SyntaxKindTopNestedWithOthersClause:
		return &TopNestedWithOthersClause{SyntaxNode: node}
	case SyntaxKindUnionOperator:
		return &UnionOperator{SyntaxNode: node}
	case SyntaxKindAsOperator:
		return &AsOperator{SyntaxNode: node}
	case SyntaxKindSerializeOperator:
		return &SerializeOperator{SyntaxNode: node}
	case SyntaxKindInvokeOperator:
		return &InvokeOperator{SyntaxNode: node}
	case SyntaxKindRenderOperator:
		return &RenderOperator{SyntaxNode: node}
	case SyntaxKindMakeGraphOperator:
		return &MakeGraphOperator{SyntaxNode: node}
	case SyntaxKindMakeGraphWithTablesAndKeysClause:
		return &MakeGraphWithTablesAndKeysClause{SyntaxNode: node}
	case SyntaxKindMakeGraphWithImplicitIdClause:
		return &MakeGraphWithImplicitIdClause{SyntaxNode: node}
	case SyntaxKindGraphMarkComponentsOperator:
		return &GraphMarkComponentsOperator{SyntaxNode: node}
	case SyntaxKindGraphWhereNodesOperator:
		return &GraphWhereNodesOperator{SyntaxNode: node}
	case SyntaxKindGraphWhereEdgesOperator:
		return &GraphWhereEdgesOperator{SyntaxNode: node}
	case SyntaxKindMakeGraphTableAndKeyClause:
		return &MakeGraphTableAndKeyClause{SyntaxNode: node}
	case SyntaxKindMakeGraphPartitionedByClause:
		return &MakeGraphPartitionedByClause{SyntaxNode: node}
	case SyntaxKindGraphToTableOperator:
		return &GraphToTableOperator{SyntaxNode: node}
	case SyntaxKindGraphToTableOutputClause:
		return &GraphToTableOutputClause{SyntaxNode: node}
	case SyntaxKindGraphToTableAsClause:
		return &GraphToTableAsClause{SyntaxNode: node}
	case SyntaxKindGraphMatchOperator:
		return &GraphMatchOperator{SyntaxNode: node}
	case SyntaxKindGraphShortestPathsOperator:
		return &GraphShortestPathsOperator{SyntaxNode: node}
	case SyntaxKindGraphMatchPattern:
		return &GraphMatchPattern{SyntaxNode: node}
	case SyntaxKindGraphMatchPatternNode:
		return &GraphMatchPatternNode{SyntaxNode: node}
	case SyntaxKindGraphMatchPatternEdge:
		return &GraphMatchPatternEdge{SyntaxNode: node}
	case SyntaxKindGraphMatchPatternEdgeRange:
		return &GraphMatchPatternEdgeRange{SyntaxNode: node}
	case SyntaxKindWhereClause:
		return &WhereClause{SyntaxNode: node}
	case SyntaxKindProjectClause:
		return &ProjectClause{SyntaxNode: node}
	case SyntaxKindNameReferenceList:
		return &NameReferenceList{SyntaxNode: node}
	case SyntaxKindRenderWithClause:
		return &RenderWithClause{SyntaxNode: node}
	case SyntaxKindPrintOperator:
		return &PrintOperator{SyntaxNode: node}
	case SyntaxKindAliasStatement:
		return &AliasStatement{SyntaxNode: node}
	case SyntaxKindLetStatement:
		return &LetStatement{SyntaxNode: node}
	case SyntaxKindFunctionDeclaration:
		return &FunctionDeclaration{SyntaxNode: node}
	case SyntaxKindFunctionParameters:
		return &FunctionParameters{SyntaxNode: node}
	case SyntaxKindFunctionParameter:
		return &FunctionParameter{SyntaxNode: node}
	case SyntaxKindDefaultValueDeclaration:
		return &DefaultValueDeclaration{SyntaxNode: node}
	case SyntaxKindFunctionBody:
		return &FunctionBody{SyntaxNode: node}
	case SyntaxKindSchemaTypeExpression:
		return &SchemaTypeExpression{SyntaxNode: node}
	case SyntaxKindExpressionStatement:
		return &ExpressionStatement{SyntaxNode: node}
	case SyntaxKindMaterializeExpression:
		return &MaterializeExpression{SyntaxNode: node}
	case SyntaxKindSetOptionStatement:
		return &SetOptionStatement{SyntaxNode: node}
	case SyntaxKindOptionValueClause:
		return &OptionValueClause{SyntaxNode: node}
	case SyntaxKindQueryParametersStatement:
		return &QueryParametersStatement{SyntaxNode: node}
	case SyntaxKindRestrictStatementWithClause:
		return &RestrictStatementWithClause{SyntaxNode: node}
	case SyntaxKindRestrictStatement:
		return &RestrictStatement{SyntaxNode: node}
	case SyntaxKindPatternStatement:
		return &PatternStatement{SyntaxNode: node}
	case SyntaxKindPatternDeclaration:
		return &PatternDeclaration{SyntaxNode: node}
	case SyntaxKindPatternPathParameter:
		return &PatternPathParameter{SyntaxNode: node}
	case SyntaxKindPatternMatch:
		return &PatternMatch{SyntaxNode: node}
	case SyntaxKindPatternPathValue:
		return &PatternPathValue{SyntaxNode: node}
	case SyntaxKindDataScopeExpression:
		return &DataScopeExpression{SyntaxNode: node}
	case SyntaxKindDataTableExpression:
		return &DataTableExpression{SyntaxNode: node}
	case SyntaxKindRowSchema:
		return &RowSchema{SyntaxNode: node}
	case SyntaxKindEvaluateRowSchema:
		return &EvaluateRowSchema{SyntaxNode: node}
	case SyntaxKindExternalDataExpression:
		return &ExternalDataExpression{SyntaxNode: node}
	case SyntaxKindContextualDataTableExpression:
		return &ContextualDataTableExpression{SyntaxNode: node}
	case SyntaxKindExternalDataWithClause:
		return &ExternalDataWithClause{SyntaxNode: node}
	case SyntaxKindInlineExternalTableKindClause:
		return &InlineExternalTableKindClause{SyntaxNode: node}
	case SyntaxKindInlineExternalTableDataFormatClause:
		return &InlineExternalTableDataFormatClause{SyntaxNode: node}
	case SyntaxKindDateTimePattern:
		return &DateTimePattern{SyntaxNode: node}
	case SyntaxKindInlineExternalTablePathFormatPartitionColumnReference:
		return &InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node}
	case SyntaxKindInlineExternalTablePathFormatClause:
		return &InlineExternalTablePathFormatClause{SyntaxNode: node}
	case SyntaxKindPartitionColumnDeclaration:
		return &PartitionColumnDeclaration{SyntaxNode: node}
	case SyntaxKindInlineExternalTablePartitionClause:
		return &InlineExternalTablePartitionClause{SyntaxNode: node}
	case SyntaxKindInlineExternalTableConnectionStringsClause:
		return &InlineExternalTableConnectionStringsClause{SyntaxNode: node}
	case SyntaxKindInlineExternalTableExpression:
		return &InlineExternalTableExpression{SyntaxNode: node}
	case SyntaxKindJoinOperator:
		return &JoinOperator{SyntaxNode: node}
	case SyntaxKindLookupOperator:
		return &LookupOperator{SyntaxNode: node}
	case SyntaxKindJoinOnClause:
		return &JoinOnClause{SyntaxNode: node}
	case SyntaxKindJoinWhereClause:
		return &JoinWhereClause{SyntaxNode: node}
	case SyntaxKindScanOperator:
		return &ScanOperator{SyntaxNode: node}
	case SyntaxKindAssertSchemaOperator:
		return &AssertSchemaOperator{SyntaxNode: node}
	case SyntaxKindScanDeclareClause:
		return &ScanDeclareClause{SyntaxNode: node}
	case SyntaxKindScanOrderByClause:
		return &ScanOrderByClause{SyntaxNode: node}
	case SyntaxKindScanPartitionByClause:
		return &ScanPartitionByClause{SyntaxNode: node}
	case SyntaxKindScanStepOutput:
		return &ScanStepOutput{SyntaxNode: node}
	case SyntaxKindScanStep:
		return &ScanStep{SyntaxNode: node}
	case SyntaxKindScanComputationClause:
		return &ScanComputationClause{SyntaxNode: node}
	case SyntaxKindScanAssignment:
		return &ScanAssignment{SyntaxNode: node}
	case SyntaxKindPartitionByOperator:
		return &PartitionByOperator{SyntaxNode: node}
	case SyntaxKindPartitionByIdClause:
		return &PartitionByIdClause{SyntaxNode: node}
	case SyntaxKindCommandWithValueClause:
		return &CommandWithValueClause{SyntaxNode: node}
	case SyntaxKindCommandWithPropertyListClause:
		return &CommandWithPropertyListClause{SyntaxNode: node}
	case SyntaxKindUnknownCommand:
		return &UnknownCommand{SyntaxNode: node}
	case SyntaxKindCustomCommand:
		return &CustomCommand{SyntaxNode: node}
	case SyntaxKindPartialCommand:
		return &PartialCommand{SyntaxNode: node}
	case SyntaxKindCommandAndSkippedTokens:
		return &CommandAndSkippedTokens{SyntaxNode: node}
	case SyntaxKindBadCommand:
		return &BadCommand{SyntaxNode: node}
	case SyntaxKindCommandBlock:
		return &CommandBlock{SyntaxNode: node}
	case SyntaxKindGraphMergeOperator:
		return &GraphMergeOperator{SyntaxNode: node}
	default:
		return node
	}
}
