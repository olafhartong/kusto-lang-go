package kustolang

// Ported from src/Kusto.Language/Parser/QueryParser.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// QueryParserAPI mirrors the upstream static QueryParser entry points. Go
// overloads are expressed with explicit Tokens, From, and WithOptions names.
type QueryParserAPI struct{}

func (QueryParserAPI) ParseQuery(text string) *QueryBlock { return ParseQuery(text) }
func (QueryParserAPI) ParseQueryWithOptions(text string, options ParseOptions) *QueryBlock {
	return ParseQueryWithOptions(text, options)
}
func (QueryParserAPI) ParseQueryTokens(tokens []LexicalToken) *QueryBlock {
	return ParseQueryTokens(tokens)
}
func (QueryParserAPI) ParseQueryTokensWithOptions(tokens []LexicalToken, options ParseOptions) *QueryBlock {
	return ParseQueryTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseQueryTokensFrom(tokens []LexicalToken, start int) *QueryBlock {
	return ParseQueryTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseQueryTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *QueryBlock {
	return ParseQueryTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseExpression(text string) Expression { return ParseExpression(text) }
func (QueryParserAPI) ParseExpressionWithOptions(text string, options ParseOptions) Expression {
	return ParseExpressionWithOptions(text, options)
}
func (QueryParserAPI) ParseExpressionTokens(tokens []LexicalToken) Expression {
	return ParseExpressionTokens(tokens)
}
func (QueryParserAPI) ParseExpressionTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseExpressionTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseExpressionTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseExpressionTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseExpressionTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return ParseExpressionTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseLiteral(text string) Expression { return ParseLiteral(text) }
func (QueryParserAPI) ParseLiteralWithOptions(text string, options ParseOptions) Expression {
	return ParseLiteralWithOptions(text, options)
}
func (QueryParserAPI) ParseLiteralTokens(tokens []LexicalToken) Expression {
	return ParseLiteralTokens(tokens)
}
func (QueryParserAPI) ParseLiteralTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseLiteralTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseLiteralTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseLiteralTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseLiteralTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return ParseLiteralTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseEntityPath(text string) Expression { return ParseEntityPath(text) }
func (QueryParserAPI) ParseEntityPathWithOptions(text string, options ParseOptions) Expression {
	return ParseEntityPathWithOptions(text, options)
}
func (QueryParserAPI) ParseEntityPathTokens(tokens []LexicalToken) Expression {
	return ParseEntityPathTokens(tokens)
}
func (QueryParserAPI) ParseEntityPathTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseEntityPathTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseEntityPathTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseEntityPathTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseEntityPathTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return ParseEntityPathTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseEntityGroup(text string) Expression { return ParseEntityGroup(text) }
func (QueryParserAPI) ParseEntityGroupWithOptions(text string, options ParseOptions) Expression {
	return ParseEntityGroupWithOptions(text, options)
}
func (QueryParserAPI) ParseEntityGroupTokens(tokens []LexicalToken) Expression {
	return ParseEntityGroupTokens(tokens)
}
func (QueryParserAPI) ParseEntityGroupTokensWithOptions(tokens []LexicalToken, options ParseOptions) Expression {
	return ParseEntityGroupTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseEntityGroupTokensFrom(tokens []LexicalToken, start int) Expression {
	return ParseEntityGroupTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseEntityGroupTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) Expression {
	return ParseEntityGroupTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseRowSchema(text string) *RowSchema { return ParseRowSchema(text) }
func (QueryParserAPI) ParseRowSchemaWithOptions(text string, options ParseOptions) *RowSchema {
	return ParseRowSchemaWithOptions(text, options)
}
func (QueryParserAPI) ParseRowSchemaTokens(tokens []LexicalToken) *RowSchema {
	return ParseRowSchemaTokens(tokens)
}
func (QueryParserAPI) ParseRowSchemaTokensWithOptions(tokens []LexicalToken, options ParseOptions) *RowSchema {
	return ParseRowSchemaTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseRowSchemaTokensFrom(tokens []LexicalToken, start int) *RowSchema {
	return ParseRowSchemaTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseRowSchemaTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *RowSchema {
	return ParseRowSchemaTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseFunctionParameters(text string) *FunctionParameters {
	return ParseFunctionParameters(text)
}
func (QueryParserAPI) ParseFunctionParametersWithOptions(text string, options ParseOptions) *FunctionParameters {
	return ParseFunctionParametersWithOptions(text, options)
}
func (QueryParserAPI) ParseFunctionParametersTokens(tokens []LexicalToken) *FunctionParameters {
	return ParseFunctionParametersTokens(tokens)
}
func (QueryParserAPI) ParseFunctionParametersTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameters {
	return ParseFunctionParametersTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseFunctionParametersTokensFrom(tokens []LexicalToken, start int) *FunctionParameters {
	return ParseFunctionParametersTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseFunctionParametersTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameters {
	return ParseFunctionParametersTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseFunctionParameter(text string) *FunctionParameter {
	return ParseFunctionParameter(text)
}
func (QueryParserAPI) ParseFunctionParameterWithOptions(text string, options ParseOptions) *FunctionParameter {
	return ParseFunctionParameterWithOptions(text, options)
}
func (QueryParserAPI) ParseFunctionParameterTokens(tokens []LexicalToken) *FunctionParameter {
	return ParseFunctionParameterTokens(tokens)
}
func (QueryParserAPI) ParseFunctionParameterTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionParameter {
	return ParseFunctionParameterTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseFunctionParameterTokensFrom(tokens []LexicalToken, start int) *FunctionParameter {
	return ParseFunctionParameterTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseFunctionParameterTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionParameter {
	return ParseFunctionParameterTokensFromWithOptions(tokens, start, options)
}

func (QueryParserAPI) ParseFunctionBody(text string) *FunctionBody { return ParseFunctionBody(text) }
func (QueryParserAPI) ParseFunctionBodyWithOptions(text string, options ParseOptions) *FunctionBody {
	return ParseFunctionBodyWithOptions(text, options)
}
func (QueryParserAPI) ParseFunctionBodyTokens(tokens []LexicalToken) *FunctionBody {
	return ParseFunctionBodyTokens(tokens)
}
func (QueryParserAPI) ParseFunctionBodyTokensWithOptions(tokens []LexicalToken, options ParseOptions) *FunctionBody {
	return ParseFunctionBodyTokensWithOptions(tokens, options)
}
func (QueryParserAPI) ParseFunctionBodyTokensFrom(tokens []LexicalToken, start int) *FunctionBody {
	return ParseFunctionBodyTokensFrom(tokens, start)
}
func (QueryParserAPI) ParseFunctionBodyTokensFromWithOptions(tokens []LexicalToken, start int, options ParseOptions) *FunctionBody {
	return ParseFunctionBodyTokensFromWithOptions(tokens, start, options)
}

// QueryParser exposes source-shaped access to the upstream static API.
var QueryParser QueryParserAPI
