package kustolang

// Ported from src/Kusto.Language/Utils/SubstringMap.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type SubstringMapEntry[T any] struct {
	Key   string
	Value T
}

type SubstringMap[T any] struct {
	root *substringMapNode[T]
}

type substringMapNode[T any] struct {
	value    SubstringMapEntry[T]
	hasValue bool
	children map[byte]*substringMapNode[T]
}

func NewSubstringMap[T any](entries []SubstringMapEntry[T]) *SubstringMap[T] {
	root := &substringMapNode[T]{}
	for _, entry := range entries {
		root.add(entry)
	}
	return &SubstringMap[T]{root: root}
}

func (m *SubstringMap[T]) GetLongestMatch(text string, start int) SubstringMapEntry[T] {
	if m == nil || m.root == nil || start < 0 || start >= len(text) {
		return SubstringMapEntry[T]{}
	}

	var bestNode *substringMapNode[T]
	node := m.root
	for pos := start; pos < len(text); pos++ {
		node = node.get(text[pos])
		if node == nil {
			break
		}
		if node.hasValue {
			bestNode = node
		}
	}

	if bestNode != nil {
		return bestNode.value
	}
	return SubstringMapEntry[T]{}
}

func (n *substringMapNode[T]) add(entry SubstringMapEntry[T]) {
	node := n
	for i := 0; i < len(entry.Key); i++ {
		ch := entry.Key[i]
		if node.children == nil {
			node.children = map[byte]*substringMapNode[T]{}
		}
		child := node.children[ch]
		if child == nil {
			child = &substringMapNode[T]{}
			node.children[ch] = child
		}
		node = child
	}
	node.value = entry
	node.hasValue = true
}

func (n *substringMapNode[T]) get(key byte) *substringMapNode[T] {
	if n == nil || n.children == nil {
		return nil
	}
	return n.children[key]
}
