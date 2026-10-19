package kustolang

// Ported from Binder_TreeBinder.cs, Binder_NodeBinder.cs and Binder_Names.cs.

func (t *treeBinder) visitPathExpression(node *PathExpression) {
	if node == nil || node.IsBound() || node.Expression() == nil || node.Selector() == nil {
		return
	}
	tx := newUnionBindingTransaction(node)
	t.visitElement(node.Expression())
	if !expressionIsBound(node.Expression()) {
		tx.rollback()
		return
	}
	if !completePathExpression(node) && !parserRecoveredMissingPathSelector(node) {
		node.SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
		return
	}

	// Bracketed selectors are evaluated in the ordinary scope, not the scope
	// introduced by the expression to the left of the dot.
	if selector, ok := node.Selector().(*BracketedExpression); ok {
		if !t.bindBracketedSelector(node.Expression(), selector) {
			tx.rollback()
			return
		}
	} else {
		oldPathScope := t.binder.pathScope
		t.binder.pathScope = pathScopeResult(node.Expression())
		t.visitElement(node.Selector())
		t.binder.pathScope = oldPathScope
		if !expressionIsBound(node.Selector()) {
			tx.rollback()
			return
		}
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

func (t *treeBinder) visitElementExpression(node *ElementExpression) {
	if node == nil || node.IsBound() || node.ChildCount() != 2 || node.Expression() == nil || node.Selector() == nil {
		return
	}
	selector, ok := node.Selector().(*BracketedExpression)
	if !ok {
		return
	}
	tx := newUnionBindingTransaction(node)
	t.visitElement(node.Expression())
	if !expressionIsBound(node.Expression()) {
		tx.rollback()
		return
	}
	if !bindableElementSelector(selector) {
		node.SetSemanticInfo(NewSemanticInfoForType(&Error.TypeSymbol))
		return
	}
	if !t.bindBracketedSelector(node.Expression(), selector) {
		tx.rollback()
		return
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

func (t *treeBinder) visitBracketedExpression(node *BracketedExpression) {
	if node == nil || node.IsBound() || !bindableUnqualifiedBracketedExpression(node) {
		return
	}
	tx := newUnionBindingTransaction(node)
	t.visitElement(node.Expression())
	if !expressionIsBound(node.Expression()) {
		tx.rollback()
		return
	}
	t.bindNode(node.SyntaxNode)
	if !node.IsBound() {
		tx.rollback()
	}
}

// Binder_NodeBinder.VisitBracketedExpression resolves unqualified literal
// names even when QueryParser synthesized the missing close bracket.
func bindableUnqualifiedBracketedExpression(node *BracketedExpression) bool {
	return node != nil && node.ChildCount() == 3 && node.OpenBracket() != nil && node.OpenBracket().Width() != 0 &&
		node.Expression() != nil && !syntaxElementIsMissing(node.Expression()) && node.CloseBracket() != nil
}

func (t *treeBinder) bindBracketedSelector(collection Expression, selector *BracketedExpression) bool {
	if !bindableElementSelector(selector) {
		return false
	}
	oldPathScope := t.binder.pathScope
	t.binder.pathScope = nil
	t.visitElement(selector.Expression())
	t.binder.pathScope = oldPathScope
	if !expressionIsBound(selector.Expression()) {
		return false
	}
	selector.SetSemanticInfo(t.binder.bindElementSelector(collection, selector))
	return true
}

func bindableElementSelector(node *BracketedExpression) bool {
	if completeBracketedExpression(node) {
		return true
	}
	return node != nil && node.ChildCount() == 3 && node.OpenBracket() != nil && node.OpenBracket().Width() != 0 &&
		node.Expression() != nil && !syntaxElementIsMissing(node.Expression()) &&
		parserMissingTokenWithCode(node.CloseBracket(), SyntaxKindCloseBracketToken, "KS005")
}

func completePathExpression(node *PathExpression) bool {
	return node != nil && node.ChildCount() == 3 && node.Expression() != nil && node.Dot() != nil &&
		node.Dot().Width() != 0 && node.Selector() != nil && !syntaxElementIsMissing(node.Expression()) &&
		!syntaxElementIsMissing(node.Selector())
}

func parserRecoveredMissingPathSelector(node *PathExpression) bool {
	if node == nil || node.ChildCount() != 3 || node.Expression() == nil || syntaxElementIsMissing(node.Expression()) ||
		node.Dot() == nil || node.Dot().Width() == 0 {
		return false
	}
	selector, ok := node.Selector().(*NameReference)
	if !ok || !syntaxElementIsMissing(selector) {
		return false
	}
	for _, diagnostic := range selector.ContainedDiagnostics() {
		if diagnostic.Code == "KS006" {
			return true
		}
	}
	return false
}

func completeBracketedExpression(node *BracketedExpression) bool {
	return node != nil && node.ChildCount() == 3 && node.OpenBracket() != nil && node.OpenBracket().Width() != 0 &&
		node.Expression() != nil && !syntaxElementIsMissing(node.Expression()) && node.CloseBracket() != nil &&
		node.CloseBracket().Width() != 0
}

func pathElementWrapperInfo(selector Expression) SemanticInfo {
	info := boundSemanticInfoForExpression(selector)
	if info == nil {
		return NewSemanticInfoForType(&Error.TypeSymbol)
	}
	if call, ok := selector.(*FunctionCallExpression); ok {
		if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok && symbol != nil {
				return NewSemanticInfoForSymbol(symbol, returnTypeFromSymbol(symbol))
			}
		}
	}
	return NewSemanticInfoForSymbol(info.ReferencedSymbol(), resultTypeOrError(selector))
}

func pathScopeResult(expression Expression) Symbol {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if result, ok := call.SyntaxNode.Payload(functionCallResultSymbolPayload); ok {
			if symbol, ok := result.(Symbol); ok && symbol != nil {
				return symbol
			}
		}
	}
	if info := boundSemanticInfoForExpression(expression); info != nil {
		if element, ok := info.ReferencedSymbol().(*EntityGroupElementSymbol); ok {
			return element
		}
		if result := GetResultType(info.ReferencedSymbol()); result != nil {
			return result
		}
		if info.ResultType != nil {
			return info.ResultType
		}
	}
	return Error
}

func (b *Binder) bindUnqualifiedBracketedExpression(node *BracketedExpression) SemanticInfo {
	index := node.Expression()
	typ := resultTypeOrError(index)
	if sameType(typ, &Error.TypeSymbol) {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	if !sameType(typ, ScalarTypes.String.Type()) {
		return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), index))
	}
	if !ExpressionIsLiteral(index) {
		return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(
			DiagnosticExpressionMustBeLiteral(), index))
	}
	name, _ := ExpressionLiteralValue(index).(string)
	info := b.ResolveNameInRowScope(b.rowScope, name, SymbolMatchDefault, false)
	if len(info.Diagnostics) != 0 {
		diagnostics := append([]Diagnostic{}, info.Diagnostics...)
		for i, diagnostic := range diagnostics {
			if !diagnostic.HasLocation() {
				diagnostics[i] = b.diagnosticAtSyntaxElement(diagnostic, index)
			}
		}
		info.Diagnostics = diagnostics
	}
	return info
}

func (b *Binder) bindElementSelector(collection Expression, selector *BracketedExpression) SemanticInfo {
	index := selector.Expression()
	indexType := resultTypeOrError(index)
	if sameType(indexType, &Error.TypeSymbol) {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	collectionType := ExpressionRawResultType(collection)
	if collectionType == nil || sameType(collectionType, &Error.TypeSymbol) {
		return NewSemanticInfoForReferencedSymbol(Error)
	}

	if array, ok := DynamicArrayForType(collectionType); ok {
		result := ScalarTypes.GetDynamic(array.ElementType)
		if !IsIntegerOrDynamic(indexType) {
			return NewSemanticInfoForType(result, b.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.Int.SymbolName(), ScalarTypes.Long.SymbolName()), index))
		}
		return NewSemanticInfoForType(result)
	}
	if bag, ok := DynamicBagForType(collectionType); ok {
		if !IsStringOrDynamic(indexType) {
			return NewSemanticInfoForType(ScalarTypes.Dynamic.Type(), b.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.String.SymbolName()), index))
		}
		if name, ok := ExpressionConstantValue(index).(string); ok {
			if property, found := bag.TryGetProperty(name); found {
				return NewSemanticInfoForType(ScalarTypes.GetDynamic(property.Type))
			}
		}
		return NewSemanticInfoForType(ScalarTypes.Dynamic.Type())
	}
	if sameType(collectionType, ScalarTypes.Dynamic.Type()) {
		if !IsStringOrDynamic(indexType) && !IsIntegerOrDynamic(indexType) {
			return NewSemanticInfoForType(ScalarTypes.Dynamic.Type(), b.diagnosticAtSyntaxElement(
				DiagnosticExpressionMustHaveType(ScalarTypes.Int.SymbolName(), ScalarTypes.Long.SymbolName(), ScalarTypes.String.SymbolName()), index))
		}
		return NewSemanticInfoForType(ScalarTypes.Dynamic.Type())
	}
	if tuple, ok := TupleForType(collectionType); ok {
		if IsInteger(indexType) && ExpressionIsConstant(index) {
			if position, ok := constantTupleIndex(ExpressionConstantValue(index)); ok && position >= 0 && position < len(tuple.Columns()) {
				column := tuple.Columns()[position]
				return NewSemanticInfoForSymbol(column, column.Type)
			}
		}
		return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
	}
	if sameType(collectionType, ScalarTypes.Unknown.Type()) {
		return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
	}
	return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(
		DiagnosticElementAccessNotAllowedInContext(), selector))
}

func constantTupleIndex(value any) (int, bool) {
	switch value := value.(type) {
	case int:
		return value, true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}

func (b *Binder) resolveNameInPathScope(name string, match SymbolMatch, allowTableMembers bool) SemanticInfo {
	scope := b.pathScope
	if database, ok := scope.(*DatabaseSymbol); ok {
		if members := databasePathMembers(database, name, match); len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		} else if len(members) > 1 {
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
		}
		if match.Has(SymbolMatchTable) {
			table := b.getOpenTable(name, database)
			if database.IsOpen() {
				return NewSemanticInfoForSymbol(table, &table.TypeSymbol)
			}
			diagnostic := DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name)
			if b.isFuzzy {
				diagnostic = DiagnosticFuzzyTableNotDefined(name)
			}
			return NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostic)
		}
	}
	if group, ok := scope.(*GroupSymbol); ok && pathScopeGroupPassesThrough(group) {
		members := distinctSymbols(pathScopeMembers(group, name, match))
		if len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		}
		if len(members) > 1 {
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
		}
		return unknownNameSemanticInfo(name, match, matchRequestsTabularName(match))
	}
	if table, ok := tableSymbolFromSymbol(scope); ok && table != nil && !allowTableMembers {
		return unknownNameSemanticInfo(name, match, false)
	}
	if group, ok := scope.(*GroupSymbol); ok && patternPathGroup(group) {
		members := FindMembers(group, name, match, false)
		if len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		}
		if len(members) > 1 {
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
		}
		return unknownNameSemanticInfo(name, match, false)
	}
	typ := returnTypeFromSymbol(scope)
	bag, bagOK := scope.(*DynamicBagSymbol)
	if !bagOK {
		bag, bagOK = DynamicBagForType(typ)
	}
	if bagOK {
		if property, found := bag.TryGetProperty(name); found && SymbolMatches(property, match) {
			return NewSemanticInfoForType(ScalarTypes.GetDynamic(property.Type))
		}
		return NewSemanticInfoForType(ScalarTypes.Dynamic.Type())
	}
	if isDynamicType(typ) {
		return NewSemanticInfoForType(ScalarTypes.Dynamic.Type())
	}
	if sameType(typ, ScalarTypes.Unknown.Type()) {
		return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
	}
	if _, ok := scope.(*ErrorSymbol); ok || sameType(typ, &Error.TypeSymbol) {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	tuple, tupleOK := scope.(*TupleSymbol)
	if !tupleOK {
		tuple, tupleOK = TupleForType(typ)
	}
	if tupleOK {
		members := FindMembers(tuple, name, match, false)
		if len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		}
		if len(members) > 1 {
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
		}
	}
	if members := FindMembers(scope, name, match, false); len(members) == 1 {
		return NewSemanticInfoForReferencedSymbol(members[0])
	} else if len(members) > 1 {
		return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
	}
	return unknownNameSemanticInfo(name, match, false)
}

func patternPathGroup(group *GroupSymbol) bool {
	if group == nil || len(group.Members()) == 0 {
		return false
	}
	for _, member := range group.Members() {
		if _, ok := member.(*VariableSymbol); !ok {
			return false
		}
	}
	return true
}

func databasePathMembers(database *DatabaseSymbol, name string, match SymbolMatch) []Symbol {
	if database == nil {
		return nil
	}
	functions := FindMembers(database, name, match&SymbolMatchFunction, false)
	zeroArgumentFunctions := functions[:0]
	for _, member := range functions {
		function, ok := member.(*FunctionSymbol)
		if !ok || function.MinArgumentCount() == 0 {
			zeroArgumentFunctions = append(zeroArgumentFunctions, member)
		}
	}
	if len(zeroArgumentFunctions) > 0 {
		return zeroArgumentFunctions
	}
	return FindMembers(database, name, match&^SymbolMatchFunction, false)
}

func pathScopeGroupPassesThrough(group *GroupSymbol) bool {
	if group == nil {
		return false
	}
	for _, member := range group.Members() {
		switch member.(type) {
		case *DatabaseSymbol, *ClusterSymbol:
			return true
		}
	}
	return false
}

func pathScopeMembers(scope Symbol, name string, match SymbolMatch) []Symbol {
	switch scope := scope.(type) {
	case *DatabaseSymbol:
		return databasePathMembers(scope, name, match)
	case *GroupSymbol:
		var members []Symbol
		for _, member := range scope.Members() {
			members = append(members, pathScopeMembers(member, name, match)...)
		}
		return members
	default:
		return FindMembers(scope, name, match, false)
	}
}
