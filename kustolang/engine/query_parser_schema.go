package kustolang

import "fmt"

// Ported from the row-schema parsing path in
// src/Kusto.Language/Parser/QueryParser.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type queryParserState struct {
	tokens         []LexicalToken
	pos            int
	options        ParseOptions
	commandGrammar *CommandGrammar
	cancellation   CancellationToken
	work           uint64
}

var noLexicalToken = NewLexicalToken(SyntaxKindNone, "", "")

func ParseRowSchema(text string) *RowSchema {
	return ParseRowSchemaWithOptions(text, DefaultParseOptions)
}

func ParseRowSchemaWithOptions(text string, options ParseOptions) *RowSchema {
	return newQueryParserStateWithOwnedTokens(LexWithOptions(text, options), options, CancellationToken{}).parseRowSchema(false)
}

func ParseRowSchemaTokens(tokens []LexicalToken) *RowSchema {
	return ParseRowSchemaTokensFromWithOptions(tokens, 0, DefaultParseOptions)
}

func ParseRowSchemaTokensWithOptions(tokens []LexicalToken, options ParseOptions) *RowSchema {
	return ParseRowSchemaTokensFromWithOptions(tokens, 0, options)
}

func ParseRowSchemaTokensFrom(tokens []LexicalToken, start int) *RowSchema {
	return ParseRowSchemaTokensFromWithOptions(tokens, start, DefaultParseOptions)
}

func ParseRowSchemaTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *RowSchema {
	return newQueryParserStateWithOptions(lexicalTokensFrom(tokens, start), options).parseRowSchema(false)
}

func lexicalTokensFrom(tokens []LexicalToken, start int) []LexicalToken {
	if start <= 0 {
		return tokens
	}
	if start >= len(tokens) {
		return nil
	}
	return tokens[start:]
}

func newQueryParserState(tokens []LexicalToken) *queryParserState {
	return newQueryParserStateWithOptions(tokens, DefaultParseOptions)
}

func newQueryParserStateWithOptions(tokens []LexicalToken, options ParseOptions) *queryParserState {
	return newQueryParserStateWithOptionsAndCancellation(tokens, options, CancellationToken{})
}

func newQueryParserStateWithOptionsAndCancellation(tokens []LexicalToken, options ParseOptions, cancellationToken CancellationToken) *queryParserState {
	cancellationToken.ThrowIfCancellationRequested()
	return newQueryParserStateWithOwnedTokens(copyLexicalTokensWithCancellation(tokens, cancellationToken), options, cancellationToken)
}

// The caller transfers a freshly lexed or already copied slice. Public APIs
// accepting caller-owned tokens must continue through the copying constructor.
func newQueryParserStateWithOwnedTokens(tokens []LexicalToken, options ParseOptions, cancellationToken CancellationToken) *queryParserState {
	cancellationToken.ThrowIfCancellationRequested()
	return &queryParserState{tokens: tokens, options: options, cancellation: cancellationToken}
}

func copyLexicalTokensWithCancellation(tokens []LexicalToken, cancellationToken CancellationToken) []LexicalToken {
	copyOfTokens := make([]LexicalToken, len(tokens))
	for i := range tokens {
		if i&255 == 0 {
			cancellationToken.ThrowIfCancellationRequested()
		}
		copyOfTokens[i] = tokens[i]
	}
	cancellationToken.ThrowIfCancellationRequested()
	return copyOfTokens
}

func (p *queryParserState) canBeIdentifier(kind SyntaxKind) bool {
	return SyntaxKindCanBeIdentifier(kind)
}

func (p *queryParserState) peekToken(offset ...int) LexicalToken {
	p.work++
	if p.work&255 == 0 {
		p.cancellation.ThrowIfCancellationRequested()
	}
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	index := p.pos + delta
	if index < 0 || index >= len(p.tokens) {
		return noLexicalToken
	}
	return p.tokens[index]
}

func (p *queryParserState) tokenTextStart(index int) int {
	if index < 0 {
		return 0
	}
	start := 0
	for i := 0; i < index && i < len(p.tokens); i++ {
		p.work++
		if p.work&255 == 0 {
			p.cancellation.ThrowIfCancellationRequested()
		}
		start += p.tokens[i].Length()
	}
	if index < len(p.tokens) {
		start += len(p.tokens[index].Trivia)
	}
	return start
}

func (p *queryParserState) parseToken(kinds ...SyntaxKind) *SyntaxToken {
	token := p.peekToken()
	if token.Kind == SyntaxKindNone {
		return nil
	}
	if len(kinds) > 0 {
		matched := false
		for _, kind := range kinds {
			if token.Kind == kind {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
	}
	p.pos++
	return SyntaxTokenFromLexical(token)
}

func (p *queryParserState) parseTokenTextAsKind(text string, kind SyntaxKind) *SyntaxToken {
	source := NewArraySource(p.tokens)
	length := MatchesLexicalText(source, p.pos, text)
	if length <= 0 {
		return nil
	}
	token := ProduceSyntaxTokenFromText(source, p.pos, length, text, kind)
	p.pos += length
	return token
}

func (p *queryParserState) parseFirstTokenTextAsKind(alternatives ...struct {
	text string
	kind SyntaxKind
}) *SyntaxToken {
	for _, alternative := range alternatives {
		if token := p.parseTokenTextAsKind(alternative.text, alternative.kind); token != nil {
			return token
		}
	}
	return nil
}

func (p *queryParserState) parseCombinedToken(length int) *SyntaxToken {
	if length <= 0 {
		return nil
	}
	first := p.peekToken()
	if first.Kind == SyntaxKindNone {
		return nil
	}
	text := first.Text
	for i := 1; i < length; i++ {
		next := p.peekToken(i)
		if next.Kind == SyntaxKindNone {
			return nil
		}
		text += next.Trivia + next.Text
	}
	p.pos += length
	return SyntaxTokenFromLexical(NewLexicalToken(first.Kind, first.Trivia, text, first.Diagnostics...))
}

func (p *queryParserState) parseRequiredToken(kind SyntaxKind) *SyntaxToken {
	if token := p.parseToken(kind); token != nil {
		return token
	}
	return createMissingToken(kind)
}

func (p *queryParserState) scanCommonListEnd(offset ...int) bool {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	switch p.peekToken(delta).Kind {
	case SyntaxKindCloseParenToken,
		SyntaxKindCloseBracketToken,
		SyntaxKindCloseBraceToken,
		SyntaxKindBarToken,
		SyntaxKindSemicolonToken,
		SyntaxKindEndOfTextToken,
		SyntaxKindNone:
		return true
	default:
		return false
	}
}

func (p *queryParserState) parseRowSchema(oneOrMore bool) *RowSchema {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseToken()
	leadingComma := p.parseToken(SyntaxKindCommaToken)
	columns := p.parseNameAndTypeDeclarationCommaList(oneOrMore)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewRowSchemaTyped(open, leadingComma, columns, close)
}

func createMissingRowSchema() *RowSchema {
	return createMissingRowSchemaWithDiagnostics()
}

func createMissingRowSchemaWithDiagnostics(diagnostics ...Diagnostic) *RowSchema {
	diagnostics = append([]Diagnostic{DiagnosticMissingSchemaDeclaration()}, diagnostics...)
	return MustNewRowSchemaTyped(
		createOmittedToken(SyntaxKindOpenParenToken),
		nil,
		AsSeparatedSyntaxListOf[*NameAndTypeDeclaration](NewTypedSyntaxList("SeparatedElement<NameAndTypeDeclaration>")),
		createOmittedToken(SyntaxKindCloseParenToken),
		diagnostics...,
	)
}

func (p *queryParserState) parseEvaluateRowSchema() *EvaluateRowSchema {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseToken()
	leadingComma := p.parseToken(SyntaxKindCommaToken)
	var asterisk *SyntaxToken
	var asteriskComma *SyntaxToken
	asterisk = p.parseToken(SyntaxKindAsteriskToken)
	asteriskComma = p.parseToken(SyntaxKindCommaToken)
	columns := p.parseNameAndTypeDeclarationCommaList(false)
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewEvaluateRowSchemaTyped(open, leadingComma, asterisk, asteriskComma, columns, close)
}

func (p *queryParserState) parseNameAndTypeDeclarationCommaList(oneOrMore bool) *SeparatedSyntaxListOf[*NameAndTypeDeclaration] {
	return p.parseNameAndTypeDeclarationCommaListWithTrailing(oneOrMore, true)
}

func (p *queryParserState) parseNameAndTypeDeclarationCommaListWithTrailing(oneOrMore bool, allowTrailingComma bool) *SeparatedSyntaxListOf[*NameAndTypeDeclaration] {
	return parseCommaList(p, "NameAndTypeDeclaration", p.parseNameAndTypeDeclaration, createMissingNameAndTypeDeclaration, p.atCommonListEnd, oneOrMore, allowTrailingComma)
}

func (p *queryParserState) parseNameAndTypeDeclaration() *NameAndTypeDeclaration {
	name := p.parseExtendedNameDeclaration()
	if name != nil {
		colon := p.parseRequiredToken(SyntaxKindColonToken)
		typ := p.parseParamType()
		if typ == nil {
			if schemaType := p.parseSchemaType(); schemaType != nil {
				typ = schemaType
			}
		}
		if typ == nil {
			typ = p.parseInvalidParamType()
		}
		if typ == nil {
			typ = createMissingType()
		}
		return MustNewNameAndTypeDeclarationTyped(name, colon, typ)
	}
	if p.peekToken().Kind == SyntaxKindColonToken {
		colon := p.parseToken()
		typ := p.parseParamType()
		if typ == nil {
			if schemaType := p.parseSchemaType(); schemaType != nil {
				typ = schemaType
			}
		}
		if typ == nil {
			typ = p.parseInvalidParamType()
		}
		if typ == nil {
			typ = createMissingType()
		}
		return MustNewNameAndTypeDeclarationTyped(p.createMissingNameDeclaration(), colon, typ)
	}
	return nil
}

func (p *queryParserState) parseSchemaType() *SchemaTypeExpression {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	if p.peekToken(1).Kind == SyntaxKindAsteriskToken {
		open := p.parseToken()
		star := MustNewStarExpressionTyped(p.parseToken())
		close := p.parseRequiredToken(SyntaxKindCloseParenToken)
		return MustNewSchemaTypeExpressionTyped(
			open,
			AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", NewSeparatedElement(star, nil))),
			close,
		)
	}
	return p.parseSchemaMultipartType()
}

func (p *queryParserState) parseSchemaMultipartType() *SchemaTypeExpression {
	if p.peekToken().Kind != SyntaxKindOpenParenToken {
		return nil
	}
	open := p.parseToken()
	columns := p.parseSchemaColumnDeclarationCommaList()
	close := p.parseRequiredToken(SyntaxKindCloseParenToken)
	return MustNewSchemaTypeExpressionTyped(open, columns, close)
}

func (p *queryParserState) parseSchemaColumnDeclarationCommaList() *SeparatedSyntaxListOf[Expression] {
	var elements []SyntaxElement
	var element Expression
	if !p.scanCommonListEnd() {
		element = p.parseNameAndTypeDeclaration()
	}
	if element == nil && p.peekToken().Kind == SyntaxKindCommaToken {
		element = createMissingNameAndTypeDeclaration()
	}
	for element != nil {
		if p.peekToken().Kind == SyntaxKindEndOfTextToken || p.peekToken().Kind == SyntaxKindNone || p.scanCommonListEnd() {
			elements = append(elements, NewSeparatedElement(element, nil))
			return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
		}
		if p.peekToken().Kind == SyntaxKindCommaToken {
			comma := p.parseToken()
			elements = append(elements, NewSeparatedElement(element, comma))
			if p.scanCommonListEnd() {
				break
			}
			next := p.parseNameAndTypeDeclaration()
			if next != nil {
				element = next
				continue
			}
			element = createMissingNameAndTypeDeclaration()
			continue
		}
		next := p.parseNameAndTypeDeclaration()
		if next != nil {
			elements = append(elements, NewSeparatedElement(element, createMissingToken(SyntaxKindCommaToken)))
			element = next
			continue
		}
		elements = append(elements, NewSeparatedElement(element, nil))
		return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
	}
	return AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>", elements...))
}

func (p *queryParserState) parseDefaultValueDeclaration() *DefaultValueDeclaration {
	equal := p.parseToken(SyntaxKindEqualToken)
	if equal == nil {
		return nil
	}
	expr := p.parseLiteral()
	if expr == nil {
		expr = p.parseIdentifierOrKeywordTokenLiteral()
	}
	if expr == nil {
		expr = p.createMissingExpression()
	}
	return MustNewDefaultValueDeclarationTyped(equal, expr)
}

func (p *queryParserState) parseFunctionParameter() *FunctionParameter {
	name := p.parseNameAndTypeDeclaration()
	if name == nil {
		return nil
	}
	return MustNewFunctionParameterTyped(name, p.parseDefaultValueDeclaration())
}

func createMissingFunctionParameter() *FunctionParameter {
	nameAndType := MustNewNameAndTypeDeclarationTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		createOmittedToken(SyntaxKindColonToken),
		MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
	)
	return MustNewFunctionParameterTyped(nameAndType, nil, DiagnosticMissingParameter())
}

func (p *queryParserState) parseRequiredFunctionParameterCommaList(atEnd func() bool) *SeparatedSyntaxListOf[*FunctionParameter] {
	list := p.parseFunctionParameterCommaList(atEnd)
	if list.Count() == 0 {
		return AsSeparatedSyntaxListOf[*FunctionParameter](NewTypedSyntaxList("SeparatedElement<FunctionParameter>", NewSeparatedElement(createMissingFunctionParameter(), nil)))
	}
	return list
}

func (p *queryParserState) parseFunctionParameterCommaList(atEnd func() bool) *SeparatedSyntaxListOf[*FunctionParameter] {
	if atEnd == nil {
		atEnd = p.atCommonListEnd
	}
	return parseCommaList(p, "FunctionParameter", p.parseFunctionParameter, createMissingFunctionParameter, atEnd, false, false)
}

func (p *queryParserState) parseNameAndOptionalTypeDeclaration() Expression {
	if p.peekToken().Kind == SyntaxKindColonToken {
		colon := p.parseToken()
		typ := p.parseParamType()
		if typ == nil {
			typ = p.parseInvalidParamType()
		}
		if typ == nil {
			typ = createMissingType()
		}
		return MustNewNameAndTypeDeclarationTyped(p.createMissingNameDeclaration(), colon, typ)
	}
	name := p.parseExtendedNameDeclaration()
	if name == nil {
		return nil
	}
	if p.peekToken().Kind != SyntaxKindColonToken {
		return name
	}
	colon := p.parseToken()
	typ := p.parseParamType()
	if typ == nil {
		typ = p.parseInvalidParamType()
	}
	if typ == nil {
		typ = createMissingType()
	}
	return MustNewNameAndTypeDeclarationTyped(name, colon, typ)
}

func (p *queryParserState) parseExtendedNameDeclaration() *NameDeclaration {
	name := p.parseExtendedName()
	if name == nil {
		return nil
	}
	return MustNewNameDeclarationTyped(name)
}

func (p *queryParserState) parseNameDeclaration() *NameDeclaration {
	name := p.parseName()
	if name == nil {
		return nil
	}
	return MustNewNameDeclarationTyped(name)
}

func (p *queryParserState) parseExtendedNameReference() *NameReference {
	name := p.parseExtendedName()
	if name == nil {
		return nil
	}
	return MustNewNameReferenceTyped(name, SymbolMatchNone)
}

func (p *queryParserState) parseExtendedName() Name {
	switch p.peekToken().Kind {
	case SyntaxKindOpenBracketToken:
		return p.parseBracketedName()
	case SyntaxKindOpenBraceToken:
		return p.parseBracedName()
	default:
		if name := p.parseIdentifierName(); name != nil {
			return name
		}
		return p.parseExtendedKeywordAsIdentifierName()
	}
}

func (p *queryParserState) parseIdentifierName() Name {
	token := p.parseKeywordOrIdentifier()
	if token == nil {
		return nil
	}
	return MustNewTokenNameTyped(token)
}

func (p *queryParserState) parseExtendedKeywordAsIdentifierName() Name {
	if !p.isExtendedKeywordAsIdentifier(p.peekToken().Kind) {
		return nil
	}
	return MustNewTokenNameTyped(p.parseToken())
}

func (p *queryParserState) parseKeywordOrIdentifier() *SyntaxToken {
	kind := p.peekToken().Kind
	if kind == SyntaxKindIdentifierToken || (IsKeywordKind(kind) && p.canBeIdentifier(kind)) {
		return p.parseToken()
	}
	return nil
}

func (p *queryParserState) parseBracketedName() Name {
	if p.scanBracketedName() <= 0 {
		return nil
	}
	open := p.parseToken()
	name := p.parseStringOrCompoundStringLiteral()
	close := p.parseRequiredToken(SyntaxKindCloseBracketToken)
	return MustNewBracketedNameTyped(open, name, close)
}

func (p *queryParserState) scanBracketedName(offset ...int) int {
	delta := 0
	if len(offset) > 0 {
		delta = offset[0]
	}
	if p.peekToken(delta).Kind != SyntaxKindOpenBracketToken {
		return -1
	}
	length := p.scanStringOrCompoundStringLiteral(delta + 1)
	if length <= 0 || p.peekToken(delta+1+length).Kind != SyntaxKindCloseBracketToken {
		return -1
	}
	return length + 2
}

func (p *queryParserState) parseBracedName() Name {
	length := p.scanBracedName()
	if length <= 2 {
		return nil
	}
	open := SyntaxTokenFromLexical(p.peekToken())
	text := ""
	for i := 1; i < length-1; i++ {
		text += p.peekToken(i).Text
	}
	name := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", text))
	close := SyntaxTokenFromLexical(p.peekToken(length - 1))
	p.pos += length
	return MustNewBracedNameTyped(open, name, close)
}

func (p *queryParserState) scanIdentifierOrKeywordAsIdentifier(offset int) bool {
	kind := p.peekToken(offset).Kind
	return kind == SyntaxKindIdentifierToken || (IsKeywordKind(kind) && p.canBeIdentifier(kind))
}

func (p *queryParserState) parseParamType() TypeExpression {
	if !isParamTypeKind(p.peekToken().Kind) {
		return nil
	}
	return MustNewPrimitiveTypeExpressionTyped(p.parseToken())
}

func (p *queryParserState) parseParamTypeExtended() TypeExpression {
	if !isExtendedParamTypeKind(p.peekToken().Kind) {
		return nil
	}
	return MustNewPrimitiveTypeExpressionTyped(p.parseToken())
}

func (p *queryParserState) parseIdentifierTypeExpression() TypeExpression {
	token := p.peekToken()
	if token.Kind == SyntaxKindIdentifierToken || (IsKeywordKind(token.Kind) && p.canBeIdentifier(token.Kind)) {
		return MustNewPrimitiveTypeExpressionTyped(p.parseToken())
	}
	return nil
}

func (p *queryParserState) parseInvalidParamType() TypeExpression {
	token := p.peekToken()
	if isExtendedParamTypeKind(token.Kind) || token.Kind == SyntaxKindIdentifierToken || (IsKeywordKind(token.Kind) && p.canBeIdentifier(token.Kind)) {
		return MustNewPrimitiveTypeExpressionTyped(p.parseToken(), DiagnosticInvalidTypeName(token.Text))
	}
	return nil
}

func createMissingNameDeclaration() *NameDeclaration {
	return MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), DiagnosticMissingName())
}

func (p *queryParserState) missingNameDiagnostic() Diagnostic {
	token := p.peekToken()
	if IsKeywordKind(token.Kind) {
		return DiagnosticMissingNameWithKeyword(token.Text)
	}
	return DiagnosticMissingName()
}

func (p *queryParserState) createMissingNameDeclaration() *NameDeclaration {
	return MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), p.missingNameDiagnostic())
}

func (p *queryParserState) createMissingNameReferenceNode() *NameReference {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, p.missingNameDiagnostic())
}

func createMissingNameAndTypeDeclaration() *NameAndTypeDeclaration {
	return MustNewNameAndTypeDeclarationTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		createOmittedToken(SyntaxKindColonToken),
		MustNewPrimitiveTypeExpressionTyped(createOmittedToken(SyntaxKindStringKeyword)),
		DiagnosticMissingParameter(),
	)
}

func createMissingType() TypeExpression {
	return MustNewPrimitiveTypeExpressionTyped(NewMissingToken(SyntaxKindIdentifierToken, Diagnostic{}), DiagnosticMissingTypeName())
}

func createMissingToken(kind SyntaxKind) *SyntaxToken {
	return CreateMissingSyntaxToken(kind)
}

func createExpectedToken(kind SyntaxKind) *SyntaxToken {
	return NewMissingToken(kind, DiagnosticTermsExpected(missingTokenText(kind)))
}

func createOmittedToken(kind SyntaxKind) *SyntaxToken {
	return NewMissingToken(kind, Diagnostic{})
}

func missingTokenText(kind SyntaxKind) string {
	if text := SyntaxText(kind); text != "" {
		return text
	}
	switch kind {
	case SyntaxKindIdentifierToken:
		return "identifier"
	case SyntaxKindStringLiteralToken:
		return "string literal"
	case SyntaxKindStringKeyword:
		return "string"
	default:
		return fmt.Sprintf("%d", kind)
	}
}

func isParamTypeKind(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindBoolKeyword,
		SyntaxKindBooleanKeyword,
		SyntaxKindDateKeyword,
		SyntaxKindDateTimeKeyword,
		SyntaxKindDecimalKeyword,
		SyntaxKindDoubleKeyword,
		SyntaxKindDynamicKeyword,
		SyntaxKindGuidKeyword,
		SyntaxKindIntKeyword,
		SyntaxKindInt64Keyword,
		SyntaxKindInt8Keyword,
		SyntaxKindLongKeyword,
		SyntaxKindRealKeyword,
		SyntaxKindStringKeyword,
		SyntaxKindTimeKeyword,
		SyntaxKindTimespanKeyword,
		SyntaxKindUniqueIdKeyword:
		return true
	default:
		return false
	}
}

func isExtendedParamTypeKind(kind SyntaxKind) bool {
	if isParamTypeKind(kind) {
		return true
	}
	switch kind {
	case SyntaxKindFloatKeyword,
		SyntaxKindInt16Keyword,
		SyntaxKindInt32Keyword,
		SyntaxKindUIntKeyword,
		SyntaxKindUInt16Keyword,
		SyntaxKindUInt32Keyword,
		SyntaxKindUInt64Keyword,
		SyntaxKindUInt8Keyword,
		SyntaxKindULongKeyword:
		return true
	default:
		return false
	}
}

func (p *queryParserState) isExtendedKeywordAsIdentifier(kind SyntaxKind) bool {
	if IsKeywordKind(kind) && p.canBeIdentifier(kind) {
		return true
	}
	return isExtendedSpecialKeywordAsIdentifier(kind)
}

func isExtendedKeywordAsIdentifier(kind SyntaxKind) bool {
	if IsKeywordKind(kind) && SyntaxKindCanBeIdentifier(kind) {
		return true
	}
	return isExtendedSpecialKeywordAsIdentifier(kind)
}

func isExtendedSpecialKeywordAsIdentifier(kind SyntaxKind) bool {
	switch kind {
	case SyntaxKindAccumulateKeyword,
		SyntaxKindAsKeyword,
		SyntaxKindByKeyword,
		SyntaxKindContainsKeyword,
		SyntaxKindConsumeKeyword,
		SyntaxKindCountKeyword,
		SyntaxKindDataTableKeyword,
		SyntaxKindDistinctKeyword,
		SyntaxKindEarliestKeyword,
		SyntaxKindExtendKeyword,
		SyntaxKindExternalDataKeyword,
		SyntaxKindInlineExternalTableKeyword,
		SyntaxKindDataFormatKeyword,
		SyntaxKindDateTimePatternKeyword,
		SyntaxKindFindKeyword,
		SyntaxKindFilterKeyword,
		SyntaxKindHasKeyword,
		SyntaxKindInKeyword,
		SyntaxKindInvokeKeyword,
		SyntaxKindLatestKeyword,
		SyntaxKindLimitKeyword,
		SyntaxKindMaterializeKeyword,
		SyntaxKindMdmKeyword,
		SyntaxKindOfKeyword,
		SyntaxKindParseKeyword,
		SyntaxKindPlotlyKeyword,
		SyntaxKindPrintKeyword,
		SyntaxKindSampleKeyword,
		SyntaxKindSampleDistinctKeyword,
		SyntaxKindScanKeyword,
		SyntaxKindSearchKeyword,
		SyntaxKindSerializeKeyword,
		SyntaxKindSetKeyword,
		SyntaxKindSqlKeyword,
		SyntaxKindSortKeyword,
		SyntaxKindSummarizeKeyword,
		SyntaxKindTakeKeyword,
		SyntaxKindTitleKeyword,
		SyntaxKindToKeyword,
		SyntaxKindTopKeyword,
		SyntaxKindToScalarKeyword,
		SyntaxKindToTableKeyword,
		SyntaxKindTopNestedKeyword,
		SyntaxKindTopHittersKeyword,
		SyntaxKindVerboseKeyword,
		SyntaxKindViewersKeyword,
		SyntaxKindWhereKeyword:
		return true
	default:
		return false
	}
}
