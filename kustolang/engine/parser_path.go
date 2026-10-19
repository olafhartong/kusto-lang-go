package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/ParserPath.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ParserAndIndex[TInput any] struct {
	Parser        Parser[TInput]
	IndexInParent int
}

func NewParserAndIndex[TInput any](parser Parser[TInput], indexInParent int) ParserAndIndex[TInput] {
	return ParserAndIndex[TInput]{
		Parser:        parser,
		IndexInParent: indexInParent,
	}
}

type ParserPath[TInput any] struct {
	parent        *ParserPath[TInput]
	parser        Parser[TInput]
	indexInParent int
	children      []*ParserPath[TInput]
}

func NewParserPath[TInput any](parserAndIndex ParserAndIndex[TInput]) *ParserPath[TInput] {
	return newParserPath(nil, parserAndIndex)
}

func newParserPath[TInput any](parent *ParserPath[TInput], parserAndIndex ParserAndIndex[TInput]) *ParserPath[TInput] {
	return &ParserPath[TInput]{
		parent:        parent,
		parser:        parserAndIndex.Parser,
		indexInParent: parserAndIndex.IndexInParent,
	}
}

func (p *ParserPath[TInput]) Parser() Parser[TInput] {
	if p == nil {
		return nil
	}
	return p.parser
}

func (p *ParserPath[TInput]) Parent() *ParserPath[TInput] {
	if p == nil {
		return nil
	}
	return p.parent
}

func (p *ParserPath[TInput]) IndexInParent() int {
	if p == nil {
		return -1
	}
	return p.indexInParent
}

func (p *ParserPath[TInput]) ChildCount() int {
	if p == nil || p.parser == nil {
		return 0
	}
	return p.parser.ChildParserCount()
}

func (p *ParserPath[TInput]) GetChild(index int) *ParserPath[TInput] {
	if p == nil || index < 0 || index >= p.ChildCount() {
		return nil
	}
	if p.children == nil && p.ChildCount() > 0 {
		children := make([]*ParserPath[TInput], p.ChildCount())
		for i := range children {
			children[i] = newParserPath(p, NewParserAndIndex(p.parser.ChildParser(i), i))
		}
		p.children = children
	}
	return p.children[index]
}

func (p *ParserPath[TInput]) DebugText() string {
	if p == nil || p.parser == nil {
		return ""
	}
	description := p.parser.Description()
	if p.parent == nil {
		return description
	}
	return p.parent.DebugText() + "/" + description
}

func ParserPathFrom[TInput any](path []ParserAndIndex[TInput]) *ParserPath[TInput] {
	var current *ParserPath[TInput]
	for _, parserAndIndex := range path {
		current = newParserPath(current, parserAndIndex)
	}
	return current
}
