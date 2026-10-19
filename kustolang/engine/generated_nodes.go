package kustolang

import (
	"fmt"
	"reflect"
)

// GeneratedSyntaxNodeChild describes one child slot from upstream
// GeneratedSyntaxNodes.cs metadata.
type GeneratedSyntaxNodeChild struct {
	Name     string
	Type     string
	Optional bool
	Hint     CompletionHint
}

// GeneratedSyntaxNodePayload describes a non-syntax property stored on an
// upstream generated syntax node, such as NameReference.Match.
type GeneratedSyntaxNodePayload struct {
	Name string
	Type string
}

// GeneratedSyntaxNodeMetadata describes one concrete upstream syntax node.
type GeneratedSyntaxNodeMetadata struct {
	Kind      SyntaxKind
	KindFixed bool
	Name      string
	BaseType  string
	Children  []GeneratedSyntaxNodeChild
	Payloads  []GeneratedSyntaxNodePayload
}

var generatedSyntaxNodeByKind map[SyntaxKind]GeneratedSyntaxNodeMetadata
var generatedSyntaxNodesByKind map[SyntaxKind][]GeneratedSyntaxNodeMetadata
var generatedSyntaxNodeByName map[string]GeneratedSyntaxNodeMetadata

func init() {
	applyCompatibilityGeneratedSyntaxCompatibility()
	generatedSyntaxNodeByKind = map[SyntaxKind]GeneratedSyntaxNodeMetadata{}
	generatedSyntaxNodesByKind = map[SyntaxKind][]GeneratedSyntaxNodeMetadata{}
	generatedSyntaxNodeByName = map[string]GeneratedSyntaxNodeMetadata{}
	for _, node := range generatedSyntaxNodeMetadata {
		if node.KindFixed {
			if _, ok := generatedSyntaxNodeByKind[node.Kind]; !ok {
				generatedSyntaxNodeByKind[node.Kind] = node
			}
			generatedSyntaxNodesByKind[node.Kind] = append(generatedSyntaxNodesByKind[node.Kind], node)
		}
		generatedSyntaxNodeByName[node.Name] = node
	}
}

func applyCompatibilityGeneratedSyntaxCompatibility() {
	// the upstream compatibility implementation allowed a lone `to typeof(...)` mv-expand
	// expression with a nil Expression child. Current upstream requires it.
	for nodeIndex := range generatedSyntaxNodeMetadata {
		node := &generatedSyntaxNodeMetadata[nodeIndex]
		if node.Name != "MvExpandExpression" {
			continue
		}
		for childIndex := range node.Children {
			if node.Children[childIndex].Name == "Expression" {
				node.Children[childIndex].Optional = true
				return
			}
		}
	}
}

func GeneratedSyntaxNodeMetadataForKind(kind SyntaxKind) (GeneratedSyntaxNodeMetadata, bool) {
	node, ok := generatedSyntaxNodeMetadataForKind(kind)
	if !ok {
		return GeneratedSyntaxNodeMetadata{}, false
	}
	return cloneGeneratedSyntaxNodeMetadata(node), true
}

func GeneratedSyntaxNodeMetadataAllForKind(kind SyntaxKind) []GeneratedSyntaxNodeMetadata {
	nodes := generatedSyntaxNodesByKind[kind]
	out := make([]GeneratedSyntaxNodeMetadata, len(nodes))
	for i, node := range nodes {
		out[i] = cloneGeneratedSyntaxNodeMetadata(node)
	}
	return out
}

func GeneratedSyntaxNodeMetadataForName(name string) (GeneratedSyntaxNodeMetadata, bool) {
	node, ok := generatedSyntaxNodeMetadataForName(name)
	if !ok {
		return GeneratedSyntaxNodeMetadata{}, false
	}
	return cloneGeneratedSyntaxNodeMetadata(node), true
}

// Internal metadata readers share tables finalized during init. Neither these
// readers nor node construction may mutate their child or payload slices.
// Exported accessors continue to return independent copies.
func generatedSyntaxNodeMetadataForKind(kind SyntaxKind) (GeneratedSyntaxNodeMetadata, bool) {
	node, ok := generatedSyntaxNodeByKind[kind]
	return node, ok
}

func generatedSyntaxNodeMetadataForName(name string) (GeneratedSyntaxNodeMetadata, bool) {
	node, ok := generatedSyntaxNodeByName[name]
	return node, ok
}

func GeneratedSyntaxNodeMetadataAll() []GeneratedSyntaxNodeMetadata {
	out := make([]GeneratedSyntaxNodeMetadata, len(generatedSyntaxNodeMetadata))
	for i, node := range generatedSyntaxNodeMetadata {
		out[i] = cloneGeneratedSyntaxNodeMetadata(node)
	}
	return out
}

func NewGeneratedSyntaxNode(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	metadata, ok := generatedSyntaxNodeMetadataForKind(kind)
	if !ok {
		return nil, fmt.Errorf("syntax kind %d has no generated node metadata", kind)
	}
	return newGeneratedSyntaxNodeFromMetadata(metadata, kind, children, diagnostics...)
}

func NewGeneratedSyntaxNodeByName(name string, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	metadata, ok := generatedSyntaxNodeMetadataForName(name)
	if !ok {
		return nil, fmt.Errorf("syntax node %q has no generated node metadata", name)
	}
	if !metadata.KindFixed {
		return nil, fmt.Errorf("syntax node %q requires an explicit syntax kind", name)
	}
	return newGeneratedSyntaxNodeFromMetadata(metadata, metadata.Kind, children, diagnostics...)
}

func NewGeneratedSyntaxNodeByNameWithKind(name string, kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	metadata, ok := generatedSyntaxNodeMetadataForName(name)
	if !ok {
		return nil, fmt.Errorf("syntax node %q has no generated node metadata", name)
	}
	if metadata.KindFixed && kind != metadata.Kind {
		return nil, fmt.Errorf("syntax node %q has fixed kind %d, got %d", name, metadata.Kind, kind)
	}
	return newGeneratedSyntaxNodeFromMetadata(metadata, kind, children, diagnostics...)
}

func newGeneratedSyntaxNodeFromMetadata(metadata GeneratedSyntaxNodeMetadata, kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) (*SyntaxNode, error) {
	if len(children) != len(metadata.Children) {
		return nil, fmt.Errorf("%s expects %d children, got %d", metadata.Name, len(metadata.Children), len(children))
	}
	for i, child := range children {
		if child == nil && !metadata.Children[i].Optional {
			return nil, fmt.Errorf("%s child %s at index %d is required", metadata.Name, metadata.Children[i].Name, i)
		}
	}
	return newSyntaxNode(kind, metadata.Name, children, diagnostics...), nil
}

func MustNewGeneratedSyntaxNode(kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	node, err := NewGeneratedSyntaxNode(kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func MustNewGeneratedSyntaxNodeByName(name string, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	node, err := NewGeneratedSyntaxNodeByName(name, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func MustNewGeneratedSyntaxNodeByNameWithKind(name string, kind SyntaxKind, children []SyntaxElement, diagnostics ...Diagnostic) *SyntaxNode {
	node, err := NewGeneratedSyntaxNodeByNameWithKind(name, kind, children, diagnostics...)
	if err != nil {
		panic(err)
	}
	return node
}

func syntaxElementFromSyntaxListOf[T SyntaxElement](list *SyntaxListOf[T]) SyntaxElement {
	if list == nil {
		return nil
	}
	return list.SyntaxList
}

func syntaxElementFromSeparatedSyntaxListOf[T SyntaxElement](list *SeparatedSyntaxListOf[T]) SyntaxElement {
	if list == nil {
		return nil
	}
	return list.SyntaxList
}

func syntaxElementFromTyped[T SyntaxElement](element T) SyntaxElement {
	if isNilSyntaxElement(element) {
		return nil
	}
	return element
}

func isNilSyntaxElement(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func cloneGeneratedSyntaxNodeMetadata(node GeneratedSyntaxNodeMetadata) GeneratedSyntaxNodeMetadata {
	node.Children = append([]GeneratedSyntaxNodeChild{}, node.Children...)
	node.Payloads = append([]GeneratedSyntaxNodePayload{}, node.Payloads...)
	return node
}
