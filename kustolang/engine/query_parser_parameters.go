package kustolang

// Ported from the query-operator-parameter region of
// src/Kusto.Language/Parser/QueryParser.cs.
// This is an initial native slice for known name=value parameters used by the
// currently ported query operators; this now includes the comma-list property
// bag form used by render with(...). More specialized operator parameter
// shapes remain pending with the larger operator grammar.

type queryOperatorParameterNameKind int

const (
	queryOperatorParameterNameDeclaredOrKnown queryOperatorParameterNameKind = iota
	queryOperatorParameterNameKnownOnly
	queryOperatorParameterNameDeclaredOnly
)

func (p *queryParserState) parseQueryOperatorParameterList(parameters []*QueryOperatorParameter, equalsNeeded bool) *SyntaxListOf[*NamedParameter] {
	return p.parseQueryOperatorParameterListWithEnd(parameters, equalsNeeded, nil)
}

func (p *queryParserState) parseQueryOperatorParameterListWithEnd(parameters []*QueryOperatorParameter, equalsNeeded bool, fnEndValue func() bool) *SyntaxListOf[*NamedParameter] {
	var elements []SyntaxElement
	for {
		parameter := p.scanQueryOperatorParameter(parameters, equalsNeeded)
		if parameter == nil {
			break
		}
		named := p.parseQueryOperatorParameterWithEnd(parameter, fnEndValue)
		if named == nil {
			break
		}
		elements = append(elements, named)
	}
	return AsSyntaxListOf[*NamedParameter](NewTypedSyntaxList("NamedParameter", elements...))
}

func (p *queryParserState) parseQueryOperatorParameterCommaList(parameters []*QueryOperatorParameter, namesAllowed queryOperatorParameterNameKind) *SeparatedSyntaxListOf[*NamedParameter] {
	endNameList := func() bool {
		return p.scanQueryOperatorParameterNameListEnd(parameters)
	}
	parseElement := func() *NamedParameter {
		return p.parseQueryOperatorParameterWithNameKind(parameters, namesAllowed, false, endNameList)
	}
	return parseCommaList(p, "NamedParameter", parseElement, createMissingNamedParameter, p.atCommonListEnd, false, false)
}

func (p *queryParserState) scanQueryOperatorParameter(parameters []*QueryOperatorParameter, equalsNeeded bool) *QueryOperatorParameter {
	_, nameLen, parameter := p.scanQueryOperatorParameterNameWithKind(parameters, queryOperatorParameterNameDeclaredOrKnown, equalsNeeded, 0)
	if nameLen == 0 {
		return nil
	}
	return parameter
}

func (p *queryParserState) atQueryOperatorParameterMissingEquals(parameters []*QueryOperatorParameter) bool {
	name, nameLen := p.scanQueryOperatorParameterName(0)
	if nameLen == 0 || findSpecificQueryOperatorParameter(name, parameters) == nil {
		return false
	}
	return p.peekToken(nameLen).Kind != SyntaxKindEqualToken
}

func (p *queryParserState) scanQueryOperatorParameterNameWithKind(parameters []*QueryOperatorParameter, namesAllowed queryOperatorParameterNameKind, equalsNeeded bool, offset int) (string, int, *QueryOperatorParameter) {
	name, nameLen := p.scanQueryOperatorParameterName(offset)
	if name == "" {
		return "", 0, nil
	}
	var parameter *QueryOperatorParameter
	if namesAllowed == queryOperatorParameterNameDeclaredOnly || namesAllowed == queryOperatorParameterNameDeclaredOrKnown {
		parameter = findSpecificQueryOperatorParameter(name, parameters)
	}
	if parameter == nil && namesAllowed == queryOperatorParameterNameKnownOnly {
		parameter = findSpecificQueryOperatorParameter(name, parameters)
	}
	if parameter == nil && (namesAllowed == queryOperatorParameterNameKnownOnly || namesAllowed == queryOperatorParameterNameDeclaredOrKnown) {
		parameter = findGloballyKnownQueryOperatorParameter(name)
	}
	if parameter == nil {
		return "", 0, nil
	}
	if parameter.HasNoEquals() || !equalsNeeded || p.peekToken(offset+nameLen).Kind == SyntaxKindEqualToken {
		return name, nameLen, parameter
	}
	return "", 0, nil
}

func (p *queryParserState) parseSpecificQueryOperatorParameter(parameter *QueryOperatorParameter, equalsNeeded bool) *NamedParameter {
	if parameter == nil {
		return nil
	}
	name, nameLen := p.scanQueryOperatorParameterName(0)
	if !queryOperatorParameterMatches(parameter, name) {
		return nil
	}
	nextKind := p.peekToken(nameLen).Kind
	if !parameter.HasNoEquals() && equalsNeeded && nextKind != SyntaxKindEqualToken {
		return nil
	}
	return p.parseQueryOperatorParameter(parameter)
}

func (p *queryParserState) parseSpecificQueryOperatorParameterWithEnd(parameters []*QueryOperatorParameter, equalsNeeded bool, fnEndNameList func() bool) *NamedParameter {
	parameter := p.scanSpecificQueryOperatorParameter(parameters, equalsNeeded, 0)
	if parameter == nil {
		return nil
	}
	return p.parseQueryOperatorParameterWithEnd(parameter, fnEndNameList)
}

func (p *queryParserState) scanSpecificQueryOperatorParameter(parameters []*QueryOperatorParameter, equalsNeeded bool, offset int) *QueryOperatorParameter {
	name, nameLen := p.scanQueryOperatorParameterName(offset)
	if name == "" {
		return nil
	}
	parameter := findSpecificQueryOperatorParameter(name, parameters)
	if parameter == nil {
		return nil
	}
	if parameter.HasNoEquals() || !equalsNeeded || p.peekToken(offset+nameLen).Kind == SyntaxKindEqualToken {
		return parameter
	}
	return nil
}

func (p *queryParserState) scanQueryOperatorParameterName(offset int) (string, int) {
	token := p.peekToken(offset)
	if token.Kind == SyntaxKindNone || token.Kind == SyntaxKindEndOfTextToken || token.Kind == SyntaxKindBarToken {
		return "", 0
	}
	if !canScanQueryOperatorParameterNamePart(token.Kind) {
		return "", 0
	}
	text := token.Text
	if p.peekToken(offset+1).Kind == SyntaxKindDotToken && canScanQueryOperatorParameterNamePart(p.peekToken(offset+2).Kind) {
		return text + "." + p.peekToken(offset+2).Text, 3
	}
	return text, 1
}

func canScanQueryOperatorParameterNamePart(kind SyntaxKind) bool {
	return kind == SyntaxKindIdentifierToken || IsKeywordKind(kind)
}

func findQueryOperatorParameter(name string, parameters []*QueryOperatorParameter) *QueryOperatorParameter {
	if parameter := findSpecificQueryOperatorParameter(name, parameters); parameter != nil {
		return parameter
	}
	return findGloballyKnownQueryOperatorParameter(name)
}

func findGloballyKnownQueryOperatorParameter(name string) *QueryOperatorParameter {
	for _, parameter := range QueryOperatorParameters.AllParameters {
		// These names belong only to graph-to-table output clauses. Keeping the
		// parameter objects in AllParameters is useful to catalog consumers, but
		// the native parser must not treat them as globally valid recovery names.
		if parameter.SymbolName() == QueryOperatorParameters.WithSourceId.SymbolName() ||
			parameter.SymbolName() == QueryOperatorParameters.WithTargetId.SymbolName() ||
			parameter.SymbolName() == QueryOperatorParameters.WithNodeId.SymbolName() {
			continue
		}
		if queryOperatorParameterMatches(parameter, name) {
			return parameter
		}
	}
	return nil
}

func findSpecificQueryOperatorParameter(name string, parameters []*QueryOperatorParameter) *QueryOperatorParameter {
	for _, parameter := range parameters {
		if queryOperatorParameterMatches(parameter, name) {
			return parameter
		}
	}
	return nil
}

func queryOperatorParameterMatches(parameter *QueryOperatorParameter, name string) bool {
	if parameter == nil {
		return false
	}
	if parameter.SymbolName() == name {
		return true
	}
	for _, alias := range parameter.Aliases() {
		if alias == name {
			return true
		}
	}
	return false
}

func (p *queryParserState) parseQueryOperatorParameter(parameter *QueryOperatorParameter) *NamedParameter {
	return p.parseQueryOperatorParameterWithEnd(parameter, nil)
}

func (p *queryParserState) parseQueryOperatorParameterWithEnd(parameter *QueryOperatorParameter, fnEndNameList func() bool) *NamedParameter {
	if parameter == nil {
		return nil
	}
	nameToken := p.parseQueryOperatorParameterNameToken()
	if nameToken == nil {
		return nil
	}
	name := MustNewNameDeclarationTyped(MustNewTokenNameTyped(nameToken))
	equal := createOmittedToken(SyntaxKindEqualToken)
	missingEqual := false
	if !parameter.HasNoEquals() {
		if p.peekToken().Kind == SyntaxKindEqualToken {
			equal = p.parseToken(SyntaxKindEqualToken)
		} else {
			equal = p.parseRequiredToken(SyntaxKindEqualToken)
			missingEqual = true
		}
	}
	value := p.parseQueryOperatorParameterValueWithMissingEqualRecovery(parameter, missingEqual, fnEndNameList)
	if isNilSyntaxElement(value) {
		value = p.createMissingValue()
	}
	return MustNewNamedParameterTyped(name, equal, value, queryOperatorParameterExpressionHint(parameter))
}

func (p *queryParserState) parseQueryOperatorParameterWithNameKind(parameters []*QueryOperatorParameter, namesAllowed queryOperatorParameterNameKind, equalsNeeded bool, fnEndNameList func() bool) *NamedParameter {
	nameText, nameLen, parameter := p.scanQueryOperatorParameterNameWithKind(parameters, namesAllowed, equalsNeeded, 0)
	if nameText == "" {
		return nil
	}
	nameToken := p.parseQueryOperatorParameterNameTokenWithText(nameText, nameLen)
	if nameToken == nil {
		return nil
	}
	name := MustNewNameDeclarationTyped(MustNewTokenNameTyped(nameToken))
	equal := p.parseRequiredToken(SyntaxKindEqualToken)
	var value Expression
	hint := CompletionHintNone
	if parameter != nil {
		value = p.parseQueryOperatorParameterValueWithEnd(parameter, fnEndNameList)
		hint = queryOperatorParameterExpressionHint(parameter)
	} else {
		value = p.parseAnyQueryOperatorParameterValue()
	}
	if isNilSyntaxElement(value) {
		value = p.createMissingValue()
	}
	return MustNewNamedParameterTyped(name, equal, value, hint)
}

func (p *queryParserState) parseQueryOperatorParameterNameToken() *SyntaxToken {
	name, nameLen := p.scanQueryOperatorParameterName(0)
	if name == "" {
		return nil
	}
	return p.parseQueryOperatorParameterNameTokenWithText(name, nameLen)
}

func (p *queryParserState) parseQueryOperatorParameterNameTokenWithText(name string, nameLen int) *SyntaxToken {
	if name == "" || nameLen <= 0 {
		return nil
	}
	first := p.peekToken()
	p.pos += nameLen
	return SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, first.Trivia, name))
}

func (p *queryParserState) parseQueryOperatorParameterValue(parameter *QueryOperatorParameter) Expression {
	return p.parseQueryOperatorParameterValueWithEnd(parameter, nil)
}

func (p *queryParserState) parseQueryOperatorParameterValueWithEnd(parameter *QueryOperatorParameter, fnEndNameList func() bool) Expression {
	if parameter == nil {
		return p.parseAnyQueryOperatorParameterValue()
	}
	switch parameter.ValueKind() {
	case QueryOperatorParameterValueAny:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return p.createMissingValue()
	case QueryOperatorParameterValueStringLiteral:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return createMissingStringLiteral()
	case QueryOperatorParameterValueBoolLiteral:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return createMissingBoolLiteral()
	case QueryOperatorParameterValueIntegerLiteral, QueryOperatorParameterValueNumericLiteral, QueryOperatorParameterValueSummableLiteral:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return createMissingLongLiteral()
	case QueryOperatorParameterValueForcedRealLiteral:
		if expr := p.parseAnyQueryOperatorParameterForcedRealValue(); expr != nil {
			return expr
		}
		return createMissingRealLiteral()
	case QueryOperatorParameterValueScalarLiteral:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return p.createMissingValue()
	case QueryOperatorParameterValueColumn:
		if expr := p.parseNameReference(); expr != nil {
			return expr
		}
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return p.createMissingNameReference()
	case QueryOperatorParameterValueColumnList:
		if fnEndNameList == nil {
			fnEndNameList = p.scanDefaultQueryOperatorParameterNameListEnd
		}
		return p.parseNameReferenceList(fnEndNameList)
	case QueryOperatorParameterValueNameDeclaration:
		if name := p.parseNameDeclaration(); name != nil {
			return name
		}
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return p.createMissingNameDeclaration()
	case QueryOperatorParameterValueString:
		if expr := p.parseFunctionCallOrPath(); expr != nil {
			return expr
		}
		return p.createMissingValue()
	case QueryOperatorParameterValueWord, QueryOperatorParameterValueWordOrNumber:
		if len(parameter.Values()) > 0 {
			if expr := p.parseTokenLiteral(parameter.Values(), parameter.IsCaseSensitive()); expr != nil {
				return expr
			}
		}
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		if len(parameter.Values()) > 0 {
			return createMissingTokenLiteral(parameter.Values()...)
		}
		return p.createMissingValue()
	default:
		if expr := p.parseAnyQueryOperatorParameterValue(); expr != nil {
			return expr
		}
		return p.createMissingValue()
	}
}

func (p *queryParserState) parseQueryOperatorParameterValueWithMissingEqualRecovery(parameter *QueryOperatorParameter, missingEqual bool, fnEndNameList func() bool) Expression {
	if !missingEqual || parameter == nil {
		return p.parseQueryOperatorParameterValueWithEnd(parameter, fnEndNameList)
	}
	switch parameter.ValueKind() {
	case QueryOperatorParameterValueWord, QueryOperatorParameterValueWordOrNumber:
		if expr := p.parseIdentifierOrKeywordTokenLiteral(); expr != nil {
			return expr
		}
	case QueryOperatorParameterValueColumn, QueryOperatorParameterValueNameDeclaration, QueryOperatorParameterValueString:
		if expr := p.parseNameReference(); expr != nil {
			return expr
		}
	}
	return p.parseQueryOperatorParameterValueWithEnd(parameter, fnEndNameList)
}

func (p *queryParserState) scanQueryOperatorParameterNameListEnd(parameters []*QueryOperatorParameter) bool {
	if p.peekToken().Kind == SyntaxKindCommaToken {
		return p.scanKnownQueryOperatorParameter(parameters, false, 1) || p.scanCommonListEnd(1)
	}
	return p.scanKnownQueryOperatorParameter(parameters, true, 0) || p.scanCommonListEnd()
}

func (p *queryParserState) scanDefaultQueryOperatorParameterNameListEnd() bool {
	return p.scanQueryOperatorParameterNameListEnd(nil)
}

func (p *queryParserState) scanKnownQueryOperatorParameter(parameters []*QueryOperatorParameter, equalsNeeded bool, offset int) bool {
	if p.scanSpecificQueryOperatorParameter(parameters, equalsNeeded, offset) != nil {
		return true
	}
	name, nameLen := p.scanQueryOperatorParameterName(offset)
	if name == "" {
		return false
	}
	parameter := findGloballyKnownQueryOperatorParameter(name)
	if parameter == nil {
		return false
	}
	return parameter.HasNoEquals() || !equalsNeeded || p.peekToken(offset+nameLen).Kind == SyntaxKindEqualToken
}

func (p *queryParserState) parseNameReferenceList(fnEndList func() bool) *NameReferenceList {
	names := p.parseNameReferenceSeparatedList(fnEndList)
	return MustNewNameReferenceListTyped(names)
}

func (p *queryParserState) parseNameReferenceSeparatedList(fnEndList func() bool) *SeparatedSyntaxListOf[*NameReference] {
	return parseCommaList(
		p,
		"NameReference",
		func() *NameReference { return p.parseNameReferenceListName(fnEndList) },
		createMissingNameReferenceNode,
		fnEndList,
		true,
		false,
	)
}

func (p *queryParserState) parseNameReferenceListName(fnEndList func() bool) *NameReference {
	if fnEndList != nil && fnEndList() {
		return nil
	}
	name := p.parseExtendedName()
	if name == nil {
		return nil
	}
	return MustNewNameReferenceTyped(name, SymbolMatchNone)
}

func (p *queryParserState) parseAnyQueryOperatorParameterValue() Expression {
	if expr := p.parseLiteral(); expr != nil {
		return expr
	}
	if expr := p.parseIdentifierOrKeywordTokenLiteral(); expr != nil {
		return expr
	}
	if expr := p.parseNameReference(); expr != nil {
		return expr
	}
	return nil
}

func (p *queryParserState) parseAnyQueryOperatorParameterForcedRealValue() Expression {
	if expr := p.parseForcedRealLiteral(); expr != nil {
		return expr
	}
	return p.parseAnyQueryOperatorParameterValue()
}

func (p *queryParserState) parseIdentifierOrKeywordTokenLiteral() Expression {
	kind := p.peekToken().Kind
	if kind == SyntaxKindIdentifierToken || (IsKeywordKind(kind) && p.canBeIdentifier(kind)) {
		return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, p.parseToken())
	}
	return nil
}

func (p *queryParserState) parseTokenLiteral(values []string, caseSensitive bool) Expression {
	token := p.peekToken()
	for _, value := range values {
		if (caseSensitive && token.Text == value) || (!caseSensitive && ordinalEqualFold(token.Text, value)) {
			return MustNewLiteralExpressionTyped(SyntaxKindTokenLiteralExpression, p.parseToken())
		}
	}
	return nil
}

func (p *queryParserState) createMissingValue() Expression {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingValue())
}

func createMissingNamedParameter() *NamedParameter {
	return MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken))),
		createOmittedToken(SyntaxKindEqualToken),
		MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone),
		CompletionHintNone,
		DiagnosticMissingParameter(),
	)
}

func createMissingValueExpression() Expression {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingValue())
}

func createMissingNameReferenceNode() *NameReference {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(createOmittedToken(SyntaxKindIdentifierToken)), SymbolMatchNone, DiagnosticMissingName())
}

func queryOperatorParameterExpressionHint(parameter *QueryOperatorParameter) CompletionHint {
	if parameter == nil {
		return CompletionHintNone
	}
	switch parameter.ValueKind() {
	case QueryOperatorParameterValueColumn, QueryOperatorParameterValueColumnList:
		return CompletionHintColumn
	default:
		return CompletionHintNone
	}
}
