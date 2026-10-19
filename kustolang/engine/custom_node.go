package kustolang

import "fmt"

// Ported from src/Kusto.Language/Syntax/CustomNode.cs. Descriptors preserve
// the upstream variable-node shape used by generated command grammar and
// syntax completion.

type CustomElementDescriptor struct {
	Name     string
	Hint     CompletionHint
	Optional bool
}

// CustomElementDescriptorFromOptions supplies the optional arguments of the
// source From(string, CompletionHint, bool) overload.
type CustomElementDescriptorFromOptions struct {
	Hint       CompletionHint
	IsOptional bool
}

var CustomElementDescriptorDefault = CustomElementDescriptor{
	Hint: CompletionHintSyntax,
}

// DefaultCustomElementDescriptor is the original Go spelling retained for
// compatibility.
var DefaultCustomElementDescriptor = CustomElementDescriptorDefault

func NewCustomElementDescriptor(name string, hint CompletionHint, optional bool) CustomElementDescriptor {
	return CustomElementDescriptor{Name: name, Hint: hint, Optional: optional}
}

func CustomElementDescriptorFrom(name string, options ...CustomElementDescriptorFromOptions) CustomElementDescriptor {
	if len(options) > 1 {
		panic("CustomElementDescriptorFrom accepts at most one options value")
	}
	hint := CompletionHintSyntax
	isOptional := false
	if len(options) == 1 {
		hint = options[0].Hint
		isOptional = options[0].IsOptional
	}
	return customElementDescriptorFrom(name, hint, isOptional)
}

func CustomElementDescriptorFromHint(hint CompletionHint, isOptional ...bool) CustomElementDescriptor {
	if len(isOptional) > 1 {
		panic("CustomElementDescriptorFromHint accepts at most one isOptional value")
	}
	optional := false
	if len(isOptional) == 1 {
		optional = isOptional[0]
	}
	return customElementDescriptorFrom("", hint, optional)
}

func CustomElementDescriptorFromIsOptional(isOptional bool) CustomElementDescriptor {
	return customElementDescriptorFrom("", CompletionHintSyntax, isOptional)
}

func customElementDescriptorFrom(name string, hint CompletionHint, isOptional bool) CustomElementDescriptor {
	descriptor := NewCustomElementDescriptor(name, hint, isOptional)
	if descriptor.IsDefault() {
		return CustomElementDescriptorDefault
	}
	return descriptor
}

func CustomNodeGetDefaultShape(count int) []CustomElementDescriptor {
	shape := make([]CustomElementDescriptor, count)
	for i := range shape {
		shape[i] = CustomElementDescriptorDefault
	}
	return shape
}

// CustomElementDescriptorDefaultShape is the original Go spelling retained
// for compatibility.
func CustomElementDescriptorDefaultShape(count int) []CustomElementDescriptor {
	return CustomNodeGetDefaultShape(count)
}

func (d CustomElementDescriptor) WithName(name string) CustomElementDescriptor {
	d.Name = name
	return d
}

func (d CustomElementDescriptor) WithHint(hint CompletionHint) CustomElementDescriptor {
	d.Hint = hint
	return d
}

func (d CustomElementDescriptor) WithOptional(optional bool) CustomElementDescriptor {
	d.Optional = optional
	return d
}

func (d CustomElementDescriptor) WithIsOptional(isOptional bool) CustomElementDescriptor {
	if d.Optional == isOptional {
		return d
	}
	// The pinned source accidentally passes this.IsOptional instead of the
	// argument when constructing the changed descriptor. Preserve that behavior
	// for source parity; WithOptional remains the behavior-correct Go API.
	return NewCustomElementDescriptor(d.Name, d.Hint, d.Optional)
}

func (d CustomElementDescriptor) IsDefault() bool {
	return d.Name == "" && d.Hint == CompletionHintSyntax && !d.Optional
}

type CustomNode struct {
	*SyntaxNode
	shape        []CustomElementDescriptor
	elementCount int
}

func NewCustomNode(elements ...SyntaxElement) *CustomNode {
	node, err := NewCustomNodeWithShape(CustomElementDescriptorDefaultShape(len(elements)), elements)
	if err != nil {
		panic(err)
	}
	return node
}

func NewCustomNodeWithShape(shape []CustomElementDescriptor, elements []SyntaxElement, diagnostics ...Diagnostic) (*CustomNode, error) {
	if shape == nil {
		return nil, fmt.Errorf("custom node shape is required")
	}
	if len(elements) > len(shape) {
		return nil, fmt.Errorf("custom node has %d elements for %d shape slots", len(elements), len(shape))
	}
	children := make([]SyntaxElement, len(shape))
	copy(children, elements)
	// The source constructor validates only supplied elements. Remaining shape
	// slots represent parser-recovered omissions and are exposed as nil children.
	for i, child := range elements {
		if child == nil && !shape[i].Optional {
			return nil, fmt.Errorf("custom node child %d is required", i)
		}
	}
	return &CustomNode{
		SyntaxNode:   NewSyntaxNode(SyntaxKindCustomNode, children, diagnostics...),
		shape:        append([]CustomElementDescriptor{}, shape...),
		elementCount: len(elements),
	}, nil
}

func (n *CustomNode) GetName(index int) string {
	if n == nil || index < 0 || index >= len(n.shape) {
		return ""
	}
	return n.shape[index].Name
}

func (n *CustomNode) IsOptional(index int) bool {
	if n == nil || index < 0 || index >= len(n.shape) {
		return false
	}
	return n.shape[index].Optional
}

func (n *CustomNode) CompletionHint(index int) CompletionHint {
	if n == nil || index < 0 || index >= len(n.shape) {
		return CompletionHintInherit
	}
	return n.shape[index].Hint
}

func (n *CustomNode) Shape() []CustomElementDescriptor {
	if n == nil {
		return nil
	}
	return append([]CustomElementDescriptor{}, n.shape...)
}

func (n *CustomNode) Clone(includeDiagnostics bool) SyntaxElement {
	if n == nil {
		return (*CustomNode)(nil)
	}
	clone, err := NewCustomNodeWithShape(
		n.Shape(),
		cloneSyntaxElements(n.Children()[:n.elementCount], includeDiagnostics),
		cloneDiagnostics(n.diagnostics, includeDiagnostics)...,
	)
	if err != nil {
		panic(err)
	}
	return clone
}
