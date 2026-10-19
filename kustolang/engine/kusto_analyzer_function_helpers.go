package kustolang

func functionCallArgumentAt(call *FunctionCallExpression, index int) Expression {
	if call == nil || index < 0 || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return nil
	}
	expressions := call.ArgumentList().Expressions()
	if index >= expressions.Count() {
		return nil
	}
	argument := expressions.At(index)
	if argument == nil {
		return nil
	}
	return argument.Element()
}

func functionCallArgumentCount(call *FunctionCallExpression) int {
	if call == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return 0
	}
	return call.ArgumentList().Expressions().Count()
}

func walkSyntaxWithTriviaStartAndAncestors(root SyntaxElement, triviaStart int, fn func(SyntaxElement, int, []SyntaxElement)) {
	walkSyntaxWithTriviaStartAndAncestorsCore(root, triviaStart, nil, fn)
}

func walkSyntaxWithTriviaStartAndAncestorsCore(root SyntaxElement, triviaStart int, ancestors []SyntaxElement, fn func(SyntaxElement, int, []SyntaxElement)) {
	if root == nil || fn == nil {
		return
	}
	fn(root, triviaStart, ancestors)
	childStart := triviaStart
	childAncestors := append(append([]SyntaxElement{}, ancestors...), root)
	for _, child := range root.Children() {
		if child != nil {
			walkSyntaxWithTriviaStartAndAncestorsCore(child, childStart, childAncestors, fn)
			childStart += child.FullWidth()
		}
	}
}
