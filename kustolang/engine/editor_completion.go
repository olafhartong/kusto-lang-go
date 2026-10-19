package kustolang

import (
	"fmt"
	"strings"
)

// Ported from:
//   src/Kusto.Language/Editor/BrackettingStyle.cs
//   src/Kusto.Language/Editor/PlacementStyle.cs
//   src/Kusto.Language/Editor/SpacingStyle.cs
//   src/Kusto.Language/Editor/PositionBias.cs
//   src/Kusto.Language/Editor/MinimalTextKind.cs
//   src/Kusto.Language/Editor/CompletionKind.cs
//   src/Kusto.Language/Editor/CompletionPriority.cs
//   src/Kusto.Language/Editor/CompletionDisplay.cs
//   src/Kusto.Language/Editor/CompletionText.cs
//   src/Kusto.Language/Editor/CompletionItem.cs
//   src/Kusto.Language/Editor/CompletionInfo.cs
//   src/Kusto.Language/Editor/CompletionOptions.cs
//   src/Kusto.Language/Editor/TextEdit.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type BrackettingStyle int

const (
	BrackettingStyleDefault BrackettingStyle = iota
	BrackettingStyleNone
	BrackettingStyleVertical
	BrackettingStyleDiagonal
)

type PlacementStyle int

const (
	PlacementStyleNone PlacementStyle = iota
	PlacementStyleNewLine
	PlacementStyleSmart
)

type SpacingStyle int

const (
	SpacingStyleAsIs SpacingStyle = iota
	SpacingStyleMinimal
	SpacingStyleOne
)

type DualSpacingStyle int

const (
	DualSpacingStyleAsIs DualSpacingStyle = iota
	DualSpacingStyleNeither
	DualSpacingStyleBefore
	DualSpacingStyleAfter
	DualSpacingStyleBoth
)

type PositionBias int

const (
	PositionBiasLeft PositionBias = iota
	PositionBiasRight
)

type MinimalTextKind int

const (
	MinimalTextKindRemoveLeadingWhitespaceAndComments MinimalTextKind = iota
	MinimalTextKindMinimizeWhitespaceAndRemoveComments
	MinimalTextKindSingleLine
)

type CompletionKind int

const (
	CompletionKindUnknown CompletionKind = iota
	CompletionKindKeyword
	CompletionKindPunctuation
	CompletionKindSyntax
	CompletionKindIdentifier
	CompletionKindExample
	CompletionKindScalarPrefix
	CompletionKindTabularPrefix
	CompletionKindTabularSuffix
	CompletionKindQueryPrefix
	CompletionKindCommandPrefix
	CompletionKindScalarInfix
	CompletionKindRenderChart
	CompletionKindColumn
	CompletionKindTable
	CompletionKindBuiltInFunction
	CompletionKindLocalFunction
	CompletionKindDatabaseFunction
	CompletionKindAggregateFunction
	CompletionKindParameter
	CompletionKindVariable
	CompletionKindDatabase
	CompletionKindCluster
	CompletionKindMaterialiedView
	CompletionKindEntityGroup
	CompletionKindGraph
	CompletionKindScalarType
	CompletionKindOption
	CompletionKindStoredQueryResult
)

type CompletionRank int

const (
	CompletionRankLiteral        CompletionRank = 1
	CompletionRankAggregate      CompletionRank = 2
	CompletionRankColumn         CompletionRank = 3
	CompletionRankTable          CompletionRank = 4
	CompletionRankEntity         CompletionRank = 5
	CompletionRankVariable       CompletionRank = 6
	CompletionRankFunction       CompletionRank = 7
	CompletionRankKeyword        CompletionRank = 8
	CompletionRankStringOperator CompletionRank = 9
	CompletionRankMathOperator   CompletionRank = 10
	CompletionRankOther          CompletionRank = 11
	CompletionRankDefault        CompletionRank = 12
)

type CompletionPriority int

const (
	CompletionPriorityTop     CompletionPriority = 1
	CompletionPriorityHigh    CompletionPriority = 2
	CompletionPriorityNormal  CompletionPriority = 3
	CompletionPriorityLow     CompletionPriority = 4
	CompletionPriorityDefault                    = CompletionPriorityNormal
)

type CompletionText struct {
	Text   string
	Caret  bool
	Select bool
}

func NewCompletionText(text string, caret bool, selectText bool) CompletionText {
	return CompletionText{Text: text, Caret: caret, Select: selectText}
}

func NewCompletionTextEdit(text string) CompletionText {
	return NewCompletionText(text, true, true)
}

const (
	defaultCompletionCursorMarker = "|"
	defaultCompletionStartMarker  = "[["
	defaultCompletionEndMarker    = "]]"
)

func ParseCompletionText(textWithMarkers string, markers ...string) []CompletionText {
	cursorMarker := defaultCompletionCursorMarker
	startMarker := defaultCompletionStartMarker
	endMarker := defaultCompletionEndMarker
	if len(markers) > 0 {
		cursorMarker = markers[0]
	}
	if len(markers) > 1 {
		startMarker = markers[1]
	}
	if len(markers) > 2 {
		endMarker = markers[2]
	}

	var list []CompletionText
	if cursorPos := strings.Index(textWithMarkers, cursorMarker); cursorPos >= 0 {
		beforeCursorText := ""
		if cursorPos > 0 {
			beforeCursorText = textWithMarkers[:cursorPos]
		}
		afterCursorText := textWithMarkers[cursorPos+len(cursorMarker):]
		list = append(list, NewCompletionText(beforeCursorText, false, false))
		list = append(list, NewCompletionText(afterCursorText, true, false))
		return ListToReadOnly(list)
	}

	pos := 0
	for pos < len(textWithMarkers) {
		startPos := strings.Index(textWithMarkers[pos:], startMarker)
		if startPos >= 0 {
			startPos += pos
			endPos := strings.Index(textWithMarkers[startPos+len(startMarker):], endMarker)
			if endPos >= 0 {
				endPos += startPos + len(startMarker)
				if startPos > pos {
					list = append(list, NewCompletionText(textWithMarkers[pos:startPos], false, false))
				}
				list = append(list, NewCompletionTextEdit(textWithMarkers[startPos+len(startMarker):endPos]))
				pos = endPos + len(endMarker)
				continue
			}
		}
		list = append(list, NewCompletionText(textWithMarkers[pos:], false, false))
		pos = len(textWithMarkers)
	}
	return ListToReadOnly(list)
}

type CompletionItem struct {
	Kind        CompletionKind
	DisplayText string
	MatchText   string
	OrderText   string
	ApplyTexts  []CompletionText
	Retrigger   bool

	rank     CompletionRank
	priority CompletionPriority
}

func (item CompletionItem) Rank() CompletionRank {
	return item.rank
}

func (item CompletionItem) Priority() CompletionPriority {
	return item.priority
}

func (item CompletionItem) ApplyTextsList() []CompletionText {
	return ListToReadOnly(item.ApplyTexts)
}

type CompletionItemOption func(*completionItemOptions)

type completionItemOptions struct {
	beforeText *string
	afterText  *string
	matchText  *string
	orderText  *string
	applyTexts []CompletionText
	rank       CompletionRank
	priority   CompletionPriority
	retrigger  bool
}

func defaultCompletionItemOptions() completionItemOptions {
	return completionItemOptions{
		rank:     CompletionRankDefault,
		priority: CompletionPriorityNormal,
	}
}

func WithCompletionBeforeText(text string) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.beforeText = &text
	}
}

func WithCompletionAfterText(text string) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.afterText = &text
	}
}

func WithCompletionMatchText(text string) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.matchText = &text
	}
}

func WithCompletionOrderText(text string) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.orderText = &text
	}
}

func WithCompletionApplyTexts(texts ...CompletionText) CompletionItemOption {
	ownedTexts := ListToReadOnly(texts)
	return func(options *completionItemOptions) {
		options.applyTexts = ownedTexts
	}
}

func WithCompletionApplyTextMarkers(textWithMarkers string) CompletionItemOption {
	return WithCompletionApplyTexts(ParseCompletionText(textWithMarkers)...)
}

func WithCompletionRank(rank CompletionRank) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.rank = rank
	}
}

func WithCompletionPriority(priority CompletionPriority) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.priority = priority
	}
}

func WithCompletionRetrigger(retrigger bool) CompletionItemOption {
	return func(options *completionItemOptions) {
		options.retrigger = retrigger
	}
}

func NewCompletionItem(kind CompletionKind, displayText string, options ...CompletionItemOption) CompletionItem {
	config := defaultCompletionItemOptions()
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}

	applyTexts := config.applyTexts
	if applyTexts == nil {
		beforeText := displayText
		if config.beforeText != nil {
			beforeText = *config.beforeText
		}
		applyTexts = createCompletionApplyTexts(beforeText, config.afterText)
	}

	matchText := displayText
	if config.matchText != nil {
		matchText = *config.matchText
	}
	orderText := matchText
	if config.orderText != nil {
		orderText = *config.orderText
	}

	return completionItemWithOwnedApplyTexts(CompletionItem{
		Kind:        kind,
		DisplayText: displayText,
		MatchText:   matchText,
		OrderText:   orderText,
		ApplyTexts:  applyTexts,
		Retrigger:   config.retrigger,
		rank:        config.rank,
		priority:    config.priority,
	})
}

func NewSyntaxCompletionItem(displayText string, options ...CompletionItemOption) CompletionItem {
	return NewCompletionItem(CompletionKindSyntax, displayText, options...)
}

func createCompletionApplyTexts(beforeText string, afterText *string) []CompletionText {
	list := []CompletionText{NewCompletionText(beforeText, false, false)}
	if afterText != nil {
		list = append(list, NewCompletionText(*afterText, true, false))
	}
	return ListToReadOnly(list)
}

func (item CompletionItem) BeforeText() string {
	if len(item.ApplyTexts) == 0 {
		return ""
	}
	caretIndex := item.CaretIndex()
	if caretIndex > 0 {
		return joinCompletionTexts(item.ApplyTexts[:caretIndex])
	}
	if caretIndex < 0 {
		return joinCompletionTexts(item.ApplyTexts)
	}
	return ""
}

func (item CompletionItem) AfterText() string {
	if len(item.ApplyTexts) == 0 {
		return ""
	}
	caretIndex := item.CaretIndex()
	if caretIndex >= 0 {
		return joinCompletionTexts(item.ApplyTexts[caretIndex:])
	}
	return ""
}

func (item CompletionItem) EditText() string {
	return item.BeforeText()
}

func (item CompletionItem) HasCaret() bool {
	return item.CaretIndex() >= 0
}

func (item CompletionItem) CaretIndex() int {
	for i, text := range item.ApplyTexts {
		if text.Caret {
			return i
		}
	}
	return -1
}

func (item CompletionItem) WithKind(kind CompletionKind) CompletionItem {
	item.Kind = kind
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithDisplayText(displayText string) CompletionItem {
	item.DisplayText = displayText
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithMatchText(matchText string) CompletionItem {
	item.MatchText = matchText
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithOrderText(orderText string) CompletionItem {
	item.OrderText = orderText
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithApplyTexts(applyTexts ...CompletionText) CompletionItem {
	item.ApplyTexts = applyTexts
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithApplyTextMarkers(textWithMarkers string) CompletionItem {
	return item.WithApplyTexts(ParseCompletionText(textWithMarkers)...)
}

func (item CompletionItem) WithBeforeText(beforeText string) CompletionItem {
	newBeforeText := NewCompletionText(beforeText, false, false)
	caretIndex := item.CaretIndex()
	if caretIndex >= 0 {
		item.ApplyTexts = ListReplaceRangeItem(item.ApplyTexts, 0, caretIndex, newBeforeText)
	} else {
		item.ApplyTexts = []CompletionText{newBeforeText}
	}
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithEditText(editText string) CompletionItem {
	return item.WithBeforeText(editText)
}

func (item CompletionItem) WithAfterText(afterText string) CompletionItem {
	newAfterText := NewCompletionText(afterText, true, false)
	caretIndex := item.CaretIndex()
	switch {
	case caretIndex > 0:
		item.ApplyTexts = ListReplaceRangeItem(item.ApplyTexts, caretIndex, len(item.ApplyTexts)-caretIndex, newAfterText)
	case caretIndex == 0:
		item.ApplyTexts = []CompletionText{newAfterText}
	default:
		item.ApplyTexts = ListAppendItem(item.ApplyTexts, newAfterText)
	}
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithRank(rank CompletionRank) CompletionItem {
	item.rank = rank
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithPriority(priority CompletionPriority) CompletionItem {
	item.priority = priority
	return completionItemWithOwnedApplyTexts(item)
}

func (item CompletionItem) WithRetrigger(retrigger bool) CompletionItem {
	item.Retrigger = retrigger
	return completionItemWithOwnedApplyTexts(item)
}

func completionItemWithOwnedApplyTexts(item CompletionItem) CompletionItem {
	item.ApplyTexts = ListToReadOnly(item.ApplyTexts)
	return item
}

func completionItemsToReadOnly(items []CompletionItem) []CompletionItem {
	if len(items) == 0 {
		return EmptyReadOnlyList[CompletionItem]()
	}
	ownedItems := make([]CompletionItem, len(items))
	for i, item := range items {
		ownedItems[i] = completionItemWithOwnedApplyTexts(item)
	}
	return ownedItems
}

func joinCompletionTexts(texts []CompletionText) string {
	var builder strings.Builder
	for _, text := range texts {
		builder.WriteString(text.Text)
	}
	return builder.String()
}

type CompletionInfo struct {
	Items      []CompletionItem
	EditStart  int
	EditLength int
}

func NewCompletionInfo(items []CompletionItem, editStart int, editLength int) CompletionInfo {
	return CompletionInfo{
		Items:      completionItemsToReadOnly(items),
		EditStart:  editStart,
		EditLength: editLength,
	}
}

func (info CompletionInfo) ItemsList() []CompletionItem {
	return completionItemsToReadOnly(info.Items)
}

var CompletionInfoEmpty = NewCompletionInfo(nil, 0, 0)

type completionOptionFlags uint32

const (
	completionOptionIncludeSymbols completionOptionFlags = 1 << iota
	completionOptionIncludeSyntax
	completionOptionIncludePunctuation
	_
	completionOptionEnableParameterInjection
	completionOptionAutoAppendWhitespace
	completionOptionIncludeExtendedSyntax

	completionOptionDefault = completionOptionIncludeSymbols |
		completionOptionIncludeSyntax |
		completionOptionIncludePunctuation |
		completionOptionAutoAppendWhitespace
)

type CompletionOptions struct {
	flags            completionOptionFlags
	IncludeFunctions IncludeFunctionKind
}

func NewCompletionOptions(flags completionOptionFlags, includeFunctionKind IncludeFunctionKind) CompletionOptions {
	return CompletionOptions{flags: flags, IncludeFunctions: includeFunctionKind}
}

func (options CompletionOptions) IncludeSymbols() bool {
	return options.flags&completionOptionIncludeSymbols != 0
}

func (options CompletionOptions) IncludeSyntax() bool {
	return options.flags&completionOptionIncludeSyntax != 0
}

func (options CompletionOptions) IncludePunctuationOnlySyntax() bool {
	return options.flags&completionOptionIncludePunctuation != 0
}

func (options CompletionOptions) EnableParameterInjection() bool {
	return options.flags&completionOptionEnableParameterInjection != 0
}

func (options CompletionOptions) AutoAppendWhitespace() bool {
	return options.flags&completionOptionAutoAppendWhitespace != 0
}

func (options CompletionOptions) IncludeExtendedSyntax() bool {
	return options.flags&completionOptionIncludeExtendedSyntax != 0
}

func (options CompletionOptions) WithIncludeSymbols(include bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionIncludeSymbols, include)
	return options
}

func (options CompletionOptions) WithIncludeFunctions(kind IncludeFunctionKind) CompletionOptions {
	options.IncludeFunctions = kind
	return options
}

func (options CompletionOptions) WithIncludeFunctionsEnabled(include bool) CompletionOptions {
	if include {
		options.IncludeFunctions = IncludeFunctionKindAll
	} else {
		options.IncludeFunctions = IncludeFunctionKindNone
	}
	return options
}

func (options CompletionOptions) WithIncludeSyntax(include bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionIncludeSyntax, include)
	return options
}

func (options CompletionOptions) WithIncludePunctuationOnlySyntax(include bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionIncludePunctuation, include)
	return options
}

func (options CompletionOptions) WithIncludeExtendedSyntax(include bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionIncludeExtendedSyntax, include)
	return options
}

func (options CompletionOptions) WithEnableParameterInjection(enable bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionEnableParameterInjection, enable)
	return options
}

func (options CompletionOptions) WithAutoAppendWhitespace(enable bool) CompletionOptions {
	options.flags = setCompletionOption(options.flags, completionOptionAutoAppendWhitespace, enable)
	return options
}

func setCompletionOption(flags completionOptionFlags, option completionOptionFlags, value bool) completionOptionFlags {
	if value {
		return flags | option
	}
	return flags &^ option
}

var CompletionOptionsDefault = NewCompletionOptions(completionOptionDefault, IncludeFunctionKindAll)

type TextEdit struct {
	Start        int
	DeleteLength int
	InsertText   string
}

type ArgumentOutOfRangeError struct {
	ParameterName string
	Message       string
}

func (e ArgumentOutOfRangeError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ParameterName == "" {
		return "argument out of range"
	}
	return fmt.Sprintf("argument %q out of range", e.ParameterName)
}

func NewTextEdit(start int, deleteLength int, insertText string) TextEdit {
	if start < 0 {
		panic(ArgumentOutOfRangeError{ParameterName: "start", Message: "negate start position"})
	}
	if deleteLength < 0 {
		panic(ArgumentOutOfRangeError{ParameterName: "deleteLength", Message: "negative delete length"})
	}
	return TextEdit{Start: start, DeleteLength: deleteLength, InsertText: insertText}
}

func NewTextReplacement(start int, deleteLength int, insertText string) TextEdit {
	return NewTextEdit(start, deleteLength, insertText)
}

func NewTextDeletion(start int, deleteLength int) TextEdit {
	return NewTextEdit(start, deleteLength, "")
}

func NewTextInsertion(start int, text string) TextEdit {
	return NewTextEdit(start, 0, text)
}

func CompletionDisplayText(symbol Symbol, nameOnly ...bool) string {
	onlyName := len(nameOnly) > 0 && nameOnly[0]
	switch s := symbol.(type) {
	case nil:
		return ""
	case *TableSymbol:
		if onlyName {
			return s.SymbolName()
		}
		if s.IsExternal() {
			return fmt.Sprintf("external_table('%s')", s.SymbolName())
		}
		return s.SymbolName()
	case *ExternalTableSymbol:
		if onlyName {
			return s.SymbolName()
		}
		return fmt.Sprintf("external_table('%s')", s.SymbolName())
	case *FunctionSymbol:
		if onlyName {
			return s.SymbolName()
		}
		signatures := s.Signatures()
		if len(signatures) == 0 {
			return s.SymbolName()
		}
		return s.SymbolName() + completionSignatureText(signatures[0])
	case *PatternSymbol:
		return s.SymbolName() + completionParametersText(s.Parameters(), false)
	case *VariableSymbol:
		if fn, ok := s.Type.(*FunctionSymbol); ok {
			return CompletionDisplayText(fn, onlyName)
		}
		return s.SymbolName()
	case *ParameterSymbol:
		if fn, ok := s.Type.(*FunctionSymbol); ok {
			return CompletionDisplayText(fn, onlyName)
		}
		return s.SymbolName()
	case *DatabaseSymbol, *ClusterSymbol, *OptionSymbol:
		return s.SymbolName()
	default:
		return symbol.SymbolName()
	}
}

func completionSignatureText(signature *Signature) string {
	if signature == nil {
		return "()"
	}
	return completionParametersText(signature.Parameters, signature.HasRepeatableParameters)
}

func completionParametersText(parameters []*Parameter, hasRepeatableParameters bool) string {
	var builder strings.Builder
	for i, parameter := range parameters {
		if i > 0 {
			builder.WriteString(", ")
		}
		name := ""
		if parameter != nil {
			name = parameter.Name
		}
		if parameter != nil && parameter.IsOptional() {
			builder.WriteString("[")
			builder.WriteString(name)
			builder.WriteString("]")
		} else {
			builder.WriteString(name)
		}
	}
	if hasRepeatableParameters {
		builder.WriteString(", ...")
	}
	return "(" + builder.String() + ")"
}
