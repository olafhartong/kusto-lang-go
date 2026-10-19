package kustolang

// Ported from src/Kusto.Language/Parser/Combinators/Parsers/MapParser.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type MapEntry[TInput comparable, TOutput any] struct {
	Key   []TInput
	Value func() TOutput
}

type MapParser[TInput comparable, TOutput any] struct {
	ParserInfo[TInput]
	root *mapParserNode[TInput, TOutput]
}

func NewMapParser[TInput comparable, TOutput any](entries []MapEntry[TInput, TOutput]) *MapParser[TInput, TOutput] {
	EnsureArgumentNotNil(entries, "entries")
	return &MapParser[TInput, TOutput]{
		ParserInfo: NewParserInfo(WithParserTraits[TInput](ParserTraits{Match: true})),
		root:       newMapParserNode(entries),
	}
}

func (p *MapParser[TInput, TOutput]) Description() string {
	return p.ParserInfo.Description(p)
}

func (p *MapParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	return parserWithTag(p.ParserInfo, tag, p, p.cloneWithInfo)
}

func (p *MapParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	return parserWithAnnotations(p.ParserInfo, annotations, p, p.cloneWithInfo)
}

func (p *MapParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	return parserWithIsHidden(p.ParserInfo, isHidden, p, p.cloneWithInfo)
}

func (p *MapParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *MapParser[TInput, TOutput]) ParseValue(source Source[TInput], start int) ParseResult[TOutput] {
	node := p.root
	length := 0
	bestLength := -1
	var bestOutput func() TOutput

	for {
		if source.IsEnd(start + length) {
			break
		}

		input := source.Peek(start + length)
		subNode, ok := node.tryGetValueNode(input)
		if !ok {
			break
		}

		length++
		if subNode.hasValue {
			bestLength = length
			bestOutput = subNode.value
		}
		node = subNode
	}

	if bestLength > 0 {
		return NewParseResult(bestLength, bestOutput())
	}
	var zero TOutput
	return NewParseResult(bestLength, zero)
}

func (p *MapParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	result := p.ParseValue(source, inputStart)
	if result.Length >= 0 {
		appendOutput(output, result.Value)
	}
	return result.Length
}

func (p *MapParser[TInput, TOutput]) Scan(source Source[TInput], start int) int {
	node := p.root
	length := 0
	bestLength := -1

	for {
		if source.IsEnd(start + length) {
			break
		}

		input := source.Peek(start + length)
		subNode, ok := node.tryGetValueNode(input)
		if !ok {
			break
		}

		length++
		if subNode.hasValue {
			bestLength = length
		}
		node = subNode
	}

	return bestLength
}

func (p *MapParser[TInput, TOutput]) cloneWithInfo(info ParserInfo[TInput]) Parser[TInput] {
	clone := *p
	clone.ParserInfo = info
	return &clone
}

type mapParserNode[TInput comparable, TOutput any] struct {
	hasValue bool
	value    func() TOutput
	nodes    map[TInput]*mapParserNode[TInput, TOutput]
}

func newMapParserNode[TInput comparable, TOutput any](entries []MapEntry[TInput, TOutput]) *mapParserNode[TInput, TOutput] {
	node := &mapParserNode[TInput, TOutput]{}
	for _, entry := range entries {
		node.add(entry.Key, entry.Value)
	}
	return node
}

func (n *mapParserNode[TInput, TOutput]) add(sequence []TInput, value func() TOutput) {
	node := n
	for _, item := range sequence {
		if node.nodes == nil {
			node.nodes = map[TInput]*mapParserNode[TInput, TOutput]{}
		}
		subNode, ok := node.nodes[item]
		if !ok {
			subNode = &mapParserNode[TInput, TOutput]{}
			node.nodes[item] = subNode
		}
		node = subNode
	}

	node.hasValue = true
	node.value = value
}

func (n *mapParserNode[TInput, TOutput]) tryGetValueNode(key TInput) (*mapParserNode[TInput, TOutput], bool) {
	if n.nodes == nil {
		return nil, false
	}
	node, ok := n.nodes[key]
	return node, ok
}
