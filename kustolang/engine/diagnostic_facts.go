package kustolang

import (
	"fmt"
	"strings"
)

// Ported from src/Kusto.Language/Diagnostics/DiagnosticFacts.cs. This file is
// intentionally focused on parser/binder diagnostics currently used by the Go
// analyzer and binder facade.

func DiagnosticMissingText(text string) Diagnostic {
	return NewDiagnostic("KS001", "Missing: "+text)
}

func DiagnosticUnexpectedCharacter(text string) Diagnostic {
	return NewDiagnostic("KS002", "Unexpected: "+text)
}

func DiagnosticMalformedToken(term string) Diagnostic {
	return NewDiagnostic("KS003", "Malformed "+term)
}

func DiagnosticMalformedLiteral() Diagnostic {
	return NewDiagnostic("KS004", "Malformed literal")
}

func DiagnosticTermsExpected(terms ...string) Diagnostic {
	if len(terms) == 1 {
		return NewDiagnostic("KS005", "Expected: "+terms[0])
	}
	return NewDiagnostic("KS005", "Expected one of: "+strings.Join(terms, ", "))
}

func DiagnosticMissingName() Diagnostic {
	return NewDiagnostic("KS006", "Missing name")
}

func DiagnosticMissingNameWithKeyword(keyword string) Diagnostic {
	return NewDiagnostic("KS006", fmt.Sprintf("Missing name: If the keyword '%s' is intended be used as the name, it needs to be bracketted as %s.", keyword, GetBracketedName(keyword)))
}

func DiagnosticMissingTypeName() Diagnostic {
	return NewDiagnostic("KS006", "Missing type name")
}

func DiagnosticMissingTypeOfLiteral() Diagnostic {
	return NewDiagnostic("KS006", "Missing typeof")
}

func DiagnosticMissingSchemaDeclaration() Diagnostic {
	return NewDiagnostic("KS006", "Missing schema declaration")
}

func DiagnosticMissingExpression() Diagnostic {
	return NewDiagnostic("KS006", "Missing expression")
}

func DiagnosticMissingStatement() Diagnostic {
	return NewDiagnostic("KS006", "Missing statement")
}

func DiagnosticMissingFunctionCall() Diagnostic {
	return NewDiagnostic("KS006", "Missing function call")
}

func DiagnosticMissingClause() Diagnostic {
	return NewDiagnostic("KS006", "Missing clause")
}

func DiagnosticMissingMaterializedViewCombineClause(name string) Diagnostic {
	return NewDiagnostic("KS006", fmt.Sprintf("Missing %s clause", name))
}

func DiagnosticMissingExpressionWithKeyword(keyword string) Diagnostic {
	return NewDiagnostic("KS006", fmt.Sprintf("Missing expression: If the keyword '%s' is intended to be part of an expression it needs to be bracketted as %s.", keyword, GetBracketedName(keyword)))
}

func DiagnosticNameRequiresBrackets(name string) Diagnostic {
	return NewDiagnostic("KS228", fmt.Sprintf("The name '%s' needs to be bracketed as %s to be used in this context.", name, GetBracketedName(name)))
}

func DiagnosticMissingValue() Diagnostic {
	return NewDiagnostic("KS006", "Missing value")
}

func DiagnosticMissingNumber() Diagnostic {
	return NewDiagnostic("KS006", "Missing number")
}

func DiagnosticMissingString() Diagnostic {
	return NewDiagnostic("KS006", "Missing string")
}

func DiagnosticMissingBoolean() Diagnostic {
	return NewDiagnostic("KS006", "Missing boolean")
}

func DiagnosticMissingParameter() Diagnostic {
	return NewDiagnostic("KS006", "Missing parameter")
}

func DiagnosticMissingFirstOrLast() Diagnostic {
	return NewDiagnostic("KS006", "Missing first or last")
}

func DiagnosticMissingJoinOnClause() Diagnostic {
	return NewDiagnostic("KS006", "Missing join on condition clause")
}

func DiagnosticMissingJsonValue() Diagnostic {
	return NewDiagnostic("KS006", "Missing json value")
}

func DiagnosticMissingJsonPair() Diagnostic {
	return NewDiagnostic("KS006", "Missing json key:value pair")
}

func DiagnosticMissingPatternMatch() Diagnostic {
	return NewDiagnostic("KS006", "Missing pattern match clause")
}

func DiagnosticAnalyzerFailure(analyzerName string, message string) Diagnostic {
	return NewDiagnostic("KS203", fmt.Sprintf("Failure in analysis '%s': %s", analyzerName, message))
}

func DiagnosticClientParametersNotSupported() Diagnostic {
	return NewDiagnostic("KS213", "Client parameters are not supported or enabled.")
}

func DiagnosticQueryTextSizeExceeded() Diagnostic {
	return NewDiagnostic("KS244", "Query text size exceeds maximum safe limit. Parsing not performed.")
}

func DiagnosticQuerySyntaxDepthExceeded() Diagnostic {
	return NewDiagnostic("KS245", "Query syntax depth exceeds maximum safe limit. Semantic analysis not performed.")
}

func DiagnosticInternalFailure() Diagnostic {
	return NewDiagnostic("KS246", "An internal failure occurred during parsing.")
}

func DiagnosticQueryOperatorExpected() Diagnostic {
	return NewDiagnostic("KS176", "Query operator expected.")
}

func DiagnosticQueryOperatorNotAllowedInContext(name string) Diagnostic {
	return NewDiagnostic("KS177", fmt.Sprintf("The query operator '%s' is not allowed in the current context.", name))
}

func DiagnosticSchemaDifferentThanExpected() Diagnostic {
	return NewDiagnostic("SEM0527", "The tabular operator provided to assert-schema has a different schema than expected.")
}

func DiagnosticMissingCommand() Diagnostic {
	return NewDiagnostic("KS300", "Missing command.")
}

func DiagnosticInvalidTypeName(name string) Diagnostic {
	return NewDiagnostic("KS160", fmt.Sprintf("The name '%s' is not a valid type name that can be used here.", name))
}

func DiagnosticParsePatternMustStartWithColumnNameOrStar() Diagnostic {
	return NewDiagnostic("KS100", "The pattern must start with a column name or *")
}

func DiagnosticParsePatternNameDoesNotFollowStringLiteral() Diagnostic {
	return NewDiagnostic("KS101", "The column name must follow a string literal")
}

func DiagnosticParsePatternStringLiteralMustFollowStar() Diagnostic {
	return NewDiagnostic("KS102", "A string literal must follow a *")
}

func DiagnosticParsePatternUsingStarAfterStringColumnIsAmbiguous() Diagnostic {
	return NewDiagnostic("KS103", "Using * after parsing a string column is ambiguous.")
}

func DiagnosticInvalidPatternPart() Diagnostic {
	return NewDiagnostic("KS104", "Invalid pattern part.")
}

func DiagnosticIdentifierNameOnly() Diagnostic {
	return NewDiagnostic("KS105", "The name must be a single identifier only.")
}

func DiagnosticIncorrectNumberOfDataValues(multiple int) Diagnostic {
	return NewDiagnostic("KS173", fmt.Sprintf("Incorrect number of data values. The values should appear in multiples of %d.", multiple))
}

func DiagnosticExpressionMustHaveType(types ...string) Diagnostic {
	if len(types) == 1 {
		return NewDiagnostic("KS141", fmt.Sprintf("The expression must have the type %s.", types[0]))
	}
	return NewDiagnostic("KS141", fmt.Sprintf("The expression must have one of the types: %s.", joinEnglish(types, ", ", " or ")))
}

func DiagnosticValueCountMustEqualParameterCount() Diagnostic {
	return NewDiagnostic("KS185", "The number of values must equal the number of parameters.")
}

func DiagnosticPackMustBeLastItemInList() Diagnostic {
	return NewDiagnostic("KS185", "The pack(*) expression must be the last item in the list.")
}

func DiagnosticPathValueWithNoPathParameter() Diagnostic {
	return NewDiagnostic("KS186", "A path value can only be specified when a path name is part of the declaration.")
}

func DiagnosticPathValueExpected() Diagnostic {
	return NewDiagnostic("KS187", "A path value is expected.")
}

func DiagnosticOperatorNotDefined(name string, argumentTypes ...string) Diagnostic {
	if len(argumentTypes) == 1 {
		return NewDiagnostic("KS106", fmt.Sprintf("The operator '%s' is not defined for the operand type %s.", name, argumentTypes[0]))
	}
	return NewDiagnostic("KS106", fmt.Sprintf("The operator '%s' is not defined for the operand types %s.", name, joinEnglish(argumentTypes, ", ", " and ")))
}

func DiagnosticTypeExpected(types ...string) Diagnostic {
	if len(types) == 1 {
		return NewDiagnostic("KS107", fmt.Sprintf("A value of type %s expected.", types[0]))
	}
	return NewDiagnostic("KS107", fmt.Sprintf("A value of type %s expected.", joinEnglish(types, ", ", " or ")))
}

func DiagnosticNoPatternMatchesArguments() Diagnostic {
	return NewDiagnostic("KS188", "No pattern matches the specified arguments.")
}

func DiagnosticScalarTypeExpected() Diagnostic {
	return NewDiagnostic("KS108", "Scalar value expected.")
}

func DiagnosticTabularValueExpected() Diagnostic {
	return NewDiagnostic("KS111", "Tabular value expected.")
}

func DiagnosticDatabaseExpected() Diagnostic {
	return NewDiagnostic("KS114", "Database expected.")
}

func DiagnosticClusterExpected() Diagnostic {
	return NewDiagnostic("KS115", "Cluster expected.")
}

func DiagnosticTableExpected() Diagnostic {
	return NewDiagnostic("KS251", "Table expected.")
}

func DiagnosticTableOrScalarExpected() Diagnostic {
	return NewDiagnostic("KS112", "A tabular or scalar value expected.")
}

func DiagnosticPluginNotAllowedInThisContext(name string) Diagnostic {
	return NewDiagnostic("KS126", fmt.Sprintf("The plug-in function '%s' is not allowed in this context.", name))
}

func DiagnosticFunctionNotDefinedWithMatchingParameters(name string, types ...string) Diagnostic {
	return NewDiagnostic("KS127", fmt.Sprintf("The function '%s' is not compatible with arguments (%s)", name, strings.Join(types, ", ")))
}

func DiagnosticAggregateNotAllowedInThisContext(name string) Diagnostic {
	return NewDiagnostic("KS144", fmt.Sprintf("The aggregate function '%s' is not allowed in this context.", name))
}

func DiagnosticElementAccessNotAllowedInContext() Diagnostic {
	return NewDiagnostic("KS147", "The element access operator [] is not allowed in this context.")
}

func DiagnosticPluginFunctionNotDefined(name string) Diagnostic {
	return NewDiagnostic("KS124", fmt.Sprintf("The plug-in function '%s' is not defined.", name))
}

func DiagnosticPluginFunctionIsNotEnabled(name string) Diagnostic {
	return NewDiagnostic("KS125", fmt.Sprintf("The plug-in function '%s' is not enabled.", name))
}

func DiagnosticFunctionRequiresArgumentList(functionName string) Diagnostic {
	return NewDiagnostic("KS117", fmt.Sprintf("The function '%s' requires an argument list.", functionName))
}

func DiagnosticArgumentCountExpected(count int) Diagnostic {
	switch count {
	case 0:
		return NewDiagnostic("KS118", "No arguments expected.")
	case 1:
		return NewDiagnostic("KS118", "1 argument expected.")
	default:
		return NewDiagnostic("KS118", fmt.Sprintf("%d arguments expected.", count))
	}
}

func DiagnosticExpressionMustBeConstantOrIdentifier() Diagnostic {
	return NewDiagnostic("KS130", "The expression must be a constant or identifier.")
}

func DiagnosticExpressionMustBeConstant() Diagnostic {
	return NewDiagnostic("KS129", "The expression must be a constant.")
}

func DiagnosticExpressionMustBeLiteral() Diagnostic {
	return NewDiagnostic("KS131", "The expression must be a literal.")
}

func DiagnosticExpressionMustBeLiteralScalarValue() Diagnostic {
	return NewDiagnostic("KS132", "The expression must be a literal scalar value.")
}

func DiagnosticExpressionMustNotBeEmpty() Diagnostic {
	return NewDiagnostic("KS133", "The expression value must not be empty.")
}

func DiagnosticMultiValuedExpressionCannotBeAssignedToVariable() Diagnostic {
	return NewDiagnostic("KS139", "The multi-valued expression cannot be assigned to a variable.")
}

func DiagnosticTypeIsNotIntervalType(intervalType, rangeType string) Diagnostic {
	return NewDiagnostic("KS178", fmt.Sprintf("The type '%s' is not an appropriate interval type for '%s'", intervalType, rangeType))
}

func DiagnosticExpressionMustBeInteger() Diagnostic {
	return NewDiagnostic("KS134", "The expression value must be an integer.")
}

func DiagnosticExpressionMustBeRealOrDecimal() Diagnostic {
	return NewDiagnostic("KS135", "The expression value must be an real or decimal number.")
}

func DiagnosticExpressionMustBeIntegerOrArray() Diagnostic {
	return NewDiagnostic("KS136", "The expression value must be an integer or a dynamic array of integers.")
}

func DiagnosticExpressionMustBeNumeric() Diagnostic {
	return NewDiagnostic("KS137", "The expression value must be a number.")
}

func DiagnosticExpressionMustBeNumericOrBool() Diagnostic {
	return NewDiagnostic("KS137", "The expression value must be a number or boolean true/false.")
}

func DiagnosticExpressionMustBeSummable() Diagnostic {
	return NewDiagnostic("KS138", "The argument value must be summable: a number, timespan or datetime.")
}

func DiagnosticExpressionMustBeStringOrArray() Diagnostic {
	return NewDiagnostic("KS233", "The expression value must be a string or a dynamic array of strings.")
}

func DiagnosticExpressionMustBeOrderable() Diagnostic {
	return NewDiagnostic("KS206", "The argument value must be orderable: a number, timespan, datetime, string or boolean.")
}

func DiagnosticExpressionMustBeDynamicArray() Diagnostic {
	return NewDiagnostic("KS234", "The expression value must be a dynamic array.")
}

func DiagnosticExpressionMustBeDynamicBag() Diagnostic {
	return NewDiagnostic("KS235", "The expression must be a dynamic bag.")
}

func DiagnosticInvalidNameInAggregateContext(name string) Diagnostic {
	return NewDiagnostic("KS236", fmt.Sprintf("The name '%s' is not an aggregate or scalar function.", name))
}

func DiagnosticInvalidNameInPlugInContext(name string) Diagnostic {
	return NewDiagnostic("KS237", fmt.Sprintf("The name '%s' is not a plug-in function.", name))
}

func DiagnosticExpressionMustHaveValue(values ...string) Diagnostic {
	if len(values) == 1 {
		return NewDiagnostic("KS140", fmt.Sprintf("The expression must be the value: %s", values[0]))
	}
	return NewDiagnostic("KS140", fmt.Sprintf("The expression must be one of the values: %s", joinEnglish(values, ", ", " or ")))
}

func DiagnosticExpressionMustNotHaveValue(values ...string) Diagnostic {
	if len(values) == 1 {
		return NewDiagnostic("KS140", fmt.Sprintf("The expression must not be the value: %s", values[0]))
	}
	return NewDiagnostic("KS140", fmt.Sprintf("The expression must not be one of the values: %s", joinEnglish(values, ", ", " or ")))
}

func DiagnosticColumnExpected() Diagnostic {
	return NewDiagnostic("KS109", "Column name expected.")
}

func DiagnosticRenameAssignmentExpected() Diagnostic {
	return NewDiagnostic("KS110", "Column rename assignment expected.")
}

func DiagnosticTypeNotAllowed(typ string) Diagnostic {
	return NewDiagnostic("KS116", fmt.Sprintf("The value of type '%s' is not allowed in this context.", typ))
}

func DiagnosticInvalidColumnDeclaration() Diagnostic {
	return NewDiagnostic("KS170", "The syntax is not a valid column declaration.")
}

func DiagnosticDuplicateColumnDeclaration(name string) Diagnostic {
	return NewDiagnostic("KS171", fmt.Sprintf("A column with the name '%s' is already declared.", name))
}

func DiagnosticInvalidTypeExpression() Diagnostic {
	return NewDiagnostic("KS172", "The syntax is not a valid type expression.")
}

func DiagnosticQueryOperatorCannotBeFirst() Diagnostic {
	return NewDiagnostic("KS174", "The operator cannot be the first operator in a query.")
}

func DiagnosticQueryOperatorMustBeFirst() Diagnostic {
	return NewDiagnostic("KS175", "The operator must be the first operator in the query.")
}

func DiagnosticColumnDeclarationExpected() Diagnostic {
	return NewDiagnostic("KS220", "Column declaration expected")
}

func DiagnosticTableHasNoColumns() Diagnostic {
	return NewDiagnostic("KS190", "The table has no columns")
}

func DiagnosticNoColumnsInScope() Diagnostic {
	return NewDiagnostic("KS199", "No columns are currently in scope.")
}

func DiagnosticStarExpressionNotAllowed() Diagnostic {
	return NewDiagnostic("KS191", "The * syntax is not allowed here.")
}

func DiagnosticStarExpressionMustBeLastArgument() Diagnostic {
	return NewDiagnostic("KS192", "The * syntax must be the last argument.")
}

func DiagnosticNamedArgumentsNotSupported() Diagnostic {
	return NewDiagnostic("KS193", "Named arguments are not supported for this function.")
}

func DiagnosticCompoundNamedArgumentsNotSupported() Diagnostic {
	return NewDiagnostic("KS194", "Compound named arguments are not supported.")
}

func DiagnosticUnnamedArgumentAfterOutOfOrderNamedArgument() Diagnostic {
	return NewDiagnostic("KS195", "All arguments after an unordered named argument must be named.")
}

func DiagnosticUnknownArgumentName() Diagnostic {
	return NewDiagnostic("KS196", "The argument name does not refer to a declared parameter.")
}

func DiagnosticMissingArgumentForParameter(parameterName string) Diagnostic {
	return NewDiagnostic("KS197", fmt.Sprintf("The argument for parameter '%s' is missing.", parameterName))
}

func DiagnosticIncompleteFragment() Diagnostic {
	return NewDiagnostic("KS198", "The incomplete fragment is unexpected.")
}

func DiagnosticUnknownDirective(name string) Diagnostic {
	return NewDiagnostic("KS250", fmt.Sprintf("Unknown directive '%s'.", name))
}

func DiagnosticErrorInExpansion(name, errors string) Diagnostic {
	return NewDiagnostic("KS200", fmt.Sprintf("Failure in expansion of '%s': %s", name, errors))
}

func DiagnosticNoCommonArgumentType() Diagnostic {
	return NewDiagnostic("KS232", "No common type can be determined from the arguments.")
}

func DiagnosticUnknownQueryOperatorParameterName(name string) Diagnostic {
	return NewDiagnostic("KS179", fmt.Sprintf("The name '%s' is not a recognized parameter for this operator.", name)).
		WithSeverity(DiagnosticSeverityWarning)
}

func DiagnosticParameterAlreadySpecified(name string) Diagnostic {
	return NewDiagnostic("KS180", fmt.Sprintf("The parameter '%s' is already specified.", name))
}

func DiagnosticInvalidJoinCondition() Diagnostic {
	return NewDiagnostic("KS182", "The join condition must be either the name of a column common to both tables or in the form $left.<column> == $right.<column>.")
}

func DiagnosticInvalidJoinConditionOperand(prefix string) Diagnostic {
	return NewDiagnostic("KS183", fmt.Sprintf("The join condition operand must be: %s.<column>", prefix))
}

func DiagnosticColumnMustExistOnBothSidesOfJoin(name string) Diagnostic {
	return NewDiagnostic("KS145", fmt.Sprintf("The column '%s' must exist on both sides of the join.", name))
}

func DiagnosticCommonJoinColumnsMustHaveSameType(name string) Diagnostic {
	return NewDiagnostic("KS227", fmt.Sprintf("The common column '%s' must have the same type on both sides of the join.", name))
}

func DiagnosticJoinKeyCannotBeDynamic() Diagnostic {
	return NewDiagnostic("KS241", "Join keys cannot be dynamic")
}

func DiagnosticJoinKeysNotComparable(leftType, rightType string) Diagnostic {
	return NewDiagnostic("KS242", fmt.Sprintf("Inconsistent data types for join keys: (%s, %s)", leftType, rightType))
}

func DiagnosticExpressionDoesNotHaveMultipleValues() Diagnostic {
	return NewDiagnostic("KS149", "The expression does not have multiple named values.")
}

func DiagnosticNameDoesNotHaveCorrespondingExpression() Diagnostic {
	return NewDiagnostic("KS150", "The name does not have a corresponding expression.")
}

func DiagnosticMissingGraphMatchPattern() Diagnostic {
	return NewDiagnostic("KS217", "Missing graph-match pattern.")
}

func DiagnosticGraphExpected() Diagnostic {
	return NewDiagnostic("KS218", "Graph expected.")
}

func DiagnosticQueryOperatorExpectsGraph() Diagnostic {
	return NewDiagnostic("KS219", "The query operator requires an input graph.")
}

func DiagnosticMissingGraphEntityType() Diagnostic {
	return NewDiagnostic("KS229", "Missing graph entity type, Expected values [nodes, edges].")
}

func DiagnosticIncorrectNumberOfOutputGraphToTableEntities() Diagnostic {
	return NewDiagnostic("KS230", "The graph-to-table operator requires 1-2 arguments (nodes, edges).")
}

func DiagnosticMissingGraphMatchPatternElement() Diagnostic {
	return NewDiagnostic("KS238", "Missing graph-match pattern element.")
}

func DiagnosticGraphMatchPatternSyntaxError(expectedElementType string, actualElementType string) Diagnostic {
	return NewDiagnostic("KS239", fmt.Sprintf("graph-match element expected type is '%s', but received '%s' type.", expectedElementType, actualElementType))
}

func DiagnosticMakeGraphDynamicNodeIDColumnNotSupported() Diagnostic {
	return NewDiagnostic("KS240", "source, target and node id columns can't be of type dynamic, consider using explicit cast")
}

func DiagnosticMakeGraphImplicitIDShouldNotBeEmpty() Diagnostic {
	return NewDiagnostic("KS243", "Implicit node ID should not be empty.")
}

func DiagnosticMultipleGraphOutputsRequireAliases() Diagnostic {
	return NewDiagnostic("KS273", "Multiple graph outputs require aliases.")
}

func DiagnosticMissingPartitionColumnDeclaration() Diagnostic {
	return NewDiagnostic("KS252", "Partition column definition is expected.")
}

func DiagnosticMissingDataFormat() Diagnostic {
	return NewDiagnostic("KS253", "DataFormat definition is expected.")
}

func DiagnosticMissingExternalTableKind() Diagnostic {
	return NewDiagnostic("KS254", "External Table Kind definition is expected.")
}

func DiagnosticMissingPathFormatTokens() Diagnostic {
	return NewDiagnostic("KS255", "External Table Path Format is empty.")
}

func DiagnosticMissingConnectionStrings() Diagnostic {
	return NewDiagnostic("KS256", "Connection strings definition is expected.")
}

func DiagnosticUnknownTokenInPathFormatDefinition() Diagnostic {
	return NewDiagnostic("KS257", "External Table Path Format syntax has unsupported expresion.")
}

func DiagnosticWrongPartitionColumnType() Diagnostic {
	return NewDiagnostic("KS258", "External Table Partition Column type is not supported.")
}

func DiagnosticWrongPartitionColumnFunction() Diagnostic {
	return NewDiagnostic("KS259", "External Table Partition Column does not support this function.")
}

func DiagnosticPartitionColumnNotUsedInPathFormat(partitionName string) Diagnostic {
	return NewDiagnostic("KS264", fmt.Sprintf("External Table Partition %s is not used in pathformat.", partitionName))
}

func DiagnosticPartitionColumnCanNotBeUsedBothDirectlyAndPattern(partitionName string) Diagnostic {
	return NewDiagnostic("KS265", fmt.Sprintf("External Table Partition %s can be used with single pattern.", partitionName))
}

func DiagnosticWrongVirtualPartitionColumnType() Diagnostic {
	return NewDiagnostic("KS266", "Virtual column could be either string or datetime.")
}

func DiagnosticWrongDataStreamType(typ string) Diagnostic {
	return NewDiagnostic("KS267", fmt.Sprintf("Data stream type %s is not supported by external table ", typ))
}

func DiagnosticWildcardedNamesNotAllowedInThisContext() Diagnostic {
	return NewDiagnostic("KS268", "Wildcarded names not allowed in this context.")
}

func DiagnosticInlineExternalTableDeltaDataFormatNotSupported() Diagnostic {
	return NewDiagnostic("KS269", "Inline external table kind=delta does not support dataformat.")
}

func DiagnosticInlineExternalTableDeltaPartitionByNotSupported() Diagnostic {
	return NewDiagnostic("KS270", "Inline external table kind=delta does not support partition by.")
}

func DiagnosticInlineExternalTableDeltaPathFormatNotSupported() Diagnostic {
	return NewDiagnostic("KS271", "Inline external table kind=delta does not support pathformat.")
}

func DiagnosticInlineExternalTableDeltaRequiresSingleRootURI() Diagnostic {
	return NewDiagnostic("KS272", "Inline external table kind=delta requires exactly one root URI.")
}

func DiagnosticFunctionExpectsArgumentCountExact(functionName string, count int) Diagnostic {
	switch count {
	case 0:
		return NewDiagnostic("KS119", fmt.Sprintf("The function '%s' expects no arguments.", functionName))
	case 1:
		return NewDiagnostic("KS119", fmt.Sprintf("The function '%s' expects 1 argument.", functionName))
	default:
		return NewDiagnostic("KS119", fmt.Sprintf("The function '%s' expects %d arguments.", functionName, count))
	}
}

func DiagnosticFunctionExpectsArgumentCountRange(functionName string, min, max int) Diagnostic {
	if min == max {
		return DiagnosticFunctionExpectsArgumentCountExact(functionName, min)
	}
	return NewDiagnostic("KS120", fmt.Sprintf("The function '%s' expects between %d and %d arguments.", functionName, min, max))
}

func DiagnosticFunctionHasIncorrectNumberOfArguments() Diagnostic {
	return NewDiagnostic("KS121", "The function call has an incorrect number of arguments.")
}

func DiagnosticIntegerLiteralExpected() Diagnostic {
	return NewDiagnostic("KS221", "Integer literal expected.")
}

func DiagnosticStringLiteralExpected() Diagnostic {
	return NewDiagnostic("KS222", "String literal expected.")
}

func DiagnosticBooleanLiteralExpected() Diagnostic {
	return NewDiagnostic("KS223", "Boolean literal expected.")
}

func DiagnosticSummableLiteralExpected() Diagnostic {
	return NewDiagnostic("KS224", "Summable literal expected.")
}

func DiagnosticNumericLiteralExpected() Diagnostic {
	return NewDiagnostic("KS225", "Numeric literal expected.")
}

func DiagnosticScalarLiteralExpected() Diagnostic {
	return NewDiagnostic("KS226", "Scalar literal expected.")
}

func DiagnosticTabularParametersMustBeDeclaredFirst() Diagnostic {
	return NewDiagnostic("KS226", "Tabular parameters must be declared first.")
}

func DiagnosticRawGuidLiteralNotAllowed() Diagnostic {
	return NewDiagnostic("KS214", "Raw guid literals are not allowed in this context, use guid(...) instead.")
}

func DiagnosticDecimalInDynamic() Diagnostic {
	return NewDiagnostic("KS215", "Decimal values are not supported in dynamic objects.")
}

func DiagnosticClusterDatabaseOrTableExpected() Diagnostic {
	return NewDiagnostic("KS216", "Cluster, Database or Table expected.")
}

func DiagnosticNameDoesNotReferToAnyKnownEntityGroup(name string) Diagnostic {
	return NewDiagnostic("KS247", fmt.Sprintf("The name '%s' does not refer to any known entity group.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownItem(name string, entities ...string) Diagnostic {
	if len(entities) > 1 {
		panic("DiagnosticNameDoesNotReferToAnyKnownItem accepts at most one entity")
	}
	entity := ""
	if len(entities) == 1 {
		entity = entities[0]
	}
	if entity == "" {
		entity = "column, table, variable or function"
	}
	return NewDiagnostic("KS142", fmt.Sprintf("The name '%s' does not refer to any known %s.", name, entity)).WithLocationText(name)
}

func DiagnosticNameRefersToMoreThanOneItem(name string) Diagnostic {
	return NewDiagnostic("KS146", fmt.Sprintf("The name '%s' refers to more than one column or variable", name))
}

func DiagnosticNameDoesNotReferToTable(name string) Diagnostic {
	return NewDiagnostic("KS181", fmt.Sprintf("The name '%s' does not refer to a table.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownDatabase(name string) Diagnostic {
	return NewDiagnostic("KS208", fmt.Sprintf("The name '%s' does not refer to any known database.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownCluster(name string) Diagnostic {
	return NewDiagnostic("KS207", fmt.Sprintf("The name '%s' either does not refer to a reachable cluster or no schema from it is currently available.", name)).
		WithSeverity(DiagnosticSeverityWarning)
}

func DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name string) Diagnostic {
	return NewDiagnostic("KS204", fmt.Sprintf("The name '%s' does not refer to any known table, tabular variable or function.", name)).WithLocationText(name)
}

func DiagnosticFuzzyUnionOperandDoesNotReferToAnyKnownTable(name string) Diagnostic {
	return NewDiagnostic("KS205", fmt.Sprintf("The fuzzy union operand '%s' does not refer to any known table, tabular variable or function.", name)).
		WithSeverity(DiagnosticSeverityWarning).
		WithLocationText(name)
}

func DiagnosticFuzzyTableNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("table", names)
}

func DiagnosticFuzzyEntityNotDefined(nameAndKind ...string) Diagnostic {
	if len(nameAndKind) > 2 {
		panic("DiagnosticFuzzyEntityNotDefined accepts at most a name and kind")
	}
	name := ""
	kind := "entity"
	if len(nameAndKind) > 0 {
		name = nameAndKind[0]
	}
	if len(nameAndKind) > 1 {
		kind = nameAndKind[1]
	}
	message := fmt.Sprintf("The fuzzy expression does not refer to any known or currently accessible %s.", kind)
	if name != "" {
		message = fmt.Sprintf("The fuzzy name '%s' does not refer to any known or currently accessible %s.", name, kind)
	}
	return NewDiagnostic("KS205", message).
		WithSeverity(DiagnosticSeverityWarning).
		WithLocationText(name)
}

func DiagnosticFuzzyClusterNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("cluster", names)
}

func DiagnosticFuzzyDatabaseNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("database", names)
}

func DiagnosticFuzzyExternalTableNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("external table", names)
}

func DiagnosticFuzzyMaterializedViewNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("materialized view", names)
}

func DiagnosticFuzzyEntityGroupNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("entity group", names)
}

func DiagnosticFuzzyFunctionNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("function", names)
}

func DiagnosticFuzzyStoredQueryResultNotDefined(names ...string) Diagnostic {
	return diagnosticFuzzyKindNotDefined("stored query result", names)
}

func diagnosticFuzzyKindNotDefined(kind string, names []string) Diagnostic {
	if len(names) > 1 {
		panic("fuzzy diagnostic factory accepts at most one name")
	}
	name := ""
	if len(names) == 1 {
		name = names[0]
	}
	return DiagnosticFuzzyEntityNotDefined(name, kind)
}

func DiagnosticNameDoesNotReferToAnyKnownExternalTable(name string) Diagnostic {
	return NewDiagnostic("KS209", fmt.Sprintf("The name '%s' does not refer to any known external table.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownMaterializedView(name string) Diagnostic {
	return NewDiagnostic("KS210", fmt.Sprintf("The name '%s' does not refer to any known materialized view.", name))
}

func DiagnosticTableOrGraphExpected() Diagnostic {
	return NewDiagnostic("KS231", "Table or graph expected.")
}

func DiagnosticVariableAlreadyDeclared(name string) Diagnostic {
	return NewDiagnostic("KS201", fmt.Sprintf("A variable with the name '%s' has already been declared.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownFunction(name string) Diagnostic {
	return NewDiagnostic("KS211", fmt.Sprintf("The name '%s' does not refer to any known function.", name)).WithLocationText(name)
}

func DiagnosticNameDoesNotReferToAnyKnownStoredQueryResult(name string) Diagnostic {
	return NewDiagnostic("KS248", fmt.Sprintf("The name '%s' does not refer to any known stored query result.", name))
}

func DiagnosticTabularValueDoesNotHaveRequiredColumns() Diagnostic {
	return NewDiagnostic("KS249", "The tabular value does not have the required columns.")
}

func DiagnosticNameDoesNotReferToAnyKnownGraphModel(name string) Diagnostic {
	return NewDiagnostic("KS260", fmt.Sprintf("The name '%s' does not refer to any known graph model.", name))
}

func DiagnosticNameDoesNotReferToAnyKnownGraphSnapshot(name string, graphModelName string) Diagnostic {
	return NewDiagnostic("KS261", fmt.Sprintf("The name '%s' does not refer to any known graph snapshot of graph model '%s'.", name, graphModelName))
}

// The following source-named factories complete the pinned public
// DiagnosticFacts surface. Some retain compatibility upstream spelling.
func DiagnosticTokenExpected(tokens ...string) Diagnostic {
	return DiagnosticTermsExpected(tokens...)
}

func DiagnosticMissingFunctionDeclaration() Diagnostic {
	return NewDiagnostic("KS006", "Missing function declaration")
}

func DiagnosticMissingAllLastOrNone() Diagnostic {
	return NewDiagnostic("KS006", "Missing all, last or none")
}

func DiagnosticSingleColumnTableExpected() Diagnostic {
	return NewDiagnostic("KS113", "A tabular value with only one column expected.")
}

func DiagnosticScalarFunctionNotDefined(name string) Diagnostic {
	return NewDiagnostic("KS122", fmt.Sprintf("The scalar function '%s' is not defined.", name))
}

func DiagnosticAggregateFunctionNotDefined(name string) Diagnostic {
	return NewDiagnostic("KS123", fmt.Sprintf("The aggregate function '%s' is not defined.", name))
}

func DiagnosticPlugInFunctionNotDefined(name string) Diagnostic {
	return DiagnosticPluginFunctionNotDefined(name)
}

func DiagnosticPlugInFunctionIsNotEnabled(name string) Diagnostic {
	return DiagnosticPluginFunctionIsNotEnabled(name)
}

func DiagnosticNameIsNotAFunction(name string) Diagnostic {
	return NewDiagnostic("KS128", fmt.Sprintf("The name '%s' does not refer to a function.", name))
}

func DiagnosticFunctionNotDefined(name string) Diagnostic {
	return NewDiagnostic("KS143", fmt.Sprintf("The function '%s' is not defined.", name))
}

func DiagnosticTheElementAccessOperatorIsNotAllowedInThisContext() Diagnostic {
	return DiagnosticElementAccessNotAllowedInContext()
}

func DiagnosticTheExpressionHasNoName() Diagnostic {
	return NewDiagnostic("KS148", "A column name cannot be inferred for this expression.")
}

func DiagnosticTheExpressionDoesNotHaveMultipleValues() Diagnostic {
	return DiagnosticExpressionDoesNotHaveMultipleValues()
}

func DiagnosticTheNameDoesNotHaveCorrespondingExpression() Diagnostic {
	return DiagnosticNameDoesNotHaveCorrespondingExpression()
}

func DiagnosticTheExpressionRefersToMoreThanOneColumn() Diagnostic {
	return NewDiagnostic("KS184", "The expression refers to more than one column.")
}

func DiagnosticDefaultValueExpected() Diagnostic {
	return NewDiagnostic("KS189", "Default value expected.")
}

func DiagnosticUnnamedArgumentAfterOutofOrderNamedArgument() Diagnostic {
	return DiagnosticUnnamedArgumentAfterOutOfOrderNamedArgument()
}

func DiagnosticMaterializedViewNameMustBeStringLiteral() Diagnostic {
	return NewDiagnostic("KS202", "Materialized view name must be a string literal")
}

func DiagnosticNameDoesNotReferToAnyKnownTable(name string) Diagnostic {
	return NewDiagnostic("KS204", fmt.Sprintf("The name '%s' does not refer to any known table, tabular variable or function.", name))
}

func DiagnosticRealLiteralExpected() Diagnostic {
	return NewDiagnostic("KS225", "Real literal expected.")
}

func DiagnosticMakeGraphDynamicNodeIdColumnNotSupported() Diagnostic {
	return DiagnosticMakeGraphDynamicNodeIDColumnNotSupported()
}

func DiagnosticMakeGraphImplicityIdShouldNotBeEmpty() Diagnostic {
	return DiagnosticMakeGraphImplicitIDShouldNotBeEmpty()
}

func DiagnosticLatestSnapshotUnknown(graphModelName string) Diagnostic {
	return NewDiagnostic("KS262", fmt.Sprintf("The lastest snapshot of graph model '%s' is unknown.", graphModelName))
}

func DiagnosticVolatileGraphUnknown(graphModelName string) Diagnostic {
	return NewDiagnostic("KS263", fmt.Sprintf("The volatile graph of model '%s' is unknown.", graphModelName))
}

func DiagnosticInlineExternalTableDeltaRequiresSingleRootUri() Diagnostic {
	return DiagnosticInlineExternalTableDeltaRequiresSingleRootURI()
}

func joinEnglish(values []string, separator, finalSeparator string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + finalSeparator + values[1]
	default:
		out := ""
		for i, value := range values {
			if i > 0 {
				if i == len(values)-1 {
					out += finalSeparator
				} else {
					out += separator
				}
			}
			out += value
		}
		return out
	}
}
