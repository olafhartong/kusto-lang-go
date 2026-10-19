package kustolang

type implicitFunctionArgument struct {
	call       *SyntaxNode
	expression Expression
	typ        *TypeSymbol
}

func invokeFunctionCall(expression Expression) (*FunctionCallExpression, bool) {
	switch expression := expression.(type) {
	case *FunctionCallExpression:
		return expression, true
	case *PathExpression:
		return invokeFunctionCall(expression.Selector())
	default:
		return nil, false
	}
}

func (t *treeBinder) visitInvoke(node *InvokeOperator) {
	if node == nil || node.Function() == nil {
		return
	}
	call, _ := invokeFunctionCall(node.Function())

	inputType := &EmptyTableSymbol.TypeSymbol
	if t.binder.rowScope != nil {
		inputType = &t.binder.rowScope.TypeSymbol
	}
	oldImplicit, oldRow := t.binder.implicitFunctionArgument, t.binder.rowScope
	var callNode *SyntaxNode
	if call != nil {
		callNode = call.SyntaxNode
	}
	implicitExpression := FakeExpressionCreate(inputType)
	if call != nil && call.Name() != nil {
		implicitExpression = call.Name()
	}
	if pipe, ok := WrapGeneratedSyntaxElement(node.Parent()).(*PipeExpression); ok &&
		simpleBinderSyntaxNode(pipe.Operator()) == node.SyntaxNode && pipe.Expression() != nil {
		implicitExpression = pipe.Expression()
	}
	t.binder.implicitFunctionArgument = &implicitFunctionArgument{
		call: callNode, expression: implicitExpression, typ: inputType,
	}
	t.binder.rowScope = nil
	defer func() { t.binder.implicitFunctionArgument, t.binder.rowScope = oldImplicit, oldRow }()
	t.visitElement(node.Function())
	t.bindNode(node.SyntaxNode)
}

func (n *simpleNodeBinder) bindInvoke(node *InvokeOperator) *SemanticInfo {
	if node == nil || node.Function() == nil {
		return nil
	}
	diagnostics := n.notFirstDiagnostics(node)
	call, ok := invokeFunctionCall(node.Function())
	if !ok {
		functionNode := simpleBinderSyntaxNode(node.Function())
		if functionNode == nil || !functionNode.IsBound() {
			return nil
		}
		return semanticInfoPtr(NewSemanticInfoForType(resultTypeOrError(node.Function()), diagnostics...))
	}
	if !call.IsBound() {
		return nil
	}
	info, _ := call.GetSemanticInfo()
	resultSymbol, _ := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
	if symbol, ok := resultSymbol.(Symbol); ok && symbol != nil {
		return semanticInfoPtr(NewSemanticInfoForSymbol(symbol, returnTypeFromSymbol(symbol), diagnostics...))
	}
	return semanticInfoPtr(NewSemanticInfoForType(info.ResultType, diagnostics...))
}
