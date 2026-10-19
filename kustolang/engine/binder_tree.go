package kustolang

import "strings"

// Ported from Binder_TreeBinder.cs. This initial native tranche binds children
// before parents and carries the left side of a pipe as row scope while binding
// its operator.

type treeBinder struct {
	binder             *Binder
	nodeBinder         *simpleNodeBinder
	inPipe             bool
	inWildcard         bool
	scanAliases        map[string]*VariableSymbol
	summarizeByColumns []*ColumnSymbol
}

func newTreeBinder(binder *Binder) *treeBinder {
	return &treeBinder{
		binder:     binder,
		nodeBinder: newSimpleNodeBinder(binder),
	}
}

func (t *treeBinder) bind(root SyntaxElement) {
	if node := simpleBinderSyntaxNode(root); node != nil {
		t.visitNode(node)
		return
	}
	t.visitElement(root)
}

func (t *treeBinder) visitElement(element SyntaxElement) {
	if t != nil && t.binder != nil {
		t.binder.cancellationToken.ThrowIfCancellationRequested()
	}
	if node := simpleBinderSyntaxNode(element); node != nil {
		t.visitNode(node)
		return
	}
	switch element := element.(type) {
	case nil, *SyntaxToken:
		return
	case *SyntaxNode:
		t.visitNode(element)
	default:
		for i := 0; i < element.ChildCount(); i++ {
			t.visitElement(element.GetChild(i))
		}
	}
}

func (t *treeBinder) visitNode(node *SyntaxNode) {
	if node == nil {
		return
	}
	if t != nil && t.binder != nil {
		t.binder.cancellationToken.ThrowIfCancellationRequested()
	}
	if pipe, ok := WrapGeneratedSyntaxNode(node).(*PipeExpression); ok {
		t.visitPipeExpression(pipe)
		return
	}
	switch operator := WrapGeneratedSyntaxNode(node).(type) {
	case *CommandBlock:
		t.visitCommandBlock(operator)
		return
	case *NameReference:
		t.visitNameReference(operator)
		return
	case *Directive:
		t.bindNode(operator.SyntaxNode)
		return
	case *AliasStatement:
		t.visitAliasStatement(operator)
		return
	case *LetStatement:
		t.visitLetStatement(operator)
		return
	case *FunctionDeclaration:
		t.visitFunctionDeclarationScope(operator, "")
		return
	case *PatternStatement:
		t.visitPatternStatement(operator)
		return
	case *PatternDeclaration:
		t.visitPatternDeclaration(operator)
		return
	case *QueryParametersStatement:
		t.visitQueryParametersStatement(operator)
		return
	case *SetOptionStatement:
		t.visitSetOptionStatement(operator)
		return
	case *BracketedExpression:
		t.visitBracketedExpression(operator)
		return
	case *PathExpression:
		t.visitPathExpression(operator)
		return
	case *ElementExpression:
		t.visitElementExpression(operator)
		return
	case *BinaryExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *PrefixUnaryExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *InExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *HasAnyExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *HasAllExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *BetweenExpression:
		t.visitScalarOperator(operator.SyntaxNode)
		return
	case *FunctionCallExpression:
		t.visitFunctionCall(operator)
		return
	case *InvokeOperator:
		t.visitInvoke(operator)
		return
	case *ToScalarExpression:
		t.visitToScalarExpression(operator)
		return
	case *ToTableExpression:
		t.visitToTableExpression(operator)
		return
	case *DynamicExpression:
		t.visitDynamicExpression(operator)
		return
	case *DataTableExpression:
		t.visitDataTable(operator)
		return
	case *ContextualDataTableExpression:
		t.visitContextualDataTable(operator)
		return
	case *ExternalDataExpression:
		t.visitExternalData(operator)
		return
	case *InlineExternalTableExpression:
		t.visitInlineExternalTable(operator)
		return
	case *MaterializedViewCombineExpression:
		t.visitMaterializedViewCombine(operator)
		return
	case *EntityGroup:
		t.visitEntityGroup(operator)
		return
	case *MacroExpandOperator:
		t.visitMacroExpand(operator)
		return
	case *DateTimePattern:
		t.visitDateTimePattern(operator)
		return
	case *SchemaTypeExpression:
		t.visitSchemaType(operator, false)
		return
	case *MaterializedViewCombineNameClause:
		t.visitMaterializedViewCombineNameClause(operator)
		return
	case *MaterializedViewCombineClause:
		t.visitMaterializedViewCombineClause(operator)
		return
	case *NameAndTypeDeclaration:
		if schema, ok := operator.Type().(*SchemaTypeExpression); ok {
			t.visitElement(operator.Name())
			t.visitSchemaType(schema, true)
			t.bindNode(operator.SyntaxNode)
			return
		}
	case *PrintOperator:
		t.visitPrint(operator)
		return
	case *RangeOperator:
		t.visitRange(operator)
		return
	case *UnionOperator:
		t.visitUnion(operator)
		return
	case *SearchOperator:
		t.visitSearch(operator)
		return
	case *FindOperator:
		t.visitFind(operator)
		return
	case *ForkOperator:
		t.visitFork(operator)
		return
	case *FacetOperator:
		t.visitFacet(operator)
		return
	case *PartitionOperator:
		t.visitPartition(operator)
		return
	case *PartitionByOperator:
		t.visitPartitionBy(operator)
		return
	case *ParseOperator:
		t.visitParseOperator(operator)
		return
	case *ParseWhereOperator:
		t.visitParseOperator(operator)
		return
	case *ParseKvOperator:
		t.visitParseOperator(operator)
		return
	case *MakeGraphOperator:
		t.visitMakeGraph(operator)
		return
	case *MakeSeriesOperator:
		t.visitMakeSeries(operator)
		return
	case *GraphMatchOperator:
		t.visitGraphMatch(operator, nil)
		return
	case *GraphMergeOperator:
		t.visitGraphMerge(operator, nil)
		return
	case *GraphShortestPathsOperator:
		t.visitGraphShortestPaths(operator, nil)
		return
	case *GraphMarkComponentsOperator:
		t.visitGraphMarkComponents(operator, nil)
		return
	case *GraphWhereNodesOperator:
		t.visitGraphWhereNodes(operator, nil)
		return
	case *GraphWhereEdgesOperator:
		t.visitGraphWhereEdges(operator, nil)
		return
	case *GraphToTableOperator:
		t.visitGraphToTable(operator, nil)
		return
	case *JoinOperator:
		t.visitJoin(operator)
		return
	case *LookupOperator:
		t.visitLookup(operator)
		return
	case *ExecuteAndCacheOperator:
		t.visitExecuteAndCache(operator)
		return
	case *ConsumeOperator:
		t.visitConsume(operator)
		return
	case *MvExpandOperator:
		t.visitMvExpand(operator)
		return
	case *MvApplyOperator:
		t.visitMvApply(operator)
		return
	case *ProjectAwayOperator:
		t.visitProjectColumns(operator.SyntaxNode, operator.Expressions())
		return
	case *ProjectKeepOperator:
		t.visitProjectColumns(operator.SyntaxNode, operator.Expressions())
		return
	case *ProjectReorderOperator:
		t.visitProjectReorder(operator)
		return
	case *ProjectByNamesOperator:
		t.visitProjectByNames(operator)
		return
	case *DistinctOperator:
		t.visitDistinct(operator)
		return
	case *SampleDistinctOperator:
		t.visitSampleDistinct(operator)
		return
	case *SummarizeOperator:
		t.visitSummarize(operator)
		return
	case *TopNestedOperator:
		t.visitTopNested(operator)
		return
	case *FilterOperator:
		t.visitFilter(operator)
		return
	case *SortOperator:
		t.visitSort(operator)
		return
	case *TopHittersOperator:
		t.visitTopHitters(operator)
		return
	case *EvaluateOperator:
		t.visitEvaluate(operator)
		return
	case *GetSchemaOperator:
		t.visitGetSchema(operator)
		return
	case *ScanOperator:
		t.visitScan(operator)
		return
	}

	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	t.bindNode(node)
}

func (t *treeBinder) visitNameReference(node *NameReference) {
	if node == nil {
		return
	}
	if syntaxElementIsMissing(node) {
		// Kusto.Language's shared ErrorInfo stores ErrorSymbol as the result
		// type, not as a referenced symbol, for parser-synthesized names.
		node.SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
		return
	}
	t.bindNode(node.SyntaxNode)
	if len(node.SemanticDiagnostics()) != 0 {
		return
	}
	table, ok := tableSymbolFromSymbol(node.ReferencedSymbol())
	if !ok || table == nil {
		return
	}
	if nameReferenceIsCommandButNotQueryPart(node) {
		t.binder.rowScope = table
	}
}

func nameReferenceIsCommandButNotQueryPart(node SyntaxElement) bool {
	current := node
	for current != nil {
		parent := current.Parent()
		switch WrapGeneratedSyntaxElement(parent).(type) {
		case *PathExpression:
			current = parent
			continue
		}
		switch parent.(type) {
		case *SeparatedElement, *SyntaxList:
			current = parent
			continue
		}
		if _, ok := WrapGeneratedSyntaxElement(parent).(Command); ok {
			return true
		}
		_, ok := parent.(*CustomNode)
		return ok
	}
	return false
}

func (t *treeBinder) visitCommandBlock(node *CommandBlock) {
	if node == nil {
		return
	}

	t.visitElement(node.Directives())
	statements := node.Statements()
	if statements == nil || statements.Count() == 0 {
		return
	}

	firstStatement := statements.At(0).Element()
	t.visitElement(firstStatement)

	if command := firstCommandDescendant(firstStatement); command != nil {
		commandExpression, ok := command.(Expression)
		if ok {
			resultType := ExpressionResultType(commandExpression)
			if resultType == nil {
				resultType = &Error.TypeSymbol
			}
			t.binder.LocalScope.AddSymbol(NewVariableSymbol("$command_results", resultType, false, nil, commandExpression))
		}
	}

	for i := 1; i < statements.Count(); i++ {
		t.visitElement(statements.At(i).Element())
	}
}

func firstCommandDescendant(root SyntaxElement) Command {
	var command Command
	WalkSyntax(root, func(element SyntaxElement) {
		if command != nil || element == root {
			return
		}
		if candidate, ok := WrapGeneratedSyntaxElement(element).(Command); ok {
			command = candidate
		}
	})
	return command
}

func (t *treeBinder) visitAliasStatement(node *AliasStatement) {
	if node == nil {
		return
	}
	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	t.bindNode(node.SyntaxNode)

	name := NameDeclarationSimpleName(node.Name())
	database, ok := aliasDatabaseResult(node.Expression())
	if name == "" || !ok {
		return
	}
	if t.binder.aliasedDatabases == nil {
		t.binder.aliasedDatabases = map[string]*DatabaseSymbol{}
	}
	t.binder.aliasedDatabases[name] = database
}

func aliasDatabaseResult(expression Expression) (*DatabaseSymbol, bool) {
	switch expression := expression.(type) {
	case *FunctionCallExpression:
		result, found := expression.SyntaxNode.Payload(functionCallResultSymbolPayload)
		database, ok := result.(*DatabaseSymbol)
		return database, found && ok
	case *PathExpression:
		return aliasDatabaseResult(expression.Selector())
	case *ParenthesizedExpression:
		return aliasDatabaseResult(expression.Expression())
	case *BracketedExpression:
		return aliasDatabaseResult(expression.Expression())
	default:
		return nil, false
	}
}

func (t *treeBinder) visitScalarOperator(node *SyntaxNode) {
	if node == nil || node.IsBound() {
		return
	}
	if !supportsScalarOperator(node) {
		for i := 0; i < node.ChildCount(); i++ {
			t.visitElement(node.GetChild(i))
		}
		return
	}
	preserveRecoveredChildren := syntaxElementHasMissingChildren(node)
	tx := newUnionBindingTransaction(node)
	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	for _, expression := range scalarOperatorArguments(WrapGeneratedSyntaxNode(node)) {
		expressionNode := simpleBinderSyntaxNode(expression)
		if expressionNode != nil && syntaxElementIsMissing(expression) {
			expressionNode.SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
		}
		if expressionNode == nil || !expressionNode.IsBound() {
			if !preserveRecoveredChildren {
				tx.rollback()
			}
			return
		}
	}
	t.bindNode(node)
	if !node.IsBound() {
		if !preserveRecoveredChildren {
			tx.rollback()
		}
	}
}

func supportsScalarOperator(node *SyntaxNode) bool {
	if node == nil || operatorKindForSyntax(node.Kind()) == OperatorKindNone {
		return false
	}
	wrapped := WrapGeneratedSyntaxNode(node)
	var token *SyntaxToken
	switch expression := wrapped.(type) {
	case *BinaryExpression:
		token = expression.Operator()
	case *PrefixUnaryExpression:
		token = expression.Operator()
	case *InExpression:
		token = expression.Operator()
		if !completeExpressionList(expression.Right()) {
			return false
		}
	case *HasAnyExpression:
		token = expression.Operator()
		if !completeExpressionList(expression.Right()) {
			return false
		}
	case *HasAllExpression:
		token = expression.Operator()
		if !completeExpressionList(expression.Right()) {
			return false
		}
	case *BetweenExpression:
		token = expression.Operator()
		if expression.Right() == nil || expression.Right().OpenParen() == nil || expression.Right().OpenParen().Width() == 0 ||
			expression.Right().DotDot() == nil || expression.Right().DotDot().Width() == 0 ||
			!mvExpansionTokenIsPresentOrParserMissing(expression.Right().CloseParen(), SyntaxKindCloseParenToken) {
			return false
		}
	default:
		return false
	}
	if token == nil || token.Width() == 0 {
		return false
	}
	arguments := scalarOperatorArguments(wrapped)
	if len(arguments) == 0 {
		return false
	}
	_, inList := wrapped.(*InExpression)
	_, hasAnyList := wrapped.(*HasAnyExpression)
	_, hasAllList := wrapped.(*HasAllExpression)
	listOperator := inList || hasAnyList || hasAllList
	for i, argument := range arguments {
		if _, binary := wrapped.(*BinaryExpression); binary && i == 1 && isMissingExpression(argument) {
			continue
		}
		if listOperator && i > 0 && isMissingExpression(argument) {
			continue
		}
		if argument == nil || syntaxElementIsMissing(argument) {
			return false
		}
	}
	return true
}

func isMissingExpression(element SyntaxElement) bool {
	if !syntaxElementIsMissing(element) {
		return false
	}
	for _, diagnostic := range element.ContainedDiagnostics() {
		if diagnostic.Code == "KS006" {
			return true
		}
	}
	return false
}

func completeExpressionList(list *ExpressionList) bool {
	return list != nil && list.OpenParen() != nil && list.OpenParen().Width() != 0 && list.Expressions() != nil &&
		list.Expressions().Count() != 0 && mvExpansionTokenIsPresentOrParserMissing(list.CloseParen(), SyntaxKindCloseParenToken)
}

func scalarOperatorArguments(node SyntaxElement) []Expression {
	switch expression := node.(type) {
	case *BinaryExpression:
		return []Expression{expression.Left(), expression.Right()}
	case *PrefixUnaryExpression:
		return []Expression{expression.Expression()}
	case *InExpression:
		return listOperatorArguments(expression.Left(), expression.Right())
	case *HasAnyExpression:
		return listOperatorArguments(expression.Left(), expression.Right())
	case *HasAllExpression:
		return listOperatorArguments(expression.Left(), expression.Right())
	case *BetweenExpression:
		if expression.Right() != nil {
			return []Expression{expression.Left(), expression.Right().First(), expression.Right().Second()}
		}
	}
	return nil
}

func listOperatorArguments(left Expression, list *ExpressionList) []Expression {
	result := []Expression{left}
	if list != nil && list.Expressions() != nil {
		for _, separated := range list.Expressions().Elements() {
			if separated != nil {
				result = append(result, separated.Element())
			}
		}
	}
	return result
}

func (t *treeBinder) visitSchemaType(node *SchemaTypeExpression, cannotBeEmpty bool) bool {
	if node == nil || node.IsBound() {
		return node != nil
	}
	if !supportsSchemaTypeExpression(node) {
		return false
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldScanAliases := t.scanAliases
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldCommonColumnsOnly := t.nodeBinder.commonColumnsOnly
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.scanAliases = oldScanAliases
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.schemaCannotBeEmpty = oldSchemaCannotBeEmpty
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		t.nodeBinder.commonColumnsOnly = oldCommonColumnsOnly
		t.nodeBinder.lookupRightRowScope = oldLookupRightRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	for _, separated := range node.Columns().Elements() {
		if separated == nil {
			return false
		}
		expression := separated.Element()
		if expression == nil || syntaxElementIsMissing(expression) {
			continue
		}
		declaration, ok := expression.(*NameAndTypeDeclaration)
		if !ok {
			t.visitElement(expression)
			continue
		}
		switch typ := declaration.Type().(type) {
		case *PrimitiveTypeExpression:
			t.visitNode(typ.SyntaxNode)
			if !typ.IsBound() {
				return false
			}
		case *SchemaTypeExpression:
			if !t.visitSchemaType(typ, true) {
				return false
			}
		}
	}

	t.nodeBinder.schemaCannotBeEmpty = cannotBeEmpty
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return false
	}
	committed = true
	return true
}

func supportsSchemaTypeExpression(node *SchemaTypeExpression) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 3 || node.OpenParen() == nil ||
		node.Columns() == nil || node.CloseParen() == nil {
		return false
	}
	list, ok := node.GetChild(1).(*SyntaxList)
	if !ok || (list.ElementType() != "SeparatedElement<Expression>" && list.ElementType() != "SeparatedElement<SyntaxElement>") {
		return false
	}
	for i := 0; i < list.Count(); i++ {
		separated, ok := list.At(i).(*SeparatedElement)
		if !ok || separated == nil || separated.Element() == nil {
			return false
		}
		expression, ok := WrapGeneratedSyntaxElement(separated.Element()).(Expression)
		if !ok || !supportsSchemaElement(expression) {
			return false
		}
	}
	return true
}

func supportsSchemaElement(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameAndTypeDeclaration:
		if expression == nil || expression.SyntaxNode == nil || expression.ChildCount() != 3 ||
			expression.Name() == nil || expression.Colon() == nil || expression.Type() == nil ||
			!supportsSchemaNameDeclaration(expression.Name()) {
			return false
		}
		switch typ := expression.Type().(type) {
		case *PrimitiveTypeExpression:
			return typ.SyntaxNode != nil && typ.ChildCount() == 1 && typ.Type() != nil
		case *SchemaTypeExpression:
			return supportsSchemaTypeExpression(typ)
		default:
			return true
		}
	case *StarExpression:
		return expression != nil && expression.SyntaxNode != nil && expression.ChildCount() == 1 && expression.AsteriskToken() != nil
	default:
		return expression != nil
	}
}

func supportsSchemaNameDeclaration(node *NameDeclaration) bool {
	return node != nil && node.SyntaxNode != nil && node.ChildCount() == 1 && node.Name() != nil
}

func (t *treeBinder) visitExecuteAndCache(node *ExecuteAndCacheOperator) {
	if !supportsExecuteAndCacheOperator(node) || node.IsBound() {
		return
	}
	t.bindNode(node.SyntaxNode)
}

func supportsExecuteAndCacheOperator(node *ExecuteAndCacheOperator) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 1 {
		return false
	}
	keyword, ok := node.GetChild(0).(*SyntaxToken)
	return ok && keyword != nil && keyword.Kind() == SyntaxKindExecuteAndCacheKeyword && keyword.Width() != 0 &&
		node.ExecuteAndCacheKeyword() == keyword
}

func (t *treeBinder) visitConsume(node *ConsumeOperator) {
	if node == nil || node.IsBound() || !supportsConsumeOperator(node) || (t.inPipe && t.binder.rowScope == nil) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldScanAliases := t.scanAliases
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSchemaCannotBeEmpty := t.nodeBinder.schemaCannotBeEmpty
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.scanAliases = oldScanAliases
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.schemaCannotBeEmpty = oldSchemaCannotBeEmpty
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		t.nodeBinder.lookupRightRowScope = oldLookupRightRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	for _, parameter := range node.Parameters().Elements() {
		if !t.bindConsumeParameterExpression(parameter.Expression()) {
			return
		}
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func supportsConsumeOperator(node *ConsumeOperator) bool {
	if node == nil || node.SyntaxNode == nil || node.ChildCount() != 2 || node.ConsumeKeyword() == nil ||
		node.ConsumeKeyword().Kind() != SyntaxKindConsumeKeyword || node.ConsumeKeyword().Width() == 0 || node.Parameters() == nil {
		return false
	}
	list, ok := node.GetChild(1).(*SyntaxList)
	if !ok || list.ElementType() != "NamedParameter" {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.SyntaxNode == nil || parameter.ChildCount() != 3 || parameter.Name() == nil ||
			parameter.Name().Name() == nil || !unionNameIsExact(parameter.Name().Name()) || parameter.EqualToken() == nil ||
			parameter.Expression() == nil || !supportsConsumeParameterExpression(parameter.Expression()) {
			return false
		}
	}
	return true
}

func supportsConsumeParameterExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		return expression.Token() != nil
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	case *NameDeclaration:
		return expression.Name() != nil && unionNameIsExact(expression.Name())
	default:
		return false
	}
}

func (t *treeBinder) bindConsumeParameterExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *LiteralExpression:
		t.visitNode(expression.SyntaxNode)
		return expression.IsBound()
	case *NameReference:
		t.visitNode(expression.SyntaxNode)
		return expression.IsBound()
	case *NameDeclaration:
		return true
	default:
		return false
	}
}

func (t *treeBinder) visitMvExpand(node *MvExpandOperator) {
	if !supportsMvExpandOperator(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !t.bindMvExpandExpression(separated.Element()) {
			return
		}
	}
	if limit := node.RowLimitClause(); limit != nil {
		t.visitElement(limit.RowLimit())
		if bound := simpleBinderSyntaxNode(limit.RowLimit()); bound == nil ||
			(!bound.IsBound() && !syntaxElementHasMissingChildren(limit.RowLimit())) {
			return
		}
	}

	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) bindMvExpandExpression(node *MvExpandExpression) bool {
	if node == nil || !t.stageMvExpansionScalar(node.Expression()) {
		return false
	}
	if clause := node.ToTypeOf(); clause != nil {
		if typeExpression := mvExpandDeclaredTypeExpression(clause); typeExpression != nil {
			symbol := declaredSymbolFromPrimitiveTypeExpression(typeExpression)
			typeExpression.SetSemanticInfo(NewSemanticInfoForSymbol(symbol, ScalarTypes.TypeLit.Type()))
			clause.TypeOf().SetSemanticInfo(NewSemanticInfoForSymbol(symbol, ScalarTypes.TypeLit.Type()))
		} else if clause.TypeOf() != nil {
			clause.TypeOf().SetSemanticInfo(*t.nodeBinder.bindTypeOfLiteral(clause.TypeOf()))
		}
	}
	node.SetSemanticInfo(NewSemanticInfoForType(mvExpandExpressionResultType(node)))
	return true
}

func supportsMvExpandOperator(node *MvExpandOperator) bool {
	if node == nil || node.MvExpandKeyword() == nil || node.MvExpandKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expressions() == nil || node.Expressions().Count() == 0 {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || !serializeNameDeclarationIsSupported(parameter.Name()) ||
			parameter.EqualToken() == nil || parameter.Expression() == nil || !supportsMvExpandParameterExpression(parameter.Expression()) {
			return false
		}
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !supportsMvExpandExpression(separated.Element()) {
			return false
		}
	}
	if limit := node.RowLimitClause(); limit != nil {
		if limit.LimitKeyword() == nil || limit.LimitKeyword().Width() == 0 || !supportsMvExpansionScalar(limit.RowLimit()) {
			return false
		}
	}
	return true
}

func supportsMvExpandParameterExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameDeclaration:
		return expression.Name() != nil && unionNameIsExact(expression.Name())
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	case *LiteralExpression:
		return expression.Token() != nil && expression.Token().Width() != 0
	default:
		return false
	}
}

func supportsMvExpandExpression(node *MvExpandExpression) bool {
	if node == nil || !supportsMvExpansionScalar(node.Expression()) {
		return false
	}
	return true
}

func mvExpandDeclaredTypeExpression(clause *ToTypeOfClause) *PrimitiveTypeExpression {
	if clause == nil || clause.ToKeyword() == nil || clause.ToKeyword().Width() == 0 || clause.TypeOf() == nil {
		return nil
	}
	typeOf := clause.TypeOf()
	if typeOf.TypeOfKeyword() == nil || typeOf.TypeOfKeyword().Width() == 0 ||
		typeOf.OpenParen() == nil || typeOf.OpenParen().Width() == 0 ||
		!mvExpansionTokenIsPresentOrParserMissing(typeOf.CloseParen(), SyntaxKindCloseParenToken) ||
		typeOf.Types() == nil || typeOf.Types().Count() != 1 {
		return nil
	}
	separated := typeOf.Types().At(0)
	if separated == nil {
		return nil
	}
	primitive, ok := separated.Element().(*PrimitiveTypeExpression)
	if !ok || primitive.Type() == nil || primitive.Type().Width() == 0 || declaredSymbolFromPrimitiveTypeExpression(primitive) == Error {
		return nil
	}
	return primitive
}

func cloneOpenTableInferredColumns(source map[*TableSymbol]map[string]*ColumnSymbol) map[*TableSymbol]map[string]*ColumnSymbol {
	if source == nil {
		return nil
	}
	clone := make(map[*TableSymbol]map[string]*ColumnSymbol, len(source))
	for table, columns := range source {
		copy := make(map[string]*ColumnSymbol, len(columns))
		for name, column := range columns {
			copy[name] = column
		}
		clone[table] = copy
	}
	return clone
}

func cloneOpenTableInferredOrder(source map[*TableSymbol][]*ColumnSymbol) map[*TableSymbol][]*ColumnSymbol {
	if source == nil {
		return nil
	}
	clone := make(map[*TableSymbol][]*ColumnSymbol, len(source))
	for table, columns := range source {
		clone[table] = append([]*ColumnSymbol(nil), columns...)
	}
	return clone
}

type unionBindingTransaction struct {
	semanticPayloads      map[*SyntaxNode]any
	hadSemantic           map[*SyntaxNode]bool
	functionCallResults   map[*SyntaxNode]any
	hadFunctionCallResult map[*SyntaxNode]bool
}

func newUnionBindingTransaction(root SyntaxElement) *unionBindingTransaction {
	tx := &unionBindingTransaction{
		semanticPayloads:      map[*SyntaxNode]any{},
		hadSemantic:           map[*SyntaxNode]bool{},
		functionCallResults:   map[*SyntaxNode]any{},
		hadFunctionCallResult: map[*SyntaxNode]bool{},
	}
	var snapshot func(SyntaxElement)
	snapshot = func(element SyntaxElement) {
		if element == nil {
			return
		}
		if node := syntaxNodeForBinderLocation(element); node != nil {
			if value, ok := node.Payload(semanticInfoPayloadName); ok {
				tx.semanticPayloads[node] = value
				tx.hadSemantic[node] = true
			} else {
				tx.hadSemantic[node] = false
			}
			if value, ok := node.Payload(functionCallResultSymbolPayload); ok {
				tx.functionCallResults[node] = value
				tx.hadFunctionCallResult[node] = true
			} else {
				tx.hadFunctionCallResult[node] = false
			}
		}
		for i := 0; i < element.ChildCount(); i++ {
			snapshot(element.GetChild(i))
		}
	}
	snapshot(root)
	return tx
}

func (tx *unionBindingTransaction) rollback() {
	if tx == nil {
		return
	}
	for node, had := range tx.hadSemantic {
		if had {
			node.setPayload(semanticInfoPayloadName, tx.semanticPayloads[node])
		} else if node.payloads != nil {
			delete(node.payloads, semanticInfoPayloadName)
		}
		if tx.hadFunctionCallResult[node] {
			node.setPayload(functionCallResultSymbolPayload, tx.functionCallResults[node])
		} else if node.payloads != nil {
			delete(node.payloads, functionCallResultSymbolPayload)
		}
	}
}

func (t *treeBinder) visitUnion(node *UnionOperator) {
	if !supportsUnionOperator(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldIsFuzzy := t.binder.isFuzzy
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		t.binder.isFuzzy = oldIsFuzzy
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	// Union inputs are table expressions and never resolve against the piped row.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	t.binder.LocalScope = NewLocalScope(oldLocalScope)
	t.inWildcard = false
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && parameter.Expression() != nil {
			t.visitElement(parameter.Expression())
		}
	}
	if fuzzy, specified := unionFuzzyEvaluation(node); specified {
		t.binder.isFuzzy = fuzzy
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !t.bindUnionOperand(separated.Element()) {
			return
		}
	}

	// Restore the incoming table only for result construction; operand binding
	// above deliberately ran with no row scope.
	t.binder.rowScope = oldRowScope
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) bindUnionOperand(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		if syntaxElementIsMissing(expression) {
			expression.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
			return true
		}
		var info SemanticInfo
		if unionNameIsWildcard(expression.Name()) {
			info = t.nodeBinder.bindWildcardedNameReference(expression)
		} else {
			const tableNamespace = SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView |
				SymbolMatchStoredQueryResult | SymbolMatchFunction | SymbolMatchView | SymbolMatchLocal | SymbolMatchTabular
			info = t.binder.ResolveNameInRowScope(nil, NameReferenceSimpleName(expression), tableNamespace, true)
			info = t.binder.bindBareFunctionReference(expression, info)
		}
		info = t.binder.locateNameReferenceDiagnostics(info, expression)
		expression.SetSemanticInfo(info)
		return true
	case *FunctionCallExpression, *PathExpression:
		t.visitElement(expression)
		return expressionIsBound(expression)
	case *ElementExpression:
		return t.bindUnionWildcardElementOperand(expression)
	case *ParenthesizedExpression:
		if parserRecoveredEmptyParenthesizedExpression(expression) {
			t.visitElement(expression)
			return expressionIsBound(expression)
		}
		return t.bindParenthesizedTableSourceOperand(expression)
	case *BracketedExpression:
		if parserRecoveredEmptyBracketedExpression(expression) {
			return t.bindParserRecoveredEmptyBracketedExpression(expression)
		}
		t.visitElement(expression)
		return expressionIsBound(expression)
	default:
		return false
	}
}

func (t *treeBinder) bindUnionWildcardElementOperand(expression *ElementExpression) bool {
	selector, ok := expression.Selector().(*NameReference)
	if !ok || !unionNameIsWildcard(selector.Name()) {
		return false
	}
	t.visitElement(expression.Expression())
	if !expressionIsBound(expression.Expression()) {
		return false
	}

	oldPathScope := t.binder.pathScope
	t.binder.pathScope = pathScopeResult(expression.Expression())
	selector.SetSemanticInfo(t.nodeBinder.bindWildcardedNameReference(selector))
	t.binder.pathScope = oldPathScope
	expression.SetSemanticInfo(pathElementWrapperInfo(selector))
	return expression.IsBound()
}

func (t *treeBinder) visitJoin(node *JoinOperator) {
	if !t.supportsJoinOperator(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	// Parameters bind in the incoming (left) row scope.
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && parameter.Expression() != nil {
			t.visitElement(parameter.Expression())
		}
	}

	// The right table expression cannot see columns from the left input.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	t.binder.LocalScope = NewLocalScope(oldLocalScope)
	t.inWildcard = false
	if !t.bindJoinOperand(node.Expression()) {
		return
	}
	right := unionExpressionTable(node.Expression())

	// Join conditions see both sides. Bare names prefer the left row scope and
	// fall back to the right; qualified operands resolve through $left/$right.
	t.binder.rowScope = oldRowScope
	t.binder.LocalScope = oldLocalScope
	t.nodeBinder.joinRightRowScope = right
	switch clause := node.ConditionClause().(type) {
	case *JoinOnClause:
		for _, separated := range clause.Expressions().Elements() {
			if separated == nil || !t.bindJoinConditionExpression(separated.Element(), oldRowScope, right) {
				return
			}
		}
	case *JoinWhereClause:
		t.visitElement(clause.Expression())
		if !expressionIsBound(clause.Expression()) {
			return
		}
	}

	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) visitLookup(node *LookupOperator) {
	if !t.supportsLookupOperator(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldJoinRightRowScope := t.nodeBinder.joinRightRowScope
	oldLookupRightRowScope := t.nodeBinder.lookupRightRowScope
	oldCommonColumnsOnly := t.nodeBinder.commonColumnsOnly
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		t.nodeBinder.joinRightRowScope = oldJoinRightRowScope
		t.nodeBinder.lookupRightRowScope = oldLookupRightRowScope
		t.nodeBinder.commonColumnsOnly = oldCommonColumnsOnly
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	// Parameters are scalar expressions in the incoming (left) row scope.
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && parameter.Expression() != nil {
			t.visitElement(parameter.Expression())
		}
	}

	// The lookup table is resolved independently from columns on the left.
	t.binder.rowScope = nil
	t.binder.scopeKind = ScopeKindNormal
	t.binder.LocalScope = NewLocalScope(oldLocalScope)
	t.inWildcard = false
	if !t.bindJoinOperand(node.Expression()) {
		return
	}
	right := unionExpressionTable(node.Expression())

	// Lookup uses the same join-condition grammar as join. The node binder
	// records validated key pairs when constructing the output schema.
	t.binder.rowScope = oldRowScope
	t.binder.LocalScope = oldLocalScope
	t.nodeBinder.joinRightRowScope = right
	t.nodeBinder.lookupRightRowScope = right
	for _, separated := range node.LookupClause().(*JoinOnClause).Expressions().Elements() {
		if separated == nil || !t.bindJoinConditionExpression(separated.Element(), oldRowScope, right) {
			return
		}
	}

	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		return
	}
	committed = true
}

func (t *treeBinder) bindJoinOperand(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		if syntaxElementIsMissing(expression) {
			expression.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
			return true
		}
		const tableNamespace = SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView |
			SymbolMatchStoredQueryResult | SymbolMatchLocal | SymbolMatchTabular
		info := t.binder.ResolveNameInRowScope(nil, NameReferenceSimpleName(expression), tableNamespace, true)
		expression.SetSemanticInfo(t.binder.locateNameReferenceDiagnostics(info, expression))
		return true
	case *FunctionCallExpression, *PathExpression:
		t.visitElement(expression)
		return expressionIsBound(expression)
	case *DataTableExpression, *ContextualDataTableExpression:
		t.visitElement(expression)
		return expressionIsBound(expression)
	case *ParenthesizedExpression:
		return t.bindParenthesizedTableSourceOperand(expression)
	default:
		return false
	}
}

func (t *treeBinder) bindParenthesizedTableSourceOperand(expression *ParenthesizedExpression) bool {
	if !supportsParenthesizedTableSourceOperand(expression) {
		return false
	}
	t.visitElement(expression.Expression())
	if !expressionIsBound(expression.Expression()) {
		return false
	}
	if table := unionExpressionTable(expression.Expression()); table != nil {
		expression.SetSemanticInfo(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
		return true
	}
	typ := resultTypeOrError(expression.Expression())
	if typ == nil || typ == &Error.TypeSymbol {
		return false
	}
	// A bound scalar is retained so the containing operator can report KS111.
	expression.SetSemanticInfo(NewSemanticInfoForType(typ))
	return true
}

func (t *treeBinder) bindJoinCommonColumn(expression Expression, left, right *TableSymbol) bool {
	switch expression := expression.(type) {
	case *NameReference:
		if syntaxElementIsMissing(expression) {
			expression.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
			return true
		}
		leftColumn := joinColumnForTable(t.binder, left, NameReferenceSimpleName(expression))
		rightColumn := joinColumnForTable(t.binder, right, NameReferenceSimpleName(expression))
		if leftColumn != nil && rightColumn != nil {
			common := commonColumnSymbol(leftColumn, rightColumn)
			expression.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(common))
		} else {
			expression.SetSemanticInfo(t.binder.ResolveNameReferenceInRowScopes(left, right, true, expression, false))
		}
		return true
	case *ParenthesizedExpression:
		if !t.bindJoinCommonColumn(expression.Expression(), left, right) {
			return false
		}
		expression.SetSemanticInfo(NewSemanticInfoForType(resultTypeOrError(expression.Expression())))
		return true
	default:
		return false
	}
}

func (t *treeBinder) bindJoinConditionExpression(expression Expression, left, right *TableSymbol) bool {
	if expression == nil {
		return false
	}
	switch expression := expression.(type) {
	case *NameReference:
		name := NameReferenceSimpleName(expression)
		if name == "$left" || name == "$right" {
			expression.SetSemanticInfo(t.binder.ResolveNameReferenceInRowScopes(left, right, false, expression, false))
			return true
		}
		return t.bindJoinCommonColumn(expression, left, right)
	case *ParenthesizedExpression:
		if !t.bindJoinConditionExpression(expression.Expression(), left, right) {
			return false
		}
		expression.SetSemanticInfo(NewSemanticInfoForType(resultTypeOrError(expression.Expression())))
		return true
	case *BinaryExpression:
		if !t.bindJoinConditionExpression(expression.Left(), left, right) ||
			!t.bindJoinConditionExpression(expression.Right(), left, right) {
			return false
		}
		if expression.Kind() == SyntaxKindAndExpression {
			expression.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Bool.Type()))
			return true
		}
		t.bindNode(expression.SyntaxNode)
		return expression.IsBound()
	case *PathExpression:
		if t.bindMalformedJoinBracketedPath(expression) {
			return true
		}
		t.visitPathExpression(expression)
		return expression.IsBound()
	case *LiteralExpression:
		t.visitElement(expression)
		return expression.IsBound()
	default:
		oldCommonColumnsOnly := t.nodeBinder.commonColumnsOnly
		t.nodeBinder.commonColumnsOnly = true
		t.visitElement(expression)
		t.nodeBinder.commonColumnsOnly = oldCommonColumnsOnly
		return expressionIsBound(expression)
	}
}

// Binder_TreeBinder's bracketed-path branch visits the selector children but
// does not bind the PathExpression itself. Preserve that recovery shape with
// an error result so the transactional Go binder can continue to KS109.
func (t *treeBinder) bindMalformedJoinBracketedPath(path *PathExpression) bool {
	if path == nil || path.Expression() == nil || path.Dot() == nil || path.Dot().Width() == 0 {
		return false
	}
	selector, ok := path.Selector().(*BracketedExpression)
	if !ok || !bindableUnqualifiedBracketedExpression(selector) || selector.CloseBracket().Width() != 0 {
		return false
	}
	t.visitElement(path.Expression())
	if !expressionIsBound(path.Expression()) {
		return false
	}
	oldPathScope := t.binder.pathScope
	t.binder.pathScope = nil
	t.visitElement(selector.Expression())
	t.binder.pathScope = oldPathScope
	if !expressionIsBound(selector.Expression()) {
		return false
	}
	selector.SetSemanticInfo(t.binder.bindElementSelector(path.Expression(), selector))
	path.SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
	return true
}

func joinColumnForTable(binder *Binder, table *TableSymbol, name string) *ColumnSymbol {
	if table == nil || name == "" {
		return nil
	}
	for _, column := range binder.declaredAndInferredColumns(table) {
		if column != nil && column.SymbolName() == name {
			return column
		}
	}
	if table.IsOpen() {
		return binder.getOpenTableInferredColumn(name, table)
	}
	return nil
}

func (t *treeBinder) supportsJoinOperator(node *JoinOperator) bool {
	if node == nil || node.JoinKeyword() == nil || node.JoinKeyword().Width() == 0 ||
		node.Parameters() == nil || !supportsJoinOperand(node.Expression()) {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || parameter.Expression() == nil ||
			!supportsJoinParameterExpression(parameter.Expression()) {
			return false
		}
	}
	if node.ConditionClause() == nil {
		return true
	}
	switch clause := node.ConditionClause().(type) {
	case *JoinOnClause:
		if clause.OnKeyword() == nil || clause.OnKeyword().Width() == 0 || clause.Expressions() == nil {
			return false
		}
		for _, separated := range clause.Expressions().Elements() {
			if separated == nil || !t.supportsJoinConditionExpression(separated.Element()) {
				return false
			}
		}
		return true
	case *JoinWhereClause:
		return clause.WhereKeyword() != nil && clause.WhereKeyword().Width() != 0 &&
			clause.Expression() != nil && t.supportsFilterScalar(clause.Expression())
	default:
		return false
	}
}

func (t *treeBinder) supportsLookupOperator(node *LookupOperator) bool {
	if node == nil || node.LookupKeyword() == nil || node.LookupKeyword().Width() == 0 ||
		node.Parameters() == nil || !supportsJoinOperand(node.Expression()) {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || parameter.Expression() == nil ||
			!supportsJoinParameterExpression(parameter.Expression()) {
			return false
		}
	}
	clause, ok := node.LookupClause().(*JoinOnClause)
	if !ok || clause.Expressions() == nil {
		return false
	}
	// A zero-width on token with no keys is the parser's recoverable missing
	// clause. Any partially injected equivalent is rejected before traversal.
	if clause.OnKeyword() == nil || (clause.OnKeyword().Width() == 0 && clause.Expressions().Count() != 0) {
		return false
	}
	for _, separated := range clause.Expressions().Elements() {
		if separated == nil || !t.supportsJoinConditionExpression(separated.Element()) {
			return false
		}
	}
	return true
}

func supportsJoinParameterExpression(expression Expression) bool {
	if reference, ok := expression.(*NameReference); ok && syntaxElementIsMissing(reference) {
		return true
	}
	return supportsUnionParameterExpression(expression)
}

func supportsJoinOperand(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || unionNameIsExact(expression.Name())
	case *FunctionCallExpression:
		return completeUnionFunctionCall(expression)
	case *DataTableExpression, *ContextualDataTableExpression:
		return true
	case *PathExpression:
		return supportsTableSourcePath(expression)
	case *ParenthesizedExpression:
		return supportsParenthesizedTableSourceOperand(expression)
	default:
		return false
	}
}

func supportsUnionWildcardElementOperand(expression *ElementExpression) bool {
	if expression == nil || expression.Expression() == nil || expression.Selector() == nil {
		return false
	}
	selector, ok := expression.Selector().(*NameReference)
	return ok && unionNameIsWildcard(selector.Name()) && supportsTableSourcePathPart(expression.Expression())
}

func supportsJoinCommonColumn(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || unionNameIsExact(expression.Name())
	case *ParenthesizedExpression:
		return expression.OpenParen() != nil && expression.OpenParen().Width() != 0 &&
			expression.CloseParen() != nil && expression.CloseParen().Width() != 0 &&
			supportsJoinCommonColumn(expression.Expression())
	default:
		return false
	}
}

func (t *treeBinder) supportsJoinConditionExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || unionNameIsExact(expression.Name())
	case *LiteralExpression:
		return expression.Token() != nil
	case *PathExpression:
		return completePathExpression(expression)
	case *ParenthesizedExpression:
		return expression.OpenParen() != nil && expression.OpenParen().Width() != 0 &&
			expression.Expression() != nil && t.supportsJoinConditionExpression(expression.Expression())
	case *BinaryExpression:
		return expression.Operator() != nil && expression.Left() != nil && expression.Right() != nil &&
			t.supportsJoinConditionExpression(expression.Left()) && t.supportsJoinConditionExpression(expression.Right())
	default:
		return t.supportsFilterScalar(expression)
	}
}

func supportsUnionOperator(node *UnionOperator) bool {
	if node == nil || node.UnionKeyword() == nil || node.UnionKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expressions() == nil {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || parameter.Expression() == nil ||
			!supportsUnionParameterExpression(parameter.Expression()) {
			return false
		}
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !supportsUnionOperand(separated.Element()) {
			return false
		}
	}
	return true
}

func unionUsesFuzzyEvaluation(node *UnionOperator) bool {
	fuzzy, specified := unionFuzzyEvaluation(node)
	return specified && fuzzy
}

func unionFuzzyEvaluation(node *UnionOperator) (bool, bool) {
	if node == nil {
		return false, false
	}
	return queryOperatorBooleanParameterValue(node.Parameters(), "isfuzzy")
}

func queryOperatorBooleanParameterValue(parameters *SyntaxListOf[*NamedParameter], name string) (bool, bool) {
	if parameters == nil {
		return false, false
	}
	for _, parameter := range parameters.Elements() {
		if parameter == nil || !strings.EqualFold(NameDeclarationSimpleName(parameter.Name()), name) {
			continue
		}
		literal, ok := parameter.Expression().(*LiteralExpression)
		if !ok || literal.Kind() != SyntaxKindBooleanLiteralExpression || literal.Token() == nil {
			return false, false
		}
		return strings.EqualFold(literal.Token().ValueText(), "true"), true
	}
	return false, false
}

func supportsUnionParameterExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameDeclaration:
		return expression.Name() != nil && unionNameIsExact(expression.Name())
	case *LiteralExpression:
		return expression.Token() != nil
	case *NameReference:
		return isCompleteProjectByNamesReference(expression)
	default:
		return false
	}
}

func supportsUnionOperand(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || unionNameIsExact(expression.Name()) || unionNameIsWildcard(expression.Name())
	case *FunctionCallExpression:
		return completeUnionFunctionCall(expression)
	case *PathExpression:
		return supportsTableSourcePath(expression)
	case *ElementExpression:
		return supportsUnionWildcardElementOperand(expression)
	case *ParenthesizedExpression:
		return supportsParenthesizedTableSourceOperand(expression)
	case *BracketedExpression:
		return parserRecoveredEmptyBracketedExpression(expression) || bindableUnqualifiedBracketedExpression(expression)
	default:
		return false
	}
}

func parserRecoveredEmptyParenthesizedExpression(expression *ParenthesizedExpression) bool {
	return expression != nil && expression.ChildCount() == 3 &&
		unionPrintRecoveryTokenIsPresent(expression.OpenParen(), SyntaxKindOpenParenToken) &&
		parserMissingElementWithCode(expression.Expression(), "KS006") &&
		parserMissingTokenWithCode(expression.CloseParen(), SyntaxKindCloseParenToken, "KS005")
}

func parserRecoveredEmptyBracketedExpression(expression *BracketedExpression) bool {
	return expression != nil && expression.ChildCount() == 3 &&
		unionPrintRecoveryTokenIsPresent(expression.OpenBracket(), SyntaxKindOpenBracketToken) &&
		parserMissingElementWithCode(expression.Expression(), "KS006") &&
		parserMissingTokenWithCode(expression.CloseBracket(), SyntaxKindCloseBracketToken, "KS005")
}

func unionPrintRecoveryTokenIsPresent(token *SyntaxToken, kind SyntaxKind) bool {
	return token != nil && token.Kind() == kind && token.Width() != 0
}

func parserMissingElementWithCode(element SyntaxElement, code string) bool {
	if element == nil || element.Width() != 0 {
		return false
	}
	for _, diagnostic := range element.ContainedDiagnostics() {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func parserMissingTokenWithCode(token *SyntaxToken, kind SyntaxKind, code string) bool {
	return token != nil && token.Kind() == kind && parserMissingElementWithCode(token, code)
}

func unionNameIsWildcard(name Name) bool {
	switch name := name.(type) {
	case *WildcardedName:
		return name.Pattern() != nil && name.Pattern().Width() != 0
	case *BracketedWildcardedName:
		if name.OpenBracket() == nil || name.OpenBracket().Width() == 0 || name.Pattern() == nil || name.Pattern().Width() == 0 || name.CloseBracket() == nil {
			return false
		}
		return name.CloseBracket().Width() != 0 ||
			(syntaxElementIsMissing(name.CloseBracket()) && len(name.CloseBracket().Diagnostics()) != 0)
	default:
		return false
	}
}

func unionNameIsExact(name Name) bool {
	switch name := name.(type) {
	case *TokenName:
		return name.Name() != nil && name.Name().Width() != 0
	case *BracketedName:
		return name.OpenBracket() != nil && name.OpenBracket().Width() != 0 && name.Name() != nil &&
			name.Name().Width() != 0 && name.CloseBracket() != nil && name.CloseBracket().Width() != 0
	default:
		return false
	}
}

func supportsParenthesizedTableSourceOperand(expression *ParenthesizedExpression) bool {
	if expression == nil || expression.OpenParen() == nil || expression.OpenParen().Width() == 0 ||
		expression.CloseParen() == nil || expression.Expression() == nil {
		return false
	}
	close := expression.CloseParen()
	return close.Width() != 0 || syntaxElementIsMissing(close)
}

func completeUnionFunctionCall(call *FunctionCallExpression) bool {
	if call == nil || call.Name() == nil || !unionNameIsExact(call.Name().Name()) || call.ArgumentList() == nil ||
		call.ArgumentList().OpenParen() == nil || call.ArgumentList().OpenParen().Width() == 0 ||
		call.ArgumentList().CloseParen() == nil || call.ArgumentList().CloseParen().Width() == 0 ||
		call.ArgumentList().Expressions() == nil {
		return false
	}
	for _, separated := range call.ArgumentList().Expressions().Elements() {
		if separated == nil || separated.Element() == nil || syntaxElementIsMissing(separated.Element()) ||
			syntaxElementHasMissingChildren(separated.Element()) {
			return false
		}
	}
	return true
}

func supportsTableSourcePath(path *PathExpression) bool {
	if !completePathExpression(path) {
		return false
	}
	return supportsTableSourcePathPart(path.Expression()) && supportsTableSourcePathPart(path.Selector())
}

func supportsTableSourcePathPart(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		return unionNameIsExact(expression.Name()) || unionNameIsWildcard(expression.Name())
	case *FunctionCallExpression:
		return completeUnionFunctionCall(expression)
	case *PathExpression:
		return supportsTableSourcePath(expression)
	default:
		return false
	}
}

func unionExpressionTable(expression Expression) *TableSymbol {
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			break
		}
		expression = parenthesized.Expression()
	}
	return expressionTableResult(expression)
}

func (t *treeBinder) visitRange(node *RangeOperator) {
	if node == nil || node.IsBound() || !supportsRangeOperator(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	oldRowScope := t.binder.rowScope
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope = oldRowScope
		t.nodeBinder.inPipe = oldNodeBinderInPipe
	}()

	t.binder.rowScope = nil
	for _, expression := range []Expression{node.From(), node.To(), node.Step()} {
		if syntaxElementIsMissing(expression) {
			if reference, ok := expression.(*NameReference); ok {
				reference.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
				continue
			}
		}
		t.visitElement(expression)
		if !expressionIsBound(expression) {
			tx.rollback()
			return
		}
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

func supportsRangeOperator(node *RangeOperator) bool {
	allowMissingKeywords := rangeHasMissingEndpointExpression(node)
	if node == nil || node.RangeToken() == nil || node.RangeToken().Width() == 0 || node.Name() == nil ||
		!supportsRangeKeywordToken(node.FromToken(), allowMissingKeywords) || node.From() == nil ||
		!supportsRangeKeywordToken(node.ToToken(), allowMissingKeywords) || node.To() == nil ||
		node.StepToken() == nil || (node.StepToken().Width() == 0 && !syntaxElementIsMissing(node.StepToken())) || node.Step() == nil {
		return false
	}
	if syntaxElementIsMissing(node.StepToken()) && !rangeHasMissingEndpointExpression(node) {
		toReference, recoveredStepIdentifier := node.To().(*NameReference)
		if !recoveredStepIdentifier || !strings.EqualFold(NameReferenceSimpleName(toReference), "step") ||
			syntaxElementHasMissingChildren(node.From()) || syntaxElementHasMissingChildren(node.To()) || syntaxElementHasMissingChildren(node.Step()) {
			return false
		}
	}
	switch node.Name().Name().(type) {
	case *TokenName, *BracketedName:
	default:
		return false
	}
	return true
}

func rangeHasMissingEndpointExpression(node *RangeOperator) bool {
	return node != nil && (syntaxElementIsMissing(node.From()) || syntaxElementIsMissing(node.To()) || syntaxElementIsMissing(node.Step()))
}

func supportsRangeKeywordToken(token *SyntaxToken, allowMissing bool) bool {
	return token != nil && (token.Width() != 0 || allowMissing && syntaxElementIsMissing(token))
}

type stagedPrintBinding struct {
	node *SyntaxNode
	info SemanticInfo
}

func (t *treeBinder) visitPrint(node *PrintOperator) {
	if !t.supportsPrint(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	committed := false
	defer func() {
		if !committed {
			tx.rollback()
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
		}
	}()
	pending := make([]stagedPrintBinding, 0, node.Expressions().Count()*2)
	for _, separated := range node.Expressions().Elements() {
		if _, ok := t.stagePrintExpression(separated.Element(), &pending); !ok {
			return
		}
	}
	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func (t *treeBinder) supportsPrint(node *PrintOperator) bool {
	if node == nil || node.PrintKeyword() == nil || node.PrintKeyword().Width() == 0 ||
		node.Expressions() == nil || node.Expressions().Count() == 0 {
		return false
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || !t.supportsPrintExpression(separated.Element()) {
			return false
		}
	}
	return true
}

func (t *treeBinder) supportsPrintExpression(expression Expression) bool {
	if expression == nil {
		return false
	}
	switch expression := expression.(type) {
	case *NameReference:
		return syntaxElementIsMissing(expression) || isCompleteProjectByNamesReference(expression)
	case *LiteralExpression:
		switch expression.Kind() {
		case SyntaxKindBooleanLiteralExpression, SyntaxKindIntLiteralExpression, SyntaxKindLongLiteralExpression,
			SyntaxKindRealLiteralExpression, SyntaxKindDecimalLiteralExpression, SyntaxKindDateTimeLiteralExpression,
			SyntaxKindTimespanLiteralExpression, SyntaxKindGuidLiteralExpression, SyntaxKindStringLiteralExpression,
			SyntaxKindNullLiteralExpression, SyntaxKindTokenLiteralExpression:
			return true
		default:
			return false
		}
	case *TypeOfLiteralExpression:
		return expression.Width() != 0
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *DynamicExpression:
		return supportsDynamicExpression(expression)
	case *ToScalarExpression:
		return supportsToScalarTableExpression(expression.SyntaxNode, expression.KindParameter(), expression.Expression())
	case *ToTableExpression:
		return supportsToScalarTableExpression(expression.SyntaxNode, expression.KindParameter(), expression.Expression())
	case *ParenthesizedExpression:
		return expression.Expression() != nil && t.supportsPrintExpression(expression.Expression())
	case *BracketedExpression:
		return parserRecoveredEmptyBracketedExpression(expression) ||
			bindableUnqualifiedBracketedExpression(expression) && t.supportsPrintExpression(expression.Expression())
	case *PathExpression:
		if !completePathExpression(expression) || !t.supportsPrintExpression(expression.Expression()) {
			return false
		}
		switch selector := expression.Selector().(type) {
		case *NameReference:
			return isCompleteProjectByNamesReference(selector)
		case *BracketedExpression:
			return completeBracketedExpression(selector) && t.supportsPrintExpression(selector.Expression())
		case *FunctionCallExpression:
			return t.supportsPrintExpression(selector)
		default:
			return false
		}
	case *ElementExpression:
		selector, ok := expression.Selector().(*BracketedExpression)
		return ok && (completeBracketedExpression(selector) || parserRecoveredEmptyBracketedExpression(selector)) &&
			t.supportsPrintExpression(expression.Expression()) &&
			t.supportsPrintExpression(selector.Expression())
	case *BinaryExpression, *PrefixUnaryExpression, *InExpression, *HasAnyExpression, *HasAllExpression, *BetweenExpression:
		if !supportsScalarOperator(simpleBinderSyntaxNode(expression)) {
			return false
		}
		for _, argument := range scalarOperatorArguments(expression) {
			if !t.supportsPrintExpression(argument) {
				return false
			}
		}
		return true
	case *FunctionCallExpression:
		if expression.Name() == nil || !isCompleteProjectByNamesReference(expression.Name()) || expression.ArgumentList() == nil ||
			expression.ArgumentList().OpenParen() == nil || expression.ArgumentList().OpenParen().Width() == 0 ||
			expression.ArgumentList().Expressions() == nil || expression.ArgumentList().CloseParen() == nil ||
			expression.ArgumentList().CloseParen().Kind() != SyntaxKindCloseParenToken {
			return false
		}
		for _, argument := range expression.ArgumentList().Expressions().Elements() {
			if argument == nil || argument.Element() == nil {
				return false
			}
			if star, ok := argument.Element().(*StarExpression); ok {
				if star.AsteriskToken() == nil || star.AsteriskToken().Width() == 0 {
					return false
				}
				continue
			}
			if !t.supportsPrintExpression(argument.Element()) {
				return false
			}
		}
		return true
	case *SimpleNamedExpression:
		return expression.Name() != nil && NameDeclarationSimpleName(expression.Name()) != "" &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 && expression.Expression() != nil &&
			t.supportsPrintExpression(expression.Expression())
	case *CompoundNamedExpression:
		return compoundRenameListIsSupported(expression.Names()) && expression.EqualToken() != nil &&
			expression.EqualToken().Width() != 0 && expression.Expression() != nil &&
			t.supportsPrintExpression(expression.Expression())
	default:
		return false
	}
}

func (t *treeBinder) stagePrintExpression(expression Expression, pending *[]stagedPrintBinding) (SemanticInfo, bool) {
	var info SemanticInfo
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		inner, ok := t.stagePrintExpression(expression.Expression(), pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *CompoundNamedExpression:
		inner, ok := t.stagePrintExpression(expression.Expression(), pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *BracketedExpression:
		if parserRecoveredEmptyBracketedExpression(expression) {
			if !t.bindParserRecoveredEmptyBracketedExpression(expression) {
				return SemanticInfo{}, false
			}
		} else {
			t.visitElement(expression)
		}
		bound := boundSemanticInfoForExpression(expression)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	default:
		t.visitElement(expression)
		bound := boundSemanticInfoForExpression(expression)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	}
	*pending = append(*pending, stagedPrintBinding{node: simpleBinderSyntaxNode(expression), info: info})
	return info, true
}

func (t *treeBinder) bindParserRecoveredEmptyBracketedExpression(expression *BracketedExpression) bool {
	if !parserRecoveredEmptyBracketedExpression(expression) {
		return false
	}
	t.visitElement(expression.Expression())
	if !expressionIsBound(expression.Expression()) {
		return false
	}
	t.bindNode(expression.SyntaxNode)
	return expression.IsBound()
}

type stagedScanBinding struct {
	node *SyntaxNode
	info SemanticInfo
}

func (t *treeBinder) visitScan(node *ScanOperator) {
	if !t.supportsScan(node) {
		return
	}

	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	oldScanAliases := t.scanAliases
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
		t.scanAliases = oldScanAliases
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	var pending []stagedScanBinding
	// Match Binder_TreeBinder.cs: parameters and source clauses are bound before
	// the declaration/step child scope exists.
	for _, parameter := range node.Parameters().Elements() {
		if expression := parameter.Expression(); expression != nil {
			if _, ok := t.stageScanExpression(expression, &pending); !ok {
				return
			}
		}
	}
	if clause := node.OrderByClause(); clause != nil {
		for _, expression := range clause.Expressions().Elements() {
			if _, ok := t.stageScanExpression(expression.Element(), &pending); !ok {
				return
			}
		}
	}
	if clause := node.PartitionByClause(); clause != nil {
		for _, expression := range clause.Expressions().Elements() {
			if _, ok := t.stageScanExpression(expression.Element(), &pending); !ok {
				return
			}
		}
	}

	t.binder.LocalScope = NewLocalScope(oldLocalScope)
	declared := make([]*ColumnSymbol, 0)
	if clause := node.DeclareClause(); clause != nil {
		for _, separated := range clause.Declarations().Elements() {
			declaration := separated.Element().NameAndType()
			column := scanDeclaredColumn(declaration)
			declared = append(declared, column)
			t.binder.LocalScope.AddSymbol(column)
			pending = append(pending, stagedScanBinding{declaration.Name().SyntaxNode, NewSemanticInfoForReferencedSymbol(column)})
		}
	}

	tupleColumns := append([]*ColumnSymbol{}, t.binder.declaredAndInferredColumns(t.binder.rowScope)...)
	tupleColumns = append(tupleColumns, declared...)
	stepTuple := NewTupleSymbol(tupleColumns, t.binder.rowScope)
	t.scanAliases = make(map[string]*VariableSymbol, node.Steps().Count())
	for _, step := range node.Steps().Elements() {
		name := NameDeclarationSimpleName(step.Name())
		variable := NewVariableSymbol(name, stepTuple, false, nil, nil)
		t.scanAliases[name] = variable
		t.binder.LocalScope.AddSymbol(variable)
		pending = append(pending, stagedScanBinding{step.Name().SyntaxNode, NewSemanticInfoForReferencedSymbol(variable)})
	}

	for _, step := range node.Steps().Elements() {
		if _, ok := t.stageScanExpression(step.Condition(), &pending); !ok {
			return
		}
		if computation := step.ComputationClause(); computation != nil {
			for _, separated := range computation.Assignments().Elements() {
				assignment := separated.Element()
				if _, ok := t.stageScanExpression(assignment.Name(), &pending); !ok {
					return
				}
				if _, ok := t.stageScanExpression(assignment.Expression(), &pending); !ok {
					return
				}
			}
		}
	}

	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func (t *treeBinder) supportsScan(node *ScanOperator) bool {
	if node == nil || node.Parameters() == nil || node.Steps() == nil {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Expression() == nil || !supportsScanParameterExpression(parameter.Expression()) {
			return false
		}
	}
	for _, clause := range []*SeparatedSyntaxListOf[Expression]{scanOrderExpressions(node), scanPartitionExpressions(node)} {
		if clause == nil {
			continue
		}
		for _, separated := range clause.Elements() {
			if separated == nil || !supportsScanExpression(separated.Element()) {
				return false
			}
		}
	}
	if clause := node.DeclareClause(); clause != nil {
		if clause.Declarations() == nil {
			return false
		}
		for _, separated := range clause.Declarations().Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().NameAndType() == nil ||
				separated.Element().NameAndType().Name() == nil || separated.Element().NameAndType().Type() == nil {
				return false
			}
		}
	}
	for _, step := range node.Steps().Elements() {
		if step == nil || step.Name() == nil || !supportsScanExpression(step.Condition()) {
			return false
		}
		if computation := step.ComputationClause(); computation != nil {
			if computation.Assignments() == nil {
				return false
			}
			for _, separated := range computation.Assignments().Elements() {
				if separated == nil || separated.Element() == nil || separated.Element().Name() == nil ||
					(!scanParserMissingNameReference(separated.Element().Name()) && !isCompleteProjectByNamesReference(separated.Element().Name())) ||
					(!scanParserMissingNameReference(separated.Element().Expression()) && !supportsScanExpression(separated.Element().Expression())) {
					return false
				}
			}
		}
	}
	return true
}

func scanOrderExpressions(node *ScanOperator) *SeparatedSyntaxListOf[Expression] {
	if node != nil && node.OrderByClause() != nil {
		return node.OrderByClause().Expressions()
	}
	return nil
}

func scanPartitionExpressions(node *ScanOperator) *SeparatedSyntaxListOf[Expression] {
	if node != nil && node.PartitionByClause() != nil {
		return node.PartitionByClause().Expressions()
	}
	return nil
}

func supportsScanParameterExpression(expression Expression) bool {
	if _, ok := expression.(*NameDeclaration); ok {
		return true
	}
	return supportsScanExpression(expression)
}

func supportsScanExpression(expression Expression) bool {
	return supportsScanScalar(expression)
}

func scanParserMissingNameReference(expression Expression) bool {
	reference, ok := expression.(*NameReference)
	return ok && reference.Width() == 0 && NameReferenceSimpleName(reference) == ""
}

func scanDeclaredColumn(declaration *NameAndTypeDeclaration) *ColumnSymbol {
	typ := &Error.TypeSymbol
	if declaration != nil {
		if declared, ok := declaredSymbolFromTypeExpression(declaration.Type()).(*TypeSymbol); ok && declared != nil {
			typ = declared
		}
	}
	return NewColumnSymbolWithSource(NameDeclarationSimpleName(declaration.Name()), typ, declaration.Name())
}

func (t *treeBinder) stageScanExpression(expression Expression, pending *[]stagedScanBinding) (SemanticInfo, bool) {
	var info SemanticInfo
	switch expression := expression.(type) {
	case *NameDeclaration:
		return SemanticInfo{}, true
	case *NameReference:
		if alias := t.scanStepAlias(NameReferenceSimpleName(expression)); alias != nil {
			info = NewSemanticInfoForReferencedSymbol(alias)
		} else {
			info = t.binder.ResolveNameReferenceInRowScope(t.binder.rowScope, expression, false)
		}
	case *LiteralExpression:
		bound := t.nodeBinder.bindLiteral(expression)
		if bound != nil {
			info = *bound
		}
	case *ParenthesizedExpression:
		inner, ok := t.stageScanExpression(expression.Expression(), pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *SimpleNamedExpression:
		inner, ok := t.stageScanExpression(expression.Expression(), pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *BinaryExpression:
		left, leftOK := t.stageScanExpression(expression.Left(), pending)
		right, rightOK := t.stageScanExpression(expression.Right(), pending)
		if !leftOK || !rightOK {
			return SemanticInfo{}, false
		}
		info = scanBinarySemanticInfo(expression, left, right)
	case *PathExpression:
		selector, simpleSelector := expression.Selector().(*NameReference)
		if simpleSelector {
			left, ok := t.stageScanExpression(expression.Expression(), pending)
			if !ok {
				return SemanticInfo{}, false
			}
			selectorInfo := t.resolveScanPathSelector(left, selector)
			*pending = append(*pending, stagedScanBinding{selector.SyntaxNode, selectorInfo})
			info = NewSemanticInfoForSymbol(selectorInfo.ReferencedSymbol(), topHittersStagedResultType(selectorInfo))
		} else {
			t.visitElement(expression)
			bound := boundSemanticInfoForExpression(expression)
			if bound == nil {
				return SemanticInfo{}, false
			}
			return *bound, true
		}
	default:
		t.visitElement(expression)
		bound := boundSemanticInfoForExpression(expression)
		if bound == nil {
			return SemanticInfo{}, false
		}
		return *bound, true
	}
	if node := simpleBinderSyntaxNode(expression); node != nil {
		*pending = append(*pending, stagedScanBinding{node, info})
	} else if path, ok := expression.(*PathExpression); ok {
		*pending = append(*pending, stagedScanBinding{path.SyntaxNode, info})
	}
	return info, true
}

func (t *treeBinder) scanStepAlias(name string) *VariableSymbol {
	if t == nil || name == "" {
		return nil
	}
	if alias := t.scanAliases[name]; alias != nil {
		return alias
	}
	if t.binder == nil || t.binder.LocalScope == nil {
		return nil
	}
	var symbols []Symbol
	t.binder.LocalScope.GetSymbols(name, SymbolMatchLocal, &symbols)
	for _, symbol := range symbols {
		variable, ok := symbol.(*VariableSymbol)
		if !ok {
			continue
		}
		if _, tuple := variable.Type.(*TupleSymbol); tuple {
			return variable
		}
	}
	return nil
}

func scanBinarySemanticInfo(expression *BinaryExpression, left, right SemanticInfo) SemanticInfo {
	switch expression.Kind() {
	case SyntaxKindLessThanExpression, SyntaxKindLessThanOrEqualExpression,
		SyntaxKindGreaterThanExpression, SyntaxKindGreaterThanOrEqualExpression,
		SyntaxKindEqualExpression, SyntaxKindNotEqualExpression, SyntaxKindEqualTildeExpression,
		SyntaxKindAndExpression, SyntaxKindOrExpression:
		return NewSemanticInfoForType(ScalarTypes.Bool.Type())
	default:
		leftType := topHittersStagedResultType(left)
		rightType := topHittersStagedResultType(right)
		if sameType(leftType, rightType) {
			return NewSemanticInfoForType(leftType)
		}
		return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
	}
}

func (t *treeBinder) resolveScanPathSelector(left SemanticInfo, selector *NameReference) SemanticInfo {
	name := NameReferenceSimpleName(selector)
	var tuple *TupleSymbol
	switch symbol := left.ReferencedSymbol().(type) {
	case *VariableSymbol:
		tuple, _ = symbol.Type.(*TupleSymbol)
	case *TupleSymbol:
		tuple = symbol
	}
	if tuple == nil && left.ResultType != nil {
		tuple = left.ResultType.tuple
	}
	if tuple != nil {
		if members := FindMembers(tuple, name, normalizeNameMatch(selector.Match()), false); len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		} else if len(members) > 1 {
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
		}
	}
	diagnostic := DiagnosticNameDoesNotReferToAnyKnownItem(name, "")
	if t.binder != nil {
		diagnostic = t.binder.diagnosticAtSyntaxElement(diagnostic, selector)
	}
	return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
}

func (t *treeBinder) visitGetSchema(node *GetSchemaOperator) {
	if !supportsGetSchema(node) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldPipedAsRowScope := t.nodeBinder.pipedAsRowScope
	oldDefaultColumnNameSuffix := t.nodeBinder.defaultColumnNameSuffix
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.nodeBinder.pipedAsRowScope = oldPipedAsRowScope
		t.nodeBinder.defaultColumnNameSuffix = oldDefaultColumnNameSuffix
	}()

	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	t.bindNode(node.SyntaxNode)
}

func supportsGetSchema(node *GetSchemaOperator) bool {
	if node == nil || node.GetSchemaKeyword() == nil || node.GetSchemaKeyword().Kind() != SyntaxKindGetSchemaKeyword ||
		node.GetSchemaKeyword().Width() == 0 {
		return false
	}
	parameter := node.KindParameter()
	if parameter == nil {
		return true
	}
	if parameter.Name() == nil || NameDeclarationSimpleName(parameter.Name()) != "kind" ||
		parameter.EqualToken() == nil || parameter.EqualToken().Kind() != SyntaxKindEqualToken {
		return false
	}
	literal, ok := parameter.Expression().(*LiteralExpression)
	return ok && literal.Token() != nil
}

func (t *treeBinder) visitTopNested(node *TopNestedOperator) {
	if !t.supportsTopNested(node) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
	}()

	pending := make([]stagedSummarizeBinding, 0)
	for _, separated := range node.Clauses().Elements() {
		clause := separated.Element()
		t.binder.scopeKind = ScopeKindNormal
		if clause.Expression() != nil {
			if _, ok := t.stageSummarizeScalar(clause.Expression(), false, &pending); !ok {
				return
			}
		}
		if _, ok := t.stageSummarizeScalar(clause.OfExpression(), false, &pending); !ok {
			return
		}
		if others := clause.WithOthersClause(); others != nil {
			if _, ok := t.stageSummarizeScalar(others.Expression(), false, &pending); !ok {
				return
			}
		}
		t.binder.scopeKind = ScopeKindAggregate
		if _, ok := t.stageTopNestedByExpression(clause.ByExpression(), &pending); !ok {
			return
		}
	}

	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
	}
	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) supportsTopNested(node *TopNestedOperator) bool {
	if node == nil || node.Clauses() == nil || node.Clauses().Count() == 0 || (t.inPipe && t.binder.rowScope == nil) {
		return false
	}
	for _, separated := range node.Clauses().Elements() {
		if separated == nil || separated.Element() == nil {
			return false
		}
		clause := separated.Element()
		if parserRecoveredMissingTopNestedClause(clause) {
			continue
		}
		if clause.TopNestedKeyword() == nil || clause.OfKeyword() == nil || clause.ByKeyword() == nil ||
			clause.OfExpression() == nil || clause.ByExpression() == nil ||
			(clause.Expression() != nil && !t.supportsSummarizeScalar(clause.Expression(), false)) ||
			!t.supportsSummarizeScalar(clause.OfExpression(), false) || !t.supportsTopNestedByExpression(clause.ByExpression()) {
			return false
		}
		if others := clause.WithOthersClause(); others != nil {
			if others.WithKeyword() == nil || others.OthersKeyword() == nil || others.Equal() == nil || others.Expression() == nil ||
				!t.supportsSummarizeScalar(others.Expression(), false) {
				return false
			}
		}
	}
	return true
}

func parserRecoveredMissingTopNestedClause(clause *TopNestedClause) bool {
	if clause == nil || clause.Width() != 0 {
		return false
	}
	for _, diagnostic := range clause.ContainedDiagnostics() {
		if diagnostic.Code == "KS006" && diagnostic.Message == "Missing clause" {
			return true
		}
	}
	return false
}

func (t *treeBinder) supportsTopNestedByExpression(expression Expression) bool {
	if ordered, ok := expression.(*OrderedExpression); ok {
		return ordered.Expression() != nil && ordered.Ordering() != nil && t.supportsSummarizeScalar(ordered.Expression(), true)
	}
	return t.supportsSummarizeScalar(expression, true)
}

func (t *treeBinder) stageTopNestedByExpression(expression Expression, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	if ordered, ok := expression.(*OrderedExpression); ok {
		inner, supported := t.stageSummarizeScalar(ordered.Expression(), true, pending)
		if !supported {
			return SemanticInfo{}, false
		}
		info := NewSemanticInfoForType(topHittersStagedResultType(inner))
		*pending = append(*pending, stagedSummarizeBinding{node: ordered.SyntaxNode, info: info})
		return info, true
	}
	return t.stageSummarizeScalar(expression, true, pending)
}

type stagedSummarizeBinding struct {
	node         *SyntaxNode
	info         SemanticInfo
	resultSymbol Symbol
}

func (t *treeBinder) visitSummarize(node *SummarizeOperator) {
	if !t.supportsSummarize(node) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldSummarizeByColumns := t.summarizeByColumns
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		t.summarizeByColumns = oldSummarizeByColumns
	}()
	t.summarizeByColumns = nil

	pending := make([]stagedSummarizeBinding, 0)
	// Match Binder_TreeBinder.cs: parameters, grouping expressions, then
	// aggregate expressions under aggregate scope.
	for _, parameter := range node.Parameters().Elements() {
		if parameter != nil && parameter.Expression() != nil {
			if _, ok := t.stageSummarizeScalar(parameter.Expression(), false, &pending); !ok {
				return
			}
		}
	}
	if by := node.ByClause(); by != nil {
		for _, separated := range by.Expressions().Elements() {
			expression := separated.Element()
			if _, ok := t.stageSummarizeScalar(expression, false, &pending); !ok {
				return
			}
			if column := stagedSummarizeResultColumn(expression, pending); column != nil {
				t.summarizeByColumns = append(t.summarizeByColumns, column)
			}
		}
	}
	t.binder.scopeKind = ScopeKindAggregate
	for _, separated := range node.Aggregates().Elements() {
		if _, ok := t.stageSummarizeScalar(separated.Element(), true, &pending); !ok {
			return
		}
	}

	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
		if binding.resultSymbol != nil {
			binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
		}
	}
	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) supportsSummarize(node *SummarizeOperator) bool {
	if node == nil || node.SummarizeKeyword() == nil || node.SummarizeKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Aggregates() == nil || (t.inPipe && t.binder.rowScope == nil) {
		return false
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || NameDeclarationSimpleName(parameter.Name()) == "" ||
			parameter.Expression() == nil || !t.supportsSummarizeScalar(parameter.Expression(), false) {
			return false
		}
	}
	if by := node.ByClause(); by != nil {
		if by.ByKeyword() == nil || by.ByKeyword().Width() == 0 || by.Expressions() == nil || by.BinClause() != nil {
			return false
		}
		for _, separated := range by.Expressions().Elements() {
			if separated == nil || !t.supportsSummarizeScalar(separated.Element(), false) {
				return false
			}
		}
	}
	for _, separated := range node.Aggregates().Elements() {
		if separated == nil || !t.supportsSummarizeScalar(separated.Element(), true) {
			return false
		}
	}
	return true
}

func (t *treeBinder) supportsSummarizeScalar(expression Expression, aggregate bool) bool {
	if expression == nil {
		return false
	}
	if syntaxElementIsMissing(expression) {
		_, ok := expression.(*NameReference)
		return ok
	}
	switch expression := expression.(type) {
	case *NameReference:
		return isCompleteProjectByNamesReference(expression)
	case *LiteralExpression:
		return expression.Width() != 0 && isSupportedTopHittersLiteral(expression)
	case *TypeOfLiteralExpression:
		return expression.Width() != 0
	case *CompoundStringLiteralExpression:
		return supportsCompoundStringLiteral(expression)
	case *DynamicExpression:
		return supportsDynamicExpression(expression)
	case *ToScalarExpression:
		return supportsToScalarTableExpression(expression.SyntaxNode, expression.KindParameter(), expression.Expression())
	case *ParenthesizedExpression:
		return t.supportsSummarizeScalar(expression.Expression(), aggregate)
	case *BracketedExpression:
		return completeBracketedExpression(expression) && t.supportsSummarizeScalar(expression.Expression(), false)
	case *PathExpression:
		if !completePathExpression(expression) || !t.supportsSummarizeScalar(expression.Expression(), false) {
			return false
		}
		switch selector := expression.Selector().(type) {
		case *NameReference:
			return isCompleteProjectByNamesReference(selector)
		case *BracketedExpression:
			return completeBracketedExpression(selector) && t.supportsSummarizeScalar(selector.Expression(), false)
		default:
			return false
		}
	case *ElementExpression:
		selector, ok := expression.Selector().(*BracketedExpression)
		return ok && (completeBracketedExpression(selector) || parserRecoveredEmptyBracketedExpression(selector)) &&
			t.supportsSummarizeScalar(expression.Expression(), false) &&
			t.supportsSummarizeScalar(selector.Expression(), false)
	case *BinaryExpression:
		return expression.Operator() != nil && expression.Operator().Width() != 0 && operatorKindForSyntax(expression.Kind()) != OperatorKindNone &&
			t.supportsSummarizeScalar(expression.Left(), false) && t.supportsSummarizeScalar(expression.Right(), false)
	case *PrefixUnaryExpression:
		return expression.Operator() != nil && expression.Operator().Width() != 0 && operatorKindForSyntax(expression.Kind()) != OperatorKindNone &&
			t.supportsSummarizeScalar(expression.Expression(), false)
	case *InExpression:
		return t.supportsSummarizeListOperator(expression.SyntaxNode, expression.Operator(), expression.Left(), expression.Right())
	case *HasAnyExpression:
		return t.supportsSummarizeListOperator(expression.SyntaxNode, expression.Operator(), expression.Left(), expression.Right())
	case *HasAllExpression:
		return t.supportsSummarizeListOperator(expression.SyntaxNode, expression.Operator(), expression.Left(), expression.Right())
	case *BetweenExpression:
		return expression.Operator() != nil && expression.Operator().Width() != 0 && operatorKindForSyntax(expression.Kind()) != OperatorKindNone &&
			expression.Right() != nil && expression.Right().OpenParen() != nil && expression.Right().OpenParen().Width() != 0 &&
			expression.Right().DotDot() != nil && expression.Right().DotDot().Width() != 0 && expression.Right().CloseParen() != nil &&
			expression.Right().CloseParen().Width() != 0 && t.supportsSummarizeScalar(expression.Left(), false) &&
			t.supportsSummarizeScalar(expression.Right().First(), false) && t.supportsSummarizeScalar(expression.Right().Second(), false)
	case *SimpleNamedExpression:
		return expression.Name() != nil && serializeNameDeclarationIsSupported(expression.Name()) &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			t.supportsSummarizeScalar(expression.Expression(), aggregate)
	case *CompoundNamedExpression:
		return compoundRenameListIsSupported(expression.Names()) && expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			t.supportsSummarizeScalar(expression.Expression(), aggregate)
	case *FunctionCallExpression:
		if expression.Name() == nil || !isCompleteProjectByNamesReference(expression.Name()) || expression.ArgumentList() == nil ||
			expression.ArgumentList().Expressions() == nil {
			return false
		}
		function := t.summarizeFunction(expression.Name(), aggregate)
		// An unresolved call is structurally supported. The upstream binder binds
		// its arguments and commits ErrorSymbol plus the name diagnostic instead
		// of abandoning the containing summarize operator.
		if function != nil && !boundedSummarizeFunction(function) {
			return false
		}
		argumentAggregate := t.summarizeArgumentIsAggregate(function, aggregate)
		for _, argument := range expression.ArgumentList().Expressions().Elements() {
			if argument == nil || (!isSummarizeStarArgument(argument.Element()) && !t.supportsSummarizeScalar(argument.Element(), argumentAggregate)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (t *treeBinder) supportsSummarizeListOperator(node *SyntaxNode, operator *SyntaxToken, left Expression, right *ExpressionList) bool {
	if node == nil || operator == nil || operator.Width() == 0 || operatorKindForSyntax(node.Kind()) == OperatorKindNone ||
		right == nil || right.OpenParen() == nil || right.OpenParen().Width() == 0 || right.Expressions() == nil ||
		right.Expressions().Count() == 0 || right.CloseParen() == nil || right.CloseParen().Width() == 0 ||
		!t.supportsSummarizeScalar(left, false) {
		return false
	}
	for _, separated := range right.Expressions().Elements() {
		if separated == nil || !t.supportsSummarizeScalar(separated.Element(), false) {
			return false
		}
	}
	return true
}

func boundedSummarizeFunction(function *FunctionSymbol) bool {
	if function == nil || len(function.Signatures()) == 0 {
		return false
	}
	for _, signature := range function.Signatures() {
		if signature == nil || (signature.ReturnKind == ReturnTypeComputed && signature.Body == "" && signature.Declaration == nil) ||
			(signature.ReturnKind == ReturnTypeCustom && signature.CustomReturnType == nil) {
			return false
		}
		if signature.Tabularity == TabularityTabular {
			return false
		}
	}
	return true
}

func (t *treeBinder) summarizeFunction(name *NameReference, aggregate bool) *FunctionSymbol {
	if t == nil || t.binder == nil || name == nil {
		return nil
	}
	globals := t.binder.Globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	simpleName := NameReferenceSimpleName(name)
	if aggregate {
		if function := globals.GetAggregate(simpleName); function != nil {
			return function
		}
		if t.binder.LocalScope != nil {
			var locals []Symbol
			t.binder.LocalScope.GetSymbols(simpleName, SymbolMatchFunction|SymbolMatchView, &locals)
			if len(locals) == 1 {
				function, _ := locals[0].(*FunctionSymbol)
				return function
			}
		}
		// Aggregate scope also exposes ordinary scalar functions. Their arguments
		// remain in aggregate scope in current Kusto.Language, allowing nested
		// aggregates while rejecting direct row-column references.
		if globals.Database != nil {
			if function := globals.Database.GetFunction(simpleName); function != nil {
				return function
			}
		}
		return globals.GetFunction(simpleName)
	}
	info := t.binder.ResolveName(simpleName, SymbolMatchFunction|SymbolMatchView, false)
	function, _ := info.ReferencedSymbol().(*FunctionSymbol)
	return function
}

func (t *treeBinder) summarizeArgumentIsAggregate(function *FunctionSymbol, aggregate bool) bool {
	if !aggregate {
		return false
	}
	globals := t.binder.Globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	return !globals.IsAggregateFunction(function)
}

func (t *treeBinder) stageSummarizeScalar(expression Expression, aggregate bool, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	var info SemanticInfo
	var resultSymbol Symbol
	switch expression := expression.(type) {
	case *NameReference:
		if syntaxElementIsMissing(expression) {
			info = NewSemanticInfoForReferencedSymbol(Error)
		} else if aggregate {
			info = t.binder.resolveAggregateNameReference(t.binder.rowScope, expression)
		} else {
			info = t.binder.ResolveNameReferenceInRowScope(t.binder.rowScope, expression, false)
		}
	case *LiteralExpression:
		bound := t.nodeBinder.bindLiteral(expression)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	case *TypeOfLiteralExpression:
		bound := t.nodeBinder.bindTypeOfLiteral(expression)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	case *CompoundStringLiteralExpression:
		bound := t.nodeBinder.bind(expression.SyntaxNode)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	case *DynamicExpression:
		oldDepth := t.binder.dynamicDepth
		t.binder.dynamicDepth++
		inner, ok := t.stageSummarizeDynamicJSONValue(expression.Expression(), pending)
		t.binder.dynamicDepth = oldDepth
		if !ok {
			return SemanticInfo{}, false
		}
		tx := newUnionBindingTransaction(expression)
		simpleBinderSyntaxNode(expression.Expression()).SetSemanticInfo(inner)
		bound := t.nodeBinder.bindDynamicExpression(expression)
		tx.rollback()
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
	case *ToScalarExpression:
		bound, ok := t.stageSummarizeToScalar(expression, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		return bound, true
	case *ParenthesizedExpression:
		inner, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *BracketedExpression:
		inner, ok := t.stageSummarizeScalar(expression.Expression(), false, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		tx := newUnionBindingTransaction(expression)
		simpleBinderSyntaxNode(expression.Expression()).SetSemanticInfo(inner)
		info = t.binder.bindUnqualifiedBracketedExpression(expression)
		tx.rollback()
	case *PathExpression:
		collection, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		switch selector := expression.Selector().(type) {
		case *NameReference:
			oldPathScope := t.binder.pathScope
			t.binder.pathScope = stagedSummarizePathScope(collection)
			selectorInfo := t.binder.ResolveName(NameReferenceSimpleName(selector), SymbolMatchDefault, false)
			t.binder.pathScope = oldPathScope
			*pending = append(*pending, stagedSummarizeBinding{node: selector.SyntaxNode, info: selectorInfo})
			info = NewSemanticInfoForSymbol(selectorInfo.ReferencedSymbol(), topHittersStagedResultType(selectorInfo), selectorInfo.Diagnostics...)
		case *BracketedExpression:
			selectorInfo, ok := t.stageSummarizeElementSelector(expression, expression.Expression(), collection, selector, pending)
			if !ok {
				return SemanticInfo{}, false
			}
			info = NewSemanticInfoForSymbol(selectorInfo.ReferencedSymbol(), topHittersStagedResultType(selectorInfo), selectorInfo.Diagnostics...)
		default:
			return SemanticInfo{}, false
		}
	case *ElementExpression:
		collection, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		selector, ok := expression.Selector().(*BracketedExpression)
		if !ok {
			return SemanticInfo{}, false
		}
		selectorInfo, ok := t.stageSummarizeElementSelector(expression, expression.Expression(), collection, selector, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForSymbol(selectorInfo.ReferencedSymbol(), topHittersStagedResultType(selectorInfo), selectorInfo.Diagnostics...)
	case *BinaryExpression:
		nestedAggregate := aggregate
		left, ok := t.stageSummarizeScalar(expression.Left(), nestedAggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		right, ok := t.stageSummarizeScalar(expression.Right(), nestedAggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		bound := t.nodeBinder.operatorInfoForTypes(
			t.nodeBinder.operatorForKind(operatorKindForSyntax(expression.Kind())),
			[]Expression{expression.Left(), expression.Right()},
			[]*TypeSymbol{topHittersStagedResultType(left), topHittersStagedResultType(right)},
			expression.Operator(), true, true,
		)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
		info.IsConstant = left.IsConstant && right.IsConstant
	case *PrefixUnaryExpression:
		operand, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		bound := t.nodeBinder.operatorInfoForTypes(
			t.nodeBinder.operatorForKind(operatorKindForSyntax(expression.Kind())),
			[]Expression{expression.Expression()},
			[]*TypeSymbol{topHittersStagedResultType(operand)},
			expression.Operator(), true, true,
		)
		if bound == nil {
			return SemanticInfo{}, false
		}
		info = *bound
		info.IsConstant = operand.IsConstant
	case *InExpression:
		arguments := listOperatorArguments(expression.Left(), expression.Right())
		bound, ok := t.stageSummarizeOperator(expression.SyntaxNode, expression.Operator(), arguments, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = bound
	case *HasAnyExpression:
		arguments := listOperatorArguments(expression.Left(), expression.Right())
		bound, ok := t.stageSummarizeOperator(expression.SyntaxNode, expression.Operator(), arguments, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = bound
	case *HasAllExpression:
		arguments := listOperatorArguments(expression.Left(), expression.Right())
		bound, ok := t.stageSummarizeOperator(expression.SyntaxNode, expression.Operator(), arguments, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = bound
	case *BetweenExpression:
		arguments := []Expression{expression.Left(), expression.Right().First(), expression.Right().Second()}
		bound, ok := t.stageSummarizeOperator(expression.SyntaxNode, expression.Operator(), arguments, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = bound
	case *SimpleNamedExpression:
		inner, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *CompoundNamedExpression:
		inner, ok := t.stageSummarizeScalar(expression.Expression(), aggregate, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		info = NewSemanticInfoForType(topHittersStagedResultType(inner))
	case *FunctionCallExpression:
		function := t.summarizeFunction(expression.Name(), aggregate)
		arguments := expression.ArgumentList().Expressions()
		argumentExpressions := make([]Expression, 0, arguments.Count())
		argumentTypes := make([]*TypeSymbol, 0, arguments.Count())
		argumentSemantics := make([]SemanticInfo, 0, arguments.Count())
		argumentBindingsStart := len(*pending)
		outerScope := t.binder.scopeKind
		argumentAggregate := t.summarizeArgumentIsAggregate(function, aggregate)
		for _, argument := range arguments.Elements() {
			var argumentInfo SemanticInfo
			var ok bool
			oldScope := t.binder.scopeKind
			t.binder.scopeKind = ArgumentScopeForFunction(t.binder.Globals, function, outerScope)
			if isSummarizeStarArgument(argument.Element()) {
				bound := t.nodeBinder.bindStarExpression(argument.Element().(*StarExpression))
				if bound != nil {
					argumentInfo, ok = *bound, true
					*pending = append(*pending, stagedSummarizeBinding{node: simpleBinderSyntaxNode(argument.Element()), info: argumentInfo})
				}
			} else {
				argumentInfo, ok = t.stageSummarizeScalar(argument.Element(), argumentAggregate, pending)
			}
			t.binder.scopeKind = oldScope
			if !ok {
				return SemanticInfo{}, false
			}
			argumentExpressions = append(argumentExpressions, argument.Element())
			argumentTypes = append(argumentTypes, topHittersStagedResultType(argumentInfo))
			argumentSemantics = append(argumentSemantics, argumentInfo)
		}
		if function == nil {
			diagnostic := DiagnosticNameDoesNotReferToAnyKnownFunction(NameReferenceSimpleName(expression.Name()))
			if !aggregate {
				globals := t.binder.Globals
				if globals == nil {
					globals = DefaultGlobalState
				}
				if aggregateFunction := globals.GetAggregate(NameReferenceSimpleName(expression.Name())); aggregateFunction != nil {
					diagnostic = DiagnosticAggregateNotAllowedInThisContext(aggregateFunction.SymbolName())
				}
			}
			nameInfo := NewSemanticInfoForReferencedSymbol(Error, t.binder.diagnosticAtSyntaxElement(
				diagnostic, expression.Name()))
			info = NewSemanticInfoForReferencedSymbol(Error)
			*pending = append(*pending, stagedSummarizeBinding{node: expression.Name().SyntaxNode, info: nameInfo})
			// Current BindFunctionCallOrPattern returns ErrorInfo after a failed
			// name. The compatibility implementation returned nil, leaving only the name
			// node bound; projection still recovered an error-typed column.
			*pending = append(*pending, stagedSummarizeBinding{node: simpleBinderSyntaxNode(expression), info: info})
			return info, true
		}

		matches := t.nodeBinder.getBestMatchingFunctionSignatures(function.Signatures(), argumentExpressions, symbolsFromTypeSymbols(argumentTypes), false)
		var signature *Signature
		if len(matches) == 1 {
			signature = matches[0]
		} else {
			signature = uniqueArgumentCountSignature(function, len(argumentExpressions))
		}
		result := ScalarTypes.Unknown.Type()
		nameInfo := SemanticInfo{}
		var calledFunctionInfo *FunctionCallInfo
		if signature != nil {
			tx := newUnionBindingTransaction(expression)
			for _, binding := range (*pending)[argumentBindingsStart:] {
				binding.node.SetSemanticInfo(binding.info)
				if binding.resultSymbol != nil {
					binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
				}
			}
			parameters := signature.ArgumentParametersForExpressions(argumentExpressions, symbolsFromTypeSymbols(argumentTypes), t.binder.Globals)
			diagnostics := t.nodeBinder.functionSignatureDiagnostics(signature, expression.Name(), argumentExpressions, argumentTypes, parameters)
			var resultDiagnostics []Diagnostic
			resultSymbol, result, calledFunctionInfo, resultDiagnostics = t.nodeBinder.functionCallResult(
				signature, expression.SyntaxNode, argumentExpressions, argumentTypes, parameters,
				argumentSemantics, t.summarizeByColumns,
			)
			diagnostics = append(diagnostics, resultDiagnostics...)
			if len(matches) == 1 {
				initialResult := result
				result = t.nodeBinder.refineDynamicFunctionCallResult(
					function, expression.SyntaxNode, argumentExpressions, argumentTypes, result,
					argumentSemantics, t.summarizeByColumns,
				)
				if result != initialResult && resultSymbol == initialResult {
					resultSymbol = result
				}
			}
			info = NewSemanticInfoForSignature(signature, result, diagnostics...)
			if calledFunctionInfo != nil {
				info = info.WithCalledFunctionInfo(calledFunctionInfo)
			}
			nameInfo = NewSemanticInfoForSignature(signature, result)
			tx.rollback()
		} else if len(matches) > 1 {
			tx := newUnionBindingTransaction(expression)
			for _, binding := range (*pending)[argumentBindingsStart:] {
				binding.node.SetSemanticInfo(binding.info)
				if binding.resultSymbol != nil {
					binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
				}
			}
			result = GetCommonSignatureReturnType(matches, argumentTypes, func(candidate *Signature) *TypeSymbol {
				parameters := candidate.ArgumentParametersForExpressions(argumentExpressions, symbolsFromTypeSymbols(argumentTypes), t.binder.Globals)
				_, candidateResult, _, _ := t.nodeBinder.functionCallResult(
					candidate, expression.SyntaxNode, argumentExpressions, argumentTypes, parameters,
					argumentSemantics, t.summarizeByColumns,
				)
				return candidateResult
			})
			diagnostics := []Diagnostic{}
			if len(argumentExpressions) == 0 && function.MinArgumentCount() > 0 {
				diagnostics = append(diagnostics, t.binder.diagnosticAtSyntaxElement(
					DiagnosticFunctionExpectsArgumentCountRange(function.SymbolName(), function.MinArgumentCount(), function.MaxArgumentCount()), expression.Name()))
			} else if !argumentTypesHaveErrorsOrUnknown(argumentTypes) {
				typeNames := make([]string, len(argumentTypes))
				for i, typ := range argumentTypes {
					typeNames[i] = typ.SymbolName()
				}
				diagnostics = append(diagnostics, t.binder.diagnosticAtSyntaxElement(
					DiagnosticFunctionNotDefinedWithMatchingParameters(function.SymbolName(), typeNames...), expression.Name()))
			}
			info = NewSemanticInfoForSymbol(function, result, diagnostics...)
			nameInfo = NewSemanticInfoForSymbol(function, result)
			tx.rollback()
		}
		if result == nil || !IsScalar(result) || isErrorTypeSymbol(result) {
			result = ScalarTypes.Unknown.Type()
		}
		if signature != nil || len(matches) > 1 {
			info.ResultType = result
			nameInfo.ResultType = result
		} else {
			typeNames := make([]string, len(argumentTypes))
			for i, typ := range argumentTypes {
				typeNames[i] = typ.SymbolName()
			}
			diagnostic := t.binder.diagnosticAtSyntaxElement(
				DiagnosticFunctionNotDefinedWithMatchingParameters(function.SymbolName(), typeNames...), expression.Name())
			info = NewSemanticInfoForSymbol(function, result, diagnostic)
			nameInfo = NewSemanticInfoForSymbol(function, result)
		}
		*pending = append(*pending, stagedSummarizeBinding{node: expression.Name().SyntaxNode, info: nameInfo})
	default:
		return SemanticInfo{}, false
	}
	*pending = append(*pending, stagedSummarizeBinding{node: simpleBinderSyntaxNode(expression), info: info, resultSymbol: resultSymbol})
	return info, true
}

func (t *treeBinder) stageSummarizeToScalar(expression *ToScalarExpression, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	tx := newUnionBindingTransaction(expression)
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)

	t.visitToScalarExpression(expression)
	root := expression.SyntaxNode
	info, bound := root.GetSemanticInfo()
	inferenceChanged := !sameOpenTableInference(
		oldInferredColumns, oldInferredOrder,
		t.binder.openTableInferredColumns, t.binder.openTableInferredOrder,
	)
	bindings := collectStagedSummarizeBindings(expression)

	t.binder.openTableInferredColumns = oldInferredColumns
	t.binder.openTableInferredOrder = oldInferredOrder
	tx.rollback()
	if !bound || inferenceChanged {
		return SemanticInfo{}, false
	}
	*pending = append(*pending, bindings...)
	return info, true
}

func collectStagedSummarizeBindings(root SyntaxElement) []stagedSummarizeBinding {
	bindings := make([]stagedSummarizeBinding, 0)
	seen := map[*SyntaxNode]bool{}
	var collect func(SyntaxElement)
	collect = func(element SyntaxElement) {
		if element == nil {
			return
		}
		if node := syntaxNodeForBinderLocation(element); node != nil && !seen[node] {
			seen[node] = true
			if info, ok := node.GetSemanticInfo(); ok {
				var result Symbol
				if value, found := node.Payload(functionCallResultSymbolPayload); found {
					result, _ = value.(Symbol)
				}
				bindings = append(bindings, stagedSummarizeBinding{node: node, info: info, resultSymbol: result})
			}
		}
		for index := 0; index < element.ChildCount(); index++ {
			collect(element.GetChild(index))
		}
	}
	collect(root)
	return bindings
}

func sameOpenTableInference(
	leftColumns map[*TableSymbol]map[string]*ColumnSymbol,
	leftOrder map[*TableSymbol][]*ColumnSymbol,
	rightColumns map[*TableSymbol]map[string]*ColumnSymbol,
	rightOrder map[*TableSymbol][]*ColumnSymbol,
) bool {
	if len(leftColumns) != len(rightColumns) || len(leftOrder) != len(rightOrder) {
		return false
	}
	for table, columns := range leftColumns {
		other, ok := rightColumns[table]
		if !ok || len(columns) != len(other) {
			return false
		}
		for name, column := range columns {
			if other[name] != column {
				return false
			}
		}
	}
	for table, columns := range leftOrder {
		other, ok := rightOrder[table]
		if !ok || len(columns) != len(other) {
			return false
		}
		for index, column := range columns {
			if other[index] != column {
				return false
			}
		}
	}
	return true
}

func (t *treeBinder) stageSummarizeDynamicJSONValue(expression Expression, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	switch expression := expression.(type) {
	case *LiteralExpression, *CompoundStringLiteralExpression, *PrefixUnaryExpression, *DynamicExpression, *NameReference:
		return t.stageSummarizeScalar(expression, false, pending)
	case *JsonArrayExpression:
		for _, separated := range expression.Values().Elements() {
			if separated == nil || separated.Element() == nil {
				return SemanticInfo{}, false
			}
			if _, ok := t.stageSummarizeDynamicJSONValue(separated.Element(), pending); !ok {
				return SemanticInfo{}, false
			}
		}
		tx := newUnionBindingTransaction(expression)
		for _, separated := range expression.Values().Elements() {
			valueInfo, ok := stagedSummarizeInfo(separated.Element(), *pending)
			if !ok {
				tx.rollback()
				return SemanticInfo{}, false
			}
			simpleBinderSyntaxNode(separated.Element()).SetSemanticInfo(valueInfo)
		}
		bound := t.nodeBinder.bindJsonArrayExpression(expression)
		tx.rollback()
		if bound == nil {
			return SemanticInfo{}, false
		}
		info := *bound
		*pending = append(*pending, stagedSummarizeBinding{node: expression.SyntaxNode, info: info})
		return info, true
	case *JsonObjectExpression:
		for _, separated := range expression.Pairs().Elements() {
			if separated == nil || separated.Element() == nil || separated.Element().Value() == nil {
				return SemanticInfo{}, false
			}
			if _, ok := t.stageSummarizeDynamicJSONValue(separated.Element().Value(), pending); !ok {
				return SemanticInfo{}, false
			}
		}
		tx := newUnionBindingTransaction(expression)
		for _, separated := range expression.Pairs().Elements() {
			value := separated.Element().Value()
			valueInfo, ok := stagedSummarizeInfo(value, *pending)
			if !ok {
				tx.rollback()
				return SemanticInfo{}, false
			}
			simpleBinderSyntaxNode(value).SetSemanticInfo(valueInfo)
		}
		bound := t.nodeBinder.bindJsonObjectExpression(expression)
		tx.rollback()
		if bound == nil {
			return SemanticInfo{}, false
		}
		info := *bound
		*pending = append(*pending, stagedSummarizeBinding{node: expression.SyntaxNode, info: info})
		return info, true
	default:
		return SemanticInfo{}, false
	}
}

func stagedSummarizeInfo(expression Expression, pending []stagedSummarizeBinding) (SemanticInfo, bool) {
	node := simpleBinderSyntaxNode(expression)
	for i := len(pending) - 1; i >= 0; i-- {
		if pending[i].node == node {
			return pending[i].info, true
		}
	}
	return SemanticInfo{}, false
}

func stagedSummarizePathScope(info SemanticInfo) Symbol {
	if symbol := info.ReferencedSymbol(); symbol != nil {
		if _, function := symbol.(*FunctionSymbol); function && info.ResultType != nil {
			return info.ResultType
		}
		if result := GetResultType(symbol); result != nil {
			return result
		}
		// Function-call semantic info references the function symbol while its
		// concrete invocation type is carried separately. Selectors must bind
		// against that result (for example parse_json(Column).property), not
		// against the function declaration itself.
		if info.ResultType != nil {
			return info.ResultType
		}
		return symbol
	}
	if info.ResultType != nil {
		return info.ResultType
	}
	return Error
}

func (t *treeBinder) stageSummarizeElementSelector(owner SyntaxElement, collection Expression, collectionInfo SemanticInfo, selector *BracketedExpression, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	indexInfo, ok := t.stageSummarizeScalar(selector.Expression(), false, pending)
	if !ok {
		return SemanticInfo{}, false
	}
	tx := newUnionBindingTransaction(owner)
	simpleBinderSyntaxNode(collection).SetSemanticInfo(collectionInfo)
	simpleBinderSyntaxNode(selector.Expression()).SetSemanticInfo(indexInfo)
	info := t.binder.bindElementSelector(collection, selector)
	tx.rollback()
	*pending = append(*pending, stagedSummarizeBinding{node: selector.SyntaxNode, info: info})
	return info, true
}

func (t *treeBinder) stageSummarizeOperator(node *SyntaxNode, location *SyntaxToken, arguments []Expression, pending *[]stagedSummarizeBinding) (SemanticInfo, bool) {
	types := make([]*TypeSymbol, len(arguments))
	constant := true
	for i, argument := range arguments {
		argumentInfo, ok := t.stageSummarizeScalar(argument, false, pending)
		if !ok {
			return SemanticInfo{}, false
		}
		types[i] = topHittersStagedResultType(argumentInfo)
		constant = constant && argumentInfo.IsConstant
	}
	bound := t.nodeBinder.operatorInfoForTypes(t.nodeBinder.operatorForKind(operatorKindForSyntax(node.Kind())), arguments, types, location, false, true)
	if bound == nil {
		return SemanticInfo{}, false
	}
	info := *bound
	info.IsConstant = constant
	return info, true
}

func isSummarizeStarArgument(expression Expression) bool {
	star, ok := expression.(*StarExpression)
	return ok && star != nil && star.AsteriskToken() != nil && star.AsteriskToken().Width() != 0
}

func stagedSummarizeResultColumn(expression Expression, pending []stagedSummarizeBinding) *ColumnSymbol {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return stagedSummarizeResultColumn(expression.Expression(), pending)
	case *SimpleNamedExpression:
		return stagedSummarizeResultColumn(expression.Expression(), pending)
	case *OrderedExpression:
		return stagedSummarizeResultColumn(expression.Expression(), pending)
	}
	node := simpleBinderSyntaxNode(expression)
	for i := len(pending) - 1; i >= 0; i-- {
		if pending[i].node == node {
			column, _ := pending[i].info.ReferencedSymbol().(*ColumnSymbol)
			return column
		}
	}
	return nil
}

func (t *treeBinder) visitTopHitters(node *TopHittersOperator) {
	if node == nil || node.TopHittersKeyword() == nil || node.TopHittersKeyword().Width() == 0 ||
		node.Expression() == nil || node.OfExpression() == nil ||
		(t.inPipe && t.binder.rowScope == nil) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	tx := newUnionBindingTransaction(node)
	committed := false
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		if !committed {
			tx.rollback()
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
		}
	}()

	if !t.supportsSummarizeScalar(node.Expression(), false) || !t.supportsSummarizeScalar(node.OfExpression(), false) ||
		(node.ByClause() != nil && !t.supportsSummarizeScalar(node.ByClause().Expression(), false)) {
		return
	}

	pending := make([]stagedSummarizeBinding, 0)
	if _, ok := t.stageSummarizeScalar(node.Expression(), false, &pending); !ok {
		return
	}
	if _, ok := t.stageSummarizeScalar(node.OfExpression(), false, &pending); !ok {
		return
	}
	if node.ByClause() != nil {
		if _, ok := t.stageSummarizeScalar(node.ByClause().Expression(), false, &pending); !ok {
			return
		}
	}
	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
		if binding.resultSymbol != nil {
			binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
		}
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func isSupportedTopHittersLiteral(expression *LiteralExpression) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind() {
	case SyntaxKindBooleanLiteralExpression,
		SyntaxKindIntLiteralExpression,
		SyntaxKindLongLiteralExpression,
		SyntaxKindRealLiteralExpression,
		SyntaxKindDecimalLiteralExpression,
		SyntaxKindDateTimeLiteralExpression,
		SyntaxKindTimespanLiteralExpression,
		SyntaxKindGuidLiteralExpression,
		SyntaxKindStringLiteralExpression,
		SyntaxKindNullLiteralExpression,
		SyntaxKindTokenLiteralExpression:
		return true
	default:
		return false
	}
}

func topHittersStagedResultType(info SemanticInfo) *TypeSymbol {
	if info.ResultType != nil {
		return info.ResultType
	}
	return &Error.TypeSymbol
}

func (t *treeBinder) visitSort(node *SortOperator) {
	if node == nil {
		return
	}
	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) visitFilter(node *FilterOperator) {
	if node == nil || node.Keyword() == nil || node.Keyword().Width() == 0 ||
		node.Parameters() == nil || node.Condition() == nil ||
		!t.supportsRecoverableFilterScalar(node.Condition()) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldNodeBinderInPipe := t.nodeBinder.inPipe
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	tx := newUnionBindingTransaction(node)
	committed := false
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		t.nodeBinder.inPipe = oldNodeBinderInPipe
		if !committed {
			tx.rollback()
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
		}
	}()

	for i := 0; i < node.ChildCount(); i++ {
		t.visitElement(node.GetChild(i))
	}
	if !expressionIsBound(node.Condition()) {
		return
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

// Binder_TreeBinder visits parser-recovered filter predicates and lets
// Binder_NodeBinder produce the operator result even when the predicate is an
// error. Keep rejecting unsupported hand-built trees that have no recovery
// diagnostics so transactional rollback remains available.
func (t *treeBinder) supportsRecoverableFilterScalar(expression Expression) bool {
	if t.supportsFilterScalar(expression) {
		return true
	}
	switch expression := expression.(type) {
	case *PathExpression:
		return parserRecoveredMissingPathSelector(expression) &&
			t.supportsRecoverableFilterScalar(expression.Expression())
	case *ParenthesizedExpression:
		return expression != nil && t.supportsRecoverableFilterScalar(expression.Expression())
	case *SimpleNamedExpression:
		return expression != nil && t.supportsRecoverableFilterScalar(expression.Expression())
	case *BinaryExpression:
		return expression != nil && expression.Operator() != nil && expression.Operator().Width() != 0 &&
			t.supportsRecoverableFilterScalar(expression.Left()) &&
			t.supportsRecoverableFilterScalar(expression.Right())
	case *InExpression, *HasAnyExpression, *HasAllExpression, *BetweenExpression, *PrefixUnaryExpression:
		return supportsScalarOperator(simpleBinderSyntaxNode(expression))
	default:
		return false
	}
}

func (t *treeBinder) supportsFilterScalar(expression Expression) bool {
	switch expression := expression.(type) {
	case *StarExpression:
		return expression != nil && expression.AsteriskToken() != nil && expression.AsteriskToken().Width() != 0
	case *ParenthesizedExpression:
		return expression != nil && t.supportsFilterScalar(expression.Expression())
	case *SimpleNamedExpression:
		return expression != nil && t.supportsFilterScalar(expression.Expression())
	case *BinaryExpression:
		return expression != nil && t.supportsFilterScalar(expression.Left()) && t.supportsFilterScalar(expression.Right())
	default:
		return t.supportsSummarizeScalar(expression, false)
	}
}

func (t *treeBinder) visitProjectReorder(node *ProjectReorderOperator) {
	if node == nil || node.ProjectReorderKeyword() == nil || node.ProjectReorderKeyword().Width() == 0 ||
		node.Expressions() == nil || (t.inPipe && t.binder.rowScope == nil) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
	}()

	type pendingReference struct {
		reference *NameReference
		ordered   *OrderedExpression
		info      SemanticInfo
	}
	pending := make([]pendingReference, 0, node.Expressions().Count())
	for _, separated := range node.Expressions().Elements() {
		if separated == nil {
			return
		}
		expression := separated.Element()
		if syntaxElementIsMissing(expression) {
			continue
		}
		var ordered *OrderedExpression
		if candidate, ok := expression.(*OrderedExpression); ok {
			ordered = candidate
			if !isSupportedProjectReorderOrdering(ordered.Ordering()) {
				return
			}
			expression = ordered.Expression()
		}
		reference, ok := expression.(*NameReference)
		if !ok || reference == nil {
			return
		}
		info, resolved := t.resolveProjectColumnReference(reference)
		if !resolved {
			pending = append(pending, pendingReference{reference: reference, ordered: ordered, info: info})
			continue
		}
		if !projectReorderReferenceIsColumn(info.ReferencedSymbol()) {
			return
		}
		pending = append(pending, pendingReference{reference: reference, ordered: ordered, info: info})
	}

	for _, binding := range pending {
		binding.reference.SetSemanticInfo(binding.info)
		if binding.ordered != nil {
			binding.ordered.SetSemanticInfo(binding.info)
		}
	}
	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) visitProjectByNames(node *ProjectByNamesOperator) {
	if node == nil || node.ProjectByNamesKeyword() == nil || node.ProjectByNamesKeyword().Width() == 0 ||
		node.Expressions() == nil {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	tx := newUnionBindingTransaction(node)
	committed := false
	var failedNode *SyntaxNode
	var failedInfo *SemanticInfo
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		if !committed {
			tx.rollback()
			if failedNode != nil && failedInfo != nil {
				failedNode.SetSemanticInfo(*failedInfo)
			}
		}
	}()

	pendingReferences := map[*NameReference]SemanticInfo{}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || separated.Element() == nil {
			return
		}
		if reference, ok := separated.Element().(*NameReference); ok {
			if syntaxElementIsMissing(reference) {
				pendingReferences[reference] = NewSemanticInfoForReferencedSymbol(Error)
				continue
			}
			if !isCompleteProjectByNamesReference(reference) {
				return
			}
			info := t.binder.ResolveNameReferenceInRowScope(t.binder.rowScope, reference, false)
			if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); unresolved || len(info.Diagnostics) != 0 {
				failedNode = reference.SyntaxNode
				failedInfo = &info
				return
			}
			pendingReferences[reference] = info
		}
	}

	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		if reference, ok := expression.(*NameReference); ok {
			reference.SetSemanticInfo(pendingReferences[reference])
		} else {
			t.visitElement(expression)
			if !expressionIsBound(expression) {
				return
			}
		}
	}

	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func (t *treeBinder) visitDistinct(node *DistinctOperator) {
	if node == nil || node.DistinctKeyword() == nil || node.DistinctKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expressions() == nil || node.Expressions().Count() == 0 ||
		(t.inPipe && t.binder.rowScope == nil) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	tx := newUnionBindingTransaction(node)
	committed := false
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		if !committed {
			tx.rollback()
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
		}
	}()

	pending := make([]stagedSummarizeBinding, 0)

	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || NameDeclarationSimpleName(parameter.Name()) == "" || parameter.Expression() == nil ||
			!t.supportsSummarizeScalar(parameter.Expression(), false) {
			return
		}
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || separated.Element() == nil {
			return
		}
		switch expression := separated.Element().(type) {
		case *StarExpression:
			if expression.Width() == 0 {
				return
			}
		default:
			if !t.supportsSummarizeScalar(expression, false) {
				return
			}
		}
	}

	for _, parameter := range node.Parameters().Elements() {
		if _, ok := t.stageSummarizeScalar(parameter.Expression(), false, &pending); !ok {
			return
		}
	}
	for _, separated := range node.Expressions().Elements() {
		if _, star := separated.Element().(*StarExpression); star {
			continue
		}
		if _, ok := t.stageSummarizeScalar(separated.Element(), false, &pending); !ok {
			return
		}
	}
	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
		if binding.resultSymbol != nil {
			binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
		}
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func (t *treeBinder) visitSampleDistinct(node *SampleDistinctOperator) {
	if node == nil || node.SampleDistinctKeyword() == nil || node.SampleDistinctKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expression() == nil || node.OfExpression() == nil ||
		(t.inPipe && t.binder.rowScope == nil) {
		return
	}

	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldInWildcard := t.inWildcard
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	tx := newUnionBindingTransaction(node)
	committed := false
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldInWildcard
		if !committed {
			tx.rollback()
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
		}
	}()

	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || parameter.Name() == nil || NameDeclarationSimpleName(parameter.Name()) == "" ||
			parameter.Expression() == nil || !t.supportsSummarizeScalar(parameter.Expression(), false) {
			return
		}
	}
	if !t.supportsSummarizeScalar(node.Expression(), false) || !t.supportsSummarizeScalar(node.OfExpression(), false) {
		return
	}

	pending := make([]stagedSummarizeBinding, 0)
	for _, parameter := range node.Parameters().Elements() {
		if _, ok := t.stageSummarizeScalar(parameter.Expression(), false, &pending); !ok {
			return
		}
	}
	if _, ok := t.stageSummarizeScalar(node.Expression(), false, &pending); !ok {
		return
	}
	if _, ok := t.stageSummarizeScalar(node.OfExpression(), false, &pending); !ok {
		return
	}
	for _, binding := range pending {
		binding.node.SetSemanticInfo(binding.info)
		if binding.resultSymbol != nil {
			binding.node.setPayload(functionCallResultSymbolPayload, binding.resultSymbol)
		}
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func isCompleteProjectByNamesReference(reference *NameReference) bool {
	if reference == nil || reference.Name() == nil || NameReferenceSimpleName(reference) == "" {
		return false
	}
	switch name := reference.Name().(type) {
	case *TokenName:
		return name.Name() != nil && name.Name().Width() != 0
	case *BracketedName:
		return name.OpenBracket() != nil && name.OpenBracket().Width() != 0 && name.Name() != nil && name.Name().Width() != 0 && name.CloseBracket() != nil && name.CloseBracket().Width() != 0
	default:
		return false
	}
}

func isSupportedProjectReorderOrdering(ordering *OrderingClause) bool {
	if ordering == nil || ordering.NullsClause() != nil || len(ordering.ContainedDiagnostics()) != 0 {
		return false
	}
	keyword := ordering.AscOrDescKeyword()
	if keyword == nil || keyword.Width() == 0 {
		return false
	}
	switch keyword.Kind() {
	case SyntaxKindAscKeyword, SyntaxKindDescKeyword, SyntaxKindGrannyAscKeyword, SyntaxKindGrannyDescKeyword:
		return true
	default:
		return false
	}
}

func projectReorderReferenceIsColumn(symbol Symbol) bool {
	switch symbol := symbol.(type) {
	case *ColumnSymbol:
		return true
	case *GroupSymbol:
		for _, member := range symbol.Members() {
			if _, ok := member.(*ColumnSymbol); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (t *treeBinder) visitProjectColumns(node *SyntaxNode, expressions *SeparatedSyntaxListOf[Expression]) {
	if node == nil || expressions == nil || (t.inPipe && t.binder.rowScope == nil) {
		return
	}

	type pendingReference struct {
		reference *NameReference
		info      SemanticInfo
	}
	pending := make([]pendingReference, 0, expressions.Count())
	for _, separated := range expressions.Elements() {
		if separated == nil {
			return
		}
		reference, ok := separated.Element().(*NameReference)
		if !ok || reference == nil {
			return
		}
		if syntaxElementIsMissing(reference) {
			pending = append(pending, pendingReference{reference: reference, info: NewSemanticInfoForReferencedSymbol(Error)})
			continue
		}
		info, resolved := t.resolveProjectColumnReference(reference)
		_ = resolved
		pending = append(pending, pendingReference{reference: reference, info: info})
	}

	for _, binding := range pending {
		binding.reference.SetSemanticInfo(binding.info)
	}
	t.bindNode(node)
}

func (t *treeBinder) resolveProjectColumnReference(reference *NameReference) (SemanticInfo, bool) {
	name := NameReferenceSimpleName(reference)
	if isProjectColumnWildcard(reference) {
		oldWildcard := t.inWildcard
		t.inWildcard = true
		defer func() { t.inWildcard = oldWildcard }()

		var symbols []Symbol
		if t.binder.rowScope != nil {
			for _, column := range t.binder.rowScope.Columns() {
				if column != nil && KustoFactsMatches(name, column.SymbolName()) {
					symbols = append(symbols, column)
				}
			}
		}
		if len(symbols) == 0 && (t.binder.rowScope == nil || !t.binder.rowScope.IsOpen()) {
			diagnostic := t.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownItem(name, ""), reference)
			return NewSemanticInfoForReferencedSymbol(Error, diagnostic), false
		}
		return projectColumnReferenceInfo(symbols), true
	}

	var symbols []Symbol
	if t.binder.rowScope != nil {
		for _, column := range t.binder.rowScope.Columns() {
			if column != nil && column.SymbolName() == name {
				symbols = append(symbols, column)
			}
		}
	}
	if len(symbols) != 0 {
		return projectColumnReferenceInfo(symbols), true
	}
	if known := t.binder.matchingNameSymbols(name, SymbolMatchDefault); len(known) != 0 {
		return projectColumnReferenceInfo(known), true
	}
	diagnostic := t.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownItem(name, ""), reference)
	return NewSemanticInfoForReferencedSymbol(Error, diagnostic), false
}

func projectColumnReferenceInfo(symbols []Symbol) SemanticInfo {
	if len(symbols) == 1 {
		return NewSemanticInfoForReferencedSymbol(symbols[0])
	}
	group := NewGroupSymbol(symbols...)
	return NewSemanticInfoForSymbol(group, &group.TypeSymbol)
}

func isProjectColumnWildcard(reference *NameReference) bool {
	if reference == nil {
		return false
	}
	switch reference.Name().(type) {
	case *WildcardedName, *BracketedWildcardedName:
		return true
	default:
		return false
	}
}

func isCompleteProjectColumnName(reference *NameReference) bool {
	if reference == nil || reference.Name() == nil || NameReferenceSimpleName(reference) == "" {
		return false
	}
	switch name := reference.Name().(type) {
	case *TokenName:
		return name.Name() != nil && name.Name().Width() != 0
	case *BracketedName:
		return name.OpenBracket() != nil && name.OpenBracket().Width() != 0 && name.Name() != nil && name.Name().Width() != 0 && name.CloseBracket() != nil && name.CloseBracket().Width() != 0
	case *WildcardedName:
		return name.Pattern() != nil && name.Pattern().Width() != 0
	case *BracketedWildcardedName:
		return name.OpenBracket() != nil && name.OpenBracket().Width() != 0 && name.Pattern() != nil && name.Pattern().Width() != 0 && name.CloseBracket() != nil && name.CloseBracket().Width() != 0
	default:
		return false
	}
}

func (t *treeBinder) visitPipeExpression(node *PipeExpression) {
	if consume, ok := node.Operator().(*ConsumeOperator); ok && !supportsConsumeOperator(consume) {
		return
	}
	if executeAndCache, ok := node.Operator().(*ExecuteAndCacheOperator); ok && !supportsExecuteAndCacheOperator(executeAndCache) {
		return
	}
	if union, ok := node.Operator().(*UnionOperator); ok && !supportsUnionOperator(union) {
		return
	}
	if join, ok := node.Operator().(*JoinOperator); ok && !t.supportsJoinOperator(join) {
		return
	}
	if lookup, ok := node.Operator().(*LookupOperator); ok && !t.supportsLookupOperator(lookup) {
		return
	}
	if mvExpand, ok := node.Operator().(*MvExpandOperator); ok && !supportsMvExpandOperator(mvExpand) {
		return
	}
	outerSourceScopeKind := t.binder.scopeKind
	t.binder.scopeKind = ScopeKindNormal
	if reference, ok := node.Expression().(*NameReference); ok {
		reference.SetSemanticInfo(t.binder.ResolveNameReferenceInRowScope(t.binder.rowScope, reference, true))
	} else {
		t.visitElement(node.Expression())
	}
	t.bindImplicitFunctionSource(node.Expression())
	t.binder.scopeKind = outerSourceScopeKind
	if containsFailedBoundedOperator(node.Expression()) {
		return
	}
	oldRowScope := t.binder.rowScope
	oldScopeKind := t.binder.scopeKind
	oldInPipe := t.inPipe
	incomingRowScope := expressionTableResult(node.Expression())
	if _, isAs := node.Operator().(*AsOperator); isAs && !expressionHasResolvedTableResult(node.Expression()) {
		incomingRowScope = nil
	}
	func() {
		defer func() {
			t.nodeBinder.pipedAsRowScope = nil
			t.binder.rowScope = oldRowScope
			t.binder.scopeKind = oldScopeKind
			t.inPipe = oldInPipe
		}()
		t.binder.rowScope = incomingRowScope
		t.binder.scopeKind = ScopeKindNormal
		t.inPipe = true
		if _, isAs := node.Operator().(*AsOperator); isAs {
			t.nodeBinder.pipedAsRowScope = incomingRowScope
		}
		switch operator := node.Operator().(type) {
		case *MakeGraphOperator:
			t.visitMakeGraph(operator)
		case *GraphMergeOperator:
			t.visitGraphMerge(operator, expressionGraphResult(node.Expression()))
		case *GraphMatchOperator:
			t.visitGraphMatch(operator, expressionGraphResult(node.Expression()))
		case *GraphShortestPathsOperator:
			t.visitGraphShortestPaths(operator, expressionGraphResult(node.Expression()))
		case *GraphMarkComponentsOperator:
			t.visitGraphMarkComponents(operator, expressionGraphResult(node.Expression()))
		case *GraphWhereNodesOperator:
			t.visitGraphWhereNodes(operator, expressionGraphResult(node.Expression()))
		case *GraphWhereEdgesOperator:
			t.visitGraphWhereEdges(operator, expressionGraphResult(node.Expression()))
		case *GraphToTableOperator:
			t.visitGraphToTable(operator, expressionGraphResult(node.Expression()))
		default:
			t.visitElement(operator)
		}
	}()
	if union, ok := node.Operator().(*UnionOperator); ok && !union.IsBound() {
		return
	}
	if join, ok := node.Operator().(*JoinOperator); ok && !join.IsBound() {
		return
	}
	if lookup, ok := node.Operator().(*LookupOperator); ok && !lookup.IsBound() {
		return
	}
	if mvExpand, ok := node.Operator().(*MvExpandOperator); ok && !mvExpand.IsBound() {
		return
	}
	if executeAndCache, ok := node.Operator().(*ExecuteAndCacheOperator); ok && !executeAndCache.IsBound() {
		return
	}
	if consume, ok := node.Operator().(*ConsumeOperator); ok && !consume.IsBound() {
		return
	}

	if asOperator, ok := node.Operator().(*AsOperator); ok && asOperator.IsBound() {
		t.publishAsAlias(asOperator, incomingRowScope)
	}

	t.bindNode(node.SyntaxNode)
}

func (t *treeBinder) bindImplicitFunctionSource(expression Expression) {
	reference, ok := expression.(*NameReference)
	if !ok || reference == nil {
		return
	}
	function, ok := reference.ReferencedSymbol().(*FunctionSymbol)
	if !ok || len(function.Signatures()) != 1 {
		return
	}
	signature := function.Signatures()[0]
	if signature == nil || signature.MinArgumentCount != 0 {
		return
	}
	result, callInfo := t.nodeBinder.bindComputedFunction(signature, nil, nil, nil)
	if result == nil || callInfo == nil {
		return
	}
	reference.setPayload(functionCallResultSymbolPayload, result)
	reference.SetSemanticInfo(NewSemanticInfoForSignature(signature, returnTypeFromSymbol(result)).WithCalledFunctionInfo(callInfo))
}

func containsFailedBoundedOperator(expression Expression) bool {
	if containsFailedRange(expression) {
		return true
	}
	switch expression := expression.(type) {
	case *DataTableExpression:
		return !expression.IsBound()
	case *ContextualDataTableExpression:
		return !expression.IsBound()
	case *ExternalDataExpression:
		return !expression.IsBound()
	case *InlineExternalTableExpression:
		return !expression.IsBound()
	case *ConsumeOperator:
		return !expression.IsBound()
	case *ExecuteAndCacheOperator:
		return !expression.IsBound()
	case *UnionOperator:
		return !expression.IsBound()
	case *JoinOperator:
		return !expression.IsBound()
	case *LookupOperator:
		return !expression.IsBound()
	case *MvExpandOperator:
		return !expression.IsBound()
	case *ParenthesizedExpression:
		return containsFailedBoundedOperator(expression.Expression())
	case *PipeExpression:
		if containsFailedBoundedOperator(expression.Expression()) {
			return true
		}
		switch operator := expression.Operator().(type) {
		case *ConsumeOperator:
			return !operator.IsBound()
		case *ExecuteAndCacheOperator:
			return !operator.IsBound()
		case *UnionOperator:
			return !operator.IsBound()
		case *JoinOperator:
			return !operator.IsBound()
		case *LookupOperator:
			return !operator.IsBound()
		case *MvExpandOperator:
			return !operator.IsBound()
		default:
			return false
		}
	default:
		return false
	}
}

func containsFailedRange(expression Expression) bool {
	switch expression := expression.(type) {
	case *RangeOperator:
		return !expression.IsBound()
	case *PipeExpression:
		if expression.IsBound() {
			return false
		}
		if containsFailedRange(expression.Expression()) {
			return true
		}
		rangeOperator, ok := expression.Operator().(*RangeOperator)
		return ok && !rangeOperator.IsBound()
	default:
		return false
	}
}

func (t *treeBinder) publishAsAlias(node *AsOperator, incomingRowScope *TableSymbol) {
	if t == nil || t.binder == nil || node == nil || incomingRowScope == nil {
		return
	}
	name := NameDeclarationSimpleName(node.Name())
	if name == "" {
		return
	}
	variable := NewVariableSymbol(name, incomingRowScope, false, nil, nil)
	if t.binder.LocalScope == nil {
		t.binder.LocalScope = NewLocalScope(nil)
	}
	t.binder.LocalScope.AddSymbol(variable)
	node.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(variable))
}

func (t *treeBinder) bindNode(node *SyntaxNode) {
	if t != nil && t.binder != nil {
		t.binder.cancellationToken.ThrowIfCancellationRequested()
	}
	if t.inPipe && t.binder.rowScope == nil {
		switch WrapGeneratedSyntaxNode(node).(type) {
		case *SampleDistinctOperator, *TopHittersOperator:
			return
		}
	}
	t.nodeBinder.inPipe = t.inPipe
	if info := t.nodeBinder.bind(node); info != nil {
		t.binder.cancellationToken.ThrowIfCancellationRequested()
		node.SetSemanticInfo(*info)
	}
}

func expressionTableResult(expression Expression) *TableSymbol {
	if expression == nil {
		return nil
	}
	if call, ok := expression.(*FunctionCallExpression); ok {
		if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok {
				if table, ok := tableSymbolFromSymbol(symbol); ok {
					return table
				}
			}
		}
	}
	if reference, ok := expression.(*NameReference); ok {
		if result, found := reference.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok {
				if table, ok := tableSymbolFromSymbol(symbol); ok {
					return table
				}
			}
		}
	}
	accessor, ok := any(expression).(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	})
	if !ok {
		return nil
	}
	info, ok := accessor.GetSemanticInfo()
	if !ok {
		return nil
	}
	switch symbol := info.ReferencedSymbol().(type) {
	case *TableSymbol:
		return symbol
	case *VariableSymbol:
		table, _ := tableSymbolFromSymbol(GetResultType(symbol.Type))
		return table
	case *ParameterSymbol:
		table, _ := tableSymbolFromSymbol(GetResultType(symbol.Type))
		return table
	case *EntityGroupElementSymbol:
		table, _ := tableSymbolFromSymbol(symbol.UnderlyingSymbol)
		return table
	default:
		table, _ := TableSymbolForType(info.ResultType)
		return table
	}
}

func expressionHasResolvedTableResult(expression Expression) bool {
	if expression == nil {
		return false
	}
	accessor, ok := any(expression).(interface {
		GetSemanticInfo() (SemanticInfo, bool)
	})
	if !ok {
		return false
	}
	_, ok = accessor.GetSemanticInfo()
	return ok && expressionTableResult(expression) != nil
}

func simpleBinderSyntaxNode(element SyntaxElement) *SyntaxNode {
	switch element := element.(type) {
	case *CommandBlock:
		return element.SyntaxNode
	case *QueryBlock:
		return element.SyntaxNode
	case *DirectiveBlock:
		return element.SyntaxNode
	case *Directive:
		return element.SyntaxNode
	case *CustomCommand:
		return element.SyntaxNode
	case *AliasStatement:
		return element.SyntaxNode
	case *LetStatement:
		return element.SyntaxNode
	case *FunctionDeclaration:
		return element.SyntaxNode
	case *PatternStatement:
		return element.SyntaxNode
	case *PatternDeclaration:
		return element.SyntaxNode
	case *QueryParametersStatement:
		return element.SyntaxNode
	case *SetOptionStatement:
		return element.SyntaxNode
	case *OptionValueClause:
		return element.SyntaxNode
	case *FunctionParameter:
		return element.SyntaxNode
	case *NameReference:
		return element.SyntaxNode
	case *LiteralExpression:
		return element.SyntaxNode
	case *TypeOfLiteralExpression:
		return element.SyntaxNode
	case *CompoundStringLiteralExpression:
		return element.SyntaxNode
	case *StarExpression:
		return element.SyntaxNode
	case *TypedColumnReference:
		return element.SyntaxNode
	case *ExpressionList:
		return element.SyntaxNode
	case *PatternPathValue:
		return element.SyntaxNode
	case *DynamicExpression:
		return element.SyntaxNode
	case *JsonArrayExpression:
		return element.SyntaxNode
	case *JsonObjectExpression:
		return element.SyntaxNode
	case *JsonPair:
		return element.SyntaxNode
	case *PrimitiveTypeExpression:
		return element.SyntaxNode
	case *SchemaTypeExpression:
		return element.SyntaxNode
	case *ToScalarExpression:
		return element.SyntaxNode
	case *ToTableExpression:
		return element.SyntaxNode
	case *DataTableExpression:
		return element.SyntaxNode
	case *ContextualDataTableExpression:
		return element.SyntaxNode
	case *ExternalDataExpression:
		return element.SyntaxNode
	case *InlineExternalTableExpression:
		return element.SyntaxNode
	case *MaterializeExpression:
		return element.SyntaxNode
	case *MaterializedViewCombineExpression:
		return element.SyntaxNode
	case *MaterializedViewCombineNameClause:
		return element.SyntaxNode
	case *MaterializedViewCombineClause:
		return element.SyntaxNode
	case *EntityGroup:
		return element.SyntaxNode
	case *MacroExpandOperator:
		return element.SyntaxNode
	case *DateTimePattern:
		return element.SyntaxNode
	case *NameAndTypeDeclaration:
		return element.SyntaxNode
	case *NameDeclaration:
		return element.SyntaxNode
	case *RenameList:
		return element.SyntaxNode
	case *ParenthesizedExpression:
		return element.SyntaxNode
	case *BracketedExpression:
		return element.SyntaxNode
	case *PathExpression:
		return element.SyntaxNode
	case *ElementExpression:
		return element.SyntaxNode
	case *DataScopeExpression:
		return element.SyntaxNode
	case *BinaryExpression:
		return element.SyntaxNode
	case *PrefixUnaryExpression:
		return element.SyntaxNode
	case *InExpression:
		return element.SyntaxNode
	case *HasAnyExpression:
		return element.SyntaxNode
	case *HasAllExpression:
		return element.SyntaxNode
	case *BetweenExpression:
		return element.SyntaxNode
	case *SimpleNamedExpression:
		return element.SyntaxNode
	case *CompoundNamedExpression:
		return element.SyntaxNode
	case *OrderedExpression:
		return element.SyntaxNode
	case *PipeExpression:
		return element.SyntaxNode
	case *BadQueryOperator:
		return element.SyntaxNode
	case *ForkExpression:
		return element.SyntaxNode
	case *ForkOperator:
		return element.SyntaxNode
	case *FacetOperator:
		return element.SyntaxNode
	case *PartitionOperator:
		return element.SyntaxNode
	case *PartitionByOperator:
		return element.SyntaxNode
	case *PartitionQuery:
		return element.SyntaxNode
	case *PartitionSubquery:
		return element.SyntaxNode
	case *InvokeOperator:
		return element.SyntaxNode
	case *ParseOperator:
		return element.SyntaxNode
	case *ParseWhereOperator:
		return element.SyntaxNode
	case *ParseKvOperator:
		return element.SyntaxNode
	case *PrintOperator:
		return element.SyntaxNode
	case *RangeOperator:
		return element.SyntaxNode
	case *UnionOperator:
		return element.SyntaxNode
	case *SearchOperator:
		return element.SyntaxNode
	case *FindOperator:
		return element.SyntaxNode
	case *EvaluateOperator:
		return element.SyntaxNode
	case *RenderOperator:
		return element.SyntaxNode
	case *ReduceByOperator:
		return element.SyntaxNode
	case *MakeSeriesOperator:
		return element.SyntaxNode
	case *MakeSeriesExpression:
		return element.SyntaxNode
	case *MakeGraphOperator:
		return element.SyntaxNode
	case *GraphMatchOperator:
		return element.SyntaxNode
	case *GraphShortestPathsOperator:
		return element.SyntaxNode
	case *GraphMarkComponentsOperator:
		return element.SyntaxNode
	case *GraphMergeOperator:
		return element.SyntaxNode
	case *GraphWhereNodesOperator:
		return element.SyntaxNode
	case *GraphWhereEdgesOperator:
		return element.SyntaxNode
	case *GraphToTableOperator:
		return element.SyntaxNode
	case *JoinOperator:
		return element.SyntaxNode
	case *LookupOperator:
		return element.SyntaxNode
	case *ExecuteAndCacheOperator:
		return element.SyntaxNode
	case *ConsumeOperator:
		return element.SyntaxNode
	case *MvExpandOperator:
		return element.SyntaxNode
	case *MvApplyOperator:
		return element.SyntaxNode
	case *MvApplyExpression:
		return element.SyntaxNode
	case *MvApplySubqueryExpression:
		return element.SyntaxNode
	case *AsOperator:
		return element.SyntaxNode
	case *CountOperator:
		return element.SyntaxNode
	case *TakeOperator:
		return element.SyntaxNode
	case *SampleOperator:
		return element.SyntaxNode
	case *SampleDistinctOperator:
		return element.SyntaxNode
	case *TopOperator:
		return element.SyntaxNode
	case *TopHittersOperator:
		return element.SyntaxNode
	case *SortOperator:
		return element.SyntaxNode
	case *SerializeOperator:
		return element.SyntaxNode
	case *FilterOperator:
		return element.SyntaxNode
	case *ProjectOperator:
		return element.SyntaxNode
	case *ProjectAwayOperator:
		return element.SyntaxNode
	case *ProjectKeepOperator:
		return element.SyntaxNode
	case *ProjectReorderOperator:
		return element.SyntaxNode
	case *ProjectByNamesOperator:
		return element.SyntaxNode
	case *DistinctOperator:
		return element.SyntaxNode
	case *SummarizeOperator:
		return element.SyntaxNode
	case *TopNestedOperator:
		return element.SyntaxNode
	case *FunctionCallExpression:
		return element.SyntaxNode
	case *AssertSchemaOperator:
		return element.SyntaxNode
	case *GetSchemaOperator:
		return element.SyntaxNode
	case *ScanOperator:
		return element.SyntaxNode
	case *ExtendOperator:
		return element.SyntaxNode
	case *ProjectRenameOperator:
		return element.SyntaxNode
	case *ExpressionStatement:
		return element.SyntaxNode
	default:
		return nil
	}
}
