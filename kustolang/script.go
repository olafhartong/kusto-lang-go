package kustolang

// Ported from src/Kusto.Language/Editor/CodeBlock.cs and CodeScript.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
	api "github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

type BlockSeparator func(text string, lineStarts []int) []int

// BlockSeparatorHandle gives an otherwise non-comparable Go function a stable
// identity for source-compatible CodeScript.WithSeparator no-op behavior.
type BlockSeparatorHandle struct {
	separator BlockSeparator
}

func NewBlockSeparatorHandle(separator BlockSeparator) *BlockSeparatorHandle {
	return &BlockSeparatorHandle{separator: separator}
}

// CodeScript splits a document into independently analyzed KQL blocks. Public
// positions use UTF-16 code units, matching the upstream editor API.
type CodeScript struct {
	script           *internal.CodeScript
	separator        BlockSeparator
	separatorHandle  *BlockSeparatorHandle
	separatorAdapter bool
}

// CodeBlock is one independently analyzed region of a CodeScript.
type CodeBlock struct {
	block *internal.CodeBlock
}

func NewCodeScript(text string, globals ...*GlobalState) *CodeScript {
	state := DefaultGlobalState
	if len(globals) > 0 && globals[0] != nil {
		state = globals[0]
	}
	return &CodeScript{script: internal.NewCodeScript(text, internal.NewKustoCodeServiceFactory(state))}
}

func NewCodeScriptWithFactory(text string, factory *CodeServiceFactory) *CodeScript {
	if factory == nil || factory.factory == nil {
		return &CodeScript{script: internal.NewCodeScript(text, internal.NewUnknownCodeServiceFactory())}
	}
	return &CodeScript{script: internal.NewCodeScript(text, factory.factory)}
}

func NewCodeBlock(start int, service *CodeService) *CodeBlock {
	var inner internal.CodeService
	if service != nil {
		inner = service.service
	}
	return wrapCodeBlock(internal.NewCodeBlock(start, inner))
}

func NewCodeScriptForEnvironment(text string, env *api.EnvironmentDefinition, local api.LocalData) (*CodeScript, error) {
	globals, err := NewGlobalStateForEnvironment(env, local)
	if err != nil {
		return nil, err
	}
	return NewCodeScript(text, globals), nil
}

func (s *CodeScript) Text() string {
	if s == nil || s.script == nil {
		return ""
	}
	return s.script.Text
}

func (s *CodeScript) Blocks() []*CodeBlock {
	if s == nil || s.script == nil {
		return []*CodeBlock{}
	}
	blocks := s.script.Blocks()
	result := make([]*CodeBlock, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, wrapCodeBlock(block))
	}
	return result
}

func (s *CodeScript) LineStarts() []int {
	if s == nil || s.script == nil {
		return []int{}
	}
	return s.script.LineStarts()
}

func (s *CodeScript) Factory() *CodeServiceFactory {
	if s == nil || s.script == nil {
		return nil
	}
	return wrapCodeServiceFactory(s.script.Factory)
}

func (s *CodeScript) Globals() *GlobalState {
	if s == nil || s.script == nil {
		return nil
	}
	factory, ok := s.script.Factory.GetFactory(func(factory internal.CodeServiceFactory) bool {
		_, isKusto := factory.(*internal.KustoCodeServiceFactory)
		return isKusto
	}).(*internal.KustoCodeServiceFactory)
	if !ok || factory == nil {
		return nil
	}
	return factory.Globals
}

func (s *CodeScript) Separator() BlockSeparator {
	if s == nil || s.script == nil {
		return nil
	}
	return s.separator
}

func (s *CodeScript) SeparatorHandle() *BlockSeparatorHandle {
	if s == nil || s.script == nil {
		return nil
	}
	return s.separatorHandle
}

func (s *CodeScript) WithText(text string) *CodeScript {
	if s == nil || s.script == nil {
		return NewCodeScript(text)
	}
	return &CodeScript{
		script:           s.script.WithText(text),
		separator:        s.separator,
		separatorHandle:  s.separatorHandle,
		separatorAdapter: s.separatorAdapter,
	}
}

func (s *CodeScript) WithFactory(factory *CodeServiceFactory) *CodeScript {
	if s == nil || s.script == nil {
		return NewCodeScriptWithFactory("", factory)
	}
	if factory == nil || factory.factory == nil {
		updated := s.script.WithFactory(internal.NewUnknownCodeServiceFactory())
		if updated == s.script {
			return s
		}
		return s.withScript(updated)
	}
	updated := s.script.WithFactory(factory.factory)
	if updated == s.script {
		return s
	}
	return s.withScript(updated)
}

func (s *CodeScript) WithTextAndFactory(text string, factory *CodeServiceFactory) *CodeScript {
	if s == nil || s.script == nil {
		return NewCodeScriptWithFactory(text, factory)
	}
	if factory == nil || factory.factory == nil {
		updated := s.script.WithTextAndFactory(text, internal.NewUnknownCodeServiceFactory())
		if updated == s.script {
			return s
		}
		return s.withScript(updated)
	}
	updated := s.script.WithTextAndFactory(text, factory.factory)
	if updated == s.script {
		return s
	}
	return s.withScript(updated)
}

func (s *CodeScript) WithGlobals(globals *GlobalState) *CodeScript {
	if s == nil || s.script == nil {
		return s
	}
	if globals == nil {
		globals = DefaultGlobalState
	}
	kustoFactory, ok := s.script.Factory.GetFactory(func(factory internal.CodeServiceFactory) bool {
		_, isKusto := factory.(*internal.KustoCodeServiceFactory)
		return isKusto
	}).(*internal.KustoCodeServiceFactory)
	if !ok || kustoFactory == nil {
		return s
	}

	factory := s.script.Factory.WithFactory(kustoFactory.WithGlobals(globals))
	updated := s.script.WithFactory(factory)
	if updated == s.script {
		return s
	}
	return s.withScript(updated)
}

func (s *CodeScript) WithSeparator(separator BlockSeparator) *CodeScript {
	return s.withSeparator(separator, nil)
}

func (s *CodeScript) WithSeparatorHandle(handle *BlockSeparatorHandle) *CodeScript {
	if s != nil && handle != nil && handle == s.separatorHandle {
		return s
	}
	var separator BlockSeparator
	if handle != nil {
		separator = handle.separator
	}
	return s.withSeparator(separator, handle)
}

func (s *CodeScript) withSeparator(separator BlockSeparator, handle *BlockSeparatorHandle) *CodeScript {
	if s == nil || s.script == nil {
		return NewCodeScript("").withSeparator(separator, handle)
	}

	adapter := s.separatorAdapter
	var converted internal.BlockSeparator
	if separator != nil {
		adapter = !adapter
		converted = adaptBlockSeparator(separator, adapter)
	}
	updated := s.script.WithSeparator(converted)
	if updated == s.script {
		return s
	}
	return &CodeScript{
		script:           updated,
		separator:        separator,
		separatorHandle:  handle,
		separatorAdapter: adapter,
	}
}

func (s *CodeScript) withScript(script *internal.CodeScript) *CodeScript {
	return &CodeScript{
		script:           script,
		separator:        s.separator,
		separatorHandle:  s.separatorHandle,
		separatorAdapter: s.separatorAdapter,
	}
}

func adaptBlockSeparator(separator BlockSeparator, alternate bool) internal.BlockSeparator {
	if alternate {
		return func(text string, lineStarts []int) []int {
			return separator(text, lineStarts)
		}
	}
	return func(text string, lineStarts []int) []int {
		return separator(text, lineStarts)
	}
}

func (s *CodeScript) GetBlockAtPosition(position int) *CodeBlock {
	if s == nil || s.script == nil {
		return nil
	}
	return wrapCodeBlock(s.script.GetBlockAtPosition(position))
}

func (s *CodeScript) TryGetTextPosition(line, lineOffset int) (position int, ok bool) {
	if s == nil || s.script == nil {
		return 0, false
	}
	return s.script.TryGetTextPosition(line, lineOffset)
}

func (s *CodeScript) TryGetLineAndOffset(position int) (line, lineOffset int, ok bool) {
	if s == nil || s.script == nil {
		return 0, 0, false
	}
	return s.script.TryGetLineAndOffset(position)
}

func wrapCodeBlock(block *internal.CodeBlock) *CodeBlock {
	if block == nil {
		return nil
	}
	return &CodeBlock{block: block}
}

func (b *CodeBlock) Start() int {
	if b == nil || b.block == nil {
		return 0
	}
	return b.block.Start
}

func (b *CodeBlock) Text() string {
	if b == nil || b.block == nil {
		return ""
	}
	return b.block.Text()
}

func (b *CodeBlock) Length() int {
	if b == nil || b.block == nil {
		return 0
	}
	return b.block.Length()
}

func (b *CodeBlock) End() int {
	if b == nil || b.block == nil {
		return 0
	}
	return b.block.End()
}

func (b *CodeBlock) Kind() string {
	if b == nil || b.block == nil {
		return ""
	}
	return b.block.Kind()
}

func (b *CodeBlock) Service() *CodeService {
	if b == nil || b.block == nil {
		return nil
	}
	return wrapCodeService(b.block.Service())
}
