package kustolang

import (
	"reflect"
	"strings"
)

// Only exact generated wrapper types may bypass Text dispatch. An external
// type embedding SyntaxNode can override Text and must retain that behavior.
// The registry is immutable after initialization and does not grow per query.
var generatedSyntaxTextTypes = func() map[reflect.Type]bool {
	types := make(map[reflect.Type]bool, len(generatedSyntaxNodeMetadata))
	for _, metadata := range generatedSyntaxNodeMetadata {
		node := &SyntaxNode{generatedNodeName: metadata.Name}
		types[reflect.TypeOf(WrapGeneratedSyntaxNode(node))] = true
	}
	return types
}()

func writeSyntaxChildrenText(b *strings.Builder, children []SyntaxElement, includeTrivia IncludeTrivia) {
	for i, child := range children {
		switch element := child.(type) {
		case nil:
			continue
		case *SyntaxToken:
			// These boundaries belong to each immediate child list, not to the
			// first and last token in the entire traversal.
			b.WriteString(element.textTrivia(includeTrivia, i == 0, i == len(children)-1))
			b.WriteString(element.text)
		case *SyntaxNode:
			writeSyntaxChildrenText(b, element.children, includeTrivia)
		case *SyntaxList:
			writeSyntaxChildrenText(b, element.children(), includeTrivia)
		case *SeparatedElement:
			writeSyntaxChildrenText(b, element.children(), includeTrivia)
		default:
			if generatedSyntaxTextTypes[reflect.TypeOf(child)] {
				node := child.(interface{ syntaxNodeRoot() *SyntaxNode }).syntaxNodeRoot()
				writeSyntaxChildrenText(b, node.children, includeTrivia)
			} else {
				b.WriteString(child.Text(includeTrivia))
			}
		}
	}
}
