package kustolang

// Ported from src/Kusto.Language/Parser/ScriptFacts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// ScriptFactsAPI mirrors the upstream static ScriptFacts class.
// Positions and ranges are UTF-16 code units, matching the .NET API.
type ScriptFactsAPI struct{}

// GetKustoBlockStarts produces the block starts for a Kusto multi-block
// document.
func (api ScriptFactsAPI) GetKustoBlockStarts(text string, lineStarts []int) []int {
	starts := api.GetLineSeparatedBlockStarts(text, lineStarts)
	starts = api.RemoveDuplicateBlockStarts(starts)
	return api.RemoveInvalidKustoBlockStarts(starts, text)
}

// GetLineSeparatedBlockStarts separates blocks at blank lines. Invalid UTF-16
// line starts produce the single initial block instead of panicking.
func (ScriptFactsAPI) GetLineSeparatedBlockStarts(text string, lineStarts []int) []int {
	byteStarts := make([]int, len(lineStarts))
	for i, start := range lineStarts {
		byteStart, ok := parserFactsUTF16ToByteOffset(text, start)
		if !ok {
			return []int{0}
		}
		byteStarts[i] = byteStart
	}

	byteBlocks := GetLineSeparatedBlockStarts(text, byteStarts)
	blocks := make([]int, 0, len(byteBlocks))
	for _, start := range byteBlocks {
		utf16Start, ok := parserFactsByteToUTF16Offset(text, start)
		if !ok {
			continue
		}
		blocks = append(blocks, utf16Start)
	}
	if len(blocks) == 0 {
		return []int{0}
	}
	return blocks
}

// RemoveDuplicateBlockStarts returns sorted, unique block starts without
// retaining or modifying the caller's slice.
func (ScriptFactsAPI) RemoveDuplicateBlockStarts(blockStarts []int) []int {
	starts := RemoveDuplicateBlockStarts(blockStarts)
	if len(starts) == 0 {
		return []int{}
	}
	return starts
}

// RemoveInvalidKustoBlockStarts removes starts that occur strictly inside a
// string or comment range.
func (api ScriptFactsAPI) RemoveInvalidKustoBlockStarts(blockStarts []int, text string) []int {
	return api.RemoveInvalidBlockStarts(blockStarts, api.GetInvalidKustoBlockStartRanges(text))
}

// RemoveInvalidBlockStarts removes starts that occur strictly inside one of
// the ordered invalid ranges.
func (ScriptFactsAPI) RemoveInvalidBlockStarts(blockStarts []int, invalidRanges []TextRange) []int {
	starts := RemoveInvalidBlockStarts(blockStarts, invalidRanges)
	if len(starts) == 0 {
		return []int{}
	}
	return starts
}

// GetInvalidKustoBlockStartRanges returns the ordered string and comment
// ranges in which a Kusto block cannot start.
func (ScriptFactsAPI) GetInvalidKustoBlockStartRanges(text string) []TextRange {
	byteRanges := GetInvalidKustoBlockStartRanges(text)
	if len(byteRanges) == 0 {
		return []TextRange{}
	}

	ranges := make([]TextRange, 0, len(byteRanges))
	for _, byteRange := range byteRanges {
		start, startOK := parserFactsByteToUTF16Offset(text, byteRange.Start)
		end, endOK := parserFactsByteToUTF16Offset(text, byteRange.End())
		if !startOK || !endOK || end < start {
			continue
		}
		ranges = append(ranges, NewTextRange(start, end-start))
	}
	return ranges
}

// ScriptFacts exposes source-shaped access to the upstream static API.
var ScriptFacts ScriptFactsAPI
