package kustolang

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoOutliner.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func KustoOutlinerGetOutlines(code *KustoCode, options OutliningOptions) OutlineInfo {
	outlines := []OutlineRange{}
	if code == nil || code.Syntax() == nil {
		return NewOutlineInfo(outlines)
	}

	if options.Queries {
		outlines = kustoOutlinerAddRootOutline(code, options, outlines)
	}

	if options.Statements {
		walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
			if statement, ok := WrapGeneratedSyntaxElement(element).(Statement); ok {
				outlines = kustoOutlinerAddStatementOutline(code, statement, triviaStart, options, outlines)
			}
		})
	}

	return NewOutlineInfo(outlines)
}

func kustoOutlinerOnSameLine(code *KustoCode, a *SyntaxToken, b *SyntaxToken) bool {
	lineA, _, okA := code.TryGetLineAndOffset(a.TextStart())
	lineB, _, okB := code.TryGetLineAndOffset(b.TextStart())
	return okA && okB && lineA == lineB
}

func kustoOutlinerAddRootOutline(code *KustoCode, options OutliningOptions, outlines []OutlineRange) []OutlineRange {
	collapsedText, ok := kustoOutlinerTryGetRootCollapsedText(code, options)
	if !ok {
		return outlines
	}
	length := TrimEnd(code.Text, 0, len(code.Text))
	return append(outlines, NewOutlineRange(0, length, collapsedText))
}

func kustoOutlinerTryGetRootCollapsedText(code *KustoCode, options OutliningOptions) (string, bool) {
	return kustoOutlinerTryGetCollapsedText(code, 0, len(code.Text), options)
}

func kustoOutlinerTryGetCollapsedText(code *KustoCode, start int, length int, options OutliningOptions) (string, bool) {
	if code == nil || code.Syntax() == nil || length < 0 {
		return "", false
	}

	var builder strings.Builder
	builderWidth := 0
	braces := false
	end := start + length
	refs := kustoServiceTokenRefs(code.Syntax())
	index := kustoOutlinerTokenIndexAtOrAfter(refs, start)

	for index >= 0 && index < len(refs) {
		token := refs[index].positionedToken()
		if token == nil || token.TextStart() >= end {
			break
		}
		switch token.RawText() {
		case "{":
			braces = true
			index = len(refs)
			continue
		case "|", ";":
			index = len(refs)
			continue
		}

		if token.TriviaWidth() > 0 && token.TriviaStart() >= start {
			if builderWidth == 0 {
				trivia := token.Trivia()
				prefix := kustoOutlinerUTF16Prefix(trivia, options.MaxCollapsedTextLength)
				builder.WriteString(prefix)
				builderWidth += utf16TextWidth(prefix)
			} else if token.RawText() != "" {
				builder.WriteString(" ")
				builderWidth++
			}
		}

		tokenWidth := utf16TextWidth(token.RawText())
		if builderWidth+tokenWidth <= options.MaxCollapsedTextLength {
			builder.WriteString(token.RawText())
			builderWidth += tokenWidth
		} else {
			break
		}
		index++
	}

	newText := builder.String()
	if IsWhitespaceOnly(newText) {
		return "", false
	}
	if braces {
		return newText + " { ... }", true
	}
	return newText + " ...", true
}

func kustoOutlinerUTF16Prefix(text string, maxWidth int) string {
	if maxWidth < 0 {
		panic("kustolang: negative collapsed text length")
	}
	width := 0
	end := 0
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		units := utf16.RuneLen(r)
		if units < 0 {
			units = 1
		}
		if width+units > maxWidth {
			break
		}
		width += units
		end += size
	}
	return text[:end]
}

func kustoOutlinerTokenIndexAtOrAfter(refs []kustoServiceTokenRef, position int) int {
	if len(refs) == 0 {
		return -1
	}
	for i, ref := range refs {
		if ref.end() > position || ref.positionedToken().TextStart() >= position {
			return i
		}
	}
	return len(refs) - 1
}

func kustoOutlinerAddStatementOutline(code *KustoCode, statement Statement, triviaStart int, options OutliningOptions, outlines []OutlineRange) []OutlineRange {
	firstToken, lastToken, ok := kustoOutlinerFirstAndLastToken(statement, triviaStart)
	if !ok || firstToken == nil || lastToken == nil || kustoOutlinerOnSameLine(code, firstToken, lastToken) {
		return outlines
	}

	length := lastToken.End() - firstToken.TextStart()
	collapsedText, ok := kustoOutlinerTryGetCollapsedText(code, firstToken.TextStart(), length, options)
	if !ok {
		return outlines
	}
	return append(outlines, NewOutlineRange(firstToken.TextStart(), length, collapsedText))
}

func kustoOutlinerFirstAndLastToken(element SyntaxElement, triviaStart int) (*SyntaxToken, *SyntaxToken, bool) {
	var refs []kustoServiceTokenRef
	collectKustoServiceTokenRefs(element, triviaStart, nil, &refs)
	if len(refs) == 0 {
		return nil, nil, false
	}
	first := refs[0].positionedToken()
	last := refs[len(refs)-1].positionedToken()
	return first, last, true
}
