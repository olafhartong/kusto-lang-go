package kustolang

// searchPredicateBinder implements search's abbreviated term grammar without
// changing normal expression binding outside parentheses and boolean and/or.
type searchPredicateBinder struct {
	binder *Binder
	tree   *treeBinder
}

func newSearchPredicateBinder(b *Binder, t *treeBinder) *searchPredicateBinder {
	return &searchPredicateBinder{binder: b, tree: t}
}

func (p *searchPredicateBinder) bind(expression Expression) {
	if expression == nil || simpleBinderSyntaxNode(expression).IsBound() {
		return
	}
	switch node := expression.(type) {
	case *StarExpression:
		node.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Bool.Type()))
	case *LiteralExpression:
		p.bindStandaloneTerm(node)
	case *NameReference:
		p.bindStandaloneTerm(node)
	case *ParenthesizedExpression:
		p.bind(node.Expression())
		node.SetSemanticInfo(NewSemanticInfoForType(resultTypeOrError(node.Expression())))
	case *BinaryExpression:
		kind := operatorKindForSyntax(node.Kind())
		if kind != OperatorKindAnd && kind != OperatorKindOr {
			p.tree.visitScalarOperator(node.SyntaxNode)
			return
		}
		p.bindBinaryOperand(node.Left(), true)
		p.bindBinaryOperand(node.Right(), true)
		p.tree.bindNode(node.SyntaxNode)
	default:
		p.tree.visitElement(expression)
	}
}

func (p *searchPredicateBinder) bindBinaryOperand(expression Expression, binary bool) {
	if star, ok := expression.(*StarExpression); ok && binary {
		star.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Dynamic.Type()))
		return
	}
	p.bind(expression)
}

func (p *searchPredicateBinder) bindStandaloneTerm(expression Expression) {
	p.tree.visitElement(expression)
	if info := boundSemanticInfoForExpression(expression); info != nil && info.IsConstant && sameType(info.ResultType, ScalarTypes.String.Type()) {
		adjusted := *info
		adjusted.ResultType = ScalarTypes.Bool.Type()
		simpleBinderSyntaxNode(expression).SetSemanticInfo(adjusted)
	}
}
