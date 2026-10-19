package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxTree.cs and SyntaxNode.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type SyntaxTree struct {
	Root             *SyntaxNode
	Original         *SyntaxTree
	OffsetInOriginal int
}

type syntaxNodeFragment interface {
	SyntaxElement
	syntaxNodeRoot() *SyntaxNode
}

func NewSyntaxTree(root *SyntaxNode, original *SyntaxTree, offsetInOriginal int) *SyntaxTree {
	if root == nil {
		root = NewSyntaxNode(SyntaxKindNone, nil)
	}
	tree := &SyntaxTree{Root: root, Original: original, OffsetInOriginal: offsetInOriginal}
	tree.assignOwnership(root)
	return tree
}

// CopyAsFragment mirrors SyntaxNode.CopyAsFragment from the upstream
// Kusto.Language syntax API. It keeps the generated wrapper type when one is
// present, while the detached tree retains the source tree and offset used to
// project diagnostics and positions back to the original query.
func CopyAsFragment[T syntaxNodeFragment](node T) T {
	clone := CopySyntaxElementAsFragment(node)
	if typed, ok := clone.(T); ok {
		return typed
	}
	panic("CopyAsFragment could not preserve the syntax node type")
}

// CopySyntaxElementAsFragment is the non-generic package boundary used by the
// public facade. The returned element preserves generated wrapper identity.
func CopySyntaxElementAsFragment(node SyntaxElement) SyntaxElement {
	if node == nil || node.IsToken() {
		panic("CopyAsFragment requires a syntax node")
	}
	clone := cloneSyntaxElementPreservingGeneratedWrapper(node, true)
	fragment, ok := clone.(syntaxNodeFragment)
	if !ok {
		panic("CopyAsFragment requires a syntax node")
	}
	root := fragment.syntaxNodeRoot()
	if root == nil {
		panic("CopyAsFragment requires a syntax node")
	}
	NewSyntaxTree(root, node.Tree(), node.ByteTriviaStart())
	return clone
}

func (t *SyntaxTree) assignOwnership(root SyntaxElement) {
	rootInfo := syntaxOwnershipOf(root)
	if rootInfo == nil {
		panic("syntax tree root does not embed ownership metadata")
	}
	// A tree may be created from a subtree or a syntax graph whose children
	// were replaced directly. Treat the supplied node as a detached fragment
	// root and rebuild ownership from the actual child structure.
	var assign func(SyntaxElement, SyntaxElement, int, int, int)
	assign = func(element, parent SyntaxElement, index, offset, triviaStart int) {
		if element == nil {
			return
		}
		info := syntaxOwnershipOf(element)
		info.self = element
		info.parent = parent
		info.indexInParent = index
		info.offsetInParent = offset
		info.tree = t
		info.triviaStart = triviaStart
		info.hasPosition = true
		childOffset := 0
		for index := 0; index < element.ChildCount(); index++ {
			child := element.GetChild(index)
			if child != nil {
				assign(child, element, index, childOffset, triviaStart+childOffset)
				childOffset += child.FullWidth()
			}
		}
	}
	assign(root, nil, 0, 0, 0)
}

func (t *SyntaxTree) GetPositionInOriginalTree(position int) int {
	for current := t; current != nil && current.Original != nil; current = current.Original {
		position += current.OffsetInOriginal
	}
	return position
}

func (t *SyntaxTree) GetPositionInOriginalTreeByte(position int) int {
	return t.GetPositionInOriginalTree(position)
}

func (t *SyntaxTree) GetPositionInOriginalTreeUTF16(position int) int {
	for current := t; current != nil && current.Original != nil; current = current.Original {
		originalText := ""
		if current.Original.Root != nil {
			originalText = current.Original.Root.Text(IncludeTriviaAll)
		}
		position += utf16TextWidthOfBytePrefix(originalText, current.OffsetInOriginal)
	}
	return position
}

func (t *SyntaxTree) triviaStartOf(target SyntaxElement) (int, bool) {
	if t == nil || t.Root == nil || target == nil {
		return 0, false
	}
	info := syntaxOwnershipOf(target)
	if info == nil || info.Tree() != t {
		return 0, false
	}
	return info.TriviaStart(), true
}

func (t *SyntaxTree) Depth() int {
	if t == nil || t.Root == nil {
		return 0
	}
	maxDepth, depth := 0, 0
	WalkSyntaxNodes(t.Root, func(*SyntaxNode) {
		depth++
		if depth > maxDepth {
			maxDepth = depth
		}
	}, func(*SyntaxNode) {
		depth--
	}, nil)
	return maxDepth
}

func (t *SyntaxTree) IsSafeToRecurse(state *GlobalState) bool {
	if t == nil {
		return true
	}
	if state == nil {
		state = DefaultGlobalState
	}
	return t.Depth() <= GlobalStateGetProperty(state, Properties.MaxAnalysisDepth)
}

func (t *SyntaxTree) Diagnostics() []Diagnostic {
	if t == nil || t.Root == nil {
		return []Diagnostic{}
	}
	return GetContainedDiagnostics(t.Root, DiagnosticsIncludeSyntactic)
}
