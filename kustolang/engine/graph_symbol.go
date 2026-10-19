package kustolang

// Ported from src/Kusto.Language/Symbols/GraphSymbol.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var EmptyTableSymbol = NewTableSymbol("")

type GraphSymbol struct {
	TypeSymbol
	EdgeShape *TableSymbol
	NodeShape *TableSymbol
}

func NewGraphSymbol(name string, edgeShape *TableSymbol, nodeShape *TableSymbol) *GraphSymbol {
	if edgeShape == nil {
		edgeShape = EmptyTableSymbol
	}
	return &GraphSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		EdgeShape:  edgeShape,
		NodeShape:  nodeShape,
	}
}

func NewGraphSymbolFromNodeShapes(name string, edgeShape *TableSymbol, nodeShapes []*TableSymbol) *GraphSymbol {
	if edgeShape == nil {
		edgeShape = EmptyTableSymbol
	}
	var nodeShape *TableSymbol
	if len(nodeShapes) > 0 {
		nodeShape = CombineTableSymbols(nodeShapes...)
	}
	return &GraphSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		EdgeShape:  edgeShape,
		NodeShape:  nodeShape,
	}
}

func NewGraphSymbolFromShapes(name string, edgeShapes []*TableSymbol, nodeShapes []*TableSymbol) *GraphSymbol {
	var edgeShape *TableSymbol
	if len(edgeShapes) > 0 {
		edgeShape = CombineTableSymbols(edgeShapes...)
	}
	var nodeShape *TableSymbol
	if len(nodeShapes) > 0 {
		nodeShape = CombineTableSymbols(nodeShapes...)
	}
	return &GraphSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		EdgeShape:  edgeShape,
		NodeShape:  nodeShape,
	}
}

func NewGraphSymbolFromSchema(name string, edgeSchema string, nodeSchema ...string) (*GraphSymbol, error) {
	edgeShape, err := TableSymbolFromSchema(edgeSchema)
	if err != nil {
		return nil, err
	}
	var nodeShape *TableSymbol
	if len(nodeSchema) > 0 {
		nodeShape, err = TableSymbolFromSchema(nodeSchema[0])
		if err != nil {
			return nil, err
		}
	}
	return NewGraphSymbol(name, edgeShape, nodeShape), nil
}

func GraphSymbolFromSchema(edgeSchema string, nodeSchema ...string) (*GraphSymbol, error) {
	return NewGraphSymbolFromSchema("", edgeSchema, nodeSchema...)
}

func (s *GraphSymbol) SymbolKind() SymbolKind {
	return SymbolKindGraph
}

func (s *GraphSymbol) SymbolTabularity() Tabularity {
	return TabularityOther
}

func (s *GraphSymbol) WithName(name string) *GraphSymbol {
	if s == nil {
		return nil
	}
	return NewGraphSymbol(name, s.EdgeShape, s.NodeShape)
}

func (s *GraphSymbol) WithEdgeShape(edgeShape *TableSymbol) *GraphSymbol {
	if s == nil {
		return nil
	}
	if s.EdgeShape == edgeShape {
		return s
	}
	return NewGraphSymbol(s.SymbolName(), edgeShape, s.NodeShape)
}

func (s *GraphSymbol) WithNodeShape(nodeShape *TableSymbol) *GraphSymbol {
	if s == nil {
		return nil
	}
	if s.NodeShape == nil {
		return s
	}
	return NewGraphSymbol(s.SymbolName(), s.EdgeShape, nodeShape)
}

func MergeGraphSymbols(leftGraph *GraphSymbol, rightGraph *GraphSymbol) *GraphSymbol {
	switch {
	case leftGraph != nil && rightGraph != nil:
		newEdgeShape := CombineTableSymbols(leftGraph.EdgeShape, rightGraph.EdgeShape)
		var newNodeShape *TableSymbol
		switch {
		case leftGraph.NodeShape != nil && rightGraph.NodeShape != nil:
			newNodeShape = CombineTableSymbols(leftGraph.NodeShape, rightGraph.NodeShape)
		case leftGraph.NodeShape != nil:
			newNodeShape = leftGraph.NodeShape
		case rightGraph.NodeShape != nil:
			newNodeShape = rightGraph.NodeShape
		}
		return NewGraphSymbol("", newEdgeShape, newNodeShape)
	case leftGraph != nil:
		return leftGraph
	case rightGraph != nil:
		return rightGraph
	default:
		return nil
	}
}

func CombineTableSymbols(tables ...*TableSymbol) *TableSymbol {
	return CombineTableSymbolsWithKind(CombineKindUnifySameName, tables...)
}

var EmptyGraphSymbol = NewGraphSymbol("", EmptyTableSymbol, EmptyTableSymbol)
