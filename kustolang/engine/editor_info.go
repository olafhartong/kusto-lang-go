package kustolang

// Ported from:
//   src/Kusto.Language/Editor/FormattedText.cs
//   src/Kusto.Language/Editor/SyntaxReference.cs
//   src/Kusto.Language/Editor/ClusterReference.cs
//   src/Kusto.Language/Editor/DatabaseReference.cs
//   src/Kusto.Language/Editor/OutlineRange.cs
//   src/Kusto.Language/Editor/OutlineInfo.cs
//   src/Kusto.Language/Editor/ClassificationKind.cs
//   src/Kusto.Language/Editor/ClassifiedRange.cs
//   src/Kusto.Language/Editor/ClassificationResult.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type FormattedText struct {
	Text     string
	Position int
}

func NewFormattedText(newText string, newPosition int) FormattedText {
	return FormattedText{Text: newText, Position: newPosition}
}

type SyntaxReference struct {
	Range TextRange
}

func NewSyntaxReference(start int, length int) SyntaxReference {
	return SyntaxReference{Range: NewTextRange(start, length)}
}

func NewSyntaxReferenceRange(rng TextRange) SyntaxReference {
	return SyntaxReference{Range: rng}
}

func (r SyntaxReference) Start() int {
	return r.Range.Start
}

func (r SyntaxReference) Length() int {
	return r.Range.Length
}

func (r SyntaxReference) End() int {
	return r.Range.End()
}

type ClusterReference struct {
	SyntaxReference
	Cluster string
}

func NewClusterReference(cluster string, start int, length int) ClusterReference {
	return ClusterReference{
		SyntaxReference: NewSyntaxReference(start, length),
		Cluster:         cluster,
	}
}

type DatabaseReference struct {
	SyntaxReference
	Database string
	Cluster  string
}

func NewDatabaseReference(database string, cluster string, start int, length int) DatabaseReference {
	return DatabaseReference{
		SyntaxReference: NewSyntaxReference(start, length),
		Database:        database,
		Cluster:         cluster,
	}
}

type OutlineRange struct {
	Start         int
	Length        int
	CollapsedText string
}

func NewOutlineRange(start int, length int, collapsedText string) OutlineRange {
	return OutlineRange{Start: start, Length: length, CollapsedText: collapsedText}
}

func NewOutlineRangeOptional(start int, length int, collapsedText *string) OutlineRange {
	text := ""
	if collapsedText != nil {
		text = *collapsedText
	}
	return NewOutlineRange(start, length, text)
}

type OutlineInfo struct {
	Ranges []OutlineRange
}

func NewOutlineInfo(ranges []OutlineRange) OutlineInfo {
	return OutlineInfo{Ranges: ListToReadOnly(ranges)}
}

func (info OutlineInfo) RangesList() []OutlineRange {
	return ListToReadOnly(info.Ranges)
}

var OutlineInfoEmpty = NewOutlineInfo(nil)

type ClassificationKind int

const (
	ClassificationKindPlainText ClassificationKind = iota
	ClassificationKindComment
	ClassificationKindPunctuation
	ClassificationKindDirective
	ClassificationKindLiteral
	ClassificationKindStringLiteral
	ClassificationKindType
	ClassificationKindColumn
	ClassificationKindTable
	ClassificationKindDatabase
	ClassificationKindFunction
	ClassificationKindParameter
	ClassificationKindVariable
	ClassificationKindIdentifier
	ClassificationKindClientParameter
	ClassificationKindQueryParameter
	ClassificationKindScalarOperator
	ClassificationKindMathOperator
	ClassificationKindQueryOperator
	ClassificationKindCommand
	ClassificationKindKeyword
	ClassificationKindMaterializedView
	ClassificationKindSchemaMember
	ClassificationKindSignatureParameter
	ClassificationKindOption
)

type ClassifiedRange struct {
	Kind   ClassificationKind
	Start  int
	Length int
}

func NewClassifiedRange(kind ClassificationKind, start int, length int) ClassifiedRange {
	return ClassifiedRange{Kind: kind, Start: start, Length: length}
}

func (r ClassifiedRange) End() int {
	return r.Start + r.Length
}

type ClassificationInfo struct {
	Classifications []ClassifiedRange
}

func NewClassificationInfo(classifications []ClassifiedRange) ClassificationInfo {
	return ClassificationInfo{Classifications: ListToReadOnly(classifications)}
}

func (info ClassificationInfo) ClassificationsList() []ClassifiedRange {
	return ListToReadOnly(info.Classifications)
}

var ClassificationInfoEmpty = NewClassificationInfo(nil)
