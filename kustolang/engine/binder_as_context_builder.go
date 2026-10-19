package kustolang

// Ported from src/Kusto.Language/Binder/Binder_AsContextBuilder.cs.
// This helper searches syntax already carrying semantic result types for
// preceding as-operator declarations and adds them to the binder local scope.

type AsContextBuilder struct {
	position int
	binder   *Binder
}

func NewAsContextBuilder(position int, binder *Binder) *AsContextBuilder {
	return &AsContextBuilder{position: position, binder: binder}
}

func (b *Binder) AddAsContext(root SyntaxElement, position int) {
	NewAsContextBuilder(position, b).Visit(root)
}

func (b *AsContextBuilder) Visit(root SyntaxElement) {
	b.visit(root, 0)
}

func (b *AsContextBuilder) visit(element SyntaxElement, triviaStart int) {
	if element == nil {
		return
	}
	wrapped := WrapGeneratedSyntaxElement(element)
	if body, ok := wrapped.(*FunctionBody); ok {
		if b.position > elementTextStart(body, triviaStart) && b.position < elementEnd(body, triviaStart) {
			b.defaultVisit(body, triviaStart)
		}
		return
	}
	if asOperator, ok := wrapped.(*AsOperator); ok {
		b.defaultVisit(asOperator, triviaStart)
		b.visitAsOperator(asOperator, triviaStart)
		return
	}
	b.defaultVisit(wrapped, triviaStart)
}

func (b *AsContextBuilder) defaultVisit(element SyntaxElement, triviaStart int) {
	childStart := triviaStart
	for _, child := range element.Children() {
		if child != nil {
			b.visit(child, childStart)
			childStart += child.ByteFullWidth()
		}
	}
}

func (b *AsContextBuilder) visitAsOperator(node *AsOperator, triviaStart int) {
	if b == nil || b.binder == nil || node == nil {
		return
	}
	name := NameDeclarationSimpleName(node.Name())
	typ := node.RawResultType()
	if name == "" || b.position <= elementEnd(node, triviaStart) || typ == nil {
		return
	}
	if b.binder.LocalScope == nil {
		b.binder.LocalScope = NewLocalScope(nil)
	}
	b.binder.LocalScope.AddSymbol(NewVariableSymbol(name, typ, false, nil, nil))
}
