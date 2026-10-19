package kustolang

import "strings"

// Ported from the graph pattern validation path in
// src/Kusto.Language/Binder/Binder_NodeBinder.cs.

type locatedGraphPatternDiagnostic struct {
	diagnostic Diagnostic
	location   SyntaxElement
}

func CheckGraphMatchPatterns(patterns *SeparatedSyntaxListOf[*GraphMatchPattern]) []Diagnostic {
	located := checkGraphMatchPatternsLocated(patterns)
	diagnostics := make([]Diagnostic, 0, len(located))
	for _, item := range located {
		diagnostic := item.diagnostic
		if item.location != nil {
			diagnostic = diagnostic.WithLocationText(strings.TrimSpace(item.location.Text(IncludeTriviaAll)))
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}

func checkGraphMatchPatternsLocated(patterns *SeparatedSyntaxListOf[*GraphMatchPattern]) []locatedGraphPatternDiagnostic {
	if patterns == nil || patterns.Count() == 0 {
		return []locatedGraphPatternDiagnostic{{DiagnosticMissingGraphMatchPattern(), patterns}}
	}
	diagnostics := []locatedGraphPatternDiagnostic{}
	for i := 0; i < patterns.Count(); i++ {
		element := patterns.At(i)
		if element == nil || element.Element() == nil {
			continue
		}
		diagnostics = append(diagnostics, checkGraphMatchPatternLocated(element.Element())...)
	}
	return diagnostics
}

func CheckGraphMatchPattern(pattern *GraphMatchPattern) []Diagnostic {
	located := checkGraphMatchPatternLocated(pattern)
	diagnostics := make([]Diagnostic, 0, len(located))
	for _, item := range located {
		diagnostic := item.diagnostic
		if item.location != nil {
			diagnostic = diagnostic.WithLocationText(strings.TrimSpace(item.location.Text(IncludeTriviaAll)))
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}

func checkGraphMatchPatternLocated(pattern *GraphMatchPattern) []locatedGraphPatternDiagnostic {
	if pattern == nil || pattern.PatternElements() == nil || pattern.PatternElements().Count() == 0 {
		return []locatedGraphPatternDiagnostic{{DiagnosticMissingGraphMatchPatternElement(), pattern}}
	}

	diagnostics := []locatedGraphPatternDiagnostic{}
	elements := pattern.PatternElements()
	for i := 0; i < elements.Count(); i++ {
		element := elements.At(i)
		expectNode := i%2 == 0
		if expectNode && !isGraphMatchPatternNode(element) {
			diagnostics = append(diagnostics, locatedGraphPatternDiagnostic{
				DiagnosticGraphMatchPatternSyntaxError("node", graphMatchPatternElementType(element)), element})
		} else if !expectNode && !isGraphMatchPatternEdge(element) {
			diagnostics = append(diagnostics, locatedGraphPatternDiagnostic{
				DiagnosticGraphMatchPatternSyntaxError("edge", graphMatchPatternElementType(element)), element})
		}
	}

	last := elements.At(elements.Count() - 1)
	if !isGraphMatchPatternNode(last) {
		diagnostics = append(diagnostics, locatedGraphPatternDiagnostic{DiagnosticMissingGraphMatchPatternElement(), last})
	}

	return diagnostics
}

func isGraphMatchPatternNode(element GraphMatchPatternNotation) bool {
	_, ok := element.(*GraphMatchPatternNode)
	return ok
}

func isGraphMatchPatternEdge(element GraphMatchPatternNotation) bool {
	_, ok := element.(*GraphMatchPatternEdge)
	return ok
}

func graphMatchPatternElementType(element GraphMatchPatternNotation) string {
	switch element.(type) {
	case *GraphMatchPatternNode:
		return "node"
	case *GraphMatchPatternEdge:
		return "edge"
	default:
		return ""
	}
}
