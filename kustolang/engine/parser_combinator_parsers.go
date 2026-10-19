package kustolang

import (
	"fmt"
	"reflect"
)

// Ported from:
//   src/Kusto.Language/Parser/Combinators/Parser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ResultPrimaryParser.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/ListPrimaryParser.cs
//   src/Kusto.Language/Parser/Combinators/RightParser.cs
//   src/Kusto.Language/Parser/Combinators/ParserExtensions.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type Parser[TInput any] interface {
	Tag() string
	Annotations() []any
	Description() string
	IsHidden() bool
	WithTag(tag string) Parser[TInput]
	WithAnnotations(annotations []any) Parser[TInput]
	WithIsHidden(isHidden bool) Parser[TInput]
	Hide() Parser[TInput]
	IsOptional() bool
	IsRequired() bool
	IsAlternation() bool
	IsSequence() bool
	IsRepetition() bool
	IsMatch() bool
	IsForward() bool
	IsNegation() bool
	IsConditional() bool
	ChildParserCount() int
	ChildParser(index int) Parser[TInput]
	ChildParserIndex(child Parser[TInput]) int
	ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int
	Scan(source Source[TInput], inputStart int) int
}

type ResultParser[TInput any, TOutput any] interface {
	Parser[TInput]
	ParseValue(source Source[TInput], inputStart int) ParseResult[TOutput]
}

type ParserTraits struct {
	Optional    bool
	Required    bool
	Alternation bool
	Sequence    bool
	Repetition  bool
	Match       bool
	Forward     bool
	Negation    bool
	Conditional bool
}

type ParserInfo[TInput any] struct {
	tag         string
	annotations []any
	description string
	hidden      bool
	traits      ParserTraits
	children    []Parser[TInput]
}

func NewParserInfo[TInput any](options ...ParserOption[TInput]) ParserInfo[TInput] {
	config := newParserConfig(options...)
	return config.info
}

func (i ParserInfo[TInput]) Tag() string {
	return i.tag
}

func (i ParserInfo[TInput]) Annotations() []any {
	return cloneAnySlice(i.annotations)
}

func (i ParserInfo[TInput]) Description(owner Parser[TInput]) string {
	base := i.description
	if base == "" {
		base = DescribeParser(owner)
		if base == "" {
			base = fmt.Sprintf("%T", owner)
		}
	}
	if i.tag == "" {
		return base
	}
	untagged := owner.WithTag("")
	if untaggedDescription := DescribeParser(untagged); untaggedDescription != "" {
		base = untaggedDescription
	}
	return i.tag + ": " + base
}

func (i ParserInfo[TInput]) IsHidden() bool {
	return i.hidden
}

func (i ParserInfo[TInput]) IsOptional() bool {
	return i.traits.Optional
}

func (i ParserInfo[TInput]) IsRequired() bool {
	return i.traits.Required
}

func (i ParserInfo[TInput]) IsAlternation() bool {
	return i.traits.Alternation
}

func (i ParserInfo[TInput]) IsSequence() bool {
	return i.traits.Sequence
}

func (i ParserInfo[TInput]) IsRepetition() bool {
	return i.traits.Repetition
}

func (i ParserInfo[TInput]) IsMatch() bool {
	return i.traits.Match
}

func (i ParserInfo[TInput]) IsForward() bool {
	return i.traits.Forward
}

func (i ParserInfo[TInput]) IsNegation() bool {
	return i.traits.Negation
}

func (i ParserInfo[TInput]) IsConditional() bool {
	return i.traits.Conditional
}

func (i ParserInfo[TInput]) ChildParserCount() int {
	return len(i.children)
}

func (i ParserInfo[TInput]) ChildParser(index int) Parser[TInput] {
	if index < 0 || index >= len(i.children) {
		return nil
	}
	return i.children[index]
}

func (i ParserInfo[TInput]) ChildParserIndex(child Parser[TInput]) int {
	for index, parser := range i.children {
		if parser == child {
			return index
		}
	}
	return -1
}

func (i ParserInfo[TInput]) WithTag(tag string) (ParserInfo[TInput], bool) {
	if tag == i.tag {
		return i, false
	}
	i.tag = tag
	return i, true
}

func (i ParserInfo[TInput]) WithAnnotations(annotations []any) (ParserInfo[TInput], bool) {
	list := cloneAnySlice(annotations)
	if reflect.DeepEqual(i.annotations, list) {
		return i, false
	}
	i.annotations = list
	return i, true
}

func (i ParserInfo[TInput]) WithIsHidden(isHidden bool) (ParserInfo[TInput], bool) {
	if isHidden == i.hidden {
		return i, false
	}
	i.hidden = isHidden
	return i, true
}

type ParserOption[TInput any] func(*parserConfig[TInput])

type parserConfig[TInput any] struct {
	info ParserInfo[TInput]
	scan func(Source[TInput], int) int
}

func WithParserTag[TInput any](tag string) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.tag = tag
	}
}

func WithParserAnnotations[TInput any](annotations []any) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.annotations = cloneAnySlice(annotations)
	}
}

func WithParserDescription[TInput any](description string) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.description = description
	}
}

func WithParserHidden[TInput any](hidden bool) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.hidden = hidden
	}
}

func WithParserTraits[TInput any](traits ParserTraits) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.traits = traits
	}
}

func WithParserChildren[TInput any](children ...Parser[TInput]) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.info.children = append([]Parser[TInput]{}, children...)
	}
}

func WithParserScan[TInput any](scan func(Source[TInput], int) int) ParserOption[TInput] {
	return func(config *parserConfig[TInput]) {
		config.scan = scan
	}
}

func newParserConfig[TInput any](options ...ParserOption[TInput]) parserConfig[TInput] {
	var config parserConfig[TInput]
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config
}

type ResultPrimaryParser[TInput any, TOutput any] struct {
	info  ParserInfo[TInput]
	parse func(Source[TInput], int) ParseResult[TOutput]
	scan  func(Source[TInput], int) int
}

func NewResultPrimaryParser[TInput any, TOutput any](parse func(Source[TInput], int) ParseResult[TOutput], options ...ParserOption[TInput]) *ResultPrimaryParser[TInput, TOutput] {
	config := newParserConfig(options...)
	return &ResultPrimaryParser[TInput, TOutput]{
		info:  config.info,
		parse: parse,
		scan:  config.scan,
	}
}

func (p *ResultPrimaryParser[TInput, TOutput]) Tag() string {
	return p.info.Tag()
}

func (p *ResultPrimaryParser[TInput, TOutput]) Annotations() []any {
	return p.info.Annotations()
}

func (p *ResultPrimaryParser[TInput, TOutput]) Description() string {
	return p.info.Description(p)
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsHidden() bool {
	return p.info.IsHidden()
}

func (p *ResultPrimaryParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	info, changed := p.info.WithTag(tag)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ResultPrimaryParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	info, changed := p.info.WithAnnotations(annotations)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ResultPrimaryParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	info, changed := p.info.WithIsHidden(isHidden)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ResultPrimaryParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsOptional() bool {
	return p.info.IsOptional()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsRequired() bool {
	return p.info.IsRequired()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsAlternation() bool {
	return p.info.IsAlternation()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsSequence() bool {
	return p.info.IsSequence()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsRepetition() bool {
	return p.info.IsRepetition()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsMatch() bool {
	return p.info.IsMatch()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsForward() bool {
	return p.info.IsForward()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsNegation() bool {
	return p.info.IsNegation()
}

func (p *ResultPrimaryParser[TInput, TOutput]) IsConditional() bool {
	return p.info.IsConditional()
}

func (p *ResultPrimaryParser[TInput, TOutput]) ChildParserCount() int {
	return p.info.ChildParserCount()
}

func (p *ResultPrimaryParser[TInput, TOutput]) ChildParser(index int) Parser[TInput] {
	return p.info.ChildParser(index)
}

func (p *ResultPrimaryParser[TInput, TOutput]) ChildParserIndex(child Parser[TInput]) int {
	return p.info.ChildParserIndex(child)
}

func (p *ResultPrimaryParser[TInput, TOutput]) ParseValue(source Source[TInput], inputStart int) ParseResult[TOutput] {
	if p.parse == nil {
		var zero TOutput
		return NewParseResult(-1, zero)
	}
	return p.parse(source, inputStart)
}

func (p *ResultPrimaryParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	result := p.ParseValue(source, inputStart)
	if result.Length >= 0 && output != nil {
		*output = append(*output, result.Value)
	}
	return result.Length
}

func (p *ResultPrimaryParser[TInput, TOutput]) Scan(source Source[TInput], inputStart int) int {
	if p.scan != nil {
		return p.scan(source, inputStart)
	}
	return p.ParseValue(source, inputStart).Length
}

type ListPrimaryParser[TInput any, TOutput any] struct {
	info  ParserInfo[TInput]
	parse func(Source[TInput], int, *[]any, int) int
	scan  func(Source[TInput], int) int
}

func NewListPrimaryParser[TInput any, TOutput any](parse func(Source[TInput], int, *[]any, int) int, options ...ParserOption[TInput]) *ListPrimaryParser[TInput, TOutput] {
	config := newParserConfig(options...)
	return &ListPrimaryParser[TInput, TOutput]{
		info:  config.info,
		parse: parse,
		scan:  config.scan,
	}
}

func (p *ListPrimaryParser[TInput, TOutput]) Tag() string {
	return p.info.Tag()
}

func (p *ListPrimaryParser[TInput, TOutput]) Annotations() []any {
	return p.info.Annotations()
}

func (p *ListPrimaryParser[TInput, TOutput]) Description() string {
	return p.info.Description(p)
}

func (p *ListPrimaryParser[TInput, TOutput]) IsHidden() bool {
	return p.info.IsHidden()
}

func (p *ListPrimaryParser[TInput, TOutput]) WithTag(tag string) Parser[TInput] {
	info, changed := p.info.WithTag(tag)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ListPrimaryParser[TInput, TOutput]) WithAnnotations(annotations []any) Parser[TInput] {
	info, changed := p.info.WithAnnotations(annotations)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ListPrimaryParser[TInput, TOutput]) WithIsHidden(isHidden bool) Parser[TInput] {
	info, changed := p.info.WithIsHidden(isHidden)
	if !changed {
		return p
	}
	clone := *p
	clone.info = info
	return &clone
}

func (p *ListPrimaryParser[TInput, TOutput]) Hide() Parser[TInput] {
	return p.WithIsHidden(true)
}

func (p *ListPrimaryParser[TInput, TOutput]) IsOptional() bool {
	return p.info.IsOptional()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsRequired() bool {
	return p.info.IsRequired()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsAlternation() bool {
	return p.info.IsAlternation()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsSequence() bool {
	return p.info.IsSequence()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsRepetition() bool {
	return p.info.IsRepetition()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsMatch() bool {
	return p.info.IsMatch()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsForward() bool {
	return p.info.IsForward()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsNegation() bool {
	return p.info.IsNegation()
}

func (p *ListPrimaryParser[TInput, TOutput]) IsConditional() bool {
	return p.info.IsConditional()
}

func (p *ListPrimaryParser[TInput, TOutput]) ChildParserCount() int {
	return p.info.ChildParserCount()
}

func (p *ListPrimaryParser[TInput, TOutput]) ChildParser(index int) Parser[TInput] {
	return p.info.ChildParser(index)
}

func (p *ListPrimaryParser[TInput, TOutput]) ChildParserIndex(child Parser[TInput]) int {
	return p.info.ChildParserIndex(child)
}

func (p *ListPrimaryParser[TInput, TOutput]) ParseValue(source Source[TInput], inputStart int) ParseResult[TOutput] {
	list := listPrimaryOutputListPool.AllocateFromPool()
	defer listPrimaryOutputListPool.ReturnToPool(list)

	n := p.ParseList(source, inputStart, list, 0)
	var value TOutput
	if n >= 0 && len(*list) > 0 {
		value = (*list)[0].(TOutput)
	}
	return NewParseResult(n, value)
}

func (p *ListPrimaryParser[TInput, TOutput]) ParseList(source Source[TInput], inputStart int, output *[]any, outputStart int) int {
	if p.parse == nil {
		return -1
	}
	return p.parse(source, inputStart, output, outputStart)
}

func (p *ListPrimaryParser[TInput, TOutput]) Scan(source Source[TInput], inputStart int) int {
	if p.scan != nil {
		return p.scan(source, inputStart)
	}
	result := p.ParseValue(source, inputStart)
	return result.Length
}

type RightParser[TInput any, TOutput any] struct {
	parser ResultParser[TInput, TOutput]
}

func NewRightParser[TInput any, TOutput any](parser ResultParser[TInput, TOutput]) RightParser[TInput, TOutput] {
	return RightParser[TInput, TOutput]{parser: parser}
}

func (p RightParser[TInput, TOutput]) Parser() ResultParser[TInput, TOutput] {
	return p.parser
}

func (p RightParser[TInput, TOutput]) WithTag(tag string) RightParser[TInput, TOutput] {
	return RightParser[TInput, TOutput]{parser: asResultParser[TInput, TOutput](p.parser.WithTag(tag))}
}

func (p RightParser[TInput, TOutput]) WithAnnotations(annotations []any) RightParser[TInput, TOutput] {
	return RightParser[TInput, TOutput]{parser: asResultParser[TInput, TOutput](p.parser.WithAnnotations(annotations))}
}

func (p RightParser[TInput, TOutput]) WithIsHidden(isHidden bool) RightParser[TInput, TOutput] {
	return RightParser[TInput, TOutput]{parser: asResultParser[TInput, TOutput](p.parser.WithIsHidden(isHidden))}
}

func (p RightParser[TInput, TOutput]) Hide() RightParser[TInput, TOutput] {
	return p.WithIsHidden(true)
}

func ParseText(parser Parser[byte], text string, output *[]any) int {
	return parser.ParseList(NewTextSource(text), 0, output, 0)
}

func ParseTextValue[TOutput any](parser ResultParser[byte, TOutput], text string) ParseResult[TOutput] {
	return parser.ParseValue(NewTextSource(text), 0)
}

func TryParseTextValue[TOutput any](parser ResultParser[byte, TOutput], text string) (TOutput, bool) {
	result := ParseTextValue(parser, text)
	return result.Value, result.Length > 0
}

func asResultParser[TInput any, TOutput any](parser Parser[TInput]) ResultParser[TInput, TOutput] {
	result, ok := parser.(ResultParser[TInput, TOutput])
	if !ok {
		panic(fmt.Sprintf("%T is not a ResultParser", parser))
	}
	return result
}

func cloneAnySlice(values []any) []any {
	if len(values) == 0 {
		return EmptyReadOnlyList[any]()
	}
	clone := make([]any, len(values))
	copy(clone, values)
	return clone
}

var listPrimaryOutputListPool = NewObjectPool(
	func() *[]any {
		list := make([]any, 0, 8)
		return &list
	},
	func(list *[]any) {
		if list == nil {
			return
		}
		clear(*list)
		*list = (*list)[:0]
	},
	50,
)
