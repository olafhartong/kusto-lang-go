package kustolang

// Ported from Binder_TreeBinder.cs and Binder_NodeBinder.cs make-series visitors
// at the revision pinned in third_party/kusto/UPSTREAM.json.

func (t *treeBinder) visitMakeSeries(node *MakeSeriesOperator) {
	if node == nil || node.IsBound() {
		return
	}
	for _, parameter := range node.Parameters().Elements() {
		t.visitElement(parameter)
	}
	oldKind := t.binder.scopeKind
	t.binder.scopeKind = ScopeKindAggregate
	t.visitElement(node.Aggregates())
	t.binder.scopeKind = oldKind
	if clause := node.OnClause(); clause != nil {
		t.visitElement(clause)
	}
	if clause := node.RangeClause(); clause != nil {
		t.visitElement(clause)
	}
	if clause := node.ByClause(); clause != nil {
		t.visitElement(clause)
	}
	t.bindNode(node.SyntaxNode)
}

func (n *simpleNodeBinder) bindMakeSeriesExpression(node *MakeSeriesExpression) *SemanticInfo {
	if node == nil {
		return nil
	}
	return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(node.Expression())))
}

func (n *simpleNodeBinder) bindMakeSeries(node *MakeSeriesOperator) *SemanticInfo {
	if node == nil || node.Parameters() == nil || node.Aggregates() == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if !n.hasPipedInput(node) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.MakeSeriesKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.MakeSeriesParameters)...)
	builder := NewProjectionBuilder()
	suffix := 1
	if by := node.ByClause(); by != nil && by.Expressions() != nil {
		for _, separated := range by.Expressions().Elements() {
			expression := separated.Element()
			makeSeriesCheckScalar(n.binder, expression, &diagnostics)
			n.projection.addSummarizeExpression(expression, builder, &diagnostics, false, &suffix)
		}
	}
	for _, separated := range node.Aggregates().Elements() {
		aggregate := separated.Element()
		if aggregate == nil {
			continue
		}
		makeSeriesCheckScalar(n.binder, aggregate.Expression(), &diagnostics)
		makeSeriesAddColumn(aggregate.Expression(), builder, &diagnostics, true, &suffix, ScalarTypes.Dynamic.Type())
	}
	if on := node.OnClause(); on != nil {
		makeSeriesCheckScalar(n.binder, on.Expression(), &diagnostics)
		makeSeriesAddColumn(on.Expression(), builder, &diagnostics, false, &suffix, ScalarTypes.Dynamic.Type())
		makeSeriesCheckRange(n.binder, node.RangeClause(), resultTypeOrError(on.Expression()), &diagnostics)
	}
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(n.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func makeSeriesAddColumn(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic, aggregate bool, suffix *int, typ *TypeSymbol) {
	if syntaxElementIsMissing(expression) {
		return
	}
	name := ""
	location := SyntaxElement(expression)
	if named, ok := expression.(*SimpleNamedExpression); ok {
		name = NameDeclarationSimpleName(named.Name())
		expression = named.Expression()
		location = named.Name()
	} else {
		name = summarizeExpressionName(expression, aggregate)
	}
	if name == "" {
		name = serializeDefaultColumnName(builder, suffix)
	}
	column := NewColumnSymbolWithSource(name, typ, expression)
	builder.Declare(column, diagnostics, location)
	if named, ok := location.(*NameDeclaration); ok {
		named.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
}

func makeSeriesCheckScalar(binder *Binder, expression Expression, diagnostics *[]Diagnostic) {
	typ := ExpressionResultType(expression)
	if typ != nil && !isErrorTypeSymbol(typ) && !sameType(typ, ScalarTypes.Unknown.Type()) && !IsScalar(typ) {
		*diagnostics = append(*diagnostics, binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
	}
}

func makeSeriesCheckAssignable(binder *Binder, expression Expression, target *TypeSymbol, diagnostics *[]Diagnostic) {
	typ := ExpressionResultType(expression)
	if typ != nil && target != nil && !isErrorTypeSymbol(typ) && !isErrorTypeSymbol(target) &&
		!sameType(typ, ScalarTypes.Unknown.Type()) && !sameType(target, ScalarTypes.Unknown.Type()) &&
		!SymbolAssignableTo(typ, target, ConversionPromotable) {
		*diagnostics = append(*diagnostics, binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(target.SymbolName()), expression))
	}
}

func makeSeriesCheckRange(binder *Binder, clause MakeSeriesRangeClause, rangeType *TypeSymbol, diagnostics *[]Diagnostic) {
	checkInterval := func(expression Expression) {
		typ := ExpressionResultType(expression)
		valid := makeSeriesIntervalTypeValid(rangeType, typ)
		if !valid && typ != nil && rangeType != nil && !isErrorTypeSymbol(typ) && !isErrorTypeSymbol(rangeType) && !sameType(typ, ScalarTypes.Unknown.Type()) && !sameType(rangeType, ScalarTypes.Unknown.Type()) {
			*diagnostics = append(*diagnostics, binder.diagnosticAtSyntaxElement(DiagnosticTypeIsNotIntervalType(typ.SymbolName(), rangeType.SymbolName()), expression))
		}
	}
	switch clause := clause.(type) {
	case *MakeSeriesInRangeClause:
		if clause.Arguments() == nil || clause.Arguments().Expressions() == nil {
			return
		}
		expressions := clause.Arguments().Expressions()
		if len(clause.ContainedDiagnostics()) == 0 && expressions.Count() != 3 {
			*diagnostics = append(*diagnostics, binder.diagnosticAtSyntaxElement(DiagnosticArgumentCountExpected(3), expressions))
		}
		if expressions.Count() == 3 {
			makeSeriesCheckAssignable(binder, expressions.At(0).Element(), rangeType, diagnostics)
			makeSeriesCheckAssignable(binder, expressions.At(1).Element(), rangeType, diagnostics)
			checkInterval(expressions.At(2).Element())
		}
	case *MakeSeriesFromToStepClause:
		if from := clause.MakeSeriesFromClause(); from != nil && from.Expression() != nil {
			makeSeriesCheckAssignable(binder, from.Expression(), rangeType, diagnostics)
		}
		if to := clause.MakeSeriesToClause(); to != nil && to.Expression() != nil {
			makeSeriesCheckAssignable(binder, to.Expression(), rangeType, diagnostics)
		}
		if step := clause.MakeSeriesStepClause(); step != nil && step.Expression() != nil {
			checkInterval(step.Expression())
		}
	}
}

func makeSeriesIntervalTypeValid(rangeType, intervalType *TypeSymbol) bool {
	if rangeType == nil || intervalType == nil {
		return false
	}
	types := []*TypeSymbol{rangeType, intervalType}
	matches := GetBestMatchingSignatures(
		Operators.Add.Signatures(),
		[]Expression{nil, nil},
		[]Symbol{rangeType, intervalType},
		false,
	)
	result := GetCommonSignatureReturnType(matches, types, func(signature *Signature) *TypeSymbol {
		return signature.GetReturnTypeForArgumentTypes(nil, types)
	})
	return result != nil && SymbolAssignableTo(result, rangeType, ConversionNone)
}
