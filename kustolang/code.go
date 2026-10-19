package kustolang

// Ported from src/Kusto.Language/KustoCode.cs. Public positions use UTF-16
// code units like .NET strings; the native parser remains byte-indexed inside
// the implementation package.

import (
	"unicode/utf8"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

// KustoCode is a parsed KQL document with optional semantic analysis.
type KustoCode struct {
	Text       string
	Kind       CodeKind
	Globals    *GlobalState
	Tree       *SyntaxTree
	ResultType *TypeSymbol
	Query      *QueryBlock

	code *internal.KustoCode
}

func wrapKustoCode(code *internal.KustoCode) *KustoCode {
	if code == nil {
		return nil
	}
	return &KustoCode{
		Text:       code.Text,
		Kind:       code.Kind,
		Globals:    code.Globals,
		Tree:       code.Tree,
		ResultType: code.ResultType,
		Query:      code.Query,
		code:       code,
	}
}

func wrapKustoCodeResult(receiver *KustoCode, code *internal.KustoCode) *KustoCode {
	if receiver != nil && receiver.code == code {
		return receiver
	}
	return wrapKustoCode(code)
}

func (c *KustoCode) Syntax() *SyntaxNode {
	if c == nil || c.code == nil {
		return nil
	}
	return c.code.Syntax()
}

func (c *KustoCode) GetDiagnostics() []Diagnostic {
	return c.GetDiagnosticsWithCancellation(CancellationToken{})
}

func (c *KustoCode) GetDiagnosticsWithCancellation(token CancellationToken) []Diagnostic {
	if c == nil || c.code == nil {
		return []Diagnostic{}
	}
	return diagnosticsToUTF16(c.code.Text, c.code.GetDiagnosticsWithCancellation(token))
}

func (c *KustoCode) GetSyntaxDiagnostics() []Diagnostic {
	return c.GetSyntaxDiagnosticsWithCancellation(CancellationToken{})
}

func (c *KustoCode) GetSyntaxDiagnosticsWithCancellation(token CancellationToken) []Diagnostic {
	if c == nil || c.code == nil {
		return []Diagnostic{}
	}
	return diagnosticsToUTF16(c.code.Text, c.code.GetSyntaxDiagnosticsWithCancellation(token))
}

func (c *KustoCode) HasSemantics() bool {
	return c != nil && c.code != nil && c.code.HasSemantics()
}

func (c *KustoCode) Analyze(globals ...*GlobalState) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.Analyze(globals...))
}

func (c *KustoCode) AnalyzeWithCancellation(globals *GlobalState, token CancellationToken) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.AnalyzeWithCancellation(globals, token))
}

func (c *KustoCode) WithGlobals(globals *GlobalState) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.WithGlobals(globals))
}

func (c *KustoCode) WithGlobalsWithCancellation(globals *GlobalState, token CancellationToken) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.WithGlobalsWithCancellation(globals, token))
}

func (c *KustoCode) WithText(text string) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.WithText(text))
}

func (c *KustoCode) WithTextWithCancellation(text string, token CancellationToken) *KustoCode {
	if c == nil || c.code == nil {
		return nil
	}
	return wrapKustoCodeResult(c, c.code.WithTextWithCancellation(text, token))
}

func (c *KustoCode) GetSymbolsInScope(position int, match SymbolMatch, include IncludeFunctionKind, tokens ...CancellationToken) []Symbol {
	if c == nil || c.code == nil {
		return []Symbol{}
	}
	token := scopeCancellationToken(tokens)
	return c.code.GetSymbolsInScopeWithCancellation(utf16PositionToByteOffset(c.code.Text, position), match, include, token)
}

// GetSymbolsInScopeDefault uses SymbolMatchAny and IncludeFunctionKindAll,
// matching the optional arguments on the pinned Kusto.Language API.
func (c *KustoCode) GetSymbolsInScopeDefault(position int, tokens ...CancellationToken) []Symbol {
	return c.GetSymbolsInScope(position, SymbolMatchAny, IncludeFunctionKindAll, tokens...)
}

func (c *KustoCode) GetSpeculativeReferencedSymbol(position int, name string, match SymbolMatch, tokens ...CancellationToken) Symbol {
	if c == nil || c.code == nil {
		return nil
	}
	token := scopeCancellationToken(tokens)
	return c.code.GetSpeculativeReferencedSymbolWithCancellation(utf16PositionToByteOffset(c.code.Text, position), name, match, token)
}

// GetSpeculativeReferencedSymbolDefault uses SymbolMatchAny, matching the
// optional argument on the pinned Kusto.Language API.
func (c *KustoCode) GetSpeculativeReferencedSymbolDefault(position int, name string, tokens ...CancellationToken) Symbol {
	return c.GetSpeculativeReferencedSymbol(position, name, SymbolMatchAny, tokens...)
}

func (c *KustoCode) GetColumnsInScope(position int, tokens ...CancellationToken) *TableSymbol {
	if c == nil || c.code == nil {
		return nil
	}
	token := scopeCancellationToken(tokens)
	return c.code.GetColumnsInScopeWithCancellation(utf16PositionToByteOffset(c.code.Text, position), token)
}

func (c *KustoCode) MaxDepth() int {
	if c == nil || c.code == nil {
		return 0
	}
	return c.code.MaxDepth()
}

func (c *KustoCode) Dialect() KustoDialect {
	if c == nil || c.code == nil {
		return KustoDialectUnknown
	}
	return c.code.Dialect()
}

func (c *KustoCode) TryGetLineAndOffset(position int) (line, lineOffset int, ok bool) {
	if c == nil || c.code == nil || position < 0 || position > byteOffsetToUTF16Position(c.code.Text, len(c.code.Text)) {
		return 0, 0, false
	}
	// Preserve positions between a surrogate pair's two UTF-16 code units.
	// Converting to a UTF-8 byte boundary first would collapse both positions.
	return TextFacts.TryGetLineAndOffset(c.code.Text, position)
}

func (c *KustoCode) GetTokenIndex(position int) int {
	if c == nil || c.code == nil {
		return 0
	}
	return c.code.GetTokenIndex(utf16PositionToByteOffset(c.code.Text, position))
}

func (c *KustoCode) GetLexicalTokens() []LexicalToken {
	if c == nil || c.code == nil {
		return []LexicalToken{}
	}
	return c.code.GetLexicalTokens()
}

func scopeCancellationToken(tokens []CancellationToken) CancellationToken {
	if len(tokens) == 0 {
		return CancellationToken{}
	}
	if len(tokens) == 1 {
		return tokens[0]
	}
	return internal.NewCancellationTokenFunc(func() bool {
		for _, token := range tokens {
			if token.Err() != nil {
				return true
			}
		}
		return false
	})
}

func diagnosticsToUTF16(text string, diagnostics []Diagnostic) []Diagnostic {
	return internal.DiagnosticsToUTF16(text, diagnostics)
}

func utf16PositionToByteOffset(text string, position int) int {
	if position <= 0 {
		return 0
	}
	units := 0
	for offset, r := range text {
		if units >= position {
			return offset
		}
		width := 1
		if r > 0xffff {
			width = 2
		}
		if units+width > position {
			return offset
		}
		units += width
	}
	return len(text)
}

func byteOffsetToUTF16Position(text string, offset int) int {
	if offset <= 0 {
		return 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	units := 0
	for index := 0; index < offset; {
		r, size := utf8.DecodeRuneInString(text[index:])
		if size == 0 || index+size > offset {
			break
		}
		if r > 0xffff {
			units += 2
		} else {
			units++
		}
		index += size
	}
	return units
}
