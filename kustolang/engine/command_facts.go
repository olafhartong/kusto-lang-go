package kustolang

// Ported from src/Kusto.Language/Parser/CommandFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func IsCommandStartLine(text string, lineStart int) bool {
	lineLength := LineLength(text, lineStart)
	indentationLength := IndentationLength(text, lineStart)
	return lineLength > indentationLength && lineStart+indentationLength < len(text) && text[lineStart+indentationLength] == '.'
}

func CommandBlockStarts(text string, lineStarts ...[]int) []int {
	starts := LineStarts(text)
	if len(lineStarts) > 0 && lineStarts[0] != nil {
		starts = lineStarts[0]
	}
	commandStarts := []int{}
	commentStart := -1
	for _, lineStart := range starts {
		if IsCommandStartLine(text, lineStart) {
			start := lineStart
			if commentStart != -1 {
				start = commentStart
			}
			if len(commandStarts) == 0 {
				start = 0
			}
			commandStarts = append(commandStarts, start)
			commentStart = -1
		} else if IsCommentLine(text, lineStart) {
			if commentStart == -1 {
				commentStart = lineStart
			}
		} else {
			commentStart = -1
		}
	}
	if len(commandStarts) == 0 {
		commandStarts = append(commandStarts, 0)
	}
	return commandStarts
}

func CommandBlockTexts(text string) []string {
	lineStarts := LineStarts(text)
	blockStarts := CommandBlockStarts(text, lineStarts)
	blockStarts = RemoveDuplicateBlockStarts(blockStarts)
	blockStarts = RemoveInvalidKustoBlockStarts(blockStarts, text)
	blockTexts := make([]string, 0, len(blockStarts))
	for i, start := range blockStarts {
		end := len(text)
		if i < len(blockStarts)-1 {
			end = blockStarts[i+1]
		}
		blockTexts = append(blockTexts, text[start:end])
	}
	return blockTexts
}

func CommandTexts(text string) []string {
	var commandTexts []string
	for _, blockText := range CommandBlockTexts(text) {
		trimmed := TrimCommentAndWhitespaceLines(blockText)
		if IsCommandStartLine(trimmed, 0) {
			commandTexts = append(commandTexts, trimmed)
		}
	}
	return commandTexts
}
