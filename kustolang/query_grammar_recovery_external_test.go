package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicQueryGrammarRecoveryFactorySignatures(t *testing.T) {
	var source kustolang.Source[kustolang.LexicalToken]
	var namedExpression kustolang.NamedExpression = kustolang.CreateMissingNamedExpression(source, 0)
	var scanAssignment *kustolang.ScanAssignment = kustolang.CreateMissingScanAssignment(source, 0)
	var joinOn kustolang.JoinConditionClause = kustolang.CreateMissingJoinOnClause(source, 0)
	var functionCall *kustolang.FunctionCallExpression = kustolang.CreateMissingFunctionCall(source, 0)
	var functionCallExpression kustolang.Expression = kustolang.CreateMissingFunctionCallExpression(source, 0)
	var schemaType *kustolang.SchemaTypeExpression = kustolang.CreateMissingSchemaType(source, 0)
	var rowSchema *kustolang.RowSchema = kustolang.CreateMissingRowSchema(source, 0)
	var kindClause *kustolang.InlineExternalTableKindClause = kustolang.CreateMissingInlineExternalTableKindClause(source, 0)
	var dataFormatClause *kustolang.InlineExternalTableDataFormatClause = kustolang.CreateMissingInlineExternalTableDataFormatClause(source, 0)
	var connectionStringsClause *kustolang.InlineExternalTableConnectionStringsClause = kustolang.CreateMissingInlineExternalTableConnectionStringsClause(source, 0)
	var pathFormat *kustolang.InlineExternalTablePathFormatPartitionColumnReference = kustolang.CreateMissingPathFormatTokens(source, 0)
	var partitionColumn *kustolang.PartitionColumnDeclaration = kustolang.CreateMissingPartitionColumnDeclaration(source, 0)
	var evaluateRowSchema *kustolang.EvaluateRowSchema = kustolang.CreateMissingEvaluateRowSchema(source, 0)
	var makeSeries *kustolang.MakeSeriesExpression = kustolang.CreateMissingMakeSeriesExpression(source, 0)
	var mvExpand *kustolang.MvExpandExpression = kustolang.CreateMissingMvExpandExpression(source, 0)
	var mvApply *kustolang.MvApplyExpression = kustolang.CreateMissingMvApplyExpression(source, 0)
	var mvApplySubquery *kustolang.MvApplySubqueryExpression = kustolang.CreateMissingMvApplySubqueryExpression(source, 0)
	var fork *kustolang.ForkExpression = kustolang.CreateMissingForkExpression(source, 0)
	var partitionOperand kustolang.PartitionOperand = kustolang.CreateMissingPartitionOperand(source, 0)
	var statement kustolang.Statement = kustolang.CreateMissingStatement(source, 0)
	var statementElement *kustolang.SeparatedElementOf[kustolang.Statement] = kustolang.CreateMissingStatementElement(source, 0)
	var nameAndType *kustolang.NameAndTypeDeclaration = kustolang.CreateMissingNameAndTypeDeclaration(source, 0)
	var functionParameter *kustolang.FunctionParameter = kustolang.CreateMissingFunctionParameter(source, 0)
	var namedParameter *kustolang.NamedParameter = kustolang.CreateMissingNamedParameter(source, 0)
	var functionDeclaration *kustolang.FunctionDeclaration = kustolang.CreateMissingFunctionDeclaration(source, 0)
	var baseClause *kustolang.MaterializedViewCombineClause = kustolang.CreateMissingMaterializedViewCombineBaseClause(source, 0)
	var deltaClause *kustolang.MaterializedViewCombineClause = kustolang.CreateMissingMaterializedViewCombineDeltaClause(source, 0)
	var aggregatesClause *kustolang.MaterializedViewCombineClause = kustolang.CreateMissingMaterializedViewCombineAggregatesClause(source, 0)
	var nameClause *kustolang.MaterializedViewCombineNameClause = kustolang.CreateMissingMaterializedViewCombineNameClause(source, 0)

	for label, element := range map[string]kustolang.SyntaxElement{
		"named expression": namedExpression, "scan assignment": scanAssignment, "join on": joinOn,
		"function call": functionCall, "function call expression": functionCallExpression,
		"schema type": schemaType, "row schema": rowSchema, "kind clause": kindClause,
		"data format clause": dataFormatClause, "connection strings clause": connectionStringsClause,
		"path format": pathFormat, "partition column": partitionColumn, "evaluate row schema": evaluateRowSchema,
		"make series": makeSeries, "mv expand": mvExpand, "mv apply": mvApply,
		"mv apply subquery": mvApplySubquery, "fork": fork, "partition operand": partitionOperand,
		"statement": statement, "statement element": statementElement, "name and type": nameAndType,
		"function parameter": functionParameter, "named parameter": namedParameter,
		"function declaration": functionDeclaration, "base clause": baseClause, "delta clause": deltaClause,
		"aggregates clause": aggregatesClause, "name clause": nameClause,
	} {
		if element == nil || element.Width() != 0 {
			t.Fatalf("%s = %#v, width %d", label, element, element.Width())
		}
	}
}

func TestPublicQueryGrammarExactJsonFactorySpellings(t *testing.T) {
	value := kustolang.CreateMissingJsonValue(nil, 0)
	pair := kustolang.CreateMissingJsonPair(nil, 0)
	if value == nil || value.Kind() != kustolang.SyntaxKindStringLiteralExpression {
		t.Fatalf("CreateMissingJsonValue() = %#v", value)
	}
	if pair == nil || pair.Kind() != kustolang.SyntaxKindJsonPair {
		t.Fatalf("CreateMissingJsonPair() = %#v", pair)
	}
}

func TestPublicCreateMissingTokenLiteralGoSpellings(t *testing.T) {
	var variadic func(kustolang.Source[kustolang.LexicalToken], int) kustolang.Expression = kustolang.CreateMissingTokenLiteral("left", "right")
	var fromList func(kustolang.Source[kustolang.LexicalToken], int) kustolang.Expression = kustolang.CreateMissingTokenLiteralFromList([]string{"left", "right"})

	for name, factory := range map[string]func(kustolang.Source[kustolang.LexicalToken], int) kustolang.Expression{
		"variadic": variadic, "list": fromList,
	} {
		t.Run(name, func(t *testing.T) {
			expression := factory(nil, 0)
			if expression.Kind() != kustolang.SyntaxKindTokenLiteralExpression || expression.Width() != 0 {
				t.Fatalf("kind/width = %v/%d", expression.Kind(), expression.Width())
			}
			got := make([]string, len(expression.Diagnostics()))
			for i, diagnostic := range expression.Diagnostics() {
				got[i] = diagnostic.Description
			}
			if !reflect.DeepEqual(got, []string{"Expected one of: left, right"}) {
				t.Fatalf("diagnostics = %#v", got)
			}
		})
	}
}
