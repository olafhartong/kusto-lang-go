package kustolang

import "fmt"

// Ported from:
//   src/Kusto.Language/Editor/AnalyzerInfo.cs
//   src/Kusto.Language/Editor/ClientParameter.cs
//   src/Kusto.Language/Editor/ClientDirectiveArgument.cs
//   src/Kusto.Language/Editor/CodeKinds.cs
//   src/Kusto.Language/Editor/CodeServiceFeatures.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	CodeKindsQuery     = "Query"
	CodeKindsCommand   = "Command"
	CodeKindsDirective = "Directive"
	CodeKindsUnknown   = "Unknown"
)

const (
	CodeServiceFeatureDiagnostics        = "Diagnostics"
	CodeServiceFeatureClassification     = "Classification"
	CodeServiceFeatureCompletion         = "Completion"
	CodeServiceFeatureOutlining          = "Outlining"
	CodeServiceFeatureQuickInfo          = "QuickInfo"
	CodeServiceFeatureClusterReferences  = "ClusterReferences"
	CodeServiceFeatureDatabaseReferences = "DatabaseReferences"
	CodeServiceFeatureMinimalText        = "MinimalText"
	CodeServiceFeatureFormatting         = "Formatting"
	CodeServiceFeatureClientParameters   = "ClientParameters"
)

type AnalyzerInfo struct {
	Name        string
	Diagnostics []Diagnostic
}

func NewAnalyzerInfo(name string, diagnostics []Diagnostic) AnalyzerInfo {
	return AnalyzerInfo{
		Name:        name,
		Diagnostics: ListToReadOnly(diagnostics),
	}
}

func (info AnalyzerInfo) DiagnosticsList() []Diagnostic {
	return ListToReadOnly(info.Diagnostics)
}

func cloneAnalyzerInfos(infos []AnalyzerInfo) []AnalyzerInfo {
	if len(infos) == 0 {
		return EmptyReadOnlyList[AnalyzerInfo]()
	}
	clones := make([]AnalyzerInfo, len(infos))
	for i, info := range infos {
		clones[i] = NewAnalyzerInfo(info.Name, info.Diagnostics)
	}
	return clones
}

type ClientParameter struct {
	SyntaxReference
	Name  string
	Index *int
}

func NewClientParameter(name string, index *int, rng TextRange) ClientParameter {
	return ClientParameter{
		SyntaxReference: NewSyntaxReferenceRange(rng),
		Name:            name,
		Index:           cloneOptionalInt(index),
	}
}

func NewClientParameterAt(name string, index *int, start int, length int) ClientParameter {
	return NewClientParameter(name, index, NewTextRange(start, length))
}

func NewClientParameterRange(name string, rng TextRange) ClientParameter {
	return NewClientParameter(name, nil, rng)
}

func NewClientParameterReference(name string, start int, length int) ClientParameter {
	return NewClientParameterAt(name, nil, start, length)
}

func (p ClientParameter) String() string {
	if p.Index != nil {
		return fmt.Sprintf("%s[%d]", p.Name, *p.Index)
	}
	return p.Name
}

func cloneOptionalInt(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

type ClientDirectiveArgument struct {
	Name  string
	Text  string
	Value any
}

func NewClientDirectiveArgument(name string, text string, value any) ClientDirectiveArgument {
	return ClientDirectiveArgument{Name: name, Text: text, Value: value}
}
