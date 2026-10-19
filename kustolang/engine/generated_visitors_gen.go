package kustolang

// Code generated from the pinned upstream Kusto.Language sources; DO NOT EDIT.

const generatedSyntaxVisitorMethodCount = 213

type SyntaxVisitor interface {
	VisitCustom(*CustomNode)
	VisitList(*SyntaxList)
	VisitSeparatedElement(*SeparatedElement)
	VisitQueryBlock(*QueryBlock)
	VisitDirectiveBlock(*DirectiveBlock)
	VisitDirective(*Directive)
	VisitSkippedTokens(*SkippedTokens)
	VisitTypeOfLiteralExpression(*TypeOfLiteralExpression)
	VisitBadQueryOperator(*BadQueryOperator)
	VisitCompoundStringLiteralExpression(*CompoundStringLiteralExpression)
	VisitTokenName(*TokenName)
	VisitBracketedName(*BracketedName)
	VisitBracedName(*BracedName)
	VisitWildcardedName(*WildcardedName)
	VisitBracketedWildcardedName(*BracketedWildcardedName)
	VisitNameDeclaration(*NameDeclaration)
	VisitNameReference(*NameReference)
	VisitLiteralExpression(*LiteralExpression)
	VisitStarExpression(*StarExpression)
	VisitAtExpression(*AtExpression)
	VisitJsonPair(*JsonPair)
	VisitJsonObjectExpression(*JsonObjectExpression)
	VisitJsonArrayExpression(*JsonArrayExpression)
	VisitDynamicExpression(*DynamicExpression)
	VisitParenthesizedExpression(*ParenthesizedExpression)
	VisitExpressionList(*ExpressionList)
	VisitExpressionCouple(*ExpressionCouple)
	VisitPrefixUnaryExpression(*PrefixUnaryExpression)
	VisitBinaryExpression(*BinaryExpression)
	VisitInExpression(*InExpression)
	VisitHasAnyExpression(*HasAnyExpression)
	VisitHasAllExpression(*HasAllExpression)
	VisitBetweenExpression(*BetweenExpression)
	VisitFunctionCallExpression(*FunctionCallExpression)
	VisitToScalarExpression(*ToScalarExpression)
	VisitToTableExpression(*ToTableExpression)
	VisitMaterializedViewCombineExpression(*MaterializedViewCombineExpression)
	VisitMaterializedViewCombineNameClause(*MaterializedViewCombineNameClause)
	VisitMaterializedViewCombineClause(*MaterializedViewCombineClause)
	VisitSimpleNamedExpression(*SimpleNamedExpression)
	VisitRenameList(*RenameList)
	VisitCompoundNamedExpression(*CompoundNamedExpression)
	VisitBracketedExpression(*BracketedExpression)
	VisitPathExpression(*PathExpression)
	VisitElementExpression(*ElementExpression)
	VisitPipeExpression(*PipeExpression)
	VisitRangeOperator(*RangeOperator)
	VisitNamedParameter(*NamedParameter)
	VisitConsumeOperator(*ConsumeOperator)
	VisitCountOperator(*CountOperator)
	VisitCountAsIdentifierClause(*CountAsIdentifierClause)
	VisitExecuteAndCacheOperator(*ExecuteAndCacheOperator)
	VisitExtendOperator(*ExtendOperator)
	VisitFacetOperator(*FacetOperator)
	VisitFacetWithOperatorClause(*FacetWithOperatorClause)
	VisitFacetWithExpressionClause(*FacetWithExpressionClause)
	VisitFilterOperator(*FilterOperator)
	VisitGetSchemaOperator(*GetSchemaOperator)
	VisitFindOperator(*FindOperator)
	VisitDataScopeClause(*DataScopeClause)
	VisitTypedColumnReference(*TypedColumnReference)
	VisitFindInClause(*FindInClause)
	VisitFindProjectClause(*FindProjectClause)
	VisitPackExpression(*PackExpression)
	VisitNameAndTypeDeclaration(*NameAndTypeDeclaration)
	VisitPrimitiveTypeExpression(*PrimitiveTypeExpression)
	VisitSearchOperator(*SearchOperator)
	VisitForkOperator(*ForkOperator)
	VisitNameEqualsClause(*NameEqualsClause)
	VisitForkExpression(*ForkExpression)
	VisitMakeSeriesOperator(*MakeSeriesOperator)
	VisitMakeSeriesExpression(*MakeSeriesExpression)
	VisitDefaultExpressionClause(*DefaultExpressionClause)
	VisitMakeSeriesOnClause(*MakeSeriesOnClause)
	VisitMakeSeriesFromClause(*MakeSeriesFromClause)
	VisitMakeSeriesToClause(*MakeSeriesToClause)
	VisitMakeSeriesStepClause(*MakeSeriesStepClause)
	VisitMakeSeriesInRangeClause(*MakeSeriesInRangeClause)
	VisitMakeSeriesFromToStepClause(*MakeSeriesFromToStepClause)
	VisitMakeSeriesByClause(*MakeSeriesByClause)
	VisitMvExpandOperator(*MvExpandOperator)
	VisitMvExpandExpression(*MvExpandExpression)
	VisitMvExpandRowLimitClause(*MvExpandRowLimitClause)
	VisitMvApplyOperator(*MvApplyOperator)
	VisitMvApplyExpression(*MvApplyExpression)
	VisitMvApplyRowLimitClause(*MvApplyRowLimitClause)
	VisitMvApplyContextIdClause(*MvApplyContextIdClause)
	VisitMvApplySubqueryExpression(*MvApplySubqueryExpression)
	VisitToTypeOfClause(*ToTypeOfClause)
	VisitEvaluateSchemaClause(*EvaluateSchemaClause)
	VisitEvaluateOperator(*EvaluateOperator)
	VisitParseOperator(*ParseOperator)
	VisitParseWhereOperator(*ParseWhereOperator)
	VisitParseKvWithClause(*ParseKvWithClause)
	VisitParseKvOperator(*ParseKvOperator)
	VisitPartitionOperator(*PartitionOperator)
	VisitPartitionQuery(*PartitionQuery)
	VisitPartitionScope(*PartitionScope)
	VisitPartitionSubquery(*PartitionSubquery)
	VisitProjectOperator(*ProjectOperator)
	VisitProjectAwayOperator(*ProjectAwayOperator)
	VisitProjectByNamesOperator(*ProjectByNamesOperator)
	VisitProjectKeepOperator(*ProjectKeepOperator)
	VisitProjectRenameOperator(*ProjectRenameOperator)
	VisitProjectReorderOperator(*ProjectReorderOperator)
	VisitSampleOperator(*SampleOperator)
	VisitSampleDistinctOperator(*SampleDistinctOperator)
	VisitEntityGroup(*EntityGroup)
	VisitReduceByOperator(*ReduceByOperator)
	VisitReduceByWithClause(*ReduceByWithClause)
	VisitSummarizeOperator(*SummarizeOperator)
	VisitMacroExpandScopeReferenceName(*MacroExpandScopeReferenceName)
	VisitMacroExpandOperator(*MacroExpandOperator)
	VisitSummarizeByClause(*SummarizeByClause)
	VisitDistinctOperator(*DistinctOperator)
	VisitTakeOperator(*TakeOperator)
	VisitSortOperator(*SortOperator)
	VisitOrderedExpression(*OrderedExpression)
	VisitOrderingClause(*OrderingClause)
	VisitOrderingNullsClause(*OrderingNullsClause)
	VisitTopHittersOperator(*TopHittersOperator)
	VisitTopHittersByClause(*TopHittersByClause)
	VisitTopOperator(*TopOperator)
	VisitTopNestedOperator(*TopNestedOperator)
	VisitTopNestedClause(*TopNestedClause)
	VisitTopNestedWithOthersClause(*TopNestedWithOthersClause)
	VisitUnionOperator(*UnionOperator)
	VisitAsOperator(*AsOperator)
	VisitSerializeOperator(*SerializeOperator)
	VisitInvokeOperator(*InvokeOperator)
	VisitRenderOperator(*RenderOperator)
	VisitMakeGraphOperator(*MakeGraphOperator)
	VisitMakeGraphWithTablesAndKeysClause(*MakeGraphWithTablesAndKeysClause)
	VisitMakeGraphWithImplicitIdClause(*MakeGraphWithImplicitIdClause)
	VisitGraphMarkComponentsOperator(*GraphMarkComponentsOperator)
	VisitGraphWhereNodesOperator(*GraphWhereNodesOperator)
	VisitGraphWhereEdgesOperator(*GraphWhereEdgesOperator)
	VisitMakeGraphTableAndKeyClause(*MakeGraphTableAndKeyClause)
	VisitMakeGraphPartitionedByClause(*MakeGraphPartitionedByClause)
	VisitGraphToTableOperator(*GraphToTableOperator)
	VisitGraphToTableOutputClause(*GraphToTableOutputClause)
	VisitGraphToTableAsClause(*GraphToTableAsClause)
	VisitGraphMatchOperator(*GraphMatchOperator)
	VisitGraphShortestPathsOperator(*GraphShortestPathsOperator)
	VisitGraphMatchPattern(*GraphMatchPattern)
	VisitGraphMatchPatternNode(*GraphMatchPatternNode)
	VisitGraphMatchPatternEdge(*GraphMatchPatternEdge)
	VisitGraphMatchPatternEdgeRange(*GraphMatchPatternEdgeRange)
	VisitWhereClause(*WhereClause)
	VisitProjectClause(*ProjectClause)
	VisitNameReferenceList(*NameReferenceList)
	VisitRenderWithClause(*RenderWithClause)
	VisitPrintOperator(*PrintOperator)
	VisitAliasStatement(*AliasStatement)
	VisitLetStatement(*LetStatement)
	VisitFunctionDeclaration(*FunctionDeclaration)
	VisitFunctionParameters(*FunctionParameters)
	VisitFunctionParameter(*FunctionParameter)
	VisitDefaultValueDeclaration(*DefaultValueDeclaration)
	VisitFunctionBody(*FunctionBody)
	VisitSchemaTypeExpression(*SchemaTypeExpression)
	VisitExpressionStatement(*ExpressionStatement)
	VisitMaterializeExpression(*MaterializeExpression)
	VisitSetOptionStatement(*SetOptionStatement)
	VisitOptionValueClause(*OptionValueClause)
	VisitQueryParametersStatement(*QueryParametersStatement)
	VisitRestrictStatementWithClause(*RestrictStatementWithClause)
	VisitRestrictStatement(*RestrictStatement)
	VisitPatternStatement(*PatternStatement)
	VisitPatternDeclaration(*PatternDeclaration)
	VisitPatternPathParameter(*PatternPathParameter)
	VisitPatternMatch(*PatternMatch)
	VisitPatternPathValue(*PatternPathValue)
	VisitDataScopeExpression(*DataScopeExpression)
	VisitDataTableExpression(*DataTableExpression)
	VisitRowSchema(*RowSchema)
	VisitEvaluateRowSchema(*EvaluateRowSchema)
	VisitExternalDataExpression(*ExternalDataExpression)
	VisitContextualDataTableExpression(*ContextualDataTableExpression)
	VisitExternalDataWithClause(*ExternalDataWithClause)
	VisitInlineExternalTableKindClause(*InlineExternalTableKindClause)
	VisitInlineExternalTableDataFormatClause(*InlineExternalTableDataFormatClause)
	VisitDateTimePattern(*DateTimePattern)
	VisitInlineExternalTablePathFormatPartitionColumnReference(*InlineExternalTablePathFormatPartitionColumnReference)
	VisitInlineExternalTablePathFormatClause(*InlineExternalTablePathFormatClause)
	VisitPartitionColumnDeclaration(*PartitionColumnDeclaration)
	VisitInlineExternalTablePartitionClause(*InlineExternalTablePartitionClause)
	VisitInlineExternalTableConnectionStringsClause(*InlineExternalTableConnectionStringsClause)
	VisitInlineExternalTableExpression(*InlineExternalTableExpression)
	VisitJoinOperator(*JoinOperator)
	VisitLookupOperator(*LookupOperator)
	VisitJoinOnClause(*JoinOnClause)
	VisitJoinWhereClause(*JoinWhereClause)
	VisitScanOperator(*ScanOperator)
	VisitAssertSchemaOperator(*AssertSchemaOperator)
	VisitScanDeclareClause(*ScanDeclareClause)
	VisitScanOrderByClause(*ScanOrderByClause)
	VisitScanPartitionByClause(*ScanPartitionByClause)
	VisitScanStepOutput(*ScanStepOutput)
	VisitScanStep(*ScanStep)
	VisitScanComputationClause(*ScanComputationClause)
	VisitScanAssignment(*ScanAssignment)
	VisitPartitionByOperator(*PartitionByOperator)
	VisitPartitionByIdClause(*PartitionByIdClause)
	VisitCommandWithValueClause(*CommandWithValueClause)
	VisitCommandWithPropertyListClause(*CommandWithPropertyListClause)
	VisitUnknownCommand(*UnknownCommand)
	VisitCustomCommand(*CustomCommand)
	VisitPartialCommand(*PartialCommand)
	VisitCommandAndSkippedTokens(*CommandAndSkippedTokens)
	VisitBadCommand(*BadCommand)
	VisitCommandBlock(*CommandBlock)
	VisitGraphMergeOperator(*GraphMergeOperator)
}

type SyntaxVisitorResult[T any] interface {
	VisitCustom(*CustomNode) T
	VisitList(*SyntaxList) T
	VisitSeparatedElement(*SeparatedElement) T
	VisitQueryBlock(*QueryBlock) T
	VisitDirectiveBlock(*DirectiveBlock) T
	VisitDirective(*Directive) T
	VisitSkippedTokens(*SkippedTokens) T
	VisitTypeOfLiteralExpression(*TypeOfLiteralExpression) T
	VisitBadQueryOperator(*BadQueryOperator) T
	VisitCompoundStringLiteralExpression(*CompoundStringLiteralExpression) T
	VisitTokenName(*TokenName) T
	VisitBracketedName(*BracketedName) T
	VisitBracedName(*BracedName) T
	VisitWildcardedName(*WildcardedName) T
	VisitBracketedWildcardedName(*BracketedWildcardedName) T
	VisitNameDeclaration(*NameDeclaration) T
	VisitNameReference(*NameReference) T
	VisitLiteralExpression(*LiteralExpression) T
	VisitStarExpression(*StarExpression) T
	VisitAtExpression(*AtExpression) T
	VisitJsonPair(*JsonPair) T
	VisitJsonObjectExpression(*JsonObjectExpression) T
	VisitJsonArrayExpression(*JsonArrayExpression) T
	VisitDynamicExpression(*DynamicExpression) T
	VisitParenthesizedExpression(*ParenthesizedExpression) T
	VisitExpressionList(*ExpressionList) T
	VisitExpressionCouple(*ExpressionCouple) T
	VisitPrefixUnaryExpression(*PrefixUnaryExpression) T
	VisitBinaryExpression(*BinaryExpression) T
	VisitInExpression(*InExpression) T
	VisitHasAnyExpression(*HasAnyExpression) T
	VisitHasAllExpression(*HasAllExpression) T
	VisitBetweenExpression(*BetweenExpression) T
	VisitFunctionCallExpression(*FunctionCallExpression) T
	VisitToScalarExpression(*ToScalarExpression) T
	VisitToTableExpression(*ToTableExpression) T
	VisitMaterializedViewCombineExpression(*MaterializedViewCombineExpression) T
	VisitMaterializedViewCombineNameClause(*MaterializedViewCombineNameClause) T
	VisitMaterializedViewCombineClause(*MaterializedViewCombineClause) T
	VisitSimpleNamedExpression(*SimpleNamedExpression) T
	VisitRenameList(*RenameList) T
	VisitCompoundNamedExpression(*CompoundNamedExpression) T
	VisitBracketedExpression(*BracketedExpression) T
	VisitPathExpression(*PathExpression) T
	VisitElementExpression(*ElementExpression) T
	VisitPipeExpression(*PipeExpression) T
	VisitRangeOperator(*RangeOperator) T
	VisitNamedParameter(*NamedParameter) T
	VisitConsumeOperator(*ConsumeOperator) T
	VisitCountOperator(*CountOperator) T
	VisitCountAsIdentifierClause(*CountAsIdentifierClause) T
	VisitExecuteAndCacheOperator(*ExecuteAndCacheOperator) T
	VisitExtendOperator(*ExtendOperator) T
	VisitFacetOperator(*FacetOperator) T
	VisitFacetWithOperatorClause(*FacetWithOperatorClause) T
	VisitFacetWithExpressionClause(*FacetWithExpressionClause) T
	VisitFilterOperator(*FilterOperator) T
	VisitGetSchemaOperator(*GetSchemaOperator) T
	VisitFindOperator(*FindOperator) T
	VisitDataScopeClause(*DataScopeClause) T
	VisitTypedColumnReference(*TypedColumnReference) T
	VisitFindInClause(*FindInClause) T
	VisitFindProjectClause(*FindProjectClause) T
	VisitPackExpression(*PackExpression) T
	VisitNameAndTypeDeclaration(*NameAndTypeDeclaration) T
	VisitPrimitiveTypeExpression(*PrimitiveTypeExpression) T
	VisitSearchOperator(*SearchOperator) T
	VisitForkOperator(*ForkOperator) T
	VisitNameEqualsClause(*NameEqualsClause) T
	VisitForkExpression(*ForkExpression) T
	VisitMakeSeriesOperator(*MakeSeriesOperator) T
	VisitMakeSeriesExpression(*MakeSeriesExpression) T
	VisitDefaultExpressionClause(*DefaultExpressionClause) T
	VisitMakeSeriesOnClause(*MakeSeriesOnClause) T
	VisitMakeSeriesFromClause(*MakeSeriesFromClause) T
	VisitMakeSeriesToClause(*MakeSeriesToClause) T
	VisitMakeSeriesStepClause(*MakeSeriesStepClause) T
	VisitMakeSeriesInRangeClause(*MakeSeriesInRangeClause) T
	VisitMakeSeriesFromToStepClause(*MakeSeriesFromToStepClause) T
	VisitMakeSeriesByClause(*MakeSeriesByClause) T
	VisitMvExpandOperator(*MvExpandOperator) T
	VisitMvExpandExpression(*MvExpandExpression) T
	VisitMvExpandRowLimitClause(*MvExpandRowLimitClause) T
	VisitMvApplyOperator(*MvApplyOperator) T
	VisitMvApplyExpression(*MvApplyExpression) T
	VisitMvApplyRowLimitClause(*MvApplyRowLimitClause) T
	VisitMvApplyContextIdClause(*MvApplyContextIdClause) T
	VisitMvApplySubqueryExpression(*MvApplySubqueryExpression) T
	VisitToTypeOfClause(*ToTypeOfClause) T
	VisitEvaluateSchemaClause(*EvaluateSchemaClause) T
	VisitEvaluateOperator(*EvaluateOperator) T
	VisitParseOperator(*ParseOperator) T
	VisitParseWhereOperator(*ParseWhereOperator) T
	VisitParseKvWithClause(*ParseKvWithClause) T
	VisitParseKvOperator(*ParseKvOperator) T
	VisitPartitionOperator(*PartitionOperator) T
	VisitPartitionQuery(*PartitionQuery) T
	VisitPartitionScope(*PartitionScope) T
	VisitPartitionSubquery(*PartitionSubquery) T
	VisitProjectOperator(*ProjectOperator) T
	VisitProjectAwayOperator(*ProjectAwayOperator) T
	VisitProjectByNamesOperator(*ProjectByNamesOperator) T
	VisitProjectKeepOperator(*ProjectKeepOperator) T
	VisitProjectRenameOperator(*ProjectRenameOperator) T
	VisitProjectReorderOperator(*ProjectReorderOperator) T
	VisitSampleOperator(*SampleOperator) T
	VisitSampleDistinctOperator(*SampleDistinctOperator) T
	VisitEntityGroup(*EntityGroup) T
	VisitReduceByOperator(*ReduceByOperator) T
	VisitReduceByWithClause(*ReduceByWithClause) T
	VisitSummarizeOperator(*SummarizeOperator) T
	VisitMacroExpandScopeReferenceName(*MacroExpandScopeReferenceName) T
	VisitMacroExpandOperator(*MacroExpandOperator) T
	VisitSummarizeByClause(*SummarizeByClause) T
	VisitDistinctOperator(*DistinctOperator) T
	VisitTakeOperator(*TakeOperator) T
	VisitSortOperator(*SortOperator) T
	VisitOrderedExpression(*OrderedExpression) T
	VisitOrderingClause(*OrderingClause) T
	VisitOrderingNullsClause(*OrderingNullsClause) T
	VisitTopHittersOperator(*TopHittersOperator) T
	VisitTopHittersByClause(*TopHittersByClause) T
	VisitTopOperator(*TopOperator) T
	VisitTopNestedOperator(*TopNestedOperator) T
	VisitTopNestedClause(*TopNestedClause) T
	VisitTopNestedWithOthersClause(*TopNestedWithOthersClause) T
	VisitUnionOperator(*UnionOperator) T
	VisitAsOperator(*AsOperator) T
	VisitSerializeOperator(*SerializeOperator) T
	VisitInvokeOperator(*InvokeOperator) T
	VisitRenderOperator(*RenderOperator) T
	VisitMakeGraphOperator(*MakeGraphOperator) T
	VisitMakeGraphWithTablesAndKeysClause(*MakeGraphWithTablesAndKeysClause) T
	VisitMakeGraphWithImplicitIdClause(*MakeGraphWithImplicitIdClause) T
	VisitGraphMarkComponentsOperator(*GraphMarkComponentsOperator) T
	VisitGraphWhereNodesOperator(*GraphWhereNodesOperator) T
	VisitGraphWhereEdgesOperator(*GraphWhereEdgesOperator) T
	VisitMakeGraphTableAndKeyClause(*MakeGraphTableAndKeyClause) T
	VisitMakeGraphPartitionedByClause(*MakeGraphPartitionedByClause) T
	VisitGraphToTableOperator(*GraphToTableOperator) T
	VisitGraphToTableOutputClause(*GraphToTableOutputClause) T
	VisitGraphToTableAsClause(*GraphToTableAsClause) T
	VisitGraphMatchOperator(*GraphMatchOperator) T
	VisitGraphShortestPathsOperator(*GraphShortestPathsOperator) T
	VisitGraphMatchPattern(*GraphMatchPattern) T
	VisitGraphMatchPatternNode(*GraphMatchPatternNode) T
	VisitGraphMatchPatternEdge(*GraphMatchPatternEdge) T
	VisitGraphMatchPatternEdgeRange(*GraphMatchPatternEdgeRange) T
	VisitWhereClause(*WhereClause) T
	VisitProjectClause(*ProjectClause) T
	VisitNameReferenceList(*NameReferenceList) T
	VisitRenderWithClause(*RenderWithClause) T
	VisitPrintOperator(*PrintOperator) T
	VisitAliasStatement(*AliasStatement) T
	VisitLetStatement(*LetStatement) T
	VisitFunctionDeclaration(*FunctionDeclaration) T
	VisitFunctionParameters(*FunctionParameters) T
	VisitFunctionParameter(*FunctionParameter) T
	VisitDefaultValueDeclaration(*DefaultValueDeclaration) T
	VisitFunctionBody(*FunctionBody) T
	VisitSchemaTypeExpression(*SchemaTypeExpression) T
	VisitExpressionStatement(*ExpressionStatement) T
	VisitMaterializeExpression(*MaterializeExpression) T
	VisitSetOptionStatement(*SetOptionStatement) T
	VisitOptionValueClause(*OptionValueClause) T
	VisitQueryParametersStatement(*QueryParametersStatement) T
	VisitRestrictStatementWithClause(*RestrictStatementWithClause) T
	VisitRestrictStatement(*RestrictStatement) T
	VisitPatternStatement(*PatternStatement) T
	VisitPatternDeclaration(*PatternDeclaration) T
	VisitPatternPathParameter(*PatternPathParameter) T
	VisitPatternMatch(*PatternMatch) T
	VisitPatternPathValue(*PatternPathValue) T
	VisitDataScopeExpression(*DataScopeExpression) T
	VisitDataTableExpression(*DataTableExpression) T
	VisitRowSchema(*RowSchema) T
	VisitEvaluateRowSchema(*EvaluateRowSchema) T
	VisitExternalDataExpression(*ExternalDataExpression) T
	VisitContextualDataTableExpression(*ContextualDataTableExpression) T
	VisitExternalDataWithClause(*ExternalDataWithClause) T
	VisitInlineExternalTableKindClause(*InlineExternalTableKindClause) T
	VisitInlineExternalTableDataFormatClause(*InlineExternalTableDataFormatClause) T
	VisitDateTimePattern(*DateTimePattern) T
	VisitInlineExternalTablePathFormatPartitionColumnReference(*InlineExternalTablePathFormatPartitionColumnReference) T
	VisitInlineExternalTablePathFormatClause(*InlineExternalTablePathFormatClause) T
	VisitPartitionColumnDeclaration(*PartitionColumnDeclaration) T
	VisitInlineExternalTablePartitionClause(*InlineExternalTablePartitionClause) T
	VisitInlineExternalTableConnectionStringsClause(*InlineExternalTableConnectionStringsClause) T
	VisitInlineExternalTableExpression(*InlineExternalTableExpression) T
	VisitJoinOperator(*JoinOperator) T
	VisitLookupOperator(*LookupOperator) T
	VisitJoinOnClause(*JoinOnClause) T
	VisitJoinWhereClause(*JoinWhereClause) T
	VisitScanOperator(*ScanOperator) T
	VisitAssertSchemaOperator(*AssertSchemaOperator) T
	VisitScanDeclareClause(*ScanDeclareClause) T
	VisitScanOrderByClause(*ScanOrderByClause) T
	VisitScanPartitionByClause(*ScanPartitionByClause) T
	VisitScanStepOutput(*ScanStepOutput) T
	VisitScanStep(*ScanStep) T
	VisitScanComputationClause(*ScanComputationClause) T
	VisitScanAssignment(*ScanAssignment) T
	VisitPartitionByOperator(*PartitionByOperator) T
	VisitPartitionByIdClause(*PartitionByIdClause) T
	VisitCommandWithValueClause(*CommandWithValueClause) T
	VisitCommandWithPropertyListClause(*CommandWithPropertyListClause) T
	VisitUnknownCommand(*UnknownCommand) T
	VisitCustomCommand(*CustomCommand) T
	VisitPartialCommand(*PartialCommand) T
	VisitCommandAndSkippedTokens(*CommandAndSkippedTokens) T
	VisitBadCommand(*BadCommand) T
	VisitCommandBlock(*CommandBlock) T
	VisitGraphMergeOperator(*GraphMergeOperator) T
}

type SyntaxVisitorFuncs struct {
	DefaultVisit                                                   func(SyntaxElement)
	VisitCustomFunc                                                func(*CustomNode)
	VisitListFunc                                                  func(*SyntaxList)
	VisitSeparatedElementFunc                                      func(*SeparatedElement)
	VisitQueryBlockFunc                                            func(*QueryBlock)
	VisitDirectiveBlockFunc                                        func(*DirectiveBlock)
	VisitDirectiveFunc                                             func(*Directive)
	VisitSkippedTokensFunc                                         func(*SkippedTokens)
	VisitTypeOfLiteralExpressionFunc                               func(*TypeOfLiteralExpression)
	VisitBadQueryOperatorFunc                                      func(*BadQueryOperator)
	VisitCompoundStringLiteralExpressionFunc                       func(*CompoundStringLiteralExpression)
	VisitTokenNameFunc                                             func(*TokenName)
	VisitBracketedNameFunc                                         func(*BracketedName)
	VisitBracedNameFunc                                            func(*BracedName)
	VisitWildcardedNameFunc                                        func(*WildcardedName)
	VisitBracketedWildcardedNameFunc                               func(*BracketedWildcardedName)
	VisitNameDeclarationFunc                                       func(*NameDeclaration)
	VisitNameReferenceFunc                                         func(*NameReference)
	VisitLiteralExpressionFunc                                     func(*LiteralExpression)
	VisitStarExpressionFunc                                        func(*StarExpression)
	VisitAtExpressionFunc                                          func(*AtExpression)
	VisitJsonPairFunc                                              func(*JsonPair)
	VisitJsonObjectExpressionFunc                                  func(*JsonObjectExpression)
	VisitJsonArrayExpressionFunc                                   func(*JsonArrayExpression)
	VisitDynamicExpressionFunc                                     func(*DynamicExpression)
	VisitParenthesizedExpressionFunc                               func(*ParenthesizedExpression)
	VisitExpressionListFunc                                        func(*ExpressionList)
	VisitExpressionCoupleFunc                                      func(*ExpressionCouple)
	VisitPrefixUnaryExpressionFunc                                 func(*PrefixUnaryExpression)
	VisitBinaryExpressionFunc                                      func(*BinaryExpression)
	VisitInExpressionFunc                                          func(*InExpression)
	VisitHasAnyExpressionFunc                                      func(*HasAnyExpression)
	VisitHasAllExpressionFunc                                      func(*HasAllExpression)
	VisitBetweenExpressionFunc                                     func(*BetweenExpression)
	VisitFunctionCallExpressionFunc                                func(*FunctionCallExpression)
	VisitToScalarExpressionFunc                                    func(*ToScalarExpression)
	VisitToTableExpressionFunc                                     func(*ToTableExpression)
	VisitMaterializedViewCombineExpressionFunc                     func(*MaterializedViewCombineExpression)
	VisitMaterializedViewCombineNameClauseFunc                     func(*MaterializedViewCombineNameClause)
	VisitMaterializedViewCombineClauseFunc                         func(*MaterializedViewCombineClause)
	VisitSimpleNamedExpressionFunc                                 func(*SimpleNamedExpression)
	VisitRenameListFunc                                            func(*RenameList)
	VisitCompoundNamedExpressionFunc                               func(*CompoundNamedExpression)
	VisitBracketedExpressionFunc                                   func(*BracketedExpression)
	VisitPathExpressionFunc                                        func(*PathExpression)
	VisitElementExpressionFunc                                     func(*ElementExpression)
	VisitPipeExpressionFunc                                        func(*PipeExpression)
	VisitRangeOperatorFunc                                         func(*RangeOperator)
	VisitNamedParameterFunc                                        func(*NamedParameter)
	VisitConsumeOperatorFunc                                       func(*ConsumeOperator)
	VisitCountOperatorFunc                                         func(*CountOperator)
	VisitCountAsIdentifierClauseFunc                               func(*CountAsIdentifierClause)
	VisitExecuteAndCacheOperatorFunc                               func(*ExecuteAndCacheOperator)
	VisitExtendOperatorFunc                                        func(*ExtendOperator)
	VisitFacetOperatorFunc                                         func(*FacetOperator)
	VisitFacetWithOperatorClauseFunc                               func(*FacetWithOperatorClause)
	VisitFacetWithExpressionClauseFunc                             func(*FacetWithExpressionClause)
	VisitFilterOperatorFunc                                        func(*FilterOperator)
	VisitGetSchemaOperatorFunc                                     func(*GetSchemaOperator)
	VisitFindOperatorFunc                                          func(*FindOperator)
	VisitDataScopeClauseFunc                                       func(*DataScopeClause)
	VisitTypedColumnReferenceFunc                                  func(*TypedColumnReference)
	VisitFindInClauseFunc                                          func(*FindInClause)
	VisitFindProjectClauseFunc                                     func(*FindProjectClause)
	VisitPackExpressionFunc                                        func(*PackExpression)
	VisitNameAndTypeDeclarationFunc                                func(*NameAndTypeDeclaration)
	VisitPrimitiveTypeExpressionFunc                               func(*PrimitiveTypeExpression)
	VisitSearchOperatorFunc                                        func(*SearchOperator)
	VisitForkOperatorFunc                                          func(*ForkOperator)
	VisitNameEqualsClauseFunc                                      func(*NameEqualsClause)
	VisitForkExpressionFunc                                        func(*ForkExpression)
	VisitMakeSeriesOperatorFunc                                    func(*MakeSeriesOperator)
	VisitMakeSeriesExpressionFunc                                  func(*MakeSeriesExpression)
	VisitDefaultExpressionClauseFunc                               func(*DefaultExpressionClause)
	VisitMakeSeriesOnClauseFunc                                    func(*MakeSeriesOnClause)
	VisitMakeSeriesFromClauseFunc                                  func(*MakeSeriesFromClause)
	VisitMakeSeriesToClauseFunc                                    func(*MakeSeriesToClause)
	VisitMakeSeriesStepClauseFunc                                  func(*MakeSeriesStepClause)
	VisitMakeSeriesInRangeClauseFunc                               func(*MakeSeriesInRangeClause)
	VisitMakeSeriesFromToStepClauseFunc                            func(*MakeSeriesFromToStepClause)
	VisitMakeSeriesByClauseFunc                                    func(*MakeSeriesByClause)
	VisitMvExpandOperatorFunc                                      func(*MvExpandOperator)
	VisitMvExpandExpressionFunc                                    func(*MvExpandExpression)
	VisitMvExpandRowLimitClauseFunc                                func(*MvExpandRowLimitClause)
	VisitMvApplyOperatorFunc                                       func(*MvApplyOperator)
	VisitMvApplyExpressionFunc                                     func(*MvApplyExpression)
	VisitMvApplyRowLimitClauseFunc                                 func(*MvApplyRowLimitClause)
	VisitMvApplyContextIdClauseFunc                                func(*MvApplyContextIdClause)
	VisitMvApplySubqueryExpressionFunc                             func(*MvApplySubqueryExpression)
	VisitToTypeOfClauseFunc                                        func(*ToTypeOfClause)
	VisitEvaluateSchemaClauseFunc                                  func(*EvaluateSchemaClause)
	VisitEvaluateOperatorFunc                                      func(*EvaluateOperator)
	VisitParseOperatorFunc                                         func(*ParseOperator)
	VisitParseWhereOperatorFunc                                    func(*ParseWhereOperator)
	VisitParseKvWithClauseFunc                                     func(*ParseKvWithClause)
	VisitParseKvOperatorFunc                                       func(*ParseKvOperator)
	VisitPartitionOperatorFunc                                     func(*PartitionOperator)
	VisitPartitionQueryFunc                                        func(*PartitionQuery)
	VisitPartitionScopeFunc                                        func(*PartitionScope)
	VisitPartitionSubqueryFunc                                     func(*PartitionSubquery)
	VisitProjectOperatorFunc                                       func(*ProjectOperator)
	VisitProjectAwayOperatorFunc                                   func(*ProjectAwayOperator)
	VisitProjectByNamesOperatorFunc                                func(*ProjectByNamesOperator)
	VisitProjectKeepOperatorFunc                                   func(*ProjectKeepOperator)
	VisitProjectRenameOperatorFunc                                 func(*ProjectRenameOperator)
	VisitProjectReorderOperatorFunc                                func(*ProjectReorderOperator)
	VisitSampleOperatorFunc                                        func(*SampleOperator)
	VisitSampleDistinctOperatorFunc                                func(*SampleDistinctOperator)
	VisitEntityGroupFunc                                           func(*EntityGroup)
	VisitReduceByOperatorFunc                                      func(*ReduceByOperator)
	VisitReduceByWithClauseFunc                                    func(*ReduceByWithClause)
	VisitSummarizeOperatorFunc                                     func(*SummarizeOperator)
	VisitMacroExpandScopeReferenceNameFunc                         func(*MacroExpandScopeReferenceName)
	VisitMacroExpandOperatorFunc                                   func(*MacroExpandOperator)
	VisitSummarizeByClauseFunc                                     func(*SummarizeByClause)
	VisitDistinctOperatorFunc                                      func(*DistinctOperator)
	VisitTakeOperatorFunc                                          func(*TakeOperator)
	VisitSortOperatorFunc                                          func(*SortOperator)
	VisitOrderedExpressionFunc                                     func(*OrderedExpression)
	VisitOrderingClauseFunc                                        func(*OrderingClause)
	VisitOrderingNullsClauseFunc                                   func(*OrderingNullsClause)
	VisitTopHittersOperatorFunc                                    func(*TopHittersOperator)
	VisitTopHittersByClauseFunc                                    func(*TopHittersByClause)
	VisitTopOperatorFunc                                           func(*TopOperator)
	VisitTopNestedOperatorFunc                                     func(*TopNestedOperator)
	VisitTopNestedClauseFunc                                       func(*TopNestedClause)
	VisitTopNestedWithOthersClauseFunc                             func(*TopNestedWithOthersClause)
	VisitUnionOperatorFunc                                         func(*UnionOperator)
	VisitAsOperatorFunc                                            func(*AsOperator)
	VisitSerializeOperatorFunc                                     func(*SerializeOperator)
	VisitInvokeOperatorFunc                                        func(*InvokeOperator)
	VisitRenderOperatorFunc                                        func(*RenderOperator)
	VisitMakeGraphOperatorFunc                                     func(*MakeGraphOperator)
	VisitMakeGraphWithTablesAndKeysClauseFunc                      func(*MakeGraphWithTablesAndKeysClause)
	VisitMakeGraphWithImplicitIdClauseFunc                         func(*MakeGraphWithImplicitIdClause)
	VisitGraphMarkComponentsOperatorFunc                           func(*GraphMarkComponentsOperator)
	VisitGraphWhereNodesOperatorFunc                               func(*GraphWhereNodesOperator)
	VisitGraphWhereEdgesOperatorFunc                               func(*GraphWhereEdgesOperator)
	VisitMakeGraphTableAndKeyClauseFunc                            func(*MakeGraphTableAndKeyClause)
	VisitMakeGraphPartitionedByClauseFunc                          func(*MakeGraphPartitionedByClause)
	VisitGraphToTableOperatorFunc                                  func(*GraphToTableOperator)
	VisitGraphToTableOutputClauseFunc                              func(*GraphToTableOutputClause)
	VisitGraphToTableAsClauseFunc                                  func(*GraphToTableAsClause)
	VisitGraphMatchOperatorFunc                                    func(*GraphMatchOperator)
	VisitGraphShortestPathsOperatorFunc                            func(*GraphShortestPathsOperator)
	VisitGraphMatchPatternFunc                                     func(*GraphMatchPattern)
	VisitGraphMatchPatternNodeFunc                                 func(*GraphMatchPatternNode)
	VisitGraphMatchPatternEdgeFunc                                 func(*GraphMatchPatternEdge)
	VisitGraphMatchPatternEdgeRangeFunc                            func(*GraphMatchPatternEdgeRange)
	VisitWhereClauseFunc                                           func(*WhereClause)
	VisitProjectClauseFunc                                         func(*ProjectClause)
	VisitNameReferenceListFunc                                     func(*NameReferenceList)
	VisitRenderWithClauseFunc                                      func(*RenderWithClause)
	VisitPrintOperatorFunc                                         func(*PrintOperator)
	VisitAliasStatementFunc                                        func(*AliasStatement)
	VisitLetStatementFunc                                          func(*LetStatement)
	VisitFunctionDeclarationFunc                                   func(*FunctionDeclaration)
	VisitFunctionParametersFunc                                    func(*FunctionParameters)
	VisitFunctionParameterFunc                                     func(*FunctionParameter)
	VisitDefaultValueDeclarationFunc                               func(*DefaultValueDeclaration)
	VisitFunctionBodyFunc                                          func(*FunctionBody)
	VisitSchemaTypeExpressionFunc                                  func(*SchemaTypeExpression)
	VisitExpressionStatementFunc                                   func(*ExpressionStatement)
	VisitMaterializeExpressionFunc                                 func(*MaterializeExpression)
	VisitSetOptionStatementFunc                                    func(*SetOptionStatement)
	VisitOptionValueClauseFunc                                     func(*OptionValueClause)
	VisitQueryParametersStatementFunc                              func(*QueryParametersStatement)
	VisitRestrictStatementWithClauseFunc                           func(*RestrictStatementWithClause)
	VisitRestrictStatementFunc                                     func(*RestrictStatement)
	VisitPatternStatementFunc                                      func(*PatternStatement)
	VisitPatternDeclarationFunc                                    func(*PatternDeclaration)
	VisitPatternPathParameterFunc                                  func(*PatternPathParameter)
	VisitPatternMatchFunc                                          func(*PatternMatch)
	VisitPatternPathValueFunc                                      func(*PatternPathValue)
	VisitDataScopeExpressionFunc                                   func(*DataScopeExpression)
	VisitDataTableExpressionFunc                                   func(*DataTableExpression)
	VisitRowSchemaFunc                                             func(*RowSchema)
	VisitEvaluateRowSchemaFunc                                     func(*EvaluateRowSchema)
	VisitExternalDataExpressionFunc                                func(*ExternalDataExpression)
	VisitContextualDataTableExpressionFunc                         func(*ContextualDataTableExpression)
	VisitExternalDataWithClauseFunc                                func(*ExternalDataWithClause)
	VisitInlineExternalTableKindClauseFunc                         func(*InlineExternalTableKindClause)
	VisitInlineExternalTableDataFormatClauseFunc                   func(*InlineExternalTableDataFormatClause)
	VisitDateTimePatternFunc                                       func(*DateTimePattern)
	VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc func(*InlineExternalTablePathFormatPartitionColumnReference)
	VisitInlineExternalTablePathFormatClauseFunc                   func(*InlineExternalTablePathFormatClause)
	VisitPartitionColumnDeclarationFunc                            func(*PartitionColumnDeclaration)
	VisitInlineExternalTablePartitionClauseFunc                    func(*InlineExternalTablePartitionClause)
	VisitInlineExternalTableConnectionStringsClauseFunc            func(*InlineExternalTableConnectionStringsClause)
	VisitInlineExternalTableExpressionFunc                         func(*InlineExternalTableExpression)
	VisitJoinOperatorFunc                                          func(*JoinOperator)
	VisitLookupOperatorFunc                                        func(*LookupOperator)
	VisitJoinOnClauseFunc                                          func(*JoinOnClause)
	VisitJoinWhereClauseFunc                                       func(*JoinWhereClause)
	VisitScanOperatorFunc                                          func(*ScanOperator)
	VisitAssertSchemaOperatorFunc                                  func(*AssertSchemaOperator)
	VisitScanDeclareClauseFunc                                     func(*ScanDeclareClause)
	VisitScanOrderByClauseFunc                                     func(*ScanOrderByClause)
	VisitScanPartitionByClauseFunc                                 func(*ScanPartitionByClause)
	VisitScanStepOutputFunc                                        func(*ScanStepOutput)
	VisitScanStepFunc                                              func(*ScanStep)
	VisitScanComputationClauseFunc                                 func(*ScanComputationClause)
	VisitScanAssignmentFunc                                        func(*ScanAssignment)
	VisitPartitionByOperatorFunc                                   func(*PartitionByOperator)
	VisitPartitionByIdClauseFunc                                   func(*PartitionByIdClause)
	VisitCommandWithValueClauseFunc                                func(*CommandWithValueClause)
	VisitCommandWithPropertyListClauseFunc                         func(*CommandWithPropertyListClause)
	VisitUnknownCommandFunc                                        func(*UnknownCommand)
	VisitCustomCommandFunc                                         func(*CustomCommand)
	VisitPartialCommandFunc                                        func(*PartialCommand)
	VisitCommandAndSkippedTokensFunc                               func(*CommandAndSkippedTokens)
	VisitBadCommandFunc                                            func(*BadCommand)
	VisitCommandBlockFunc                                          func(*CommandBlock)
	VisitGraphMergeOperatorFunc                                    func(*GraphMergeOperator)
}

func (v SyntaxVisitorFuncs) defaultVisit(element SyntaxElement) {
	if v.DefaultVisit != nil {
		v.DefaultVisit(element)
	}
}

func (v SyntaxVisitorFuncs) VisitCustom(node *CustomNode) {
	if v.VisitCustomFunc != nil {
		v.VisitCustomFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitList(node *SyntaxList) {
	if v.VisitListFunc != nil {
		v.VisitListFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSeparatedElement(node *SeparatedElement) {
	if v.VisitSeparatedElementFunc != nil {
		v.VisitSeparatedElementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitQueryBlock(node *QueryBlock) {
	if v.VisitQueryBlockFunc != nil {
		v.VisitQueryBlockFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDirectiveBlock(node *DirectiveBlock) {
	if v.VisitDirectiveBlockFunc != nil {
		v.VisitDirectiveBlockFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDirective(node *Directive) {
	if v.VisitDirectiveFunc != nil {
		v.VisitDirectiveFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSkippedTokens(node *SkippedTokens) {
	if v.VisitSkippedTokensFunc != nil {
		v.VisitSkippedTokensFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTypeOfLiteralExpression(node *TypeOfLiteralExpression) {
	if v.VisitTypeOfLiteralExpressionFunc != nil {
		v.VisitTypeOfLiteralExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBadQueryOperator(node *BadQueryOperator) {
	if v.VisitBadQueryOperatorFunc != nil {
		v.VisitBadQueryOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCompoundStringLiteralExpression(node *CompoundStringLiteralExpression) {
	if v.VisitCompoundStringLiteralExpressionFunc != nil {
		v.VisitCompoundStringLiteralExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTokenName(node *TokenName) {
	if v.VisitTokenNameFunc != nil {
		v.VisitTokenNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBracketedName(node *BracketedName) {
	if v.VisitBracketedNameFunc != nil {
		v.VisitBracketedNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBracedName(node *BracedName) {
	if v.VisitBracedNameFunc != nil {
		v.VisitBracedNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitWildcardedName(node *WildcardedName) {
	if v.VisitWildcardedNameFunc != nil {
		v.VisitWildcardedNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBracketedWildcardedName(node *BracketedWildcardedName) {
	if v.VisitBracketedWildcardedNameFunc != nil {
		v.VisitBracketedWildcardedNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNameDeclaration(node *NameDeclaration) {
	if v.VisitNameDeclarationFunc != nil {
		v.VisitNameDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNameReference(node *NameReference) {
	if v.VisitNameReferenceFunc != nil {
		v.VisitNameReferenceFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitLiteralExpression(node *LiteralExpression) {
	if v.VisitLiteralExpressionFunc != nil {
		v.VisitLiteralExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitStarExpression(node *StarExpression) {
	if v.VisitStarExpressionFunc != nil {
		v.VisitStarExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitAtExpression(node *AtExpression) {
	if v.VisitAtExpressionFunc != nil {
		v.VisitAtExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJsonPair(node *JsonPair) {
	if v.VisitJsonPairFunc != nil {
		v.VisitJsonPairFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJsonObjectExpression(node *JsonObjectExpression) {
	if v.VisitJsonObjectExpressionFunc != nil {
		v.VisitJsonObjectExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJsonArrayExpression(node *JsonArrayExpression) {
	if v.VisitJsonArrayExpressionFunc != nil {
		v.VisitJsonArrayExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDynamicExpression(node *DynamicExpression) {
	if v.VisitDynamicExpressionFunc != nil {
		v.VisitDynamicExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitParenthesizedExpression(node *ParenthesizedExpression) {
	if v.VisitParenthesizedExpressionFunc != nil {
		v.VisitParenthesizedExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExpressionList(node *ExpressionList) {
	if v.VisitExpressionListFunc != nil {
		v.VisitExpressionListFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExpressionCouple(node *ExpressionCouple) {
	if v.VisitExpressionCoupleFunc != nil {
		v.VisitExpressionCoupleFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPrefixUnaryExpression(node *PrefixUnaryExpression) {
	if v.VisitPrefixUnaryExpressionFunc != nil {
		v.VisitPrefixUnaryExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBinaryExpression(node *BinaryExpression) {
	if v.VisitBinaryExpressionFunc != nil {
		v.VisitBinaryExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInExpression(node *InExpression) {
	if v.VisitInExpressionFunc != nil {
		v.VisitInExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitHasAnyExpression(node *HasAnyExpression) {
	if v.VisitHasAnyExpressionFunc != nil {
		v.VisitHasAnyExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitHasAllExpression(node *HasAllExpression) {
	if v.VisitHasAllExpressionFunc != nil {
		v.VisitHasAllExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBetweenExpression(node *BetweenExpression) {
	if v.VisitBetweenExpressionFunc != nil {
		v.VisitBetweenExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFunctionCallExpression(node *FunctionCallExpression) {
	if v.VisitFunctionCallExpressionFunc != nil {
		v.VisitFunctionCallExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitToScalarExpression(node *ToScalarExpression) {
	if v.VisitToScalarExpressionFunc != nil {
		v.VisitToScalarExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitToTableExpression(node *ToTableExpression) {
	if v.VisitToTableExpressionFunc != nil {
		v.VisitToTableExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMaterializedViewCombineExpression(node *MaterializedViewCombineExpression) {
	if v.VisitMaterializedViewCombineExpressionFunc != nil {
		v.VisitMaterializedViewCombineExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMaterializedViewCombineNameClause(node *MaterializedViewCombineNameClause) {
	if v.VisitMaterializedViewCombineNameClauseFunc != nil {
		v.VisitMaterializedViewCombineNameClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMaterializedViewCombineClause(node *MaterializedViewCombineClause) {
	if v.VisitMaterializedViewCombineClauseFunc != nil {
		v.VisitMaterializedViewCombineClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSimpleNamedExpression(node *SimpleNamedExpression) {
	if v.VisitSimpleNamedExpressionFunc != nil {
		v.VisitSimpleNamedExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRenameList(node *RenameList) {
	if v.VisitRenameListFunc != nil {
		v.VisitRenameListFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCompoundNamedExpression(node *CompoundNamedExpression) {
	if v.VisitCompoundNamedExpressionFunc != nil {
		v.VisitCompoundNamedExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBracketedExpression(node *BracketedExpression) {
	if v.VisitBracketedExpressionFunc != nil {
		v.VisitBracketedExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPathExpression(node *PathExpression) {
	if v.VisitPathExpressionFunc != nil {
		v.VisitPathExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitElementExpression(node *ElementExpression) {
	if v.VisitElementExpressionFunc != nil {
		v.VisitElementExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPipeExpression(node *PipeExpression) {
	if v.VisitPipeExpressionFunc != nil {
		v.VisitPipeExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRangeOperator(node *RangeOperator) {
	if v.VisitRangeOperatorFunc != nil {
		v.VisitRangeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNamedParameter(node *NamedParameter) {
	if v.VisitNamedParameterFunc != nil {
		v.VisitNamedParameterFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitConsumeOperator(node *ConsumeOperator) {
	if v.VisitConsumeOperatorFunc != nil {
		v.VisitConsumeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCountOperator(node *CountOperator) {
	if v.VisitCountOperatorFunc != nil {
		v.VisitCountOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCountAsIdentifierClause(node *CountAsIdentifierClause) {
	if v.VisitCountAsIdentifierClauseFunc != nil {
		v.VisitCountAsIdentifierClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExecuteAndCacheOperator(node *ExecuteAndCacheOperator) {
	if v.VisitExecuteAndCacheOperatorFunc != nil {
		v.VisitExecuteAndCacheOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExtendOperator(node *ExtendOperator) {
	if v.VisitExtendOperatorFunc != nil {
		v.VisitExtendOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFacetOperator(node *FacetOperator) {
	if v.VisitFacetOperatorFunc != nil {
		v.VisitFacetOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFacetWithOperatorClause(node *FacetWithOperatorClause) {
	if v.VisitFacetWithOperatorClauseFunc != nil {
		v.VisitFacetWithOperatorClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFacetWithExpressionClause(node *FacetWithExpressionClause) {
	if v.VisitFacetWithExpressionClauseFunc != nil {
		v.VisitFacetWithExpressionClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFilterOperator(node *FilterOperator) {
	if v.VisitFilterOperatorFunc != nil {
		v.VisitFilterOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGetSchemaOperator(node *GetSchemaOperator) {
	if v.VisitGetSchemaOperatorFunc != nil {
		v.VisitGetSchemaOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFindOperator(node *FindOperator) {
	if v.VisitFindOperatorFunc != nil {
		v.VisitFindOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDataScopeClause(node *DataScopeClause) {
	if v.VisitDataScopeClauseFunc != nil {
		v.VisitDataScopeClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTypedColumnReference(node *TypedColumnReference) {
	if v.VisitTypedColumnReferenceFunc != nil {
		v.VisitTypedColumnReferenceFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFindInClause(node *FindInClause) {
	if v.VisitFindInClauseFunc != nil {
		v.VisitFindInClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFindProjectClause(node *FindProjectClause) {
	if v.VisitFindProjectClauseFunc != nil {
		v.VisitFindProjectClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPackExpression(node *PackExpression) {
	if v.VisitPackExpressionFunc != nil {
		v.VisitPackExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNameAndTypeDeclaration(node *NameAndTypeDeclaration) {
	if v.VisitNameAndTypeDeclarationFunc != nil {
		v.VisitNameAndTypeDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPrimitiveTypeExpression(node *PrimitiveTypeExpression) {
	if v.VisitPrimitiveTypeExpressionFunc != nil {
		v.VisitPrimitiveTypeExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSearchOperator(node *SearchOperator) {
	if v.VisitSearchOperatorFunc != nil {
		v.VisitSearchOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitForkOperator(node *ForkOperator) {
	if v.VisitForkOperatorFunc != nil {
		v.VisitForkOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNameEqualsClause(node *NameEqualsClause) {
	if v.VisitNameEqualsClauseFunc != nil {
		v.VisitNameEqualsClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitForkExpression(node *ForkExpression) {
	if v.VisitForkExpressionFunc != nil {
		v.VisitForkExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesOperator(node *MakeSeriesOperator) {
	if v.VisitMakeSeriesOperatorFunc != nil {
		v.VisitMakeSeriesOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesExpression(node *MakeSeriesExpression) {
	if v.VisitMakeSeriesExpressionFunc != nil {
		v.VisitMakeSeriesExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDefaultExpressionClause(node *DefaultExpressionClause) {
	if v.VisitDefaultExpressionClauseFunc != nil {
		v.VisitDefaultExpressionClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesOnClause(node *MakeSeriesOnClause) {
	if v.VisitMakeSeriesOnClauseFunc != nil {
		v.VisitMakeSeriesOnClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesFromClause(node *MakeSeriesFromClause) {
	if v.VisitMakeSeriesFromClauseFunc != nil {
		v.VisitMakeSeriesFromClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesToClause(node *MakeSeriesToClause) {
	if v.VisitMakeSeriesToClauseFunc != nil {
		v.VisitMakeSeriesToClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesStepClause(node *MakeSeriesStepClause) {
	if v.VisitMakeSeriesStepClauseFunc != nil {
		v.VisitMakeSeriesStepClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesInRangeClause(node *MakeSeriesInRangeClause) {
	if v.VisitMakeSeriesInRangeClauseFunc != nil {
		v.VisitMakeSeriesInRangeClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesFromToStepClause(node *MakeSeriesFromToStepClause) {
	if v.VisitMakeSeriesFromToStepClauseFunc != nil {
		v.VisitMakeSeriesFromToStepClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeSeriesByClause(node *MakeSeriesByClause) {
	if v.VisitMakeSeriesByClauseFunc != nil {
		v.VisitMakeSeriesByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvExpandOperator(node *MvExpandOperator) {
	if v.VisitMvExpandOperatorFunc != nil {
		v.VisitMvExpandOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvExpandExpression(node *MvExpandExpression) {
	if v.VisitMvExpandExpressionFunc != nil {
		v.VisitMvExpandExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvExpandRowLimitClause(node *MvExpandRowLimitClause) {
	if v.VisitMvExpandRowLimitClauseFunc != nil {
		v.VisitMvExpandRowLimitClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvApplyOperator(node *MvApplyOperator) {
	if v.VisitMvApplyOperatorFunc != nil {
		v.VisitMvApplyOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvApplyExpression(node *MvApplyExpression) {
	if v.VisitMvApplyExpressionFunc != nil {
		v.VisitMvApplyExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvApplyRowLimitClause(node *MvApplyRowLimitClause) {
	if v.VisitMvApplyRowLimitClauseFunc != nil {
		v.VisitMvApplyRowLimitClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvApplyContextIdClause(node *MvApplyContextIdClause) {
	if v.VisitMvApplyContextIdClauseFunc != nil {
		v.VisitMvApplyContextIdClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMvApplySubqueryExpression(node *MvApplySubqueryExpression) {
	if v.VisitMvApplySubqueryExpressionFunc != nil {
		v.VisitMvApplySubqueryExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitToTypeOfClause(node *ToTypeOfClause) {
	if v.VisitToTypeOfClauseFunc != nil {
		v.VisitToTypeOfClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitEvaluateSchemaClause(node *EvaluateSchemaClause) {
	if v.VisitEvaluateSchemaClauseFunc != nil {
		v.VisitEvaluateSchemaClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitEvaluateOperator(node *EvaluateOperator) {
	if v.VisitEvaluateOperatorFunc != nil {
		v.VisitEvaluateOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitParseOperator(node *ParseOperator) {
	if v.VisitParseOperatorFunc != nil {
		v.VisitParseOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitParseWhereOperator(node *ParseWhereOperator) {
	if v.VisitParseWhereOperatorFunc != nil {
		v.VisitParseWhereOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitParseKvWithClause(node *ParseKvWithClause) {
	if v.VisitParseKvWithClauseFunc != nil {
		v.VisitParseKvWithClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitParseKvOperator(node *ParseKvOperator) {
	if v.VisitParseKvOperatorFunc != nil {
		v.VisitParseKvOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionOperator(node *PartitionOperator) {
	if v.VisitPartitionOperatorFunc != nil {
		v.VisitPartitionOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionQuery(node *PartitionQuery) {
	if v.VisitPartitionQueryFunc != nil {
		v.VisitPartitionQueryFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionScope(node *PartitionScope) {
	if v.VisitPartitionScopeFunc != nil {
		v.VisitPartitionScopeFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionSubquery(node *PartitionSubquery) {
	if v.VisitPartitionSubqueryFunc != nil {
		v.VisitPartitionSubqueryFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectOperator(node *ProjectOperator) {
	if v.VisitProjectOperatorFunc != nil {
		v.VisitProjectOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectAwayOperator(node *ProjectAwayOperator) {
	if v.VisitProjectAwayOperatorFunc != nil {
		v.VisitProjectAwayOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectByNamesOperator(node *ProjectByNamesOperator) {
	if v.VisitProjectByNamesOperatorFunc != nil {
		v.VisitProjectByNamesOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectKeepOperator(node *ProjectKeepOperator) {
	if v.VisitProjectKeepOperatorFunc != nil {
		v.VisitProjectKeepOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectRenameOperator(node *ProjectRenameOperator) {
	if v.VisitProjectRenameOperatorFunc != nil {
		v.VisitProjectRenameOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectReorderOperator(node *ProjectReorderOperator) {
	if v.VisitProjectReorderOperatorFunc != nil {
		v.VisitProjectReorderOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSampleOperator(node *SampleOperator) {
	if v.VisitSampleOperatorFunc != nil {
		v.VisitSampleOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSampleDistinctOperator(node *SampleDistinctOperator) {
	if v.VisitSampleDistinctOperatorFunc != nil {
		v.VisitSampleDistinctOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitEntityGroup(node *EntityGroup) {
	if v.VisitEntityGroupFunc != nil {
		v.VisitEntityGroupFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitReduceByOperator(node *ReduceByOperator) {
	if v.VisitReduceByOperatorFunc != nil {
		v.VisitReduceByOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitReduceByWithClause(node *ReduceByWithClause) {
	if v.VisitReduceByWithClauseFunc != nil {
		v.VisitReduceByWithClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSummarizeOperator(node *SummarizeOperator) {
	if v.VisitSummarizeOperatorFunc != nil {
		v.VisitSummarizeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMacroExpandScopeReferenceName(node *MacroExpandScopeReferenceName) {
	if v.VisitMacroExpandScopeReferenceNameFunc != nil {
		v.VisitMacroExpandScopeReferenceNameFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMacroExpandOperator(node *MacroExpandOperator) {
	if v.VisitMacroExpandOperatorFunc != nil {
		v.VisitMacroExpandOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSummarizeByClause(node *SummarizeByClause) {
	if v.VisitSummarizeByClauseFunc != nil {
		v.VisitSummarizeByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDistinctOperator(node *DistinctOperator) {
	if v.VisitDistinctOperatorFunc != nil {
		v.VisitDistinctOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTakeOperator(node *TakeOperator) {
	if v.VisitTakeOperatorFunc != nil {
		v.VisitTakeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSortOperator(node *SortOperator) {
	if v.VisitSortOperatorFunc != nil {
		v.VisitSortOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitOrderedExpression(node *OrderedExpression) {
	if v.VisitOrderedExpressionFunc != nil {
		v.VisitOrderedExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitOrderingClause(node *OrderingClause) {
	if v.VisitOrderingClauseFunc != nil {
		v.VisitOrderingClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitOrderingNullsClause(node *OrderingNullsClause) {
	if v.VisitOrderingNullsClauseFunc != nil {
		v.VisitOrderingNullsClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopHittersOperator(node *TopHittersOperator) {
	if v.VisitTopHittersOperatorFunc != nil {
		v.VisitTopHittersOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopHittersByClause(node *TopHittersByClause) {
	if v.VisitTopHittersByClauseFunc != nil {
		v.VisitTopHittersByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopOperator(node *TopOperator) {
	if v.VisitTopOperatorFunc != nil {
		v.VisitTopOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopNestedOperator(node *TopNestedOperator) {
	if v.VisitTopNestedOperatorFunc != nil {
		v.VisitTopNestedOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopNestedClause(node *TopNestedClause) {
	if v.VisitTopNestedClauseFunc != nil {
		v.VisitTopNestedClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitTopNestedWithOthersClause(node *TopNestedWithOthersClause) {
	if v.VisitTopNestedWithOthersClauseFunc != nil {
		v.VisitTopNestedWithOthersClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitUnionOperator(node *UnionOperator) {
	if v.VisitUnionOperatorFunc != nil {
		v.VisitUnionOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitAsOperator(node *AsOperator) {
	if v.VisitAsOperatorFunc != nil {
		v.VisitAsOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSerializeOperator(node *SerializeOperator) {
	if v.VisitSerializeOperatorFunc != nil {
		v.VisitSerializeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInvokeOperator(node *InvokeOperator) {
	if v.VisitInvokeOperatorFunc != nil {
		v.VisitInvokeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRenderOperator(node *RenderOperator) {
	if v.VisitRenderOperatorFunc != nil {
		v.VisitRenderOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeGraphOperator(node *MakeGraphOperator) {
	if v.VisitMakeGraphOperatorFunc != nil {
		v.VisitMakeGraphOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeGraphWithTablesAndKeysClause(node *MakeGraphWithTablesAndKeysClause) {
	if v.VisitMakeGraphWithTablesAndKeysClauseFunc != nil {
		v.VisitMakeGraphWithTablesAndKeysClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeGraphWithImplicitIdClause(node *MakeGraphWithImplicitIdClause) {
	if v.VisitMakeGraphWithImplicitIdClauseFunc != nil {
		v.VisitMakeGraphWithImplicitIdClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMarkComponentsOperator(node *GraphMarkComponentsOperator) {
	if v.VisitGraphMarkComponentsOperatorFunc != nil {
		v.VisitGraphMarkComponentsOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphWhereNodesOperator(node *GraphWhereNodesOperator) {
	if v.VisitGraphWhereNodesOperatorFunc != nil {
		v.VisitGraphWhereNodesOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphWhereEdgesOperator(node *GraphWhereEdgesOperator) {
	if v.VisitGraphWhereEdgesOperatorFunc != nil {
		v.VisitGraphWhereEdgesOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeGraphTableAndKeyClause(node *MakeGraphTableAndKeyClause) {
	if v.VisitMakeGraphTableAndKeyClauseFunc != nil {
		v.VisitMakeGraphTableAndKeyClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMakeGraphPartitionedByClause(node *MakeGraphPartitionedByClause) {
	if v.VisitMakeGraphPartitionedByClauseFunc != nil {
		v.VisitMakeGraphPartitionedByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphToTableOperator(node *GraphToTableOperator) {
	if v.VisitGraphToTableOperatorFunc != nil {
		v.VisitGraphToTableOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphToTableOutputClause(node *GraphToTableOutputClause) {
	if v.VisitGraphToTableOutputClauseFunc != nil {
		v.VisitGraphToTableOutputClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphToTableAsClause(node *GraphToTableAsClause) {
	if v.VisitGraphToTableAsClauseFunc != nil {
		v.VisitGraphToTableAsClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMatchOperator(node *GraphMatchOperator) {
	if v.VisitGraphMatchOperatorFunc != nil {
		v.VisitGraphMatchOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphShortestPathsOperator(node *GraphShortestPathsOperator) {
	if v.VisitGraphShortestPathsOperatorFunc != nil {
		v.VisitGraphShortestPathsOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMatchPattern(node *GraphMatchPattern) {
	if v.VisitGraphMatchPatternFunc != nil {
		v.VisitGraphMatchPatternFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMatchPatternNode(node *GraphMatchPatternNode) {
	if v.VisitGraphMatchPatternNodeFunc != nil {
		v.VisitGraphMatchPatternNodeFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMatchPatternEdge(node *GraphMatchPatternEdge) {
	if v.VisitGraphMatchPatternEdgeFunc != nil {
		v.VisitGraphMatchPatternEdgeFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMatchPatternEdgeRange(node *GraphMatchPatternEdgeRange) {
	if v.VisitGraphMatchPatternEdgeRangeFunc != nil {
		v.VisitGraphMatchPatternEdgeRangeFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitWhereClause(node *WhereClause) {
	if v.VisitWhereClauseFunc != nil {
		v.VisitWhereClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitProjectClause(node *ProjectClause) {
	if v.VisitProjectClauseFunc != nil {
		v.VisitProjectClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitNameReferenceList(node *NameReferenceList) {
	if v.VisitNameReferenceListFunc != nil {
		v.VisitNameReferenceListFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRenderWithClause(node *RenderWithClause) {
	if v.VisitRenderWithClauseFunc != nil {
		v.VisitRenderWithClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPrintOperator(node *PrintOperator) {
	if v.VisitPrintOperatorFunc != nil {
		v.VisitPrintOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitAliasStatement(node *AliasStatement) {
	if v.VisitAliasStatementFunc != nil {
		v.VisitAliasStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitLetStatement(node *LetStatement) {
	if v.VisitLetStatementFunc != nil {
		v.VisitLetStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFunctionDeclaration(node *FunctionDeclaration) {
	if v.VisitFunctionDeclarationFunc != nil {
		v.VisitFunctionDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFunctionParameters(node *FunctionParameters) {
	if v.VisitFunctionParametersFunc != nil {
		v.VisitFunctionParametersFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFunctionParameter(node *FunctionParameter) {
	if v.VisitFunctionParameterFunc != nil {
		v.VisitFunctionParameterFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDefaultValueDeclaration(node *DefaultValueDeclaration) {
	if v.VisitDefaultValueDeclarationFunc != nil {
		v.VisitDefaultValueDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitFunctionBody(node *FunctionBody) {
	if v.VisitFunctionBodyFunc != nil {
		v.VisitFunctionBodyFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSchemaTypeExpression(node *SchemaTypeExpression) {
	if v.VisitSchemaTypeExpressionFunc != nil {
		v.VisitSchemaTypeExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExpressionStatement(node *ExpressionStatement) {
	if v.VisitExpressionStatementFunc != nil {
		v.VisitExpressionStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitMaterializeExpression(node *MaterializeExpression) {
	if v.VisitMaterializeExpressionFunc != nil {
		v.VisitMaterializeExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitSetOptionStatement(node *SetOptionStatement) {
	if v.VisitSetOptionStatementFunc != nil {
		v.VisitSetOptionStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitOptionValueClause(node *OptionValueClause) {
	if v.VisitOptionValueClauseFunc != nil {
		v.VisitOptionValueClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitQueryParametersStatement(node *QueryParametersStatement) {
	if v.VisitQueryParametersStatementFunc != nil {
		v.VisitQueryParametersStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRestrictStatementWithClause(node *RestrictStatementWithClause) {
	if v.VisitRestrictStatementWithClauseFunc != nil {
		v.VisitRestrictStatementWithClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRestrictStatement(node *RestrictStatement) {
	if v.VisitRestrictStatementFunc != nil {
		v.VisitRestrictStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPatternStatement(node *PatternStatement) {
	if v.VisitPatternStatementFunc != nil {
		v.VisitPatternStatementFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPatternDeclaration(node *PatternDeclaration) {
	if v.VisitPatternDeclarationFunc != nil {
		v.VisitPatternDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPatternPathParameter(node *PatternPathParameter) {
	if v.VisitPatternPathParameterFunc != nil {
		v.VisitPatternPathParameterFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPatternMatch(node *PatternMatch) {
	if v.VisitPatternMatchFunc != nil {
		v.VisitPatternMatchFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPatternPathValue(node *PatternPathValue) {
	if v.VisitPatternPathValueFunc != nil {
		v.VisitPatternPathValueFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDataScopeExpression(node *DataScopeExpression) {
	if v.VisitDataScopeExpressionFunc != nil {
		v.VisitDataScopeExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDataTableExpression(node *DataTableExpression) {
	if v.VisitDataTableExpressionFunc != nil {
		v.VisitDataTableExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitRowSchema(node *RowSchema) {
	if v.VisitRowSchemaFunc != nil {
		v.VisitRowSchemaFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitEvaluateRowSchema(node *EvaluateRowSchema) {
	if v.VisitEvaluateRowSchemaFunc != nil {
		v.VisitEvaluateRowSchemaFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExternalDataExpression(node *ExternalDataExpression) {
	if v.VisitExternalDataExpressionFunc != nil {
		v.VisitExternalDataExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitContextualDataTableExpression(node *ContextualDataTableExpression) {
	if v.VisitContextualDataTableExpressionFunc != nil {
		v.VisitContextualDataTableExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitExternalDataWithClause(node *ExternalDataWithClause) {
	if v.VisitExternalDataWithClauseFunc != nil {
		v.VisitExternalDataWithClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTableKindClause(node *InlineExternalTableKindClause) {
	if v.VisitInlineExternalTableKindClauseFunc != nil {
		v.VisitInlineExternalTableKindClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTableDataFormatClause(node *InlineExternalTableDataFormatClause) {
	if v.VisitInlineExternalTableDataFormatClauseFunc != nil {
		v.VisitInlineExternalTableDataFormatClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitDateTimePattern(node *DateTimePattern) {
	if v.VisitDateTimePatternFunc != nil {
		v.VisitDateTimePatternFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTablePathFormatPartitionColumnReference(node *InlineExternalTablePathFormatPartitionColumnReference) {
	if v.VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc != nil {
		v.VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTablePathFormatClause(node *InlineExternalTablePathFormatClause) {
	if v.VisitInlineExternalTablePathFormatClauseFunc != nil {
		v.VisitInlineExternalTablePathFormatClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionColumnDeclaration(node *PartitionColumnDeclaration) {
	if v.VisitPartitionColumnDeclarationFunc != nil {
		v.VisitPartitionColumnDeclarationFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTablePartitionClause(node *InlineExternalTablePartitionClause) {
	if v.VisitInlineExternalTablePartitionClauseFunc != nil {
		v.VisitInlineExternalTablePartitionClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTableConnectionStringsClause(node *InlineExternalTableConnectionStringsClause) {
	if v.VisitInlineExternalTableConnectionStringsClauseFunc != nil {
		v.VisitInlineExternalTableConnectionStringsClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitInlineExternalTableExpression(node *InlineExternalTableExpression) {
	if v.VisitInlineExternalTableExpressionFunc != nil {
		v.VisitInlineExternalTableExpressionFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJoinOperator(node *JoinOperator) {
	if v.VisitJoinOperatorFunc != nil {
		v.VisitJoinOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitLookupOperator(node *LookupOperator) {
	if v.VisitLookupOperatorFunc != nil {
		v.VisitLookupOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJoinOnClause(node *JoinOnClause) {
	if v.VisitJoinOnClauseFunc != nil {
		v.VisitJoinOnClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitJoinWhereClause(node *JoinWhereClause) {
	if v.VisitJoinWhereClauseFunc != nil {
		v.VisitJoinWhereClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanOperator(node *ScanOperator) {
	if v.VisitScanOperatorFunc != nil {
		v.VisitScanOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitAssertSchemaOperator(node *AssertSchemaOperator) {
	if v.VisitAssertSchemaOperatorFunc != nil {
		v.VisitAssertSchemaOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanDeclareClause(node *ScanDeclareClause) {
	if v.VisitScanDeclareClauseFunc != nil {
		v.VisitScanDeclareClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanOrderByClause(node *ScanOrderByClause) {
	if v.VisitScanOrderByClauseFunc != nil {
		v.VisitScanOrderByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanPartitionByClause(node *ScanPartitionByClause) {
	if v.VisitScanPartitionByClauseFunc != nil {
		v.VisitScanPartitionByClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanStepOutput(node *ScanStepOutput) {
	if v.VisitScanStepOutputFunc != nil {
		v.VisitScanStepOutputFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanStep(node *ScanStep) {
	if v.VisitScanStepFunc != nil {
		v.VisitScanStepFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanComputationClause(node *ScanComputationClause) {
	if v.VisitScanComputationClauseFunc != nil {
		v.VisitScanComputationClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitScanAssignment(node *ScanAssignment) {
	if v.VisitScanAssignmentFunc != nil {
		v.VisitScanAssignmentFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionByOperator(node *PartitionByOperator) {
	if v.VisitPartitionByOperatorFunc != nil {
		v.VisitPartitionByOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartitionByIdClause(node *PartitionByIdClause) {
	if v.VisitPartitionByIdClauseFunc != nil {
		v.VisitPartitionByIdClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCommandWithValueClause(node *CommandWithValueClause) {
	if v.VisitCommandWithValueClauseFunc != nil {
		v.VisitCommandWithValueClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCommandWithPropertyListClause(node *CommandWithPropertyListClause) {
	if v.VisitCommandWithPropertyListClauseFunc != nil {
		v.VisitCommandWithPropertyListClauseFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitUnknownCommand(node *UnknownCommand) {
	if v.VisitUnknownCommandFunc != nil {
		v.VisitUnknownCommandFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCustomCommand(node *CustomCommand) {
	if v.VisitCustomCommandFunc != nil {
		v.VisitCustomCommandFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitPartialCommand(node *PartialCommand) {
	if v.VisitPartialCommandFunc != nil {
		v.VisitPartialCommandFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCommandAndSkippedTokens(node *CommandAndSkippedTokens) {
	if v.VisitCommandAndSkippedTokensFunc != nil {
		v.VisitCommandAndSkippedTokensFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitBadCommand(node *BadCommand) {
	if v.VisitBadCommandFunc != nil {
		v.VisitBadCommandFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitCommandBlock(node *CommandBlock) {
	if v.VisitCommandBlockFunc != nil {
		v.VisitCommandBlockFunc(node)
		return
	}
	v.defaultVisit(node)
}

func (v SyntaxVisitorFuncs) VisitGraphMergeOperator(node *GraphMergeOperator) {
	if v.VisitGraphMergeOperatorFunc != nil {
		v.VisitGraphMergeOperatorFunc(node)
		return
	}
	v.defaultVisit(node)
}

type SyntaxVisitorResultFuncs[T any] struct {
	DefaultVisit                                                   func(SyntaxElement) T
	VisitCustomFunc                                                func(*CustomNode) T
	VisitListFunc                                                  func(*SyntaxList) T
	VisitSeparatedElementFunc                                      func(*SeparatedElement) T
	VisitQueryBlockFunc                                            func(*QueryBlock) T
	VisitDirectiveBlockFunc                                        func(*DirectiveBlock) T
	VisitDirectiveFunc                                             func(*Directive) T
	VisitSkippedTokensFunc                                         func(*SkippedTokens) T
	VisitTypeOfLiteralExpressionFunc                               func(*TypeOfLiteralExpression) T
	VisitBadQueryOperatorFunc                                      func(*BadQueryOperator) T
	VisitCompoundStringLiteralExpressionFunc                       func(*CompoundStringLiteralExpression) T
	VisitTokenNameFunc                                             func(*TokenName) T
	VisitBracketedNameFunc                                         func(*BracketedName) T
	VisitBracedNameFunc                                            func(*BracedName) T
	VisitWildcardedNameFunc                                        func(*WildcardedName) T
	VisitBracketedWildcardedNameFunc                               func(*BracketedWildcardedName) T
	VisitNameDeclarationFunc                                       func(*NameDeclaration) T
	VisitNameReferenceFunc                                         func(*NameReference) T
	VisitLiteralExpressionFunc                                     func(*LiteralExpression) T
	VisitStarExpressionFunc                                        func(*StarExpression) T
	VisitAtExpressionFunc                                          func(*AtExpression) T
	VisitJsonPairFunc                                              func(*JsonPair) T
	VisitJsonObjectExpressionFunc                                  func(*JsonObjectExpression) T
	VisitJsonArrayExpressionFunc                                   func(*JsonArrayExpression) T
	VisitDynamicExpressionFunc                                     func(*DynamicExpression) T
	VisitParenthesizedExpressionFunc                               func(*ParenthesizedExpression) T
	VisitExpressionListFunc                                        func(*ExpressionList) T
	VisitExpressionCoupleFunc                                      func(*ExpressionCouple) T
	VisitPrefixUnaryExpressionFunc                                 func(*PrefixUnaryExpression) T
	VisitBinaryExpressionFunc                                      func(*BinaryExpression) T
	VisitInExpressionFunc                                          func(*InExpression) T
	VisitHasAnyExpressionFunc                                      func(*HasAnyExpression) T
	VisitHasAllExpressionFunc                                      func(*HasAllExpression) T
	VisitBetweenExpressionFunc                                     func(*BetweenExpression) T
	VisitFunctionCallExpressionFunc                                func(*FunctionCallExpression) T
	VisitToScalarExpressionFunc                                    func(*ToScalarExpression) T
	VisitToTableExpressionFunc                                     func(*ToTableExpression) T
	VisitMaterializedViewCombineExpressionFunc                     func(*MaterializedViewCombineExpression) T
	VisitMaterializedViewCombineNameClauseFunc                     func(*MaterializedViewCombineNameClause) T
	VisitMaterializedViewCombineClauseFunc                         func(*MaterializedViewCombineClause) T
	VisitSimpleNamedExpressionFunc                                 func(*SimpleNamedExpression) T
	VisitRenameListFunc                                            func(*RenameList) T
	VisitCompoundNamedExpressionFunc                               func(*CompoundNamedExpression) T
	VisitBracketedExpressionFunc                                   func(*BracketedExpression) T
	VisitPathExpressionFunc                                        func(*PathExpression) T
	VisitElementExpressionFunc                                     func(*ElementExpression) T
	VisitPipeExpressionFunc                                        func(*PipeExpression) T
	VisitRangeOperatorFunc                                         func(*RangeOperator) T
	VisitNamedParameterFunc                                        func(*NamedParameter) T
	VisitConsumeOperatorFunc                                       func(*ConsumeOperator) T
	VisitCountOperatorFunc                                         func(*CountOperator) T
	VisitCountAsIdentifierClauseFunc                               func(*CountAsIdentifierClause) T
	VisitExecuteAndCacheOperatorFunc                               func(*ExecuteAndCacheOperator) T
	VisitExtendOperatorFunc                                        func(*ExtendOperator) T
	VisitFacetOperatorFunc                                         func(*FacetOperator) T
	VisitFacetWithOperatorClauseFunc                               func(*FacetWithOperatorClause) T
	VisitFacetWithExpressionClauseFunc                             func(*FacetWithExpressionClause) T
	VisitFilterOperatorFunc                                        func(*FilterOperator) T
	VisitGetSchemaOperatorFunc                                     func(*GetSchemaOperator) T
	VisitFindOperatorFunc                                          func(*FindOperator) T
	VisitDataScopeClauseFunc                                       func(*DataScopeClause) T
	VisitTypedColumnReferenceFunc                                  func(*TypedColumnReference) T
	VisitFindInClauseFunc                                          func(*FindInClause) T
	VisitFindProjectClauseFunc                                     func(*FindProjectClause) T
	VisitPackExpressionFunc                                        func(*PackExpression) T
	VisitNameAndTypeDeclarationFunc                                func(*NameAndTypeDeclaration) T
	VisitPrimitiveTypeExpressionFunc                               func(*PrimitiveTypeExpression) T
	VisitSearchOperatorFunc                                        func(*SearchOperator) T
	VisitForkOperatorFunc                                          func(*ForkOperator) T
	VisitNameEqualsClauseFunc                                      func(*NameEqualsClause) T
	VisitForkExpressionFunc                                        func(*ForkExpression) T
	VisitMakeSeriesOperatorFunc                                    func(*MakeSeriesOperator) T
	VisitMakeSeriesExpressionFunc                                  func(*MakeSeriesExpression) T
	VisitDefaultExpressionClauseFunc                               func(*DefaultExpressionClause) T
	VisitMakeSeriesOnClauseFunc                                    func(*MakeSeriesOnClause) T
	VisitMakeSeriesFromClauseFunc                                  func(*MakeSeriesFromClause) T
	VisitMakeSeriesToClauseFunc                                    func(*MakeSeriesToClause) T
	VisitMakeSeriesStepClauseFunc                                  func(*MakeSeriesStepClause) T
	VisitMakeSeriesInRangeClauseFunc                               func(*MakeSeriesInRangeClause) T
	VisitMakeSeriesFromToStepClauseFunc                            func(*MakeSeriesFromToStepClause) T
	VisitMakeSeriesByClauseFunc                                    func(*MakeSeriesByClause) T
	VisitMvExpandOperatorFunc                                      func(*MvExpandOperator) T
	VisitMvExpandExpressionFunc                                    func(*MvExpandExpression) T
	VisitMvExpandRowLimitClauseFunc                                func(*MvExpandRowLimitClause) T
	VisitMvApplyOperatorFunc                                       func(*MvApplyOperator) T
	VisitMvApplyExpressionFunc                                     func(*MvApplyExpression) T
	VisitMvApplyRowLimitClauseFunc                                 func(*MvApplyRowLimitClause) T
	VisitMvApplyContextIdClauseFunc                                func(*MvApplyContextIdClause) T
	VisitMvApplySubqueryExpressionFunc                             func(*MvApplySubqueryExpression) T
	VisitToTypeOfClauseFunc                                        func(*ToTypeOfClause) T
	VisitEvaluateSchemaClauseFunc                                  func(*EvaluateSchemaClause) T
	VisitEvaluateOperatorFunc                                      func(*EvaluateOperator) T
	VisitParseOperatorFunc                                         func(*ParseOperator) T
	VisitParseWhereOperatorFunc                                    func(*ParseWhereOperator) T
	VisitParseKvWithClauseFunc                                     func(*ParseKvWithClause) T
	VisitParseKvOperatorFunc                                       func(*ParseKvOperator) T
	VisitPartitionOperatorFunc                                     func(*PartitionOperator) T
	VisitPartitionQueryFunc                                        func(*PartitionQuery) T
	VisitPartitionScopeFunc                                        func(*PartitionScope) T
	VisitPartitionSubqueryFunc                                     func(*PartitionSubquery) T
	VisitProjectOperatorFunc                                       func(*ProjectOperator) T
	VisitProjectAwayOperatorFunc                                   func(*ProjectAwayOperator) T
	VisitProjectByNamesOperatorFunc                                func(*ProjectByNamesOperator) T
	VisitProjectKeepOperatorFunc                                   func(*ProjectKeepOperator) T
	VisitProjectRenameOperatorFunc                                 func(*ProjectRenameOperator) T
	VisitProjectReorderOperatorFunc                                func(*ProjectReorderOperator) T
	VisitSampleOperatorFunc                                        func(*SampleOperator) T
	VisitSampleDistinctOperatorFunc                                func(*SampleDistinctOperator) T
	VisitEntityGroupFunc                                           func(*EntityGroup) T
	VisitReduceByOperatorFunc                                      func(*ReduceByOperator) T
	VisitReduceByWithClauseFunc                                    func(*ReduceByWithClause) T
	VisitSummarizeOperatorFunc                                     func(*SummarizeOperator) T
	VisitMacroExpandScopeReferenceNameFunc                         func(*MacroExpandScopeReferenceName) T
	VisitMacroExpandOperatorFunc                                   func(*MacroExpandOperator) T
	VisitSummarizeByClauseFunc                                     func(*SummarizeByClause) T
	VisitDistinctOperatorFunc                                      func(*DistinctOperator) T
	VisitTakeOperatorFunc                                          func(*TakeOperator) T
	VisitSortOperatorFunc                                          func(*SortOperator) T
	VisitOrderedExpressionFunc                                     func(*OrderedExpression) T
	VisitOrderingClauseFunc                                        func(*OrderingClause) T
	VisitOrderingNullsClauseFunc                                   func(*OrderingNullsClause) T
	VisitTopHittersOperatorFunc                                    func(*TopHittersOperator) T
	VisitTopHittersByClauseFunc                                    func(*TopHittersByClause) T
	VisitTopOperatorFunc                                           func(*TopOperator) T
	VisitTopNestedOperatorFunc                                     func(*TopNestedOperator) T
	VisitTopNestedClauseFunc                                       func(*TopNestedClause) T
	VisitTopNestedWithOthersClauseFunc                             func(*TopNestedWithOthersClause) T
	VisitUnionOperatorFunc                                         func(*UnionOperator) T
	VisitAsOperatorFunc                                            func(*AsOperator) T
	VisitSerializeOperatorFunc                                     func(*SerializeOperator) T
	VisitInvokeOperatorFunc                                        func(*InvokeOperator) T
	VisitRenderOperatorFunc                                        func(*RenderOperator) T
	VisitMakeGraphOperatorFunc                                     func(*MakeGraphOperator) T
	VisitMakeGraphWithTablesAndKeysClauseFunc                      func(*MakeGraphWithTablesAndKeysClause) T
	VisitMakeGraphWithImplicitIdClauseFunc                         func(*MakeGraphWithImplicitIdClause) T
	VisitGraphMarkComponentsOperatorFunc                           func(*GraphMarkComponentsOperator) T
	VisitGraphWhereNodesOperatorFunc                               func(*GraphWhereNodesOperator) T
	VisitGraphWhereEdgesOperatorFunc                               func(*GraphWhereEdgesOperator) T
	VisitMakeGraphTableAndKeyClauseFunc                            func(*MakeGraphTableAndKeyClause) T
	VisitMakeGraphPartitionedByClauseFunc                          func(*MakeGraphPartitionedByClause) T
	VisitGraphToTableOperatorFunc                                  func(*GraphToTableOperator) T
	VisitGraphToTableOutputClauseFunc                              func(*GraphToTableOutputClause) T
	VisitGraphToTableAsClauseFunc                                  func(*GraphToTableAsClause) T
	VisitGraphMatchOperatorFunc                                    func(*GraphMatchOperator) T
	VisitGraphShortestPathsOperatorFunc                            func(*GraphShortestPathsOperator) T
	VisitGraphMatchPatternFunc                                     func(*GraphMatchPattern) T
	VisitGraphMatchPatternNodeFunc                                 func(*GraphMatchPatternNode) T
	VisitGraphMatchPatternEdgeFunc                                 func(*GraphMatchPatternEdge) T
	VisitGraphMatchPatternEdgeRangeFunc                            func(*GraphMatchPatternEdgeRange) T
	VisitWhereClauseFunc                                           func(*WhereClause) T
	VisitProjectClauseFunc                                         func(*ProjectClause) T
	VisitNameReferenceListFunc                                     func(*NameReferenceList) T
	VisitRenderWithClauseFunc                                      func(*RenderWithClause) T
	VisitPrintOperatorFunc                                         func(*PrintOperator) T
	VisitAliasStatementFunc                                        func(*AliasStatement) T
	VisitLetStatementFunc                                          func(*LetStatement) T
	VisitFunctionDeclarationFunc                                   func(*FunctionDeclaration) T
	VisitFunctionParametersFunc                                    func(*FunctionParameters) T
	VisitFunctionParameterFunc                                     func(*FunctionParameter) T
	VisitDefaultValueDeclarationFunc                               func(*DefaultValueDeclaration) T
	VisitFunctionBodyFunc                                          func(*FunctionBody) T
	VisitSchemaTypeExpressionFunc                                  func(*SchemaTypeExpression) T
	VisitExpressionStatementFunc                                   func(*ExpressionStatement) T
	VisitMaterializeExpressionFunc                                 func(*MaterializeExpression) T
	VisitSetOptionStatementFunc                                    func(*SetOptionStatement) T
	VisitOptionValueClauseFunc                                     func(*OptionValueClause) T
	VisitQueryParametersStatementFunc                              func(*QueryParametersStatement) T
	VisitRestrictStatementWithClauseFunc                           func(*RestrictStatementWithClause) T
	VisitRestrictStatementFunc                                     func(*RestrictStatement) T
	VisitPatternStatementFunc                                      func(*PatternStatement) T
	VisitPatternDeclarationFunc                                    func(*PatternDeclaration) T
	VisitPatternPathParameterFunc                                  func(*PatternPathParameter) T
	VisitPatternMatchFunc                                          func(*PatternMatch) T
	VisitPatternPathValueFunc                                      func(*PatternPathValue) T
	VisitDataScopeExpressionFunc                                   func(*DataScopeExpression) T
	VisitDataTableExpressionFunc                                   func(*DataTableExpression) T
	VisitRowSchemaFunc                                             func(*RowSchema) T
	VisitEvaluateRowSchemaFunc                                     func(*EvaluateRowSchema) T
	VisitExternalDataExpressionFunc                                func(*ExternalDataExpression) T
	VisitContextualDataTableExpressionFunc                         func(*ContextualDataTableExpression) T
	VisitExternalDataWithClauseFunc                                func(*ExternalDataWithClause) T
	VisitInlineExternalTableKindClauseFunc                         func(*InlineExternalTableKindClause) T
	VisitInlineExternalTableDataFormatClauseFunc                   func(*InlineExternalTableDataFormatClause) T
	VisitDateTimePatternFunc                                       func(*DateTimePattern) T
	VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc func(*InlineExternalTablePathFormatPartitionColumnReference) T
	VisitInlineExternalTablePathFormatClauseFunc                   func(*InlineExternalTablePathFormatClause) T
	VisitPartitionColumnDeclarationFunc                            func(*PartitionColumnDeclaration) T
	VisitInlineExternalTablePartitionClauseFunc                    func(*InlineExternalTablePartitionClause) T
	VisitInlineExternalTableConnectionStringsClauseFunc            func(*InlineExternalTableConnectionStringsClause) T
	VisitInlineExternalTableExpressionFunc                         func(*InlineExternalTableExpression) T
	VisitJoinOperatorFunc                                          func(*JoinOperator) T
	VisitLookupOperatorFunc                                        func(*LookupOperator) T
	VisitJoinOnClauseFunc                                          func(*JoinOnClause) T
	VisitJoinWhereClauseFunc                                       func(*JoinWhereClause) T
	VisitScanOperatorFunc                                          func(*ScanOperator) T
	VisitAssertSchemaOperatorFunc                                  func(*AssertSchemaOperator) T
	VisitScanDeclareClauseFunc                                     func(*ScanDeclareClause) T
	VisitScanOrderByClauseFunc                                     func(*ScanOrderByClause) T
	VisitScanPartitionByClauseFunc                                 func(*ScanPartitionByClause) T
	VisitScanStepOutputFunc                                        func(*ScanStepOutput) T
	VisitScanStepFunc                                              func(*ScanStep) T
	VisitScanComputationClauseFunc                                 func(*ScanComputationClause) T
	VisitScanAssignmentFunc                                        func(*ScanAssignment) T
	VisitPartitionByOperatorFunc                                   func(*PartitionByOperator) T
	VisitPartitionByIdClauseFunc                                   func(*PartitionByIdClause) T
	VisitCommandWithValueClauseFunc                                func(*CommandWithValueClause) T
	VisitCommandWithPropertyListClauseFunc                         func(*CommandWithPropertyListClause) T
	VisitUnknownCommandFunc                                        func(*UnknownCommand) T
	VisitCustomCommandFunc                                         func(*CustomCommand) T
	VisitPartialCommandFunc                                        func(*PartialCommand) T
	VisitCommandAndSkippedTokensFunc                               func(*CommandAndSkippedTokens) T
	VisitBadCommandFunc                                            func(*BadCommand) T
	VisitCommandBlockFunc                                          func(*CommandBlock) T
	VisitGraphMergeOperatorFunc                                    func(*GraphMergeOperator) T
}

func (v SyntaxVisitorResultFuncs[T]) defaultVisit(element SyntaxElement) T {
	if v.DefaultVisit != nil {
		return v.DefaultVisit(element)
	}
	var zero T
	return zero
}

func (v SyntaxVisitorResultFuncs[T]) VisitCustom(node *CustomNode) T {
	if v.VisitCustomFunc != nil {
		return v.VisitCustomFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitList(node *SyntaxList) T {
	if v.VisitListFunc != nil {
		return v.VisitListFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSeparatedElement(node *SeparatedElement) T {
	if v.VisitSeparatedElementFunc != nil {
		return v.VisitSeparatedElementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitQueryBlock(node *QueryBlock) T {
	if v.VisitQueryBlockFunc != nil {
		return v.VisitQueryBlockFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDirectiveBlock(node *DirectiveBlock) T {
	if v.VisitDirectiveBlockFunc != nil {
		return v.VisitDirectiveBlockFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDirective(node *Directive) T {
	if v.VisitDirectiveFunc != nil {
		return v.VisitDirectiveFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSkippedTokens(node *SkippedTokens) T {
	if v.VisitSkippedTokensFunc != nil {
		return v.VisitSkippedTokensFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTypeOfLiteralExpression(node *TypeOfLiteralExpression) T {
	if v.VisitTypeOfLiteralExpressionFunc != nil {
		return v.VisitTypeOfLiteralExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBadQueryOperator(node *BadQueryOperator) T {
	if v.VisitBadQueryOperatorFunc != nil {
		return v.VisitBadQueryOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCompoundStringLiteralExpression(node *CompoundStringLiteralExpression) T {
	if v.VisitCompoundStringLiteralExpressionFunc != nil {
		return v.VisitCompoundStringLiteralExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTokenName(node *TokenName) T {
	if v.VisitTokenNameFunc != nil {
		return v.VisitTokenNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBracketedName(node *BracketedName) T {
	if v.VisitBracketedNameFunc != nil {
		return v.VisitBracketedNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBracedName(node *BracedName) T {
	if v.VisitBracedNameFunc != nil {
		return v.VisitBracedNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitWildcardedName(node *WildcardedName) T {
	if v.VisitWildcardedNameFunc != nil {
		return v.VisitWildcardedNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBracketedWildcardedName(node *BracketedWildcardedName) T {
	if v.VisitBracketedWildcardedNameFunc != nil {
		return v.VisitBracketedWildcardedNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNameDeclaration(node *NameDeclaration) T {
	if v.VisitNameDeclarationFunc != nil {
		return v.VisitNameDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNameReference(node *NameReference) T {
	if v.VisitNameReferenceFunc != nil {
		return v.VisitNameReferenceFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitLiteralExpression(node *LiteralExpression) T {
	if v.VisitLiteralExpressionFunc != nil {
		return v.VisitLiteralExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitStarExpression(node *StarExpression) T {
	if v.VisitStarExpressionFunc != nil {
		return v.VisitStarExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitAtExpression(node *AtExpression) T {
	if v.VisitAtExpressionFunc != nil {
		return v.VisitAtExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJsonPair(node *JsonPair) T {
	if v.VisitJsonPairFunc != nil {
		return v.VisitJsonPairFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJsonObjectExpression(node *JsonObjectExpression) T {
	if v.VisitJsonObjectExpressionFunc != nil {
		return v.VisitJsonObjectExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJsonArrayExpression(node *JsonArrayExpression) T {
	if v.VisitJsonArrayExpressionFunc != nil {
		return v.VisitJsonArrayExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDynamicExpression(node *DynamicExpression) T {
	if v.VisitDynamicExpressionFunc != nil {
		return v.VisitDynamicExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitParenthesizedExpression(node *ParenthesizedExpression) T {
	if v.VisitParenthesizedExpressionFunc != nil {
		return v.VisitParenthesizedExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExpressionList(node *ExpressionList) T {
	if v.VisitExpressionListFunc != nil {
		return v.VisitExpressionListFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExpressionCouple(node *ExpressionCouple) T {
	if v.VisitExpressionCoupleFunc != nil {
		return v.VisitExpressionCoupleFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPrefixUnaryExpression(node *PrefixUnaryExpression) T {
	if v.VisitPrefixUnaryExpressionFunc != nil {
		return v.VisitPrefixUnaryExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBinaryExpression(node *BinaryExpression) T {
	if v.VisitBinaryExpressionFunc != nil {
		return v.VisitBinaryExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInExpression(node *InExpression) T {
	if v.VisitInExpressionFunc != nil {
		return v.VisitInExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitHasAnyExpression(node *HasAnyExpression) T {
	if v.VisitHasAnyExpressionFunc != nil {
		return v.VisitHasAnyExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitHasAllExpression(node *HasAllExpression) T {
	if v.VisitHasAllExpressionFunc != nil {
		return v.VisitHasAllExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBetweenExpression(node *BetweenExpression) T {
	if v.VisitBetweenExpressionFunc != nil {
		return v.VisitBetweenExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFunctionCallExpression(node *FunctionCallExpression) T {
	if v.VisitFunctionCallExpressionFunc != nil {
		return v.VisitFunctionCallExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitToScalarExpression(node *ToScalarExpression) T {
	if v.VisitToScalarExpressionFunc != nil {
		return v.VisitToScalarExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitToTableExpression(node *ToTableExpression) T {
	if v.VisitToTableExpressionFunc != nil {
		return v.VisitToTableExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMaterializedViewCombineExpression(node *MaterializedViewCombineExpression) T {
	if v.VisitMaterializedViewCombineExpressionFunc != nil {
		return v.VisitMaterializedViewCombineExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMaterializedViewCombineNameClause(node *MaterializedViewCombineNameClause) T {
	if v.VisitMaterializedViewCombineNameClauseFunc != nil {
		return v.VisitMaterializedViewCombineNameClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMaterializedViewCombineClause(node *MaterializedViewCombineClause) T {
	if v.VisitMaterializedViewCombineClauseFunc != nil {
		return v.VisitMaterializedViewCombineClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSimpleNamedExpression(node *SimpleNamedExpression) T {
	if v.VisitSimpleNamedExpressionFunc != nil {
		return v.VisitSimpleNamedExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRenameList(node *RenameList) T {
	if v.VisitRenameListFunc != nil {
		return v.VisitRenameListFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCompoundNamedExpression(node *CompoundNamedExpression) T {
	if v.VisitCompoundNamedExpressionFunc != nil {
		return v.VisitCompoundNamedExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBracketedExpression(node *BracketedExpression) T {
	if v.VisitBracketedExpressionFunc != nil {
		return v.VisitBracketedExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPathExpression(node *PathExpression) T {
	if v.VisitPathExpressionFunc != nil {
		return v.VisitPathExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitElementExpression(node *ElementExpression) T {
	if v.VisitElementExpressionFunc != nil {
		return v.VisitElementExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPipeExpression(node *PipeExpression) T {
	if v.VisitPipeExpressionFunc != nil {
		return v.VisitPipeExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRangeOperator(node *RangeOperator) T {
	if v.VisitRangeOperatorFunc != nil {
		return v.VisitRangeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNamedParameter(node *NamedParameter) T {
	if v.VisitNamedParameterFunc != nil {
		return v.VisitNamedParameterFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitConsumeOperator(node *ConsumeOperator) T {
	if v.VisitConsumeOperatorFunc != nil {
		return v.VisitConsumeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCountOperator(node *CountOperator) T {
	if v.VisitCountOperatorFunc != nil {
		return v.VisitCountOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCountAsIdentifierClause(node *CountAsIdentifierClause) T {
	if v.VisitCountAsIdentifierClauseFunc != nil {
		return v.VisitCountAsIdentifierClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExecuteAndCacheOperator(node *ExecuteAndCacheOperator) T {
	if v.VisitExecuteAndCacheOperatorFunc != nil {
		return v.VisitExecuteAndCacheOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExtendOperator(node *ExtendOperator) T {
	if v.VisitExtendOperatorFunc != nil {
		return v.VisitExtendOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFacetOperator(node *FacetOperator) T {
	if v.VisitFacetOperatorFunc != nil {
		return v.VisitFacetOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFacetWithOperatorClause(node *FacetWithOperatorClause) T {
	if v.VisitFacetWithOperatorClauseFunc != nil {
		return v.VisitFacetWithOperatorClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFacetWithExpressionClause(node *FacetWithExpressionClause) T {
	if v.VisitFacetWithExpressionClauseFunc != nil {
		return v.VisitFacetWithExpressionClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFilterOperator(node *FilterOperator) T {
	if v.VisitFilterOperatorFunc != nil {
		return v.VisitFilterOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGetSchemaOperator(node *GetSchemaOperator) T {
	if v.VisitGetSchemaOperatorFunc != nil {
		return v.VisitGetSchemaOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFindOperator(node *FindOperator) T {
	if v.VisitFindOperatorFunc != nil {
		return v.VisitFindOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDataScopeClause(node *DataScopeClause) T {
	if v.VisitDataScopeClauseFunc != nil {
		return v.VisitDataScopeClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTypedColumnReference(node *TypedColumnReference) T {
	if v.VisitTypedColumnReferenceFunc != nil {
		return v.VisitTypedColumnReferenceFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFindInClause(node *FindInClause) T {
	if v.VisitFindInClauseFunc != nil {
		return v.VisitFindInClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFindProjectClause(node *FindProjectClause) T {
	if v.VisitFindProjectClauseFunc != nil {
		return v.VisitFindProjectClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPackExpression(node *PackExpression) T {
	if v.VisitPackExpressionFunc != nil {
		return v.VisitPackExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNameAndTypeDeclaration(node *NameAndTypeDeclaration) T {
	if v.VisitNameAndTypeDeclarationFunc != nil {
		return v.VisitNameAndTypeDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPrimitiveTypeExpression(node *PrimitiveTypeExpression) T {
	if v.VisitPrimitiveTypeExpressionFunc != nil {
		return v.VisitPrimitiveTypeExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSearchOperator(node *SearchOperator) T {
	if v.VisitSearchOperatorFunc != nil {
		return v.VisitSearchOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitForkOperator(node *ForkOperator) T {
	if v.VisitForkOperatorFunc != nil {
		return v.VisitForkOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNameEqualsClause(node *NameEqualsClause) T {
	if v.VisitNameEqualsClauseFunc != nil {
		return v.VisitNameEqualsClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitForkExpression(node *ForkExpression) T {
	if v.VisitForkExpressionFunc != nil {
		return v.VisitForkExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesOperator(node *MakeSeriesOperator) T {
	if v.VisitMakeSeriesOperatorFunc != nil {
		return v.VisitMakeSeriesOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesExpression(node *MakeSeriesExpression) T {
	if v.VisitMakeSeriesExpressionFunc != nil {
		return v.VisitMakeSeriesExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDefaultExpressionClause(node *DefaultExpressionClause) T {
	if v.VisitDefaultExpressionClauseFunc != nil {
		return v.VisitDefaultExpressionClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesOnClause(node *MakeSeriesOnClause) T {
	if v.VisitMakeSeriesOnClauseFunc != nil {
		return v.VisitMakeSeriesOnClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesFromClause(node *MakeSeriesFromClause) T {
	if v.VisitMakeSeriesFromClauseFunc != nil {
		return v.VisitMakeSeriesFromClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesToClause(node *MakeSeriesToClause) T {
	if v.VisitMakeSeriesToClauseFunc != nil {
		return v.VisitMakeSeriesToClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesStepClause(node *MakeSeriesStepClause) T {
	if v.VisitMakeSeriesStepClauseFunc != nil {
		return v.VisitMakeSeriesStepClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesInRangeClause(node *MakeSeriesInRangeClause) T {
	if v.VisitMakeSeriesInRangeClauseFunc != nil {
		return v.VisitMakeSeriesInRangeClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesFromToStepClause(node *MakeSeriesFromToStepClause) T {
	if v.VisitMakeSeriesFromToStepClauseFunc != nil {
		return v.VisitMakeSeriesFromToStepClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeSeriesByClause(node *MakeSeriesByClause) T {
	if v.VisitMakeSeriesByClauseFunc != nil {
		return v.VisitMakeSeriesByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvExpandOperator(node *MvExpandOperator) T {
	if v.VisitMvExpandOperatorFunc != nil {
		return v.VisitMvExpandOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvExpandExpression(node *MvExpandExpression) T {
	if v.VisitMvExpandExpressionFunc != nil {
		return v.VisitMvExpandExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvExpandRowLimitClause(node *MvExpandRowLimitClause) T {
	if v.VisitMvExpandRowLimitClauseFunc != nil {
		return v.VisitMvExpandRowLimitClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvApplyOperator(node *MvApplyOperator) T {
	if v.VisitMvApplyOperatorFunc != nil {
		return v.VisitMvApplyOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvApplyExpression(node *MvApplyExpression) T {
	if v.VisitMvApplyExpressionFunc != nil {
		return v.VisitMvApplyExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvApplyRowLimitClause(node *MvApplyRowLimitClause) T {
	if v.VisitMvApplyRowLimitClauseFunc != nil {
		return v.VisitMvApplyRowLimitClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvApplyContextIdClause(node *MvApplyContextIdClause) T {
	if v.VisitMvApplyContextIdClauseFunc != nil {
		return v.VisitMvApplyContextIdClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMvApplySubqueryExpression(node *MvApplySubqueryExpression) T {
	if v.VisitMvApplySubqueryExpressionFunc != nil {
		return v.VisitMvApplySubqueryExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitToTypeOfClause(node *ToTypeOfClause) T {
	if v.VisitToTypeOfClauseFunc != nil {
		return v.VisitToTypeOfClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitEvaluateSchemaClause(node *EvaluateSchemaClause) T {
	if v.VisitEvaluateSchemaClauseFunc != nil {
		return v.VisitEvaluateSchemaClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitEvaluateOperator(node *EvaluateOperator) T {
	if v.VisitEvaluateOperatorFunc != nil {
		return v.VisitEvaluateOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitParseOperator(node *ParseOperator) T {
	if v.VisitParseOperatorFunc != nil {
		return v.VisitParseOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitParseWhereOperator(node *ParseWhereOperator) T {
	if v.VisitParseWhereOperatorFunc != nil {
		return v.VisitParseWhereOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitParseKvWithClause(node *ParseKvWithClause) T {
	if v.VisitParseKvWithClauseFunc != nil {
		return v.VisitParseKvWithClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitParseKvOperator(node *ParseKvOperator) T {
	if v.VisitParseKvOperatorFunc != nil {
		return v.VisitParseKvOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionOperator(node *PartitionOperator) T {
	if v.VisitPartitionOperatorFunc != nil {
		return v.VisitPartitionOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionQuery(node *PartitionQuery) T {
	if v.VisitPartitionQueryFunc != nil {
		return v.VisitPartitionQueryFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionScope(node *PartitionScope) T {
	if v.VisitPartitionScopeFunc != nil {
		return v.VisitPartitionScopeFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionSubquery(node *PartitionSubquery) T {
	if v.VisitPartitionSubqueryFunc != nil {
		return v.VisitPartitionSubqueryFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectOperator(node *ProjectOperator) T {
	if v.VisitProjectOperatorFunc != nil {
		return v.VisitProjectOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectAwayOperator(node *ProjectAwayOperator) T {
	if v.VisitProjectAwayOperatorFunc != nil {
		return v.VisitProjectAwayOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectByNamesOperator(node *ProjectByNamesOperator) T {
	if v.VisitProjectByNamesOperatorFunc != nil {
		return v.VisitProjectByNamesOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectKeepOperator(node *ProjectKeepOperator) T {
	if v.VisitProjectKeepOperatorFunc != nil {
		return v.VisitProjectKeepOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectRenameOperator(node *ProjectRenameOperator) T {
	if v.VisitProjectRenameOperatorFunc != nil {
		return v.VisitProjectRenameOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectReorderOperator(node *ProjectReorderOperator) T {
	if v.VisitProjectReorderOperatorFunc != nil {
		return v.VisitProjectReorderOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSampleOperator(node *SampleOperator) T {
	if v.VisitSampleOperatorFunc != nil {
		return v.VisitSampleOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSampleDistinctOperator(node *SampleDistinctOperator) T {
	if v.VisitSampleDistinctOperatorFunc != nil {
		return v.VisitSampleDistinctOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitEntityGroup(node *EntityGroup) T {
	if v.VisitEntityGroupFunc != nil {
		return v.VisitEntityGroupFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitReduceByOperator(node *ReduceByOperator) T {
	if v.VisitReduceByOperatorFunc != nil {
		return v.VisitReduceByOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitReduceByWithClause(node *ReduceByWithClause) T {
	if v.VisitReduceByWithClauseFunc != nil {
		return v.VisitReduceByWithClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSummarizeOperator(node *SummarizeOperator) T {
	if v.VisitSummarizeOperatorFunc != nil {
		return v.VisitSummarizeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMacroExpandScopeReferenceName(node *MacroExpandScopeReferenceName) T {
	if v.VisitMacroExpandScopeReferenceNameFunc != nil {
		return v.VisitMacroExpandScopeReferenceNameFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMacroExpandOperator(node *MacroExpandOperator) T {
	if v.VisitMacroExpandOperatorFunc != nil {
		return v.VisitMacroExpandOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSummarizeByClause(node *SummarizeByClause) T {
	if v.VisitSummarizeByClauseFunc != nil {
		return v.VisitSummarizeByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDistinctOperator(node *DistinctOperator) T {
	if v.VisitDistinctOperatorFunc != nil {
		return v.VisitDistinctOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTakeOperator(node *TakeOperator) T {
	if v.VisitTakeOperatorFunc != nil {
		return v.VisitTakeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSortOperator(node *SortOperator) T {
	if v.VisitSortOperatorFunc != nil {
		return v.VisitSortOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitOrderedExpression(node *OrderedExpression) T {
	if v.VisitOrderedExpressionFunc != nil {
		return v.VisitOrderedExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitOrderingClause(node *OrderingClause) T {
	if v.VisitOrderingClauseFunc != nil {
		return v.VisitOrderingClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitOrderingNullsClause(node *OrderingNullsClause) T {
	if v.VisitOrderingNullsClauseFunc != nil {
		return v.VisitOrderingNullsClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopHittersOperator(node *TopHittersOperator) T {
	if v.VisitTopHittersOperatorFunc != nil {
		return v.VisitTopHittersOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopHittersByClause(node *TopHittersByClause) T {
	if v.VisitTopHittersByClauseFunc != nil {
		return v.VisitTopHittersByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopOperator(node *TopOperator) T {
	if v.VisitTopOperatorFunc != nil {
		return v.VisitTopOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopNestedOperator(node *TopNestedOperator) T {
	if v.VisitTopNestedOperatorFunc != nil {
		return v.VisitTopNestedOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopNestedClause(node *TopNestedClause) T {
	if v.VisitTopNestedClauseFunc != nil {
		return v.VisitTopNestedClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitTopNestedWithOthersClause(node *TopNestedWithOthersClause) T {
	if v.VisitTopNestedWithOthersClauseFunc != nil {
		return v.VisitTopNestedWithOthersClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitUnionOperator(node *UnionOperator) T {
	if v.VisitUnionOperatorFunc != nil {
		return v.VisitUnionOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitAsOperator(node *AsOperator) T {
	if v.VisitAsOperatorFunc != nil {
		return v.VisitAsOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSerializeOperator(node *SerializeOperator) T {
	if v.VisitSerializeOperatorFunc != nil {
		return v.VisitSerializeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInvokeOperator(node *InvokeOperator) T {
	if v.VisitInvokeOperatorFunc != nil {
		return v.VisitInvokeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRenderOperator(node *RenderOperator) T {
	if v.VisitRenderOperatorFunc != nil {
		return v.VisitRenderOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeGraphOperator(node *MakeGraphOperator) T {
	if v.VisitMakeGraphOperatorFunc != nil {
		return v.VisitMakeGraphOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeGraphWithTablesAndKeysClause(node *MakeGraphWithTablesAndKeysClause) T {
	if v.VisitMakeGraphWithTablesAndKeysClauseFunc != nil {
		return v.VisitMakeGraphWithTablesAndKeysClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeGraphWithImplicitIdClause(node *MakeGraphWithImplicitIdClause) T {
	if v.VisitMakeGraphWithImplicitIdClauseFunc != nil {
		return v.VisitMakeGraphWithImplicitIdClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMarkComponentsOperator(node *GraphMarkComponentsOperator) T {
	if v.VisitGraphMarkComponentsOperatorFunc != nil {
		return v.VisitGraphMarkComponentsOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphWhereNodesOperator(node *GraphWhereNodesOperator) T {
	if v.VisitGraphWhereNodesOperatorFunc != nil {
		return v.VisitGraphWhereNodesOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphWhereEdgesOperator(node *GraphWhereEdgesOperator) T {
	if v.VisitGraphWhereEdgesOperatorFunc != nil {
		return v.VisitGraphWhereEdgesOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeGraphTableAndKeyClause(node *MakeGraphTableAndKeyClause) T {
	if v.VisitMakeGraphTableAndKeyClauseFunc != nil {
		return v.VisitMakeGraphTableAndKeyClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMakeGraphPartitionedByClause(node *MakeGraphPartitionedByClause) T {
	if v.VisitMakeGraphPartitionedByClauseFunc != nil {
		return v.VisitMakeGraphPartitionedByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphToTableOperator(node *GraphToTableOperator) T {
	if v.VisitGraphToTableOperatorFunc != nil {
		return v.VisitGraphToTableOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphToTableOutputClause(node *GraphToTableOutputClause) T {
	if v.VisitGraphToTableOutputClauseFunc != nil {
		return v.VisitGraphToTableOutputClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphToTableAsClause(node *GraphToTableAsClause) T {
	if v.VisitGraphToTableAsClauseFunc != nil {
		return v.VisitGraphToTableAsClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMatchOperator(node *GraphMatchOperator) T {
	if v.VisitGraphMatchOperatorFunc != nil {
		return v.VisitGraphMatchOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphShortestPathsOperator(node *GraphShortestPathsOperator) T {
	if v.VisitGraphShortestPathsOperatorFunc != nil {
		return v.VisitGraphShortestPathsOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMatchPattern(node *GraphMatchPattern) T {
	if v.VisitGraphMatchPatternFunc != nil {
		return v.VisitGraphMatchPatternFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMatchPatternNode(node *GraphMatchPatternNode) T {
	if v.VisitGraphMatchPatternNodeFunc != nil {
		return v.VisitGraphMatchPatternNodeFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMatchPatternEdge(node *GraphMatchPatternEdge) T {
	if v.VisitGraphMatchPatternEdgeFunc != nil {
		return v.VisitGraphMatchPatternEdgeFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMatchPatternEdgeRange(node *GraphMatchPatternEdgeRange) T {
	if v.VisitGraphMatchPatternEdgeRangeFunc != nil {
		return v.VisitGraphMatchPatternEdgeRangeFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitWhereClause(node *WhereClause) T {
	if v.VisitWhereClauseFunc != nil {
		return v.VisitWhereClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitProjectClause(node *ProjectClause) T {
	if v.VisitProjectClauseFunc != nil {
		return v.VisitProjectClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitNameReferenceList(node *NameReferenceList) T {
	if v.VisitNameReferenceListFunc != nil {
		return v.VisitNameReferenceListFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRenderWithClause(node *RenderWithClause) T {
	if v.VisitRenderWithClauseFunc != nil {
		return v.VisitRenderWithClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPrintOperator(node *PrintOperator) T {
	if v.VisitPrintOperatorFunc != nil {
		return v.VisitPrintOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitAliasStatement(node *AliasStatement) T {
	if v.VisitAliasStatementFunc != nil {
		return v.VisitAliasStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitLetStatement(node *LetStatement) T {
	if v.VisitLetStatementFunc != nil {
		return v.VisitLetStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFunctionDeclaration(node *FunctionDeclaration) T {
	if v.VisitFunctionDeclarationFunc != nil {
		return v.VisitFunctionDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFunctionParameters(node *FunctionParameters) T {
	if v.VisitFunctionParametersFunc != nil {
		return v.VisitFunctionParametersFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFunctionParameter(node *FunctionParameter) T {
	if v.VisitFunctionParameterFunc != nil {
		return v.VisitFunctionParameterFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDefaultValueDeclaration(node *DefaultValueDeclaration) T {
	if v.VisitDefaultValueDeclarationFunc != nil {
		return v.VisitDefaultValueDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitFunctionBody(node *FunctionBody) T {
	if v.VisitFunctionBodyFunc != nil {
		return v.VisitFunctionBodyFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSchemaTypeExpression(node *SchemaTypeExpression) T {
	if v.VisitSchemaTypeExpressionFunc != nil {
		return v.VisitSchemaTypeExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExpressionStatement(node *ExpressionStatement) T {
	if v.VisitExpressionStatementFunc != nil {
		return v.VisitExpressionStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitMaterializeExpression(node *MaterializeExpression) T {
	if v.VisitMaterializeExpressionFunc != nil {
		return v.VisitMaterializeExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitSetOptionStatement(node *SetOptionStatement) T {
	if v.VisitSetOptionStatementFunc != nil {
		return v.VisitSetOptionStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitOptionValueClause(node *OptionValueClause) T {
	if v.VisitOptionValueClauseFunc != nil {
		return v.VisitOptionValueClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitQueryParametersStatement(node *QueryParametersStatement) T {
	if v.VisitQueryParametersStatementFunc != nil {
		return v.VisitQueryParametersStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRestrictStatementWithClause(node *RestrictStatementWithClause) T {
	if v.VisitRestrictStatementWithClauseFunc != nil {
		return v.VisitRestrictStatementWithClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRestrictStatement(node *RestrictStatement) T {
	if v.VisitRestrictStatementFunc != nil {
		return v.VisitRestrictStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPatternStatement(node *PatternStatement) T {
	if v.VisitPatternStatementFunc != nil {
		return v.VisitPatternStatementFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPatternDeclaration(node *PatternDeclaration) T {
	if v.VisitPatternDeclarationFunc != nil {
		return v.VisitPatternDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPatternPathParameter(node *PatternPathParameter) T {
	if v.VisitPatternPathParameterFunc != nil {
		return v.VisitPatternPathParameterFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPatternMatch(node *PatternMatch) T {
	if v.VisitPatternMatchFunc != nil {
		return v.VisitPatternMatchFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPatternPathValue(node *PatternPathValue) T {
	if v.VisitPatternPathValueFunc != nil {
		return v.VisitPatternPathValueFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDataScopeExpression(node *DataScopeExpression) T {
	if v.VisitDataScopeExpressionFunc != nil {
		return v.VisitDataScopeExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDataTableExpression(node *DataTableExpression) T {
	if v.VisitDataTableExpressionFunc != nil {
		return v.VisitDataTableExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitRowSchema(node *RowSchema) T {
	if v.VisitRowSchemaFunc != nil {
		return v.VisitRowSchemaFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitEvaluateRowSchema(node *EvaluateRowSchema) T {
	if v.VisitEvaluateRowSchemaFunc != nil {
		return v.VisitEvaluateRowSchemaFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExternalDataExpression(node *ExternalDataExpression) T {
	if v.VisitExternalDataExpressionFunc != nil {
		return v.VisitExternalDataExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitContextualDataTableExpression(node *ContextualDataTableExpression) T {
	if v.VisitContextualDataTableExpressionFunc != nil {
		return v.VisitContextualDataTableExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitExternalDataWithClause(node *ExternalDataWithClause) T {
	if v.VisitExternalDataWithClauseFunc != nil {
		return v.VisitExternalDataWithClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTableKindClause(node *InlineExternalTableKindClause) T {
	if v.VisitInlineExternalTableKindClauseFunc != nil {
		return v.VisitInlineExternalTableKindClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTableDataFormatClause(node *InlineExternalTableDataFormatClause) T {
	if v.VisitInlineExternalTableDataFormatClauseFunc != nil {
		return v.VisitInlineExternalTableDataFormatClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitDateTimePattern(node *DateTimePattern) T {
	if v.VisitDateTimePatternFunc != nil {
		return v.VisitDateTimePatternFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTablePathFormatPartitionColumnReference(node *InlineExternalTablePathFormatPartitionColumnReference) T {
	if v.VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc != nil {
		return v.VisitInlineExternalTablePathFormatPartitionColumnReferenceFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTablePathFormatClause(node *InlineExternalTablePathFormatClause) T {
	if v.VisitInlineExternalTablePathFormatClauseFunc != nil {
		return v.VisitInlineExternalTablePathFormatClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionColumnDeclaration(node *PartitionColumnDeclaration) T {
	if v.VisitPartitionColumnDeclarationFunc != nil {
		return v.VisitPartitionColumnDeclarationFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTablePartitionClause(node *InlineExternalTablePartitionClause) T {
	if v.VisitInlineExternalTablePartitionClauseFunc != nil {
		return v.VisitInlineExternalTablePartitionClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTableConnectionStringsClause(node *InlineExternalTableConnectionStringsClause) T {
	if v.VisitInlineExternalTableConnectionStringsClauseFunc != nil {
		return v.VisitInlineExternalTableConnectionStringsClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitInlineExternalTableExpression(node *InlineExternalTableExpression) T {
	if v.VisitInlineExternalTableExpressionFunc != nil {
		return v.VisitInlineExternalTableExpressionFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJoinOperator(node *JoinOperator) T {
	if v.VisitJoinOperatorFunc != nil {
		return v.VisitJoinOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitLookupOperator(node *LookupOperator) T {
	if v.VisitLookupOperatorFunc != nil {
		return v.VisitLookupOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJoinOnClause(node *JoinOnClause) T {
	if v.VisitJoinOnClauseFunc != nil {
		return v.VisitJoinOnClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitJoinWhereClause(node *JoinWhereClause) T {
	if v.VisitJoinWhereClauseFunc != nil {
		return v.VisitJoinWhereClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanOperator(node *ScanOperator) T {
	if v.VisitScanOperatorFunc != nil {
		return v.VisitScanOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitAssertSchemaOperator(node *AssertSchemaOperator) T {
	if v.VisitAssertSchemaOperatorFunc != nil {
		return v.VisitAssertSchemaOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanDeclareClause(node *ScanDeclareClause) T {
	if v.VisitScanDeclareClauseFunc != nil {
		return v.VisitScanDeclareClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanOrderByClause(node *ScanOrderByClause) T {
	if v.VisitScanOrderByClauseFunc != nil {
		return v.VisitScanOrderByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanPartitionByClause(node *ScanPartitionByClause) T {
	if v.VisitScanPartitionByClauseFunc != nil {
		return v.VisitScanPartitionByClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanStepOutput(node *ScanStepOutput) T {
	if v.VisitScanStepOutputFunc != nil {
		return v.VisitScanStepOutputFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanStep(node *ScanStep) T {
	if v.VisitScanStepFunc != nil {
		return v.VisitScanStepFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanComputationClause(node *ScanComputationClause) T {
	if v.VisitScanComputationClauseFunc != nil {
		return v.VisitScanComputationClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitScanAssignment(node *ScanAssignment) T {
	if v.VisitScanAssignmentFunc != nil {
		return v.VisitScanAssignmentFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionByOperator(node *PartitionByOperator) T {
	if v.VisitPartitionByOperatorFunc != nil {
		return v.VisitPartitionByOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartitionByIdClause(node *PartitionByIdClause) T {
	if v.VisitPartitionByIdClauseFunc != nil {
		return v.VisitPartitionByIdClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCommandWithValueClause(node *CommandWithValueClause) T {
	if v.VisitCommandWithValueClauseFunc != nil {
		return v.VisitCommandWithValueClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCommandWithPropertyListClause(node *CommandWithPropertyListClause) T {
	if v.VisitCommandWithPropertyListClauseFunc != nil {
		return v.VisitCommandWithPropertyListClauseFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitUnknownCommand(node *UnknownCommand) T {
	if v.VisitUnknownCommandFunc != nil {
		return v.VisitUnknownCommandFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCustomCommand(node *CustomCommand) T {
	if v.VisitCustomCommandFunc != nil {
		return v.VisitCustomCommandFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitPartialCommand(node *PartialCommand) T {
	if v.VisitPartialCommandFunc != nil {
		return v.VisitPartialCommandFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCommandAndSkippedTokens(node *CommandAndSkippedTokens) T {
	if v.VisitCommandAndSkippedTokensFunc != nil {
		return v.VisitCommandAndSkippedTokensFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitBadCommand(node *BadCommand) T {
	if v.VisitBadCommandFunc != nil {
		return v.VisitBadCommandFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitCommandBlock(node *CommandBlock) T {
	if v.VisitCommandBlockFunc != nil {
		return v.VisitCommandBlockFunc(node)
	}
	return v.defaultVisit(node)
}

func (v SyntaxVisitorResultFuncs[T]) VisitGraphMergeOperator(node *GraphMergeOperator) T {
	if v.VisitGraphMergeOperatorFunc != nil {
		return v.VisitGraphMergeOperatorFunc(node)
	}
	return v.defaultVisit(node)
}

func AcceptSyntaxVisitor(element SyntaxElement, visitor SyntaxVisitor) {
	if element == nil || visitor == nil {
		return
	}
	switch node := element.(type) {
	case *CustomNode:
		visitor.VisitCustom(node)
		return
	case *SyntaxList:
		visitor.VisitList(node)
		return
	case *SeparatedElement:
		visitor.VisitSeparatedElement(node)
		return
	case *QueryBlock:
		visitor.VisitQueryBlock(node)
		return
	case *DirectiveBlock:
		visitor.VisitDirectiveBlock(node)
		return
	case *Directive:
		visitor.VisitDirective(node)
		return
	case *SkippedTokens:
		visitor.VisitSkippedTokens(node)
		return
	case *TypeOfLiteralExpression:
		visitor.VisitTypeOfLiteralExpression(node)
		return
	case *BadQueryOperator:
		visitor.VisitBadQueryOperator(node)
		return
	case *CompoundStringLiteralExpression:
		visitor.VisitCompoundStringLiteralExpression(node)
		return
	case *TokenName:
		visitor.VisitTokenName(node)
		return
	case *BracketedName:
		visitor.VisitBracketedName(node)
		return
	case *BracedName:
		visitor.VisitBracedName(node)
		return
	case *WildcardedName:
		visitor.VisitWildcardedName(node)
		return
	case *BracketedWildcardedName:
		visitor.VisitBracketedWildcardedName(node)
		return
	case *NameDeclaration:
		visitor.VisitNameDeclaration(node)
		return
	case *NameReference:
		visitor.VisitNameReference(node)
		return
	case *LiteralExpression:
		visitor.VisitLiteralExpression(node)
		return
	case *StarExpression:
		visitor.VisitStarExpression(node)
		return
	case *AtExpression:
		visitor.VisitAtExpression(node)
		return
	case *JsonPair:
		visitor.VisitJsonPair(node)
		return
	case *JsonObjectExpression:
		visitor.VisitJsonObjectExpression(node)
		return
	case *JsonArrayExpression:
		visitor.VisitJsonArrayExpression(node)
		return
	case *DynamicExpression:
		visitor.VisitDynamicExpression(node)
		return
	case *ParenthesizedExpression:
		visitor.VisitParenthesizedExpression(node)
		return
	case *ExpressionList:
		visitor.VisitExpressionList(node)
		return
	case *ExpressionCouple:
		visitor.VisitExpressionCouple(node)
		return
	case *PrefixUnaryExpression:
		visitor.VisitPrefixUnaryExpression(node)
		return
	case *BinaryExpression:
		visitor.VisitBinaryExpression(node)
		return
	case *InExpression:
		visitor.VisitInExpression(node)
		return
	case *HasAnyExpression:
		visitor.VisitHasAnyExpression(node)
		return
	case *HasAllExpression:
		visitor.VisitHasAllExpression(node)
		return
	case *BetweenExpression:
		visitor.VisitBetweenExpression(node)
		return
	case *FunctionCallExpression:
		visitor.VisitFunctionCallExpression(node)
		return
	case *ToScalarExpression:
		visitor.VisitToScalarExpression(node)
		return
	case *ToTableExpression:
		visitor.VisitToTableExpression(node)
		return
	case *MaterializedViewCombineExpression:
		visitor.VisitMaterializedViewCombineExpression(node)
		return
	case *MaterializedViewCombineNameClause:
		visitor.VisitMaterializedViewCombineNameClause(node)
		return
	case *MaterializedViewCombineClause:
		visitor.VisitMaterializedViewCombineClause(node)
		return
	case *SimpleNamedExpression:
		visitor.VisitSimpleNamedExpression(node)
		return
	case *RenameList:
		visitor.VisitRenameList(node)
		return
	case *CompoundNamedExpression:
		visitor.VisitCompoundNamedExpression(node)
		return
	case *BracketedExpression:
		visitor.VisitBracketedExpression(node)
		return
	case *PathExpression:
		visitor.VisitPathExpression(node)
		return
	case *ElementExpression:
		visitor.VisitElementExpression(node)
		return
	case *PipeExpression:
		visitor.VisitPipeExpression(node)
		return
	case *RangeOperator:
		visitor.VisitRangeOperator(node)
		return
	case *NamedParameter:
		visitor.VisitNamedParameter(node)
		return
	case *ConsumeOperator:
		visitor.VisitConsumeOperator(node)
		return
	case *CountOperator:
		visitor.VisitCountOperator(node)
		return
	case *CountAsIdentifierClause:
		visitor.VisitCountAsIdentifierClause(node)
		return
	case *ExecuteAndCacheOperator:
		visitor.VisitExecuteAndCacheOperator(node)
		return
	case *ExtendOperator:
		visitor.VisitExtendOperator(node)
		return
	case *FacetOperator:
		visitor.VisitFacetOperator(node)
		return
	case *FacetWithOperatorClause:
		visitor.VisitFacetWithOperatorClause(node)
		return
	case *FacetWithExpressionClause:
		visitor.VisitFacetWithExpressionClause(node)
		return
	case *FilterOperator:
		visitor.VisitFilterOperator(node)
		return
	case *GetSchemaOperator:
		visitor.VisitGetSchemaOperator(node)
		return
	case *FindOperator:
		visitor.VisitFindOperator(node)
		return
	case *DataScopeClause:
		visitor.VisitDataScopeClause(node)
		return
	case *TypedColumnReference:
		visitor.VisitTypedColumnReference(node)
		return
	case *FindInClause:
		visitor.VisitFindInClause(node)
		return
	case *FindProjectClause:
		visitor.VisitFindProjectClause(node)
		return
	case *PackExpression:
		visitor.VisitPackExpression(node)
		return
	case *NameAndTypeDeclaration:
		visitor.VisitNameAndTypeDeclaration(node)
		return
	case *PrimitiveTypeExpression:
		visitor.VisitPrimitiveTypeExpression(node)
		return
	case *SearchOperator:
		visitor.VisitSearchOperator(node)
		return
	case *ForkOperator:
		visitor.VisitForkOperator(node)
		return
	case *NameEqualsClause:
		visitor.VisitNameEqualsClause(node)
		return
	case *ForkExpression:
		visitor.VisitForkExpression(node)
		return
	case *MakeSeriesOperator:
		visitor.VisitMakeSeriesOperator(node)
		return
	case *MakeSeriesExpression:
		visitor.VisitMakeSeriesExpression(node)
		return
	case *DefaultExpressionClause:
		visitor.VisitDefaultExpressionClause(node)
		return
	case *MakeSeriesOnClause:
		visitor.VisitMakeSeriesOnClause(node)
		return
	case *MakeSeriesFromClause:
		visitor.VisitMakeSeriesFromClause(node)
		return
	case *MakeSeriesToClause:
		visitor.VisitMakeSeriesToClause(node)
		return
	case *MakeSeriesStepClause:
		visitor.VisitMakeSeriesStepClause(node)
		return
	case *MakeSeriesInRangeClause:
		visitor.VisitMakeSeriesInRangeClause(node)
		return
	case *MakeSeriesFromToStepClause:
		visitor.VisitMakeSeriesFromToStepClause(node)
		return
	case *MakeSeriesByClause:
		visitor.VisitMakeSeriesByClause(node)
		return
	case *MvExpandOperator:
		visitor.VisitMvExpandOperator(node)
		return
	case *MvExpandExpression:
		visitor.VisitMvExpandExpression(node)
		return
	case *MvExpandRowLimitClause:
		visitor.VisitMvExpandRowLimitClause(node)
		return
	case *MvApplyOperator:
		visitor.VisitMvApplyOperator(node)
		return
	case *MvApplyExpression:
		visitor.VisitMvApplyExpression(node)
		return
	case *MvApplyRowLimitClause:
		visitor.VisitMvApplyRowLimitClause(node)
		return
	case *MvApplyContextIdClause:
		visitor.VisitMvApplyContextIdClause(node)
		return
	case *MvApplySubqueryExpression:
		visitor.VisitMvApplySubqueryExpression(node)
		return
	case *ToTypeOfClause:
		visitor.VisitToTypeOfClause(node)
		return
	case *EvaluateSchemaClause:
		visitor.VisitEvaluateSchemaClause(node)
		return
	case *EvaluateOperator:
		visitor.VisitEvaluateOperator(node)
		return
	case *ParseOperator:
		visitor.VisitParseOperator(node)
		return
	case *ParseWhereOperator:
		visitor.VisitParseWhereOperator(node)
		return
	case *ParseKvWithClause:
		visitor.VisitParseKvWithClause(node)
		return
	case *ParseKvOperator:
		visitor.VisitParseKvOperator(node)
		return
	case *PartitionOperator:
		visitor.VisitPartitionOperator(node)
		return
	case *PartitionQuery:
		visitor.VisitPartitionQuery(node)
		return
	case *PartitionScope:
		visitor.VisitPartitionScope(node)
		return
	case *PartitionSubquery:
		visitor.VisitPartitionSubquery(node)
		return
	case *ProjectOperator:
		visitor.VisitProjectOperator(node)
		return
	case *ProjectAwayOperator:
		visitor.VisitProjectAwayOperator(node)
		return
	case *ProjectByNamesOperator:
		visitor.VisitProjectByNamesOperator(node)
		return
	case *ProjectKeepOperator:
		visitor.VisitProjectKeepOperator(node)
		return
	case *ProjectRenameOperator:
		visitor.VisitProjectRenameOperator(node)
		return
	case *ProjectReorderOperator:
		visitor.VisitProjectReorderOperator(node)
		return
	case *SampleOperator:
		visitor.VisitSampleOperator(node)
		return
	case *SampleDistinctOperator:
		visitor.VisitSampleDistinctOperator(node)
		return
	case *EntityGroup:
		visitor.VisitEntityGroup(node)
		return
	case *ReduceByOperator:
		visitor.VisitReduceByOperator(node)
		return
	case *ReduceByWithClause:
		visitor.VisitReduceByWithClause(node)
		return
	case *SummarizeOperator:
		visitor.VisitSummarizeOperator(node)
		return
	case *MacroExpandScopeReferenceName:
		visitor.VisitMacroExpandScopeReferenceName(node)
		return
	case *MacroExpandOperator:
		visitor.VisitMacroExpandOperator(node)
		return
	case *SummarizeByClause:
		visitor.VisitSummarizeByClause(node)
		return
	case *DistinctOperator:
		visitor.VisitDistinctOperator(node)
		return
	case *TakeOperator:
		visitor.VisitTakeOperator(node)
		return
	case *SortOperator:
		visitor.VisitSortOperator(node)
		return
	case *OrderedExpression:
		visitor.VisitOrderedExpression(node)
		return
	case *OrderingClause:
		visitor.VisitOrderingClause(node)
		return
	case *OrderingNullsClause:
		visitor.VisitOrderingNullsClause(node)
		return
	case *TopHittersOperator:
		visitor.VisitTopHittersOperator(node)
		return
	case *TopHittersByClause:
		visitor.VisitTopHittersByClause(node)
		return
	case *TopOperator:
		visitor.VisitTopOperator(node)
		return
	case *TopNestedOperator:
		visitor.VisitTopNestedOperator(node)
		return
	case *TopNestedClause:
		visitor.VisitTopNestedClause(node)
		return
	case *TopNestedWithOthersClause:
		visitor.VisitTopNestedWithOthersClause(node)
		return
	case *UnionOperator:
		visitor.VisitUnionOperator(node)
		return
	case *AsOperator:
		visitor.VisitAsOperator(node)
		return
	case *SerializeOperator:
		visitor.VisitSerializeOperator(node)
		return
	case *InvokeOperator:
		visitor.VisitInvokeOperator(node)
		return
	case *RenderOperator:
		visitor.VisitRenderOperator(node)
		return
	case *MakeGraphOperator:
		visitor.VisitMakeGraphOperator(node)
		return
	case *MakeGraphWithTablesAndKeysClause:
		visitor.VisitMakeGraphWithTablesAndKeysClause(node)
		return
	case *MakeGraphWithImplicitIdClause:
		visitor.VisitMakeGraphWithImplicitIdClause(node)
		return
	case *GraphMarkComponentsOperator:
		visitor.VisitGraphMarkComponentsOperator(node)
		return
	case *GraphWhereNodesOperator:
		visitor.VisitGraphWhereNodesOperator(node)
		return
	case *GraphWhereEdgesOperator:
		visitor.VisitGraphWhereEdgesOperator(node)
		return
	case *MakeGraphTableAndKeyClause:
		visitor.VisitMakeGraphTableAndKeyClause(node)
		return
	case *MakeGraphPartitionedByClause:
		visitor.VisitMakeGraphPartitionedByClause(node)
		return
	case *GraphToTableOperator:
		visitor.VisitGraphToTableOperator(node)
		return
	case *GraphToTableOutputClause:
		visitor.VisitGraphToTableOutputClause(node)
		return
	case *GraphToTableAsClause:
		visitor.VisitGraphToTableAsClause(node)
		return
	case *GraphMatchOperator:
		visitor.VisitGraphMatchOperator(node)
		return
	case *GraphShortestPathsOperator:
		visitor.VisitGraphShortestPathsOperator(node)
		return
	case *GraphMatchPattern:
		visitor.VisitGraphMatchPattern(node)
		return
	case *GraphMatchPatternNode:
		visitor.VisitGraphMatchPatternNode(node)
		return
	case *GraphMatchPatternEdge:
		visitor.VisitGraphMatchPatternEdge(node)
		return
	case *GraphMatchPatternEdgeRange:
		visitor.VisitGraphMatchPatternEdgeRange(node)
		return
	case *WhereClause:
		visitor.VisitWhereClause(node)
		return
	case *ProjectClause:
		visitor.VisitProjectClause(node)
		return
	case *NameReferenceList:
		visitor.VisitNameReferenceList(node)
		return
	case *RenderWithClause:
		visitor.VisitRenderWithClause(node)
		return
	case *PrintOperator:
		visitor.VisitPrintOperator(node)
		return
	case *AliasStatement:
		visitor.VisitAliasStatement(node)
		return
	case *LetStatement:
		visitor.VisitLetStatement(node)
		return
	case *FunctionDeclaration:
		visitor.VisitFunctionDeclaration(node)
		return
	case *FunctionParameters:
		visitor.VisitFunctionParameters(node)
		return
	case *FunctionParameter:
		visitor.VisitFunctionParameter(node)
		return
	case *DefaultValueDeclaration:
		visitor.VisitDefaultValueDeclaration(node)
		return
	case *FunctionBody:
		visitor.VisitFunctionBody(node)
		return
	case *SchemaTypeExpression:
		visitor.VisitSchemaTypeExpression(node)
		return
	case *ExpressionStatement:
		visitor.VisitExpressionStatement(node)
		return
	case *MaterializeExpression:
		visitor.VisitMaterializeExpression(node)
		return
	case *SetOptionStatement:
		visitor.VisitSetOptionStatement(node)
		return
	case *OptionValueClause:
		visitor.VisitOptionValueClause(node)
		return
	case *QueryParametersStatement:
		visitor.VisitQueryParametersStatement(node)
		return
	case *RestrictStatementWithClause:
		visitor.VisitRestrictStatementWithClause(node)
		return
	case *RestrictStatement:
		visitor.VisitRestrictStatement(node)
		return
	case *PatternStatement:
		visitor.VisitPatternStatement(node)
		return
	case *PatternDeclaration:
		visitor.VisitPatternDeclaration(node)
		return
	case *PatternPathParameter:
		visitor.VisitPatternPathParameter(node)
		return
	case *PatternMatch:
		visitor.VisitPatternMatch(node)
		return
	case *PatternPathValue:
		visitor.VisitPatternPathValue(node)
		return
	case *DataScopeExpression:
		visitor.VisitDataScopeExpression(node)
		return
	case *DataTableExpression:
		visitor.VisitDataTableExpression(node)
		return
	case *RowSchema:
		visitor.VisitRowSchema(node)
		return
	case *EvaluateRowSchema:
		visitor.VisitEvaluateRowSchema(node)
		return
	case *ExternalDataExpression:
		visitor.VisitExternalDataExpression(node)
		return
	case *ContextualDataTableExpression:
		visitor.VisitContextualDataTableExpression(node)
		return
	case *ExternalDataWithClause:
		visitor.VisitExternalDataWithClause(node)
		return
	case *InlineExternalTableKindClause:
		visitor.VisitInlineExternalTableKindClause(node)
		return
	case *InlineExternalTableDataFormatClause:
		visitor.VisitInlineExternalTableDataFormatClause(node)
		return
	case *DateTimePattern:
		visitor.VisitDateTimePattern(node)
		return
	case *InlineExternalTablePathFormatPartitionColumnReference:
		visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(node)
		return
	case *InlineExternalTablePathFormatClause:
		visitor.VisitInlineExternalTablePathFormatClause(node)
		return
	case *PartitionColumnDeclaration:
		visitor.VisitPartitionColumnDeclaration(node)
		return
	case *InlineExternalTablePartitionClause:
		visitor.VisitInlineExternalTablePartitionClause(node)
		return
	case *InlineExternalTableConnectionStringsClause:
		visitor.VisitInlineExternalTableConnectionStringsClause(node)
		return
	case *InlineExternalTableExpression:
		visitor.VisitInlineExternalTableExpression(node)
		return
	case *JoinOperator:
		visitor.VisitJoinOperator(node)
		return
	case *LookupOperator:
		visitor.VisitLookupOperator(node)
		return
	case *JoinOnClause:
		visitor.VisitJoinOnClause(node)
		return
	case *JoinWhereClause:
		visitor.VisitJoinWhereClause(node)
		return
	case *ScanOperator:
		visitor.VisitScanOperator(node)
		return
	case *AssertSchemaOperator:
		visitor.VisitAssertSchemaOperator(node)
		return
	case *ScanDeclareClause:
		visitor.VisitScanDeclareClause(node)
		return
	case *ScanOrderByClause:
		visitor.VisitScanOrderByClause(node)
		return
	case *ScanPartitionByClause:
		visitor.VisitScanPartitionByClause(node)
		return
	case *ScanStepOutput:
		visitor.VisitScanStepOutput(node)
		return
	case *ScanStep:
		visitor.VisitScanStep(node)
		return
	case *ScanComputationClause:
		visitor.VisitScanComputationClause(node)
		return
	case *ScanAssignment:
		visitor.VisitScanAssignment(node)
		return
	case *PartitionByOperator:
		visitor.VisitPartitionByOperator(node)
		return
	case *PartitionByIdClause:
		visitor.VisitPartitionByIdClause(node)
		return
	case *CommandWithValueClause:
		visitor.VisitCommandWithValueClause(node)
		return
	case *CommandWithPropertyListClause:
		visitor.VisitCommandWithPropertyListClause(node)
		return
	case *UnknownCommand:
		visitor.VisitUnknownCommand(node)
		return
	case *CustomCommand:
		visitor.VisitCustomCommand(node)
		return
	case *PartialCommand:
		visitor.VisitPartialCommand(node)
		return
	case *CommandAndSkippedTokens:
		visitor.VisitCommandAndSkippedTokens(node)
		return
	case *BadCommand:
		visitor.VisitBadCommand(node)
		return
	case *CommandBlock:
		visitor.VisitCommandBlock(node)
		return
	case *GraphMergeOperator:
		visitor.VisitGraphMergeOperator(node)
		return
	case *SyntaxNode:
		switch node.GeneratedNodeName() {
		case "QueryBlock":
			visitor.VisitQueryBlock(&QueryBlock{SyntaxNode: node})
			return
		case "DirectiveBlock":
			visitor.VisitDirectiveBlock(&DirectiveBlock{SyntaxNode: node})
			return
		case "Directive":
			visitor.VisitDirective(&Directive{SyntaxNode: node})
			return
		case "SkippedTokens":
			visitor.VisitSkippedTokens(&SkippedTokens{SyntaxNode: node})
			return
		case "TypeOfLiteralExpression":
			visitor.VisitTypeOfLiteralExpression(&TypeOfLiteralExpression{SyntaxNode: node})
			return
		case "BadQueryOperator":
			visitor.VisitBadQueryOperator(&BadQueryOperator{SyntaxNode: node})
			return
		case "CompoundStringLiteralExpression":
			visitor.VisitCompoundStringLiteralExpression(&CompoundStringLiteralExpression{SyntaxNode: node})
			return
		case "TokenName":
			visitor.VisitTokenName(&TokenName{SyntaxNode: node})
			return
		case "BracketedName":
			visitor.VisitBracketedName(&BracketedName{SyntaxNode: node})
			return
		case "BracedName":
			visitor.VisitBracedName(&BracedName{SyntaxNode: node})
			return
		case "WildcardedName":
			visitor.VisitWildcardedName(&WildcardedName{SyntaxNode: node})
			return
		case "BracketedWildcardedName":
			visitor.VisitBracketedWildcardedName(&BracketedWildcardedName{SyntaxNode: node})
			return
		case "NameDeclaration":
			visitor.VisitNameDeclaration(&NameDeclaration{SyntaxNode: node})
			return
		case "NameReference":
			visitor.VisitNameReference(&NameReference{SyntaxNode: node})
			return
		case "LiteralExpression":
			visitor.VisitLiteralExpression(&LiteralExpression{SyntaxNode: node})
			return
		case "StarExpression":
			visitor.VisitStarExpression(&StarExpression{SyntaxNode: node})
			return
		case "AtExpression":
			visitor.VisitAtExpression(&AtExpression{SyntaxNode: node})
			return
		case "JsonPair":
			visitor.VisitJsonPair(&JsonPair{SyntaxNode: node})
			return
		case "JsonObjectExpression":
			visitor.VisitJsonObjectExpression(&JsonObjectExpression{SyntaxNode: node})
			return
		case "JsonArrayExpression":
			visitor.VisitJsonArrayExpression(&JsonArrayExpression{SyntaxNode: node})
			return
		case "DynamicExpression":
			visitor.VisitDynamicExpression(&DynamicExpression{SyntaxNode: node})
			return
		case "ParenthesizedExpression":
			visitor.VisitParenthesizedExpression(&ParenthesizedExpression{SyntaxNode: node})
			return
		case "ExpressionList":
			visitor.VisitExpressionList(&ExpressionList{SyntaxNode: node})
			return
		case "ExpressionCouple":
			visitor.VisitExpressionCouple(&ExpressionCouple{SyntaxNode: node})
			return
		case "PrefixUnaryExpression":
			visitor.VisitPrefixUnaryExpression(&PrefixUnaryExpression{SyntaxNode: node})
			return
		case "BinaryExpression":
			visitor.VisitBinaryExpression(&BinaryExpression{SyntaxNode: node})
			return
		case "InExpression":
			visitor.VisitInExpression(&InExpression{SyntaxNode: node})
			return
		case "HasAnyExpression":
			visitor.VisitHasAnyExpression(&HasAnyExpression{SyntaxNode: node})
			return
		case "HasAllExpression":
			visitor.VisitHasAllExpression(&HasAllExpression{SyntaxNode: node})
			return
		case "BetweenExpression":
			visitor.VisitBetweenExpression(&BetweenExpression{SyntaxNode: node})
			return
		case "FunctionCallExpression":
			visitor.VisitFunctionCallExpression(&FunctionCallExpression{SyntaxNode: node})
			return
		case "ToScalarExpression":
			visitor.VisitToScalarExpression(&ToScalarExpression{SyntaxNode: node})
			return
		case "ToTableExpression":
			visitor.VisitToTableExpression(&ToTableExpression{SyntaxNode: node})
			return
		case "MaterializedViewCombineExpression":
			visitor.VisitMaterializedViewCombineExpression(&MaterializedViewCombineExpression{SyntaxNode: node})
			return
		case "MaterializedViewCombineNameClause":
			visitor.VisitMaterializedViewCombineNameClause(&MaterializedViewCombineNameClause{SyntaxNode: node})
			return
		case "MaterializedViewCombineClause":
			visitor.VisitMaterializedViewCombineClause(&MaterializedViewCombineClause{SyntaxNode: node})
			return
		case "SimpleNamedExpression":
			visitor.VisitSimpleNamedExpression(&SimpleNamedExpression{SyntaxNode: node})
			return
		case "RenameList":
			visitor.VisitRenameList(&RenameList{SyntaxNode: node})
			return
		case "CompoundNamedExpression":
			visitor.VisitCompoundNamedExpression(&CompoundNamedExpression{SyntaxNode: node})
			return
		case "BracketedExpression":
			visitor.VisitBracketedExpression(&BracketedExpression{SyntaxNode: node})
			return
		case "PathExpression":
			visitor.VisitPathExpression(&PathExpression{SyntaxNode: node})
			return
		case "ElementExpression":
			visitor.VisitElementExpression(&ElementExpression{SyntaxNode: node})
			return
		case "PipeExpression":
			visitor.VisitPipeExpression(&PipeExpression{SyntaxNode: node})
			return
		case "RangeOperator":
			visitor.VisitRangeOperator(&RangeOperator{SyntaxNode: node})
			return
		case "NamedParameter":
			visitor.VisitNamedParameter(&NamedParameter{SyntaxNode: node})
			return
		case "ConsumeOperator":
			visitor.VisitConsumeOperator(&ConsumeOperator{SyntaxNode: node})
			return
		case "CountOperator":
			visitor.VisitCountOperator(&CountOperator{SyntaxNode: node})
			return
		case "CountAsIdentifierClause":
			visitor.VisitCountAsIdentifierClause(&CountAsIdentifierClause{SyntaxNode: node})
			return
		case "ExecuteAndCacheOperator":
			visitor.VisitExecuteAndCacheOperator(&ExecuteAndCacheOperator{SyntaxNode: node})
			return
		case "ExtendOperator":
			visitor.VisitExtendOperator(&ExtendOperator{SyntaxNode: node})
			return
		case "FacetOperator":
			visitor.VisitFacetOperator(&FacetOperator{SyntaxNode: node})
			return
		case "FacetWithOperatorClause":
			visitor.VisitFacetWithOperatorClause(&FacetWithOperatorClause{SyntaxNode: node})
			return
		case "FacetWithExpressionClause":
			visitor.VisitFacetWithExpressionClause(&FacetWithExpressionClause{SyntaxNode: node})
			return
		case "FilterOperator":
			visitor.VisitFilterOperator(&FilterOperator{SyntaxNode: node})
			return
		case "GetSchemaOperator":
			visitor.VisitGetSchemaOperator(&GetSchemaOperator{SyntaxNode: node})
			return
		case "FindOperator":
			visitor.VisitFindOperator(&FindOperator{SyntaxNode: node})
			return
		case "DataScopeClause":
			visitor.VisitDataScopeClause(&DataScopeClause{SyntaxNode: node})
			return
		case "TypedColumnReference":
			visitor.VisitTypedColumnReference(&TypedColumnReference{SyntaxNode: node})
			return
		case "FindInClause":
			visitor.VisitFindInClause(&FindInClause{SyntaxNode: node})
			return
		case "FindProjectClause":
			visitor.VisitFindProjectClause(&FindProjectClause{SyntaxNode: node})
			return
		case "PackExpression":
			visitor.VisitPackExpression(&PackExpression{SyntaxNode: node})
			return
		case "NameAndTypeDeclaration":
			visitor.VisitNameAndTypeDeclaration(&NameAndTypeDeclaration{SyntaxNode: node})
			return
		case "PrimitiveTypeExpression":
			visitor.VisitPrimitiveTypeExpression(&PrimitiveTypeExpression{SyntaxNode: node})
			return
		case "SearchOperator":
			visitor.VisitSearchOperator(&SearchOperator{SyntaxNode: node})
			return
		case "ForkOperator":
			visitor.VisitForkOperator(&ForkOperator{SyntaxNode: node})
			return
		case "NameEqualsClause":
			visitor.VisitNameEqualsClause(&NameEqualsClause{SyntaxNode: node})
			return
		case "ForkExpression":
			visitor.VisitForkExpression(&ForkExpression{SyntaxNode: node})
			return
		case "MakeSeriesOperator":
			visitor.VisitMakeSeriesOperator(&MakeSeriesOperator{SyntaxNode: node})
			return
		case "MakeSeriesExpression":
			visitor.VisitMakeSeriesExpression(&MakeSeriesExpression{SyntaxNode: node})
			return
		case "DefaultExpressionClause":
			visitor.VisitDefaultExpressionClause(&DefaultExpressionClause{SyntaxNode: node})
			return
		case "MakeSeriesOnClause":
			visitor.VisitMakeSeriesOnClause(&MakeSeriesOnClause{SyntaxNode: node})
			return
		case "MakeSeriesFromClause":
			visitor.VisitMakeSeriesFromClause(&MakeSeriesFromClause{SyntaxNode: node})
			return
		case "MakeSeriesToClause":
			visitor.VisitMakeSeriesToClause(&MakeSeriesToClause{SyntaxNode: node})
			return
		case "MakeSeriesStepClause":
			visitor.VisitMakeSeriesStepClause(&MakeSeriesStepClause{SyntaxNode: node})
			return
		case "MakeSeriesInRangeClause":
			visitor.VisitMakeSeriesInRangeClause(&MakeSeriesInRangeClause{SyntaxNode: node})
			return
		case "MakeSeriesFromToStepClause":
			visitor.VisitMakeSeriesFromToStepClause(&MakeSeriesFromToStepClause{SyntaxNode: node})
			return
		case "MakeSeriesByClause":
			visitor.VisitMakeSeriesByClause(&MakeSeriesByClause{SyntaxNode: node})
			return
		case "MvExpandOperator":
			visitor.VisitMvExpandOperator(&MvExpandOperator{SyntaxNode: node})
			return
		case "MvExpandExpression":
			visitor.VisitMvExpandExpression(&MvExpandExpression{SyntaxNode: node})
			return
		case "MvExpandRowLimitClause":
			visitor.VisitMvExpandRowLimitClause(&MvExpandRowLimitClause{SyntaxNode: node})
			return
		case "MvApplyOperator":
			visitor.VisitMvApplyOperator(&MvApplyOperator{SyntaxNode: node})
			return
		case "MvApplyExpression":
			visitor.VisitMvApplyExpression(&MvApplyExpression{SyntaxNode: node})
			return
		case "MvApplyRowLimitClause":
			visitor.VisitMvApplyRowLimitClause(&MvApplyRowLimitClause{SyntaxNode: node})
			return
		case "MvApplyContextIdClause":
			visitor.VisitMvApplyContextIdClause(&MvApplyContextIdClause{SyntaxNode: node})
			return
		case "MvApplySubqueryExpression":
			visitor.VisitMvApplySubqueryExpression(&MvApplySubqueryExpression{SyntaxNode: node})
			return
		case "ToTypeOfClause":
			visitor.VisitToTypeOfClause(&ToTypeOfClause{SyntaxNode: node})
			return
		case "EvaluateSchemaClause":
			visitor.VisitEvaluateSchemaClause(&EvaluateSchemaClause{SyntaxNode: node})
			return
		case "EvaluateOperator":
			visitor.VisitEvaluateOperator(&EvaluateOperator{SyntaxNode: node})
			return
		case "ParseOperator":
			visitor.VisitParseOperator(&ParseOperator{SyntaxNode: node})
			return
		case "ParseWhereOperator":
			visitor.VisitParseWhereOperator(&ParseWhereOperator{SyntaxNode: node})
			return
		case "ParseKvWithClause":
			visitor.VisitParseKvWithClause(&ParseKvWithClause{SyntaxNode: node})
			return
		case "ParseKvOperator":
			visitor.VisitParseKvOperator(&ParseKvOperator{SyntaxNode: node})
			return
		case "PartitionOperator":
			visitor.VisitPartitionOperator(&PartitionOperator{SyntaxNode: node})
			return
		case "PartitionQuery":
			visitor.VisitPartitionQuery(&PartitionQuery{SyntaxNode: node})
			return
		case "PartitionScope":
			visitor.VisitPartitionScope(&PartitionScope{SyntaxNode: node})
			return
		case "PartitionSubquery":
			visitor.VisitPartitionSubquery(&PartitionSubquery{SyntaxNode: node})
			return
		case "ProjectOperator":
			visitor.VisitProjectOperator(&ProjectOperator{SyntaxNode: node})
			return
		case "ProjectAwayOperator":
			visitor.VisitProjectAwayOperator(&ProjectAwayOperator{SyntaxNode: node})
			return
		case "ProjectByNamesOperator":
			visitor.VisitProjectByNamesOperator(&ProjectByNamesOperator{SyntaxNode: node})
			return
		case "ProjectKeepOperator":
			visitor.VisitProjectKeepOperator(&ProjectKeepOperator{SyntaxNode: node})
			return
		case "ProjectRenameOperator":
			visitor.VisitProjectRenameOperator(&ProjectRenameOperator{SyntaxNode: node})
			return
		case "ProjectReorderOperator":
			visitor.VisitProjectReorderOperator(&ProjectReorderOperator{SyntaxNode: node})
			return
		case "SampleOperator":
			visitor.VisitSampleOperator(&SampleOperator{SyntaxNode: node})
			return
		case "SampleDistinctOperator":
			visitor.VisitSampleDistinctOperator(&SampleDistinctOperator{SyntaxNode: node})
			return
		case "EntityGroup":
			visitor.VisitEntityGroup(&EntityGroup{SyntaxNode: node})
			return
		case "ReduceByOperator":
			visitor.VisitReduceByOperator(&ReduceByOperator{SyntaxNode: node})
			return
		case "ReduceByWithClause":
			visitor.VisitReduceByWithClause(&ReduceByWithClause{SyntaxNode: node})
			return
		case "SummarizeOperator":
			visitor.VisitSummarizeOperator(&SummarizeOperator{SyntaxNode: node})
			return
		case "MacroExpandScopeReferenceName":
			visitor.VisitMacroExpandScopeReferenceName(&MacroExpandScopeReferenceName{SyntaxNode: node})
			return
		case "MacroExpandOperator":
			visitor.VisitMacroExpandOperator(&MacroExpandOperator{SyntaxNode: node})
			return
		case "SummarizeByClause":
			visitor.VisitSummarizeByClause(&SummarizeByClause{SyntaxNode: node})
			return
		case "DistinctOperator":
			visitor.VisitDistinctOperator(&DistinctOperator{SyntaxNode: node})
			return
		case "TakeOperator":
			visitor.VisitTakeOperator(&TakeOperator{SyntaxNode: node})
			return
		case "SortOperator":
			visitor.VisitSortOperator(&SortOperator{SyntaxNode: node})
			return
		case "OrderedExpression":
			visitor.VisitOrderedExpression(&OrderedExpression{SyntaxNode: node})
			return
		case "OrderingClause":
			visitor.VisitOrderingClause(&OrderingClause{SyntaxNode: node})
			return
		case "OrderingNullsClause":
			visitor.VisitOrderingNullsClause(&OrderingNullsClause{SyntaxNode: node})
			return
		case "TopHittersOperator":
			visitor.VisitTopHittersOperator(&TopHittersOperator{SyntaxNode: node})
			return
		case "TopHittersByClause":
			visitor.VisitTopHittersByClause(&TopHittersByClause{SyntaxNode: node})
			return
		case "TopOperator":
			visitor.VisitTopOperator(&TopOperator{SyntaxNode: node})
			return
		case "TopNestedOperator":
			visitor.VisitTopNestedOperator(&TopNestedOperator{SyntaxNode: node})
			return
		case "TopNestedClause":
			visitor.VisitTopNestedClause(&TopNestedClause{SyntaxNode: node})
			return
		case "TopNestedWithOthersClause":
			visitor.VisitTopNestedWithOthersClause(&TopNestedWithOthersClause{SyntaxNode: node})
			return
		case "UnionOperator":
			visitor.VisitUnionOperator(&UnionOperator{SyntaxNode: node})
			return
		case "AsOperator":
			visitor.VisitAsOperator(&AsOperator{SyntaxNode: node})
			return
		case "SerializeOperator":
			visitor.VisitSerializeOperator(&SerializeOperator{SyntaxNode: node})
			return
		case "InvokeOperator":
			visitor.VisitInvokeOperator(&InvokeOperator{SyntaxNode: node})
			return
		case "RenderOperator":
			visitor.VisitRenderOperator(&RenderOperator{SyntaxNode: node})
			return
		case "MakeGraphOperator":
			visitor.VisitMakeGraphOperator(&MakeGraphOperator{SyntaxNode: node})
			return
		case "MakeGraphWithTablesAndKeysClause":
			visitor.VisitMakeGraphWithTablesAndKeysClause(&MakeGraphWithTablesAndKeysClause{SyntaxNode: node})
			return
		case "MakeGraphWithImplicitIdClause":
			visitor.VisitMakeGraphWithImplicitIdClause(&MakeGraphWithImplicitIdClause{SyntaxNode: node})
			return
		case "GraphMarkComponentsOperator":
			visitor.VisitGraphMarkComponentsOperator(&GraphMarkComponentsOperator{SyntaxNode: node})
			return
		case "GraphWhereNodesOperator":
			visitor.VisitGraphWhereNodesOperator(&GraphWhereNodesOperator{SyntaxNode: node})
			return
		case "GraphWhereEdgesOperator":
			visitor.VisitGraphWhereEdgesOperator(&GraphWhereEdgesOperator{SyntaxNode: node})
			return
		case "MakeGraphTableAndKeyClause":
			visitor.VisitMakeGraphTableAndKeyClause(&MakeGraphTableAndKeyClause{SyntaxNode: node})
			return
		case "MakeGraphPartitionedByClause":
			visitor.VisitMakeGraphPartitionedByClause(&MakeGraphPartitionedByClause{SyntaxNode: node})
			return
		case "GraphToTableOperator":
			visitor.VisitGraphToTableOperator(&GraphToTableOperator{SyntaxNode: node})
			return
		case "GraphToTableOutputClause":
			visitor.VisitGraphToTableOutputClause(&GraphToTableOutputClause{SyntaxNode: node})
			return
		case "GraphToTableAsClause":
			visitor.VisitGraphToTableAsClause(&GraphToTableAsClause{SyntaxNode: node})
			return
		case "GraphMatchOperator":
			visitor.VisitGraphMatchOperator(&GraphMatchOperator{SyntaxNode: node})
			return
		case "GraphShortestPathsOperator":
			visitor.VisitGraphShortestPathsOperator(&GraphShortestPathsOperator{SyntaxNode: node})
			return
		case "GraphMatchPattern":
			visitor.VisitGraphMatchPattern(&GraphMatchPattern{SyntaxNode: node})
			return
		case "GraphMatchPatternNode":
			visitor.VisitGraphMatchPatternNode(&GraphMatchPatternNode{SyntaxNode: node})
			return
		case "GraphMatchPatternEdge":
			visitor.VisitGraphMatchPatternEdge(&GraphMatchPatternEdge{SyntaxNode: node})
			return
		case "GraphMatchPatternEdgeRange":
			visitor.VisitGraphMatchPatternEdgeRange(&GraphMatchPatternEdgeRange{SyntaxNode: node})
			return
		case "WhereClause":
			visitor.VisitWhereClause(&WhereClause{SyntaxNode: node})
			return
		case "ProjectClause":
			visitor.VisitProjectClause(&ProjectClause{SyntaxNode: node})
			return
		case "NameReferenceList":
			visitor.VisitNameReferenceList(&NameReferenceList{SyntaxNode: node})
			return
		case "RenderWithClause":
			visitor.VisitRenderWithClause(&RenderWithClause{SyntaxNode: node})
			return
		case "PrintOperator":
			visitor.VisitPrintOperator(&PrintOperator{SyntaxNode: node})
			return
		case "AliasStatement":
			visitor.VisitAliasStatement(&AliasStatement{SyntaxNode: node})
			return
		case "LetStatement":
			visitor.VisitLetStatement(&LetStatement{SyntaxNode: node})
			return
		case "FunctionDeclaration":
			visitor.VisitFunctionDeclaration(&FunctionDeclaration{SyntaxNode: node})
			return
		case "FunctionParameters":
			visitor.VisitFunctionParameters(&FunctionParameters{SyntaxNode: node})
			return
		case "FunctionParameter":
			visitor.VisitFunctionParameter(&FunctionParameter{SyntaxNode: node})
			return
		case "DefaultValueDeclaration":
			visitor.VisitDefaultValueDeclaration(&DefaultValueDeclaration{SyntaxNode: node})
			return
		case "FunctionBody":
			visitor.VisitFunctionBody(&FunctionBody{SyntaxNode: node})
			return
		case "SchemaTypeExpression":
			visitor.VisitSchemaTypeExpression(&SchemaTypeExpression{SyntaxNode: node})
			return
		case "ExpressionStatement":
			visitor.VisitExpressionStatement(&ExpressionStatement{SyntaxNode: node})
			return
		case "MaterializeExpression":
			visitor.VisitMaterializeExpression(&MaterializeExpression{SyntaxNode: node})
			return
		case "SetOptionStatement":
			visitor.VisitSetOptionStatement(&SetOptionStatement{SyntaxNode: node})
			return
		case "OptionValueClause":
			visitor.VisitOptionValueClause(&OptionValueClause{SyntaxNode: node})
			return
		case "QueryParametersStatement":
			visitor.VisitQueryParametersStatement(&QueryParametersStatement{SyntaxNode: node})
			return
		case "RestrictStatementWithClause":
			visitor.VisitRestrictStatementWithClause(&RestrictStatementWithClause{SyntaxNode: node})
			return
		case "RestrictStatement":
			visitor.VisitRestrictStatement(&RestrictStatement{SyntaxNode: node})
			return
		case "PatternStatement":
			visitor.VisitPatternStatement(&PatternStatement{SyntaxNode: node})
			return
		case "PatternDeclaration":
			visitor.VisitPatternDeclaration(&PatternDeclaration{SyntaxNode: node})
			return
		case "PatternPathParameter":
			visitor.VisitPatternPathParameter(&PatternPathParameter{SyntaxNode: node})
			return
		case "PatternMatch":
			visitor.VisitPatternMatch(&PatternMatch{SyntaxNode: node})
			return
		case "PatternPathValue":
			visitor.VisitPatternPathValue(&PatternPathValue{SyntaxNode: node})
			return
		case "DataScopeExpression":
			visitor.VisitDataScopeExpression(&DataScopeExpression{SyntaxNode: node})
			return
		case "DataTableExpression":
			visitor.VisitDataTableExpression(&DataTableExpression{SyntaxNode: node})
			return
		case "RowSchema":
			visitor.VisitRowSchema(&RowSchema{SyntaxNode: node})
			return
		case "EvaluateRowSchema":
			visitor.VisitEvaluateRowSchema(&EvaluateRowSchema{SyntaxNode: node})
			return
		case "ExternalDataExpression":
			visitor.VisitExternalDataExpression(&ExternalDataExpression{SyntaxNode: node})
			return
		case "ContextualDataTableExpression":
			visitor.VisitContextualDataTableExpression(&ContextualDataTableExpression{SyntaxNode: node})
			return
		case "ExternalDataWithClause":
			visitor.VisitExternalDataWithClause(&ExternalDataWithClause{SyntaxNode: node})
			return
		case "InlineExternalTableKindClause":
			visitor.VisitInlineExternalTableKindClause(&InlineExternalTableKindClause{SyntaxNode: node})
			return
		case "InlineExternalTableDataFormatClause":
			visitor.VisitInlineExternalTableDataFormatClause(&InlineExternalTableDataFormatClause{SyntaxNode: node})
			return
		case "DateTimePattern":
			visitor.VisitDateTimePattern(&DateTimePattern{SyntaxNode: node})
			return
		case "InlineExternalTablePathFormatPartitionColumnReference":
			visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(&InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node})
			return
		case "InlineExternalTablePathFormatClause":
			visitor.VisitInlineExternalTablePathFormatClause(&InlineExternalTablePathFormatClause{SyntaxNode: node})
			return
		case "PartitionColumnDeclaration":
			visitor.VisitPartitionColumnDeclaration(&PartitionColumnDeclaration{SyntaxNode: node})
			return
		case "InlineExternalTablePartitionClause":
			visitor.VisitInlineExternalTablePartitionClause(&InlineExternalTablePartitionClause{SyntaxNode: node})
			return
		case "InlineExternalTableConnectionStringsClause":
			visitor.VisitInlineExternalTableConnectionStringsClause(&InlineExternalTableConnectionStringsClause{SyntaxNode: node})
			return
		case "InlineExternalTableExpression":
			visitor.VisitInlineExternalTableExpression(&InlineExternalTableExpression{SyntaxNode: node})
			return
		case "JoinOperator":
			visitor.VisitJoinOperator(&JoinOperator{SyntaxNode: node})
			return
		case "LookupOperator":
			visitor.VisitLookupOperator(&LookupOperator{SyntaxNode: node})
			return
		case "JoinOnClause":
			visitor.VisitJoinOnClause(&JoinOnClause{SyntaxNode: node})
			return
		case "JoinWhereClause":
			visitor.VisitJoinWhereClause(&JoinWhereClause{SyntaxNode: node})
			return
		case "ScanOperator":
			visitor.VisitScanOperator(&ScanOperator{SyntaxNode: node})
			return
		case "AssertSchemaOperator":
			visitor.VisitAssertSchemaOperator(&AssertSchemaOperator{SyntaxNode: node})
			return
		case "ScanDeclareClause":
			visitor.VisitScanDeclareClause(&ScanDeclareClause{SyntaxNode: node})
			return
		case "ScanOrderByClause":
			visitor.VisitScanOrderByClause(&ScanOrderByClause{SyntaxNode: node})
			return
		case "ScanPartitionByClause":
			visitor.VisitScanPartitionByClause(&ScanPartitionByClause{SyntaxNode: node})
			return
		case "ScanStepOutput":
			visitor.VisitScanStepOutput(&ScanStepOutput{SyntaxNode: node})
			return
		case "ScanStep":
			visitor.VisitScanStep(&ScanStep{SyntaxNode: node})
			return
		case "ScanComputationClause":
			visitor.VisitScanComputationClause(&ScanComputationClause{SyntaxNode: node})
			return
		case "ScanAssignment":
			visitor.VisitScanAssignment(&ScanAssignment{SyntaxNode: node})
			return
		case "PartitionByOperator":
			visitor.VisitPartitionByOperator(&PartitionByOperator{SyntaxNode: node})
			return
		case "PartitionByIdClause":
			visitor.VisitPartitionByIdClause(&PartitionByIdClause{SyntaxNode: node})
			return
		case "CommandWithValueClause":
			visitor.VisitCommandWithValueClause(&CommandWithValueClause{SyntaxNode: node})
			return
		case "CommandWithPropertyListClause":
			visitor.VisitCommandWithPropertyListClause(&CommandWithPropertyListClause{SyntaxNode: node})
			return
		case "UnknownCommand":
			visitor.VisitUnknownCommand(&UnknownCommand{SyntaxNode: node})
			return
		case "CustomCommand":
			visitor.VisitCustomCommand(&CustomCommand{SyntaxNode: node})
			return
		case "PartialCommand":
			visitor.VisitPartialCommand(&PartialCommand{SyntaxNode: node})
			return
		case "CommandAndSkippedTokens":
			visitor.VisitCommandAndSkippedTokens(&CommandAndSkippedTokens{SyntaxNode: node})
			return
		case "BadCommand":
			visitor.VisitBadCommand(&BadCommand{SyntaxNode: node})
			return
		case "CommandBlock":
			visitor.VisitCommandBlock(&CommandBlock{SyntaxNode: node})
			return
		case "GraphMergeOperator":
			visitor.VisitGraphMergeOperator(&GraphMergeOperator{SyntaxNode: node})
			return
		}
		switch node.Kind() {
		case SyntaxKindQueryBlock:
			visitor.VisitQueryBlock(&QueryBlock{SyntaxNode: node})
			return
		case SyntaxKindDirectiveBlock:
			visitor.VisitDirectiveBlock(&DirectiveBlock{SyntaxNode: node})
			return
		case SyntaxKindDirective:
			visitor.VisitDirective(&Directive{SyntaxNode: node})
			return
		case SyntaxKindSkippedTokens:
			visitor.VisitSkippedTokens(&SkippedTokens{SyntaxNode: node})
			return
		case SyntaxKindTypeOfLiteralExpression:
			visitor.VisitTypeOfLiteralExpression(&TypeOfLiteralExpression{SyntaxNode: node})
			return
		case SyntaxKindBadQueryOperator:
			visitor.VisitBadQueryOperator(&BadQueryOperator{SyntaxNode: node})
			return
		case SyntaxKindCompoundStringLiteralExpression:
			visitor.VisitCompoundStringLiteralExpression(&CompoundStringLiteralExpression{SyntaxNode: node})
			return
		case SyntaxKindTokenName:
			visitor.VisitTokenName(&TokenName{SyntaxNode: node})
			return
		case SyntaxKindBracketedName:
			visitor.VisitBracketedName(&BracketedName{SyntaxNode: node})
			return
		case SyntaxKindBracedName:
			visitor.VisitBracedName(&BracedName{SyntaxNode: node})
			return
		case SyntaxKindWildcardedName:
			visitor.VisitWildcardedName(&WildcardedName{SyntaxNode: node})
			return
		case SyntaxKindBracketedWildcardedName:
			visitor.VisitBracketedWildcardedName(&BracketedWildcardedName{SyntaxNode: node})
			return
		case SyntaxKindNameDeclaration:
			visitor.VisitNameDeclaration(&NameDeclaration{SyntaxNode: node})
			return
		case SyntaxKindNameReference:
			visitor.VisitNameReference(&NameReference{SyntaxNode: node})
			return
		case SyntaxKindStarExpression:
			visitor.VisitStarExpression(&StarExpression{SyntaxNode: node})
			return
		case SyntaxKindAtExpression:
			visitor.VisitAtExpression(&AtExpression{SyntaxNode: node})
			return
		case SyntaxKindJsonPair:
			visitor.VisitJsonPair(&JsonPair{SyntaxNode: node})
			return
		case SyntaxKindJsonObjectExpression:
			visitor.VisitJsonObjectExpression(&JsonObjectExpression{SyntaxNode: node})
			return
		case SyntaxKindJsonArrayExpression:
			visitor.VisitJsonArrayExpression(&JsonArrayExpression{SyntaxNode: node})
			return
		case SyntaxKindDynamicExpression:
			visitor.VisitDynamicExpression(&DynamicExpression{SyntaxNode: node})
			return
		case SyntaxKindParenthesizedExpression:
			visitor.VisitParenthesizedExpression(&ParenthesizedExpression{SyntaxNode: node})
			return
		case SyntaxKindExpressionList:
			visitor.VisitExpressionList(&ExpressionList{SyntaxNode: node})
			return
		case SyntaxKindExpressionCouple:
			visitor.VisitExpressionCouple(&ExpressionCouple{SyntaxNode: node})
			return
		case SyntaxKindFunctionCallExpression:
			visitor.VisitFunctionCallExpression(&FunctionCallExpression{SyntaxNode: node})
			return
		case SyntaxKindToScalarExpression:
			visitor.VisitToScalarExpression(&ToScalarExpression{SyntaxNode: node})
			return
		case SyntaxKindToTableExpression:
			visitor.VisitToTableExpression(&ToTableExpression{SyntaxNode: node})
			return
		case SyntaxKindMaterializedViewCombineExpression:
			visitor.VisitMaterializedViewCombineExpression(&MaterializedViewCombineExpression{SyntaxNode: node})
			return
		case SyntaxKindMaterializedViewCombineClause:
			visitor.VisitMaterializedViewCombineClause(&MaterializedViewCombineClause{SyntaxNode: node})
			return
		case SyntaxKindSimpleNamedExpression:
			visitor.VisitSimpleNamedExpression(&SimpleNamedExpression{SyntaxNode: node})
			return
		case SyntaxKindRenameList:
			visitor.VisitRenameList(&RenameList{SyntaxNode: node})
			return
		case SyntaxKindCompoundNamedExpression:
			visitor.VisitCompoundNamedExpression(&CompoundNamedExpression{SyntaxNode: node})
			return
		case SyntaxKindBracketedExpression:
			visitor.VisitBracketedExpression(&BracketedExpression{SyntaxNode: node})
			return
		case SyntaxKindPathExpression:
			visitor.VisitPathExpression(&PathExpression{SyntaxNode: node})
			return
		case SyntaxKindElementExpression:
			visitor.VisitElementExpression(&ElementExpression{SyntaxNode: node})
			return
		case SyntaxKindPipeExpression:
			visitor.VisitPipeExpression(&PipeExpression{SyntaxNode: node})
			return
		case SyntaxKindRangeOperator:
			visitor.VisitRangeOperator(&RangeOperator{SyntaxNode: node})
			return
		case SyntaxKindNamedParameter:
			visitor.VisitNamedParameter(&NamedParameter{SyntaxNode: node})
			return
		case SyntaxKindConsumeOperator:
			visitor.VisitConsumeOperator(&ConsumeOperator{SyntaxNode: node})
			return
		case SyntaxKindCountOperator:
			visitor.VisitCountOperator(&CountOperator{SyntaxNode: node})
			return
		case SyntaxKindCountAsIdentifierClause:
			visitor.VisitCountAsIdentifierClause(&CountAsIdentifierClause{SyntaxNode: node})
			return
		case SyntaxKindExecuteAndCacheOperator:
			visitor.VisitExecuteAndCacheOperator(&ExecuteAndCacheOperator{SyntaxNode: node})
			return
		case SyntaxKindExtendOperator:
			visitor.VisitExtendOperator(&ExtendOperator{SyntaxNode: node})
			return
		case SyntaxKindFacetOperator:
			visitor.VisitFacetOperator(&FacetOperator{SyntaxNode: node})
			return
		case SyntaxKindFacetWithOperatorClause:
			visitor.VisitFacetWithOperatorClause(&FacetWithOperatorClause{SyntaxNode: node})
			return
		case SyntaxKindFacetWithExpressionClause:
			visitor.VisitFacetWithExpressionClause(&FacetWithExpressionClause{SyntaxNode: node})
			return
		case SyntaxKindFilterOperator:
			visitor.VisitFilterOperator(&FilterOperator{SyntaxNode: node})
			return
		case SyntaxKindGetSchemaOperator:
			visitor.VisitGetSchemaOperator(&GetSchemaOperator{SyntaxNode: node})
			return
		case SyntaxKindFindOperator:
			visitor.VisitFindOperator(&FindOperator{SyntaxNode: node})
			return
		case SyntaxKindDataScopeClause:
			visitor.VisitDataScopeClause(&DataScopeClause{SyntaxNode: node})
			return
		case SyntaxKindTypedColumnReference:
			visitor.VisitTypedColumnReference(&TypedColumnReference{SyntaxNode: node})
			return
		case SyntaxKindFindInClause:
			visitor.VisitFindInClause(&FindInClause{SyntaxNode: node})
			return
		case SyntaxKindFindProjectClause:
			visitor.VisitFindProjectClause(&FindProjectClause{SyntaxNode: node})
			return
		case SyntaxKindPackExpression:
			visitor.VisitPackExpression(&PackExpression{SyntaxNode: node})
			return
		case SyntaxKindNameAndTypeDeclaration:
			visitor.VisitNameAndTypeDeclaration(&NameAndTypeDeclaration{SyntaxNode: node})
			return
		case SyntaxKindPrimitiveTypeExpression:
			visitor.VisitPrimitiveTypeExpression(&PrimitiveTypeExpression{SyntaxNode: node})
			return
		case SyntaxKindSearchOperator:
			visitor.VisitSearchOperator(&SearchOperator{SyntaxNode: node})
			return
		case SyntaxKindForkOperator:
			visitor.VisitForkOperator(&ForkOperator{SyntaxNode: node})
			return
		case SyntaxKindNameEqualsClause:
			visitor.VisitNameEqualsClause(&NameEqualsClause{SyntaxNode: node})
			return
		case SyntaxKindForkExpression:
			visitor.VisitForkExpression(&ForkExpression{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesOperator:
			visitor.VisitMakeSeriesOperator(&MakeSeriesOperator{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesExpression:
			visitor.VisitMakeSeriesExpression(&MakeSeriesExpression{SyntaxNode: node})
			return
		case SyntaxKindDefaultExpressionClause:
			visitor.VisitDefaultExpressionClause(&DefaultExpressionClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesOnClause:
			visitor.VisitMakeSeriesOnClause(&MakeSeriesOnClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesFromClause:
			visitor.VisitMakeSeriesFromClause(&MakeSeriesFromClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesToClause:
			visitor.VisitMakeSeriesToClause(&MakeSeriesToClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesStepClause:
			visitor.VisitMakeSeriesStepClause(&MakeSeriesStepClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesInRangeClause:
			visitor.VisitMakeSeriesInRangeClause(&MakeSeriesInRangeClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesFromToStepClause:
			visitor.VisitMakeSeriesFromToStepClause(&MakeSeriesFromToStepClause{SyntaxNode: node})
			return
		case SyntaxKindMakeSeriesByClause:
			visitor.VisitMakeSeriesByClause(&MakeSeriesByClause{SyntaxNode: node})
			return
		case SyntaxKindMvExpandOperator:
			visitor.VisitMvExpandOperator(&MvExpandOperator{SyntaxNode: node})
			return
		case SyntaxKindMvExpandExpression:
			visitor.VisitMvExpandExpression(&MvExpandExpression{SyntaxNode: node})
			return
		case SyntaxKindMvExpandRowLimitClause:
			visitor.VisitMvExpandRowLimitClause(&MvExpandRowLimitClause{SyntaxNode: node})
			return
		case SyntaxKindMvApplyOperator:
			visitor.VisitMvApplyOperator(&MvApplyOperator{SyntaxNode: node})
			return
		case SyntaxKindMvApplyExpression:
			visitor.VisitMvApplyExpression(&MvApplyExpression{SyntaxNode: node})
			return
		case SyntaxKindMvApplyRowLimitClause:
			visitor.VisitMvApplyRowLimitClause(&MvApplyRowLimitClause{SyntaxNode: node})
			return
		case SyntaxKindMvApplyContextIdClause:
			visitor.VisitMvApplyContextIdClause(&MvApplyContextIdClause{SyntaxNode: node})
			return
		case SyntaxKindMvApplySubqueryExpression:
			visitor.VisitMvApplySubqueryExpression(&MvApplySubqueryExpression{SyntaxNode: node})
			return
		case SyntaxKindToTypeOfClause:
			visitor.VisitToTypeOfClause(&ToTypeOfClause{SyntaxNode: node})
			return
		case SyntaxKindEvaluateSchemaClause:
			visitor.VisitEvaluateSchemaClause(&EvaluateSchemaClause{SyntaxNode: node})
			return
		case SyntaxKindEvaluateOperator:
			visitor.VisitEvaluateOperator(&EvaluateOperator{SyntaxNode: node})
			return
		case SyntaxKindParseOperator:
			visitor.VisitParseOperator(&ParseOperator{SyntaxNode: node})
			return
		case SyntaxKindParseWhereOperator:
			visitor.VisitParseWhereOperator(&ParseWhereOperator{SyntaxNode: node})
			return
		case SyntaxKindParseKvWithClause:
			visitor.VisitParseKvWithClause(&ParseKvWithClause{SyntaxNode: node})
			return
		case SyntaxKindParseKvOperator:
			visitor.VisitParseKvOperator(&ParseKvOperator{SyntaxNode: node})
			return
		case SyntaxKindPartitionOperator:
			visitor.VisitPartitionOperator(&PartitionOperator{SyntaxNode: node})
			return
		case SyntaxKindPartitionQuery:
			visitor.VisitPartitionQuery(&PartitionQuery{SyntaxNode: node})
			return
		case SyntaxKindPartitionScope:
			visitor.VisitPartitionScope(&PartitionScope{SyntaxNode: node})
			return
		case SyntaxKindPartitionSubquery:
			visitor.VisitPartitionSubquery(&PartitionSubquery{SyntaxNode: node})
			return
		case SyntaxKindProjectOperator:
			visitor.VisitProjectOperator(&ProjectOperator{SyntaxNode: node})
			return
		case SyntaxKindProjectAwayOperator:
			visitor.VisitProjectAwayOperator(&ProjectAwayOperator{SyntaxNode: node})
			return
		case SyntaxKindProjectByNamesOperator:
			visitor.VisitProjectByNamesOperator(&ProjectByNamesOperator{SyntaxNode: node})
			return
		case SyntaxKindProjectKeepOperator:
			visitor.VisitProjectKeepOperator(&ProjectKeepOperator{SyntaxNode: node})
			return
		case SyntaxKindProjectRenameOperator:
			visitor.VisitProjectRenameOperator(&ProjectRenameOperator{SyntaxNode: node})
			return
		case SyntaxKindProjectReorderOperator:
			visitor.VisitProjectReorderOperator(&ProjectReorderOperator{SyntaxNode: node})
			return
		case SyntaxKindSampleOperator:
			visitor.VisitSampleOperator(&SampleOperator{SyntaxNode: node})
			return
		case SyntaxKindSampleDistinctOperator:
			visitor.VisitSampleDistinctOperator(&SampleDistinctOperator{SyntaxNode: node})
			return
		case SyntaxKindEntityGroupExpression:
			visitor.VisitEntityGroup(&EntityGroup{SyntaxNode: node})
			return
		case SyntaxKindReduceByOperator:
			visitor.VisitReduceByOperator(&ReduceByOperator{SyntaxNode: node})
			return
		case SyntaxKindReduceByWithClause:
			visitor.VisitReduceByWithClause(&ReduceByWithClause{SyntaxNode: node})
			return
		case SyntaxKindSummarizeOperator:
			visitor.VisitSummarizeOperator(&SummarizeOperator{SyntaxNode: node})
			return
		case SyntaxKindMacroExpandScopeReferenceName:
			visitor.VisitMacroExpandScopeReferenceName(&MacroExpandScopeReferenceName{SyntaxNode: node})
			return
		case SyntaxKindMacroExpandOperator:
			visitor.VisitMacroExpandOperator(&MacroExpandOperator{SyntaxNode: node})
			return
		case SyntaxKindSummarizeByClause:
			visitor.VisitSummarizeByClause(&SummarizeByClause{SyntaxNode: node})
			return
		case SyntaxKindDistinctOperator:
			visitor.VisitDistinctOperator(&DistinctOperator{SyntaxNode: node})
			return
		case SyntaxKindTakeOperator:
			visitor.VisitTakeOperator(&TakeOperator{SyntaxNode: node})
			return
		case SyntaxKindSortOperator:
			visitor.VisitSortOperator(&SortOperator{SyntaxNode: node})
			return
		case SyntaxKindOrderedExpression:
			visitor.VisitOrderedExpression(&OrderedExpression{SyntaxNode: node})
			return
		case SyntaxKindOrderingClause:
			visitor.VisitOrderingClause(&OrderingClause{SyntaxNode: node})
			return
		case SyntaxKindOrderingNullsClause:
			visitor.VisitOrderingNullsClause(&OrderingNullsClause{SyntaxNode: node})
			return
		case SyntaxKindTopHittersOperator:
			visitor.VisitTopHittersOperator(&TopHittersOperator{SyntaxNode: node})
			return
		case SyntaxKindTopHittersByClause:
			visitor.VisitTopHittersByClause(&TopHittersByClause{SyntaxNode: node})
			return
		case SyntaxKindTopOperator:
			visitor.VisitTopOperator(&TopOperator{SyntaxNode: node})
			return
		case SyntaxKindTopNestedOperator:
			visitor.VisitTopNestedOperator(&TopNestedOperator{SyntaxNode: node})
			return
		case SyntaxKindTopNestedClause:
			visitor.VisitTopNestedClause(&TopNestedClause{SyntaxNode: node})
			return
		case SyntaxKindTopNestedWithOthersClause:
			visitor.VisitTopNestedWithOthersClause(&TopNestedWithOthersClause{SyntaxNode: node})
			return
		case SyntaxKindUnionOperator:
			visitor.VisitUnionOperator(&UnionOperator{SyntaxNode: node})
			return
		case SyntaxKindAsOperator:
			visitor.VisitAsOperator(&AsOperator{SyntaxNode: node})
			return
		case SyntaxKindSerializeOperator:
			visitor.VisitSerializeOperator(&SerializeOperator{SyntaxNode: node})
			return
		case SyntaxKindInvokeOperator:
			visitor.VisitInvokeOperator(&InvokeOperator{SyntaxNode: node})
			return
		case SyntaxKindRenderOperator:
			visitor.VisitRenderOperator(&RenderOperator{SyntaxNode: node})
			return
		case SyntaxKindMakeGraphOperator:
			visitor.VisitMakeGraphOperator(&MakeGraphOperator{SyntaxNode: node})
			return
		case SyntaxKindMakeGraphWithTablesAndKeysClause:
			visitor.VisitMakeGraphWithTablesAndKeysClause(&MakeGraphWithTablesAndKeysClause{SyntaxNode: node})
			return
		case SyntaxKindMakeGraphWithImplicitIdClause:
			visitor.VisitMakeGraphWithImplicitIdClause(&MakeGraphWithImplicitIdClause{SyntaxNode: node})
			return
		case SyntaxKindGraphMarkComponentsOperator:
			visitor.VisitGraphMarkComponentsOperator(&GraphMarkComponentsOperator{SyntaxNode: node})
			return
		case SyntaxKindGraphWhereNodesOperator:
			visitor.VisitGraphWhereNodesOperator(&GraphWhereNodesOperator{SyntaxNode: node})
			return
		case SyntaxKindGraphWhereEdgesOperator:
			visitor.VisitGraphWhereEdgesOperator(&GraphWhereEdgesOperator{SyntaxNode: node})
			return
		case SyntaxKindMakeGraphTableAndKeyClause:
			visitor.VisitMakeGraphTableAndKeyClause(&MakeGraphTableAndKeyClause{SyntaxNode: node})
			return
		case SyntaxKindMakeGraphPartitionedByClause:
			visitor.VisitMakeGraphPartitionedByClause(&MakeGraphPartitionedByClause{SyntaxNode: node})
			return
		case SyntaxKindGraphToTableOperator:
			visitor.VisitGraphToTableOperator(&GraphToTableOperator{SyntaxNode: node})
			return
		case SyntaxKindGraphToTableOutputClause:
			visitor.VisitGraphToTableOutputClause(&GraphToTableOutputClause{SyntaxNode: node})
			return
		case SyntaxKindGraphToTableAsClause:
			visitor.VisitGraphToTableAsClause(&GraphToTableAsClause{SyntaxNode: node})
			return
		case SyntaxKindGraphMatchOperator:
			visitor.VisitGraphMatchOperator(&GraphMatchOperator{SyntaxNode: node})
			return
		case SyntaxKindGraphShortestPathsOperator:
			visitor.VisitGraphShortestPathsOperator(&GraphShortestPathsOperator{SyntaxNode: node})
			return
		case SyntaxKindGraphMatchPattern:
			visitor.VisitGraphMatchPattern(&GraphMatchPattern{SyntaxNode: node})
			return
		case SyntaxKindGraphMatchPatternNode:
			visitor.VisitGraphMatchPatternNode(&GraphMatchPatternNode{SyntaxNode: node})
			return
		case SyntaxKindGraphMatchPatternEdge:
			visitor.VisitGraphMatchPatternEdge(&GraphMatchPatternEdge{SyntaxNode: node})
			return
		case SyntaxKindGraphMatchPatternEdgeRange:
			visitor.VisitGraphMatchPatternEdgeRange(&GraphMatchPatternEdgeRange{SyntaxNode: node})
			return
		case SyntaxKindWhereClause:
			visitor.VisitWhereClause(&WhereClause{SyntaxNode: node})
			return
		case SyntaxKindProjectClause:
			visitor.VisitProjectClause(&ProjectClause{SyntaxNode: node})
			return
		case SyntaxKindNameReferenceList:
			visitor.VisitNameReferenceList(&NameReferenceList{SyntaxNode: node})
			return
		case SyntaxKindRenderWithClause:
			visitor.VisitRenderWithClause(&RenderWithClause{SyntaxNode: node})
			return
		case SyntaxKindPrintOperator:
			visitor.VisitPrintOperator(&PrintOperator{SyntaxNode: node})
			return
		case SyntaxKindAliasStatement:
			visitor.VisitAliasStatement(&AliasStatement{SyntaxNode: node})
			return
		case SyntaxKindLetStatement:
			visitor.VisitLetStatement(&LetStatement{SyntaxNode: node})
			return
		case SyntaxKindFunctionDeclaration:
			visitor.VisitFunctionDeclaration(&FunctionDeclaration{SyntaxNode: node})
			return
		case SyntaxKindFunctionParameters:
			visitor.VisitFunctionParameters(&FunctionParameters{SyntaxNode: node})
			return
		case SyntaxKindFunctionParameter:
			visitor.VisitFunctionParameter(&FunctionParameter{SyntaxNode: node})
			return
		case SyntaxKindDefaultValueDeclaration:
			visitor.VisitDefaultValueDeclaration(&DefaultValueDeclaration{SyntaxNode: node})
			return
		case SyntaxKindFunctionBody:
			visitor.VisitFunctionBody(&FunctionBody{SyntaxNode: node})
			return
		case SyntaxKindSchemaTypeExpression:
			visitor.VisitSchemaTypeExpression(&SchemaTypeExpression{SyntaxNode: node})
			return
		case SyntaxKindExpressionStatement:
			visitor.VisitExpressionStatement(&ExpressionStatement{SyntaxNode: node})
			return
		case SyntaxKindMaterializeExpression:
			visitor.VisitMaterializeExpression(&MaterializeExpression{SyntaxNode: node})
			return
		case SyntaxKindSetOptionStatement:
			visitor.VisitSetOptionStatement(&SetOptionStatement{SyntaxNode: node})
			return
		case SyntaxKindOptionValueClause:
			visitor.VisitOptionValueClause(&OptionValueClause{SyntaxNode: node})
			return
		case SyntaxKindQueryParametersStatement:
			visitor.VisitQueryParametersStatement(&QueryParametersStatement{SyntaxNode: node})
			return
		case SyntaxKindRestrictStatementWithClause:
			visitor.VisitRestrictStatementWithClause(&RestrictStatementWithClause{SyntaxNode: node})
			return
		case SyntaxKindRestrictStatement:
			visitor.VisitRestrictStatement(&RestrictStatement{SyntaxNode: node})
			return
		case SyntaxKindPatternStatement:
			visitor.VisitPatternStatement(&PatternStatement{SyntaxNode: node})
			return
		case SyntaxKindPatternDeclaration:
			visitor.VisitPatternDeclaration(&PatternDeclaration{SyntaxNode: node})
			return
		case SyntaxKindPatternPathParameter:
			visitor.VisitPatternPathParameter(&PatternPathParameter{SyntaxNode: node})
			return
		case SyntaxKindPatternMatch:
			visitor.VisitPatternMatch(&PatternMatch{SyntaxNode: node})
			return
		case SyntaxKindPatternPathValue:
			visitor.VisitPatternPathValue(&PatternPathValue{SyntaxNode: node})
			return
		case SyntaxKindDataScopeExpression:
			visitor.VisitDataScopeExpression(&DataScopeExpression{SyntaxNode: node})
			return
		case SyntaxKindDataTableExpression:
			visitor.VisitDataTableExpression(&DataTableExpression{SyntaxNode: node})
			return
		case SyntaxKindRowSchema:
			visitor.VisitRowSchema(&RowSchema{SyntaxNode: node})
			return
		case SyntaxKindEvaluateRowSchema:
			visitor.VisitEvaluateRowSchema(&EvaluateRowSchema{SyntaxNode: node})
			return
		case SyntaxKindExternalDataExpression:
			visitor.VisitExternalDataExpression(&ExternalDataExpression{SyntaxNode: node})
			return
		case SyntaxKindContextualDataTableExpression:
			visitor.VisitContextualDataTableExpression(&ContextualDataTableExpression{SyntaxNode: node})
			return
		case SyntaxKindExternalDataWithClause:
			visitor.VisitExternalDataWithClause(&ExternalDataWithClause{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTableKindClause:
			visitor.VisitInlineExternalTableKindClause(&InlineExternalTableKindClause{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTableDataFormatClause:
			visitor.VisitInlineExternalTableDataFormatClause(&InlineExternalTableDataFormatClause{SyntaxNode: node})
			return
		case SyntaxKindDateTimePattern:
			visitor.VisitDateTimePattern(&DateTimePattern{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTablePathFormatPartitionColumnReference:
			visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(&InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTablePathFormatClause:
			visitor.VisitInlineExternalTablePathFormatClause(&InlineExternalTablePathFormatClause{SyntaxNode: node})
			return
		case SyntaxKindPartitionColumnDeclaration:
			visitor.VisitPartitionColumnDeclaration(&PartitionColumnDeclaration{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTablePartitionClause:
			visitor.VisitInlineExternalTablePartitionClause(&InlineExternalTablePartitionClause{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTableConnectionStringsClause:
			visitor.VisitInlineExternalTableConnectionStringsClause(&InlineExternalTableConnectionStringsClause{SyntaxNode: node})
			return
		case SyntaxKindInlineExternalTableExpression:
			visitor.VisitInlineExternalTableExpression(&InlineExternalTableExpression{SyntaxNode: node})
			return
		case SyntaxKindJoinOperator:
			visitor.VisitJoinOperator(&JoinOperator{SyntaxNode: node})
			return
		case SyntaxKindLookupOperator:
			visitor.VisitLookupOperator(&LookupOperator{SyntaxNode: node})
			return
		case SyntaxKindJoinOnClause:
			visitor.VisitJoinOnClause(&JoinOnClause{SyntaxNode: node})
			return
		case SyntaxKindJoinWhereClause:
			visitor.VisitJoinWhereClause(&JoinWhereClause{SyntaxNode: node})
			return
		case SyntaxKindScanOperator:
			visitor.VisitScanOperator(&ScanOperator{SyntaxNode: node})
			return
		case SyntaxKindAssertSchemaOperator:
			visitor.VisitAssertSchemaOperator(&AssertSchemaOperator{SyntaxNode: node})
			return
		case SyntaxKindScanDeclareClause:
			visitor.VisitScanDeclareClause(&ScanDeclareClause{SyntaxNode: node})
			return
		case SyntaxKindScanOrderByClause:
			visitor.VisitScanOrderByClause(&ScanOrderByClause{SyntaxNode: node})
			return
		case SyntaxKindScanPartitionByClause:
			visitor.VisitScanPartitionByClause(&ScanPartitionByClause{SyntaxNode: node})
			return
		case SyntaxKindScanStepOutput:
			visitor.VisitScanStepOutput(&ScanStepOutput{SyntaxNode: node})
			return
		case SyntaxKindScanStep:
			visitor.VisitScanStep(&ScanStep{SyntaxNode: node})
			return
		case SyntaxKindScanComputationClause:
			visitor.VisitScanComputationClause(&ScanComputationClause{SyntaxNode: node})
			return
		case SyntaxKindScanAssignment:
			visitor.VisitScanAssignment(&ScanAssignment{SyntaxNode: node})
			return
		case SyntaxKindPartitionByOperator:
			visitor.VisitPartitionByOperator(&PartitionByOperator{SyntaxNode: node})
			return
		case SyntaxKindPartitionByIdClause:
			visitor.VisitPartitionByIdClause(&PartitionByIdClause{SyntaxNode: node})
			return
		case SyntaxKindCommandWithValueClause:
			visitor.VisitCommandWithValueClause(&CommandWithValueClause{SyntaxNode: node})
			return
		case SyntaxKindCommandWithPropertyListClause:
			visitor.VisitCommandWithPropertyListClause(&CommandWithPropertyListClause{SyntaxNode: node})
			return
		case SyntaxKindUnknownCommand:
			visitor.VisitUnknownCommand(&UnknownCommand{SyntaxNode: node})
			return
		case SyntaxKindCustomCommand:
			visitor.VisitCustomCommand(&CustomCommand{SyntaxNode: node})
			return
		case SyntaxKindPartialCommand:
			visitor.VisitPartialCommand(&PartialCommand{SyntaxNode: node})
			return
		case SyntaxKindCommandAndSkippedTokens:
			visitor.VisitCommandAndSkippedTokens(&CommandAndSkippedTokens{SyntaxNode: node})
			return
		case SyntaxKindBadCommand:
			visitor.VisitBadCommand(&BadCommand{SyntaxNode: node})
			return
		case SyntaxKindCommandBlock:
			visitor.VisitCommandBlock(&CommandBlock{SyntaxNode: node})
			return
		case SyntaxKindGraphMergeOperator:
			visitor.VisitGraphMergeOperator(&GraphMergeOperator{SyntaxNode: node})
			return
		}
	}
}

func AcceptSyntaxVisitorResult[T any](element SyntaxElement, visitor SyntaxVisitorResult[T]) T {
	var zero T
	if element == nil || visitor == nil {
		return zero
	}
	switch node := element.(type) {
	case *CustomNode:
		return visitor.VisitCustom(node)
	case *SyntaxList:
		return visitor.VisitList(node)
	case *SeparatedElement:
		return visitor.VisitSeparatedElement(node)
	case *QueryBlock:
		return visitor.VisitQueryBlock(node)
	case *DirectiveBlock:
		return visitor.VisitDirectiveBlock(node)
	case *Directive:
		return visitor.VisitDirective(node)
	case *SkippedTokens:
		return visitor.VisitSkippedTokens(node)
	case *TypeOfLiteralExpression:
		return visitor.VisitTypeOfLiteralExpression(node)
	case *BadQueryOperator:
		return visitor.VisitBadQueryOperator(node)
	case *CompoundStringLiteralExpression:
		return visitor.VisitCompoundStringLiteralExpression(node)
	case *TokenName:
		return visitor.VisitTokenName(node)
	case *BracketedName:
		return visitor.VisitBracketedName(node)
	case *BracedName:
		return visitor.VisitBracedName(node)
	case *WildcardedName:
		return visitor.VisitWildcardedName(node)
	case *BracketedWildcardedName:
		return visitor.VisitBracketedWildcardedName(node)
	case *NameDeclaration:
		return visitor.VisitNameDeclaration(node)
	case *NameReference:
		return visitor.VisitNameReference(node)
	case *LiteralExpression:
		return visitor.VisitLiteralExpression(node)
	case *StarExpression:
		return visitor.VisitStarExpression(node)
	case *AtExpression:
		return visitor.VisitAtExpression(node)
	case *JsonPair:
		return visitor.VisitJsonPair(node)
	case *JsonObjectExpression:
		return visitor.VisitJsonObjectExpression(node)
	case *JsonArrayExpression:
		return visitor.VisitJsonArrayExpression(node)
	case *DynamicExpression:
		return visitor.VisitDynamicExpression(node)
	case *ParenthesizedExpression:
		return visitor.VisitParenthesizedExpression(node)
	case *ExpressionList:
		return visitor.VisitExpressionList(node)
	case *ExpressionCouple:
		return visitor.VisitExpressionCouple(node)
	case *PrefixUnaryExpression:
		return visitor.VisitPrefixUnaryExpression(node)
	case *BinaryExpression:
		return visitor.VisitBinaryExpression(node)
	case *InExpression:
		return visitor.VisitInExpression(node)
	case *HasAnyExpression:
		return visitor.VisitHasAnyExpression(node)
	case *HasAllExpression:
		return visitor.VisitHasAllExpression(node)
	case *BetweenExpression:
		return visitor.VisitBetweenExpression(node)
	case *FunctionCallExpression:
		return visitor.VisitFunctionCallExpression(node)
	case *ToScalarExpression:
		return visitor.VisitToScalarExpression(node)
	case *ToTableExpression:
		return visitor.VisitToTableExpression(node)
	case *MaterializedViewCombineExpression:
		return visitor.VisitMaterializedViewCombineExpression(node)
	case *MaterializedViewCombineNameClause:
		return visitor.VisitMaterializedViewCombineNameClause(node)
	case *MaterializedViewCombineClause:
		return visitor.VisitMaterializedViewCombineClause(node)
	case *SimpleNamedExpression:
		return visitor.VisitSimpleNamedExpression(node)
	case *RenameList:
		return visitor.VisitRenameList(node)
	case *CompoundNamedExpression:
		return visitor.VisitCompoundNamedExpression(node)
	case *BracketedExpression:
		return visitor.VisitBracketedExpression(node)
	case *PathExpression:
		return visitor.VisitPathExpression(node)
	case *ElementExpression:
		return visitor.VisitElementExpression(node)
	case *PipeExpression:
		return visitor.VisitPipeExpression(node)
	case *RangeOperator:
		return visitor.VisitRangeOperator(node)
	case *NamedParameter:
		return visitor.VisitNamedParameter(node)
	case *ConsumeOperator:
		return visitor.VisitConsumeOperator(node)
	case *CountOperator:
		return visitor.VisitCountOperator(node)
	case *CountAsIdentifierClause:
		return visitor.VisitCountAsIdentifierClause(node)
	case *ExecuteAndCacheOperator:
		return visitor.VisitExecuteAndCacheOperator(node)
	case *ExtendOperator:
		return visitor.VisitExtendOperator(node)
	case *FacetOperator:
		return visitor.VisitFacetOperator(node)
	case *FacetWithOperatorClause:
		return visitor.VisitFacetWithOperatorClause(node)
	case *FacetWithExpressionClause:
		return visitor.VisitFacetWithExpressionClause(node)
	case *FilterOperator:
		return visitor.VisitFilterOperator(node)
	case *GetSchemaOperator:
		return visitor.VisitGetSchemaOperator(node)
	case *FindOperator:
		return visitor.VisitFindOperator(node)
	case *DataScopeClause:
		return visitor.VisitDataScopeClause(node)
	case *TypedColumnReference:
		return visitor.VisitTypedColumnReference(node)
	case *FindInClause:
		return visitor.VisitFindInClause(node)
	case *FindProjectClause:
		return visitor.VisitFindProjectClause(node)
	case *PackExpression:
		return visitor.VisitPackExpression(node)
	case *NameAndTypeDeclaration:
		return visitor.VisitNameAndTypeDeclaration(node)
	case *PrimitiveTypeExpression:
		return visitor.VisitPrimitiveTypeExpression(node)
	case *SearchOperator:
		return visitor.VisitSearchOperator(node)
	case *ForkOperator:
		return visitor.VisitForkOperator(node)
	case *NameEqualsClause:
		return visitor.VisitNameEqualsClause(node)
	case *ForkExpression:
		return visitor.VisitForkExpression(node)
	case *MakeSeriesOperator:
		return visitor.VisitMakeSeriesOperator(node)
	case *MakeSeriesExpression:
		return visitor.VisitMakeSeriesExpression(node)
	case *DefaultExpressionClause:
		return visitor.VisitDefaultExpressionClause(node)
	case *MakeSeriesOnClause:
		return visitor.VisitMakeSeriesOnClause(node)
	case *MakeSeriesFromClause:
		return visitor.VisitMakeSeriesFromClause(node)
	case *MakeSeriesToClause:
		return visitor.VisitMakeSeriesToClause(node)
	case *MakeSeriesStepClause:
		return visitor.VisitMakeSeriesStepClause(node)
	case *MakeSeriesInRangeClause:
		return visitor.VisitMakeSeriesInRangeClause(node)
	case *MakeSeriesFromToStepClause:
		return visitor.VisitMakeSeriesFromToStepClause(node)
	case *MakeSeriesByClause:
		return visitor.VisitMakeSeriesByClause(node)
	case *MvExpandOperator:
		return visitor.VisitMvExpandOperator(node)
	case *MvExpandExpression:
		return visitor.VisitMvExpandExpression(node)
	case *MvExpandRowLimitClause:
		return visitor.VisitMvExpandRowLimitClause(node)
	case *MvApplyOperator:
		return visitor.VisitMvApplyOperator(node)
	case *MvApplyExpression:
		return visitor.VisitMvApplyExpression(node)
	case *MvApplyRowLimitClause:
		return visitor.VisitMvApplyRowLimitClause(node)
	case *MvApplyContextIdClause:
		return visitor.VisitMvApplyContextIdClause(node)
	case *MvApplySubqueryExpression:
		return visitor.VisitMvApplySubqueryExpression(node)
	case *ToTypeOfClause:
		return visitor.VisitToTypeOfClause(node)
	case *EvaluateSchemaClause:
		return visitor.VisitEvaluateSchemaClause(node)
	case *EvaluateOperator:
		return visitor.VisitEvaluateOperator(node)
	case *ParseOperator:
		return visitor.VisitParseOperator(node)
	case *ParseWhereOperator:
		return visitor.VisitParseWhereOperator(node)
	case *ParseKvWithClause:
		return visitor.VisitParseKvWithClause(node)
	case *ParseKvOperator:
		return visitor.VisitParseKvOperator(node)
	case *PartitionOperator:
		return visitor.VisitPartitionOperator(node)
	case *PartitionQuery:
		return visitor.VisitPartitionQuery(node)
	case *PartitionScope:
		return visitor.VisitPartitionScope(node)
	case *PartitionSubquery:
		return visitor.VisitPartitionSubquery(node)
	case *ProjectOperator:
		return visitor.VisitProjectOperator(node)
	case *ProjectAwayOperator:
		return visitor.VisitProjectAwayOperator(node)
	case *ProjectByNamesOperator:
		return visitor.VisitProjectByNamesOperator(node)
	case *ProjectKeepOperator:
		return visitor.VisitProjectKeepOperator(node)
	case *ProjectRenameOperator:
		return visitor.VisitProjectRenameOperator(node)
	case *ProjectReorderOperator:
		return visitor.VisitProjectReorderOperator(node)
	case *SampleOperator:
		return visitor.VisitSampleOperator(node)
	case *SampleDistinctOperator:
		return visitor.VisitSampleDistinctOperator(node)
	case *EntityGroup:
		return visitor.VisitEntityGroup(node)
	case *ReduceByOperator:
		return visitor.VisitReduceByOperator(node)
	case *ReduceByWithClause:
		return visitor.VisitReduceByWithClause(node)
	case *SummarizeOperator:
		return visitor.VisitSummarizeOperator(node)
	case *MacroExpandScopeReferenceName:
		return visitor.VisitMacroExpandScopeReferenceName(node)
	case *MacroExpandOperator:
		return visitor.VisitMacroExpandOperator(node)
	case *SummarizeByClause:
		return visitor.VisitSummarizeByClause(node)
	case *DistinctOperator:
		return visitor.VisitDistinctOperator(node)
	case *TakeOperator:
		return visitor.VisitTakeOperator(node)
	case *SortOperator:
		return visitor.VisitSortOperator(node)
	case *OrderedExpression:
		return visitor.VisitOrderedExpression(node)
	case *OrderingClause:
		return visitor.VisitOrderingClause(node)
	case *OrderingNullsClause:
		return visitor.VisitOrderingNullsClause(node)
	case *TopHittersOperator:
		return visitor.VisitTopHittersOperator(node)
	case *TopHittersByClause:
		return visitor.VisitTopHittersByClause(node)
	case *TopOperator:
		return visitor.VisitTopOperator(node)
	case *TopNestedOperator:
		return visitor.VisitTopNestedOperator(node)
	case *TopNestedClause:
		return visitor.VisitTopNestedClause(node)
	case *TopNestedWithOthersClause:
		return visitor.VisitTopNestedWithOthersClause(node)
	case *UnionOperator:
		return visitor.VisitUnionOperator(node)
	case *AsOperator:
		return visitor.VisitAsOperator(node)
	case *SerializeOperator:
		return visitor.VisitSerializeOperator(node)
	case *InvokeOperator:
		return visitor.VisitInvokeOperator(node)
	case *RenderOperator:
		return visitor.VisitRenderOperator(node)
	case *MakeGraphOperator:
		return visitor.VisitMakeGraphOperator(node)
	case *MakeGraphWithTablesAndKeysClause:
		return visitor.VisitMakeGraphWithTablesAndKeysClause(node)
	case *MakeGraphWithImplicitIdClause:
		return visitor.VisitMakeGraphWithImplicitIdClause(node)
	case *GraphMarkComponentsOperator:
		return visitor.VisitGraphMarkComponentsOperator(node)
	case *GraphWhereNodesOperator:
		return visitor.VisitGraphWhereNodesOperator(node)
	case *GraphWhereEdgesOperator:
		return visitor.VisitGraphWhereEdgesOperator(node)
	case *MakeGraphTableAndKeyClause:
		return visitor.VisitMakeGraphTableAndKeyClause(node)
	case *MakeGraphPartitionedByClause:
		return visitor.VisitMakeGraphPartitionedByClause(node)
	case *GraphToTableOperator:
		return visitor.VisitGraphToTableOperator(node)
	case *GraphToTableOutputClause:
		return visitor.VisitGraphToTableOutputClause(node)
	case *GraphToTableAsClause:
		return visitor.VisitGraphToTableAsClause(node)
	case *GraphMatchOperator:
		return visitor.VisitGraphMatchOperator(node)
	case *GraphShortestPathsOperator:
		return visitor.VisitGraphShortestPathsOperator(node)
	case *GraphMatchPattern:
		return visitor.VisitGraphMatchPattern(node)
	case *GraphMatchPatternNode:
		return visitor.VisitGraphMatchPatternNode(node)
	case *GraphMatchPatternEdge:
		return visitor.VisitGraphMatchPatternEdge(node)
	case *GraphMatchPatternEdgeRange:
		return visitor.VisitGraphMatchPatternEdgeRange(node)
	case *WhereClause:
		return visitor.VisitWhereClause(node)
	case *ProjectClause:
		return visitor.VisitProjectClause(node)
	case *NameReferenceList:
		return visitor.VisitNameReferenceList(node)
	case *RenderWithClause:
		return visitor.VisitRenderWithClause(node)
	case *PrintOperator:
		return visitor.VisitPrintOperator(node)
	case *AliasStatement:
		return visitor.VisitAliasStatement(node)
	case *LetStatement:
		return visitor.VisitLetStatement(node)
	case *FunctionDeclaration:
		return visitor.VisitFunctionDeclaration(node)
	case *FunctionParameters:
		return visitor.VisitFunctionParameters(node)
	case *FunctionParameter:
		return visitor.VisitFunctionParameter(node)
	case *DefaultValueDeclaration:
		return visitor.VisitDefaultValueDeclaration(node)
	case *FunctionBody:
		return visitor.VisitFunctionBody(node)
	case *SchemaTypeExpression:
		return visitor.VisitSchemaTypeExpression(node)
	case *ExpressionStatement:
		return visitor.VisitExpressionStatement(node)
	case *MaterializeExpression:
		return visitor.VisitMaterializeExpression(node)
	case *SetOptionStatement:
		return visitor.VisitSetOptionStatement(node)
	case *OptionValueClause:
		return visitor.VisitOptionValueClause(node)
	case *QueryParametersStatement:
		return visitor.VisitQueryParametersStatement(node)
	case *RestrictStatementWithClause:
		return visitor.VisitRestrictStatementWithClause(node)
	case *RestrictStatement:
		return visitor.VisitRestrictStatement(node)
	case *PatternStatement:
		return visitor.VisitPatternStatement(node)
	case *PatternDeclaration:
		return visitor.VisitPatternDeclaration(node)
	case *PatternPathParameter:
		return visitor.VisitPatternPathParameter(node)
	case *PatternMatch:
		return visitor.VisitPatternMatch(node)
	case *PatternPathValue:
		return visitor.VisitPatternPathValue(node)
	case *DataScopeExpression:
		return visitor.VisitDataScopeExpression(node)
	case *DataTableExpression:
		return visitor.VisitDataTableExpression(node)
	case *RowSchema:
		return visitor.VisitRowSchema(node)
	case *EvaluateRowSchema:
		return visitor.VisitEvaluateRowSchema(node)
	case *ExternalDataExpression:
		return visitor.VisitExternalDataExpression(node)
	case *ContextualDataTableExpression:
		return visitor.VisitContextualDataTableExpression(node)
	case *ExternalDataWithClause:
		return visitor.VisitExternalDataWithClause(node)
	case *InlineExternalTableKindClause:
		return visitor.VisitInlineExternalTableKindClause(node)
	case *InlineExternalTableDataFormatClause:
		return visitor.VisitInlineExternalTableDataFormatClause(node)
	case *DateTimePattern:
		return visitor.VisitDateTimePattern(node)
	case *InlineExternalTablePathFormatPartitionColumnReference:
		return visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(node)
	case *InlineExternalTablePathFormatClause:
		return visitor.VisitInlineExternalTablePathFormatClause(node)
	case *PartitionColumnDeclaration:
		return visitor.VisitPartitionColumnDeclaration(node)
	case *InlineExternalTablePartitionClause:
		return visitor.VisitInlineExternalTablePartitionClause(node)
	case *InlineExternalTableConnectionStringsClause:
		return visitor.VisitInlineExternalTableConnectionStringsClause(node)
	case *InlineExternalTableExpression:
		return visitor.VisitInlineExternalTableExpression(node)
	case *JoinOperator:
		return visitor.VisitJoinOperator(node)
	case *LookupOperator:
		return visitor.VisitLookupOperator(node)
	case *JoinOnClause:
		return visitor.VisitJoinOnClause(node)
	case *JoinWhereClause:
		return visitor.VisitJoinWhereClause(node)
	case *ScanOperator:
		return visitor.VisitScanOperator(node)
	case *AssertSchemaOperator:
		return visitor.VisitAssertSchemaOperator(node)
	case *ScanDeclareClause:
		return visitor.VisitScanDeclareClause(node)
	case *ScanOrderByClause:
		return visitor.VisitScanOrderByClause(node)
	case *ScanPartitionByClause:
		return visitor.VisitScanPartitionByClause(node)
	case *ScanStepOutput:
		return visitor.VisitScanStepOutput(node)
	case *ScanStep:
		return visitor.VisitScanStep(node)
	case *ScanComputationClause:
		return visitor.VisitScanComputationClause(node)
	case *ScanAssignment:
		return visitor.VisitScanAssignment(node)
	case *PartitionByOperator:
		return visitor.VisitPartitionByOperator(node)
	case *PartitionByIdClause:
		return visitor.VisitPartitionByIdClause(node)
	case *CommandWithValueClause:
		return visitor.VisitCommandWithValueClause(node)
	case *CommandWithPropertyListClause:
		return visitor.VisitCommandWithPropertyListClause(node)
	case *UnknownCommand:
		return visitor.VisitUnknownCommand(node)
	case *CustomCommand:
		return visitor.VisitCustomCommand(node)
	case *PartialCommand:
		return visitor.VisitPartialCommand(node)
	case *CommandAndSkippedTokens:
		return visitor.VisitCommandAndSkippedTokens(node)
	case *BadCommand:
		return visitor.VisitBadCommand(node)
	case *CommandBlock:
		return visitor.VisitCommandBlock(node)
	case *GraphMergeOperator:
		return visitor.VisitGraphMergeOperator(node)
	case *SyntaxNode:
		switch node.GeneratedNodeName() {
		case "QueryBlock":
			return visitor.VisitQueryBlock(&QueryBlock{SyntaxNode: node})
		case "DirectiveBlock":
			return visitor.VisitDirectiveBlock(&DirectiveBlock{SyntaxNode: node})
		case "Directive":
			return visitor.VisitDirective(&Directive{SyntaxNode: node})
		case "SkippedTokens":
			return visitor.VisitSkippedTokens(&SkippedTokens{SyntaxNode: node})
		case "TypeOfLiteralExpression":
			return visitor.VisitTypeOfLiteralExpression(&TypeOfLiteralExpression{SyntaxNode: node})
		case "BadQueryOperator":
			return visitor.VisitBadQueryOperator(&BadQueryOperator{SyntaxNode: node})
		case "CompoundStringLiteralExpression":
			return visitor.VisitCompoundStringLiteralExpression(&CompoundStringLiteralExpression{SyntaxNode: node})
		case "TokenName":
			return visitor.VisitTokenName(&TokenName{SyntaxNode: node})
		case "BracketedName":
			return visitor.VisitBracketedName(&BracketedName{SyntaxNode: node})
		case "BracedName":
			return visitor.VisitBracedName(&BracedName{SyntaxNode: node})
		case "WildcardedName":
			return visitor.VisitWildcardedName(&WildcardedName{SyntaxNode: node})
		case "BracketedWildcardedName":
			return visitor.VisitBracketedWildcardedName(&BracketedWildcardedName{SyntaxNode: node})
		case "NameDeclaration":
			return visitor.VisitNameDeclaration(&NameDeclaration{SyntaxNode: node})
		case "NameReference":
			return visitor.VisitNameReference(&NameReference{SyntaxNode: node})
		case "LiteralExpression":
			return visitor.VisitLiteralExpression(&LiteralExpression{SyntaxNode: node})
		case "StarExpression":
			return visitor.VisitStarExpression(&StarExpression{SyntaxNode: node})
		case "AtExpression":
			return visitor.VisitAtExpression(&AtExpression{SyntaxNode: node})
		case "JsonPair":
			return visitor.VisitJsonPair(&JsonPair{SyntaxNode: node})
		case "JsonObjectExpression":
			return visitor.VisitJsonObjectExpression(&JsonObjectExpression{SyntaxNode: node})
		case "JsonArrayExpression":
			return visitor.VisitJsonArrayExpression(&JsonArrayExpression{SyntaxNode: node})
		case "DynamicExpression":
			return visitor.VisitDynamicExpression(&DynamicExpression{SyntaxNode: node})
		case "ParenthesizedExpression":
			return visitor.VisitParenthesizedExpression(&ParenthesizedExpression{SyntaxNode: node})
		case "ExpressionList":
			return visitor.VisitExpressionList(&ExpressionList{SyntaxNode: node})
		case "ExpressionCouple":
			return visitor.VisitExpressionCouple(&ExpressionCouple{SyntaxNode: node})
		case "PrefixUnaryExpression":
			return visitor.VisitPrefixUnaryExpression(&PrefixUnaryExpression{SyntaxNode: node})
		case "BinaryExpression":
			return visitor.VisitBinaryExpression(&BinaryExpression{SyntaxNode: node})
		case "InExpression":
			return visitor.VisitInExpression(&InExpression{SyntaxNode: node})
		case "HasAnyExpression":
			return visitor.VisitHasAnyExpression(&HasAnyExpression{SyntaxNode: node})
		case "HasAllExpression":
			return visitor.VisitHasAllExpression(&HasAllExpression{SyntaxNode: node})
		case "BetweenExpression":
			return visitor.VisitBetweenExpression(&BetweenExpression{SyntaxNode: node})
		case "FunctionCallExpression":
			return visitor.VisitFunctionCallExpression(&FunctionCallExpression{SyntaxNode: node})
		case "ToScalarExpression":
			return visitor.VisitToScalarExpression(&ToScalarExpression{SyntaxNode: node})
		case "ToTableExpression":
			return visitor.VisitToTableExpression(&ToTableExpression{SyntaxNode: node})
		case "MaterializedViewCombineExpression":
			return visitor.VisitMaterializedViewCombineExpression(&MaterializedViewCombineExpression{SyntaxNode: node})
		case "MaterializedViewCombineNameClause":
			return visitor.VisitMaterializedViewCombineNameClause(&MaterializedViewCombineNameClause{SyntaxNode: node})
		case "MaterializedViewCombineClause":
			return visitor.VisitMaterializedViewCombineClause(&MaterializedViewCombineClause{SyntaxNode: node})
		case "SimpleNamedExpression":
			return visitor.VisitSimpleNamedExpression(&SimpleNamedExpression{SyntaxNode: node})
		case "RenameList":
			return visitor.VisitRenameList(&RenameList{SyntaxNode: node})
		case "CompoundNamedExpression":
			return visitor.VisitCompoundNamedExpression(&CompoundNamedExpression{SyntaxNode: node})
		case "BracketedExpression":
			return visitor.VisitBracketedExpression(&BracketedExpression{SyntaxNode: node})
		case "PathExpression":
			return visitor.VisitPathExpression(&PathExpression{SyntaxNode: node})
		case "ElementExpression":
			return visitor.VisitElementExpression(&ElementExpression{SyntaxNode: node})
		case "PipeExpression":
			return visitor.VisitPipeExpression(&PipeExpression{SyntaxNode: node})
		case "RangeOperator":
			return visitor.VisitRangeOperator(&RangeOperator{SyntaxNode: node})
		case "NamedParameter":
			return visitor.VisitNamedParameter(&NamedParameter{SyntaxNode: node})
		case "ConsumeOperator":
			return visitor.VisitConsumeOperator(&ConsumeOperator{SyntaxNode: node})
		case "CountOperator":
			return visitor.VisitCountOperator(&CountOperator{SyntaxNode: node})
		case "CountAsIdentifierClause":
			return visitor.VisitCountAsIdentifierClause(&CountAsIdentifierClause{SyntaxNode: node})
		case "ExecuteAndCacheOperator":
			return visitor.VisitExecuteAndCacheOperator(&ExecuteAndCacheOperator{SyntaxNode: node})
		case "ExtendOperator":
			return visitor.VisitExtendOperator(&ExtendOperator{SyntaxNode: node})
		case "FacetOperator":
			return visitor.VisitFacetOperator(&FacetOperator{SyntaxNode: node})
		case "FacetWithOperatorClause":
			return visitor.VisitFacetWithOperatorClause(&FacetWithOperatorClause{SyntaxNode: node})
		case "FacetWithExpressionClause":
			return visitor.VisitFacetWithExpressionClause(&FacetWithExpressionClause{SyntaxNode: node})
		case "FilterOperator":
			return visitor.VisitFilterOperator(&FilterOperator{SyntaxNode: node})
		case "GetSchemaOperator":
			return visitor.VisitGetSchemaOperator(&GetSchemaOperator{SyntaxNode: node})
		case "FindOperator":
			return visitor.VisitFindOperator(&FindOperator{SyntaxNode: node})
		case "DataScopeClause":
			return visitor.VisitDataScopeClause(&DataScopeClause{SyntaxNode: node})
		case "TypedColumnReference":
			return visitor.VisitTypedColumnReference(&TypedColumnReference{SyntaxNode: node})
		case "FindInClause":
			return visitor.VisitFindInClause(&FindInClause{SyntaxNode: node})
		case "FindProjectClause":
			return visitor.VisitFindProjectClause(&FindProjectClause{SyntaxNode: node})
		case "PackExpression":
			return visitor.VisitPackExpression(&PackExpression{SyntaxNode: node})
		case "NameAndTypeDeclaration":
			return visitor.VisitNameAndTypeDeclaration(&NameAndTypeDeclaration{SyntaxNode: node})
		case "PrimitiveTypeExpression":
			return visitor.VisitPrimitiveTypeExpression(&PrimitiveTypeExpression{SyntaxNode: node})
		case "SearchOperator":
			return visitor.VisitSearchOperator(&SearchOperator{SyntaxNode: node})
		case "ForkOperator":
			return visitor.VisitForkOperator(&ForkOperator{SyntaxNode: node})
		case "NameEqualsClause":
			return visitor.VisitNameEqualsClause(&NameEqualsClause{SyntaxNode: node})
		case "ForkExpression":
			return visitor.VisitForkExpression(&ForkExpression{SyntaxNode: node})
		case "MakeSeriesOperator":
			return visitor.VisitMakeSeriesOperator(&MakeSeriesOperator{SyntaxNode: node})
		case "MakeSeriesExpression":
			return visitor.VisitMakeSeriesExpression(&MakeSeriesExpression{SyntaxNode: node})
		case "DefaultExpressionClause":
			return visitor.VisitDefaultExpressionClause(&DefaultExpressionClause{SyntaxNode: node})
		case "MakeSeriesOnClause":
			return visitor.VisitMakeSeriesOnClause(&MakeSeriesOnClause{SyntaxNode: node})
		case "MakeSeriesFromClause":
			return visitor.VisitMakeSeriesFromClause(&MakeSeriesFromClause{SyntaxNode: node})
		case "MakeSeriesToClause":
			return visitor.VisitMakeSeriesToClause(&MakeSeriesToClause{SyntaxNode: node})
		case "MakeSeriesStepClause":
			return visitor.VisitMakeSeriesStepClause(&MakeSeriesStepClause{SyntaxNode: node})
		case "MakeSeriesInRangeClause":
			return visitor.VisitMakeSeriesInRangeClause(&MakeSeriesInRangeClause{SyntaxNode: node})
		case "MakeSeriesFromToStepClause":
			return visitor.VisitMakeSeriesFromToStepClause(&MakeSeriesFromToStepClause{SyntaxNode: node})
		case "MakeSeriesByClause":
			return visitor.VisitMakeSeriesByClause(&MakeSeriesByClause{SyntaxNode: node})
		case "MvExpandOperator":
			return visitor.VisitMvExpandOperator(&MvExpandOperator{SyntaxNode: node})
		case "MvExpandExpression":
			return visitor.VisitMvExpandExpression(&MvExpandExpression{SyntaxNode: node})
		case "MvExpandRowLimitClause":
			return visitor.VisitMvExpandRowLimitClause(&MvExpandRowLimitClause{SyntaxNode: node})
		case "MvApplyOperator":
			return visitor.VisitMvApplyOperator(&MvApplyOperator{SyntaxNode: node})
		case "MvApplyExpression":
			return visitor.VisitMvApplyExpression(&MvApplyExpression{SyntaxNode: node})
		case "MvApplyRowLimitClause":
			return visitor.VisitMvApplyRowLimitClause(&MvApplyRowLimitClause{SyntaxNode: node})
		case "MvApplyContextIdClause":
			return visitor.VisitMvApplyContextIdClause(&MvApplyContextIdClause{SyntaxNode: node})
		case "MvApplySubqueryExpression":
			return visitor.VisitMvApplySubqueryExpression(&MvApplySubqueryExpression{SyntaxNode: node})
		case "ToTypeOfClause":
			return visitor.VisitToTypeOfClause(&ToTypeOfClause{SyntaxNode: node})
		case "EvaluateSchemaClause":
			return visitor.VisitEvaluateSchemaClause(&EvaluateSchemaClause{SyntaxNode: node})
		case "EvaluateOperator":
			return visitor.VisitEvaluateOperator(&EvaluateOperator{SyntaxNode: node})
		case "ParseOperator":
			return visitor.VisitParseOperator(&ParseOperator{SyntaxNode: node})
		case "ParseWhereOperator":
			return visitor.VisitParseWhereOperator(&ParseWhereOperator{SyntaxNode: node})
		case "ParseKvWithClause":
			return visitor.VisitParseKvWithClause(&ParseKvWithClause{SyntaxNode: node})
		case "ParseKvOperator":
			return visitor.VisitParseKvOperator(&ParseKvOperator{SyntaxNode: node})
		case "PartitionOperator":
			return visitor.VisitPartitionOperator(&PartitionOperator{SyntaxNode: node})
		case "PartitionQuery":
			return visitor.VisitPartitionQuery(&PartitionQuery{SyntaxNode: node})
		case "PartitionScope":
			return visitor.VisitPartitionScope(&PartitionScope{SyntaxNode: node})
		case "PartitionSubquery":
			return visitor.VisitPartitionSubquery(&PartitionSubquery{SyntaxNode: node})
		case "ProjectOperator":
			return visitor.VisitProjectOperator(&ProjectOperator{SyntaxNode: node})
		case "ProjectAwayOperator":
			return visitor.VisitProjectAwayOperator(&ProjectAwayOperator{SyntaxNode: node})
		case "ProjectByNamesOperator":
			return visitor.VisitProjectByNamesOperator(&ProjectByNamesOperator{SyntaxNode: node})
		case "ProjectKeepOperator":
			return visitor.VisitProjectKeepOperator(&ProjectKeepOperator{SyntaxNode: node})
		case "ProjectRenameOperator":
			return visitor.VisitProjectRenameOperator(&ProjectRenameOperator{SyntaxNode: node})
		case "ProjectReorderOperator":
			return visitor.VisitProjectReorderOperator(&ProjectReorderOperator{SyntaxNode: node})
		case "SampleOperator":
			return visitor.VisitSampleOperator(&SampleOperator{SyntaxNode: node})
		case "SampleDistinctOperator":
			return visitor.VisitSampleDistinctOperator(&SampleDistinctOperator{SyntaxNode: node})
		case "EntityGroup":
			return visitor.VisitEntityGroup(&EntityGroup{SyntaxNode: node})
		case "ReduceByOperator":
			return visitor.VisitReduceByOperator(&ReduceByOperator{SyntaxNode: node})
		case "ReduceByWithClause":
			return visitor.VisitReduceByWithClause(&ReduceByWithClause{SyntaxNode: node})
		case "SummarizeOperator":
			return visitor.VisitSummarizeOperator(&SummarizeOperator{SyntaxNode: node})
		case "MacroExpandScopeReferenceName":
			return visitor.VisitMacroExpandScopeReferenceName(&MacroExpandScopeReferenceName{SyntaxNode: node})
		case "MacroExpandOperator":
			return visitor.VisitMacroExpandOperator(&MacroExpandOperator{SyntaxNode: node})
		case "SummarizeByClause":
			return visitor.VisitSummarizeByClause(&SummarizeByClause{SyntaxNode: node})
		case "DistinctOperator":
			return visitor.VisitDistinctOperator(&DistinctOperator{SyntaxNode: node})
		case "TakeOperator":
			return visitor.VisitTakeOperator(&TakeOperator{SyntaxNode: node})
		case "SortOperator":
			return visitor.VisitSortOperator(&SortOperator{SyntaxNode: node})
		case "OrderedExpression":
			return visitor.VisitOrderedExpression(&OrderedExpression{SyntaxNode: node})
		case "OrderingClause":
			return visitor.VisitOrderingClause(&OrderingClause{SyntaxNode: node})
		case "OrderingNullsClause":
			return visitor.VisitOrderingNullsClause(&OrderingNullsClause{SyntaxNode: node})
		case "TopHittersOperator":
			return visitor.VisitTopHittersOperator(&TopHittersOperator{SyntaxNode: node})
		case "TopHittersByClause":
			return visitor.VisitTopHittersByClause(&TopHittersByClause{SyntaxNode: node})
		case "TopOperator":
			return visitor.VisitTopOperator(&TopOperator{SyntaxNode: node})
		case "TopNestedOperator":
			return visitor.VisitTopNestedOperator(&TopNestedOperator{SyntaxNode: node})
		case "TopNestedClause":
			return visitor.VisitTopNestedClause(&TopNestedClause{SyntaxNode: node})
		case "TopNestedWithOthersClause":
			return visitor.VisitTopNestedWithOthersClause(&TopNestedWithOthersClause{SyntaxNode: node})
		case "UnionOperator":
			return visitor.VisitUnionOperator(&UnionOperator{SyntaxNode: node})
		case "AsOperator":
			return visitor.VisitAsOperator(&AsOperator{SyntaxNode: node})
		case "SerializeOperator":
			return visitor.VisitSerializeOperator(&SerializeOperator{SyntaxNode: node})
		case "InvokeOperator":
			return visitor.VisitInvokeOperator(&InvokeOperator{SyntaxNode: node})
		case "RenderOperator":
			return visitor.VisitRenderOperator(&RenderOperator{SyntaxNode: node})
		case "MakeGraphOperator":
			return visitor.VisitMakeGraphOperator(&MakeGraphOperator{SyntaxNode: node})
		case "MakeGraphWithTablesAndKeysClause":
			return visitor.VisitMakeGraphWithTablesAndKeysClause(&MakeGraphWithTablesAndKeysClause{SyntaxNode: node})
		case "MakeGraphWithImplicitIdClause":
			return visitor.VisitMakeGraphWithImplicitIdClause(&MakeGraphWithImplicitIdClause{SyntaxNode: node})
		case "GraphMarkComponentsOperator":
			return visitor.VisitGraphMarkComponentsOperator(&GraphMarkComponentsOperator{SyntaxNode: node})
		case "GraphWhereNodesOperator":
			return visitor.VisitGraphWhereNodesOperator(&GraphWhereNodesOperator{SyntaxNode: node})
		case "GraphWhereEdgesOperator":
			return visitor.VisitGraphWhereEdgesOperator(&GraphWhereEdgesOperator{SyntaxNode: node})
		case "MakeGraphTableAndKeyClause":
			return visitor.VisitMakeGraphTableAndKeyClause(&MakeGraphTableAndKeyClause{SyntaxNode: node})
		case "MakeGraphPartitionedByClause":
			return visitor.VisitMakeGraphPartitionedByClause(&MakeGraphPartitionedByClause{SyntaxNode: node})
		case "GraphToTableOperator":
			return visitor.VisitGraphToTableOperator(&GraphToTableOperator{SyntaxNode: node})
		case "GraphToTableOutputClause":
			return visitor.VisitGraphToTableOutputClause(&GraphToTableOutputClause{SyntaxNode: node})
		case "GraphToTableAsClause":
			return visitor.VisitGraphToTableAsClause(&GraphToTableAsClause{SyntaxNode: node})
		case "GraphMatchOperator":
			return visitor.VisitGraphMatchOperator(&GraphMatchOperator{SyntaxNode: node})
		case "GraphShortestPathsOperator":
			return visitor.VisitGraphShortestPathsOperator(&GraphShortestPathsOperator{SyntaxNode: node})
		case "GraphMatchPattern":
			return visitor.VisitGraphMatchPattern(&GraphMatchPattern{SyntaxNode: node})
		case "GraphMatchPatternNode":
			return visitor.VisitGraphMatchPatternNode(&GraphMatchPatternNode{SyntaxNode: node})
		case "GraphMatchPatternEdge":
			return visitor.VisitGraphMatchPatternEdge(&GraphMatchPatternEdge{SyntaxNode: node})
		case "GraphMatchPatternEdgeRange":
			return visitor.VisitGraphMatchPatternEdgeRange(&GraphMatchPatternEdgeRange{SyntaxNode: node})
		case "WhereClause":
			return visitor.VisitWhereClause(&WhereClause{SyntaxNode: node})
		case "ProjectClause":
			return visitor.VisitProjectClause(&ProjectClause{SyntaxNode: node})
		case "NameReferenceList":
			return visitor.VisitNameReferenceList(&NameReferenceList{SyntaxNode: node})
		case "RenderWithClause":
			return visitor.VisitRenderWithClause(&RenderWithClause{SyntaxNode: node})
		case "PrintOperator":
			return visitor.VisitPrintOperator(&PrintOperator{SyntaxNode: node})
		case "AliasStatement":
			return visitor.VisitAliasStatement(&AliasStatement{SyntaxNode: node})
		case "LetStatement":
			return visitor.VisitLetStatement(&LetStatement{SyntaxNode: node})
		case "FunctionDeclaration":
			return visitor.VisitFunctionDeclaration(&FunctionDeclaration{SyntaxNode: node})
		case "FunctionParameters":
			return visitor.VisitFunctionParameters(&FunctionParameters{SyntaxNode: node})
		case "FunctionParameter":
			return visitor.VisitFunctionParameter(&FunctionParameter{SyntaxNode: node})
		case "DefaultValueDeclaration":
			return visitor.VisitDefaultValueDeclaration(&DefaultValueDeclaration{SyntaxNode: node})
		case "FunctionBody":
			return visitor.VisitFunctionBody(&FunctionBody{SyntaxNode: node})
		case "SchemaTypeExpression":
			return visitor.VisitSchemaTypeExpression(&SchemaTypeExpression{SyntaxNode: node})
		case "ExpressionStatement":
			return visitor.VisitExpressionStatement(&ExpressionStatement{SyntaxNode: node})
		case "MaterializeExpression":
			return visitor.VisitMaterializeExpression(&MaterializeExpression{SyntaxNode: node})
		case "SetOptionStatement":
			return visitor.VisitSetOptionStatement(&SetOptionStatement{SyntaxNode: node})
		case "OptionValueClause":
			return visitor.VisitOptionValueClause(&OptionValueClause{SyntaxNode: node})
		case "QueryParametersStatement":
			return visitor.VisitQueryParametersStatement(&QueryParametersStatement{SyntaxNode: node})
		case "RestrictStatementWithClause":
			return visitor.VisitRestrictStatementWithClause(&RestrictStatementWithClause{SyntaxNode: node})
		case "RestrictStatement":
			return visitor.VisitRestrictStatement(&RestrictStatement{SyntaxNode: node})
		case "PatternStatement":
			return visitor.VisitPatternStatement(&PatternStatement{SyntaxNode: node})
		case "PatternDeclaration":
			return visitor.VisitPatternDeclaration(&PatternDeclaration{SyntaxNode: node})
		case "PatternPathParameter":
			return visitor.VisitPatternPathParameter(&PatternPathParameter{SyntaxNode: node})
		case "PatternMatch":
			return visitor.VisitPatternMatch(&PatternMatch{SyntaxNode: node})
		case "PatternPathValue":
			return visitor.VisitPatternPathValue(&PatternPathValue{SyntaxNode: node})
		case "DataScopeExpression":
			return visitor.VisitDataScopeExpression(&DataScopeExpression{SyntaxNode: node})
		case "DataTableExpression":
			return visitor.VisitDataTableExpression(&DataTableExpression{SyntaxNode: node})
		case "RowSchema":
			return visitor.VisitRowSchema(&RowSchema{SyntaxNode: node})
		case "EvaluateRowSchema":
			return visitor.VisitEvaluateRowSchema(&EvaluateRowSchema{SyntaxNode: node})
		case "ExternalDataExpression":
			return visitor.VisitExternalDataExpression(&ExternalDataExpression{SyntaxNode: node})
		case "ContextualDataTableExpression":
			return visitor.VisitContextualDataTableExpression(&ContextualDataTableExpression{SyntaxNode: node})
		case "ExternalDataWithClause":
			return visitor.VisitExternalDataWithClause(&ExternalDataWithClause{SyntaxNode: node})
		case "InlineExternalTableKindClause":
			return visitor.VisitInlineExternalTableKindClause(&InlineExternalTableKindClause{SyntaxNode: node})
		case "InlineExternalTableDataFormatClause":
			return visitor.VisitInlineExternalTableDataFormatClause(&InlineExternalTableDataFormatClause{SyntaxNode: node})
		case "DateTimePattern":
			return visitor.VisitDateTimePattern(&DateTimePattern{SyntaxNode: node})
		case "InlineExternalTablePathFormatPartitionColumnReference":
			return visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(&InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node})
		case "InlineExternalTablePathFormatClause":
			return visitor.VisitInlineExternalTablePathFormatClause(&InlineExternalTablePathFormatClause{SyntaxNode: node})
		case "PartitionColumnDeclaration":
			return visitor.VisitPartitionColumnDeclaration(&PartitionColumnDeclaration{SyntaxNode: node})
		case "InlineExternalTablePartitionClause":
			return visitor.VisitInlineExternalTablePartitionClause(&InlineExternalTablePartitionClause{SyntaxNode: node})
		case "InlineExternalTableConnectionStringsClause":
			return visitor.VisitInlineExternalTableConnectionStringsClause(&InlineExternalTableConnectionStringsClause{SyntaxNode: node})
		case "InlineExternalTableExpression":
			return visitor.VisitInlineExternalTableExpression(&InlineExternalTableExpression{SyntaxNode: node})
		case "JoinOperator":
			return visitor.VisitJoinOperator(&JoinOperator{SyntaxNode: node})
		case "LookupOperator":
			return visitor.VisitLookupOperator(&LookupOperator{SyntaxNode: node})
		case "JoinOnClause":
			return visitor.VisitJoinOnClause(&JoinOnClause{SyntaxNode: node})
		case "JoinWhereClause":
			return visitor.VisitJoinWhereClause(&JoinWhereClause{SyntaxNode: node})
		case "ScanOperator":
			return visitor.VisitScanOperator(&ScanOperator{SyntaxNode: node})
		case "AssertSchemaOperator":
			return visitor.VisitAssertSchemaOperator(&AssertSchemaOperator{SyntaxNode: node})
		case "ScanDeclareClause":
			return visitor.VisitScanDeclareClause(&ScanDeclareClause{SyntaxNode: node})
		case "ScanOrderByClause":
			return visitor.VisitScanOrderByClause(&ScanOrderByClause{SyntaxNode: node})
		case "ScanPartitionByClause":
			return visitor.VisitScanPartitionByClause(&ScanPartitionByClause{SyntaxNode: node})
		case "ScanStepOutput":
			return visitor.VisitScanStepOutput(&ScanStepOutput{SyntaxNode: node})
		case "ScanStep":
			return visitor.VisitScanStep(&ScanStep{SyntaxNode: node})
		case "ScanComputationClause":
			return visitor.VisitScanComputationClause(&ScanComputationClause{SyntaxNode: node})
		case "ScanAssignment":
			return visitor.VisitScanAssignment(&ScanAssignment{SyntaxNode: node})
		case "PartitionByOperator":
			return visitor.VisitPartitionByOperator(&PartitionByOperator{SyntaxNode: node})
		case "PartitionByIdClause":
			return visitor.VisitPartitionByIdClause(&PartitionByIdClause{SyntaxNode: node})
		case "CommandWithValueClause":
			return visitor.VisitCommandWithValueClause(&CommandWithValueClause{SyntaxNode: node})
		case "CommandWithPropertyListClause":
			return visitor.VisitCommandWithPropertyListClause(&CommandWithPropertyListClause{SyntaxNode: node})
		case "UnknownCommand":
			return visitor.VisitUnknownCommand(&UnknownCommand{SyntaxNode: node})
		case "CustomCommand":
			return visitor.VisitCustomCommand(&CustomCommand{SyntaxNode: node})
		case "PartialCommand":
			return visitor.VisitPartialCommand(&PartialCommand{SyntaxNode: node})
		case "CommandAndSkippedTokens":
			return visitor.VisitCommandAndSkippedTokens(&CommandAndSkippedTokens{SyntaxNode: node})
		case "BadCommand":
			return visitor.VisitBadCommand(&BadCommand{SyntaxNode: node})
		case "CommandBlock":
			return visitor.VisitCommandBlock(&CommandBlock{SyntaxNode: node})
		case "GraphMergeOperator":
			return visitor.VisitGraphMergeOperator(&GraphMergeOperator{SyntaxNode: node})
		}
		switch node.Kind() {
		case SyntaxKindQueryBlock:
			return visitor.VisitQueryBlock(&QueryBlock{SyntaxNode: node})
		case SyntaxKindDirectiveBlock:
			return visitor.VisitDirectiveBlock(&DirectiveBlock{SyntaxNode: node})
		case SyntaxKindDirective:
			return visitor.VisitDirective(&Directive{SyntaxNode: node})
		case SyntaxKindSkippedTokens:
			return visitor.VisitSkippedTokens(&SkippedTokens{SyntaxNode: node})
		case SyntaxKindTypeOfLiteralExpression:
			return visitor.VisitTypeOfLiteralExpression(&TypeOfLiteralExpression{SyntaxNode: node})
		case SyntaxKindBadQueryOperator:
			return visitor.VisitBadQueryOperator(&BadQueryOperator{SyntaxNode: node})
		case SyntaxKindCompoundStringLiteralExpression:
			return visitor.VisitCompoundStringLiteralExpression(&CompoundStringLiteralExpression{SyntaxNode: node})
		case SyntaxKindTokenName:
			return visitor.VisitTokenName(&TokenName{SyntaxNode: node})
		case SyntaxKindBracketedName:
			return visitor.VisitBracketedName(&BracketedName{SyntaxNode: node})
		case SyntaxKindBracedName:
			return visitor.VisitBracedName(&BracedName{SyntaxNode: node})
		case SyntaxKindWildcardedName:
			return visitor.VisitWildcardedName(&WildcardedName{SyntaxNode: node})
		case SyntaxKindBracketedWildcardedName:
			return visitor.VisitBracketedWildcardedName(&BracketedWildcardedName{SyntaxNode: node})
		case SyntaxKindNameDeclaration:
			return visitor.VisitNameDeclaration(&NameDeclaration{SyntaxNode: node})
		case SyntaxKindNameReference:
			return visitor.VisitNameReference(&NameReference{SyntaxNode: node})
		case SyntaxKindStarExpression:
			return visitor.VisitStarExpression(&StarExpression{SyntaxNode: node})
		case SyntaxKindAtExpression:
			return visitor.VisitAtExpression(&AtExpression{SyntaxNode: node})
		case SyntaxKindJsonPair:
			return visitor.VisitJsonPair(&JsonPair{SyntaxNode: node})
		case SyntaxKindJsonObjectExpression:
			return visitor.VisitJsonObjectExpression(&JsonObjectExpression{SyntaxNode: node})
		case SyntaxKindJsonArrayExpression:
			return visitor.VisitJsonArrayExpression(&JsonArrayExpression{SyntaxNode: node})
		case SyntaxKindDynamicExpression:
			return visitor.VisitDynamicExpression(&DynamicExpression{SyntaxNode: node})
		case SyntaxKindParenthesizedExpression:
			return visitor.VisitParenthesizedExpression(&ParenthesizedExpression{SyntaxNode: node})
		case SyntaxKindExpressionList:
			return visitor.VisitExpressionList(&ExpressionList{SyntaxNode: node})
		case SyntaxKindExpressionCouple:
			return visitor.VisitExpressionCouple(&ExpressionCouple{SyntaxNode: node})
		case SyntaxKindFunctionCallExpression:
			return visitor.VisitFunctionCallExpression(&FunctionCallExpression{SyntaxNode: node})
		case SyntaxKindToScalarExpression:
			return visitor.VisitToScalarExpression(&ToScalarExpression{SyntaxNode: node})
		case SyntaxKindToTableExpression:
			return visitor.VisitToTableExpression(&ToTableExpression{SyntaxNode: node})
		case SyntaxKindMaterializedViewCombineExpression:
			return visitor.VisitMaterializedViewCombineExpression(&MaterializedViewCombineExpression{SyntaxNode: node})
		case SyntaxKindMaterializedViewCombineClause:
			return visitor.VisitMaterializedViewCombineClause(&MaterializedViewCombineClause{SyntaxNode: node})
		case SyntaxKindSimpleNamedExpression:
			return visitor.VisitSimpleNamedExpression(&SimpleNamedExpression{SyntaxNode: node})
		case SyntaxKindRenameList:
			return visitor.VisitRenameList(&RenameList{SyntaxNode: node})
		case SyntaxKindCompoundNamedExpression:
			return visitor.VisitCompoundNamedExpression(&CompoundNamedExpression{SyntaxNode: node})
		case SyntaxKindBracketedExpression:
			return visitor.VisitBracketedExpression(&BracketedExpression{SyntaxNode: node})
		case SyntaxKindPathExpression:
			return visitor.VisitPathExpression(&PathExpression{SyntaxNode: node})
		case SyntaxKindElementExpression:
			return visitor.VisitElementExpression(&ElementExpression{SyntaxNode: node})
		case SyntaxKindPipeExpression:
			return visitor.VisitPipeExpression(&PipeExpression{SyntaxNode: node})
		case SyntaxKindRangeOperator:
			return visitor.VisitRangeOperator(&RangeOperator{SyntaxNode: node})
		case SyntaxKindNamedParameter:
			return visitor.VisitNamedParameter(&NamedParameter{SyntaxNode: node})
		case SyntaxKindConsumeOperator:
			return visitor.VisitConsumeOperator(&ConsumeOperator{SyntaxNode: node})
		case SyntaxKindCountOperator:
			return visitor.VisitCountOperator(&CountOperator{SyntaxNode: node})
		case SyntaxKindCountAsIdentifierClause:
			return visitor.VisitCountAsIdentifierClause(&CountAsIdentifierClause{SyntaxNode: node})
		case SyntaxKindExecuteAndCacheOperator:
			return visitor.VisitExecuteAndCacheOperator(&ExecuteAndCacheOperator{SyntaxNode: node})
		case SyntaxKindExtendOperator:
			return visitor.VisitExtendOperator(&ExtendOperator{SyntaxNode: node})
		case SyntaxKindFacetOperator:
			return visitor.VisitFacetOperator(&FacetOperator{SyntaxNode: node})
		case SyntaxKindFacetWithOperatorClause:
			return visitor.VisitFacetWithOperatorClause(&FacetWithOperatorClause{SyntaxNode: node})
		case SyntaxKindFacetWithExpressionClause:
			return visitor.VisitFacetWithExpressionClause(&FacetWithExpressionClause{SyntaxNode: node})
		case SyntaxKindFilterOperator:
			return visitor.VisitFilterOperator(&FilterOperator{SyntaxNode: node})
		case SyntaxKindGetSchemaOperator:
			return visitor.VisitGetSchemaOperator(&GetSchemaOperator{SyntaxNode: node})
		case SyntaxKindFindOperator:
			return visitor.VisitFindOperator(&FindOperator{SyntaxNode: node})
		case SyntaxKindDataScopeClause:
			return visitor.VisitDataScopeClause(&DataScopeClause{SyntaxNode: node})
		case SyntaxKindTypedColumnReference:
			return visitor.VisitTypedColumnReference(&TypedColumnReference{SyntaxNode: node})
		case SyntaxKindFindInClause:
			return visitor.VisitFindInClause(&FindInClause{SyntaxNode: node})
		case SyntaxKindFindProjectClause:
			return visitor.VisitFindProjectClause(&FindProjectClause{SyntaxNode: node})
		case SyntaxKindPackExpression:
			return visitor.VisitPackExpression(&PackExpression{SyntaxNode: node})
		case SyntaxKindNameAndTypeDeclaration:
			return visitor.VisitNameAndTypeDeclaration(&NameAndTypeDeclaration{SyntaxNode: node})
		case SyntaxKindPrimitiveTypeExpression:
			return visitor.VisitPrimitiveTypeExpression(&PrimitiveTypeExpression{SyntaxNode: node})
		case SyntaxKindSearchOperator:
			return visitor.VisitSearchOperator(&SearchOperator{SyntaxNode: node})
		case SyntaxKindForkOperator:
			return visitor.VisitForkOperator(&ForkOperator{SyntaxNode: node})
		case SyntaxKindNameEqualsClause:
			return visitor.VisitNameEqualsClause(&NameEqualsClause{SyntaxNode: node})
		case SyntaxKindForkExpression:
			return visitor.VisitForkExpression(&ForkExpression{SyntaxNode: node})
		case SyntaxKindMakeSeriesOperator:
			return visitor.VisitMakeSeriesOperator(&MakeSeriesOperator{SyntaxNode: node})
		case SyntaxKindMakeSeriesExpression:
			return visitor.VisitMakeSeriesExpression(&MakeSeriesExpression{SyntaxNode: node})
		case SyntaxKindDefaultExpressionClause:
			return visitor.VisitDefaultExpressionClause(&DefaultExpressionClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesOnClause:
			return visitor.VisitMakeSeriesOnClause(&MakeSeriesOnClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesFromClause:
			return visitor.VisitMakeSeriesFromClause(&MakeSeriesFromClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesToClause:
			return visitor.VisitMakeSeriesToClause(&MakeSeriesToClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesStepClause:
			return visitor.VisitMakeSeriesStepClause(&MakeSeriesStepClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesInRangeClause:
			return visitor.VisitMakeSeriesInRangeClause(&MakeSeriesInRangeClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesFromToStepClause:
			return visitor.VisitMakeSeriesFromToStepClause(&MakeSeriesFromToStepClause{SyntaxNode: node})
		case SyntaxKindMakeSeriesByClause:
			return visitor.VisitMakeSeriesByClause(&MakeSeriesByClause{SyntaxNode: node})
		case SyntaxKindMvExpandOperator:
			return visitor.VisitMvExpandOperator(&MvExpandOperator{SyntaxNode: node})
		case SyntaxKindMvExpandExpression:
			return visitor.VisitMvExpandExpression(&MvExpandExpression{SyntaxNode: node})
		case SyntaxKindMvExpandRowLimitClause:
			return visitor.VisitMvExpandRowLimitClause(&MvExpandRowLimitClause{SyntaxNode: node})
		case SyntaxKindMvApplyOperator:
			return visitor.VisitMvApplyOperator(&MvApplyOperator{SyntaxNode: node})
		case SyntaxKindMvApplyExpression:
			return visitor.VisitMvApplyExpression(&MvApplyExpression{SyntaxNode: node})
		case SyntaxKindMvApplyRowLimitClause:
			return visitor.VisitMvApplyRowLimitClause(&MvApplyRowLimitClause{SyntaxNode: node})
		case SyntaxKindMvApplyContextIdClause:
			return visitor.VisitMvApplyContextIdClause(&MvApplyContextIdClause{SyntaxNode: node})
		case SyntaxKindMvApplySubqueryExpression:
			return visitor.VisitMvApplySubqueryExpression(&MvApplySubqueryExpression{SyntaxNode: node})
		case SyntaxKindToTypeOfClause:
			return visitor.VisitToTypeOfClause(&ToTypeOfClause{SyntaxNode: node})
		case SyntaxKindEvaluateSchemaClause:
			return visitor.VisitEvaluateSchemaClause(&EvaluateSchemaClause{SyntaxNode: node})
		case SyntaxKindEvaluateOperator:
			return visitor.VisitEvaluateOperator(&EvaluateOperator{SyntaxNode: node})
		case SyntaxKindParseOperator:
			return visitor.VisitParseOperator(&ParseOperator{SyntaxNode: node})
		case SyntaxKindParseWhereOperator:
			return visitor.VisitParseWhereOperator(&ParseWhereOperator{SyntaxNode: node})
		case SyntaxKindParseKvWithClause:
			return visitor.VisitParseKvWithClause(&ParseKvWithClause{SyntaxNode: node})
		case SyntaxKindParseKvOperator:
			return visitor.VisitParseKvOperator(&ParseKvOperator{SyntaxNode: node})
		case SyntaxKindPartitionOperator:
			return visitor.VisitPartitionOperator(&PartitionOperator{SyntaxNode: node})
		case SyntaxKindPartitionQuery:
			return visitor.VisitPartitionQuery(&PartitionQuery{SyntaxNode: node})
		case SyntaxKindPartitionScope:
			return visitor.VisitPartitionScope(&PartitionScope{SyntaxNode: node})
		case SyntaxKindPartitionSubquery:
			return visitor.VisitPartitionSubquery(&PartitionSubquery{SyntaxNode: node})
		case SyntaxKindProjectOperator:
			return visitor.VisitProjectOperator(&ProjectOperator{SyntaxNode: node})
		case SyntaxKindProjectAwayOperator:
			return visitor.VisitProjectAwayOperator(&ProjectAwayOperator{SyntaxNode: node})
		case SyntaxKindProjectByNamesOperator:
			return visitor.VisitProjectByNamesOperator(&ProjectByNamesOperator{SyntaxNode: node})
		case SyntaxKindProjectKeepOperator:
			return visitor.VisitProjectKeepOperator(&ProjectKeepOperator{SyntaxNode: node})
		case SyntaxKindProjectRenameOperator:
			return visitor.VisitProjectRenameOperator(&ProjectRenameOperator{SyntaxNode: node})
		case SyntaxKindProjectReorderOperator:
			return visitor.VisitProjectReorderOperator(&ProjectReorderOperator{SyntaxNode: node})
		case SyntaxKindSampleOperator:
			return visitor.VisitSampleOperator(&SampleOperator{SyntaxNode: node})
		case SyntaxKindSampleDistinctOperator:
			return visitor.VisitSampleDistinctOperator(&SampleDistinctOperator{SyntaxNode: node})
		case SyntaxKindEntityGroupExpression:
			return visitor.VisitEntityGroup(&EntityGroup{SyntaxNode: node})
		case SyntaxKindReduceByOperator:
			return visitor.VisitReduceByOperator(&ReduceByOperator{SyntaxNode: node})
		case SyntaxKindReduceByWithClause:
			return visitor.VisitReduceByWithClause(&ReduceByWithClause{SyntaxNode: node})
		case SyntaxKindSummarizeOperator:
			return visitor.VisitSummarizeOperator(&SummarizeOperator{SyntaxNode: node})
		case SyntaxKindMacroExpandScopeReferenceName:
			return visitor.VisitMacroExpandScopeReferenceName(&MacroExpandScopeReferenceName{SyntaxNode: node})
		case SyntaxKindMacroExpandOperator:
			return visitor.VisitMacroExpandOperator(&MacroExpandOperator{SyntaxNode: node})
		case SyntaxKindSummarizeByClause:
			return visitor.VisitSummarizeByClause(&SummarizeByClause{SyntaxNode: node})
		case SyntaxKindDistinctOperator:
			return visitor.VisitDistinctOperator(&DistinctOperator{SyntaxNode: node})
		case SyntaxKindTakeOperator:
			return visitor.VisitTakeOperator(&TakeOperator{SyntaxNode: node})
		case SyntaxKindSortOperator:
			return visitor.VisitSortOperator(&SortOperator{SyntaxNode: node})
		case SyntaxKindOrderedExpression:
			return visitor.VisitOrderedExpression(&OrderedExpression{SyntaxNode: node})
		case SyntaxKindOrderingClause:
			return visitor.VisitOrderingClause(&OrderingClause{SyntaxNode: node})
		case SyntaxKindOrderingNullsClause:
			return visitor.VisitOrderingNullsClause(&OrderingNullsClause{SyntaxNode: node})
		case SyntaxKindTopHittersOperator:
			return visitor.VisitTopHittersOperator(&TopHittersOperator{SyntaxNode: node})
		case SyntaxKindTopHittersByClause:
			return visitor.VisitTopHittersByClause(&TopHittersByClause{SyntaxNode: node})
		case SyntaxKindTopOperator:
			return visitor.VisitTopOperator(&TopOperator{SyntaxNode: node})
		case SyntaxKindTopNestedOperator:
			return visitor.VisitTopNestedOperator(&TopNestedOperator{SyntaxNode: node})
		case SyntaxKindTopNestedClause:
			return visitor.VisitTopNestedClause(&TopNestedClause{SyntaxNode: node})
		case SyntaxKindTopNestedWithOthersClause:
			return visitor.VisitTopNestedWithOthersClause(&TopNestedWithOthersClause{SyntaxNode: node})
		case SyntaxKindUnionOperator:
			return visitor.VisitUnionOperator(&UnionOperator{SyntaxNode: node})
		case SyntaxKindAsOperator:
			return visitor.VisitAsOperator(&AsOperator{SyntaxNode: node})
		case SyntaxKindSerializeOperator:
			return visitor.VisitSerializeOperator(&SerializeOperator{SyntaxNode: node})
		case SyntaxKindInvokeOperator:
			return visitor.VisitInvokeOperator(&InvokeOperator{SyntaxNode: node})
		case SyntaxKindRenderOperator:
			return visitor.VisitRenderOperator(&RenderOperator{SyntaxNode: node})
		case SyntaxKindMakeGraphOperator:
			return visitor.VisitMakeGraphOperator(&MakeGraphOperator{SyntaxNode: node})
		case SyntaxKindMakeGraphWithTablesAndKeysClause:
			return visitor.VisitMakeGraphWithTablesAndKeysClause(&MakeGraphWithTablesAndKeysClause{SyntaxNode: node})
		case SyntaxKindMakeGraphWithImplicitIdClause:
			return visitor.VisitMakeGraphWithImplicitIdClause(&MakeGraphWithImplicitIdClause{SyntaxNode: node})
		case SyntaxKindGraphMarkComponentsOperator:
			return visitor.VisitGraphMarkComponentsOperator(&GraphMarkComponentsOperator{SyntaxNode: node})
		case SyntaxKindGraphWhereNodesOperator:
			return visitor.VisitGraphWhereNodesOperator(&GraphWhereNodesOperator{SyntaxNode: node})
		case SyntaxKindGraphWhereEdgesOperator:
			return visitor.VisitGraphWhereEdgesOperator(&GraphWhereEdgesOperator{SyntaxNode: node})
		case SyntaxKindMakeGraphTableAndKeyClause:
			return visitor.VisitMakeGraphTableAndKeyClause(&MakeGraphTableAndKeyClause{SyntaxNode: node})
		case SyntaxKindMakeGraphPartitionedByClause:
			return visitor.VisitMakeGraphPartitionedByClause(&MakeGraphPartitionedByClause{SyntaxNode: node})
		case SyntaxKindGraphToTableOperator:
			return visitor.VisitGraphToTableOperator(&GraphToTableOperator{SyntaxNode: node})
		case SyntaxKindGraphToTableOutputClause:
			return visitor.VisitGraphToTableOutputClause(&GraphToTableOutputClause{SyntaxNode: node})
		case SyntaxKindGraphToTableAsClause:
			return visitor.VisitGraphToTableAsClause(&GraphToTableAsClause{SyntaxNode: node})
		case SyntaxKindGraphMatchOperator:
			return visitor.VisitGraphMatchOperator(&GraphMatchOperator{SyntaxNode: node})
		case SyntaxKindGraphShortestPathsOperator:
			return visitor.VisitGraphShortestPathsOperator(&GraphShortestPathsOperator{SyntaxNode: node})
		case SyntaxKindGraphMatchPattern:
			return visitor.VisitGraphMatchPattern(&GraphMatchPattern{SyntaxNode: node})
		case SyntaxKindGraphMatchPatternNode:
			return visitor.VisitGraphMatchPatternNode(&GraphMatchPatternNode{SyntaxNode: node})
		case SyntaxKindGraphMatchPatternEdge:
			return visitor.VisitGraphMatchPatternEdge(&GraphMatchPatternEdge{SyntaxNode: node})
		case SyntaxKindGraphMatchPatternEdgeRange:
			return visitor.VisitGraphMatchPatternEdgeRange(&GraphMatchPatternEdgeRange{SyntaxNode: node})
		case SyntaxKindWhereClause:
			return visitor.VisitWhereClause(&WhereClause{SyntaxNode: node})
		case SyntaxKindProjectClause:
			return visitor.VisitProjectClause(&ProjectClause{SyntaxNode: node})
		case SyntaxKindNameReferenceList:
			return visitor.VisitNameReferenceList(&NameReferenceList{SyntaxNode: node})
		case SyntaxKindRenderWithClause:
			return visitor.VisitRenderWithClause(&RenderWithClause{SyntaxNode: node})
		case SyntaxKindPrintOperator:
			return visitor.VisitPrintOperator(&PrintOperator{SyntaxNode: node})
		case SyntaxKindAliasStatement:
			return visitor.VisitAliasStatement(&AliasStatement{SyntaxNode: node})
		case SyntaxKindLetStatement:
			return visitor.VisitLetStatement(&LetStatement{SyntaxNode: node})
		case SyntaxKindFunctionDeclaration:
			return visitor.VisitFunctionDeclaration(&FunctionDeclaration{SyntaxNode: node})
		case SyntaxKindFunctionParameters:
			return visitor.VisitFunctionParameters(&FunctionParameters{SyntaxNode: node})
		case SyntaxKindFunctionParameter:
			return visitor.VisitFunctionParameter(&FunctionParameter{SyntaxNode: node})
		case SyntaxKindDefaultValueDeclaration:
			return visitor.VisitDefaultValueDeclaration(&DefaultValueDeclaration{SyntaxNode: node})
		case SyntaxKindFunctionBody:
			return visitor.VisitFunctionBody(&FunctionBody{SyntaxNode: node})
		case SyntaxKindSchemaTypeExpression:
			return visitor.VisitSchemaTypeExpression(&SchemaTypeExpression{SyntaxNode: node})
		case SyntaxKindExpressionStatement:
			return visitor.VisitExpressionStatement(&ExpressionStatement{SyntaxNode: node})
		case SyntaxKindMaterializeExpression:
			return visitor.VisitMaterializeExpression(&MaterializeExpression{SyntaxNode: node})
		case SyntaxKindSetOptionStatement:
			return visitor.VisitSetOptionStatement(&SetOptionStatement{SyntaxNode: node})
		case SyntaxKindOptionValueClause:
			return visitor.VisitOptionValueClause(&OptionValueClause{SyntaxNode: node})
		case SyntaxKindQueryParametersStatement:
			return visitor.VisitQueryParametersStatement(&QueryParametersStatement{SyntaxNode: node})
		case SyntaxKindRestrictStatementWithClause:
			return visitor.VisitRestrictStatementWithClause(&RestrictStatementWithClause{SyntaxNode: node})
		case SyntaxKindRestrictStatement:
			return visitor.VisitRestrictStatement(&RestrictStatement{SyntaxNode: node})
		case SyntaxKindPatternStatement:
			return visitor.VisitPatternStatement(&PatternStatement{SyntaxNode: node})
		case SyntaxKindPatternDeclaration:
			return visitor.VisitPatternDeclaration(&PatternDeclaration{SyntaxNode: node})
		case SyntaxKindPatternPathParameter:
			return visitor.VisitPatternPathParameter(&PatternPathParameter{SyntaxNode: node})
		case SyntaxKindPatternMatch:
			return visitor.VisitPatternMatch(&PatternMatch{SyntaxNode: node})
		case SyntaxKindPatternPathValue:
			return visitor.VisitPatternPathValue(&PatternPathValue{SyntaxNode: node})
		case SyntaxKindDataScopeExpression:
			return visitor.VisitDataScopeExpression(&DataScopeExpression{SyntaxNode: node})
		case SyntaxKindDataTableExpression:
			return visitor.VisitDataTableExpression(&DataTableExpression{SyntaxNode: node})
		case SyntaxKindRowSchema:
			return visitor.VisitRowSchema(&RowSchema{SyntaxNode: node})
		case SyntaxKindEvaluateRowSchema:
			return visitor.VisitEvaluateRowSchema(&EvaluateRowSchema{SyntaxNode: node})
		case SyntaxKindExternalDataExpression:
			return visitor.VisitExternalDataExpression(&ExternalDataExpression{SyntaxNode: node})
		case SyntaxKindContextualDataTableExpression:
			return visitor.VisitContextualDataTableExpression(&ContextualDataTableExpression{SyntaxNode: node})
		case SyntaxKindExternalDataWithClause:
			return visitor.VisitExternalDataWithClause(&ExternalDataWithClause{SyntaxNode: node})
		case SyntaxKindInlineExternalTableKindClause:
			return visitor.VisitInlineExternalTableKindClause(&InlineExternalTableKindClause{SyntaxNode: node})
		case SyntaxKindInlineExternalTableDataFormatClause:
			return visitor.VisitInlineExternalTableDataFormatClause(&InlineExternalTableDataFormatClause{SyntaxNode: node})
		case SyntaxKindDateTimePattern:
			return visitor.VisitDateTimePattern(&DateTimePattern{SyntaxNode: node})
		case SyntaxKindInlineExternalTablePathFormatPartitionColumnReference:
			return visitor.VisitInlineExternalTablePathFormatPartitionColumnReference(&InlineExternalTablePathFormatPartitionColumnReference{SyntaxNode: node})
		case SyntaxKindInlineExternalTablePathFormatClause:
			return visitor.VisitInlineExternalTablePathFormatClause(&InlineExternalTablePathFormatClause{SyntaxNode: node})
		case SyntaxKindPartitionColumnDeclaration:
			return visitor.VisitPartitionColumnDeclaration(&PartitionColumnDeclaration{SyntaxNode: node})
		case SyntaxKindInlineExternalTablePartitionClause:
			return visitor.VisitInlineExternalTablePartitionClause(&InlineExternalTablePartitionClause{SyntaxNode: node})
		case SyntaxKindInlineExternalTableConnectionStringsClause:
			return visitor.VisitInlineExternalTableConnectionStringsClause(&InlineExternalTableConnectionStringsClause{SyntaxNode: node})
		case SyntaxKindInlineExternalTableExpression:
			return visitor.VisitInlineExternalTableExpression(&InlineExternalTableExpression{SyntaxNode: node})
		case SyntaxKindJoinOperator:
			return visitor.VisitJoinOperator(&JoinOperator{SyntaxNode: node})
		case SyntaxKindLookupOperator:
			return visitor.VisitLookupOperator(&LookupOperator{SyntaxNode: node})
		case SyntaxKindJoinOnClause:
			return visitor.VisitJoinOnClause(&JoinOnClause{SyntaxNode: node})
		case SyntaxKindJoinWhereClause:
			return visitor.VisitJoinWhereClause(&JoinWhereClause{SyntaxNode: node})
		case SyntaxKindScanOperator:
			return visitor.VisitScanOperator(&ScanOperator{SyntaxNode: node})
		case SyntaxKindAssertSchemaOperator:
			return visitor.VisitAssertSchemaOperator(&AssertSchemaOperator{SyntaxNode: node})
		case SyntaxKindScanDeclareClause:
			return visitor.VisitScanDeclareClause(&ScanDeclareClause{SyntaxNode: node})
		case SyntaxKindScanOrderByClause:
			return visitor.VisitScanOrderByClause(&ScanOrderByClause{SyntaxNode: node})
		case SyntaxKindScanPartitionByClause:
			return visitor.VisitScanPartitionByClause(&ScanPartitionByClause{SyntaxNode: node})
		case SyntaxKindScanStepOutput:
			return visitor.VisitScanStepOutput(&ScanStepOutput{SyntaxNode: node})
		case SyntaxKindScanStep:
			return visitor.VisitScanStep(&ScanStep{SyntaxNode: node})
		case SyntaxKindScanComputationClause:
			return visitor.VisitScanComputationClause(&ScanComputationClause{SyntaxNode: node})
		case SyntaxKindScanAssignment:
			return visitor.VisitScanAssignment(&ScanAssignment{SyntaxNode: node})
		case SyntaxKindPartitionByOperator:
			return visitor.VisitPartitionByOperator(&PartitionByOperator{SyntaxNode: node})
		case SyntaxKindPartitionByIdClause:
			return visitor.VisitPartitionByIdClause(&PartitionByIdClause{SyntaxNode: node})
		case SyntaxKindCommandWithValueClause:
			return visitor.VisitCommandWithValueClause(&CommandWithValueClause{SyntaxNode: node})
		case SyntaxKindCommandWithPropertyListClause:
			return visitor.VisitCommandWithPropertyListClause(&CommandWithPropertyListClause{SyntaxNode: node})
		case SyntaxKindUnknownCommand:
			return visitor.VisitUnknownCommand(&UnknownCommand{SyntaxNode: node})
		case SyntaxKindCustomCommand:
			return visitor.VisitCustomCommand(&CustomCommand{SyntaxNode: node})
		case SyntaxKindPartialCommand:
			return visitor.VisitPartialCommand(&PartialCommand{SyntaxNode: node})
		case SyntaxKindCommandAndSkippedTokens:
			return visitor.VisitCommandAndSkippedTokens(&CommandAndSkippedTokens{SyntaxNode: node})
		case SyntaxKindBadCommand:
			return visitor.VisitBadCommand(&BadCommand{SyntaxNode: node})
		case SyntaxKindCommandBlock:
			return visitor.VisitCommandBlock(&CommandBlock{SyntaxNode: node})
		case SyntaxKindGraphMergeOperator:
			return visitor.VisitGraphMergeOperator(&GraphMergeOperator{SyntaxNode: node})
		}
	}
	return zero
}
