package kustolang

// Ported from src/Kusto.Language/Parser/CommentFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// CommentFactsAPI mirrors the upstream static CommentFacts class.
// Positions are UTF-16 code-unit offsets, matching the .NET API.
type CommentFactsAPI struct{}

// GetCommentTexts gets the text of all full-line comments in text.
func (CommentFactsAPI) GetCommentTexts(text string) []string {
	comments := GetCommentTexts(text)
	if len(comments) == 0 {
		return []string{}
	}
	return append([]string(nil), comments...)
}

// IsCommentLine reports whether the line at lineStart appears to be a comment
// line. An omitted lineStart means zero. Invalid UTF-16 positions return false.
func (CommentFactsAPI) IsCommentLine(text string, lineStart ...int) bool {
	start, ok := parserFactsOptionalUTF16Position(text, lineStart)
	return ok && IsCommentLine(text, start)
}

// GetCommentLineText gets the trimmed comment text for a comment line. An
// omitted lineStart means zero. Invalid UTF-16 positions return an empty string.
func (CommentFactsAPI) GetCommentLineText(text string, lineStart ...int) string {
	start, ok := parserFactsOptionalUTF16Position(text, lineStart)
	if !ok {
		return ""
	}
	return GetCommentLineText(text, start)
}

// TrimCommentAndWhitespaceLines removes leading and trailing comment and
// whitespace-only lines.
func (CommentFactsAPI) TrimCommentAndWhitespaceLines(text string) string {
	return TrimCommentAndWhitespaceLines(text)
}

// CommentFacts exposes source-shaped access to the upstream static API.
var CommentFacts CommentFactsAPI
