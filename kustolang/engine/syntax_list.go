package kustolang

import "strings"

// Ported from src/Kusto.Language/Syntax/SyntaxList.cs and SeparatedElement.cs.
// This keeps the upstream child/list shape while using Go's non-generic
// SyntaxElement interface for the incremental parser port.

type SyntaxList struct {
	syntaxOwnership
	elements     []SyntaxElement
	elementType  string
	diagnostics  []Diagnostic
	semanticInfo *SemanticInfo
}

func NewSyntaxList(elements ...SyntaxElement) *SyntaxList {
	return NewTypedSyntaxList("", elements...)
}

func NewSyntaxListWithDiagnostics(elements []SyntaxElement, diagnostics ...Diagnostic) *SyntaxList {
	return NewTypedSyntaxListWithDiagnostics("", elements, diagnostics...)
}

func NewTypedSyntaxList(elementType string, elements ...SyntaxElement) *SyntaxList {
	return NewTypedSyntaxListWithDiagnostics(elementType, elements)
}

func NewTypedSyntaxListWithDiagnostics(elementType string, elements []SyntaxElement, diagnostics ...Diagnostic) *SyntaxList {
	list := &SyntaxList{
		elements:    append([]SyntaxElement{}, elements...),
		elementType: elementType,
		diagnostics: append([]Diagnostic{}, diagnostics...),
	}
	initializeSyntaxElement(list)
	return list
}

func EmptySyntaxList() *SyntaxList {
	return NewSyntaxList()
}

func (l *SyntaxList) Kind() SyntaxKind {
	return SyntaxKindList
}

func (l *SyntaxList) ElementType() string {
	if l == nil {
		return ""
	}
	return l.elementType
}

func (l *SyntaxList) Count() int {
	if l == nil {
		return 0
	}
	return len(l.elements)
}

func (l *SyntaxList) At(index int) SyntaxElement {
	if l == nil {
		return nil
	}
	return l.elements[index]
}

func (l *SyntaxList) ChildCount() int {
	return l.Count()
}

func (l *SyntaxList) Child(index int) SyntaxElement {
	return l.At(index)
}

func (l *SyntaxList) GetChild(index int) SyntaxElement {
	return l.At(index)
}

func (l *SyntaxList) GetName(index int) string {
	return ""
}

func (l *SyntaxList) CompletionHint(index int) CompletionHint {
	return CompletionHintInherit
}

func (l *SyntaxList) IsOptional(index int) bool {
	return false
}

func (l *SyntaxList) Width() int {
	if l == nil {
		return 0
	}
	return l.FullWidth() - l.TriviaWidth()
}

func (l *SyntaxList) ByteWidth() int {
	return l.Width()
}

func (l *SyntaxList) UTF16Width() int {
	if l == nil {
		return 0
	}
	return l.UTF16FullWidth() - l.UTF16TriviaWidth()
}

func (l *SyntaxList) TriviaWidth() int {
	children := l.children()
	if len(children) == 0 || children[0] == nil {
		return 0
	}
	return children[0].TriviaWidth()
}

func (l *SyntaxList) ByteTriviaWidth() int {
	return l.TriviaWidth()
}

func (l *SyntaxList) UTF16TriviaWidth() int {
	children := l.children()
	if len(children) == 0 || children[0] == nil {
		return 0
	}
	return children[0].UTF16TriviaWidth()
}

func (l *SyntaxList) FullWidth() int {
	width := 0
	for _, child := range l.children() {
		if child != nil {
			width += child.FullWidth()
		}
	}
	return width
}

func (l *SyntaxList) ByteFullWidth() int {
	return l.FullWidth()
}

func (l *SyntaxList) UTF16FullWidth() int {
	width := 0
	for _, child := range l.children() {
		if child != nil {
			width += child.UTF16FullWidth()
		}
	}
	return width
}

func (l *SyntaxList) IsToken() bool {
	return false
}

func (l *SyntaxList) Diagnostics() []Diagnostic {
	if l == nil {
		return []Diagnostic{}
	}
	return append([]Diagnostic{}, l.diagnostics...)
}

func (l *SyntaxList) ContainedDiagnostics() []Diagnostic {
	var out []Diagnostic
	if l != nil {
		out = append(out, l.diagnostics...)
	}
	for _, child := range l.children() {
		if child != nil {
			out = append(out, child.ContainedDiagnostics()...)
		}
	}
	if out == nil {
		return []Diagnostic{}
	}
	return out
}

func (l *SyntaxList) GetContainedSyntaxDiagnostics() []Diagnostic {
	return GetContainedDiagnosticsPreserveDuplicates(l, DiagnosticsIncludeSyntactic)
}

func (l *SyntaxList) Children() []SyntaxElement {
	return append([]SyntaxElement{}, l.children()...)
}

func (l *SyntaxList) Elements() []SyntaxElement {
	return l.Children()
}

func (l *SyntaxList) Text(includeTrivia IncludeTrivia) string {
	return syntaxChildrenText(l.children(), includeTrivia)
}

func (l *SyntaxList) Clone(includeDiagnostics bool) SyntaxElement {
	if l == nil {
		return (*SyntaxList)(nil)
	}
	clone := NewTypedSyntaxListWithDiagnostics(
		l.elementType,
		cloneSyntaxElements(l.elements, includeDiagnostics),
		cloneDiagnostics(l.diagnostics, includeDiagnostics)...,
	)
	if l.semanticInfo != nil {
		clone.SetSemanticInfo(*l.semanticInfo)
	}
	return clone
}

func (l *SyntaxList) children() []SyntaxElement {
	if l == nil {
		return nil
	}
	return l.elements
}

type SyntaxListOf[T SyntaxElement] struct {
	*SyntaxList
}

func AsSyntaxListOf[T SyntaxElement](list *SyntaxList) *SyntaxListOf[T] {
	if list == nil {
		return nil
	}
	return &SyntaxListOf[T]{SyntaxList: list}
}

func (l *SyntaxListOf[T]) At(index int) T {
	var zero T
	if l == nil || l.SyntaxList == nil {
		return zero
	}
	element := l.SyntaxList.At(index)
	if element == nil {
		return zero
	}
	if typed, ok := WrapGeneratedSyntaxElement(element).(T); ok {
		return typed
	}
	return zero
}

func (l *SyntaxListOf[T]) Elements() []T {
	if l == nil || l.SyntaxList == nil {
		return nil
	}
	out := make([]T, 0, l.Count())
	for i := 0; i < l.Count(); i++ {
		out = append(out, l.At(i))
	}
	return out
}

type SeparatedElementOf[T SyntaxElement] struct {
	*SeparatedElement
}

func AsSeparatedElementOf[T SyntaxElement](element *SeparatedElement) *SeparatedElementOf[T] {
	if element == nil {
		return nil
	}
	return &SeparatedElementOf[T]{SeparatedElement: element}
}

func (e *SeparatedElementOf[T]) Element() T {
	var zero T
	if e == nil || e.SeparatedElement == nil {
		return zero
	}
	element := e.SeparatedElement.Element()
	if element == nil {
		return zero
	}
	if typed, ok := WrapGeneratedSyntaxElement(element).(T); ok {
		return typed
	}
	return zero
}

type SeparatedSyntaxListOf[T SyntaxElement] struct {
	*SyntaxList
}

func AsSeparatedSyntaxListOf[T SyntaxElement](list *SyntaxList) *SeparatedSyntaxListOf[T] {
	if list == nil {
		return nil
	}
	return &SeparatedSyntaxListOf[T]{SyntaxList: list}
}

func (l *SeparatedSyntaxListOf[T]) At(index int) *SeparatedElementOf[T] {
	if l == nil || l.SyntaxList == nil {
		return nil
	}
	element, _ := l.SyntaxList.At(index).(*SeparatedElement)
	return AsSeparatedElementOf[T](element)
}

func (l *SeparatedSyntaxListOf[T]) Elements() []*SeparatedElementOf[T] {
	if l == nil || l.SyntaxList == nil {
		return nil
	}
	out := make([]*SeparatedElementOf[T], 0, l.Count())
	for i := 0; i < l.Count(); i++ {
		out = append(out, l.At(i))
	}
	return out
}

type SeparatedElement struct {
	syntaxOwnership
	element     SyntaxElement
	separator   *SyntaxToken
	diagnostics []Diagnostic
}

func NewSeparatedElement(element SyntaxElement, separator *SyntaxToken) *SeparatedElement {
	return NewSeparatedElementWithDiagnostics(element, separator)
}

func NewSeparatedElementWithDiagnostics(element SyntaxElement, separator *SyntaxToken, diagnostics ...Diagnostic) *SeparatedElement {
	separated := &SeparatedElement{
		element:     element,
		separator:   separator,
		diagnostics: append([]Diagnostic{}, diagnostics...),
	}
	initializeSyntaxElement(separated)
	return separated
}

func EmptySeparatedElement() *SeparatedElement {
	return NewSeparatedElement(nil, nil)
}

func (e *SeparatedElement) Kind() SyntaxKind {
	return SyntaxKindSeparatedElement
}

func (e *SeparatedElement) Element() SyntaxElement {
	if e == nil {
		return nil
	}
	return e.element
}

func (e *SeparatedElement) Separator() *SyntaxToken {
	if e == nil {
		return nil
	}
	return e.separator
}

func (e *SeparatedElement) ChildCount() int {
	return 2
}

func (e *SeparatedElement) Child(index int) SyntaxElement {
	if e == nil {
		return nil
	}
	switch index {
	case 0:
		return e.element
	case 1:
		if e.separator == nil {
			return nil
		}
		return e.separator
	default:
		panic("syntax child index out of range")
	}
}

func (e *SeparatedElement) GetChild(index int) SyntaxElement {
	return e.Child(index)
}

func (e *SeparatedElement) GetName(index int) string {
	switch index {
	case 0:
		return "Element"
	case 1:
		return "Separator"
	default:
		return ""
	}
}

func (e *SeparatedElement) CompletionHint(index int) CompletionHint {
	switch index {
	case 0:
		return CompletionHintInherit
	case 1:
		return CompletionHintSyntax
	default:
		return CompletionHintInherit
	}
}

func (e *SeparatedElement) IsOptional(index int) bool {
	return index == 1
}

func (e *SeparatedElement) Width() int {
	if e == nil {
		return 0
	}
	return e.FullWidth() - e.TriviaWidth()
}

func (e *SeparatedElement) ByteWidth() int {
	return e.Width()
}

func (e *SeparatedElement) UTF16Width() int {
	if e == nil {
		return 0
	}
	return e.UTF16FullWidth() - e.UTF16TriviaWidth()
}

func (e *SeparatedElement) TriviaWidth() int {
	children := e.children()
	if len(children) == 0 || children[0] == nil {
		return 0
	}
	return children[0].TriviaWidth()
}

func (e *SeparatedElement) ByteTriviaWidth() int {
	return e.TriviaWidth()
}

func (e *SeparatedElement) UTF16TriviaWidth() int {
	children := e.children()
	if len(children) == 0 || children[0] == nil {
		return 0
	}
	return children[0].UTF16TriviaWidth()
}

func (e *SeparatedElement) FullWidth() int {
	width := 0
	for _, child := range e.children() {
		if child != nil {
			width += child.FullWidth()
		}
	}
	return width
}

func (e *SeparatedElement) ByteFullWidth() int {
	return e.FullWidth()
}

func (e *SeparatedElement) UTF16FullWidth() int {
	width := 0
	for _, child := range e.children() {
		if child != nil {
			width += child.UTF16FullWidth()
		}
	}
	return width
}

func (e *SeparatedElement) IsToken() bool {
	return false
}

func (e *SeparatedElement) Diagnostics() []Diagnostic {
	if e == nil {
		return []Diagnostic{}
	}
	return append([]Diagnostic{}, e.diagnostics...)
}

func (e *SeparatedElement) ContainedDiagnostics() []Diagnostic {
	var out []Diagnostic
	if e != nil {
		out = append(out, e.diagnostics...)
	}
	for _, child := range e.children() {
		if child != nil {
			out = append(out, child.ContainedDiagnostics()...)
		}
	}
	if out == nil {
		return []Diagnostic{}
	}
	return out
}

func (e *SeparatedElement) GetContainedSyntaxDiagnostics() []Diagnostic {
	return GetContainedDiagnosticsPreserveDuplicates(e, DiagnosticsIncludeSyntactic)
}

func (e *SeparatedElement) Children() []SyntaxElement {
	return append([]SyntaxElement{}, e.children()...)
}

func (e *SeparatedElement) Text(includeTrivia IncludeTrivia) string {
	return syntaxChildrenText(e.children(), includeTrivia)
}

func (e *SeparatedElement) Clone(includeDiagnostics bool) SyntaxElement {
	if e == nil {
		return (*SeparatedElement)(nil)
	}
	var element SyntaxElement
	if e.element != nil {
		element = e.element.Clone(includeDiagnostics)
	}
	var separator *SyntaxToken
	if e.separator != nil {
		separator, _ = e.separator.Clone(includeDiagnostics).(*SyntaxToken)
	}
	return NewSeparatedElementWithDiagnostics(
		element,
		separator,
		cloneDiagnostics(e.diagnostics, includeDiagnostics)...,
	)
}

func (e *SeparatedElement) children() []SyntaxElement {
	if e == nil {
		return nil
	}
	if e.separator == nil {
		return []SyntaxElement{e.element, nil}
	}
	return []SyntaxElement{e.element, e.separator}
}

func syntaxChildrenText(children []SyntaxElement, includeTrivia IncludeTrivia) string {
	var b strings.Builder
	writeSyntaxChildrenText(&b, children, includeTrivia)
	return b.String()
}
