package kustolang

// Ported from the following files at upstream commit
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856:
//
//   src/Kusto.Language/Editor/OutliningOptions.cs
//   src/Kusto.Language/Editor/OutlineRange.cs
//   src/Kusto.Language/Editor/OutlineInfo.cs
//
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// OutliningOptions controls which portions of a query can be collapsed and
// how their collapsed text is displayed.
type OutliningOptions struct {
	MaxCollapsedTextLength int
	Queries                bool
	Statements             bool
}

// OutliningOptionsDefault matches Kusto.Language's default outlining options.
var OutliningOptionsDefault = OutliningOptions{
	MaxCollapsedTextLength: 120,
	Queries:                true,
	Statements:             false,
}

// WithMaxCollapesedTextlength returns a copy with the collapsed-text limit
// changed. The misspelling is retained for parity with Kusto.Language.
func (options OutliningOptions) WithMaxCollapesedTextlength(length int) OutliningOptions {
	options.MaxCollapsedTextLength = length
	return options
}

// WithMaxCollapsedTextLength is the correctly spelled compatibility helper.
func (options OutliningOptions) WithMaxCollapsedTextLength(length int) OutliningOptions {
	return options.WithMaxCollapesedTextlength(length)
}

// WithQueries returns a copy with whole-query outlining enabled or disabled.
func (options OutliningOptions) WithQueries(enable bool) OutliningOptions {
	options.Queries = enable
	return options
}

// WithStatements returns a copy with statement outlining enabled or disabled.
func (options OutliningOptions) WithStatements(enable bool) OutliningOptions {
	options.Statements = enable
	return options
}

func (options OutliningOptions) toInternal() internal.OutliningOptions {
	return internal.OutliningOptions{
		MaxCollapsedTextLength: options.MaxCollapsedTextLength,
		Queries:                options.Queries,
		Statements:             options.Statements,
	}
}

func outliningOptionsFromInternal(options internal.OutliningOptions) OutliningOptions {
	return OutliningOptions{
		MaxCollapsedTextLength: options.MaxCollapsedTextLength,
		Queries:                options.Queries,
		Statements:             options.Statements,
	}
}

// OutlineRange describes a text range that can be expanded or collapsed.
type OutlineRange struct {
	Start         int
	Length        int
	CollapsedText string
}

// NewOutlineRange creates an outline range. A Go string cannot be nil, so an
// absent collapsed text is represented by the empty string as in the C# API.
func NewOutlineRange(start, length int, collapsedText string) OutlineRange {
	return OutlineRange{Start: start, Length: length, CollapsedText: collapsedText}
}

// NewOutlineRangeOptional creates an outline range from optional collapsed
// text, treating nil as the empty string.
func NewOutlineRangeOptional(start, length int, collapsedText *string) OutlineRange {
	text := ""
	if collapsedText != nil {
		text = *collapsedText
	}
	return NewOutlineRange(start, length, text)
}

func (outline OutlineRange) toInternal() internal.OutlineRange {
	return internal.NewOutlineRange(outline.Start, outline.Length, outline.CollapsedText)
}

func outlineRangeFromInternal(outline internal.OutlineRange) OutlineRange {
	return NewOutlineRange(outline.Start, outline.Length, outline.CollapsedText)
}

// OutlineInfo contains the ranges that can be expanded or collapsed.
type OutlineInfo struct {
	Ranges []OutlineRange
}

// NewOutlineInfo creates an outline result that owns its range collection.
// A nil collection is normalized to a non-nil empty collection.
func NewOutlineInfo(ranges []OutlineRange) OutlineInfo {
	return OutlineInfo{Ranges: cloneOutlineRanges(ranges)}
}

// RangesList returns a copy of the outline ranges.
func (info OutlineInfo) RangesList() []OutlineRange {
	return cloneOutlineRanges(info.Ranges)
}

// OutlineInfoEmpty is the empty outlining result.
var OutlineInfoEmpty = NewOutlineInfo(nil)

func (info OutlineInfo) toInternal() internal.OutlineInfo {
	ranges := make([]internal.OutlineRange, len(info.Ranges))
	for i, outline := range info.Ranges {
		ranges[i] = outline.toInternal()
	}
	return internal.NewOutlineInfo(ranges)
}

func outlineInfoFromInternal(info internal.OutlineInfo) OutlineInfo {
	ranges := make([]OutlineRange, len(info.Ranges))
	for i, outline := range info.Ranges {
		ranges[i] = outlineRangeFromInternal(outline)
	}
	return NewOutlineInfo(ranges)
}

func cloneOutlineRanges(ranges []OutlineRange) []OutlineRange {
	if len(ranges) == 0 {
		return []OutlineRange{}
	}
	cloned := make([]OutlineRange, len(ranges))
	copy(cloned, ranges)
	return cloned
}
