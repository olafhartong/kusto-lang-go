package kustolang

// Ported from the public grammar surface under src/Kusto.Language/Parser.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// QueryGrammar exposes the typed productions used by the native query parser.
type QueryGrammar = internal.QueryGrammar

// CommandGrammar exposes the command block parser and predefined command rules.
type CommandGrammar = internal.CommandGrammar

// ServerCommandGrammar preserves the concrete server grammar returned by the
// source-shaped typed factories while exposing its common CommandGrammar.
type ServerCommandGrammar = internal.ServerCommandGrammar
type EngineCommandGrammar = internal.EngineCommandGrammar
type DataManagerCommandGrammar = internal.DataManagerCommandGrammar
type ClusterManagerCommandGrammar = internal.ClusterManagerCommandGrammar
type AriaBridgeCommandGrammar = internal.AriaBridgeCommandGrammar

type PredefinedRuleParsers = internal.PredefinedRuleParsers

func QueryGrammarFrom(globals *GlobalState) *QueryGrammar {
	return internal.QueryGrammarFrom(globals)
}

func CommandGrammarFrom(globals *GlobalState) *CommandGrammar {
	return internal.CommandGrammarFrom(globals)
}

func ServerCommandGrammarFrom(globals *GlobalState) ServerCommandGrammar {
	return internal.ServerCommandGrammarFrom(globals)
}

func NewCommandGrammar(globals *GlobalState) *CommandGrammar {
	return internal.NewCommandGrammar(globals)
}

func CreateCommandGrammar(globals *GlobalState) *CommandGrammar {
	return internal.CreateCommandGrammar(globals)
}

func CreateServerCommandGrammar(globals *GlobalState) ServerCommandGrammar {
	return internal.CreateServerCommandGrammar(globals)
}

func NewEngineCommandGrammar(globals *GlobalState) *EngineCommandGrammar {
	return internal.NewEngineCommandGrammar(globals)
}

func NewDataManagerCommandGrammar(globals *GlobalState) *DataManagerCommandGrammar {
	return internal.NewDataManagerCommandGrammar(globals)
}

func NewClusterManagerCommandGrammar(globals *GlobalState) *ClusterManagerCommandGrammar {
	return internal.NewClusterManagerCommandGrammar(globals)
}

func NewAriaBridgeCommandGrammar(globals *GlobalState) *AriaBridgeCommandGrammar {
	return internal.NewAriaBridgeCommandGrammar(globals)
}

func NewPredefinedRuleParsers(inputs ...ResultParser[LexicalToken, SyntaxElement]) *PredefinedRuleParsers {
	return internal.NewPredefinedRuleParsers(inputs...)
}

func NewPredefinedRuleParsersWithGrammar(grammar *QueryGrammar, inputs ...ResultParser[LexicalToken, SyntaxElement]) *PredefinedRuleParsers {
	return internal.NewPredefinedRuleParsersWithGrammar(grammar, inputs...)
}

func CreateMissingNameDeclaration(source Source[LexicalToken], start int) *NameDeclaration {
	return internal.QueryGrammarCreateMissingNameDeclaration(source, start)
}

func CreateMissingNameDeclarationExpression(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingNameDeclarationExpression(source, start)
}

func CreateMissingNameReference(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingNameReference(source, start)
}

func CreateMissingNameToken(source Source[LexicalToken], start int) *SyntaxToken {
	return internal.QueryGrammarCreateMissingNameToken(source, start)
}

func CreateMissingExpression(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingExpression(source, start)
}

func CreateMissingValue(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingValue(source, start)
}

func CreateMissingType(source Source[LexicalToken], start int) TypeExpression {
	return internal.QueryGrammarCreateMissingType(source, start)
}

func CreateMissingLongLiteral(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingLongLiteral(source, start)
}

func CreateMissingRealLiteral(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingRealLiteral(source, start)
}

func CreateMissingStringLiteral(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingStringLiteral(source, start)
}

func CreateMissingBooleanLiteral(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingBooleanLiteral(source, start)
}

func CreateMissingTypeOfLiteral(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingTypeOfLiteral(source, start)
}

func CreateMissingJSONValue(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingJSONValue(source, start)
}

func CreateMissingJSONPair(source Source[LexicalToken], start int) *JsonPair {
	return internal.QueryGrammarCreateMissingJSONPair(source, start)
}

func CreateMissingQueryOperator(source Source[LexicalToken], start int) QueryOperator {
	return internal.QueryGrammarCreateMissingQueryOperator(source, start)
}

func CreateMissingQueryOperatorExpression(source Source[LexicalToken], start int) Expression {
	return internal.QueryGrammarCreateMissingQueryOperatorExpression(source, start)
}

// QueryGrammarStringOperatorMap returns a copy so callers cannot alter parser behavior.
func QueryGrammarStringOperatorMap() map[SyntaxKind]SyntaxKind {
	return internal.QueryGrammarStringOperatorMap()
}
