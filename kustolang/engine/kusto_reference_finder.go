package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoReferenceFinder.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type kustoReferenceLocation struct {
	element     SyntaxElement
	triviaStart int
}

type kustoReferenceFinder struct {
	code       *KustoCode
	activeRoot map[*SyntaxNode]bool
}

func newKustoReferenceFinder(code *KustoCode) *kustoReferenceFinder {
	return &kustoReferenceFinder{code: code, activeRoot: map[*SyntaxNode]bool{}}
}

func (f *kustoReferenceFinder) getClusterReferences(root *SyntaxNode, location *kustoReferenceLocation, refs *[]ClusterReference, cancellationToken CancellationToken) {
	if f == nil || root == nil || refs == nil || f.activeRoot[root] {
		return
	}
	f.activeRoot[root] = true
	defer delete(f.activeRoot, root)

	f.walk(root, 0, nil, func(element SyntaxElement, triviaStart int, parent SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		expression, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if !ok {
			return
		}
		if call, ok := expression.(*FunctionCallExpression); ok && call.ReferencedSymbol() == Functions.Cluster {
			if reference, ok := f.clusterReference(call, triviaStart, location); ok {
				*refs = append(*refs, reference)
			}
			return
		}
		node := simpleBinderSyntaxNode(expression)
		if node == nil {
			return
		}
		facts := node.GetCalledFunctionFacts()
		if facts == nil || !facts.HasClusterCall() {
			return
		}
		if body := node.GetCalledFunctionBody(); body != nil {
			callLocation := location
			if callLocation == nil {
				callLocation = bestFunctionCallLocation(expression, triviaStart)
			}
			f.getClusterReferences(body, callLocation, refs, cancellationToken)
		}
	}, func(element SyntaxElement) {
		for _, alternate := range referenceFinderAlternates(element) {
			f.getClusterReferences(alternate, location, refs, cancellationToken)
		}
	})
}

func (f *kustoReferenceFinder) getDatabaseReferences(root *SyntaxNode, location *kustoReferenceLocation, currentCluster *ClusterSymbol, currentDatabase *DatabaseSymbol, refs *[]DatabaseReference, cancellationToken CancellationToken) {
	if f == nil || root == nil || refs == nil || f.activeRoot[root] {
		return
	}
	f.activeRoot[root] = true
	defer delete(f.activeRoot, root)

	f.walk(root, 0, nil, func(element SyntaxElement, triviaStart int, parent SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		expression, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if !ok {
			return
		}
		if call, ok := expression.(*FunctionCallExpression); ok && call.ReferencedSymbol() == Functions.Database {
			if reference, ok := f.databaseReference(call, triviaStart, parent, location, currentCluster); ok {
				*refs = append(*refs, reference)
			}
			return
		}
		node := simpleBinderSyntaxNode(expression)
		if node == nil {
			return
		}
		facts := node.GetCalledFunctionFacts()
		if facts == nil || !facts.HasDatabaseCall() {
			return
		}
		if body := node.GetCalledFunctionBody(); body != nil {
			database := currentDatabase
			cluster := currentCluster
			if f.code != nil && f.code.Globals != nil {
				if owner := f.code.Globals.GetDatabase(expressionReferencedSymbol(expression)); owner != nil {
					database = owner
					if ownerCluster := f.code.Globals.GetClusterForDatabase(owner); ownerCluster != nil {
						cluster = ownerCluster
					}
				}
			}
			callLocation := location
			if callLocation == nil {
				callLocation = bestFunctionCallLocation(expression, triviaStart)
			}
			f.getDatabaseReferences(body, callLocation, cluster, database, refs, cancellationToken)
		}
	}, func(element SyntaxElement) {
		for _, alternate := range referenceFinderAlternates(element) {
			f.getDatabaseReferences(alternate, location, currentCluster, currentDatabase, refs, cancellationToken)
		}
	})
}

func (f *kustoReferenceFinder) clusterReference(call *FunctionCallExpression, callStart int, location *kustoReferenceLocation) (ClusterReference, bool) {
	argument, ok := constantStringCallArgument(call, 0)
	if !ok {
		return ClusterReference{}, false
	}
	name, _ := ExpressionConstantValue(argument).(string)
	referenceLocation := location
	if referenceLocation == nil {
		referenceLocation = syntaxLocationWithin(call, callStart, argument)
	}
	if referenceLocation == nil {
		return ClusterReference{}, false
	}
	start, length := f.publicReferenceRange(referenceLocation)
	return NewClusterReference(name, start, length), true
}

func (f *kustoReferenceFinder) databaseReference(call *FunctionCallExpression, callStart int, parent SyntaxElement, location *kustoReferenceLocation, currentCluster *ClusterSymbol) (DatabaseReference, bool) {
	argument, ok := constantStringCallArgument(call, 0)
	if !ok {
		return DatabaseReference{}, false
	}
	databaseName, _ := ExpressionConstantValue(argument).(string)
	clusterName := ""
	hasPathCluster := false
	if path, ok := WrapGeneratedSyntaxElement(parent).(*PathExpression); ok && path.Selector() == call {
		if cluster, ok := pathScopeResult(path.Expression()).(*ClusterSymbol); ok {
			clusterName = cluster.SymbolName()
			hasPathCluster = true
		}
	}
	if !hasPathCluster && currentCluster != nil {
		clusterName = currentCluster.SymbolName()
	}
	if !hasPathCluster && clusterName == "" && f != nil && f.code != nil && f.code.Globals != nil && f.code.Globals.Cluster != nil {
		clusterName = f.code.Globals.Cluster.SymbolName()
	}
	referenceLocation := location
	if referenceLocation == nil {
		referenceLocation = syntaxLocationWithin(call, callStart, argument)
	}
	if referenceLocation == nil {
		return DatabaseReference{}, false
	}
	start, length := f.publicReferenceRange(referenceLocation)
	return NewDatabaseReference(databaseName, clusterName, start, length), true
}

func (f *kustoReferenceFinder) publicReferenceRange(location *kustoReferenceLocation) (int, int) {
	if location == nil || location.element == nil {
		return 0, 0
	}
	byteStart := elementTextStart(location.element, location.triviaStart)
	byteEnd := byteStart + location.element.Width()
	if tree := syntaxElementTree(location.element); tree != nil {
		if triviaStart, found := tree.triviaStartOf(location.element); found {
			byteStart = tree.GetPositionInOriginalTree(elementTextStart(location.element, triviaStart))
			byteEnd = tree.GetPositionInOriginalTree(elementTextStart(location.element, triviaStart) + location.element.Width())
		}
	}
	text := ""
	if f != nil && f.code != nil {
		text = f.code.Text
	}
	start := byteOffsetToUTF16Position(text, byteStart)
	return start, byteOffsetToUTF16Position(text, byteEnd) - start
}

func constantStringCallArgument(call *FunctionCallExpression, index int) (Expression, bool) {
	if call == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil || index < 0 || index >= call.ArgumentList().Expressions().Count() {
		return nil, false
	}
	argument := call.ArgumentList().Expressions().At(index).Element()
	if argument == nil || !ExpressionIsConstant(argument) {
		return nil, false
	}
	_, ok := ExpressionConstantValue(argument).(string)
	return argument, ok
}

func bestFunctionCallLocation(expression Expression, triviaStart int) *kustoReferenceLocation {
	if call, ok := expression.(*FunctionCallExpression); ok && call.Name() != nil {
		return syntaxLocationWithin(call, triviaStart, call.Name())
	}
	if element, ok := expression.(SyntaxElement); ok {
		return &kustoReferenceLocation{element: element, triviaStart: triviaStart}
	}
	return nil
}

func syntaxLocationWithin(root SyntaxElement, rootStart int, target SyntaxElement) *kustoReferenceLocation {
	if start, ok := syntaxElementTriviaStartNormalized(root, target, rootStart); ok {
		return &kustoReferenceLocation{element: target, triviaStart: start}
	}
	return nil
}

func syntaxElementTriviaStartNormalized(root, target SyntaxElement, triviaStart int) (int, bool) {
	if root == nil || target == nil {
		return 0, false
	}
	rootIdentity := root
	if node := simpleBinderSyntaxNode(root); node != nil {
		rootIdentity = node
	}
	targetIdentity := target
	if node := simpleBinderSyntaxNode(target); node != nil {
		targetIdentity = node
	}
	if rootIdentity == targetIdentity {
		return triviaStart, true
	}
	childStart := triviaStart
	for _, child := range root.Children() {
		if child != nil {
			if start, ok := syntaxElementTriviaStartNormalized(child, targetIdentity, childStart); ok {
				return start, true
			}
			childStart += child.FullWidth()
		}
	}
	return 0, false
}

func (f *kustoReferenceFinder) walk(root SyntaxElement, triviaStart int, parent SyntaxElement, before func(SyntaxElement, int, SyntaxElement), after func(SyntaxElement)) {
	if root == nil {
		return
	}
	before(root, triviaStart, parent)
	childStart := triviaStart
	for _, child := range root.Children() {
		if child != nil {
			f.walk(child, childStart, root, before, after)
			childStart += child.FullWidth()
		}
	}
	if list, ok := root.(*SyntaxList); ok {
		after(list)
	} else if node := syntaxNodeForBinderLocation(root); node != nil {
		after(node)
	}
}

func referenceFinderAlternates(element SyntaxElement) []*SyntaxNode {
	switch element := element.(type) {
	case *SyntaxNode:
		return element.Alternates()
	case *SyntaxList:
		return element.Alternates()
	default:
		return nil
	}
}
