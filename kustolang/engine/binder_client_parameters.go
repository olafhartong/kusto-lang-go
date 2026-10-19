package kustolang

// Ported from the client-parameter branch of
// src/Kusto.Language/Binder/Binder_NodeBinder.cs and the tabular-context helper
// from src/Kusto.Language/Binder/Binder_Names.cs. This is a narrow syntax-tree
// binder slice; the full tree binder remains separate future work.

func (b *Binder) BindClientParameterReferences(root SyntaxElement) int {
	if root == nil {
		return 0
	}
	count := 0
	nextTokens := nextSyntaxTokensByToken(root)
	walkSyntaxWithTriviaStartAndAncestors(root, 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if !ok {
			return
		}
		if _, ok := reference.Name().(*BracedName); !ok {
			return
		}
		info := b.ResolveClientParameterReference(NameReferenceSimpleName(reference), nameReferenceInTabularContext(element, ancestors))
		if len(info.Diagnostics) > 0 {
			info = info.WithDiagnostics(diagnosticsWithElementLocation(element, triviaStart, info.Diagnostics, nextTokens)...)
		}
		reference.SetSemanticInfo(info)
		count++
	})
	return count
}

func nameReferenceInTabularContext(element SyntaxElement, ancestors []SyntaxElement) bool {
	if element == nil {
		return false
	}
	current := element
	index := len(ancestors) - 1

	if index >= 0 {
		parent := ancestors[index]
		if _, ok := WrapGeneratedSyntaxElement(parent).(*FunctionCallExpression); ok && syntaxElementChildIndex(parent, current) == 0 {
			current = parent
			index--
		}
	}

	for index >= 0 {
		parent := ancestors[index]
		wrapped := WrapGeneratedSyntaxElement(parent)
		if _, ok := wrapped.(*ParenthesizedExpression); ok {
			current = parent
			index--
			continue
		}
		if _, ok := parent.(*SyntaxList); ok {
			current = parent
			index--
			continue
		}
		if _, ok := parent.(*SeparatedElement); ok {
			current = parent
			index--
			continue
		}
		break
	}

	if index < 0 {
		return false
	}

	parent := ancestors[index]
	wrappedParent := WrapGeneratedSyntaxElement(parent)
	if _, ok := wrappedParent.(*PipeExpression); ok && syntaxElementChildIndex(parent, current) == 0 {
		return true
	}

	if _, ok := wrappedParent.(*PathExpression); ok && syntaxElementChildIndex(parent, current) == 2 {
		if pathExpressionHasDatabaseResult(parent) {
			return true
		}
	}

	childIndex := syntaxElementChildIndex(parent, current)
	if childIndex < 0 {
		return false
	}
	return completionHintImpliesTabular(parent.CompletionHint(childIndex))
}

func syntaxElementInTabularContext(element SyntaxElement) bool {
	if element == nil {
		return false
	}
	var reversed []SyntaxElement
	for parent := element.Parent(); parent != nil; parent = parent.Parent() {
		reversed = append(reversed, parent)
	}
	ancestors := make([]SyntaxElement, len(reversed))
	for i := range reversed {
		ancestors[len(reversed)-1-i] = reversed[i]
	}
	return nameReferenceInTabularContext(element, ancestors)
}

func pathExpressionHasDatabaseResult(path SyntaxElement) bool {
	if path == nil || path.ChildCount() == 0 {
		return false
	}
	left, _ := WrapGeneratedSyntaxElement(path.GetChild(0)).(interface {
		ReferencedSymbol() Symbol
	})
	if left == nil {
		return false
	}
	switch symbol := GetResultType(left.ReferencedSymbol()).(type) {
	case *DatabaseSymbol:
		return true
	case *GroupSymbol:
		for _, member := range symbol.Members() {
			if _, ok := member.(*DatabaseSymbol); ok {
				return true
			}
		}
	}
	return false
}

func syntaxElementChildIndex(parent SyntaxElement, child SyntaxElement) int {
	if parent == nil || child == nil {
		return -1
	}
	for index, current := range parent.Children() {
		if current == child {
			return index
		}
	}
	return -1
}

func completionHintImpliesTabular(hint CompletionHint) bool {
	return hint&CompletionHintTable != 0 ||
		hint&CompletionHintTabular != 0 ||
		hint&CompletionHintNonScalar != 0 ||
		hint&CompletionHintMaterializedView != 0 ||
		hint&CompletionHintExternalTable != 0
}
