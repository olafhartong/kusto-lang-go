package kustolang

import "strings"

// Ported from src/Kusto.Language/Parser/CommentFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func IsCommentLine(text string, lineStart int) bool {
	lineLength := LineLength(text, lineStart)
	indentationLength := IndentationLength(text, lineStart)
	return indentationLength < lineLength && isCommentStart(text, lineStart+indentationLength)
}

func CommentLineText(text string, lineStart int) string {
	indentation := IndentationLength(text, lineStart)
	if isCommentStart(text, lineStart+indentation) {
		lineEnd := LineEnd(text, lineStart)
		commentStart := lineStart + indentation + 2
		return strings.TrimSpace(text[commentStart:lineEnd])
	}
	return ""
}

func CommentTexts(text string) []string {
	if !strings.Contains(text, "/") {
		return nil
	}
	var comments []string
	for _, start := range LineStarts(text) {
		if IsCommentLine(text, start) {
			comments = append(comments, CommentLineText(text, start))
		}
	}
	return comments
}

func TrimCommentAndWhitespaceLines(text string) string {
	lineStarts := LineStarts(text)
	startLine := 0
	for startLine < len(lineStarts) && (IsCommentLine(text, lineStarts[startLine]) || IsWhitespaceLine(text, lineStarts[startLine])) {
		startLine++
	}

	endLine := len(lineStarts) - 1
	for endLine >= 0 && (IsCommentLine(text, lineStarts[endLine]) || IsWhitespaceLine(text, lineStarts[endLine])) {
		endLine--
	}

	if endLine < startLine {
		return ""
	}
	start := lineStarts[startLine]
	end := lineStarts[endLine] + LineLength(text, lineStarts[endLine])
	return text[start:end]
}

func StripLineComments(text string) string {
	var out strings.Builder
	for _, token := range Lex(text) {
		out.WriteString(stripCommentsFromTrivia(token.Trivia))
		out.WriteString(token.Text)
	}
	return out.String()
}

func stripCommentsFromTrivia(trivia string) string {
	if !strings.Contains(trivia, "//") {
		return trivia
	}
	var out strings.Builder
	for pos := 0; pos < len(trivia); {
		if isCommentStart(trivia, pos) {
			for pos < len(trivia) && LineBreakLength(trivia, pos) == 0 {
				pos++
			}
			continue
		}
		lineBreakLength := LineBreakLength(trivia, pos)
		if lineBreakLength > 0 {
			out.WriteString(trivia[pos : pos+lineBreakLength])
			pos += lineBreakLength
			continue
		}
		out.WriteByte(trivia[pos])
		pos++
	}
	return out.String()
}

func isCommentStart(text string, position int) bool {
	return position < len(text)-1 && text[position] == '/' && text[position+1] == '/'
}
