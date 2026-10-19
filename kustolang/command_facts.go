package kustolang

// Ported from src/Kusto.Language/Parser/CommandFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// CommandFactsAPI mirrors the upstream static CommandFacts class. Positions
// are UTF-16 code-unit offsets, matching System.String.
type CommandFactsAPI struct{}

// IsCommandStartLine reports whether the line beginning at lineStart is a
// Kusto command line.
func (CommandFactsAPI) IsCommandStartLine(text string, lineStart int) bool {
	units := UTF16CodeUnits(text)
	if lineStart < 0 || lineStart >= len(units) {
		return false
	}
	lineLength := textFactsLineLength(units, lineStart, false)
	indentationLength := textFactsWhitespaceCount(units, lineStart)
	return lineLength > indentationLength && units[lineStart+indentationLength] == '.'
}

// GetCommandBlockStarts returns command block starts. An optional line-start
// list corresponds to the upstream overload accepting IReadOnlyList<int>.
func (api CommandFactsAPI) GetCommandBlockStarts(text string, lineStarts ...[]int) []int {
	starts := TextFacts.GetLineStarts(text)
	if len(lineStarts) > 0 && lineStarts[0] != nil {
		starts = append([]int(nil), lineStarts[0]...)
	}

	commandStarts := make([]int, 0)
	commentStart := -1
	for _, lineStart := range starts {
		if api.IsCommandStartLine(text, lineStart) {
			start := lineStart
			if commentStart != -1 {
				start = commentStart
			}
			if len(commandStarts) == 0 {
				start = 0
			}
			commandStarts = append(commandStarts, start)
			commentStart = -1
			continue
		}
		if commandFactsIsCommentLine(text, lineStart) {
			if commentStart == -1 {
				commentStart = lineStart
			}
		} else {
			commentStart = -1
		}
	}
	if len(commandStarts) == 0 {
		return []int{0}
	}
	return commandStarts
}

// GetCommandBlockTexts returns each command block, retaining comments and
// whitespace associated with the block.
func (api CommandFactsAPI) GetCommandBlockTexts(text string) []string {
	units := UTF16CodeUnits(text)
	starts := api.GetCommandBlockStarts(text, TextFacts.GetLineStarts(text))
	starts = ScriptFacts.RemoveDuplicateBlockStarts(starts)
	starts = ScriptFacts.RemoveInvalidKustoBlockStarts(starts, text)

	blocks := make([]string, 0, len(starts))
	for i, start := range starts {
		end := len(units)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		if start < 0 || start > end || end > len(units) {
			continue
		}
		blocks = append(blocks, textFactsSubstring(units, start, end-start))
	}
	return blocks
}

// GetCommandTexts returns command blocks with leading and trailing comment
// and whitespace lines removed.
func (api CommandFactsAPI) GetCommandTexts(text string) []string {
	commands := make([]string, 0)
	for _, block := range api.GetCommandBlockTexts(text) {
		trimmed := CommentFacts.TrimCommentAndWhitespaceLines(block)
		if api.IsCommandStartLine(trimmed, 0) {
			commands = append(commands, trimmed)
		}
	}
	return commands
}

func commandFactsIsCommentLine(text string, lineStart int) bool {
	units := UTF16CodeUnits(text)
	if lineStart < 0 || lineStart >= len(units) {
		return false
	}
	lineLength := textFactsLineLength(units, lineStart, false)
	indentationLength := textFactsWhitespaceCount(units, lineStart)
	commentStart := lineStart + indentationLength
	return indentationLength < lineLength && commentStart+1 < len(units) && units[commentStart] == '/' && units[commentStart+1] == '/'
}

// CommandFacts exposes source-shaped access to the upstream static API.
var CommandFacts CommandFactsAPI
