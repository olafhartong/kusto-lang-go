package kustolang

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Parser/KustoFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	MultiLineStringQuote          = "```"
	AlternateMultiLineStringQuote = "~~~"
)

func HasPipedInput(queryOperator QueryOperator) bool {
	if queryOperator == nil {
		return false
	}
	if pipe, ok := WrapGeneratedSyntaxElement(queryOperator.Parent()).(*PipeExpression); ok && pipe.Operator() == queryOperator {
		return true
	}
	return isChildOfPipeStartingExpression(queryOperator)
}

func isChildOfPipeStartingExpression(element SyntaxElement) bool {
	if element == nil || element.Parent() == nil {
		return false
	}
	parent := WrapGeneratedSyntaxElement(element.Parent())
	switch node := parent.(type) {
	case *ForkExpression:
		return node.Expression() == element
	case *PartitionSubquery:
		return node.Subquery() == element
	case *MvApplySubqueryExpression:
		return node.Expression() == element
	case *PartitionByOperator:
		return node.Subquery() == element
	case *MakeGraphPartitionedByClause:
		return node.Subquery() == element
	case *FacetWithExpressionClause:
		return node.Expression() == element
	case *FacetWithOperatorClause:
		return node.Operator() == element
	case *MaterializedViewCombineClause:
		combine, ok := WrapGeneratedSyntaxElement(node.Parent()).(*MaterializedViewCombineExpression)
		return ok && combine.AggregationsClause() == node
	}
	if _, ok := parent.(Expression); ok {
		return isChildOfPipeStartingExpression(parent)
	}
	return false
}

func GetMultiLineStringLiteral(text string) string {
	return MultiLineStringQuote + text + MultiLineStringQuote
}

func GetStringLiteralValue(literal string) string {
	if literal == "" {
		return ""
	}
	start := 0
	if literal[0] == 'h' || literal[0] == 'H' {
		start++
	}
	verbatim := false
	if start < len(literal) && literal[start] == '@' {
		start++
		verbatim = true
	}
	if value, ok := multilineStringLiteralValue(literal[start:], MultiLineStringQuote); ok {
		return value
	}
	if value, ok := multilineStringLiteralValue(literal[start:], AlternateMultiLineStringQuote); ok {
		return value
	}
	if start >= len(literal) {
		return ""
	}
	quote := literal[start]
	bracketed := quote == '['
	endQuote := quote
	if bracketed {
		endQuote = ']'
	}
	end := len(literal)
	if end > 0 && literal[end-1] == endQuote {
		end--
	}
	start++
	if end <= start {
		return ""
	}
	body := literal[start:end]
	if !verbatim && !bracketed && strings.Contains(body, `\`) {
		return decodeStringEscapes(body)
	}
	if verbatim && !bracketed {
		return strings.ReplaceAll(body, string([]byte{endQuote, endQuote}), string(endQuote))
	}
	return body
}

func GetExpressionResultName(expression Expression, defaultName string, rowScope *TableSymbol, globals ...*GlobalState) string {
	var state *GlobalState
	if len(globals) > 0 {
		state = globals[0]
	}
	return expressionResultName(expression, defaultName, rowScope, state)
}

func IsHostName(name, hostName string) bool {
	return ordinalEqualFold(name, hostName)
}

func GetHostName(clusterURIOrName string) string {
	_, host, _, _, _ := getURIParts(clusterURIOrName)
	return host
}

func GetHostAndPath(clusterURIOrName string) (string, string) {
	_, host, _, path, _ := getURIParts(clusterURIOrName)
	return host, path
}

func getURIParts(uri string) (scheme, hostname, port, path, query string) {
	hostnameStart := 0
	if prefixStart := strings.Index(uri, "://"); prefixStart > 0 {
		scheme = uri[:prefixStart]
		hostnameStart = prefixStart + 3
	}

	position := hostnameStart
	for position < len(uri) {
		ch, width := utf8.DecodeRuneInString(uri[position:])
		if !isHostNameRune(ch) {
			break
		}
		position += width
	}
	if hostnameStart == 0 && position == len(uri) {
		hostname = uri
	} else {
		hostname = uri[hostnameStart:position]
	}

	if position < len(uri) && uri[position] == ':' {
		position++
		portStart := position
		for position < len(uri) {
			ch, width := utf8.DecodeRuneInString(uri[position:])
			if !unicode.IsDigit(ch) {
				break
			}
			position += width
		}
		if position > portStart {
			port = uri[portStart:position]
		}
	}

	if position < len(uri) && uri[position] == '/' {
		position++
		pathStart := position
		for position < len(uri) && uri[position] != '?' {
			position++
		}
		if position > pathStart {
			path = uri[pathStart:position]
		}
	}
	if position < len(uri) && uri[position] == '?' {
		query = uri[position:]
	}
	return scheme, hostname, port, path, query
}

func isHostNameRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '.' || ch == '_'
}

func GetFullHostName(hostname, defaultDomainSuffix string) string {
	if IsPossibleShortHostName(hostname) && strings.HasPrefix(defaultDomainSuffix, ".") {
		return hostname + defaultDomainSuffix
	}
	return hostname
}

func HasShortHostName(hostname, defaultDomainSuffix string) bool {
	if !strings.HasPrefix(defaultDomainSuffix, ".") || utf16TextWidth(hostname) <= utf16TextWidth(defaultDomainSuffix) || !ordinalHasSuffix(hostname, defaultDomainSuffix) {
		return false
	}
	units := utf16.Encode([]rune(hostname))
	return possibleShortHostNamePart(string(utf16.Decode(units[:len(units)-utf16TextWidth(defaultDomainSuffix)])))
}

func IsShortHostName(name, hostName, defaultDomainSuffix string) bool {
	return HasShortHostName(hostName, defaultDomainSuffix) && utf16TextWidth(name)+utf16TextWidth(defaultDomainSuffix) == utf16TextWidth(hostName) && ordinalHasPrefix(hostName, name)
}

func IsPossibleShortHostName(name string) bool {
	return possibleShortHostNamePart(name)
}

func possibleShortHostNamePart(name string) bool {
	return name != "" && !strings.HasSuffix(name, ".") && strings.Count(name, ".") < 2
}

func GetShortHostName(hostName, defaultDomainSuffix string) string {
	if HasShortHostName(hostName, defaultDomainSuffix) {
		return hostName[:len(hostName)-len(defaultDomainSuffix)]
	}
	return ""
}

func GetSyntaxTabularity(element SyntaxElement, globals ...*GlobalState) Tabularity {
	if element == nil {
		return TabularityNone
	}
	state := DefaultGlobalState
	if len(globals) > 0 && globals[0] != nil {
		state = globals[0]
	}
	wrapper := WrapGeneratedSyntaxElement(element)
	switch node := wrapper.(type) {
	case *ExpressionStatement:
		return GetSyntaxTabularity(node.Expression(), state)
	case *SyntaxList:
		if node.Count() == 0 {
			return TabularityNone
		}
		last := node.At(node.Count() - 1)
		if separated, ok := last.(*SeparatedElement); ok {
			last = separated.Element()
		}
		return GetSyntaxTabularity(last, state)
	case QueryOperator, *PipeExpression, Command, *DataTableExpression, *ToTableExpression:
		return TabularityTabular
	case *BinaryExpression, *PrefixUnaryExpression, *InExpression, *HasAnyExpression, *HasAllExpression,
		*BetweenExpression, *ElementExpression, *BracketedExpression, *LiteralExpression,
		*CompoundStringLiteralExpression, *DynamicExpression, *ToScalarExpression:
		return TabularityScalar
	case *FunctionCallExpression:
		if function := state.GetFunction(NameReferenceSimpleName(node.Name())); function != nil {
			return function.SymbolTabularity()
		}
		return TabularityUnknown
	case Statement:
		return TabularityNone
	default:
		return TabularityUnknown
	}
}

func GetLiteralType(kind SyntaxKind) *ScalarSymbol {
	switch kind {
	case SyntaxKindBooleanLiteralExpression, SyntaxKindBooleanLiteralToken:
		return ScalarTypes.Bool
	case SyntaxKindStringLiteralExpression, SyntaxKindStringLiteralToken, SyntaxKindCompoundStringLiteralExpression:
		return ScalarTypes.String
	case SyntaxKindIntLiteralExpression, SyntaxKindIntLiteralToken:
		return ScalarTypes.Int
	case SyntaxKindLongLiteralExpression, SyntaxKindLongLiteralToken:
		return ScalarTypes.Long
	case SyntaxKindRealLiteralExpression, SyntaxKindRealLiteralToken:
		return ScalarTypes.Real
	case SyntaxKindDecimalLiteralExpression, SyntaxKindDecimalLiteralToken:
		return ScalarTypes.Decimal
	case SyntaxKindTimespanLiteralExpression, SyntaxKindTimespanLiteralToken:
		return ScalarTypes.TimeSpan
	case SyntaxKindDateTimeLiteralExpression, SyntaxKindDateTimeLiteralToken:
		return ScalarTypes.DateTime
	case SyntaxKindGuidLiteralExpression, SyntaxKindGuidLiteralToken, SyntaxKindRawGuidLiteralToken:
		return ScalarTypes.Guid
	case SyntaxKindTypeOfLiteralExpression:
		return ScalarTypes.TypeLit
	case SyntaxKindDynamicExpression:
		return ScalarTypes.Dynamic
	default:
		return ScalarTypes.Unknown
	}
}
