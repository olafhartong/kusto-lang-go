package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxElement.cs and SyntaxNode.cs.
// This is a compact Go representation that supports tree ownership,
// child navigation, widths and contained syntax diagnostics.

type IncludeTrivia int

const (
	IncludeTriviaAll IncludeTrivia = iota
	IncludeTriviaInterior
	IncludeTriviaMinimal
	IncludeTriviaSingleLine
)

type SyntaxElement interface {
	Kind() SyntaxKind
	Parent() SyntaxElement
	Root() SyntaxElement
	Tree() *SyntaxTree
	IndexInParent() int
	OffsetInParent() int
	ByteOffsetInParent() int
	UTF16OffsetInParent() int
	TriviaStart() int
	ByteTriviaStart() int
	UTF16TriviaStart() int
	TextStart() int
	ByteTextStart() int
	UTF16TextStart() int
	End() int
	ByteEnd() int
	UTF16End() int
	Width() int
	ByteWidth() int
	UTF16Width() int
	TriviaWidth() int
	ByteTriviaWidth() int
	UTF16TriviaWidth() int
	FullWidth() int
	ByteFullWidth() int
	UTF16FullWidth() int
	IsToken() bool
	ChildCount() int
	GetChild(index int) SyntaxElement
	GetName(index int) string
	CompletionHint(index int) CompletionHint
	IsOptional(index int) bool
	ContainsSyntaxDiagnostics() bool
	HasSyntaxDiagnostics() bool
	IsMissing() bool
	HasMissingChildren() bool
	Diagnostics() []Diagnostic
	GetContainedSyntaxDiagnostics() []Diagnostic
	ContainedDiagnostics() []Diagnostic
	Children() []SyntaxElement
	Text(includeTrivia IncludeTrivia) string
	GetChildIndex(child SyntaxElement) int
	Depth() int
	GetDescendantIndex(descendant SyntaxElement) int
	IsAncestorOf(descendant SyntaxElement) bool
	NameInParent() string
	GetNextSibling(includeZeroWidthElements ...bool) SyntaxElement
	GetPreviousSibling(includeZeroWidthElements ...bool) SyntaxElement
	GetFirstToken(includeZeroWidthTokens ...bool) *SyntaxToken
	GetLastToken(includeZeroWidthTokens ...bool) *SyntaxToken
	GetTokens(includeZeroWidthTokens ...bool) []*SyntaxToken
	WalkTokens(action func(*SyntaxToken))
	WalkElements(action func(SyntaxElement))
	GetTokenAt(position int) *SyntaxToken
	GetTokenAtByte(position int) *SyntaxToken
	GetTokenAtUTF16(position int) *SyntaxToken
	GetNodeAt(position, length int) *SyntaxNode
	GetNodeAtByte(position, length int) *SyntaxNode
	GetNodeAtUTF16(position, length int) *SyntaxNode
	GetPositionInOriginalTree(position int) int
	GetPositionInOriginalTreeByte(position int) int
	GetPositionInOriginalTreeUTF16(position int) int
	Clone(includeDiagnostics bool) SyntaxElement
}

// WithDiagnostics mirrors SyntaxElement.WithDiagnostics from
// Kusto.Language. It returns a clone with the supplied diagnostics at its
// root. As in the upstream implementation, an empty collection leaves the
// cloned element's existing diagnostics intact.
func WithDiagnostics(element SyntaxElement, diagnostics ...Diagnostic) SyntaxElement {
	clone := cloneSyntaxElementPreservingGeneratedWrapper(element, true)
	if len(diagnostics) != 0 {
		setSyntaxElementDiagnostics(clone, diagnostics)
	}
	return clone
}

// WithAdditionalDiagnostics mirrors SyntaxElement.WithAdditionalDiagnostics
// from Kusto.Language. The original element remains unchanged.
func WithAdditionalDiagnostics(element SyntaxElement, diagnostics ...Diagnostic) SyntaxElement {
	if element == nil {
		return nil
	}
	combined := append(element.Diagnostics(), diagnostics...)
	return WithDiagnostics(element, combined...)
}

// Generated Go node wrappers inherit SyntaxNode.Clone, which returns the
// underlying node. Restore the generated wrapper so extension-style helpers
// have the same dynamic node shape as their Kusto.Language counterparts.
func cloneSyntaxElementPreservingGeneratedWrapper(element SyntaxElement, includeDiagnostics bool) SyntaxElement {
	if element == nil {
		return nil
	}
	clone := element.Clone(includeDiagnostics)
	if node, ok := clone.(*SyntaxNode); ok {
		return WrapGeneratedSyntaxElement(node)
	}
	return clone
}

func setSyntaxElementDiagnostics(element SyntaxElement, diagnostics []Diagnostic) {
	if element == nil {
		return
	}
	copyDiagnostics := append([]Diagnostic{}, diagnostics...)
	if node, ok := element.(syntaxNodeFragment); ok {
		if root := node.syntaxNodeRoot(); root != nil {
			root.diagnostics = copyDiagnostics
		}
		return
	}
	switch element := element.(type) {
	case *SyntaxToken:
		if element != nil {
			element.diagnostics = copyDiagnostics
		}
	case *SyntaxList:
		if element != nil {
			element.diagnostics = copyDiagnostics
		}
	case *SeparatedElement:
		if element != nil {
			element.diagnostics = copyDiagnostics
		}
	default:
		panic("syntax element does not support diagnostics")
	}
}

// syntaxOwnership is embedded in each concrete syntax element. The owner is
// the stable, non-wrapper identity used by the generated node layer.
type syntaxOwnership struct {
	self           SyntaxElement
	parent         SyntaxElement
	tree           *SyntaxTree
	indexInParent  int
	offsetInParent int
	triviaStart    int
	hasPosition    bool
}

func (o *syntaxOwnership) syntaxOwnershipInfo() *syntaxOwnership { return o }

// ContainsSyntaxDiagnostics mirrors SyntaxElement.ContainsSyntaxDiagnostics.
// Go computes this from the present tree rather than maintaining the source
// implementation's aggregated flag, which also keeps direct test tree edits
// observable.
func (o *syntaxOwnership) ContainsSyntaxDiagnostics() bool {
	return o != nil && o.self != nil && len(o.self.ContainedDiagnostics()) != 0
}

// HasSyntaxDiagnostics mirrors SyntaxElement.HasSyntaxDiagnostics and only
// considers diagnostics owned by this element, not its descendants.
func (o *syntaxOwnership) HasSyntaxDiagnostics() bool {
	return o != nil && o.self != nil && len(o.self.Diagnostics()) != 0
}

// IsMissing mirrors SyntaxElement.IsMissing.
func (o *syntaxOwnership) IsMissing() bool {
	return o != nil && o.self != nil && o.self.Width() == 0 && o.ContainsSyntaxDiagnostics()
}

// GetContainedSyntaxDiagnostics mirrors SyntaxElement.GetContainedSyntaxDiagnostics.
// Only syntactic diagnostics live on syntax elements; semantic diagnostics are
// attached separately through SemanticInfo.
func (o *syntaxOwnership) GetContainedSyntaxDiagnostics() []Diagnostic {
	if o == nil || o.self == nil || !o.self.ContainsSyntaxDiagnostics() {
		return []Diagnostic{}
	}
	return GetContainedDiagnosticsPreserveDuplicates(o.self, DiagnosticsIncludeSyntactic)
}

func (o *syntaxOwnership) Parent() SyntaxElement {
	if o == nil {
		return nil
	}
	return o.parent
}

func (o *syntaxOwnership) Root() SyntaxElement {
	if o == nil {
		return nil
	}
	element := o.self
	for element != nil && element.Parent() != nil {
		element = element.Parent()
	}
	return element
}

func (o *syntaxOwnership) Tree() *SyntaxTree {
	if o == nil {
		return nil
	}
	if o.tree != nil {
		return o.tree
	}
	for element := o.parent; element != nil; element = element.Parent() {
		if info := syntaxOwnershipOf(element); info != nil && info.tree != nil {
			return info.tree
		}
	}
	return nil
}

func (o *syntaxOwnership) IndexInParent() int {
	if o == nil {
		return 0
	}
	return o.indexInParent
}

func (o *syntaxOwnership) OffsetInParent() int {
	if o == nil {
		return 0
	}
	return o.offsetInParent
}

func (o *syntaxOwnership) ByteOffsetInParent() int {
	return o.OffsetInParent()
}

func (o *syntaxOwnership) UTF16OffsetInParent() int {
	if o == nil || o.parent == nil {
		return 0
	}
	width := 0
	for index := 0; index < o.indexInParent; index++ {
		if sibling := o.parent.GetChild(index); sibling != nil {
			width += sibling.UTF16FullWidth()
		}
	}
	return width
}

func (o *syntaxOwnership) TriviaStart() int {
	if o == nil {
		return 0
	}
	if o.hasPosition {
		return o.triviaStart
	}
	start := 0
	for element := o.self; element != nil && element.Parent() != nil; element = element.Parent() {
		start += element.OffsetInParent()
	}
	return start
}

func (o *syntaxOwnership) ByteTriviaStart() int {
	return o.TriviaStart()
}

func (o *syntaxOwnership) UTF16TriviaStart() int {
	if o == nil {
		return 0
	}
	start := 0
	for element := o.self; element != nil && element.Parent() != nil; element = element.Parent() {
		start += element.UTF16OffsetInParent()
	}
	return start
}

func (o *syntaxOwnership) TextStart() int {
	if o == nil || o.self == nil {
		return 0
	}
	return o.TriviaStart() + o.self.TriviaWidth()
}

func (o *syntaxOwnership) ByteTextStart() int {
	return o.TextStart()
}

func (o *syntaxOwnership) UTF16TextStart() int {
	if o == nil || o.self == nil {
		return 0
	}
	return o.UTF16TriviaStart() + o.self.UTF16TriviaWidth()
}

func (o *syntaxOwnership) End() int {
	if o == nil || o.self == nil {
		return 0
	}
	return o.TriviaStart() + o.self.FullWidth()
}

func (o *syntaxOwnership) ByteEnd() int {
	return o.End()
}

func (o *syntaxOwnership) UTF16End() int {
	if o == nil || o.self == nil {
		return 0
	}
	return o.UTF16TriviaStart() + o.self.UTF16FullWidth()
}

func (o *syntaxOwnership) GetNextSibling(includeZeroWidthElements ...bool) SyntaxElement {
	return getSyntaxSibling(o, 1, optionalBool(includeZeroWidthElements))
}

func (o *syntaxOwnership) GetPreviousSibling(includeZeroWidthElements ...bool) SyntaxElement {
	return getSyntaxSibling(o, -1, optionalBool(includeZeroWidthElements))
}

func (o *syntaxOwnership) GetFirstToken(includeZeroWidthTokens ...bool) *SyntaxToken {
	return navigationFirstToken(o.self, optionalBool(includeZeroWidthTokens))
}

func (o *syntaxOwnership) GetLastToken(includeZeroWidthTokens ...bool) *SyntaxToken {
	return navigationLastToken(o.self, optionalBool(includeZeroWidthTokens))
}

func (o *syntaxOwnership) GetTokenAt(position int) *SyntaxToken {
	return syntaxTokenAt(o.self, position)
}

func (o *syntaxOwnership) GetTokenAtByte(position int) *SyntaxToken {
	return o.GetTokenAt(position)
}

func (o *syntaxOwnership) GetTokenAtUTF16(position int) *SyntaxToken {
	return syntaxTokenAtUTF16(o.self, position)
}

func (o *syntaxOwnership) GetNodeAt(position, length int) *SyntaxNode {
	return syntaxNodeAt(o.self, position, length)
}

func (o *syntaxOwnership) GetNodeAtByte(position, length int) *SyntaxNode {
	return o.GetNodeAt(position, length)
}

func (o *syntaxOwnership) GetNodeAtUTF16(position, length int) *SyntaxNode {
	return syntaxNodeAtUTF16(o.self, position, length)
}

func (o *syntaxOwnership) GetPositionInOriginalTree(position int) int {
	if tree := o.Tree(); tree != nil {
		return tree.GetPositionInOriginalTree(position)
	}
	return position
}

func (o *syntaxOwnership) GetPositionInOriginalTreeByte(position int) int {
	return o.GetPositionInOriginalTree(position)
}

func (o *syntaxOwnership) GetPositionInOriginalTreeUTF16(position int) int {
	if tree := o.Tree(); tree != nil {
		return tree.GetPositionInOriginalTreeUTF16(position)
	}
	return position
}

func (o *syntaxOwnership) Depth() int {
	if o == nil {
		return 0
	}
	return SyntaxDepth(o.self)
}

func (o *syntaxOwnership) NameInParent() string {
	if o == nil {
		return ""
	}
	return SyntaxNameInParent(o.self)
}

func (o *syntaxOwnership) GetChildIndex(child SyntaxElement) int {
	if o == nil {
		return -1
	}
	return SyntaxGetChildIndex(o.self, child)
}

func (o *syntaxOwnership) GetDescendantIndex(descendant SyntaxElement) int {
	if o == nil {
		return -1
	}
	return SyntaxGetDescendantIndex(o.self, descendant)
}

func (o *syntaxOwnership) IsAncestorOf(descendant SyntaxElement) bool {
	return o != nil && SyntaxIsAncestorOf(o.self, descendant)
}

func (o *syntaxOwnership) HasMissingChildren() bool {
	return o != nil && SyntaxHasMissingChildren(o.self)
}

func (o *syntaxOwnership) GetTokens(includeZeroWidthTokens ...bool) []*SyntaxToken {
	if o == nil {
		return []*SyntaxToken{}
	}
	return GetSyntaxTokens(o.self, includeZeroWidthTokens...)
}

func (o *syntaxOwnership) WalkTokens(action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensByte(o.self, o.ByteTriviaStart(), o.ByteEnd(), action)
	}
}

func (o *syntaxOwnership) WalkTokensRange(start, end int, action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensByte(o.self, start, end, action)
	}
}

func (o *syntaxOwnership) WalkTokensByte(action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensByte(o.self, o.ByteTriviaStart(), o.ByteEnd(), action)
	}
}

func (o *syntaxOwnership) WalkTokensRangeByte(start, end int, action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensByte(o.self, start, end, action)
	}
}

func (o *syntaxOwnership) WalkTokensUTF16(action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensUTF16(o.self, o.UTF16TriviaStart(), o.UTF16End(), action)
	}
}

func (o *syntaxOwnership) WalkTokensRangeUTF16(start, end int, action func(*SyntaxToken)) {
	if o != nil {
		WalkSyntaxTokensUTF16(o.self, start, end, action)
	}
}

func (o *syntaxOwnership) WalkElements(action func(SyntaxElement)) {
	if o != nil {
		WalkSyntaxElements(o.self, action, nil, nil)
	}
}

func syntaxOwnershipOf(element SyntaxElement) *syntaxOwnership {
	if element == nil {
		return nil
	}
	owned, _ := element.(interface{ syntaxOwnershipInfo() *syntaxOwnership })
	if owned == nil {
		return nil
	}
	return owned.syntaxOwnershipInfo()
}

func initializeSyntaxElement(element SyntaxElement) {
	info := syntaxOwnershipOf(element)
	if info == nil {
		panic("syntax element does not embed ownership metadata")
	}
	info.self = element
	offset := 0
	for index := 0; index < element.ChildCount(); index++ {
		child := element.GetChild(index)
		if child == nil {
			continue
		}
		assignDetachedSyntaxOwnership(child, element, index, offset)
		offset += child.FullWidth()
	}
}

func assignDetachedSyntaxOwnership(element, parent SyntaxElement, index, offset int) {
	info := syntaxOwnershipOf(element)
	if info == nil {
		panic("syntax child does not embed ownership metadata")
	}
	if info.parent != nil || info.tree != nil {
		panic(InvalidOperationError{Message: "The syntax element already has a parent."})
	}
	info.self = element
	info.parent = parent
	info.tree = nil
	info.indexInParent = index
	info.offsetInParent = offset
	info.hasPosition = false
}

type SyntaxNode struct {
	syntaxOwnership
	kind              SyntaxKind
	hasMatch          bool
	match             SymbolMatch // Fits the padding before generatedNodeName on 64-bit targets.
	generatedNodeName string
	children          []SyntaxElement
	diagnostics       []Diagnostic
	payloads          map[string]any
}

func NewSyntaxNode(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	return newSyntaxNode(kind, "", children, diagnostics...)
}

func newSyntaxNode(kind SyntaxKind, generatedNodeName string, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	node := &SyntaxNode{
		kind:              kind,
		generatedNodeName: generatedNodeName,
		children:          append([]SyntaxElement{}, children...),
		diagnostics:       append([]Diagnostic{}, diagnostics...),
	}
	initializeSyntaxElement(node)
	return node
}

func (n *SyntaxNode) Kind() SyntaxKind {
	if n == nil {
		return SyntaxKindNone
	}
	return n.kind
}

func (n *SyntaxNode) Width() int {
	if n == nil {
		return 0
	}
	return n.FullWidth() - n.TriviaWidth()
}

func (n *SyntaxNode) ByteWidth() int {
	return n.Width()
}

func (n *SyntaxNode) UTF16Width() int {
	if n == nil {
		return 0
	}
	return n.UTF16FullWidth() - n.UTF16TriviaWidth()
}

func (n *SyntaxNode) TriviaWidth() int {
	if len(n.children) == 0 || n.children[0] == nil {
		return 0
	}
	return n.children[0].TriviaWidth()
}

func (n *SyntaxNode) ByteTriviaWidth() int {
	return n.TriviaWidth()
}

func (n *SyntaxNode) UTF16TriviaWidth() int {
	if n == nil || len(n.children) == 0 || n.children[0] == nil {
		return 0
	}
	return n.children[0].UTF16TriviaWidth()
}

func (n *SyntaxNode) FullWidth() int {
	width := 0
	for _, child := range n.children {
		if child != nil {
			width += child.FullWidth()
		}
	}
	return width
}

func (n *SyntaxNode) ByteFullWidth() int {
	return n.FullWidth()
}

func (n *SyntaxNode) UTF16FullWidth() int {
	if n == nil {
		return 0
	}
	width := 0
	for _, child := range n.children {
		if child != nil {
			width += child.UTF16FullWidth()
		}
	}
	return width
}

func (n *SyntaxNode) IsToken() bool {
	return false
}

func (n *SyntaxNode) syntaxNodeElement() {}

// syntaxNodeRoot lets wrapper nodes expose their embedded syntax node without
// requiring every generated node type to repeat an unwrapping method.
func (n *SyntaxNode) syntaxNodeRoot() *SyntaxNode {
	return n
}

func (n *SyntaxNode) ChildCount() int {
	if n == nil {
		return 0
	}
	return len(n.children)
}

func (n *SyntaxNode) GetChild(index int) SyntaxElement {
	if n == nil {
		return nil
	}
	return n.children[index]
}

func (n *SyntaxNode) Child(index int) SyntaxElement {
	return n.GetChild(index)
}

func (n *SyntaxNode) GetName(index int) string {
	if n == nil {
		return ""
	}
	metadata, ok := n.generatedMetadata()
	if !ok || index < 0 || index >= len(metadata.Children) {
		return ""
	}
	return metadata.Children[index].Name
}

func (n *SyntaxNode) IsOptional(index int) bool {
	if n == nil {
		return false
	}
	metadata, ok := n.generatedMetadata()
	if !ok || index < 0 || index >= len(metadata.Children) {
		return false
	}
	return metadata.Children[index].Optional
}

func (n *SyntaxNode) CompletionHint(index int) CompletionHint {
	if n == nil {
		return CompletionHintInherit
	}
	if n.generatedNodeName == "NamedParameter" && index == 2 {
		if payload, ok := n.Payload("ExpressionHint"); ok {
			if hint, ok := payload.(CompletionHint); ok {
				return hint
			}
		}
		return CompletionHintNone
	}
	metadata, ok := n.generatedMetadata()
	if !ok || index < 0 || index >= len(metadata.Children) {
		return CompletionHintInherit
	}
	return metadata.Children[index].Hint
}

func (n *SyntaxNode) GeneratedNodeName() string {
	if n == nil {
		return ""
	}
	return n.generatedNodeName
}

func (n *SyntaxNode) Payload(name string) (any, bool) {
	if n == nil {
		return nil, false
	}
	if name == "Match" && n.hasMatch {
		return n.match, true
	}
	value, ok := n.payloads[name]
	return cloneSyntaxNodePayload(value), ok
}

func (n *SyntaxNode) setPayload(name string, value any) {
	if n == nil || name == "" {
		return
	}
	if name == "Match" {
		if match, ok := value.(SymbolMatch); ok {
			n.match, n.hasMatch = match, true
			delete(n.payloads, name)
			return
		}
		n.hasMatch = false
	}
	if n.payloads == nil {
		n.payloads = map[string]any{}
	}
	n.payloads[name] = cloneSyntaxNodePayload(value)
}

func (n *SyntaxNode) generatedMetadata() (GeneratedSyntaxNodeMetadata, bool) {
	if n == nil {
		return GeneratedSyntaxNodeMetadata{}, false
	}
	if n.generatedNodeName != "" {
		return generatedSyntaxNodeMetadataForName(n.generatedNodeName)
	}
	return generatedSyntaxNodeMetadataForKind(n.kind)
}

func (n *SyntaxNode) Diagnostics() []Diagnostic {
	return append([]Diagnostic{}, n.diagnostics...)
}

func (n *SyntaxNode) ContainedDiagnostics() []Diagnostic {
	var out []Diagnostic
	out = append(out, n.diagnostics...)
	for _, child := range n.children {
		if child != nil {
			out = append(out, child.ContainedDiagnostics()...)
		}
	}
	if out == nil {
		return []Diagnostic{}
	}
	return out
}

func (n *SyntaxNode) Children() []SyntaxElement {
	return append([]SyntaxElement{}, n.children...)
}

func (n *SyntaxNode) Text(includeTrivia IncludeTrivia) string {
	return syntaxChildrenText(n.children, includeTrivia)
}

func (n *SyntaxNode) Clone(includeDiagnostics bool) SyntaxElement {
	if n == nil {
		return (*SyntaxNode)(nil)
	}
	children := cloneSyntaxElements(n.children, includeDiagnostics)
	clone := newSyntaxNode(n.kind, n.generatedNodeName, children, cloneDiagnostics(n.diagnostics, includeDiagnostics)...)
	clone.match, clone.hasMatch = n.match, n.hasMatch
	clone.payloads = cloneSyntaxNodePayloads(n.payloads)
	return clone
}

func (n *SyntaxNode) GetOriginalNode() *SyntaxNode {
	return GetOriginalSyntaxNode(n)
}

func WalkSyntax(root SyntaxElement, fn func(SyntaxElement)) {
	if root == nil || fn == nil {
		return
	}
	fn(root)
	for _, child := range root.Children() {
		WalkSyntax(child, fn)
	}
}

func cloneSyntaxElements(elements []SyntaxElement, includeDiagnostics bool) []SyntaxElement {
	if elements == nil {
		return nil
	}
	out := make([]SyntaxElement, len(elements))
	for i, element := range elements {
		if element != nil {
			out[i] = element.Clone(includeDiagnostics)
		}
	}
	return out
}

func cloneDiagnostics(diagnostics []Diagnostic, includeDiagnostics bool) []Diagnostic {
	if !includeDiagnostics || len(diagnostics) == 0 {
		return nil
	}
	return append([]Diagnostic{}, diagnostics...)
}

func cloneSyntaxNodePayloads(payloads map[string]any) map[string]any {
	if len(payloads) == 0 {
		return nil
	}
	out := make(map[string]any, len(payloads))
	for name, value := range payloads {
		out[name] = cloneSyntaxNodePayload(value)
	}
	return out
}

func cloneSyntaxNodePayload(value any) any {
	switch typed := value.(type) {
	case []string:
		return append([]string{}, typed...)
	default:
		return value
	}
}
