package kustolang

import (
	"strconv"
	"strings"
	"sync"
)

// Ported from:
//   src/Kusto.Language/Editor/CommonCodeService.cs
//   src/Kusto.Language/Editor/UnknownCodeService.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CommonCodeService struct {
	text                 string
	clientParametersOnce sync.Once
	clientParameters     []ClientParameter
}

func NewCommonCodeService(text string) *CommonCodeService {
	return &CommonCodeService{text: text}
}

func (s *CommonCodeService) Kind() string {
	return ""
}

func (s *CommonCodeService) Text() string {
	if s == nil {
		return ""
	}
	return s.text
}

func (s *CommonCodeService) IsFeatureSupported(feature string, position int) bool {
	return true
}

func (s *CommonCodeService) TryGetCachedDiagnostics() ([]Diagnostic, bool) {
	return nil, false
}

func (s *CommonCodeService) GetDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return EmptyReadOnlyList[Diagnostic]()
}

func (s *CommonCodeService) TryGetCachedAnalyzerDiagnostics() ([]Diagnostic, bool) {
	return nil, false
}

func (s *CommonCodeService) GetAnalyzerDiagnostics(waitForAnalysis bool, cancellationToken CancellationToken) []Diagnostic {
	return EmptyReadOnlyList[Diagnostic]()
}

func (s *CommonCodeService) GetAnalyzers() []AnalyzerInfo {
	return EmptyReadOnlyList[AnalyzerInfo]()
}

func (s *CommonCodeService) GetCodeActions(position int, selectionStart int, selectionLength int, options CodeActionOptions, waitForAnalysis bool, actorName string, cancellationToken CancellationToken) CodeActionInfo {
	return CodeActionInfoNoActions
}

func (s *CommonCodeService) ApplyCodeAction(action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	return CodeActionResultNothing
}

func (s *CommonCodeService) GetClassifications(start int, length int, clipToRange bool, waitForAnalysis bool, cancellationToken CancellationToken) ClassificationInfo {
	classifications := []ClassifiedRange{NewClassifiedRange(ClassificationKindPlainText, start, length)}
	clientParameterClassifications := s.GetClientParametersClassifications(start, length, clipToRange)
	return NewClassificationInfo(CommonCodeServiceAddClassifications(classifications, clientParameterClassifications))
}

func (s *CommonCodeService) GetClientParametersClassifications(start int, length int, clipToRange bool) []ClassifiedRange {
	return clientParameterClassifications(s.GetClientParameters(), start, length, clipToRange)
}

func (s *CommonCodeService) getClientParametersClassificationsBytes(start int, length int, clipToRange bool) []ClassifiedRange {
	return clientParameterClassifications(s.getClientParametersBytes(), start, length, clipToRange)
}

func clientParameterClassifications(clientParameters []ClientParameter, start int, length int, clipToRange bool) []ClassifiedRange {
	end := start + length
	if len(clientParameters) == 0 {
		return EmptyReadOnlyList[ClassifiedRange]()
	}

	list := make([]ClassifiedRange, 0, len(clientParameters))
	for _, parameter := range clientParameters {
		if parameter.Start()+parameter.Length() < start || parameter.Start()+parameter.Length() > start+length {
			continue
		}

		parameterStart := parameter.Start()
		parameterEnd := parameter.End()
		parameterLength := parameter.Length()
		if clipToRange {
			if parameterStart < start {
				parameterStart = start
			}
			if parameterEnd > end {
				parameterEnd = end
			}
			parameterLength = parameterEnd - parameterStart
		}

		list = append(list, NewClassifiedRange(ClassificationKindClientParameter, parameterStart, parameterLength))
	}

	return ListToReadOnly(list)
}

func CommonCodeServiceAddClassifications(classifications1 []ClassifiedRange, classifications2 []ClassifiedRange) []ClassifiedRange {
	if len(classifications2) == 0 {
		return ListToReadOnly(classifications1)
	}
	if len(classifications1) == 0 {
		return ListToReadOnly(classifications2)
	}

	list := make([]ClassifiedRange, 0, len(classifications1)+len(classifications2))
	index1 := 0
	index2 := 0
	var class1 *ClassifiedRange
	var class2 *ClassifiedRange

	for {
		if class1 == nil && index1 < len(classifications1) {
			value := classifications1[index1]
			class1 = &value
		}
		if class2 == nil && index2 < len(classifications2) {
			value := classifications2[index2]
			class2 = &value
		}

		switch {
		case class1 == nil && class2 == nil:
			return ListToReadOnly(list)
		case class1 != nil && class2 == nil:
			list = append(list, *class1)
			class1 = nil
			index1++
		case class2 != nil && class1 == nil:
			list = append(list, *class2)
			class2 = nil
			index2++
		case class1.End() < class2.Start:
			list = append(list, *class1)
			class1 = nil
			index1++
		case class2.End() < class1.Start:
			list = append(list, *class2)
			class2 = nil
			index2++
		case class1.Start >= class2.Start && class1.End() <= class2.End():
			class1 = nil
			index1++
		case class1.Start < class2.Start:
			length := class2.Start - class1.Start
			list = append(list, NewClassifiedRange(class1.Kind, class1.Start, length))
			if class1.Length > length {
				value := NewClassifiedRange(class1.Kind, class1.Start+length, class1.Length-length)
				class1 = &value
			} else {
				class1 = nil
				index1++
			}
		case class2.Start < class1.Start:
			length := class1.Start - class2.Start
			list = append(list, NewClassifiedRange(class2.Kind, class2.Start, length))
			if class2.Length > length {
				value := NewClassifiedRange(class2.Kind, class2.Start+length, class2.Length-length)
				class2 = &value
			} else {
				class2 = nil
				index2++
			}
		case class1.End() <= class2.End():
			class1 = nil
			index1++
		default:
			length := class2.Length
			list = append(list, *class2)
			class2 = nil
			index2++
			value := NewClassifiedRange(class1.Kind, class1.Start+length, class1.Length-length)
			class1 = &value
		}
	}
}

func (s *CommonCodeService) GetOutlines(options OutliningOptions, cancellationToken CancellationToken) OutlineInfo {
	text := s.Text()
	tokens := Lex(text)
	if len(tokens) > 0 && tokens[0].Text != "" {
		start := 0
		end := start + TrimEnd(text)
		collapsedTextStart := len(tokens[0].Trivia)
		nextLineBreakStart := NextLineBreakStart(text, collapsedTextStart)
		collapsedTextEnd := len(text)
		if nextLineBreakStart >= 0 {
			collapsedTextEnd = nextLineBreakStart
		}
		collapsedTextLength := TrimEnd(text, collapsedTextStart, collapsedTextEnd-collapsedTextStart)
		collapsedText := text[collapsedTextStart:collapsedTextStart+collapsedTextLength] + " ..."
		convertedStart, convertedLength := byteRangeToUTF16(text, start, end-start)
		return NewOutlineInfo([]OutlineRange{NewOutlineRange(convertedStart, convertedLength, collapsedText)})
	}
	return OutlineInfoEmpty
}

func (s *CommonCodeService) ShouldAutoComplete(position int, key rune, cancellationToken CancellationToken) bool {
	return false
}

func (s *CommonCodeService) GetCompletionItems(position int, options CompletionOptions, cancellationToken CancellationToken) CompletionInfo {
	return CompletionInfoEmpty
}

func (s *CommonCodeService) GetQuickInfo(position int, options QuickInfoOptions, cancellationToken CancellationToken) QuickInfo {
	return QuickInfoEmpty
}

func (s *CommonCodeService) GetElement(position int, cancellationToken CancellationToken) TextRange {
	return NewTextRange(0, 0)
}

func (s *CommonCodeService) GetRelatedElements(position int, options FindRelatedOptions, cancellationToken CancellationToken) RelatedInfo {
	return RelatedInfoEmpty
}

func (s *CommonCodeService) GetClusterReferences(cancellationToken CancellationToken) []ClusterReference {
	return EmptyReadOnlyList[ClusterReference]()
}

func (s *CommonCodeService) GetDatabaseReferences(cancellationToken CancellationToken) []DatabaseReference {
	return EmptyReadOnlyList[DatabaseReference]()
}

func (s *CommonCodeService) GetMinimalText(kind MinimalTextKind, cancellationToken CancellationToken) string {
	return tokensMinimalText(Lex(s.Text()), commonCodeServiceIncludeTrivia(kind))
}

func commonCodeServiceIncludeTrivia(kind MinimalTextKind) IncludeTrivia {
	return KustoCodeServiceIncludeTrivia(kind)
}

func tokensMinimalText(tokens []LexicalToken, includeTrivia IncludeTrivia) string {
	var builder strings.Builder
	for i, token := range tokens {
		syntaxToken := SyntaxTokenFromLexical(token)
		builder.WriteString(syntaxToken.TextWithContext(includeTrivia, i == 0, i == len(tokens)-1))
	}
	return builder.String()
}

func (s *CommonCodeService) GetFormattedText(options FormattingOptions, caretPosition int, cancellationToken CancellationToken) FormattedText {
	return NewFormattedText(s.Text(), caretPosition)
}

func (s *CommonCodeService) GetClientParameters() []ClientParameter {
	if s == nil {
		return EmptyReadOnlyList[ClientParameter]()
	}
	return clientParametersToUTF16(s.text, s.getClientParametersBytes())
}

func (s *CommonCodeService) getClientParametersBytes() []ClientParameter {
	if s == nil {
		return EmptyReadOnlyList[ClientParameter]()
	}
	s.clientParametersOnce.Do(func() {
		s.clientParameters = s.scanClientParameters(s.text)
	})
	return ListToReadOnly(s.clientParameters)
}

func clientParametersToUTF16(text string, parameters []ClientParameter) []ClientParameter {
	converted := make([]ClientParameter, 0, len(parameters))
	for _, parameter := range parameters {
		start, length := byteRangeToUTF16(text, parameter.Start(), parameter.Length())
		converted = append(converted, NewClientParameterAt(parameter.Name, parameter.Index, start, length))
	}
	return ListToReadOnly(converted)
}

func (s *CommonCodeService) scanClientParameters(blockText string) []ClientParameter {
	var list []ClientParameter
	start := 0
	for start < len(blockText) {
		openBrace := strings.Index(blockText[start:], "{")
		if openBrace < 0 {
			break
		}
		openBrace += start
		length, nameStart, nameLength, indexStart, indexLength := ScanClientParameterParts(blockText, openBrace)
		if length > 0 {
			var index *int
			if indexLength > 0 {
				if value, err := strconv.ParseInt(blockText[indexStart:indexStart+indexLength], 10, 32); err == nil {
					parsed := int(value)
					index = &parsed
				}
			}
			list = append(list, NewClientParameterAt(blockText[nameStart:nameStart+nameLength], index, openBrace, length))
			start = openBrace + length
		} else {
			start = openBrace + 1
		}
	}
	return ListToReadOnly(list)
}

type UnknownCodeService struct {
	*CommonCodeService
}

func NewUnknownCodeService(text string) *UnknownCodeService {
	return &UnknownCodeService{CommonCodeService: NewCommonCodeService(text)}
}

func (s *UnknownCodeService) Kind() string {
	return CodeKindsUnknown
}
