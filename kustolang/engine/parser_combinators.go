package kustolang

import (
	"reflect"
	"sync"
)

// Ported from:
//   src/Kusto.Language/Parser/Combinators/ParseResult.cs
//   src/Kusto.Language/Parser/Combinators/OffsetValue.cs
//   src/Kusto.Language/Parser/Combinators/Source.cs
//   src/Kusto.Language/Parser/Combinators/ArraySource.cs
//   src/Kusto.Language/Parser/Combinators/TextSource.cs
//   src/Kusto.Language/Parser/Combinators/LimitSource.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/SourceConsumer.cs
//   src/Kusto.Language/Parser/Combinators/Parsers/SourceProducer.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ParseResult[T any] struct {
	Length int
	Value  T
}

func NewParseResult[T any](length int, value T) ParseResult[T] {
	return ParseResult[T]{Length: length, Value: value}
}

func (r ParseResult[T]) Succeeded() bool {
	return r.Length >= 0
}

type OffsetValue[T any] struct {
	Offset int
	Value  T
}

func NewOffsetValue[T any](offset int, value T) OffsetValue[T] {
	return OffsetValue[T]{Offset: offset, Value: value}
}

type Source[T any] interface {
	Peek(n ...int) T
	IsEnd(n ...int) bool
	Cache() *SourceCache
}

type cancellationSource[T any] struct {
	Source[T]
	cancellation CancellationToken
	work         uint64
}

func newCancellationSource[T any](source Source[T], cancellationToken CancellationToken) Source[T] {
	return &cancellationSource[T]{Source: source, cancellation: cancellationToken}
}

func (s *cancellationSource[T]) poll() {
	s.work++
	if s.work&255 == 0 {
		s.cancellation.ThrowIfCancellationRequested()
	}
}

func (s *cancellationSource[T]) Peek(n ...int) T {
	s.poll()
	return s.Source.Peek(n...)
}

func (s *cancellationSource[T]) IsEnd(n ...int) bool {
	s.poll()
	return s.Source.IsEnd(n...)
}

type SourceConsumer[TInput any] func(source Source[TInput], start int) int

type SourceProducer[TInput any, TOutput any] func(source Source[TInput], start int, length int) TOutput

type sourceBase struct {
	mu    sync.Mutex
	cache *SourceCache
}

func (s *sourceBase) Cache() *SourceCache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache == nil {
		s.cache = NewSourceCache()
	}
	return s.cache
}

type SourceCache struct {
	values map[reflect.Type]any
}

func NewSourceCache() *SourceCache {
	return &SourceCache{values: map[reflect.Type]any{}}
}

func SourceCacheGetOrCreate[T any](cache *SourceCache) T {
	var zero T
	if cache == nil {
		return zero
	}

	key := sourceCacheKey[T]()
	if value, ok := cache.values[key]; ok {
		return value.(T)
	}

	value := sourceCacheNewValue[T](key)
	cache.values[key] = value
	return value
}

func SourceCacheGetOrCreateFunc[T any](cache *SourceCache, creator func() T) T {
	var zero T
	if cache == nil {
		return zero
	}

	key := sourceCacheKey[T]()
	if value, ok := cache.values[key]; ok {
		return value.(T)
	}

	// The upstream source invokes the factory twice on a miss and stores the
	// second value. Preserve that observable behavior for parity.
	_ = creator()
	value := creator()
	cache.values[key] = value
	return value
}

func SourceCacheTryGetValue[T any](cache *SourceCache) (T, bool) {
	var zero T
	if cache == nil {
		return zero, false
	}
	key := sourceCacheKey[T]()
	if value, ok := cache.values[key]; ok {
		return value.(T), true
	}
	return zero, false
}

func sourceCacheKey[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func sourceCacheNewValue[T any](typ reflect.Type) T {
	if typ.Kind() == reflect.Pointer {
		return reflect.New(typ.Elem()).Interface().(T)
	}
	return reflect.New(typ).Elem().Interface().(T)
}

type ArraySource[T any] struct {
	sourceBase
	input  []T
	offset int
	end    int
}

func NewArraySource[T any](input []T, startAndLength ...int) *ArraySource[T] {
	start := 0
	length := -1
	if len(startAndLength) > 0 {
		start = startAndLength[0]
	}
	if len(startAndLength) > 1 {
		length = startAndLength[1]
	}

	end := len(input)
	if length >= 0 {
		remaining := len(input) - start
		if remaining < 0 {
			remaining = 0
		}
		if length < remaining {
			end = start + length
		}
	}

	return &ArraySource[T]{input: input, offset: start, end: end}
}

func (s *ArraySource[T]) Peek(n ...int) T {
	var zero T
	offset := optionalOffset(n)
	if s == nil || s.offset+offset >= s.end || s.offset+offset < 0 {
		return zero
	}
	return s.input[s.offset+offset]
}

func (s *ArraySource[T]) IsEnd(n ...int) bool {
	offset := optionalOffset(n)
	return s == nil || s.offset+offset >= s.end
}

type TextSource struct {
	sourceBase
	source  string
	offset  int
	end     int
	strings *StringTable
}

func NewTextSource(source string, offsetAndLength ...int) *TextSource {
	offset := 0
	length := len(source)
	if len(offsetAndLength) > 0 {
		offset = offsetAndLength[0]
		length = len(source) - offset
	}
	if len(offsetAndLength) > 1 {
		length = offsetAndLength[1]
	}
	return &TextSource{source: source, offset: offset, end: offset + length}
}

func (s *TextSource) Peek(n ...int) byte {
	offset := optionalOffset(n)
	if s == nil || s.offset+offset >= s.end || s.offset+offset < 0 {
		return 0
	}
	return s.source[s.offset+offset]
}

func (s *TextSource) IsEnd(n ...int) bool {
	offset := optionalOffset(n)
	return s == nil || s.offset+offset >= s.end
}

func (s *TextSource) PeekText(args ...int) string {
	start := 0
	length := 0
	if len(args) == 1 {
		length = args[0]
	} else if len(args) == 2 {
		start = args[0]
		length = args[1]
	} else {
		panic(ArgumentError{ParameterName: "args", Message: "PeekText expects length or start and length."})
	}

	if s.strings == nil {
		s.strings = NewStringTable()
	}
	return s.strings.AddRange(s.source, s.offset+start, length)
}

func (s *TextSource) Matches(start int, text string, ignoreCase ...bool) bool {
	if s == nil {
		return false
	}
	offs := s.offset + start
	if len(ignoreCase) > 0 {
		if ignoreCase[0] {
			return offs >= 0 && offs <= s.end && s.end <= len(s.source) && scanOrdinalText(s, start, text) >= 0
		}
		return s.matchesAt(offs, text, false)
	}
	if offs < 0 || offs >= s.end || len(text) == 0 || s.source[offs] != text[0] {
		return false
	}
	return s.matchesAt(offs, text, false)
}

func (s *TextSource) matchesAt(offset int, text string, ignoreCase bool) bool {
	if s == nil || offset < 0 || offset > s.end || len(text) > s.end-offset || offset+len(text) > len(s.source) {
		return false
	}
	got := s.source[offset : offset+len(text)]
	if ignoreCase {
		return ordinalEqualFold(got, text)
	}
	return got == text
}

func (s *TextSource) Position() int {
	if s == nil {
		return 0
	}
	return s.offset
}

type LimitSource[T any] struct {
	sourceBase
	source Source[T]
	limit  int
}

func NewLimitSource[T any](source Source[T], limit int) *LimitSource[T] {
	return &LimitSource[T]{source: source, limit: limit}
}

func (s *LimitSource[T]) IsEnd(n ...int) bool {
	offset := optionalOffset(n)
	return s == nil || offset >= s.limit || s.source.IsEnd(offset)
}

func (s *LimitSource[T]) Peek(n ...int) T {
	offset := optionalOffset(n)
	return s.source.Peek(offset)
}

func optionalOffset(offsets []int) int {
	if len(offsets) > 0 {
		return offsets[0]
	}
	return 0
}
