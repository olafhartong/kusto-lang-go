package kustolang

// Ported from the navigation and traversal surface in
// src/Kusto.Language/Syntax/SyntaxElement.cs and SyntaxNode.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func SyntaxIsMissing(element SyntaxElement) bool {
	return element != nil && element.IsMissing()
}

func SyntaxHasMissingChildren(element SyntaxElement) bool {
	if element == nil || len(element.ContainedDiagnostics()) == 0 {
		return false
	}
	for index := element.ChildCount() - 1; index >= 0; index-- {
		child := element.GetChild(index)
		if child != nil && (SyntaxIsMissing(child) || SyntaxHasMissingChildren(child)) {
			return true
		}
	}
	return false
}

func SyntaxGetChildIndex(element, child SyntaxElement) int {
	if element != nil && child != nil && child.Parent() == element {
		return child.IndexInParent()
	}
	return -1
}

func SyntaxDepth(element SyntaxElement) int {
	depth := 0
	for element != nil && element.Parent() != nil {
		depth++
		element = element.Parent()
	}
	return depth
}

func SyntaxGetCommonAncestor(a, b SyntaxElement) *SyntaxNode {
	if a == nil || b == nil {
		return nil
	}
	da, db := SyntaxDepth(a), SyntaxDepth(b)
	for da > db && da > 0 {
		a = a.Parent()
		da--
	}
	for db > da && db > 0 {
		b = b.Parent()
		db--
	}
	if da > 0 && a.Parent() == b.Parent() {
		return syntaxNodeForBinderLocation(a.Parent())
	}
	return nil
}

func SyntaxGetDescendantIndex(element, descendant SyntaxElement) int {
	if element == nil || descendant == nil {
		return -1
	}
	if descendant.Parent() != element {
		depth, descendantDepth := SyntaxDepth(element), SyntaxDepth(descendant)
		for descendant != nil && descendantDepth > depth+1 {
			descendant = descendant.Parent()
			descendantDepth--
		}
	}
	if descendant != nil && descendant.Parent() == element {
		return descendant.IndexInParent()
	}
	return -1
}

func SyntaxIsAncestorOf(element, descendant SyntaxElement) bool {
	if element == nil {
		return false
	}
	for descendant != nil {
		if descendant.Parent() == element {
			return true
		}
		descendant = descendant.Parent()
	}
	return false
}

func SyntaxNameInParent(element SyntaxElement) string {
	if element == nil || element.Parent() == nil {
		return ""
	}
	parent := element.Parent()
	return parent.GetName(SyntaxGetChildIndex(parent, element))
}

func syntaxElementAs[T SyntaxElement](element SyntaxElement) (T, bool) {
	var zero T
	if element == nil {
		return zero, false
	}
	if typed, ok := element.(T); ok {
		return typed, true
	}
	if wrapped := WrapGeneratedSyntaxElement(element); wrapped != nil {
		if typed, ok := wrapped.(T); ok {
			return typed, true
		}
	}
	return zero, false
}

func GetFirstAncestor[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	if element == nil {
		var zero T
		return zero
	}
	return firstSyntaxAncestor(element.Parent(), predicate)
}

func GetFirstAncestorOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return firstSyntaxAncestor(element, predicate)
}

func firstSyntaxAncestor[T SyntaxElement](element SyntaxElement, predicate []func(T) bool) T {
	for current := element; current != nil; current = current.Parent() {
		if typed, ok := syntaxElementAs[T](current); ok && (len(predicate) == 0 || predicate[0] == nil || predicate[0](typed)) {
			return typed
		}
	}
	var zero T
	return zero
}

func GetAncestors[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	if element == nil {
		return []T{}
	}
	return syntaxAncestors(element.Parent(), predicate)
}

func GetAncestorsOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return syntaxAncestors(element, predicate)
}

func syntaxAncestors[T SyntaxElement](element SyntaxElement, predicate []func(T) bool) []T {
	out := []T{}
	for current := element; current != nil; current = current.Parent() {
		if typed, ok := syntaxElementAs[T](current); ok && (len(predicate) == 0 || predicate[0] == nil || predicate[0](typed)) {
			out = append(out, typed)
		}
	}
	return out
}

func GetFirstDescendant[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return firstSyntaxDescendant(element, false, predicate)
}

func GetFirstDescendantOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) T {
	return firstSyntaxDescendant(element, true, predicate)
}

func firstSyntaxDescendant[T SyntaxElement](element SyntaxElement, includeSelf bool, predicate []func(T) bool) T {
	var result T
	found := false
	WalkSyntaxElements(element, func(current SyntaxElement) {
		if found || (!includeSelf && current == element) {
			return
		}
		if typed, ok := syntaxElementAs[T](current); ok && (len(predicate) == 0 || predicate[0] == nil || predicate[0](typed)) {
			result = typed
			found = true
		}
	}, nil, func(SyntaxElement) bool { return !found })
	return result
}

func GetDescendants[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return syntaxDescendants(element, false, predicate)
}

func GetDescendantsOrSelf[T SyntaxElement](element SyntaxElement, predicate ...func(T) bool) []T {
	return syntaxDescendants(element, true, predicate)
}

func syntaxDescendants[T SyntaxElement](element SyntaxElement, includeSelf bool, predicate []func(T) bool) []T {
	out := []T{}
	WalkSyntaxElements(element, func(current SyntaxElement) {
		if !includeSelf && current == element {
			return
		}
		if typed, ok := syntaxElementAs[T](current); ok && (len(predicate) == 0 || predicate[0] == nil || predicate[0](typed)) {
			out = append(out, typed)
		}
	}, nil, nil)
	return out
}

func GetSyntaxTokens(element SyntaxElement, includeZeroWidthTokens ...bool) []*SyntaxToken {
	includeZeroWidth := optionalBool(includeZeroWidthTokens)
	tokens := []*SyntaxToken{}
	WalkSyntaxElements(element, func(current SyntaxElement) {
		if token, ok := current.(*SyntaxToken); ok && (includeZeroWidth || token.FullWidth() > 0) {
			tokens = append(tokens, token)
		}
	}, nil, nil)
	return tokens
}

func WalkSyntaxTokens(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	WalkSyntaxTokensByte(element, start, end, action)
}

func WalkSyntaxTokensByte(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	if element == nil || action == nil {
		return
	}
	if start < element.ByteTriviaStart() {
		start = element.ByteTriviaStart()
	}
	if end > element.ByteEnd() {
		end = element.ByteEnd()
	}
	if start >= end {
		return
	}
	for _, token := range GetSyntaxTokens(element) {
		if token.ByteEnd() > start && token.ByteTriviaStart() < end {
			action(token)
		}
	}
}

func WalkSyntaxTokensUTF16(element SyntaxElement, start, end int, action func(*SyntaxToken)) {
	if element == nil || action == nil {
		return
	}
	if start < element.UTF16TriviaStart() {
		start = element.UTF16TriviaStart()
	}
	if end > element.UTF16End() {
		end = element.UTF16End()
	}
	if start >= end {
		return
	}
	for _, token := range GetSyntaxTokens(element) {
		if token.UTF16End() > start && token.UTF16TriviaStart() < end {
			action(token)
		}
	}
}

type syntaxWalkFrame struct {
	element SyntaxElement
	next    int
	descend bool
}

func WalkSyntaxElements(root SyntaxElement, before, after func(SyntaxElement), descend func(SyntaxElement) bool) {
	if root == nil {
		return
	}
	if before != nil {
		before(root)
	}
	rootDescend := descend == nil || descend(root)
	stack := []syntaxWalkFrame{{element: root, descend: rootDescend}}
	for len(stack) > 0 {
		frame := &stack[len(stack)-1]
		if frame.descend && frame.next < frame.element.ChildCount() {
			child := frame.element.GetChild(frame.next)
			frame.next++
			if child == nil {
				continue
			}
			if before != nil {
				before(child)
			}
			childDescend := descend == nil || descend(child)
			stack = append(stack, syntaxWalkFrame{element: child, descend: childDescend})
			continue
		}
		if after != nil {
			after(frame.element)
		}
		stack = stack[:len(stack)-1]
	}
}

func WalkSyntaxNodes(root *SyntaxNode, before, after func(*SyntaxNode), descend func(*SyntaxNode) bool) {
	WalkSyntaxElements(root, func(element SyntaxElement) {
		if node := syntaxNodeForBinderLocation(element); node != nil && before != nil {
			before(node)
		}
	}, func(element SyntaxElement) {
		if node := syntaxNodeForBinderLocation(element); node != nil && after != nil {
			after(node)
		}
	}, func(element SyntaxElement) bool {
		node := syntaxNodeForBinderLocation(element)
		return node == nil || descend == nil || descend(node)
	})
}

func GetOriginalSyntaxNode(node *SyntaxNode) *SyntaxNode {
	for node != nil && node.Tree() != nil && node.Tree().Original != nil {
		tree := node.Tree()
		start := tree.OffsetInOriginal + node.TextStart()
		original := tree.Original.Root.GetNodeAt(start, node.Width())
		if original == nil || original == node {
			break
		}
		node = original
	}
	return node
}
