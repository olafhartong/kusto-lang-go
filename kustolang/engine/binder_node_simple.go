package kustolang

import (
	"strings"
)

// Ported from the corresponding bounded methods in Binder_NodeBinder.cs.

const recoveredTabularStatementSourcePayload = "RecoveredTabularStatementSource"

type simpleNodeBinder struct {
	binder                  *Binder
	projection              *simpleProjectionBinder
	pipedAsRowScope         *TableSymbol
	joinRightRowScope       *TableSymbol
	lookupRightRowScope     *TableSymbol
	commonColumnsOnly       bool
	inPipe                  bool
	schemaCannotBeEmpty     bool
	defaultColumnNameSuffix int
}

func newSimpleNodeBinder(binder *Binder) *simpleNodeBinder {
	nodeBinder := &simpleNodeBinder{
		binder:                  binder,
		defaultColumnNameSuffix: 1,
	}
	nodeBinder.projection = newSimpleProjectionBinder(binder, &nodeBinder.defaultColumnNameSuffix, &nodeBinder.inPipe)
	return nodeBinder
}

func (n *simpleNodeBinder) bind(node *SyntaxNode) *SemanticInfo {
	switch syntax := WrapGeneratedSyntaxNode(node).(type) {
	case *Directive:
		return n.binder.bindDirective(syntax)
	case *CustomCommand:
		command := n.binder.Globals.GetCommand(syntax.CommandKind())
		if command == nil {
			return nil
		}
		var result *TypeSymbol
		if table := command.ResultType(); table != nil {
			result = &table.TypeSymbol
		}
		return semanticInfoPtr(NewSemanticInfoForSymbol(command, result))
	case *NameReference:
		if recovered, ok := syntax.SyntaxNode.Payload(recoveredTabularStatementSourcePayload); ok && recovered == true {
			return semanticInfoPtr(n.binder.ResolveNameReferenceInRowScope(n.binder.rowScope, syntax, true))
		}
		if recovered, ok := syntax.SyntaxNode.Payload(declarationRootPipeOnlyRecoveryPayload); ok && recovered == true {
			diagnostic := n.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownItem(NameReferenceSimpleName(syntax), ""), syntax)
			return semanticInfoPtr(NewSemanticInfoForType(ScalarTypes.Unknown.Type(), diagnostic))
		}
		switch syntax.Name().(type) {
		case *WildcardedName, *BracketedWildcardedName:
			return semanticInfoPtr(n.bindWildcardedNameReference(syntax))
		}
		if n.binder != nil && n.binder.scopeKind == ScopeKindPlugIn {
			return semanticInfoPtr(n.binder.resolvePlugInNameReference(n.binder.rowScope, syntax))
		}
		if n.binder != nil && n.binder.aggregateArgumentScope {
			return semanticInfoPtr(n.binder.resolveAggregateNameReference(n.binder.rowScope, syntax))
		}
		if n.joinRightRowScope != nil {
			return semanticInfoPtr(n.binder.ResolveNameReferenceInRowScopes(
				n.binder.rowScope, n.joinRightRowScope, n.commonColumnsOnly, syntax, false))
		}
		return semanticInfoPtr(n.binder.ResolveNameReferenceInRowScope(n.binder.rowScope, syntax, syntaxElementInTabularContext(syntax)))
	case *LiteralExpression:
		return n.bindLiteral(syntax)
	case *TypeOfLiteralExpression:
		return n.bindTypeOfLiteral(syntax)
	case *CompoundStringLiteralExpression:
		return semanticInfoPtr(NewSemanticInfoForType(ScalarTypes.String.Type()).WithIsConstant(true))
	case *StarExpression:
		return n.bindStarExpression(syntax)
	case *TypedColumnReference:
		return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(syntax.Column())))
	case *DynamicExpression:
		return n.bindDynamicExpression(syntax)
	case *JsonArrayExpression:
		return n.bindJsonArrayExpression(syntax)
	case *JsonObjectExpression:
		return n.bindJsonObjectExpression(syntax)
	case *JsonPair:
		return nil
	case *PrimitiveTypeExpression:
		return n.bindPrimitiveType(syntax)
	case *SchemaTypeExpression:
		return n.bindSchemaType(syntax)
	case *NameAndTypeDeclaration:
		return semanticInfoPtr(NewSemanticInfoForSymbol(Void, &Void.TypeSymbol))
	case *FunctionParameter:
		return n.bindFunctionParameter(syntax)
	case *PatternDeclaration:
		return n.bindPatternDeclaration(syntax)
	case *QueryParametersStatement:
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	case *SetOptionStatement:
		return n.bindSetOptionStatement(syntax)
	case *OptionValueClause:
		return semanticInfoPtr(NewSemanticInfoWithDiagnostics())
	case *DataTableExpression:
		return n.bindDataTable(syntax)
	case *ContextualDataTableExpression:
		return n.bindContextualDataTable(syntax)
	case *ExternalDataExpression:
		return n.bindExternalData(syntax)
	case *InlineExternalTableExpression:
		return n.bindInlineExternalTable(syntax)
	case *MaterializeExpression:
		return n.bindMaterializeExpression(syntax)
	case *MaterializedViewCombineExpression:
		return n.bindMaterializedViewCombine(syntax)
	case *MaterializedViewCombineNameClause:
		return n.bindMaterializedViewCombineNameClause(syntax)
	case *MaterializedViewCombineClause:
		return nil
	case *EntityGroup:
		return n.bindEntityGroup(syntax)
	case *MacroExpandOperator:
		return n.bindMacroExpand(syntax)
	case *DateTimePattern:
		return n.bindDateTimePattern(syntax)
	case *ParenthesizedExpression:
		return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(syntax.Expression())))
	case *BracketedExpression:
		return semanticInfoPtr(n.binder.bindUnqualifiedBracketedExpression(syntax))
	case *PathExpression:
		return semanticInfoPtr(pathElementWrapperInfo(syntax.Selector()))
	case *ElementExpression:
		return semanticInfoPtr(pathElementWrapperInfo(syntax.Selector()))
	case *DataScopeExpression:
		return n.bindDataScopeExpression(syntax)
	case *ToScalarExpression:
		return n.bindToScalarExpression(syntax)
	case *ToTableExpression:
		return n.bindToTableExpression(syntax)
	case *BinaryExpression:
		return n.bindBinary(syntax)
	case *PrefixUnaryExpression:
		return n.bindPrefixUnary(syntax)
	case *InExpression:
		return n.bindIn(syntax)
	case *HasAnyExpression:
		return n.bindHasAny(syntax)
	case *HasAllExpression:
		return n.bindHasAll(syntax)
	case *BetweenExpression:
		return n.bindBetween(syntax)
	case *FunctionCallExpression:
		return n.bindFunctionCall(syntax)
	case *InvokeOperator:
		return n.bindInvoke(syntax)
	case *SimpleNamedExpression:
		return n.bindSimpleNamed(syntax)
	case *CompoundNamedExpression:
		return n.bindCompoundNamed(syntax)
	case *OrderedExpression:
		return n.bindOrdered(syntax)
	case *PipeExpression:
		operatorInfo := boundSemanticInfoForExpression(syntax.Operator())
		if operatorInfo == nil {
			return nil
		}
		pipeInfo := *operatorInfo
		pipeInfo.Diagnostics = []Diagnostic{}
		return &pipeInfo
	case *BadQueryOperator:
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(Error))
	case *ForkExpression:
		return n.bindForkExpression(syntax)
	case *PrintOperator:
		return n.projection.bindPrint(syntax, n.inPipe)
	case *RangeOperator:
		return n.bindRange(syntax)
	case *UnionOperator:
		return n.bindUnion(syntax)
	case *SearchOperator:
		return n.bindSearch(syntax)
	case *FindOperator:
		return n.bindFind(syntax)
	case *RenderOperator:
		return n.bindRender(syntax)
	case *ReduceByOperator:
		return n.bindReduceBy(syntax)
	case *ForkOperator:
		return n.bindFork(syntax)
	case *FacetOperator:
		return n.bindFacet(syntax)
	case *PartitionOperator:
		return n.bindPartition(syntax)
	case *PartitionByOperator:
		return n.bindPartitionBy(syntax)
	case *PartitionQuery:
		return n.bindPartitionOperand(syntax.Query())
	case *PartitionSubquery:
		return n.bindPartitionOperand(syntax.Subquery())
	case *EvaluateOperator:
		return n.bindEvaluate(syntax)
	case *ParseOperator:
		return n.bindParse(syntax)
	case *ParseWhereOperator:
		return n.bindParseWhere(syntax)
	case *ParseKvOperator:
		return n.bindParseKv(syntax)
	case *MakeGraphOperator:
		return n.bindMakeGraph(syntax)
	case *MakeGraphPartitionedByClause:
		return n.bindMakeGraphPartitionedByClause(syntax)
	case *MakeSeriesOperator:
		return n.bindMakeSeries(syntax)
	case *MakeSeriesExpression:
		return n.bindMakeSeriesExpression(syntax)
	case *GraphMatchOperator:
		return n.bindGraphMatch(syntax, nil)
	case *GraphMergeOperator:
		return n.bindGraphMerge(syntax, nil, expressionGraphResult(syntax.Graph()))
	case *GraphShortestPathsOperator:
		return n.bindGraphShortestPaths(syntax, nil)
	case *GraphMarkComponentsOperator:
		return n.bindGraphMarkComponents(syntax, nil)
	case *GraphWhereNodesOperator:
		return n.bindGraphWhereNodes(syntax, nil)
	case *GraphWhereEdgesOperator:
		return n.bindGraphWhereEdges(syntax, nil)
	case *GraphToTableOperator:
		return n.bindGraphToTable(syntax, nil)
	case *JoinOperator:
		return n.bindJoin(syntax)
	case *LookupOperator:
		return n.bindLookup(syntax)
	case *ExecuteAndCacheOperator:
		return n.bindExecuteAndCache(syntax)
	case *ConsumeOperator:
		return n.bindConsume(syntax)
	case *MvExpandOperator:
		return n.projection.bindMvExpand(syntax, n.inPipe)
	case *MvApplyOperator:
		return n.bindMvApply(syntax)
	case *MvApplyExpression:
		return n.bindMvApplyExpression(syntax)
	case *MvApplySubqueryExpression:
		return n.bindMvApplySubqueryExpression(syntax)
	case *AsOperator:
		return n.bindAs(syntax)
	case *CountOperator:
		return n.bindCount(syntax)
	case *TakeOperator:
		return n.bindTake(syntax)
	case *SampleOperator:
		return n.bindSample(syntax)
	case *SampleDistinctOperator:
		return n.bindSampleDistinct(syntax)
	case *TopOperator:
		return n.bindTop(syntax)
	case *TopHittersOperator:
		return n.bindTopHitters(syntax)
	case *SortOperator:
		return n.bindSort(syntax)
	case *SerializeOperator:
		return n.bindSerialize(syntax)
	case *FilterOperator:
		return n.bindFilter(syntax)
	case *ProjectOperator:
		return n.projection.bindProject(syntax)
	case *ProjectAwayOperator:
		return n.projection.bindProjectAway(syntax)
	case *ProjectKeepOperator:
		return n.projection.bindProjectKeep(syntax)
	case *ProjectReorderOperator:
		return n.projection.bindProjectReorder(syntax)
	case *ProjectByNamesOperator:
		return n.projection.bindProjectByNames(syntax)
	case *ExtendOperator:
		return n.projection.bindExtend(syntax)
	case *ProjectRenameOperator:
		return n.projection.bindProjectRename(syntax)
	case *DistinctOperator:
		return n.projection.bindDistinct(syntax)
	case *SummarizeOperator:
		return n.projection.bindSummarize(syntax)
	case *TopNestedOperator:
		return n.projection.bindTopNested(syntax)
	case *AssertSchemaOperator:
		return n.bindAssertSchema(syntax)
	case *GetSchemaOperator:
		return n.bindGetSchema(syntax)
	case *ScanOperator:
		return n.bindScan(syntax)
	case *ExpressionStatement:
		return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(syntax.Expression())))
	default:
		return nil
	}
}

// Ported from Binder_NodeBinder.VisitWildcardedNameReference. Wildcards are
// legal under union and restrict statements and preserve candidate order in
// the resulting group.
func (n *simpleNodeBinder) bindWildcardedNameReference(node *NameReference) SemanticInfo {
	pattern := NameReferenceSimpleName(node)
	restrict := GetFirstAncestor[*RestrictStatement](node)
	if GetFirstAncestor[*UnionOperator](node) == nil && restrict == nil {
		return NewSemanticInfoForReferencedSymbol(Error, n.binder.diagnosticAtSyntaxElement(
			DiagnosticWildcardedNamesNotAllowedInThisContext(), node))
	}
	var databases []*DatabaseSymbol
	switch scope := n.binder.pathScope.(type) {
	case *DatabaseSymbol:
		databases = append(databases, scope)
	case *GroupSymbol:
		for _, member := range scope.Members() {
			if database, ok := member.(*DatabaseSymbol); ok {
				databases = append(databases, database)
			} else {
				databases = nil
				break
			}
		}
	default:
		if n.binder.currentDatabase != nil {
			databases = append(databases, n.binder.currentDatabase)
		}
	}

	var matches []Symbol
	for _, database := range databases {
		for _, table := range database.Tables() {
			if KustoFactsMatches(pattern, table.SymbolName()) {
				matches = append(matches, table)
			}
		}
	}
	if n.binder.pathScope == nil && n.binder.LocalScope != nil {
		var localViews []Symbol
		n.binder.LocalScope.GetSymbols("", SymbolMatchView, &localViews)
		for _, candidate := range localViews {
			view, ok := candidate.(*FunctionSymbol)
			if ok && view.IsView() && view.MinArgumentCount() == 0 && KustoFactsMatches(pattern, view.SymbolName()) {
				matches = append(matches, view)
			}
		}
	}
	if len(matches) == 1 {
		return NewSemanticInfoForReferencedSymbol(matches[0])
	}
	if len(matches) > 1 {
		return NewSemanticInfoForReferencedSymbol(NewGroupSymbol(matches...))
	}
	diagnostic := DiagnosticNameDoesNotReferToAnyKnownItem(pattern, "table")
	return NewSemanticInfoForReferencedSymbol(Error, n.binder.diagnosticAtSyntaxElement(diagnostic, node))
}

// Ported from Binder_NodeBinder.cs VisitStarExpression.
func (n *simpleNodeBinder) bindStarExpression(node *StarExpression) *SemanticInfo {
	if starIsFunctionArgument(node) {
		columns := n.binder.declaredAndInferredColumns(n.binder.rowScope)
		members := make([]Symbol, 0, len(columns))
		for _, column := range columns {
			members = append(members, column)
		}
		group := NewGroupSymbol(members...)
		tuple := NewTupleSymbol(columns)
		return semanticInfoPtr(NewSemanticInfoForSymbol(group, &tuple.TypeSymbol))
	}
	if parent := node.Parent(); parent != nil {
		if binary, ok := WrapGeneratedSyntaxElement(parent).(*BinaryExpression); ok && extractSameElement(binary.Left(), node) {
			return semanticInfoPtr(NewSemanticInfoForType(ScalarTypes.Dynamic.Type()))
		}
	}
	return nil
}

func starIsFunctionArgument(node *StarExpression) bool {
	if node == nil || node.Parent() == nil {
		return false
	}
	separated, ok := node.Parent().(*SeparatedElement)
	if !ok || separated.Parent() == nil {
		return false
	}
	list, ok := separated.Parent().(*SyntaxList)
	if !ok || list.Parent() == nil {
		return false
	}
	expressionList, ok := WrapGeneratedSyntaxElement(list.Parent()).(*ExpressionList)
	if !ok || expressionList.Parent() == nil {
		return false
	}
	_, ok = WrapGeneratedSyntaxElement(expressionList.Parent()).(*FunctionCallExpression)
	return ok
}

// Ported from Binder_NodeBinder.cs VisitTypeOfLiteralExpression.
func (n *simpleNodeBinder) bindTypeOfLiteral(node *TypeOfLiteralExpression) *SemanticInfo {
	diagnostics := make([]Diagnostic, 0)
	types := node.Types()
	if types != nil && types.Count() == 1 {
		if separated := types.At(0); separated != nil {
			if primitive, ok := separated.Element().(*PrimitiveTypeExpression); ok {
				symbol := schemaPrimitiveTypeSymbol(primitive)
				if symbol == Error && len(primitive.ContainedDiagnostics()) == 0 {
					name := ""
					if primitive.Type() != nil {
						name = primitive.Type().ValueText()
					}
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidTypeName(name), primitive.Type()))
				}
				return semanticInfoPtr(NewSemanticInfoForSymbol(symbol, ScalarTypes.TypeLit.Type(), diagnostics...).WithIsConstant(true))
			}
		}
	}

	columns := make([]*ColumnSymbol, 0)
	if types != nil {
		for _, separated := range types.Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			element := separated.Element()
			switch expression := element.(type) {
			case *StarExpression:
				if n.binder.rowScope != nil {
					columns = append(columns, n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
				} else {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticNoColumnsInScope(), expression))
				}
			case *NameAndTypeDeclaration:
				typ := n.declaredSchemaColumnType(expression.Type(), &diagnostics)
				columns = append(columns, newSchemaColumnSymbol(NameDeclarationSimpleName(expression.Name()), typ))
			default:
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidColumnDeclaration(), element))
			}
		}
	}
	table := NewTableSymbol("", columns...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, ScalarTypes.TypeLit.Type(), diagnostics...).WithIsConstant(true))
}

// Ported from Binder_NodeBinder.cs VisitMaterializeExpression. Current Kusto
// permits graph results; the compatibility implementation binder required a tabular result.
func (n *simpleNodeBinder) bindMaterializeExpression(node *MaterializeExpression) *SemanticInfo {
	expression := node.Expression()
	var diagnostic *Diagnostic
	if graph := expressionGraphResult(expression); graph != nil {
		info := NewSemanticInfoForReferencedSymbol(graph)
		if diagnostic != nil {
			info = info.WithDiagnostics(*diagnostic)
		}
		return semanticInfoPtr(info)
	}
	if table := expressionTableResult(expression); table != nil {
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(table))
	}

	typ := resultTypeOrError(expression)
	if typ != &Error.TypeSymbol {
		d := DiagnosticTableOrGraphExpected()
		d = n.binder.diagnosticAtSyntaxElement(d, expression)
		diagnostic = &d
	}
	if diagnostic != nil {
		return semanticInfoPtr(NewSemanticInfoForType(typ, *diagnostic))
	}
	return semanticInfoPtr(NewSemanticInfoForType(typ))
}

// Ported from Binder_NodeBinder.cs VisitDataScopeExpression. The diagnostic
// text changed when the upstream tabular check was generalized after the compatibility baseline.
func (n *simpleNodeBinder) bindDataScopeExpression(node *DataScopeExpression) *SemanticInfo {
	expression := node.Expression()
	if table := expressionTableResult(expression); table != nil {
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(table))
	}

	typ := resultTypeOrError(expression)
	var diagnostics []Diagnostic
	if typ != &Error.TypeSymbol {
		diagnostic := DiagnosticTabularValueExpected()
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(diagnostic, expression))
	}
	if graph := expressionGraphResult(expression); graph != nil {
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(graph, diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForType(typ, diagnostics...))
}

func (n *simpleNodeBinder) bindPrimitiveType(node *PrimitiveTypeExpression) *SemanticInfo {
	diagnostics := make([]Diagnostic, 0, 1)
	symbol := schemaPrimitiveTypeSymbol(node)
	if symbol == Error && node != nil && node.Type() != nil && node.Type().Width() != 0 &&
		!syntaxElementIsMissing(node) && len(node.ContainedDiagnostics()) == 0 {
		name := ""
		if node.Type() != nil {
			name = node.Type().ValueText()
		}
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidTypeName(name), node.Type()))
	}
	return semanticInfoPtr(NewSemanticInfoForSymbol(symbol, &Void.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindSchemaType(node *SchemaTypeExpression) *SemanticInfo {
	diagnostics := make([]Diagnostic, 0)
	table := n.declaredSchemaType(node, &diagnostics)
	if table == nil {
		return nil
	}
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &Void.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) declaredSchemaType(node *SchemaTypeExpression, diagnostics *[]Diagnostic) *TableSymbol {
	if node == nil || node.Columns() == nil {
		return nil
	}
	columnsSyntax := node.Columns()
	if n.schemaCannotBeEmpty && columnsSyntax.Count() == 0 {
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnDeclarationExpected(), node))
	}
	if columnsSyntax.Count() == 1 {
		if separated := columnsSyntax.At(0); separated != nil {
			if _, ok := separated.Element().(*StarExpression); ok {
				return NewTableSymbol("")
			}
		}
	}

	columns := make([]*ColumnSymbol, 0, columnsSyntax.Count())
	for _, separated := range columnsSyntax.Elements() {
		if separated == nil {
			return nil
		}
		expression := separated.Element()
		if expression == nil || syntaxElementIsMissing(expression) {
			continue
		}
		declaration, ok := expression.(*NameAndTypeDeclaration)
		if !ok {
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidColumnDeclaration(), expression))
			continue
		}

		typ := n.declaredSchemaColumnType(declaration.Type(), diagnostics)
		column := newSchemaColumnSymbol(NameDeclarationSimpleName(declaration.Name()), typ)
		columns = append(columns, column)
		declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
	return NewTableSymbol("", columns...)
}

func (n *simpleNodeBinder) declaredSchemaColumnType(expression TypeExpression, diagnostics *[]Diagnostic) *TypeSymbol {
	switch expression := expression.(type) {
	case *PrimitiveTypeExpression:
		if info, ok := expression.GetSemanticInfo(); ok {
			if symbol, ok := info.ReferencedSymbol().(*TypeSymbol); ok {
				return symbol
			}
			if info.ReferencedSymbol() == Error {
				return &Error.TypeSymbol
			}
		}
		symbol := schemaPrimitiveTypeSymbol(expression)
		if typ, ok := symbol.(*TypeSymbol); ok {
			return typ
		}
		if len(expression.ContainedDiagnostics()) == 0 {
			name := ""
			if expression.Type() != nil {
				name = expression.Type().ValueText()
			}
			*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidTypeName(name), expression.Type()))
		}
		return &Error.TypeSymbol
	case *SchemaTypeExpression:
		if info, ok := expression.GetSemanticInfo(); ok {
			if table, ok := info.ReferencedSymbol().(*TableSymbol); ok {
				return &table.TypeSymbol
			}
		}
		table := n.declaredSchemaType(expression, diagnostics)
		if table != nil {
			return &table.TypeSymbol
		}
		return &Error.TypeSymbol
	default:
		*diagnostics = append(*diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidTypeExpression(), expression))
		return &Error.TypeSymbol
	}
}

func schemaPrimitiveTypeSymbol(expression *PrimitiveTypeExpression) Symbol {
	if expression == nil || expression.Type() == nil || expression.Type().Width() == 0 {
		return Error
	}
	scalar := ScalarTypes.Get(expression.Type().RawText())
	if scalar == nil {
		return Error
	}
	return scalar.Type()
}

func newSchemaColumnSymbol(name string, typ *TypeSymbol) *ColumnSymbol {
	return &ColumnSymbol{BaseSymbol: NewBaseSymbol(name), Type: typ}
}

func (n *simpleNodeBinder) mvExpandDiagnostics(node *MvExpandOperator, inPipe bool) []Diagnostic {
	var diagnostics []Diagnostic
	if !inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.MvExpandKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.MvExpandParameters)...)
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || separated.Element() == nil {
			continue
		}
		expression := separated.Element().Expression()
		typ := ExpressionResultType(mvExpandUnderlyingExpression(expression))
		if topDiagnosticTypeKnown(typ) && !isDynamicType(typ) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.Dynamic.SymbolName()), expression))
		}
	}
	if limit := node.RowLimitClause(); limit != nil {
		diagnostics = append(diagnostics, n.integerDiagnostics(limit.RowLimit())...)
	}
	return diagnostics
}

func (n *simpleNodeBinder) bindUnion(node *UnionOperator) *SemanticInfo {
	if node == nil || node.Expressions() == nil {
		return nil
	}
	diagnostics := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.UnionParameters)
	tables := make([]*TableSymbol, 0, node.Expressions().Count()+1)
	if n.inPipe && n.binder.rowScope != nil {
		tables = append(tables, n.binder.rowScope)
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil {
			continue
		}
		expression := separated.Element()
		before := len(tables)
		addUnionExpressionTables(expression, &tables)
		if len(tables) != before {
			continue
		}
		if !syntaxElementIsMissing(expression) && resultTypeOrError(expression) != &Error.TypeSymbol {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), expression))
		}
	}

	unifiedTable := n.binder.getTableOfColumnsUnifiedByNameAndType(tables)
	columns := unifiedTable.Columns()
	if sourceName := unionWithSourceName(node); sourceName != "" {
		columns = append([]*ColumnSymbol{NewColumnSymbol(sourceName, ScalarTypes.String.Type())}, columns...)
	}
	result := NewTableSymbol("", columns...).WithIsOpen(unifiedTable.IsOpen()).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func addUnionExpressionTables(expression Expression, tables *[]*TableSymbol) {
	if expression == nil || tables == nil {
		return
	}
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			break
		}
		expression = parenthesized.Expression()
	}
	if table := unionExpressionTable(expression); table != nil {
		*tables = append(*tables, table)
		return
	}
	addUnionSymbolTables(expressionReferencedSymbol(expression), tables)
}

func addUnionSymbolTables(symbol Symbol, tables *[]*TableSymbol) {
	switch symbol := symbol.(type) {
	case *GroupSymbol:
		for _, member := range symbol.Members() {
			addUnionSymbolTables(member, tables)
		}
	case *VariableSymbol:
		addUnionSymbolTables(symbol.Type, tables)
	case *ParameterSymbol:
		addUnionSymbolTables(symbol.Type, tables)
	case *EntityGroupElementSymbol:
		addUnionSymbolTables(symbol.UnderlyingSymbol, tables)
	default:
		if table, ok := tableSymbolFromSymbol(symbol); ok {
			*tables = append(*tables, table)
		}
	}
}

func (n *simpleNodeBinder) bindJoin(node *JoinOperator) *SemanticInfo {
	if node == nil || node.Parameters() == nil || node.Expression() == nil {
		return nil
	}

	diagnostics := make([]Diagnostic, 0)
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.JoinKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.JoinParameters)...)
	if unionExpressionTable(node.Expression()) == nil && resultTypeOrError(node.Expression()) != &Error.TypeSymbol {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), node.Expression()))
	}

	switch clause := node.ConditionClause().(type) {
	case *JoinOnClause:
		for _, separated := range clause.Expressions().Elements() {
			if separated != nil {
				diagnostics = append(diagnostics, n.joinConditionDiagnostics(separated.Element(), n.binder.rowScope, n.joinRightRowScope, nil)...)
			}
		}
	case *JoinWhereClause:
		if typ := ExpressionResultType(clause.Expression()); topDiagnosticTypeKnown(typ) && !sameType(typ, ScalarTypes.Bool.Type()) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.Bool.SymbolName()), clause.Expression()))
		}
	case nil:
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMissingJoinOnClause(), node))
	}

	joinKind := joinKindParameterValue(node)
	includeLeft := !joinIsRightAntiOrSemi(joinKind)
	includeRight := !joinIsLeftAntiOrSemi(joinKind)
	columns := make([]*ColumnSymbol, 0)
	open := false
	if includeLeft && n.binder.rowScope != nil {
		columns = append(columns, n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
		open = open || n.binder.rowScope.IsOpen()
	}
	if includeRight && n.joinRightRowScope != nil {
		columns = append(columns, n.binder.declaredAndInferredColumns(n.joinRightRowScope)...)
		open = open || n.joinRightRowScope.IsOpen()
	}
	result := NewTableSymbol("", makeColumnNamesUnique(columns)...).WithIsOpen(open)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindLookup(node *LookupOperator) *SemanticInfo {
	if node == nil || node.Parameters() == nil || node.Expression() == nil || node.LookupClause() == nil {
		return nil
	}

	diagnostics := make([]Diagnostic, 0)
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.LookupKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.LookupParameters)...)
	if unionExpressionTable(node.Expression()) == nil && resultTypeOrError(node.Expression()) != &Error.TypeSymbol {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), node.Expression()))
	}

	clause, _ := node.LookupClause().(*JoinOnClause)
	joinColumns := make([]joinColumnPair, 0)
	if clause != nil {
		for _, separated := range clause.Expressions().Elements() {
			if separated != nil {
				diagnostics = append(diagnostics, n.joinConditionDiagnostics(separated.Element(), n.binder.rowScope, n.lookupRightRowScope, &joinColumns)...)
			}
		}
	}

	columns := NewProjectionBuilder()
	for _, column := range n.binder.declaredAndInferredColumns(n.binder.rowScope) {
		if pair, ok := firstJoinColumnPairForLeft(joinColumns, column); ok {
			columns.Add(commonColumnSymbol(pair.left, pair.right))
		} else {
			columns.Add(column)
		}
	}
	rightColumns := append([]*ColumnSymbol(nil), n.binder.declaredAndInferredColumns(n.lookupRightRowScope)...)
	for i := len(rightColumns) - 1; i >= 0; i-- {
		if joinColumnPairsContainRight(joinColumns, rightColumns[i]) {
			rightColumns = append(rightColumns[:i], rightColumns[i+1:]...)
		}
	}
	columns.AddRange(rightColumns)
	open := (n.binder.rowScope != nil && n.binder.rowScope.IsOpen()) ||
		(n.lookupRightRowScope != nil && n.lookupRightRowScope.IsOpen())
	result := NewTableSymbol("", columns.GetProjection()...).WithIsOpen(open)
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func removeJoinConditionParentheses(expression Expression) Expression {
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			break
		}
		expression = parenthesized.Expression()
	}
	return expression
}

type joinColumnPair struct {
	left  *ColumnSymbol
	right *ColumnSymbol
}

func firstJoinColumnPairForLeft(pairs []joinColumnPair, column *ColumnSymbol) (joinColumnPair, bool) {
	for _, pair := range pairs {
		if pair.left == column {
			return pair, true
		}
	}
	return joinColumnPair{}, false
}

func joinColumnPairsContainRight(pairs []joinColumnPair, column *ColumnSymbol) bool {
	for _, pair := range pairs {
		if pair.right == column {
			return true
		}
	}
	return false
}

func (n *simpleNodeBinder) joinConditionDiagnostics(expression Expression, leftTable, rightTable *TableSymbol, pairs *[]joinColumnPair) []Diagnostic {
	condition := removeJoinConditionParentheses(expression)
	switch condition := condition.(type) {
	case *BinaryExpression:
		switch condition.Kind() {
		case SyntaxKindEqualExpression:
			return n.joinEqualityDiagnostics(condition, pairs)
		case SyntaxKindAndExpression:
			diagnostics := n.joinConditionDiagnostics(condition.Left(), leftTable, rightTable, pairs)
			return append(diagnostics, n.joinConditionDiagnostics(condition.Right(), leftTable, rightTable, pairs)...)
		default:
			return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidJoinCondition(), condition)}
		}
	case *NameReference:
		diagnostics := n.commonColumnDiagnostics(condition, leftTable, rightTable)
		if len(diagnostics) == 0 && pairs != nil {
			left := joinColumnForTable(n.binder, leftTable, NameReferenceSimpleName(condition))
			right := joinColumnForTable(n.binder, rightTable, NameReferenceSimpleName(condition))
			if left != nil && right != nil {
				*pairs = append(*pairs, joinColumnPair{left: left, right: right})
			}
		}
		return diagnostics
	default:
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidJoinCondition(), condition)}
	}
}

func (n *simpleNodeBinder) joinEqualityDiagnostics(expression *BinaryExpression, pairs *[]joinColumnPair) []Diagnostic {
	left, leftDiagnostics := n.joinEqualityOperand(expression.Left(), "$left")
	right, rightDiagnostics := n.joinEqualityOperand(expression.Right(), "$right")
	diagnostics := append(leftDiagnostics, rightDiagnostics...)
	if left == nil || right == nil {
		return diagnostics
	}
	if comparableDiagnostics := n.joinKeyComparableDiagnostics(expression.Left(), expression.Right()); len(comparableDiagnostics) != 0 {
		return append(diagnostics, comparableDiagnostics...)
	}
	if pairs != nil {
		*pairs = append(*pairs, joinColumnPair{left: left, right: right})
	}
	return diagnostics
}

func (n *simpleNodeBinder) joinEqualityOperand(operand Expression, prefix string) (*ColumnSymbol, []Diagnostic) {
	operand = removeJoinConditionParentheses(operand)
	path, ok := operand.(*PathExpression)
	if !ok || joinConditionReferencedName(path.Expression()) != prefix {
		return nil, []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticInvalidJoinConditionOperand(prefix), operand)}
	}
	column, ok := expressionReferencedSymbol(path).(*ColumnSymbol)
	if !ok {
		// Binder.CheckIsColumn does not add KS109 when an already-bound
		// operand has an error result. Keep the missing-close recovery below
		// distinct: that malformed bracketed path is expected to produce KS109.
		if sameType(resultTypeOrError(operand), &Error.TypeSymbol) {
			if selector, bracketed := path.Selector().(*BracketedExpression); !bracketed ||
				selector.CloseBracket() == nil || selector.CloseBracket().Width() != 0 {
				return nil, nil
			}
		}
		return nil, []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), operand)}
	}
	return column, nil
}

func joinConditionReferencedName(expression Expression) string {
	reference, _ := expression.(*NameReference)
	return NameReferenceSimpleName(reference)
}

func (n *simpleNodeBinder) joinKeyComparableDiagnostics(left, right Expression) []Diagnostic {
	leftType := ExpressionResultType(left)
	rightType := ExpressionResultType(right)
	if isDynamicType(leftType) {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticJoinKeyCannotBeDynamic(), left)}
	}
	if isDynamicType(rightType) {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticJoinKeyCannotBeDynamic(), right)}
	}
	if !joinKeyTypeKnown(leftType) || !joinKeyTypeKnown(rightType) {
		return nil
	}
	if !sameType(PromoteToLong(leftType), PromoteToLong(rightType)) {
		return []Diagnostic{DiagnosticJoinKeysNotComparable(leftType.SymbolName(), rightType.SymbolName())}
	}
	return nil
}

func joinKeyTypeKnown(typ *TypeSymbol) bool {
	return typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Unknown.Type())
}

func (n *simpleNodeBinder) commonColumnDiagnostics(expression Expression, leftTable, rightTable *TableSymbol) []Diagnostic {
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			break
		}
		expression = parenthesized.Expression()
	}
	reference, ok := expression.(*NameReference)
	if !ok || syntaxElementIsMissing(reference) {
		return nil
	}
	name := NameReferenceSimpleName(reference)
	left := joinColumnForTable(n.binder, leftTable, name)
	if left == nil {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticColumnMustExistOnBothSidesOfJoin(name), reference)}
	}
	right := joinColumnForTable(n.binder, rightTable, name)
	if right == nil {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticColumnMustExistOnBothSidesOfJoin(name), reference)}
	}
	if !sameType(left.Type, right.Type) && !sameType(left.Type, ScalarTypes.Unknown.Type()) && !sameType(right.Type, ScalarTypes.Unknown.Type()) {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticCommonJoinColumnsMustHaveSameType(name), reference)}
	}
	return nil
}

func joinKindParameterValue(node *JoinOperator) string {
	if node == nil || node.Parameters() == nil {
		return ""
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || !ordinalEqualFold(NameDeclarationSimpleName(parameter.Name()), "kind") {
			continue
		}
		if value := ExpressionLiteralValueInfo(parameter.Expression()); value != nil {
			return strings.ToLower(value.ValueText)
		}
		if reference, ok := parameter.Expression().(*NameReference); ok {
			return strings.ToLower(NameReferenceSimpleName(reference))
		}
		if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
			return strings.ToLower(NameDeclarationSimpleName(declaration))
		}
	}
	return ""
}

func joinIsLeftAntiOrSemi(kind string) bool {
	switch kind {
	case "anti", "leftanti", "leftantisemi", "leftsemi":
		return true
	default:
		return false
	}
}

func joinIsRightAntiOrSemi(kind string) bool {
	switch kind {
	case "rightanti", "rightantisemi", "rightsemi":
		return true
	default:
		return false
	}
}

func unionWithSourceName(node *UnionOperator) string {
	if node == nil || node.Parameters() == nil {
		return ""
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.UnionParameters) != QueryOperatorParameters.WithSource {
			continue
		}
		if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
			return NameDeclarationSimpleName(declaration)
		}
	}
	return ""
}

func (n *simpleNodeBinder) bindRange(node *RangeOperator) *SemanticInfo {
	if node == nil || node.Name() == nil || node.From() == nil || node.To() == nil || node.Step() == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorMustBeFirst(), node.RangeToken()))
	}
	for _, expression := range []Expression{node.From(), node.To(), node.Step()} {
		typ := ExpressionResultType(expression)
		if rangeTypeKnown(typ) && !IsSummable(typ) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeSummable(), expression))
		}
	}
	if rangeExpressionIsConstantIntegerZero(node.Step()) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustNotHaveValue("0"), node.Step()))
	}

	fromType := ExpressionResultType(node.From())
	toType := ExpressionResultType(node.To())
	stepType := ExpressionResultType(node.Step())
	if syntaxElementIsMissing(node.StepToken()) && isErrorTypeSymbol(toType) {
		toType = fromType
	}
	var rangeType *TypeSymbol
	{
		fromToType := GetCommonScalarType([]*TypeSymbol{fromType, toType})
		if fromToType == nil {
			fromToType = ScalarTypes.Unknown.Type()
		}
		if sameType(fromToType, ScalarTypes.Int.Type()) && sameType(stepType, ScalarTypes.Int.Type()) {
			rangeType = ScalarTypes.Int.Type()
		} else {
			addInfo := n.operatorInfoForTypes(
				n.operatorForKind(OperatorKindAdd),
				[]Expression{node.From(), node.Step()},
				[]*TypeSymbol{fromToType, stepType},
				node.Step(),
				true,
				true,
			)
			if addInfo == nil || addInfo.ResultType == nil {
				rangeType = ScalarTypes.Unknown.Type()
			} else {
				rangeType = addInfo.ResultType
				diagnostics = append(diagnostics, addInfo.Diagnostics...)
			}
		}
	}
	column := NewColumnSymbol(NameDeclarationSimpleName(node.Name()), rangeType)
	table := NewTableSymbol("", column)
	node.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func rangeTypeKnown(typ *TypeSymbol) bool {
	return typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Unknown.Type())
}

func rangeExpressionIsConstantIntegerZero(expression Expression) bool {
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			break
		}
		expression = parenthesized.Expression()
	}
	switch value := ExpressionConstantValue(expression).(type) {
	case int:
		return value == 0
	case int32:
		return value == 0
	case int64:
		return value == 0
	default:
		return false
	}
}

func (n *simpleNodeBinder) bindScan(node *ScanOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.ScanKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.ScanParameters)...)

	columns := append([]*ColumnSymbol{}, n.binder.declaredAndInferredColumns(n.binder.rowScope)...)
	if clause := node.DeclareClause(); clause != nil {
		for _, separated := range clause.Declarations().Elements() {
			columns = append(columns, scanDeclaredColumn(separated.Element().NameAndType()))
		}
	}
	matchID := "match_id"
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || parameter.Expression() == nil {
			continue
		}
		declaration := findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.ScanParameters)
		if declaration != QueryOperatorParameters.WithMatchId {
			continue
		}
		if name, ok := parameter.Expression().(*NameDeclaration); ok && NameDeclarationSimpleName(name) != "" {
			matchID = NameDeclarationSimpleName(name)
		}
		break
	}
	columns = append(columns, NewColumnSymbol(matchID, ScalarTypes.Long.Type()))
	table := NewTableSymbol("", columns...).WithInheritableProperties(n.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

var (
	getSchemaResult = NewTableSymbol("",
		NewColumnSymbol("ColumnName", ScalarTypes.String.Type()),
		NewColumnSymbol("ColumnOrdinal", ScalarTypes.Long.Type()),
		NewColumnSymbol("DataType", ScalarTypes.String.Type()),
		NewColumnSymbol("ColumnType", ScalarTypes.String.Type()),
	)
	getSchemaCslResult = NewTableSymbol("",
		NewColumnSymbol("Schema", ScalarTypes.String.Type()),
	)
)

func (n *simpleNodeBinder) bindGetSchema(node *GetSchemaOperator) *SemanticInfo {
	table := getSchemaResult
	diagnostics := make([]Diagnostic, 0, 1)
	if getSchemaHasCslKind(node) {
		table = getSchemaCslResult
	} else if parameter := node.KindParameter(); parameter != nil && parameter.Expression() != nil {
		expression := parameter.Expression()
		typ := ExpressionResultType(expression)
		if expression.Kind() != SyntaxKindTokenLiteralExpression && topDiagnosticTypeKnown(typ) &&
			!sameType(typ, ScalarTypes.String.Type()) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), expression))
		}
	}
	// Both upstream profiles allocate CheckNotFirstInPipe diagnostics here but
	// return static SemanticInfo instances that do not contain that list.
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindExecuteAndCache(node *ExecuteAndCacheOperator) *SemanticInfo {
	if !supportsExecuteAndCacheOperator(node) {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticQueryOperatorCannotBeFirst(), node.ExecuteAndCacheKeyword()))
	}
	table := n.rowScopeResult()
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindConsume(node *ConsumeOperator) *SemanticInfo {
	if !supportsConsumeOperator(node) {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticQueryOperatorCannotBeFirst(), node.ConsumeKeyword()))
	}
	diagnostics = append(diagnostics,
		queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.ConsumeParameters)...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(Void, &Void.TypeSymbol, diagnostics...))
}

func getSchemaHasCslKind(node *GetSchemaOperator) bool {
	if node == nil || node.KindParameter() == nil {
		return false
	}
	literal, ok := node.KindParameter().Expression().(*LiteralExpression)
	return ok && literal.Kind() == SyntaxKindTokenLiteralExpression && literal.Token() != nil &&
		literal.Token().RawText() == "csl"
}

func (n *simpleNodeBinder) bindAssertSchema(node *AssertSchemaOperator) *SemanticInfo {
	if node == nil || node.AssertSchemaKeyword() == nil || node.AssertSchemaKeyword().Width() == 0 || node.Schema() == nil {
		return nil
	}
	table := n.rowScopeResult()
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
}

func (n *simpleNodeBinder) bindAs(node *AsOperator) *SemanticInfo {
	if node == nil || node.AsKeyword() == nil || node.Parameters() == nil || node.Name() == nil {
		return nil
	}

	// Upstream binds parameter names but does not include parameter validation
	// diagnostics in the AsOperator semantic result.
	_ = queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.AsParameters)
	table := n.rowScopeResult()
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
}

func (n *simpleNodeBinder) bindSimpleNamed(node *SimpleNamedExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(node.Expression())))
}

func (n *simpleNodeBinder) bindCompoundNamed(node *CompoundNamedExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(node.Expression())))
}

func (n *simpleNodeBinder) bindOrdered(node *OrderedExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	return boundSemanticInfoForExpression(node.Expression())
}

func (n *simpleNodeBinder) bindLiteral(node *LiteralExpression) *SemanticInfo {
	var result *TypeSymbol
	switch node.Kind() {
	case SyntaxKindBooleanLiteralExpression:
		result = ScalarTypes.Bool.Type()
	case SyntaxKindIntLiteralExpression:
		result = ScalarTypes.Int.Type()
	case SyntaxKindLongLiteralExpression:
		result = ScalarTypes.Long.Type()
	case SyntaxKindRealLiteralExpression:
		result = ScalarTypes.Real.Type()
	case SyntaxKindDecimalLiteralExpression:
		result = ScalarTypes.Decimal.Type()
		if n.binder != nil && n.binder.dynamicDepth != 0 {
			info := NewSemanticInfoForType(result, n.binder.diagnosticAtSyntaxElement(DiagnosticDecimalInDynamic(), node))
			return &info
		}
	case SyntaxKindDateTimeLiteralExpression:
		result = ScalarTypes.DateTime.Type()
	case SyntaxKindTimespanLiteralExpression:
		result = ScalarTypes.TimeSpan.Type()
	case SyntaxKindGuidLiteralExpression:
		result = ScalarTypes.Guid.Type()
	case SyntaxKindStringLiteralExpression:
		result = ScalarTypes.String.Type()
	case SyntaxKindNullLiteralExpression:
		result = ScalarTypes.Null.Type()
	case SyntaxKindTokenLiteralExpression:
		info := NewSemanticInfoForType(&Void.TypeSymbol)
		return &info
	default:
		return nil
	}
	info := NewSemanticInfoForType(result).WithIsConstant(true)
	return &info
}

func (n *simpleNodeBinder) bindCount(node *CountOperator) *SemanticInfo {
	diagnostics := n.notFirstDiagnostics(node)
	name := "Count"
	if clause := node.AsIdentifier(); clause != nil && clause.Identifier() != nil {
		name = clause.Identifier().ValueText()
	}
	table := NewTableSymbol("", NewColumnSymbol(name, ScalarTypes.Long.Type()))
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindTake(node *TakeOperator) *SemanticInfo {
	diagnostics := n.notFirstDiagnostics(node)
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), nil)...)
	diagnostics = append(diagnostics, n.integerDiagnostics(node.Expression())...)
	table := n.rowScopeResult().WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindSample(node *SampleOperator) *SemanticInfo {
	diagnostics := n.notFirstDiagnostics(node)
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), nil)...)
	diagnostics = append(diagnostics, n.integerDiagnostics(node.Expression())...)
	table := n.rowScopeResult().WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindSampleDistinct(node *SampleDistinctOperator) *SemanticInfo {
	if node == nil || node.SampleDistinctKeyword() == nil || node.SampleDistinctKeyword().RawText() == "" ||
		node.Parameters() == nil || node.Expression() == nil || node.OfExpression() == nil {
		return nil
	}

	diagnostics := n.notFirstDiagnostics(node)
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.SampleDistinctParameters)...)
	diagnostics = append(diagnostics, n.integerDiagnostics(node.Expression())...)

	ofType := ExpressionResultType(node.OfExpression())
	_, directColumn := expressionReferencedSymbol(node.OfExpression()).(*ColumnSymbol)
	if !directColumn {
		if ofType != nil && ofType != &Error.TypeSymbol {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), node.OfExpression()))
		}
	} else if isDynamicType(ofType) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTypeNotAllowed(ofType.SymbolName()), node.OfExpression()))
	}
	column := n.sampleDistinctColumnForExpression(node.OfExpression())

	table := NewTableSymbol("", column).WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) sampleDistinctColumnForExpression(expression Expression) *ColumnSymbol {
	name := expressionResultName(expression, "Column1", n.binder.rowScope, n.binder.Globals)
	typ := resultTypeOrError(expression)
	if source := projectionResultColumn(expression); source != nil {
		if name == "" || name == source.SymbolName() {
			return source
		}
		return NewColumnSymbolFull(name, typ, "", []*ColumnSymbol{source}, expression, nil)
	}
	column := NewColumnSymbolWithSource(name, typ, expression)
	if named, ok := expression.(*SimpleNamedExpression); ok && named.Name() != nil {
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
	return column
}

func (n *simpleNodeBinder) bindTop(node *TopOperator) *SemanticInfo {
	if node == nil {
		return nil
	}

	diagnostics := n.notFirstDiagnostics(node)
	diagnostics = append(diagnostics, n.integerDiagnostics(node.Expression())...)
	if typ := ExpressionResultType(node.ByExpression()); topDiagnosticTypeKnown(typ) && !expressionResultIsScalar(node.ByExpression()) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), node.ByExpression()))
	}
	table := n.rowScopeResult().WithIsSorted(true)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindTopHitters(node *TopHittersOperator) *SemanticInfo {
	if node == nil || node.Expression() == nil || node.OfExpression() == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.TopHittersKeyword()))
	}
	diagnostics = append(diagnostics, n.integerDiagnostics(node.Expression())...)
	if topHittersExpressionTypeKnown(node.OfExpression()) {
		if _, ok := directReferencedColumn(node.OfExpression()); !ok {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), node.OfExpression()))
		}
	}

	builder := NewProjectionBuilder()
	n.addTopHittersOfProjection(node.OfExpression(), builder, &diagnostics)
	if clause := node.ByClause(); clause != nil {
		by := clause.Expression()
		if topHittersExpressionTypeKnown(by) && !IsNumeric(ExpressionResultType(by)) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeNumeric(), by))
		}
		name := "approximate_sum_" + expressionResultName(by, "", n.binder.rowScope, n.binder.Globals)
		builder.Add(topHittersColumnForExpression(by, name, ExpressionResultType(by)))
	} else {
		name := "approximate_count_" + expressionResultName(node.OfExpression(), "", n.binder.rowScope, n.binder.Globals)
		builder.Add(topHittersColumnForExpression(node.OfExpression(), name, ScalarTypes.Long.Type()))
	}

	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(n.binder.rowScope).WithIsSorted(true)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) addTopHittersOfProjection(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic) {
	if syntaxElementIsMissing(expression) {
		return
	}
	if compound, ok := expression.(*CompoundNamedExpression); ok {
		if n.projection.addCompoundNamedExpression(compound, builder, diagnostics, compoundProjectionDefault) {
			return
		}
	}
	if named, ok := expression.(*SimpleNamedExpression); ok {
		if tuple, function := summarizeTupleResult(named.Expression()); tuple != nil {
			columns := tuple.Columns()
			if len(columns) > 0 {
				first := columns[0]
				column := NewColumnSymbolFull(NameDeclarationSimpleName(named.Name()), first.Type, "", []*ColumnSymbol{first}, first.Source, nil)
				builder.Declare(column, diagnostics, named.Name(), true)
				named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				addSummarizeTupleColumns(columns[1:], function, builder)
			}
			return
		}
		column := NewColumnSymbolWithSource(NameDeclarationSimpleName(named.Name()), resultTypeOrError(named.Expression()), named.Expression())
		builder.Declare(column, diagnostics, named.Name(), true)
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		return
	}
	if tuple, function := summarizeTupleResult(expression); tuple != nil {
		addSummarizeTupleColumns(tuple.Columns(), function, builder)
		return
	}
	if column := projectionResultColumn(expression); column != nil {
		builder.Declare(column, diagnostics, expression)
		return
	}
	name := expressionResultName(expression, "", n.binder.rowScope, n.binder.Globals)
	if name == "" {
		name = n.nextTopHittersDefaultColumnName()
	}
	builder.Add(NewColumnSymbolWithSource(name, resultTypeOrError(expression), expression))
}

func (n *simpleNodeBinder) nextTopHittersDefaultColumnName() string {
	return n.projection.nextDefaultColumnName(nil)
}

func topHittersColumnForExpression(expression Expression, name string, typ *TypeSymbol) *ColumnSymbol {
	if column := projectionResultColumn(expression); column != nil {
		if column.SymbolName() == name && sameType(column.Type, typ) {
			return column
		}
		return NewColumnSymbolFull(name, typ, "", []*ColumnSymbol{column}, expression, nil)
	}
	return NewColumnSymbolWithSource(name, typ, expression)
}

func topHittersExpressionTypeKnown(expression Expression) bool {
	typ := ExpressionResultType(expression)
	return typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Unknown.Type())
}

func topDiagnosticTypeKnown(typ *TypeSymbol) bool {
	return typ != nil && typ != &Error.TypeSymbol && !sameType(typ, ScalarTypes.Unknown.Type())
}

func (n *simpleNodeBinder) bindSort(node *SortOperator) *SemanticInfo {
	if node == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.GetChild(0)))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.SortParameters)...)

	allScalar := true
	if expressions := node.Expressions(); expressions != nil {
		for _, separated := range expressions.Elements() {
			if separated == nil || separated.Element() == nil {
				continue
			}
			expression := separated.Element()
			typ := ExpressionResultType(expression)
			if typ == nil || expressionResultIsScalar(expression) {
				continue
			}
			if typ != &Error.TypeSymbol {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
			}
			allScalar = false
			break
		}
		if allScalar {
			for _, separated := range expressions.Elements() {
				if separated == nil || separated.Element() == nil {
					continue
				}
				expression := separated.Element()
				if typ := ExpressionResultType(expression); isDynamicType(typ) {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTypeNotAllowed(typ.SymbolName()), expression))
					break
				}
			}
		}
	}
	table := n.rowScopeResult().WithIsSorted(true)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) bindSerialize(node *SerializeOperator) *SemanticInfo {
	if node == nil || node.SerializeKeyword() == nil || node.SerializeKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expressions() == nil ||
		(!n.projection.supportsSerializeExpressions(node.Expressions()) && !serializeHasDetachedScanTail(node.Expressions())) {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.SerializeKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.SerializedParameters)...)

	// CheckIsScalar reports only the first failing expression.
	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		if typ := ExpressionResultType(expression); typ != nil && typ != &Error.TypeSymbol && !expressionResultIsScalar(expression) {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
			break
		}
	}

	builder := NewProjectionBuilder()
	if n.binder.rowScope != nil {
		builder.AddRange(n.binder.declaredAndInferredColumns(n.binder.rowScope), ProjectionAddRangeOptions{DoNotRepeat: true})
	}
	n.projection.addSerializeExpressions(node.Expressions(), builder, &diagnostics)
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(n.binder.rowScope).WithIsSerialized(true)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func serializeHasDetachedScanTail(expressions *SeparatedSyntaxListOf[Expression]) bool {
	if expressions == nil || expressions.Count() == 0 {
		return false
	}
	first, ok := expressions.At(0).Element().(*NameReference)
	if !ok || first == nil {
		return false
	}
	token := first.GetFirstToken()
	return token != nil && token.Kind() == SyntaxKindScanKeyword
}

func (n *simpleNodeBinder) bindFilter(node *FilterOperator) *SemanticInfo {
	if node == nil || node.Keyword() == nil || node.Keyword().Width() == 0 || node.Parameters() == nil || node.Condition() == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if !n.inPipe {
		diagnostics = append(diagnostics,
			n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.Keyword()))
	}
	diagnostics = append(diagnostics,
		queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.FilterParameters)...)

	// Kusto.Language Binder_NodeBinder.CheckIsExactType suppresses this
	// diagnostic only for Error and Unknown, including parser-recovered nodes.
	typ := resultTypeOrError(node.Condition())
	if typ != &Error.TypeSymbol &&
		!sameType(typ, ScalarTypes.Unknown.Type()) && !sameType(typ, ScalarTypes.Bool.Type()) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustHaveType(ScalarTypes.Bool.SymbolName()), node.Condition()))
	}
	table := n.rowScopeResult()
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (n *simpleNodeBinder) notFirstDiagnostics(operator SyntaxElement) []Diagnostic {
	if !n.hasPipedInput(operator) {
		location := operator
		if operator != nil && operator.ChildCount() != 0 && operator.GetChild(0) != nil {
			location = operator.GetChild(0)
		}
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), location)}
	}
	return []Diagnostic{}
}

func (n *simpleNodeBinder) hasPipedInput(operator SyntaxElement) bool {
	queryOperator, _ := WrapGeneratedSyntaxElement(operator).(QueryOperator)
	return n.inPipe || HasPipedInput(queryOperator)
}

func (n *simpleNodeBinder) rowScopeResult() *TableSymbol {
	rowScope := n.binder.rowScope
	if rowScope == nil {
		return NewTableSymbol("")
	}
	return NewTableSymbol("", n.binder.declaredAndInferredColumns(rowScope)...).WithInheritableProperties(rowScope)
}

func (n *simpleNodeBinder) integerDiagnostics(expression Expression) []Diagnostic {
	typ := ExpressionResultType(expression)
	if !topDiagnosticTypeKnown(typ) || IsInteger(typ) {
		return []Diagnostic{}
	}
	return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeInteger(), expression)}
}

func hasSupportedExpressionResult(expression Expression) bool {
	typ := ExpressionRawResultType(expression)
	return typ != nil && typ != &Error.TypeSymbol
}

func expressionResultIsScalar(expression Expression) bool {
	if accessor, ok := expression.(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	}); ok {
		if info, ok := accessor.GetSemanticInfo(); ok && info.ReferencedSymbol() != nil {
			return IsScalar(info.ReferencedSymbol())
		}
	}
	return IsScalar(ExpressionResultType(expression))
}

func isSupportedBoundScalar(expression Expression) bool {
	if expression == nil || !hasSupportedExpressionResult(expression) || !expressionResultIsScalar(expression) {
		return false
	}
	switch expression := expression.(type) {
	case *NameReference, *LiteralExpression:
		return true
	case *ParenthesizedExpression:
		return isSupportedBoundScalar(expression.Expression())
	case *BinaryExpression:
		switch expression.Kind() {
		case SyntaxKindLessThanExpression,
			SyntaxKindLessThanOrEqualExpression,
			SyntaxKindGreaterThanExpression,
			SyntaxKindGreaterThanOrEqualExpression,
			SyntaxKindEqualExpression,
			SyntaxKindNotEqualExpression,
			SyntaxKindEqualTildeExpression,
			SyntaxKindAndExpression,
			SyntaxKindOrExpression:
			return isSupportedBoundScalar(expression.Left()) && isSupportedBoundScalar(expression.Right())
		}
	}
	return false
}

func directReferencedColumn(expression Expression) (*ColumnSymbol, bool) {
	reference, ok := expression.(*NameReference)
	if !ok {
		return nil, false
	}
	column, ok := reference.ReferencedSymbol().(*ColumnSymbol)
	return column, ok
}

func isDynamicType(typ *TypeSymbol) bool {
	return sameType(typ, ScalarTypes.Dynamic.Type()) || isDynamicTypeSymbol(typ)
}

func resultTypeOrError(expression Expression) *TypeSymbol {
	if typ := ExpressionRawResultType(expression); typ != nil {
		return typ
	}
	return &Error.TypeSymbol
}

func boundSemanticInfoForExpression(element SyntaxElement) *SemanticInfo {
	if element != nil {
		if accessor, ok := element.(interface {
			GetSemanticInfo() (SemanticInfo, bool)
		}); ok {
			if info, ok := accessor.GetSemanticInfo(); ok {
				copy := info
				return &copy
			}
		}
	}
	return nil
}

func semanticInfoPtr(info SemanticInfo) *SemanticInfo {
	return &info
}
