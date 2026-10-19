package kustolang

// Ported from the MacroExpandOperator portions of Binder_TreeBinder.cs and
// Binder_NodeBinder.cs.

func (t *treeBinder) visitEntityGroup(node *EntityGroup) {
	if node == nil || node.IsBound() || node.Entities() == nil {
		return
	}
	tx := newUnionBindingTransaction(node)
	for _, item := range node.Entities().Elements() {
		if item == nil || item.Element() == nil {
			tx.rollback()
			return
		}
		t.visitElement(item.Element())
		if syntax := simpleBinderSyntaxNode(item.Element()); syntax != nil && !syntax.IsBound() {
			tx.rollback()
			return
		}
	}
	t.bindNode(node.SyntaxNode)
}

func (n *simpleNodeBinder) bindEntityGroup(node *EntityGroup) *SemanticInfo {
	if node == nil || node.Entities() == nil {
		return nil
	}
	members := make([]Symbol, 0, node.Entities().Count())
	diagnostics := make([]Diagnostic, 0)
	var expectedKind SymbolKind
	hasExpectedKind := false
	for _, item := range node.Entities().Elements() {
		if item == nil || item.Element() == nil {
			continue
		}
		expression := item.Element()
		symbol := kustoCodeExpressionResultSymbol(expression)
		if kind, valid := entityGroupMemberKind(symbol); valid {
			if !hasExpectedKind {
				expectedKind = kind
				hasExpectedKind = true
				members = append(members, symbol)
				continue
			}
			if kind == expectedKind {
				members = append(members, symbol)
				continue
			}
		}
		if hasExpectedKind {
			var diagnostic Diagnostic
			switch expectedKind {
			case SymbolKindTable:
				diagnostic = DiagnosticTableExpected()
			case SymbolKindDatabase:
				diagnostic = DiagnosticDatabaseExpected()
			case SymbolKindCluster:
				diagnostic = DiagnosticClusterExpected()
			}
			if diagnostic.Code != "" {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(diagnostic, expression))
			}
		} else {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticClusterDatabaseOrTableExpected(), expression))
		}
	}
	group := NewEntityGroupSymbol("", members...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(group, &group.TypeSymbol, diagnostics...))
}

func entityGroupMemberKind(symbol Symbol) (SymbolKind, bool) {
	switch symbol.(type) {
	case *TableSymbol, *MaterializedViewSymbol, *ExternalTableSymbol, *StoredQueryResultSymbol:
		return SymbolKindTable, true
	case *DatabaseSymbol:
		return SymbolKindDatabase, true
	case *ClusterSymbol:
		return SymbolKindCluster, true
	default:
		return SymbolKindNone, false
	}
}

func (t *treeBinder) visitMacroExpand(node *MacroExpandOperator) {
	if node == nil || node.IsBound() || node.EntityGroup() == nil || node.StatementList() == nil {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldRowScope := t.binder.rowScope
	oldPathScope := t.binder.pathScope
	oldScopeKind := t.binder.scopeKind
	oldLocalScope := t.binder.LocalScope
	oldInPipe := t.inPipe
	oldWildcard := t.inWildcard
	oldNodeInPipe := t.nodeBinder.inPipe
	oldPiped := t.nodeBinder.pipedAsRowScope
	oldIsFuzzy := t.binder.isFuzzy
	oldInferredColumns := cloneOpenTableInferredColumns(t.binder.openTableInferredColumns)
	oldInferredOrder := cloneOpenTableInferredOrder(t.binder.openTableInferredOrder)
	defer func() {
		t.binder.rowScope = oldRowScope
		t.binder.pathScope = oldPathScope
		t.binder.scopeKind = oldScopeKind
		t.binder.LocalScope = oldLocalScope
		t.inPipe = oldInPipe
		t.inWildcard = oldWildcard
		t.nodeBinder.inPipe = oldNodeInPipe
		t.nodeBinder.pipedAsRowScope = oldPiped
		t.binder.isFuzzy = oldIsFuzzy
		if !committed {
			t.binder.openTableInferredColumns = oldInferredColumns
			t.binder.openTableInferredOrder = oldInferredOrder
			tx.rollback()
		}
	}()

	// Parameters and the group are bound in the enclosing scope.
	t.visitElement(node.Parameters())
	if fuzzy, specified := queryOperatorBooleanParameterValue(node.Parameters(), "isfuzzy"); specified {
		t.binder.isFuzzy = fuzzy
	}
	t.visitElement(node.EntityGroup())
	// Fuzziness applies only to entity evaluation, not to the macro body.
	t.binder.isFuzzy = false
	group, ok := expressionReferencedSymbol(node.EntityGroup()).(*EntityGroupSymbol)
	if !ok {
		if variable, variableOK := expressionReferencedSymbol(node.EntityGroup()).(*VariableSymbol); variableOK {
			group, ok = variable.Type.(*EntityGroupSymbol)
		}
	}
	if !ok || group == nil {
		t.binder.LocalScope = NewLocalScope(oldLocalScope)
		t.binder.LocalScope.AddSymbol(NewEntityGroupElementSymbol(macroExpandScopeName(node, nil), nil, nil))
		for _, statement := range node.StatementList().Elements() {
			if statement != nil && statement.Element() != nil {
				t.visitElement(statement.Element())
			}
		}
		result := NewTableSymbol("")
		node.SetSemanticInfo(NewSemanticInfoForSymbol(result, &result.TypeSymbol))
		committed = true
		return
	}

	scopeName := group.SymbolName()
	var declaredScope *EntityGroupElementSymbol
	if clause := node.ScopeReferenceName(); clause != nil && clause.EntityGroupReferenceName() != nil {
		declaration := clause.EntityGroupReferenceName()
		scopeName = NameDeclarationSimpleName(declaration)
		declaredScope = NewEntityGroupElementSymbol(scopeName, group, nil)
		declaration.SetSemanticInfo(NewSemanticInfoForSymbol(declaredScope, &declaredScope.TypeSymbol))
	}
	tables := make([][]*ColumnSymbol, 0, len(group.Members()))
	alternates := make([]*SyntaxNode, 0, len(group.Members()))
	memberDiagnostics := make([]Diagnostic, 0)
	bodies := make([]*SeparatedSyntaxListOf[Statement], len(group.Members()))
	original := node.Tree()
	offset := 0
	if original != nil {
		if list := node.StatementList().SyntaxList; list != nil && list.Tree() == original {
			offset = list.ByteTriviaStart()
		}
	}
	for index, member := range group.Members() {
		if declaredScope != nil && declaredScope.UnderlyingSymbol == member {
			bodies[index] = node.StatementList()
			continue
		}
		clonedList, _ := node.StatementList().SyntaxList.Clone(true).(*SyntaxList)
		bodies[index] = AsSeparatedSyntaxListOf[Statement](clonedList)
		alternate := NewSyntaxNode(SyntaxKindList, []SyntaxElement{clonedList})
		NewSyntaxTree(alternate, original, offset)
		t.binder.appendSyntaxLocations(clonedList)
		alternates = append(alternates, alternate)
	}
	for index, member := range group.Members() {
		body := bodies[index]
		local := NewLocalScope(oldLocalScope)
		current := NewEntityGroupElementSymbol(scopeName, group, member)
		local.AddSymbol(current)
		t.binder.LocalScope = local
		t.binder.rowScope = nil
		t.binder.pathScope = nil
		t.inPipe = false
		t.nodeBinder.inPipe = false
		for _, statement := range body.Elements() {
			if statement != nil && statement.Element() != nil {
				t.visitElement(statement.Element())
			}
		}
		result := macroExpandBodyResult(body)
		if result == nil {
			continue
		}
		typ := ExpressionRawResultType(result)
		if table := exactExpressionTable(result); table != nil {
			tables = append(tables, table.Columns())
		} else if typ != nil && typ != &Error.TypeSymbol {
			diagnostic := t.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), result)
			if body.SyntaxList != node.StatementList().SyntaxList && diagnostic.HasLocation() {
				if resultNode := simpleBinderSyntaxNode(result); resultNode != nil {
					diagnostic = diagnostic.WithByteLocation(resultNode.GetPositionInOriginalTreeByte(diagnostic.Start), diagnostic.Length)
				}
			}
			memberDiagnostics = append(memberDiagnostics, diagnostic)
		}
	}

	columns := CombineColumns(CombineKindUnifySameNameAndType, tables...)
	result := NewTableSymbol("", columns...)
	info := NewSemanticInfoForSymbol(result, &result.TypeSymbol, memberDiagnostics...)
	node.SetSemanticInfo(info)
	if len(alternates) > 0 {
		statementsInfo := NewSemanticInfoForType(nil).WithAlternates(alternates)
		node.StatementList().SyntaxList.SetSemanticInfo(statementsInfo)
	}
	committed = true
}

func macroExpandScopeName(node *MacroExpandOperator, group *EntityGroupSymbol) string {
	if node != nil {
		if clause := node.ScopeReferenceName(); clause != nil && clause.EntityGroupReferenceName() != nil {
			if name := NameDeclarationSimpleName(clause.EntityGroupReferenceName()); name != "" {
				return name
			}
		}
		if reference, ok := node.EntityGroup().(*NameReference); ok {
			if name := NameReferenceSimpleName(reference); name != "" {
				return name
			}
		}
	}
	if group != nil && group.SymbolName() != "" {
		return group.SymbolName()
	}
	return "$scope"
}

func (n *simpleNodeBinder) bindMacroExpand(node *MacroExpandOperator) *SemanticInfo {
	if node == nil {
		return nil
	}
	diagnostics := make([]Diagnostic, 0)
	name := expressionReferencedSymbolName(node.EntityGroup())
	if _, ok := expressionReferencedSymbol(node.EntityGroup()).(*EntityGroupSymbol); !ok {
		if variable, variableOK := expressionReferencedSymbol(node.EntityGroup()).(*VariableSymbol); !variableOK || variable.Type == nil || variable.Type.SymbolKind() != SymbolKindEntityGroup {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownEntityGroup(name), node.EntityGroup()))
		}
	}
	result := NewTableSymbol("")
	return semanticInfoPtr(NewSemanticInfoForSymbol(result, &result.TypeSymbol, diagnostics...))
}

func macroExpandBodyResult(list *SeparatedSyntaxListOf[Statement]) Expression {
	if list != nil {
		for index := 0; index < list.Count(); index++ {
			item := list.At(index)
			if item == nil {
				continue
			}
			if statement, ok := item.Element().(*ExpressionStatement); ok {
				return statement.Expression()
			}
		}
	}
	return nil
}
