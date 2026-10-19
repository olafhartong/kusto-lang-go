package kustolang

import "fmt"

// Ported from Binder_TreeBinder.cs VisitForkOperator. Each branch starts from
// the table entering fork; a preceding branch must not become the next scope.
func (t *treeBinder) visitFork(node *ForkOperator) {
	if node == nil || node.Expressions() == nil || node.IsBound() {
		return
	}
	oldRowScope := t.binder.rowScope
	oldInPipe := t.inPipe
	defer func() {
		t.binder.rowScope = oldRowScope
		t.inPipe = oldInPipe
	}()

	for _, branch := range node.Expressions().Elements() {
		t.binder.rowScope = oldRowScope
		t.inPipe = true
		t.visitElement(branch)
	}
	t.binder.rowScope = oldRowScope
	t.inPipe = oldInPipe
	t.bindNode(node.SyntaxNode)
}

// Ported from Binder_NodeBinder.cs VisitForkExpression.
func (n *simpleNodeBinder) bindForkExpression(node *ForkExpression) *SemanticInfo {
	if node == nil || node.Expression() == nil {
		return nil
	}
	typ := ExpressionRawResultType(node.Expression())
	if typ == nil {
		return nil
	}
	if node.NameEquals() != nil {
		if table := expressionTableResult(node.Expression()); table != nil {
			table = NewTableSymbol(NameDeclarationSimpleName(node.NameEquals().Name()),
				n.binder.declaredAndInferredColumns(table)...).WithInheritableProperties(table)
			return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
		}
	}
	return semanticInfoPtr(NewSemanticInfoForType(typ))
}

// Ported from Binder_NodeBinder.cs VisitForkOperator.
func (n *simpleNodeBinder) bindFork(node *ForkOperator) *SemanticInfo {
	if node == nil || node.Expressions() == nil {
		return nil
	}
	diagnostics := make([]Diagnostic, 0)
	if !n.inPipe {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
			DiagnosticQueryOperatorCannotBeFirst(), node.ForkKeyword()))
	}

	members := make([]Symbol, 0, node.Expressions().Count())
	for _, branch := range node.Expressions().Elements() {
		if branch == nil || branch.Expression() == nil {
			continue
		}
		expression := branch.Expression()
		typ := ExpressionRawResultType(expression)
		table := expressionTableResult(expression)
		if table == nil && typ != nil && typ != &Error.TypeSymbol && typ.SymbolTabularity() != TabularityTabular {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticTabularValueExpected(), expression))
		}
		diagnostics = append(diagnostics, n.forkQueryOperatorDiagnostics(expression)...)
		if table != nil {
			name := "Results"
			if len(members) > 0 {
				name = fmt.Sprintf("Results_%d", len(members)+1)
			}
			result := NewTableSymbol(name, n.binder.declaredAndInferredColumns(table)...).
				WithInheritableProperties(table)
			members = append(members, result)
		}
	}

	group := NewGroupSymbol(members...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(group, &group.TypeSymbol, diagnostics...))
}

var forkOperatorKinds = map[SyntaxKind]struct{}{
	SyntaxKindAsOperator: {}, SyntaxKindCountOperator: {}, SyntaxKindDistinctOperator: {},
	SyntaxKindExecuteAndCacheOperator: {}, SyntaxKindExtendOperator: {}, SyntaxKindFilterOperator: {},
	SyntaxKindInvokeOperator: {}, SyntaxKindMvExpandOperator: {}, SyntaxKindParseOperator: {},
	SyntaxKindParseWhereOperator: {}, SyntaxKindParseKvOperator: {}, SyntaxKindProjectOperator: {},
	SyntaxKindProjectAwayOperator: {}, SyntaxKindProjectByNamesOperator: {}, SyntaxKindProjectKeepOperator: {},
	SyntaxKindProjectRenameOperator: {}, SyntaxKindProjectReorderOperator: {}, SyntaxKindReduceByOperator: {},
	SyntaxKindSampleDistinctOperator: {}, SyntaxKindSampleOperator: {}, SyntaxKindSortOperator: {},
	SyntaxKindSummarizeOperator: {}, SyntaxKindTakeOperator: {}, SyntaxKindTopHittersOperator: {},
	SyntaxKindTopNestedOperator: {}, SyntaxKindTopOperator: {},
}

func (n *simpleNodeBinder) forkQueryOperatorDiagnostics(expression Expression) []Diagnostic {
	var diagnostics []Diagnostic
	root := expression
	for {
		pipe, ok := root.(*PipeExpression)
		if !ok {
			break
		}
		diagnostics = append(diagnostics, n.forkQueryOperatorDiagnostic(pipe.Operator())...)
		root = pipe.Expression()
	}
	if operator, ok := root.(QueryOperator); ok {
		diagnostics = append(diagnostics, n.forkQueryOperatorDiagnostic(operator)...)
	} else {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorExpected(), root))
	}
	return diagnostics
}

func (n *simpleNodeBinder) forkQueryOperatorDiagnostic(operator QueryOperator) []Diagnostic {
	if operator == nil {
		return nil
	}
	if _, allowed := forkOperatorKinds[operator.Kind()]; allowed || len(operator.ContainedDiagnostics()) > 0 {
		return nil
	}
	keyword := firstSyntaxToken(operator)
	name := ""
	if keyword != nil {
		name = keyword.RawText()
	}
	return []Diagnostic{n.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorNotAllowedInContext(name), keyword)}
}

func firstSyntaxToken(element SyntaxElement) *SyntaxToken {
	if token, ok := element.(*SyntaxToken); ok {
		return token
	}
	if element == nil {
		return nil
	}
	for _, child := range element.Children() {
		if token := firstSyntaxToken(child); token != nil {
			return token
		}
	}
	return nil
}
