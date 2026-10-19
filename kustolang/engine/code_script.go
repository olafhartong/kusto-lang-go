package kustolang

import (
	"fmt"
	"reflect"
	"sort"
	"unicode/utf8"
)

// Ported from:
//   src/Kusto.Language/Editor/CodeBlock.cs
//   src/Kusto.Language/Editor/CodeServiceFactory.cs
//   src/Kusto.Language/Editor/CodeScript.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CodeBlock struct {
	Start       int
	codeService *OffsetCodeService
}

func NewCodeBlock(start int, service CodeService) *CodeBlock {
	if service == nil {
		service = NewUnknownCodeService("")
	}
	var offsetService *OffsetCodeService
	if existing, ok := service.(*OffsetCodeService); ok {
		offsetService = existing.WithOffset(start)
	} else {
		offsetService = NewOffsetCodeService(service, start)
	}
	return &CodeBlock{Start: start, codeService: offsetService}
}

func (b *CodeBlock) Text() string {
	if b == nil || b.codeService == nil {
		return ""
	}
	return b.codeService.Text()
}

func (b *CodeBlock) Length() int {
	text := b.Text()
	return byteOffsetToUTF16Position(text, len(text))
}

func (b *CodeBlock) End() int {
	if b == nil {
		return 0
	}
	return b.Start + b.Length()
}

func (b *CodeBlock) Kind() string {
	if b == nil || b.codeService == nil {
		return ""
	}
	return b.codeService.Kind()
}

func (b *CodeBlock) Service() CodeService {
	if b == nil {
		return nil
	}
	return b.codeService
}

func (b *CodeBlock) WithStart(start int) *CodeBlock {
	if b == nil {
		return nil
	}
	return NewCodeBlock(start, b.Service())
}

type CodeServiceFactory interface {
	TryGetCodeService(text string) (CodeService, bool)
	GetFactory(match func(CodeServiceFactory) bool) CodeServiceFactory
	WithFactory(factory CodeServiceFactory) CodeServiceFactory
	factoryKey() string
}

func CodeServiceFactoryKey(factory CodeServiceFactory) string {
	if factory == nil {
		return ""
	}
	return factory.factoryKey()
}

type CodeServiceFactoryFunc struct {
	key string
	try func(string) (CodeService, bool)
}

func NewCodeServiceFactoryFunc(key string, try func(string) (CodeService, bool)) *CodeServiceFactoryFunc {
	return &CodeServiceFactoryFunc{key: key, try: try}
}

func NewUnknownCodeServiceFactory() CodeServiceFactory {
	return NewCodeServiceFactoryFunc("UnknownCodeServiceFactory", func(text string) (CodeService, bool) {
		return NewUnknownCodeService(text), true
	})
}

func (f *CodeServiceFactoryFunc) TryGetCodeService(text string) (CodeService, bool) {
	if f == nil || f.try == nil {
		return nil, false
	}
	return f.try(text)
}

func (f *CodeServiceFactoryFunc) GetFactory(match func(CodeServiceFactory) bool) CodeServiceFactory {
	if f != nil && (match == nil || match(f)) {
		return f
	}
	return nil
}

func (f *CodeServiceFactoryFunc) WithFactory(factory CodeServiceFactory) CodeServiceFactory {
	return codeServiceFactoryWithFactory(f, factory)
}

func (f *CodeServiceFactoryFunc) factoryKey() string {
	if f != nil && f.key != "" {
		return f.key
	}
	return fmt.Sprintf("%T", f)
}

type aggregateCodeServiceFactory struct {
	factories []CodeServiceFactory
}

func newAggregateCodeServiceFactory(factories []CodeServiceFactory) *aggregateCodeServiceFactory {
	list := make([]CodeServiceFactory, 0, len(factories))
	for _, factory := range factories {
		if factory != nil {
			list = append(list, factory)
		}
	}
	return &aggregateCodeServiceFactory{factories: ListToReadOnly(list)}
}

func (f *aggregateCodeServiceFactory) TryGetCodeService(text string) (CodeService, bool) {
	if f == nil {
		return nil, false
	}
	for _, factory := range f.factories {
		if service, ok := factory.TryGetCodeService(text); ok {
			return service, true
		}
	}
	return nil, false
}

func (f *aggregateCodeServiceFactory) GetFactory(match func(CodeServiceFactory) bool) CodeServiceFactory {
	if f == nil {
		return nil
	}
	for _, factory := range f.factories {
		if match == nil || match(factory) {
			return factory
		}
	}
	return nil
}

func (f *aggregateCodeServiceFactory) WithFactory(factory CodeServiceFactory) CodeServiceFactory {
	if f == nil {
		return factory
	}
	if factory == nil {
		return f
	}
	for _, existing := range f.factories {
		if sameCodeServiceFactory(existing, factory) {
			return f
		}
	}
	list := ListToReadOnly(f.factories)
	replaced := false
	for i, existing := range list {
		if existing.factoryKey() == factory.factoryKey() {
			list[i] = factory
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, factory)
	}
	return newAggregateCodeServiceFactory(list)
}

func (f *aggregateCodeServiceFactory) factoryKey() string {
	return "AggregateCodeServiceFactory"
}

func codeServiceFactoryWithFactory(base CodeServiceFactory, factory CodeServiceFactory) CodeServiceFactory {
	if factory == nil {
		return base
	}
	if base == nil {
		return factory
	}
	if aggregate, ok := factory.(*aggregateCodeServiceFactory); ok {
		return aggregate.WithFactory(base)
	}
	if factory.factoryKey() == base.factoryKey() {
		return factory
	}
	return newAggregateCodeServiceFactory([]CodeServiceFactory{base}).WithFactory(factory)
}

func sameCodeServiceFactory(left CodeServiceFactory, right CodeServiceFactory) bool {
	if left == nil || right == nil {
		return left == right
	}
	lv := reflect.ValueOf(left)
	rv := reflect.ValueOf(right)
	return lv.IsValid() && rv.IsValid() &&
		lv.Kind() == reflect.Pointer &&
		rv.Kind() == reflect.Pointer &&
		lv.Type() == rv.Type() &&
		lv.Pointer() == rv.Pointer()
}

type BlockSeparator func(text string, lineStarts []int) []int

type CodeScript struct {
	Text      string
	Factory   CodeServiceFactory
	Separator BlockSeparator

	lineStarts     []int
	byteLineStarts []int
	blocks         []*CodeBlock
}

func NewCodeScript(text string, factory CodeServiceFactory) *CodeScript {
	return createCodeScript(text, factory, nil, nil)
}

func (s *CodeScript) Blocks() []*CodeBlock {
	if s == nil {
		return EmptyReadOnlyList[*CodeBlock]()
	}
	return ListToReadOnly(s.blocks)
}

func (s *CodeScript) LineStarts() []int {
	if s == nil {
		return EmptyReadOnlyList[int]()
	}
	return ListToReadOnly(s.lineStarts)
}

func (s *CodeScript) WithText(newText string) *CodeScript {
	if s == nil {
		return NewCodeScript(newText, nil)
	}
	return createCodeScript(newText, s.Factory, s.blocks, s.Separator)
}

func (s *CodeScript) WithFactory(factory CodeServiceFactory) *CodeScript {
	if s == nil {
		return NewCodeScript("", factory)
	}
	if !sameCodeServiceFactory(factory, s.Factory) {
		return createCodeScript(s.Text, factory, nil, s.Separator)
	}
	return s
}

func (s *CodeScript) WithSeparator(separator BlockSeparator) *CodeScript {
	if s == nil {
		return NewCodeScript("", nil).WithSeparator(separator)
	}
	if !sameBlockSeparator(separator, s.Separator) {
		return createCodeScript(s.Text, s.Factory, s.blocks, separator)
	}
	return s
}

func (s *CodeScript) WithTextAndFactory(newText string, factory CodeServiceFactory) *CodeScript {
	if s == nil {
		return NewCodeScript(newText, factory)
	}
	return s.WithText(newText).WithFactory(factory)
}

func createCodeScript(text string, factory CodeServiceFactory, existingBlocks []*CodeBlock, separator BlockSeparator) *CodeScript {
	if factory == nil {
		factory = NewCodeServiceFactoryFunc("NoCodeServiceFactory", func(string) (CodeService, bool) {
			return nil, false
		})
	}
	byteLineStarts := LineStarts(text)
	lineStarts := bytePositionsToUTF16(text, byteLineStarts)
	blockStarts := []int(nil)
	if separator != nil {
		for _, position := range separator(text, lineStarts) {
			blockStarts = append(blockStarts, utf16PositionToByteOffset(text, position))
		}
	} else {
		blockStarts = KustoBlockStarts(text, byteLineStarts)
	}
	if len(blockStarts) == 0 {
		blockStarts = []int{0}
	}

	var existingBlockMap *TextKeyedDictionary[*CodeBlock]
	if len(existingBlocks) > 0 {
		existingBlockMap = ToTextKeyedDictionary(existingBlocks,
			func(block *CodeBlock) string {
				if block == nil {
					return ""
				}
				return block.Text()
			},
			func(block *CodeBlock) *CodeBlock { return block },
		)
	}

	blocks := make([]*CodeBlock, 0, len(blockStarts))
	for i, start := range blockStarts {
		if start < 0 {
			start = 0
		}
		if start > len(text) {
			start = len(text)
		}
		length := len(text) - start
		if i+1 < len(blockStarts) {
			length = blockStarts[i+1] - start
		}
		if length < 0 {
			length = 0
		}
		if start+length > len(text) {
			length = len(text) - start
		}

		var block *CodeBlock
		publicStart := byteOffsetToUTF16Position(text, start)
		if existingBlockMap != nil {
			if existing, ok := existingBlockMap.TryGetValue(text, start, length); ok {
				block = existing.WithStart(publicStart)
			}
		}
		if block == nil {
			blockText := text[start : start+length]
			service, ok := factory.TryGetCodeService(blockText)
			if !ok {
				service = NewUnknownCodeService(blockText)
			}
			block = NewCodeBlock(publicStart, service)
		}
		blocks = append(blocks, block)
	}

	return &CodeScript{
		Text:           text,
		Factory:        factory,
		Separator:      separator,
		lineStarts:     ListToReadOnly(lineStarts),
		byteLineStarts: ListToReadOnly(byteLineStarts),
		blocks:         ListToReadOnly(blocks),
	}
}

func (s *CodeScript) GetBlockAtPosition(position int) *CodeBlock {
	if s == nil || len(s.blocks) == 0 {
		return nil
	}
	switch {
	case position == 0:
		return s.blocks[0]
	case position == byteOffsetToUTF16Position(s.Text, len(s.Text)):
		return s.blocks[len(s.blocks)-1]
	case position >= 0 && position < byteOffsetToUTF16Position(s.Text, len(s.Text)):
		index := sort.Search(len(s.blocks), func(i int) bool {
			return position < s.blocks[i].End()
		})
		if index >= 0 && index < len(s.blocks) && position >= s.blocks[index].Start {
			return s.blocks[index]
		}
	}
	return nil
}

func (s *CodeScript) TryGetTextPosition(line int, lineOffset int) (position int, ok bool) {
	if s == nil {
		return 0, false
	}
	line--
	lineOffset--
	if line < 0 || line >= len(s.lineStarts) {
		return 0, false
	}
	lineStart := s.lineStarts[line]
	byteLineStart := s.byteLineStarts[line]
	byteLineEnd := len(s.Text)
	if line < len(s.byteLineStarts)-1 {
		byteLineEnd = s.byteLineStarts[line+1]
	}
	for byteLineEnd > byteLineStart {
		r, size := utf8.DecodeLastRuneInString(s.Text[byteLineStart:byteLineEnd])
		if size == 0 || !IsLineBreakStart(r) {
			break
		}
		byteLineEnd -= size
	}
	lineEnd := byteOffsetToUTF16Position(s.Text, byteLineEnd)
	lineLength := lineEnd - lineStart
	if lineOffset < lineLength+1 {
		return lineStart + lineOffset, true
	}
	return 0, false
}

func (s *CodeScript) TryGetLineAndOffset(position int) (line int, lineOffset int, ok bool) {
	if s == nil {
		return 0, 0, false
	}
	if position < 0 || position > byteOffsetToUTF16Position(s.Text, len(s.Text)) {
		return 0, 0, false
	}
	return TryGetLineAndOffset(s.lineStarts, position)
}

func bytePositionsToUTF16(text string, positions []int) []int {
	converted := make([]int, len(positions))
	for i, position := range positions {
		converted[i] = byteOffsetToUTF16Position(text, position)
	}
	return converted
}

func sameBlockSeparator(left BlockSeparator, right BlockSeparator) bool {
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	if !leftValue.IsValid() || !rightValue.IsValid() {
		return !leftValue.IsValid() && !rightValue.IsValid()
	}
	return leftValue.Pointer() == rightValue.Pointer()
}
